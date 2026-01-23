package api

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// MetricsHandlers contains all metrics-related HTTP handlers
type MetricsHandlers struct {
	api *APIServer
}

// NewMetricsHandlers creates new metrics handlers
func NewMetricsHandlers(api *APIServer) *MetricsHandlers {
	return &MetricsHandlers{api: api}
}

// HandleCurrentMetrics gets current aggregated metrics
func (h *MetricsHandlers) HandleCurrentMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all active tests
	allTests := h.api.testManager.GetAllTests()
	activeTests := make([]*TestSession, 0)
	
	for _, test := range allTests {
		if test.Status == "running" {
			activeTests = append(activeTests, test)
		}
	}

	// Aggregate metrics from all active tests
	aggregatedMetrics := h.aggregateMetrics(activeTests)

	// Add system metrics
	systemMetrics := h.getSystemMetrics()
	
	response := map[string]interface{}{
		"test_metrics":   aggregatedMetrics,
		"system_metrics": systemMetrics,
		"active_tests":   len(activeTests),
		"timestamp":      time.Now(),
	}

	h.api.sendSuccess(w, response)
}

// HandleHistoricalMetrics gets historical metrics
func (h *MetricsHandlers) HandleHistoricalMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	testID := r.URL.Query().Get("test_id")
	startTimeStr := r.URL.Query().Get("start_time")
	endTimeStr := r.URL.Query().Get("end_time")
	interval := r.URL.Query().Get("interval")

	// Set defaults
	if interval == "" {
		interval = "1m"
	}

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			h.api.sendError(w, "Invalid start_time format", http.StatusBadRequest)
			return
		}
	} else {
		startTime = time.Now().Add(-1 * time.Hour) // Default: last hour
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			h.api.sendError(w, "Invalid end_time format", http.StatusBadRequest)
			return
		}
	} else {
		endTime = time.Now()
	}

	// Get historical data (placeholder implementation)
	historicalData := h.getHistoricalMetrics(testID, startTime, endTime, interval)

	response := map[string]interface{}{
		"test_id":    testID,
		"start_time": startTime,
		"end_time":   endTime,
		"interval":   interval,
		"data":       historicalData,
	}

	h.api.sendSuccess(w, response)
}

// HandlePrometheusMetrics returns metrics in Prometheus format
func (h *MetricsHandlers) HandlePrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all tests
	allTests := h.api.testManager.GetAllTests()
	
	// Generate Prometheus metrics
	prometheusMetrics := h.generatePrometheusMetrics(allTests)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(prometheusMetrics))
}

// aggregateMetrics aggregates metrics from multiple tests
func (h *MetricsHandlers) aggregateMetrics(tests []*TestSession) map[string]interface{} {
	if len(tests) == 0 {
		return map[string]interface{}{
			"total_throughput": 0.0,
			"avg_latency":      0.0,
			"total_packets":    0,
			"packet_loss":      0.0,
		}
	}

	var totalThroughput, totalLatency, totalPacketLoss float64
	var totalPackets int64
	validLatencyCount := 0

	for _, test := range tests {
		if test.Metrics != nil {
			if throughput, ok := test.Metrics["throughput"].(float64); ok {
				totalThroughput += throughput
			}
			if latency, ok := test.Metrics["latency"].(float64); ok {
				totalLatency += latency
				validLatencyCount++
			}
			if packets, ok := test.Metrics["packets"].(int64); ok {
				totalPackets += packets
			}
			if loss, ok := test.Metrics["packet_loss"].(float64); ok {
				totalPacketLoss += loss
			}
		}
	}

	avgLatency := 0.0
	if validLatencyCount > 0 {
		avgLatency = totalLatency / float64(validLatencyCount)
	}

	avgPacketLoss := 0.0
	if len(tests) > 0 {
		avgPacketLoss = totalPacketLoss / float64(len(tests))
	}

	return map[string]interface{}{
		"total_throughput": totalThroughput,
		"avg_latency":      avgLatency,
		"total_packets":    totalPackets,
		"packet_loss":      avgPacketLoss,
		"active_tests":     len(tests),
	}
}

