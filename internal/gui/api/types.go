package api

import (
	"net/http"
	"time"
)

// APIServer handles REST API requests
type APIServer struct {
	testManager TestManager
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// TestManager interface for managing tests
type TestManager interface {
	GetAllTests() []*TestSession
	GetTest(id string) *TestSession
	StartTest(config TestConfig) *TestSession
	StopTest(id string) error
	GetActiveTestCount() int
	GetTotalTestCount() int
}

// TestSession represents a test session
type TestSession struct {
	ID          string                 `json:"id"`
	Config      TestConfig             `json:"config"`
	Status      string                 `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Metrics     map[string]interface{} `json:"metrics"`
	Logs        []string               `json:"logs"`
}

// TestConfig represents test configuration
type TestConfig struct {
	Mode                string        `json:"mode"`
	Duration            time.Duration `json:"duration"`
	Connections         int           `json:"connections"`
	Streams             int           `json:"streams"`
	Addr                string        `json:"addr"`
	PacketSize          int           `json:"packet_size"`
	Rate                int           `json:"rate"`
	CongestionControl   string        `json:"congestion_control"`
	Prometheus          bool          `json:"prometheus"`
	FECEnabled          bool          `json:"fec_enabled"`
	FECRedundancy       float64       `json:"fec_redundancy"`
	PQCEnabled          bool          `json:"pqc_enabled"`
	EmulateLatency      string        `json:"emulate_latency"`
	EmulateLoss         float64       `json:"emulate_loss"`
	EmulateDup          float64       `json:"emulate_dup"`
}

// NetworkPreset represents a network configuration preset
type NetworkPreset struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Latency     string `json:"latency"`
	Bandwidth   string `json:"bandwidth"`
	Loss        string `json:"loss"`
}

// TestProfile represents a test configuration profile
type TestProfile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Duration    string `json:"duration"`
	Connections int    `json:"connections"`
	Streams     int    `json:"streams"`
	Rate        int    `json:"rate"`
}

// SystemStatus represents system status information
type SystemStatus struct {
	Version     string    `json:"version"`
	Uptime      string    `json:"uptime"`
	ActiveTests int       `json:"active_tests"`
	TotalTests  int       `json:"total_tests"`
	Memory      MemInfo   `json:"memory"`
	CPU         CPUInfo   `json:"cpu"`
	Timestamp   time.Time `json:"timestamp"`
}

// MemInfo represents memory information
type MemInfo struct {
	Used      uint64  `json:"used"`
	Available uint64  `json:"available"`
	Total     uint64  `json:"total"`
	Percent   float64 `json:"percent"`
}

// CPUInfo represents CPU information
type CPUInfo struct {
	Cores   int     `json:"cores"`
	Usage   float64 `json:"usage"`
	LoadAvg float64 `json:"load_avg"`
}

// HealthStatus represents health check status
type HealthStatus struct {
	Status    string            `json:"status"`
	Checks    map[string]string `json:"checks"`
	Timestamp time.Time         `json:"timestamp"`
}

// Handler represents an HTTP handler function
type Handler func(w http.ResponseWriter, r *http.Request)

// NewAPIServer creates a new API server
func NewAPIServer(testManager TestManager) *APIServer {
	return &APIServer{
		testManager: testManager,
	}
}