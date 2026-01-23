// Package api provides REST API handlers for the QUIC test suite GUI
//
// This package contains modular HTTP handlers for managing tests, metrics,
// configuration, and system status through a REST API interface.
package api

import "net/http"

// RegisterRoutes registers all API routes with the provided mux
func (api *APIServer) RegisterRoutes(mux *http.ServeMux) {
	// Create handler instances
	testHandlers := NewTestHandlers(api)
	metricsHandlers := NewMetricsHandlers(api)
	configHandlers := NewConfigHandlers(api)
	systemHandlers := NewSystemHandlers(api)

	// Test management routes
	mux.HandleFunc("/api/tests", testHandlers.HandleTests)
	mux.HandleFunc("/api/tests/", testHandlers.HandleTestByID)
	
	// Metrics routes
	mux.HandleFunc("/api/metrics/current", metricsHandlers.HandleCurrentMetrics)
	mux.HandleFunc("/api/metrics/history", metricsHandlers.HandleHistoricalMetrics)
	mux.HandleFunc("/api/metrics/prometheus", metricsHandlers.HandlePrometheusMetrics)
	
	// Configuration routes
	mux.HandleFunc("/api/config/presets", configHandlers.HandleConfigPresets)
	mux.HandleFunc("/api/config/profiles", configHandlers.HandleConfigProfiles)
	
	// System routes
	mux.HandleFunc("/api/system/status", systemHandlers.HandleSystemStatus)
	mux.HandleFunc("/api/health", systemHandlers.HandleHealthCheck)
	mux.HandleFunc("/api/ws/metrics", systemHandlers.HandleWebSocketMetrics)
}

// RegisterRoutesWithMiddleware registers routes with middleware applied
func (api *APIServer) RegisterRoutesWithMiddleware(mux *http.ServeMux) {
	// Create handler instances
	testHandlers := NewTestHandlers(api)
	metricsHandlers := NewMetricsHandlers(api)
	configHandlers := NewConfigHandlers(api)
	systemHandlers := NewSystemHandlers(api)

	// Apply middleware chain
	chain := func(h http.HandlerFunc) http.Handler {
		return api.enableCORS(
			api.logRequests(
				api.rateLimit(
					api.authenticate(h))))
	}

	// Test management routes
	mux.Handle("/api/tests", chain(testHandlers.HandleTests))
	mux.Handle("/api/tests/", chain(testHandlers.HandleTestByID))
	
	// Metrics routes
	mux.Handle("/api/metrics/current", chain(metricsHandlers.HandleCurrentMetrics))
	mux.Handle("/api/metrics/history", chain(metricsHandlers.HandleHistoricalMetrics))
	mux.Handle("/api/metrics/prometheus", chain(metricsHandlers.HandlePrometheusMetrics))
	
	// Configuration routes
	mux.Handle("/api/config/presets", chain(configHandlers.HandleConfigPresets))
	mux.Handle("/api/config/profiles", chain(configHandlers.HandleConfigProfiles))
	
	// System routes
	mux.Handle("/api/system/status", chain(systemHandlers.HandleSystemStatus))
	mux.Handle("/api/health", chain(systemHandlers.HandleHealthCheck))
	mux.Handle("/api/ws/metrics", chain(systemHandlers.HandleWebSocketMetrics))
}

// GetAPIInfo returns information about the API module
func GetAPIInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "QUIC Test Suite API",
		"version":     "1.0.0",
		"description": "REST API for QUIC test management and monitoring",
		"endpoints": map[string][]string{
			"tests": {
				"GET /api/tests",
				"POST /api/tests", 
				"GET /api/tests/{id}",
				"DELETE /api/tests/{id}",
			},
			"metrics": {
				"GET /api/metrics/current",
				"GET /api/metrics/history",
				"GET /api/metrics/prometheus",
			},
			"config": {
				"GET /api/config/presets",
				"GET /api/config/profiles",
			},
			"system": {
				"GET /api/system/status",
				"GET /api/health",
				"GET /api/ws/metrics",
			},
		},
		"features": []string{
			"Test lifecycle management",
			"Real-time metrics collection",
			"Historical metrics queries",
			"Prometheus metrics export",
			"Configuration presets",
			"System health monitoring",
			"CORS support",
			"Request logging",
			"Rate limiting",
			"Authentication ready",
		},
	}
}