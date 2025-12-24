package http3

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/quic-go/quic-go/http3"
)

// LoadTester performs HTTP/3 load testing
type LoadTester struct {
	config  *LoadTestConfig
	results *LoadTestResults
	client  *http.Client
	mu      sync.RWMutex
}

// LoadTestConfig holds HTTP/3 load test configuration
type LoadTestConfig struct {
	TargetURL              string            `json:"target_url"`
	Duration               time.Duration     `json:"duration"`
	ConcurrentConnections  int               `json:"concurrent_connections"`
	RequestsPerConnection  int               `json:"requests_per_connection"`
	RequestPattern         string            `json:"request_pattern"` // "sequential", "parallel", "burst"
	Headers                map[string]string `json:"headers,omitempty"`
	Method                 string            `json:"method"`
	BodySize               int               `json:"body_size"`
	ThinkTime              time.Duration     `json:"think_time"`
	TLSConfig              *tls.Config       `json:"-"`
	FollowRedirects        bool              `json:"follow_redirects"`
	Timeout                time.Duration     `json:"timeout"`
	UserAgent              string            `json:"user_agent"`
}

// LoadTestResults holds HTTP/3 load test results
type LoadTestResults struct {
	LoadTestID         string                 `json:"load_test_id"`
	Status             string                 `json:"status"` // "running", "completed", "failed"
	CreatedAt          time.Time              `json:"created_at"`
	StartedAt          *time.Time             `json:"started_at,omitempty"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
	Config             *LoadTestConfig        `json:"config"`
	
	// Results
	TotalRequests      int64                  `json:"total_requests"`
	SuccessfulRequests int64                  `json:"successful_requests"`
	FailedRequests     int64                  `json:"failed_requests"`
	AvgResponseTime    float64                `json:"avg_response_time_ms"`
	P50ResponseTime    float64                `json:"p50_response_time_ms"`
	P95ResponseTime    float64                `json:"p95_response_time_ms"`
	P99ResponseTime    float64                `json:"p99_response_time_ms"`
	RequestsPerSecond  float64                `json:"requests_per_second"`
	BytesTransferred   int64                  `json:"bytes_transferred"`
	ErrorRate          float64                `json:"error_rate"`
	StatusCodes        map[string]int64       `json:"status_codes"`
	Errors             map[string]int64       `json:"errors"`
	
	// Detailed metrics
	ResponseTimes      []float64              `json:"-"` // Not exported in JSON
	ConnectionMetrics  *ConnectionMetrics     `json:"connection_metrics"`
	
	mu sync.RWMutex
}

// ConnectionMetrics holds connection-level metrics
type ConnectionMetrics struct {
	ConnectionsCreated   int64   `json:"connections_created"`
	ConnectionsReused    int64   `json:"connections_reused"`
	ConnectionsFailed    int64   `json:"connections_failed"`
	AvgConnectionTime    float64 `json:"avg_connection_time_ms"`
	TLSHandshakeTime     float64 `json:"avg_tls_handshake_time_ms"`
	DNSLookupTime        float64 `json:"avg_dns_lookup_time_ms"`
	
	mu sync.RWMutex
}

// RequestResult holds individual request result
type RequestResult struct {
	StartTime      time.Time
	EndTime        time.Time
	StatusCode     int
	ResponseSize   int64
	Error          error
	ConnectionTime time.Duration
	DNSTime        time.Duration
	TLSTime        time.Duration
}

// NewLoadTester creates a new HTTP/3 load tester
func NewLoadTester(config *LoadTestConfig) *LoadTester {
	loadTestID := fmt.Sprintf("http3_load_%d", time.Now().Unix())
	
	results := &LoadTestResults{
		LoadTestID:        loadTestID,
		Status:            "created",
		CreatedAt:         time.Now(),
		Config:            config,
		StatusCodes:       make(map[string]int64),
		Errors:            make(map[string]int64),
		ResponseTimes:     make([]float64, 0),
		ConnectionMetrics: &ConnectionMetrics{},
	}
	
	// Configure HTTP/3 client
	tlsConfig := config.TLSConfig
	if tlsConfig == nil {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true, // For testing
		}
	}
	
	roundTripper := &http3.RoundTripper{
		TLSClientConfig: tlsConfig,
	}
	
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	client := &http.Client{
		Transport: roundTripper,
		Timeout:   timeout,
	}
	
	if !config.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	
	return &LoadTester{
		config:  config,
		results: results,
		client:  client,
	}
}

// Start starts the load test
func (lt *LoadTester) Start(ctx context.Context) error {
	lt.results.mu.Lock()
	lt.results.Status = "running"
	now := time.Now()
	lt.results.StartedAt = &now
	lt.results.mu.Unlock()
	
	// Create context with timeout
	testCtx, cancel := context.WithTimeout(ctx, lt.config.Duration)
	defer cancel()
	
	// Start load test
	return lt.runLoadTest(testCtx)
}

// runLoadTest executes the load test
func (lt *LoadTester) runLoadTest(ctx context.Context) error {
	var wg sync.WaitGroup
	resultsChan := make(chan *RequestResult, lt.config.ConcurrentConnections*lt.config.RequestsPerConnection)
	
	// Start result collector
	go lt.collectResults(ctx, resultsChan)
	
	// Start concurrent connections
	for i := 0; i < lt.config.ConcurrentConnections; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()
			lt.runConnection(ctx, connID, resultsChan)
		}(i)
	}
	
	// Wait for all connections to complete
	wg.Wait()
	close(resultsChan)
	
	// Finalize results
	lt.finalizeResults()
	
	return nil
}

// runConnection runs requests for a single connection
func (lt *LoadTester) runConnection(ctx context.Context, connID int, resultsChan chan<- *RequestResult) {
	switch lt.config.RequestPattern {
	case "parallel":
		lt.runParallelRequests(ctx, connID, resultsChan)
	case "burst":
		lt.runBurstRequests(ctx, connID, resultsChan)
	default: // "sequential"
		lt.runSequentialRequests(ctx, connID, resultsChan)
	}
}

// runSequentialRequests runs requests sequentially
func (lt *LoadTester) runSequentialRequests(ctx context.Context, connID int, resultsChan chan<- *RequestResult) {
	for i := 0; i < lt.config.RequestsPerConnection; i++ {
		select {
		case <-ctx.Done():
			return
		default:
		}
		
		result := lt.executeRequest(ctx, connID, i)
		resultsChan <- result
		
		// Think time between requests
		if lt.config.ThinkTime > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(lt.config.ThinkTime):
			}
		}
	}
}

// runParallelRequests runs requests in parallel
func (lt *LoadTester) runParallelRequests(ctx context.Context, connID int, resultsChan chan<- *RequestResult) {
	var wg sync.WaitGroup
	
	for i := 0; i < lt.config.RequestsPerConnection; i++ {
		wg.Add(1)
		go func(reqID int) {
			defer wg.Done()
			
			select {
			case <-ctx.Done():
				return
			default:
			}
			
			result := lt.executeRequest(ctx, connID, reqID)
			resultsChan <- result
		}(i)
	}
	
	wg.Wait()
}

// runBurstRequests runs requests in bursts
func (lt *LoadTester) runBurstRequests(ctx context.Context, connID int, resultsChan chan<- *RequestResult) {
	burstSize := 10 // 10 requests per burst
	burstInterval := 1 * time.Second
	
	for burst := 0; burst < (lt.config.RequestsPerConnection+burstSize-1)/burstSize; burst++ {
		var wg sync.WaitGroup
		
		// Execute burst
		for i := 0; i < burstSize && burst*burstSize+i < lt.config.RequestsPerConnection; i++ {
			wg.Add(1)
			go func(reqID int) {
				defer wg.Done()
				
				select {
				case <-ctx.Done():
					return
				default:
				}
				
				result := lt.executeRequest(ctx, connID, reqID)
				resultsChan <- result
			}(burst*burstSize + i)
		}
		
		wg.Wait()
		
		// Wait between bursts
		if burst < (lt.config.RequestsPerConnection+burstSize-1)/burstSize-1 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(burstInterval):
			}
		}
	}
}

// executeRequest executes a single HTTP request
func (lt *LoadTester) executeRequest(ctx context.Context, connID, reqID int) *RequestResult {
	result := &RequestResult{
		StartTime: time.Now(),
	}
	
	// Create request
	method := lt.config.Method
	if method == "" {
		method = "GET"
	}
	
	var body io.Reader
	if lt.config.BodySize > 0 {
		body = strings.NewReader(strings.Repeat("x", lt.config.BodySize))
	}
	
	req, err := http.NewRequestWithContext(ctx, method, lt.config.TargetURL, body)
	if err != nil {
		result.EndTime = time.Now()
		result.Error = err
		return result
	}
	
	// Set headers
	userAgent := lt.config.UserAgent
	if userAgent == "" {
		userAgent = "QUIC-Test-Suite/1.0"
	}
	req.Header.Set("User-Agent", userAgent)
	
	for key, value := range lt.config.Headers {
		req.Header.Set(key, value)
	}
	
	// Execute request
	resp, err := lt.client.Do(req)
	result.EndTime = time.Now()
	
	if err != nil {
		result.Error = err
		return result
	}
	defer resp.Body.Close()
	
	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = err
		return result
	}
	
	result.StatusCode = resp.StatusCode
	result.ResponseSize = int64(len(bodyBytes))
	
	return result
}

// collectResults collects and processes request results
func (lt *LoadTester) collectResults(ctx context.Context, resultsChan <-chan *RequestResult) {
	for {
		select {
		case <-ctx.Done():
			return
		case result, ok := <-resultsChan:
			if !ok {
				return
			}
			
			lt.processResult(result)
		}
	}
}

// processResult processes a single request result
func (lt *LoadTester) processResult(result *RequestResult) {
	lt.results.mu.Lock()
	defer lt.results.mu.Unlock()
	
	atomic.AddInt64(&lt.results.TotalRequests, 1)
	
	if result.Error != nil {
		atomic.AddInt64(&lt.results.FailedRequests, 1)
		
		errorType := "unknown"
		if result.Error != nil {
			errorType = result.Error.Error()
		}
		lt.results.Errors[errorType]++
	} else {
		atomic.AddInt64(&lt.results.SuccessfulRequests, 1)
		atomic.AddInt64(&lt.results.BytesTransferred, result.ResponseSize)
		
		// Record status code
		statusCode := fmt.Sprintf("%d", result.StatusCode)
		lt.results.StatusCodes[statusCode]++
		
		// Record response time
		responseTime := float64(result.EndTime.Sub(result.StartTime).Nanoseconds()) / 1e6
		lt.results.ResponseTimes = append(lt.results.ResponseTimes, responseTime)
	}
}

// finalizeResults calculates final statistics
func (lt *LoadTester) finalizeResults() {
	lt.results.mu.Lock()
	defer lt.results.mu.Unlock()
	
	now := time.Now()
	lt.results.CompletedAt = &now
	lt.results.Status = "completed"
	
	// Calculate response time statistics
	if len(lt.results.ResponseTimes) > 0 {
		// Sort response times for percentile calculation
		times := make([]float64, len(lt.results.ResponseTimes))
		copy(times, lt.results.ResponseTimes)
		
		// Simple sort (for production, use a more efficient algorithm)
		for i := 0; i < len(times); i++ {
			for j := i + 1; j < len(times); j++ {
				if times[i] > times[j] {
					times[i], times[j] = times[j], times[i]
				}
			}
		}
		
		// Calculate average
		sum := 0.0
		for _, t := range times {
			sum += t
		}
		lt.results.AvgResponseTime = sum / float64(len(times))
		
		// Calculate percentiles
		lt.results.P50ResponseTime = times[len(times)*50/100]
		lt.results.P95ResponseTime = times[len(times)*95/100]
		lt.results.P99ResponseTime = times[len(times)*99/100]
	}
	
	// Calculate requests per second
	if lt.results.StartedAt != nil && lt.results.CompletedAt != nil {
		duration := lt.results.CompletedAt.Sub(*lt.results.StartedAt).Seconds()
		if duration > 0 {
			lt.results.RequestsPerSecond = float64(lt.results.TotalRequests) / duration
		}
	}
	
	// Calculate error rate
	if lt.results.TotalRequests > 0 {
		lt.results.ErrorRate = float64(lt.results.FailedRequests) / float64(lt.results.TotalRequests)
	}
}

// GetResults returns the current test results
func (lt *LoadTester) GetResults() *LoadTestResults {
	lt.results.mu.RLock()
	defer lt.results.mu.RUnlock()
	
	// Return a copy (without response times array for performance)
	results := *lt.results
	results.ResponseTimes = nil
	
	return &results
}

// Stop stops the load test
func (lt *LoadTester) Stop() {
	lt.results.mu.Lock()
	defer lt.results.mu.Unlock()
	
	if lt.results.Status == "running" {
		lt.results.Status = "stopped"
		now := time.Now()
		lt.results.CompletedAt = &now
	}
}

// Close cleans up resources
func (lt *LoadTester) Close() error {
	if transport, ok := lt.client.Transport.(*http3.RoundTripper); ok {
		return transport.Close()
	}
	return nil
}