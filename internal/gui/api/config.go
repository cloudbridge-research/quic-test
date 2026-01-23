package api

import "net/http"

// ConfigHandlers contains all configuration-related HTTP handlers
type ConfigHandlers struct {
	api *APIServer
}

// NewConfigHandlers creates new config handlers
func NewConfigHandlers(api *APIServer) *ConfigHandlers {
	return &ConfigHandlers{api: api}
}

// HandleConfigPresets returns available configuration presets
func (h *ConfigHandlers) HandleConfigPresets(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	presets := h.getNetworkPresets()
	h.api.sendSuccess(w, presets)
}

// HandleConfigProfiles returns available test profiles
func (h *ConfigHandlers) HandleConfigProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	profiles := h.getTestProfiles()
	h.api.sendSuccess(w, profiles)
}

// getNetworkPresets returns predefined network presets
func (h *ConfigHandlers) getNetworkPresets() []NetworkPreset {
	return []NetworkPreset{
		{
			Name:        "fiber",
			Description: "Fiber optic connection (low latency, high bandwidth)",
			Latency:     "5ms",
			Bandwidth:   "1000Mbps",
			Loss:        "0.01%",
		},
		{
			Name:        "mobile",
			Description: "4G/LTE mobile connection",
			Latency:     "50ms",
			Bandwidth:   "50Mbps",
			Loss:        "1%",
		},
		{
			Name:        "satellite",
			Description: "Satellite connection (high latency)",
			Latency:     "600ms",
			Bandwidth:   "25Mbps",
			Loss:        "2%",
		},
		{
			Name:        "wifi",
			Description: "WiFi connection",
			Latency:     "20ms",
			Bandwidth:   "100Mbps",
			Loss:        "0.5%",
		},
		{
			Name:        "ethernet",
			Description: "Ethernet LAN connection",
			Latency:     "1ms",
			Bandwidth:   "1000Mbps",
			Loss:        "0.001%",
		},
	}
}

// getTestProfiles returns predefined test profiles
func (h *ConfigHandlers) getTestProfiles() []TestProfile {
	return []TestProfile{
		{
			Name:        "quick",
			Description: "Quick performance test (30 seconds)",
			Duration:    "30s",
			Connections: 1,
			Streams:     2,
			Rate:        100,
		},
		{
			Name:        "standard",
			Description: "Standard performance test (2 minutes)",
			Duration:    "120s",
			Connections: 2,
			Streams:     4,
			Rate:        200,
		},
		{
			Name:        "intensive",
			Description: "Intensive load test (5 minutes)",
			Duration:    "300s",
			Connections: 4,
			Streams:     8,
			Rate:        500,
		},
		{
			Name:        "endurance",
			Description: "Long-running endurance test (30 minutes)",
			Duration:    "1800s",
			Connections: 2,
			Streams:     4,
			Rate:        100,
		},
		{
			Name:        "stress",
			Description: "High-load stress test (10 minutes)",
			Duration:    "600s",
			Connections: 10,
			Streams:     20,
			Rate:        1000,
		},
	}
}