package cmd

import (
	"fmt"
	"time"
	"quic-test/client"
	"quic-test/internal"
)

// TestHandler обрабатывает команду test
type TestHandler struct {
	*BaseHandler
	configHelper *ConfigHelper
	flagsParser  *FlagsParser
}

// NewTestHandler создает новый обработчик команды test
func NewTestHandler() *TestHandler {
	return &TestHandler{
		BaseHandler:  NewBaseHandler(NewGlobalContext()),
		configHelper: NewConfigHelper(),
		flagsParser:  NewFlagsParser(),
	}
}

// Run выполняет команду test
func (h *TestHandler) Run(args []string) error {
	fmt.Println("Запуск в режиме теста (сервер+клиент)...")
	
	// Парсим конфигурацию из аргументов
	_, config := h.flagsParser.ParseFlags()
	
	// Валидируем конфигурацию
	if err := h.flagsParser.ValidateFlags(config); err != nil {
		return fmt.Errorf("ошибка валидации конфигурации: %w", err)
	}
	
	// Создаем конфигурацию теста
	cfg := h.configHelper.CreateTestConfig("test", config)
	cfg.Duration = 30 * time.Second // По умолчанию 30 секунд
	
	// Запускаем тест
	return h.runTest(cfg)
}

// runTest запускает интегрированный тест
func (h *TestHandler) runTest(cfg internal.TestConfig) error {
	fmt.Printf("Запуск интегрированного теста на %s\n", cfg.Addr)
	fmt.Printf("Длительность: %v\n", cfg.Duration)
	fmt.Printf("Соединений: %d, Потоков: %d\n", cfg.Connections, cfg.Streams)
	fmt.Printf("Размер пакета: %d, Скорость: %d пакетов/сек\n", cfg.PacketSize, cfg.Rate)
	
	if cfg.Prometheus {
		fmt.Println("Prometheus метрики включены")
	}
	
	// TODO: В будущем здесь должен быть одновременный запуск сервера и клиента
	// Пока запускаем только клиент, предполагая что сервер уже работает
	fmt.Println("Примечание: Убедитесь, что сервер уже запущен на указанном адресе")
	
	// Запускаем клиент
	client.Run(cfg)
	return nil
}

// GetTestInfo возвращает информацию о тестировании
func (h *TestHandler) GetTestInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "QUIC Integration Test",
		"description": "Интегрированное тестирование QUIC протокола",
		"features": []string{
			"Автоматическое тестирование производительности",
			"Измерение латентности и пропускной способности",
			"Тестирование различных сценариев нагрузки",
			"Сбор детальных метрик",
			"Анализ поведения congestion control",
		},
		"test_scenarios": []string{
			"Базовое подключение",
			"Множественные соединения",
			"Высокая нагрузка",
			"Длительное соединение",
			"Переменная нагрузка",
		},
	}
}

// RunBenchmark запускает бенчмарк тест
func (h *TestHandler) RunBenchmark(config map[string]interface{}) error {
	fmt.Println("Запуск бенчмарк теста...")
	
	// Создаем различные конфигурации для бенчмарка
	benchmarkConfigs := h.createBenchmarkConfigs(config)
	
	for i, cfg := range benchmarkConfigs {
		fmt.Printf("Запуск бенчмарка %d/%d: %s\n", i+1, len(benchmarkConfigs), cfg.Description)
		
		testConfig := h.configHelper.CreateTestConfig("test", config)
		testConfig.Connections = cfg.Connections
		testConfig.Streams = cfg.Streams
		testConfig.Rate = cfg.Rate
		testConfig.Duration = cfg.Duration
		
		if err := h.runTest(testConfig); err != nil {
			fmt.Printf("Ошибка в бенчмарке %d: %v\n", i+1, err)
			continue
		}
		
		// Пауза между тестами
		time.Sleep(2 * time.Second)
	}
	
	return nil
}

// BenchmarkConfig конфигурация для бенчмарка
type BenchmarkConfig struct {
	Description string
	Connections int
	Streams     int
	Rate        int
	Duration    time.Duration
}

// createBenchmarkConfigs создает набор конфигураций для бенчмарка
func (h *TestHandler) createBenchmarkConfigs(baseConfig map[string]interface{}) []BenchmarkConfig {
	return []BenchmarkConfig{
		{
			Description: "Базовый тест (1 соединение)",
			Connections: 1,
			Streams:     1,
			Rate:        100,
			Duration:    10 * time.Second,
		},
		{
			Description: "Множественные соединения",
			Connections: 5,
			Streams:     2,
			Rate:        200,
			Duration:    15 * time.Second,
		},
		{
			Description: "Высокая нагрузка",
			Connections: 10,
			Streams:     5,
			Rate:        500,
			Duration:    20 * time.Second,
		},
		{
			Description: "Стресс тест",
			Connections: 20,
			Streams:     10,
			Rate:        1000,
			Duration:    30 * time.Second,
		},
	}
}

// ValidateTestConfig проверяет конфигурацию теста
func (h *TestHandler) ValidateTestConfig(config map[string]interface{}) error {
	// Базовая валидация
	if err := h.flagsParser.ValidateFlags(config); err != nil {
		return err
	}
	
	// Проверяем длительность теста
	duration := h.configHelper.GetInt(config, "duration")
	if duration > 0 && (duration < 1 || duration > 3600) {
		return fmt.Errorf("длительность теста должна быть от 1 до 3600 секунд")
	}
	
	return nil
}