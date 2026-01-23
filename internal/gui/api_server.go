package gui

import (
	"time"
	"quic-test/internal"
	"quic-test/internal/gui/api"
)

// APIServer is an alias for the modular API server
type APIServer = api.APIServer

// APIResponse is an alias for the modular API response
type APIResponse = api.APIResponse

// NewAPIServer creates a new API server using the modular implementation
func NewAPIServer() *APIServer {
	// Create a test manager adapter
	testManager := &TestManagerAdapter{
		manager: NewTestManager(),
	}
	
	return api.NewAPIServer(testManager)
}

// TestManagerAdapter adapts the existing TestManager to the API interface
type TestManagerAdapter struct {
	manager *TestManager
}

// GetAllTests returns all tests
func (a *TestManagerAdapter) GetAllTests() []*api.TestSession {
	tests := a.manager.GetAllTests()
	result := make([]*api.TestSession, len(tests))
	
	for i, test := range tests {
		result[i] = &api.TestSession{
			ID:        test.ID,
			Config:    convertTestConfig(test.Config),
			Status:    test.Status,
			StartTime: test.StartTime,
			EndTime:   test.EndTime,
			Metrics:   test.Metrics,
			Logs:      test.Logs,
		}
	}
	
	return result
}

// GetTest returns a specific test
func (a *TestManagerAdapter) GetTest(id string) *api.TestSession {
	test := a.manager.GetTest(id)
	if test == nil {
		return nil
	}
	
	return &api.TestSession{
		ID:        test.ID,
		Config:    convertTestConfig(test.Config),
		Status:    test.Status,
		StartTime: test.StartTime,
		EndTime:   test.EndTime,
		Metrics:   test.Metrics,
		Logs:      test.Logs,
	}
}

// StartTest starts a new test
func (a *TestManagerAdapter) StartTest(config api.TestConfig) *api.TestSession {
	internalConfig := convertFromAPIConfig(config)
	test := a.manager.StartTest(internalConfig)
	
	return &api.TestSession{
		ID:        test.ID,
		Config:    config,
		Status:    test.Status,
		StartTime: test.StartTime,
		EndTime:   test.EndTime,
		Metrics:   test.Metrics,
		Logs:      test.Logs,
	}
}

// StopTest stops a test
func (a *TestManagerAdapter) StopTest(id string) error {
	return a.manager.StopTest(id)
}

// GetActiveTestCount returns the number of active tests
func (a *TestManagerAdapter) GetActiveTestCount() int {
	return a.manager.GetActiveTestCount()
}

// GetTotalTestCount returns the total number of tests
func (a *TestManagerAdapter) GetTotalTestCount() int {
	return a.manager.GetTotalTestCount()
}

// convertTestConfig converts internal TestConfig to API TestConfig
func convertTestConfig(internal internal.TestConfig) api.TestConfig {
	return api.TestConfig{
		Mode:              internal.Mode,
		Duration:          internal.Duration,
		Connections:       internal.Connections,
		Streams:           internal.Streams,
		Addr:              internal.Addr,
		PacketSize:        internal.PacketSize,
		Rate:              internal.Rate,
		CongestionControl: internal.CongestionControl,
		Prometheus:        internal.Prometheus,
		FECEnabled:        internal.FECEnabled,
		FECRedundancy:     internal.FECRedundancy,
		PQCEnabled:        internal.PQCEnabled,
		EmulateLatency:    internal.EmulateLatency.String(),
		EmulateLoss:       internal.EmulateLoss,
		EmulateDup:        internal.EmulateDup,
	}
}

// convertFromAPIConfig converts API TestConfig to internal TestConfig
func convertFromAPIConfig(apiConfig api.TestConfig) internal.TestConfig {
	// Parse duration string to time.Duration
	emulateLatency, _ := time.ParseDuration(apiConfig.EmulateLatency)
	
	return internal.TestConfig{
		Mode:              apiConfig.Mode,
		Duration:          apiConfig.Duration,
		Connections:       apiConfig.Connections,
		Streams:           apiConfig.Streams,
		Addr:              apiConfig.Addr,
		PacketSize:        apiConfig.PacketSize,
		Rate:              apiConfig.Rate,
		CongestionControl: apiConfig.CongestionControl,
		Prometheus:        apiConfig.Prometheus,
		FECEnabled:        apiConfig.FECEnabled,
		FECRedundancy:     apiConfig.FECRedundancy,
		PQCEnabled:        apiConfig.PQCEnabled,
		EmulateLatency:    emulateLatency,
		EmulateLoss:       apiConfig.EmulateLoss,
		EmulateDup:        apiConfig.EmulateDup,
	}
}

// GetAPIModuleInfo returns information about the API module
func GetAPIModuleInfo() map[string]interface{} {
	return api.GetAPIInfo()
}