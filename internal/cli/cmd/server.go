package cmd

import (
	"fmt"
	"quic-test/internal"
	"quic-test/server"
)

// ServerHandler обрабатывает команду server
type ServerHandler struct {
	*BaseHandler
	configHelper *ConfigHelper
	flagsParser  *FlagsParser
}

// NewServerHandler создает новый обработчик команды server
func NewServerHandler() *ServerHandler {
	return &ServerHandler{
		BaseHandler:  NewBaseHandler(NewGlobalContext()),
		configHelper: NewConfigHelper(),
		flagsParser:  NewFlagsParser(),
	}
}

// Run выполняет команду server
func (h *ServerHandler) Run(args []string) error {
	fmt.Println("Запуск в режиме сервера...")
	
	// Парсим конфигурацию из аргументов
	_, config := h.flagsParser.ParseFlags()
	
	// Валидируем конфигурацию
	if err := h.flagsParser.ValidateFlags(config); err != nil {
		return fmt.Errorf("ошибка валидации конфигурации: %w", err)
	}
	
	// Создаем конфигурацию теста
	cfg := h.configHelper.CreateTestConfig("server", config)
	
	// Запускаем сервер
	return h.runServer(cfg)
}

// runServer запускает QUIC сервер
func (h *ServerHandler) runServer(cfg internal.TestConfig) error {
	fmt.Printf("Запуск QUIC сервера на %s\n", cfg.Addr)
	
	if cfg.Prometheus {
		fmt.Println("Prometheus метрики включены")
	}
	
	// Запускаем сервер
	server.Run(cfg)
	return nil
}

// GetServerInfo возвращает информацию о сервере
func (h *ServerHandler) GetServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "QUIC Server",
		"description": "Высокопроизводительный QUIC сервер для тестирования",
		"features": []string{
			"HTTP/3 поддержка",
			"Множественные соединения",
			"Prometheus метрики",
			"TLS 1.3 шифрование",
			"Congestion control алгоритмы",
		},
		"supported_algorithms": []string{
			"BBR",
			"BBRv2", 
			"BBRv3",
			"CUBIC",
			"NewReno",
		},
	}
}

// ValidateServerConfig проверяет конфигурацию сервера
func (h *ServerHandler) ValidateServerConfig(config map[string]interface{}) error {
	// Проверяем наличие сертификата и ключа
	certPath := h.configHelper.GetString(config, "cert")
	keyPath := h.configHelper.GetString(config, "key")
	
	if certPath == "" {
		return fmt.Errorf("путь к сертификату не может быть пустым")
	}
	
	if keyPath == "" {
		return fmt.Errorf("путь к приватному ключу не может быть пустым")
	}
	
	// Проверяем адрес
	addr := h.configHelper.GetString(config, "addr")
	if addr == "" {
		return fmt.Errorf("адрес сервера не может быть пустым")
	}
	
	return nil
}