package quic

import (
	"context"
	"fmt"
	"time"
)

// ScenarioStep представляет шаг сценария
type ScenarioStep struct {
	Type        string                 `yaml:"type"`
	Duration    time.Duration          `yaml:"duration"`
	Concurrency int                    `yaml:"concurrency"`
	Parameters  map[string]interface{} `yaml:"parameters"`
	Expected    map[string]interface{} `yaml:"expected"`
}

// QUICSpecificScenario представляет QUIC-специфичный сценарий
type QUICSpecificScenario struct {
	id          string
	name        string
	description string
	steps       []ScenarioStep
}

// ID возвращает идентификатор сценария
func (s *QUICSpecificScenario) ID() string {
	return s.id
}

// Name возвращает имя сценария
func (s *QUICSpecificScenario) Name() string {
	return s.name
}

// Description возвращает описание сценария
func (s *QUICSpecificScenario) Description() string {
	return s.description
}

// Steps возвращает шаги сценария
func (s *QUICSpecificScenario) Steps() []ScenarioStep {
	return s.steps
}

// Validate проверяет корректность сценария
func (s *QUICSpecificScenario) Validate() error {
	if s.id == "" {
		return fmt.Errorf("scenario ID cannot be empty")
	}
	if s.name == "" {
		return fmt.Errorf("scenario name cannot be empty")
	}
	if len(s.steps) == 0 {
		return fmt.Errorf("scenario must have at least one step")
	}
	return nil
}

// ScenarioExecutor выполняет сценарий
type ScenarioExecutor interface {
	ExecuteStep(ctx context.Context, step ScenarioStep) error
	GetMetrics() map[string]interface{}
}

// Интерфейсы для зависимостей
type Transport interface {
	Dial(ctx context.Context, addr string) (Connection, error)
}

type Connection interface {
	OpenStream() (Stream, error)
	SendDatagram(data []byte) error
	Close() error
}

type Stream interface {
	Write(data []byte) (int, error)
	Read(data []byte) (int, error)
	Close() error
}

type MetricsCollector interface {
	GetAllMetrics() map[string]interface{}
}

// NewQUICSpecificScenario создает новый QUIC-специфичный сценарий
func NewQUICSpecificScenario(id, name, description string, steps []ScenarioStep) *QUICSpecificScenario {
	return &QUICSpecificScenario{
		id:          id,
		name:        name,
		description: description,
		steps:       steps,
	}
}