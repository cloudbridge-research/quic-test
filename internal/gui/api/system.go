package api

import (
	"net/http"
	"runtime"
	"time"
)

// SystemHandlers contains all system-related HTTP handlers
type SystemHandlers struct {
	api       *APIServer
	startTime time.Time
}

// NewSystemHandlers creates new system handlers
func NewSystemHandlers(api *APIServer) *SystemHandlers {
	return &SystemHandlers{
		api:       api,
		startTime: time.Now(),
	}
}

// HandleSystemStatus returns system status information
func (h *SystemHandlers) HandleSystemStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := h.getSystemStatus()
	h.api.sendSuccess(w, status)
}

// HandleHealthCheck returns health status
func (h *SystemHandlers) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.api.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health := h.getHealthStatus()
	
	// Set appropriate status code based on health
	statusCode := http.StatusOK
	if health.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}
	
	w.WriteHeader(statusCode)
	h.api.sendSuccess(w, health)
}

// HandleWebSocketMetrics handles WebSocket connections for real-time metrics
func (h *SystemHandlers) HandleWebSocketMetrics(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder for WebSocket implementation
	// In a real implementation, this would upgrade the connection to WebSocket
	h.api.sendError(w, "WebSocket not implemented yet", http.StatusNotImplemented)
}

// getSystemStatus returns current system status
func (h *SystemHandlers) getSystemStatus() SystemStatus {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(h.startTime)
	
	return SystemStatus{
		Version:     "1.0.0", // This should come from build info
		Uptime:      uptime.String(),
		ActiveTests: h.api.testManager.GetActiveTestCount(),
		TotalTests:  h.api.testManager.GetTotalTestCount(),
		Memory: MemInfo{
			Used:      m.Alloc,
			Available: m.Sys - m.Alloc,
			Total:     m.Sys,
			Percent:   float64(m.Alloc) / float64(m.Sys) * 100,
		},
		CPU: CPUInfo{
			Cores:   runtime.NumCPU(),
			Usage:   h.getCPUUsage(),
			LoadAvg: h.getLoadAverage(),
		},
		Timestamp: time.Now(),
	}
}

// getHealthStatus returns current health status
func (h *SystemHandlers) getHealthStatus() HealthStatus {
	checks := make(map[string]string)
	
	// Check memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryPercent := float64(m.Alloc) / float64(m.Sys) * 100
	
	if memoryPercent < 80 {
		checks["memory"] = "healthy"
	} else if memoryPercent < 90 {
		checks["memory"] = "warning"
	} else {
		checks["memory"] = "critical"
	}
	
	// Check goroutine count
	goroutines := runtime.NumGoroutine()
	if goroutines < 1000 {
		checks["goroutines"] = "healthy"
	} else if goroutines < 5000 {
		checks["goroutines"] = "warning"
	} else {
		checks["goroutines"] = "critical"
	}
	
	// Check test manager
	activeTests := h.api.testManager.GetActiveTestCount()
	if activeTests < 100 {
		checks["test_manager"] = "healthy"
	} else if activeTests < 500 {
		checks["test_manager"] = "warning"
	} else {
		checks["test_manager"] = "critical"
	}
	
	// Determine overall status
	status := "healthy"
	for _, check := range checks {
		if check == "critical" {
			status = "critical"
			break
		} else if check == "warning" && status == "healthy" {
			status = "warning"
		}
	}
	
	return HealthStatus{
		Status:    status,
		Checks:    checks,
		Timestamp: time.Now(),
	}
}

// getCPUUsage returns CPU usage percentage (placeholder)
func (h *SystemHandlers) getCPUUsage() float64 {
	// This is a placeholder implementation
	// In a real system, this would calculate actual CPU usage
	return 25.0
}

// getLoadAverage returns system load average (placeholder)
func (h *SystemHandlers) getLoadAverage() float64 {
	// This is a placeholder implementation
	// In a real system, this would get actual load average
	return 1.5
}