package quic

import "time"

// QUICSpecificScenarios содержит все QUIC-специфичные сценарии
var QUICSpecificScenarios = map[string]*QUICSpecificScenario{
	"version_negotiation": {
		id:          "version_negotiation",
		name:        "Version Negotiation",
		description: "Тестирование переговоров версии QUIC",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    10 * time.Second,
				Concurrency: 10,
				Parameters: map[string]interface{}{
					"versions": []string{"v1", "v2", "draft-29"},
					"force_version_negotiation": true,
				},
				Expected: map[string]interface{}{
					"success_rate": 0.95,
					"max_handshake_time": "5s",
				},
			},
		},
	},
	
	"retry_scenario": {
		id:          "retry_scenario",
		name:        "Retry Scenario",
		description: "Тестирование механизма Retry в QUIC",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    15 * time.Second,
				Concurrency: 5,
				Parameters: map[string]interface{}{
					"force_retry": true,
					"retry_delay": "100ms",
				},
				Expected: map[string]interface{}{
					"retry_rate": 0.8,
					"success_rate": 0.9,
				},
			},
		},
	},
	
	"zero_rtt_load": {
		id:          "zero_rtt_load",
		name:        "0-RTT Load Test",
		description: "Нагрузочное тестирование 0-RTT соединений",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"enable_0rtt": true,
					"session_resumption": true,
				},
				Expected: map[string]interface{}{
					"handshake_time": "1s",
				},
			},
			{
				Type:        "zero_rtt_data",
				Duration:    30 * time.Second,
				Concurrency: 20,
				Parameters: map[string]interface{}{
					"data_size": 1024,
					"packet_rate": 100,
				},
				Expected: map[string]interface{}{
					"zero_rtt_success_rate": 0.95,
					"max_latency": "50ms",
				},
			},
		},
	},
	
	"key_update_load": {
		id:          "key_update_load",
		name:        "Key Update Load Test",
		description: "Тестирование обновления ключей под нагрузкой",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"enable_key_update": true,
					"key_update_interval": "30s",
				},
				Expected: map[string]interface{}{
					"handshake_time": "2s",
				},
			},
			{
				Type:        "streams",
				Duration:    60 * time.Second,
				Concurrency: 10,
				Parameters: map[string]interface{}{
					"stream_count": 100,
					"data_size": 4096,
					"packet_rate": 200,
				},
				Expected: map[string]interface{}{
					"key_updates": 2,
					"success_rate": 0.98,
				},
			},
		},
	},
	
	"mtu_probe": {
		id:          "mtu_probe",
		name:        "MTU Probe Test",
		description: "Тестирование обнаружения MTU и PMTUD",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"enable_pmtud": true,
					"initial_mtu": 1200,
				},
				Expected: map[string]interface{}{
					"handshake_time": "3s",
				},
			},
			{
				Type:        "mtu_probe",
				Duration:    20 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"probe_sizes": []int{1200, 1400, 1500, 9000},
					"probe_interval": "5s",
				},
				Expected: map[string]interface{}{
					"mtu_discovered": 1500,
					"probe_success_rate": 0.9,
				},
			},
		},
	},
	
	"ecn_test": {
		id:          "ecn_test",
		name:        "ECN Test",
		description: "Тестирование поддержки ECN (Explicit Congestion Notification)",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"enable_ecn": true,
				},
				Expected: map[string]interface{}{
					"handshake_time": "2s",
				},
			},
			{
				Type:        "congestion_test",
				Duration:    30 * time.Second,
				Concurrency: 5,
				Parameters: map[string]interface{}{
					"congestion_control": "cubic",
					"packet_rate": 500,
					"data_size": 1400,
				},
				Expected: map[string]interface{}{
					"ecn_marked_packets": 0.1,
					"congestion_events": 5,
				},
			},
		},
	},
}