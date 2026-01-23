package api

import (
	"encoding/json"
	"net/http"
	"time"
)

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

// CORS middleware for handling cross-origin requests
func (api *APIServer) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logging middleware for API requests
func (api *APIServer) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapper, r)
		
		duration := time.Since(start)
		
		// Log the request (in a real implementation, use a proper logger)
		// fmt.Printf("[%s] %s %s - %d - %v\n", 
		//     start.Format("2006-01-02 15:04:05"), 
		//     r.Method, 
		//     r.URL.Path, 
		//     wrapper.statusCode, 
		//     duration)
		_ = duration // Suppress unused variable warning
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Rate limiting middleware (placeholder)
func (api *APIServer) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a placeholder for rate limiting
		// In a real implementation, this would check rate limits per IP/user
		next.ServeHTTP(w, r)
	})
}

// Authentication middleware (placeholder)
func (api *APIServer) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a placeholder for authentication
		// In a real implementation, this would validate API keys or JWT tokens
		next.ServeHTTP(w, r)
	})
}

// Validation helpers
func validateTestConfig(config *TestConfig) error {
	if config.Mode == "" {
		config.Mode = "test"
	}
	
	if config.Duration <= 0 {
		config.Duration = 60 * time.Second
	}
	
	if config.Connections <= 0 {
		config.Connections = 1
	}
	
	if config.Streams <= 0 {
		config.Streams = 1
	}
	
	if config.PacketSize <= 0 {
		config.PacketSize = 1200
	}
	
	if config.Rate <= 0 {
		config.Rate = 100
	}
	
	return nil
}

// Helper function to safely get string from interface{}
func getString(value interface{}, defaultValue string) string {
	if str, ok := value.(string); ok {
		return str
	}
	return defaultValue
}

// Helper function to safely get int from interface{}
func getInt(value interface{}, defaultValue int) int {
	if num, ok := value.(float64); ok {
		return int(num)
	}
	if num, ok := value.(int); ok {
		return num
	}
	return defaultValue
}

// Helper function to safely get float64 from interface{}
func getFloat64(value interface{}, defaultValue float64) float64 {
	if num, ok := value.(float64); ok {
		return num
	}
	if num, ok := value.(int); ok {
		return float64(num)
	}
	return defaultValue
}

// Helper function to safely get bool from interface{}
func getBool(value interface{}, defaultValue bool) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	return defaultValue
}