// getSystemMetrics returns current system metrics
func (h *MetricsHandlers) getSystemMetrics() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"memory": map[string]interface{}{
			"alloc":      m.Alloc,
			"total_alloc": m.TotalAlloc,
			"sys":        m.Sys,
			"num_gc":     m.NumGC,
		},
		"goroutines": runtime.NumGoroutine(),
		"cpu_cores":  runtime.NumCPU(),
	}
}

// getHistoricalMetrics returns historical metrics (placeholder)
func (h *MetricsHandlers) getHistoricalMetrics(testID string, startTime, endTime time.Time, interval string) []map[string]interface{} {
	// This is a placeholder implementation
	// In a real system, this would query a time-series database
	
	data := make([]map[string]interface{}, 0)
	
	// Generate sample data points
	current := startTime
	for current.Before(endTime) {
		dataPoint := map[string]interface{}{
			"timestamp":  current,
			"throughput": 100.0 + float64(current.Unix()%100),
			"latency":    50.0 + float64(current.Unix()%50),
			"packets":    1000 + current.Unix()%1000,
		}
		data = append(data, dataPoint)
		
		// Increment by interval
		switch interval {
		case "1s":
			current = current.Add(time.Second)
		case "5s":
			current = current.Add(5 * time.Second)
		case "1m":
			current = current.Add(time.Minute)
		case "5m":
			current = current.Add(5 * time.Minute)
		default:
			current = current.Add(time.Minute)
		}
	}
	
	return data
}

// generatePrometheusMetrics generates metrics in Prometheus format
func (h *MetricsHandlers) generatePrometheusMetrics(tests []*TestSession) string {
	var builder strings.Builder
	
	// Write header
	builder.WriteString("# HELP quic_test_active_tests Number of active tests\n")
	builder.WriteString("# TYPE quic_test_active_tests gauge\n")
	
	activeCount := 0
	for _, test := range tests {
		if test.Status == "running" {
			activeCount++
		}
	}
	builder.WriteString(fmt.Sprintf("quic_test_active_tests %d\n", activeCount))
	
	// Total tests
	builder.WriteString("# HELP quic_test_total_tests Total number of tests\n")
	builder.WriteString("# TYPE quic_test_total_tests counter\n")
	builder.WriteString(fmt.Sprintf("quic_test_total_tests %d\n", len(tests)))
	
	// Per-test metrics
	builder.WriteString("# HELP quic_test_throughput Test throughput in Mbps\n")
	builder.WriteString("# TYPE quic_test_throughput gauge\n")
	
	builder.WriteString("# HELP quic_test_latency Test latency in milliseconds\n")
	builder.WriteString("# TYPE quic_test_latency gauge\n")
	
	builder.WriteString("# HELP quic_test_packet_loss Test packet loss rate\n")
	builder.WriteString("# TYPE quic_test_packet_loss gauge\n")
	
	for _, test := range tests {
		labels := fmt.Sprintf(`{test_id="%s",status="%s"}`, test.ID, test.Status)
		
		if test.Metrics != nil {
			if throughput, ok := test.Metrics["throughput"].(float64); ok {
				builder.WriteString(fmt.Sprintf("quic_test_throughput%s %f\n", labels, throughput))
			}
			if latency, ok := test.Metrics["latency"].(float64); ok {
				builder.WriteString(fmt.Sprintf("quic_test_latency%s %f\n", labels, latency))
			}
			if loss, ok := test.Metrics["packet_loss"].(float64); ok {
				builder.WriteString(fmt.Sprintf("quic_test_packet_loss%s %f\n", labels, loss))
			}
		}
	}
	
	// System metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	builder.WriteString("# HELP quic_test_memory_alloc Allocated memory in bytes\n")
	builder.WriteString("# TYPE quic_test_memory_alloc gauge\n")
	builder.WriteString(fmt.Sprintf("quic_test_memory_alloc %d\n", m.Alloc))
	
	builder.WriteString("# HELP quic_test_goroutines Number of goroutines\n")
	builder.WriteString("# TYPE quic_test_goroutines gauge\n")
	builder.WriteString(fmt.Sprintf("quic_test_goroutines %d\n", runtime.NumGoroutine()))
	
	return builder.String()
}