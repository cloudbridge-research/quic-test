package scenarios

import (
	"context"
	"time"
	"quic-test/internal/scenarios/quic"
)

// Aliases for backward compatibility
type QUICSpecificScenario = quic.QUICSpecificScenario
type ScenarioStep = quic.ScenarioStep
type ScenarioExecutor = quic.ScenarioExecutor
type QUICScenarioExecutor = quic.QUICScenarioExecutor
type Transport = quic.Transport
type Connection = quic.Connection
type Stream = quic.Stream
type MetricsCollector = quic.MetricsCollector

// QUICSpecificScenarios provides access to all scenarios for backward compatibility
var QUICSpecificScenarios = quic.GetAllScenarios()

// GetQUICSpecificScenario returns a QUIC-specific scenario by ID
func GetQUICSpecificScenario(id string) (*QUICSpecificScenario, error) {
	return quic.GetQUICSpecificScenario(id)
}

// ListQUICSpecificScenarios returns a list of all QUIC-specific scenarios
func ListQUICSpecificScenarios() []string {
	return quic.ListQUICSpecificScenarios()
}

// NewQUICScenarioExecutor creates a new QUIC scenario executor
func NewQUICScenarioExecutor(transport Transport, metrics MetricsCollector) *QUICScenarioExecutor {
	return quic.NewQUICScenarioExecutor(transport, metrics)
}

// GetQUICScenarioRegistry creates a new scenario registry
func GetQUICScenarioRegistry() *quic.ScenarioRegistry {
	return quic.NewScenarioRegistry()
}

// GetQUICExecutorFactory creates a new executor factory
func GetQUICExecutorFactory() *quic.ExecutorFactory {
	return quic.NewExecutorFactory()
}

// GetQUICModuleInfo returns information about the QUIC scenarios module
func GetQUICModuleInfo() map[string]interface{} {
	return quic.GetModuleInfo()
}