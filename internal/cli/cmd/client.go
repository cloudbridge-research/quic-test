package cmd

import (
	"fmt"
	"quic-test/client"
	"quic-test/internal"
)

// ClientHandler обрабатывает команду client
type ClientHandler struct {
	*BaseHandler
	configHelper *ConfigHelper
	flagsParser  *FlagsParser
}

// NewClientHandler создает новый обработчик команды client
func NewClientHandler() *ClientHandler {
	return &ClientHandler{
		BaseHandler:  NewBaseHandler(NewGlobalContext()),
		configHelper: NewConfigHelper(),
		flagsParser:  NewFlagsParser(),
	}
}

// Run выполняет команду client
func (h *ClientHandler) Run(args []string) error {
	fmt.Println("Запуск в режиме клиента...")
	
	// Парсим конфигурацию из аргументов
	_, config := h.flagsParser.ParseFlags()
	
	// Валидируем конфигурацию
	if err := h.flagsParser.ValidateFlags(config); err != nil {
		return fmt.Errorf("ошибка валидации конфигурации: %w", err)
	}
	
	// Дополнительная валидация для клиента
	if err := h.ValidateClientConfig(config); err != nil {
		return fmt.Errorf("ошибка валидации конфигурации клиента: %w", err)
	}
	
	// Создаем конфигурацию теста
	cfg := h.configHelper.CreateTestConfig("client", config)
	
	// Запускаем клиент
	return h.runClient(cfg)
}

// runClient запускает QUIC клиент
func (h *ClientHandler) runClient(cfg internal.TestConfig) error {
	fmt.Printf("Подключение к серверу %s\n", cfg.Addr)
	fmt.Printf("Соединений: %d, Потоков: %d\n", cfg.Connections, cfg.Streams)
	fmt.Printf("Размер пакета: %d, Скорость: %d пакетов/сек\n", cfg.PacketSize, cfg.Rate)
	fmt.Printf("Паттерн отправки: %s\n", cfg.Pattern)
	
	if cfg.Prometheus {
		fmt.Println("Prometheus метрики включены")
	}
	
	if cfg.AIEnabled {
		fmt.Printf("AI-маршрутизация включена (сервис: %s)\n", cfg.AIServiceURL)
	}
	
	// Запускаем клиент
	client.Run(cfg)
	return nil
}

// GetClientInfo возвращает информацию о клиенте
func (h *ClientHandler) GetClientInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "QUIC Client",
		"description": "Высокопроизводительный QUIC клиент для нагрузочного тестирования",
		"features": []string{
			"Множественные соединения",
			"Параллельные потоки",
			"Различные паттерны отправки",
			"AI-оптимизированная маршрутизация",
			"Prometheus метрики",
			"Настраиваемые параметры нагрузки",
		},
		"patterns": []string{
			"burst",   // Пакетная отправка
			"steady",  // Равномерная отправка
			"random",  // Случайная отправка
		},
	}
}

// ValidateClientConfig проверяет конфигурацию клиента
func (h *ClientHandler) ValidateClientConfig(config map[string]interface{}) error {
	// Проверяем параметры соединения
	connections := h.configHelper.GetInt(config, "connections")
	if connections <= 0 || connections > 1000 {
		return fmt.Errorf("количество соединений должно быть от 1 до 1000")
	}
	
	streams := h.configHelper.GetInt(config, "streams")
	if streams <= 0 || streams > 10000 {
		return fmt.Errorf("количество потоков должно быть от 1 до 10000")
	}
	
	// Проверяем параметры пакетов
	packetSize := h.configHelper.GetInt(config, "packetSize")
	if packetSize < 64 || packetSize > 65535 {
		return fmt.Errorf("размер пакета должен быть от 64 до 65535 байт")
	}
	
	rate := h.configHelper.GetInt(config, "rate")
	if rate <= 0 || rate > 100000 {
		return fmt.Errorf("скорость отправки должна быть от 1 до 100000 пакетов/сек")
	}
	
	// Проверяем паттерн отправки
	pattern := h.configHelper.GetString(config, "pattern")
	validPatterns := map[string]bool{
		"burst":  true,
		"steady": true,
		"random": true,
	}
	
	if !validPatterns[pattern] {
		return fmt.Errorf("неподдерживаемый паттерн отправки: %s (доступны: burst, steady, random)", pattern)
	}
	
	// Проверяем AI конфигурацию
	if h.configHelper.GetBool(config, "aiEnabled") {
		aiServiceURL := h.configHelper.GetString(config, "aiServiceURL")
		if aiServiceURL == "" {
			return fmt.Errorf("URL сервиса AI не может быть пустым при включенной AI-маршрутизации")
		}
	}
	
	return nil
}