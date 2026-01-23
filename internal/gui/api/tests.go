package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TestHandlers contains all test-related HTTP handlers
type TestHandlers struct {
	api *APIServer
}

// NewTestHandlers creates new test handlers
func NewTestHandlers(api *APIServer) *TestHandlers {
	return &TestHandlers{api: api}
}

// HandleTests handles /api/tests endpoint
func (h *TestHandlers) HandleTests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.handleListTests(w, r)
	case "POST":
		h.handleCreateTest(w, r)
	default:
		h.api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleTestByID handles /api/tests/{id} endpoint
func (h *TestHandlers) HandleTestByID(w http.ResponseWriter, r *http.Request) {
	testID := strings.TrimPrefix(r.URL.Path, "/api/tests/")
	if testID == "" {
		h.api.sendError(w, "Test ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		h.handleGetTest(w, r, testID)
	case "DELETE":
		h.handleStopTest(w, r, testID)
	default:
		h.api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListTests lists all tests
func (h *TestHandlers) handleListTests(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Set defaults
	limit := 50
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get all tests
	allTests := h.api.testManager.GetAllTests()

	// Filter by status if specified
	var filteredTests []*TestSession
	if status != "" {
		for _, test := range allTests {
			if test.Status == status {
				filteredTests = append(filteredTests, test)
			}
		}
	} else {
		filteredTests = allTests
	}

	// Apply pagination
	total := len(filteredTests)
	start := offset
	end := offset + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedTests := filteredTests[start:end]

	response := map[string]interface{}{
		"tests":  paginatedTests,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	h.api.sendSuccess(w, response)
}

// handleCreateTest creates a new test
func (h *TestHandlers) handleCreateTest(w http.ResponseWriter, r *http.Request) {
	var rawConfig map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawConfig); err != nil {
		h.api.sendError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	config, err := h.parseTestConfig(rawConfig)
	if err != nil {
		h.api.sendError(w, "Invalid configuration: "+err.Error(), http.StatusBadRequest)
		return
	}

	session := h.api.testManager.StartTest(*config)
	h.api.sendSuccess(w, session)
}

// parseTestConfig converts raw JSON map to TestConfig
func (h *TestHandlers) parseTestConfig(raw map[string]interface{}) (*TestConfig, error) {
	config := &TestConfig{}

	// Mode (required)
	if mode, ok := raw["mode"].(string); ok {
		config.Mode = mode
	} else {
		config.Mode = "test" // default
	}

	// Duration
	if durationStr, ok := raw["duration"].(string); ok {
		if duration, err := time.ParseDuration(durationStr); err == nil {
			config.Duration = duration
		}
	}
	if config.Duration == 0 {
		config.Duration = 60 * time.Second // default
	}

	// Connections
	if connections, ok := raw["connections"].(float64); ok {
		config.Connections = int(connections)
	} else {
		config.Connections = 2 // default
	}

	// Streams
	if streams, ok := raw["streams"].(float64); ok {
		config.Streams = int(streams)
	} else {
		config.Streams = 4 // default
	}

	// Address
	if addr, ok := raw["addr"].(string); ok {
		config.Addr = addr
	} else {
		config.Addr = "localhost:9000" // default
	}

	// Packet size
	if packetSize, ok := raw["packet_size"].(float64); ok {
		config.PacketSize = int(packetSize)
	} else {
		config.PacketSize = 1200 // default
	}

	// Rate
	if rate, ok := raw["rate"].(float64); ok {
		config.Rate = int(rate)
	} else {
		config.Rate = 100 // default
	}

	// Congestion control
	if cc, ok := raw["congestion_control"].(string); ok {
		config.CongestionControl = cc
	}

	// Prometheus
	if prometheus, ok := raw["prometheus"].(bool); ok {
		config.Prometheus = prometheus
	}

	// FEC
	if fecEnabled, ok := raw["fec_enabled"].(bool); ok {
		config.FECEnabled = fecEnabled
	}

	if fecRedundancy, ok := raw["fec_redundancy"].(float64); ok {
		config.FECRedundancy = fecRedundancy
	} else {
		config.FECRedundancy = 0.10 // default 10%
	}

	// PQC
	if pqcEnabled, ok := raw["pqc_enabled"].(bool); ok {
		config.PQCEnabled = pqcEnabled
	}

	// Network emulation
	if emulateLatency, ok := raw["emulate_latency"].(string); ok {
		config.EmulateLatency = emulateLatency
	}

	if emulateLoss, ok := raw["emulate_loss"].(float64); ok {
		config.EmulateLoss = emulateLoss
	}

	if emulateDup, ok := raw["emulate_dup"].(float64); ok {
		config.EmulateDup = emulateDup
	}

	return config, nil
}

// handleGetTest gets a specific test
func (h *TestHandlers) handleGetTest(w http.ResponseWriter, r *http.Request, testID string) {
	session := h.api.testManager.GetTest(testID)
	if session == nil {
		h.api.sendError(w, "Test not found", http.StatusNotFound)
		return
	}

	h.api.sendSuccess(w, session)
}

// handleStopTest stops a test
func (h *TestHandlers) handleStopTest(w http.ResponseWriter, r *http.Request, testID string) {
	if err := h.api.testManager.StopTest(testID); err != nil {
		h.api.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]string{
		"message": "Test stopped successfully",
	}
	h.api.sendSuccess(w, response)
}