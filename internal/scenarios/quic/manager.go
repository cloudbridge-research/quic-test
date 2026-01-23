package quic

import "fmt"

// GetQUICSpecificScenario возвращает QUIC-специфичный сценарий по ID
func GetQUICSpecificScenario(id string) (*QUICSpecificScenario, error) {
	allScenarios := GetAllScenarios()
	scenario, exists := allScenarios[id]
	if !exists {
		return nil, fmt.Errorf("scenario with ID '%s' not found", id)
	}
	return scenario, nil
}

// ListQUICSpecificScenarios возвращает список всех QUIC-специфичных сценариев
func ListQUICSpecificScenarios() []string {
	allScenarios := GetAllScenarios()
	scenarios := make([]string, 0, len(allScenarios))
	for id := range allScenarios {
		scenarios = append(scenarios, id)
	}
	return scenarios
}

// GetScenariosByCategory возвращает сценарии по категории
func GetScenariosByCategory(category string) map[string]*QUICSpecificScenario {
	allScenarios := GetAllScenarios()
	result := make(map[string]*QUICSpecificScenario)
	
	for id, scenario := range allScenarios {
		if matchesCategory(scenario, category) {
			result[id] = scenario
		}
	}
	
	return result
}

// matchesCategory проверяет, соответствует ли сценарий категории
func matchesCategory(scenario *QUICSpecificScenario, category string) bool {
	switch category {
	case "handshake":
		return containsStepType(scenario, "handshake")
	case "performance":
		return containsStepType(scenario, "streams_test") || 
			   containsStepType(scenario, "datagrams_test") ||
			   containsStepType(scenario, "congestion_test")
	case "security":
		return containsStepType(scenario, "key_update") ||
			   containsStepType(scenario, "zero_rtt_data")
	case "network":
		return containsStepType(scenario, "mtu_probe") ||
			   containsStepType(scenario, "ecn_test") ||
			   containsStepType(scenario, "nat_rebind")
	case "migration":
		return containsStepType(scenario, "migration_test")
	case "flow_control":
		return containsStepType(scenario, "flow_control_test")
	default:
		return false
	}
}

// containsStepType проверяет, содержит ли сценарий шаг определенного типа
func containsStepType(scenario *QUICSpecificScenario, stepType string) bool {
	for _, step := range scenario.steps {
		if step.Type == stepType {
			return true
		}
	}
	return false
}

// ValidateAllScenarios проверяет корректность всех сценариев
func ValidateAllScenarios() []error {
	allScenarios := GetAllScenarios()
	var errors []error
	
	for id, scenario := range allScenarios {
		if err := scenario.Validate(); err != nil {
			errors = append(errors, fmt.Errorf("scenario '%s': %w", id, err))
		}
	}
	
	return errors
}

// GetScenarioInfo возвращает информацию о сценарии
func GetScenarioInfo(id string) (map[string]interface{}, error) {
	scenario, err := GetQUICSpecificScenario(id)
	if err != nil {
		return nil, err
	}
	
	info := map[string]interface{}{
		"id":          scenario.ID(),
		"name":        scenario.Name(),
		"description": scenario.Description(),
		"step_count":  len(scenario.Steps()),
		"steps":       make([]map[string]interface{}, 0, len(scenario.Steps())),
	}
	
	// Добавляем информацию о шагах
	for i, step := range scenario.Steps() {
		stepInfo := map[string]interface{}{
			"index":       i,
			"type":        step.Type,
			"duration":    step.Duration.String(),
			"concurrency": step.Concurrency,
			"parameters":  step.Parameters,
			"expected":    step.Expected,
		}
		info["steps"] = append(info["steps"].([]map[string]interface{}), stepInfo)
	}
	
	return info, nil
}

// GetScenarioCategories возвращает список доступных категорий
func GetScenarioCategories() []string {
	return []string{
		"handshake",
		"performance", 
		"security",
		"network",
		"migration",
		"flow_control",
	}
}