package gui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"quic-test/internal"
)

// APIServer handles REST API requests
type APIServer struct {
	testManager *TestManager
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewAPIServer creates a new API server
func NewAPIServer() *APIServer {
	return &APIServer{
		testManager: NewTestManager(),
	}
}

// RegisterRoutes registers API routes
func (api *APIServer) RegisterRoutes(mux *http.ServeMux) {
	// Test management
	mux.HandleFunc("/api/tests", api.handleTests)
	mux.HandleFunc("/api/tests/", api.handleTestByID)
	
	// Metrics
	mux.HandleFunc("/api/metrics/current", api.handleCurrentMetrics)
	mux.HandleFunc("/api/metrics/history", api.handleHistoricalMetrics)
	mux.HandleFunc("/api/metrics/prometheus", api.handlePrometheusMetrics)
	
	// Configuration
	mux.HandleFunc("/api/config/presets", api.handleConfigPresets)
	mux.HandleFunc("/api/config/profiles", api.handleConfigProfiles)
	
	// System
	mux.HandleFunc("/api/system/status", api.handleSystemStatus)
	mux.HandleFunc("/api/system/health", api.handleHealthCheck)
	
	// WebSocket endpoint (placeholder)
	mux.HandleFunc("/api/ws/metrics", api.handleWebSocketMetrics)
}

// handleTests handles /api/tests endpoint
func (api *APIServer) handleTests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		api.handleListTests(w, r)
	case "POST":
		api.handleCreateTest(w, r)
	default:
		api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTestByID handles /api/tests/{id} endpoint
func (api *APIServer) handleTestByID(w http.ResponseWriter, r *http.Request) {
	testID := strings.TrimPrefix(r.URL.Path, "/api/tests/")
	if testID == "" {
		api.sendError(w, "Test ID required", http.StatusBadRequest)
		return
	}
	
	switch r.Method {
	case "GET":
		api.handleGetTest(w, r, testID)
	case "DELETE":
		api.handleStopTest(w, r, testID)
	default:
		api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListTests lists all tests
func (api *APIServer) handleListTests(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	// Get all tests
	allTests := api.testManager.GetAllTests()
	
	// Filter by status if specified
	var filteredTests []*TestSession
	for _, test := range allTests {
		if status == "" || test.Status == status {
			filteredTests = append(filteredTests, test)
		}
	}
	
	// Apply pagination
	total := len(filteredTests)
	start := offset
	end := offset + limit
	
	if start >= total {
		filteredTests = []*TestSession{}
	} else {
		if end > total {
			end = total
		}
		filteredTests = filteredTests[start:end]
	}
	
	response := struct {
		Tests  []*TestSession `json:"tests"`
		Total  int            `json:"total"`
		Limit  int            `json:"limit"`
		Offset int            `json:"offset"`
	}{
		Tests:  filteredTests,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	
	api.sendSuccess(w, response)
}

// handleCreateTest creates a new test
func (api *APIServer) handleCreateTest(w http.ResponseWriter, r *http.Request) {
	var rawConfig map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawConfig); err != nil {
		api.sendError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Convert raw config to TestConfig
	config, err := api.parseTestConfig(rawConfig)
	if err != nil {
		api.sendError(w, "Invalid configuration: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		api.sendError(w, "Invalid configuration: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Start test
	session := api.testManager.StartTest(*config)
	api.sendSuccess(w, session)
}

// parseTestConfig converts raw JSON map to TestConfig
func (api *APIServer) parseTestConfig(raw map[string]interface{}) (*internal.TestConfig, error) {
	config := &internal.TestConfig{}
	
	// Parse basic fields
	if v, ok := raw["mode"].(string); ok && v != "" {
		config.Mode = v
	} else {
		config.Mode = "test" // default mode
	}
	if v, ok := raw["addr"].(string); ok && v != "" {
		config.Addr = v
	} else {
		config.Addr = "localhost:9000" // default address
	}
	if v, ok := raw["connections"].(float64); ok {
		config.Connections = int(v)
	} else if v, ok := raw["connections"].(string); ok {
		if v == "" {
			config.Connections = 2 // default value
		} else if parsed, err := strconv.Atoi(v); err == nil {
			config.Connections = parsed
		} else {
			return nil, fmt.Errorf("invalid connections value: %s", v)
		}
	} else {
		config.Connections = 2 // default value
	}
	
	if v, ok := raw["streams"].(float64); ok {
		config.Streams = int(v)
	} else if v, ok := raw["streams"].(string); ok {
		if v == "" {
			config.Streams = 4 // default value
		} else if parsed, err := strconv.Atoi(v); err == nil {
			config.Streams = parsed
		} else {
			return nil, fmt.Errorf("invalid streams value: %s", v)
		}
	} else {
		config.Streams = 4 // default value
	}
	if v, ok := raw["packet_size"].(float64); ok {
		config.PacketSize = int(v)
	} else if v, ok := raw["packet_size"].(string); ok {
		if v == "" {
			config.PacketSize = 1200 // default value
		} else if parsed, err := strconv.Atoi(v); err == nil {
			config.PacketSize = parsed
		} else {
			return nil, fmt.Errorf("invalid packet_size value: %s", v)
		}
	} else {
		config.PacketSize = 1200 // default value
	}
	
	if v, ok := raw["rate"].(float64); ok {
		config.Rate = int(v)
	} else if v, ok := raw["rate"].(string); ok {
		if v == "" {
			config.Rate = 100 // default value
		} else if parsed, err := strconv.Atoi(v); err == nil {
			config.Rate = parsed
		} else {
			return nil, fmt.Errorf("invalid rate value: %s", v)
		}
	} else {
		config.Rate = 100 // default value
	}
	if v, ok := raw["prometheus"].(bool); ok {
		config.Prometheus = v
	}
	if v, ok := raw["fec_enabled"].(bool); ok {
		config.FECEnabled = v
	}
	if v, ok := raw["fec_redundancy"].(float64); ok {
		config.FECRedundancy = v
	}
	if v, ok := raw["pqc_enabled"].(bool); ok {
		config.PQCEnabled = v
	}
	if v, ok := raw["congestion_control"].(string); ok {
		config.CongestionControl = v
	}
	
	// Parse duration fields
	if v, ok := raw["duration"].(string); ok {
		if d, err := time.ParseDuration(v); err == nil {
			config.Duration = d
		} else {
			return nil, fmt.Errorf("invalid duration format: %s", v)
		}
	} else if v, ok := raw["duration"].(float64); ok {
		// Handle case where duration comes as nanoseconds (number)
		config.Duration = time.Duration(int64(v))
	} else {
		config.Duration = 60 * time.Second // default 60 seconds
	}
	
	if v, ok := raw["emulate_latency"].(string); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.EmulateLatency = d
		} else {
			return nil, fmt.Errorf("invalid emulate_latency format: %s", v)
		}
	}
	
	// Parse float fields
	if v, ok := raw["emulate_loss"].(float64); ok {
		config.EmulateLoss = v
	}
	if v, ok := raw["emulate_dup"].(float64); ok {
		config.EmulateDup = v
	}
	
	return config, nil
}

// handleGetTest gets a specific test
func (api *APIServer) handleGetTest(w http.ResponseWriter, r *http.Request, testID string) {
	session := api.testManager.GetTest(testID)
	if session == nil {
		api.sendError(w, "Test not found", http.StatusNotFound)
		return
	}
	
	api.sendSuccess(w, session)
}

// handleStopTest stops a test
func (api *APIServer) handleStopTest(w http.ResponseWriter, r *http.Request, testID string) {
	if err := api.testManager.StopTest(testID); err != nil {
		api.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	api.sendSuccess(w, map[string]string{
		"message": "Test stopped successfully",
	})
}

// handleCurrentMetrics gets current aggregated metrics
func (api *APIServer) handleCurrentMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Aggregate metrics from all active tests
	activeTests := api.testManager.GetAllTests()
	
	aggregatedMetrics := map[string]interface{}{
		"active_tests":     0,
		"total_connections": 0,
		"avg_latency_ms":   0.0,
		"total_throughput_mbps": 0.0,
		"avg_packet_loss":  0.0,
		"total_errors":     0,
	}
	
	activeCount := 0
	latencySum := 0.0
	throughputSum := 0.0
	lossSum := 0.0
	
	for _, test := range activeTests {
		if test.Status == "running" {
			activeCount++
			metrics := test.GetMetrics()
			
			if connections, ok := metrics["connections"].(int); ok {
				aggregatedMetrics["total_connections"] = aggregatedMetrics["total_connections"].(int) + connections
			}
			
			if latency, ok := metrics["latency_ms"].(float64); ok {
				latencySum += latency
			}
			
			if throughput, ok := metrics["throughput_mbps"].(float64); ok {
				throughputSum += throughput
			}
			
			if loss, ok := metrics["packet_loss"].(float64); ok {
				lossSum += loss
			}
			
			if errors, ok := metrics["errors"].(int); ok {
				aggregatedMetrics["total_errors"] = aggregatedMetrics["total_errors"].(int) + errors
			}
		}
	}
	
	aggregatedMetrics["active_tests"] = activeCount
	
	if activeCount > 0 {
		aggregatedMetrics["avg_latency_ms"] = latencySum / float64(activeCount)
		aggregatedMetrics["avg_packet_loss"] = lossSum / float64(activeCount)
	}
	
	aggregatedMetrics["total_throughput_mbps"] = throughputSum
	
	api.sendSuccess(w, aggregatedMetrics)
}

// handleHistoricalMetrics gets historical metrics
func (api *APIServer) handleHistoricalMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if r.Method != "GET" {
		api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	testID := r.URL.Query().Get("test_id")
	_ = r.URL.Query().Get("start_time") // startTimeStr - unused for now
	_ = r.URL.Query().Get("end_time")   // endTimeStr - unused for now
	interval := r.URL.Query().Get("interval")
	
	if interval == "" {
		interval = "5s"
	}
	
	// For now, return placeholder data
	// In a real implementation, this would query a time-series database
	historicalData := map[string]interface{}{
		"test_id":  testID,
		"interval": interval,
		"metrics": []map[string]interface{}{
			{
				"timestamp":      time.Now().Add(-60 * time.Second),
				"latency_ms":     45.2,
				"throughput_mbps": 125.8,
				"packet_loss":    0.01,
			},
			{
				"timestamp":      time.Now().Add(-30 * time.Second),
				"latency_ms":     47.1,
				"throughput_mbps": 128.3,
				"packet_loss":    0.008,
			},
			{
				"timestamp":      time.Now(),
				"latency_ms":     44.8,
				"throughput_mbps": 131.2,
				"packet_loss":    0.012,
			},
		},
	}
	
	api.sendSuccess(w, historicalData)
}

// handlePrometheusMetrics returns metrics in Prometheus format
func (api *APIServer) handlePrometheusMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	
	// Generate Prometheus metrics
	activeTests := api.testManager.GetAllTests()
	
	metrics := []string{
		"# HELP quic_test_active_tests Number of active tests",
		"# TYPE quic_test_active_tests gauge",
	}
	
	activeCount := 0
	for _, test := range activeTests {
		if test.Status == "running" {
			activeCount++
		}
	}
	
	metrics = append(metrics, fmt.Sprintf("quic_test_active_tests %d", activeCount))
	
	// Add per-test metrics
	for _, test := range activeTests {
		if test.Status == "running" {
			testMetrics := test.GetMetrics()
			
			if latency, ok := testMetrics["latency_ms"].(float64); ok {
				metrics = append(metrics, fmt.Sprintf("quic_test_latency_ms{test_id=\"%s\"} %.2f", test.ID, latency))
			}
			
			if throughput, ok := testMetrics["throughput_mbps"].(float64); ok {
				metrics = append(metrics, fmt.Sprintf("quic_test_throughput_mbps{test_id=\"%s\"} %.2f", test.ID, throughput))
			}
			
			if loss, ok := testMetrics["packet_loss"].(float64); ok {
				metrics = append(metrics, fmt.Sprintf("quic_test_packet_loss{test_id=\"%s\"} %.4f", test.ID, loss))
			}
		}
	}
	
	w.Write([]byte(strings.Join(metrics, "\n") + "\n"))
}

// handleConfigPresets returns available configuration presets
func (api *APIServer) handleConfigPresets(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	presets := getNetworkPresets()
	api.sendSuccess(w, presets)
}

// handleConfigProfiles returns available test profiles
func (api *APIServer) handleConfigProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	profiles := getTestProfiles()
	api.sendSuccess(w, profiles)
}

// handleSystemStatus returns system status information
func (api *APIServer) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	status := map[string]interface{}{
		"uptime":       time.Since(startTime).String(),
		"active_tests": api.testManager.GetActiveTestCount(),
		"total_tests":  api.testManager.GetTotalTestCount(),
		"version":      "1.0.0",
		"build_time":   "2024-01-01T00:00:00Z",
	}
	
	api.sendSuccess(w, status)
}

// handleHealthCheck returns health status
func (api *APIServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"checks": map[string]string{
			"api_server":   "ok",
			"test_manager": "ok",
		},
	}
	
	api.sendSuccess(w, health)
}

// handleWebSocketMetrics handles WebSocket connections for real-time metrics
func (api *APIServer) handleWebSocketMetrics(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder for WebSocket implementation
	// In a real implementation, this would upgrade the connection to WebSocket
	// and stream real-time metrics updates
	
	api.sendError(w, "WebSocket not implemented yet", http.StatusNotImplemented)
}

// Helper methods

// sendSuccess sends a successful API response
func (api *APIServer) sendSuccess(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// sendError sends an error API response
func (api *APIServer) sendError(w http.ResponseWriter, message string, statusCode int) {
	response := APIResponse{
		Success:   false,
		Error:     message,
		Timestamp: time.Now(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}