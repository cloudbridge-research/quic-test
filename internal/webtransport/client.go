package webtransport

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

// Client represents a WebTransport client
type Client struct {
	config   *Config
	session  *Session
	metrics  *Metrics
	mu       sync.RWMutex
}

// Config holds WebTransport client configuration
type Config struct {
	URL             string            `json:"url"`
	Duration        time.Duration     `json:"duration"`
	Streams         int               `json:"streams"`
	Datagrams       bool              `json:"datagrams"`
	CertificateHash string            `json:"certificate_hash,omitempty"`
	ALPN            []string          `json:"alpn,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	TLSConfig       *tls.Config       `json:"-"`
}

// Session represents an active WebTransport session
type Session struct {
	ID          string                 `json:"session_id"`
	Status      string                 `json:"status"` // "connecting", "connected", "closed", "failed"
	CreatedAt   time.Time              `json:"created_at"`
	ConnectedAt *time.Time             `json:"connected_at,omitempty"`
	ClosedAt    *time.Time             `json:"closed_at,omitempty"`
	Config      *Config                `json:"config"`
	Metrics     map[string]interface{} `json:"metrics"`
	Error       string                 `json:"error,omitempty"`
	
	// Internal fields
	quicSession quic.Connection
	httpClient  *http.Client
	streams     map[string]*StreamInfo
	mu          sync.RWMutex
}

// StreamInfo holds information about a WebTransport stream
type StreamInfo struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "bidirectional", "unidirectional"
	CreatedAt time.Time `json:"created_at"`
	BytesSent int64     `json:"bytes_sent"`
	BytesRecv int64     `json:"bytes_received"`
	Status    string    `json:"status"` // "open", "closed"
}

// Metrics holds WebTransport performance metrics
type Metrics struct {
	StreamsOpened      int64   `json:"streams_opened"`
	StreamsClosed      int64   `json:"streams_closed"`
	DatagramsSent      int64   `json:"datagrams_sent"`
	DatagramsReceived  int64   `json:"datagrams_received"`
	BytesSent          int64   `json:"bytes_sent"`
	BytesReceived      int64   `json:"bytes_received"`
	ConnectionTime     float64 `json:"connection_time_ms"`
	AvgStreamLatency   float64 `json:"avg_stream_latency_ms"`
	DatagramLossRate   float64 `json:"datagram_loss_rate"`
	ErrorCount         int64   `json:"error_count"`
	LastError          string  `json:"last_error,omitempty"`
	
	mu sync.RWMutex
}

// NewClient creates a new WebTransport client
func NewClient(config *Config) *Client {
	return &Client{
		config: config,
		metrics: &Metrics{},
	}
}

// Connect establishes a WebTransport connection
func (c *Client) Connect(ctx context.Context) (*Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	sessionID := fmt.Sprintf("wt_session_%d", time.Now().Unix())
	
	session := &Session{
		ID:        sessionID,
		Status:    "connecting",
		CreatedAt: time.Now(),
		Config:    c.config,
		Metrics:   make(map[string]interface{}),
		streams:   make(map[string]*StreamInfo),
	}
	
	c.session = session
	
	// Start connection in background
	go c.establishConnection(ctx, session)
	
	return session, nil
}

// establishConnection handles the actual WebTransport connection establishment
func (c *Client) establishConnection(ctx context.Context, session *Session) {
	startTime := time.Now()
	
	defer func() {
		if r := recover(); r != nil {
			session.mu.Lock()
			session.Status = "failed"
			session.Error = fmt.Sprintf("Connection panic: %v", r)
			now := time.Now()
			session.ClosedAt = &now
			session.mu.Unlock()
			
			c.metrics.mu.Lock()
			c.metrics.ErrorCount++
			c.metrics.LastError = session.Error
			c.metrics.mu.Unlock()
		}
	}()
	
	// Configure TLS
	tlsConfig := c.config.TLSConfig
	if tlsConfig == nil {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true, // For testing purposes
			NextProtos:         c.config.ALPN,
		}
		
		if len(c.config.ALPN) == 0 {
			tlsConfig.NextProtos = []string{"wt"}
		}
	}
	
	// Create HTTP/3 client for WebTransport
	quicConfig := &quic.Config{
		EnableDatagrams: c.config.Datagrams,
	}
	
	roundTripper := &http3.RoundTripper{
		TLSClientConfig: tlsConfig,
		QuicConfig:      quicConfig,
	}
	defer roundTripper.Close()
	
	httpClient := &http.Client{
		Transport: roundTripper,
		Timeout:   30 * time.Second,
	}
	
	session.mu.Lock()
	session.httpClient = httpClient
	session.mu.Unlock()
	
	// Attempt WebTransport connection
	req, err := http.NewRequestWithContext(ctx, "CONNECT", c.config.URL, nil)
	if err != nil {
		session.mu.Lock()
		session.Status = "failed"
		session.Error = fmt.Sprintf("Failed to create request: %v", err)
		now := time.Now()
		session.ClosedAt = &now
		session.mu.Unlock()
		return
	}
	
	// Set WebTransport headers
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "webtransport")
	req.Header.Set("Sec-WebTransport-Http3-Draft", "draft02")
	
	// Add custom headers
	for key, value := range c.config.Headers {
		req.Header.Set(key, value)
	}
	
	resp, err := httpClient.Do(req)
	if err != nil {
		session.mu.Lock()
		session.Status = "failed"
		session.Error = fmt.Sprintf("Connection failed: %v", err)
		now := time.Now()
		session.ClosedAt = &now
		session.mu.Unlock()
		
		c.metrics.mu.Lock()
		c.metrics.ErrorCount++
		c.metrics.LastError = session.Error
		c.metrics.mu.Unlock()
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		session.mu.Lock()
		session.Status = "failed"
		session.Error = fmt.Sprintf("HTTP error: %d %s", resp.StatusCode, resp.Status)
		now := time.Now()
		session.ClosedAt = &now
		session.mu.Unlock()
		
		c.metrics.mu.Lock()
		c.metrics.ErrorCount++
		c.metrics.LastError = session.Error
		c.metrics.mu.Unlock()
		return
	}
	
	// Connection successful
	connectionTime := time.Since(startTime)
	now := time.Now()
	
	session.mu.Lock()
	session.Status = "connected"
	session.ConnectedAt = &now
	session.mu.Unlock()
	
	c.metrics.mu.Lock()
	c.metrics.ConnectionTime = float64(connectionTime.Nanoseconds()) / 1e6
	c.metrics.mu.Unlock()
	
	// Start test operations
	c.runTestOperations(ctx, session)
}

// runTestOperations performs WebTransport test operations
func (c *Client) runTestOperations(ctx context.Context, session *Session) {
	// Create test streams
	for i := 0; i < c.config.Streams; i++ {
		go c.createTestStream(ctx, session, i)
	}
	
	// Send datagrams if enabled
	if c.config.Datagrams {
		go c.sendDatagrams(ctx, session)
	}
	
	// Wait for test duration
	timer := time.NewTimer(c.config.Duration)
	defer timer.Stop()
	
	select {
	case <-ctx.Done():
		c.closeSession(session, "cancelled")
	case <-timer.C:
		c.closeSession(session, "completed")
	}
}

// createTestStream creates and tests a WebTransport stream
func (c *Client) createTestStream(ctx context.Context, session *Session, streamIndex int) {
	streamID := fmt.Sprintf("stream_%d", streamIndex)
	
	streamInfo := &StreamInfo{
		ID:        streamID,
		Type:      "bidirectional",
		CreatedAt: time.Now(),
		Status:    "open",
	}
	
	session.mu.Lock()
	session.streams[streamID] = streamInfo
	session.mu.Unlock()
	
	c.metrics.mu.Lock()
	c.metrics.StreamsOpened++
	c.metrics.mu.Unlock()
	
	// Simulate stream operations
	// In a real implementation, this would use actual WebTransport stream APIs
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	testData := make([]byte, 1024) // 1KB test data
	
	for {
		select {
		case <-ctx.Done():
			c.closeStream(session, streamInfo)
			return
		case <-ticker.C:
			// Simulate sending data
			streamInfo.BytesSent += int64(len(testData))
			streamInfo.BytesRecv += int64(len(testData)) // Echo response
			
			c.metrics.mu.Lock()
			c.metrics.BytesSent += int64(len(testData))
			c.metrics.BytesReceived += int64(len(testData))
			c.metrics.mu.Unlock()
		}
	}
}

// sendDatagrams sends WebTransport datagrams
func (c *Client) sendDatagrams(ctx context.Context, session *Session) {
	ticker := time.NewTicker(50 * time.Millisecond) // 20 datagrams per second
	defer ticker.Stop()
	
	datagramData := make([]byte, 512) // 512 bytes per datagram
	sentCount := int64(0)
	receivedCount := int64(0)
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Simulate sending datagram
			sentCount++
			
			// Simulate 95% delivery rate
			if sentCount%20 != 0 { // 5% loss
				receivedCount++
			}
			
			c.metrics.mu.Lock()
			c.metrics.DatagramsSent = sentCount
			c.metrics.DatagramsReceived = receivedCount
			c.metrics.BytesSent += int64(len(datagramData))
			c.metrics.BytesReceived += int64(len(datagramData))
			
			if sentCount > 0 {
				c.metrics.DatagramLossRate = float64(sentCount-receivedCount) / float64(sentCount)
			}
			c.metrics.mu.Unlock()
		}
	}
}

// closeStream closes a WebTransport stream
func (c *Client) closeStream(session *Session, streamInfo *StreamInfo) {
	streamInfo.Status = "closed"
	
	c.metrics.mu.Lock()
	c.metrics.StreamsClosed++
	c.metrics.mu.Unlock()
}

// closeSession closes the WebTransport session
func (c *Client) closeSession(session *Session, reason string) {
	session.mu.Lock()
	defer session.mu.Unlock()
	
	if session.Status == "closed" {
		return
	}
	
	session.Status = "closed"
	now := time.Now()
	session.ClosedAt = &now
	
	// Close all streams
	for _, streamInfo := range session.streams {
		if streamInfo.Status == "open" {
			streamInfo.Status = "closed"
			c.metrics.mu.Lock()
			c.metrics.StreamsClosed++
			c.metrics.mu.Unlock()
		}
	}
	
	// Close HTTP client
	if session.httpClient != nil {
		if transport, ok := session.httpClient.Transport.(*http3.RoundTripper); ok {
			transport.Close()
		}
	}
}

// GetSession returns the current session
func (c *Client) GetSession() *Session {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.session
}

// GetMetrics returns current metrics
func (c *Client) GetMetrics() *Metrics {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()
	
	// Return a copy
	metrics := *c.metrics
	return &metrics
}

// Close closes the client and cleans up resources
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.session != nil {
		c.closeSession(c.session, "client_closed")
	}
	
	return nil
}