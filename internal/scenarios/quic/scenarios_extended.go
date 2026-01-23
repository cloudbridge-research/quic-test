package quic

import "time"

// ExtendedQUICScenarios содержит дополнительные QUIC-специфичные сценарии
var ExtendedQUICScenarios = map[string]*QUICSpecificScenario{
	"nat_rebinding": {
		id:          "nat_rebinding",
		name:        "NAT Rebinding Test",
		description: "Тестирование переподключения при изменении NAT",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"enable_nat_rebinding": true,
				},
				Expected: map[string]interface{}{
					"handshake_time": "2s",
				},
			},
			{
				Type:        "nat_rebind",
				Duration:    20 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"rebind_interval": "10s",
					"rebind_delay": "2s",
				},
				Expected: map[string]interface{}{
					"rebind_success_rate": 0.9,
					"recovery_time": "5s",
				},
			},
		},
	},
	
	"flow_control_limits": {
		id:          "flow_control_limits",
		name:        "Flow Control Limits Test",
		description: "Тестирование лимитов flow control",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"max_stream_data": 1024 * 1024, // 1MB
					"max_connection_data": 10 * 1024 * 1024, // 10MB
				},
				Expected: map[string]interface{}{
					"handshake_time": "2s",
				},
			},
			{
				Type:        "flow_control_test",
				Duration:    30 * time.Second,
				Concurrency: 5,
				Parameters: map[string]interface{}{
					"stream_count": 1000,
					"data_size": 2 * 1024 * 1024, // 2MB per stream
					"packet_rate": 1000,
				},
				Expected: map[string]interface{}{
					"flow_control_events": 10,
					"success_rate": 0.95,
				},
			},
		},
	},
	
	"datagrams_vs_streams": {
		id:          "datagrams_vs_streams",
		name:        "Datagrams vs Streams Test",
		description: "Сравнение производительности datagrams и streams",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"enable_datagrams": true,
					"enable_streams": true,
				},
				Expected: map[string]interface{}{
					"handshake_time": "2s",
				},
			},
			{
				Type:        "datagrams_test",
				Duration:    20 * time.Second,
				Concurrency: 10,
				Parameters: map[string]interface{}{
					"datagram_size": 1200,
					"datagram_rate": 500,
				},
				Expected: map[string]interface{}{
					"datagram_success_rate": 0.98,
					"datagram_latency": "10ms",
				},
			},
			{
				Type:        "streams_test",
				Duration:    20 * time.Second,
				Concurrency: 10,
				Parameters: map[string]interface{}{
					"stream_count": 100,
					"stream_data_size": 4096,
					"stream_rate": 200,
				},
				Expected: map[string]interface{}{
					"stream_success_rate": 0.99,
					"stream_latency": "15ms",
				},
			},
		},
	},
	
	"congestion_control_switch": {
		id:          "congestion_control_switch",
		name:        "Congestion Control Switch Test",
		description: "Тестирование переключения алгоритмов управления перегрузкой",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"congestion_control": "cubic",
				},
				Expected: map[string]interface{}{
					"handshake_time": "2s",
				},
			},
			{
				Type:        "congestion_test_cubic",
				Duration:    30 * time.Second,
				Concurrency: 5,
				Parameters: map[string]interface{}{
					"congestion_control": "cubic",
					"packet_rate": 1000,
					"data_size": 1400,
				},
				Expected: map[string]interface{}{
					"throughput": 50.0, // Mbps
					"congestion_events": 5,
				},
			},
			{
				Type:        "congestion_test_bbr",
				Duration:    30 * time.Second,
				Concurrency: 5,
				Parameters: map[string]interface{}{
					"congestion_control": "bbr",
					"packet_rate": 1000,
					"data_size": 1400,
				},
				Expected: map[string]interface{}{
					"throughput": 60.0, // Mbps
					"congestion_events": 3,
				},
			},
		},
	},
	
	"connection_migration": {
		id:          "connection_migration",
		name:        "Connection Migration Test",
		description: "Тестирование миграции соединения между адресами",
		steps: []ScenarioStep{
			{
				Type:        "handshake",
				Duration:    5 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"enable_connection_migration": true,
				},
				Expected: map[string]interface{}{
					"handshake_time": "2s",
				},
			},
			{
				Type:        "migration_test",
				Duration:    30 * time.Second,
				Concurrency: 1,
				Parameters: map[string]interface{}{
					"migration_interval": "15s",
					"new_addresses": []string{"192.168.1.100", "192.168.1.101"},
				},
				Expected: map[string]interface{}{
					"migration_success_rate": 0.95,
					"migration_time": "2s",
				},
			},
		},
	},
}

// GetAllScenarios объединяет все сценарии из разных файлов
func GetAllScenarios() map[string]*QUICSpecificScenario {
	allScenarios := make(map[string]*QUICSpecificScenario)
	
	// Копируем основные сценарии
	for id, scenario := range QUICSpecificScenarios {
		allScenarios[id] = scenario
	}
	
	// Копируем расширенные сценарии
	for id, scenario := range ExtendedQUICScenarios {
		allScenarios[id] = scenario
	}
	
	return allScenarios
}