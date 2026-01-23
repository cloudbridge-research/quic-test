// Package quic provides QUIC-specific test scenarios and execution framework
//
// This package contains a comprehensive set of QUIC protocol test scenarios
// including handshake tests, performance tests, security tests, and network
// condition tests. It provides both scenario definitions and execution framework.
package quic

// ScenarioRegistry provides access to all QUIC scenarios
type ScenarioRegistry struct{}

// NewScenarioRegistry creates a new scenario registry
func NewScenarioRegistry() *ScenarioRegistry {
	return &ScenarioRegistry{}
}

// GetScenario returns a scenario by ID
func (r *ScenarioRegistry) GetScenario(id string) (*QUICSpecificScenario, error) {
	return GetQUICSpecificScenario(id)
}

// ListScenarios returns all available scenario IDs
func (r *ScenarioRegistry) ListScenarios() []string {
	return ListQUICSpecificScenarios()
}

// GetScenariosByCategory returns scenarios filtered by category
func (r *ScenarioRegistry) GetScenariosByCategory(category string) map[string]*QUICSpecificScenario {
	return GetScenariosByCategory(category)
}

// ValidateScenarios validates all scenarios
func (r *ScenarioRegistry) ValidateScenarios() []error {
	return ValidateAllScenarios()
}

// GetCategories returns available scenario categories
func (r *ScenarioRegistry) GetCategories() []string {
	return GetScenarioCategories()
}

// ExecutorFactory creates scenario executors
type ExecutorFactory struct{}

// NewExecutorFactory creates a new executor factory
func NewExecutorFactory() *ExecutorFactory {
	return &ExecutorFactory{}
}

// CreateExecutor creates a new QUIC scenario executor
func (f *ExecutorFactory) CreateExecutor(transport Transport, metrics MetricsCollector) ScenarioExecutor {
	return NewQUICScenarioExecutor(transport, metrics)
}

// GetModuleInfo returns information about the QUIC scenarios module
func GetModuleInfo() map[string]interface{} {
	allScenarios := GetAllScenarios()
	
	return map[string]interface{}{
		"name":            "QUIC Test Scenarios",
		"version":         "1.0.0",
		"description":     "Comprehensive QUIC protocol test scenarios",
		"scenario_count":  len(allScenarios),
		"categories":      GetScenarioCategories(),
		"features": []string{
			"Version negotiation tests",
			"0-RTT data tests", 
			"Key update tests",
			"MTU discovery tests",
			"ECN support tests",
			"NAT rebinding tests",
			"Flow control tests",
			"Connection migration tests",
			"Performance comparison tests",
			"Congestion control tests",
		},
	}
}