package cmd

import (
	"quic-test/internal"
	"quic-test/internal/dashboard"
	"quic-test/internal/quic"
)

// Command представляет команду CLI
type Command struct {
	Name        string
	Description string
	Handler     func(args []string) error
}

// CommandRegistry содержит все доступные команды
type CommandRegistry struct {
	commands map[string]Command
}

// NewCommandRegistry создает новый реестр команд
func NewCommandRegistry() *CommandRegistry {
	registry := &CommandRegistry{
		commands: make(map[string]Command),
	}
	
	// Регистрируем все команды
	registry.registerCommands()
	return registry
}

// registerCommands регистрирует все доступные команды
func (r *CommandRegistry) registerCommands() {
	serverHandler := NewServerHandler()
	clientHandler := NewClientHandler()
	testHandler := NewTestHandler()
	dashboardHandler := NewDashboardHandler()
	specialHandler := NewSpecialHandler()

	r.commands = map[string]Command{
		"server": {
			Name:        "server",
			Description: "Запуск QUIC сервера",
			Handler:     serverHandler.Run,
		},
		"client": {
			Name:        "client",
			Description: "Запуск QUIC клиента",
			Handler:     clientHandler.Run,
		},
		"test": {
			Name:        "test",
			Description: "Запуск тестирования (сервер+клиент)",
			Handler:     testHandler.Run,
		},
		"dashboard": {
			Name:        "dashboard",
			Description: "Запуск веб-интерфейса",
			Handler:     dashboardHandler.Run,
		},
		"masque": {
			Name:        "masque",
			Description: "Запуск MASQUE тестирования",
			Handler:     specialHandler.RunMASQUE,
		},
		"ice": {
			Name:        "ice",
			Description: "Запуск ICE/STUN/TURN тестирования",
			Handler:     specialHandler.RunICE,
		},
		"enhanced": {
			Name:        "enhanced",
			Description: "Запуск расширенного тестирования (MASQUE + ICE + QUIC)",
			Handler:     specialHandler.RunEnhanced,
		},
	}
}

// GetCommand возвращает команду по имени
func (r *CommandRegistry) GetCommand(name string) (Command, bool) {
	cmd, exists := r.commands[name]
	return cmd, exists
}

// GetAllCommands возвращает все команды
func (r *CommandRegistry) GetAllCommands() map[string]Command {
	return r.commands
}

// GlobalContext содержит глобальные объекты для команд
type GlobalContext struct {
	MetricsManager *dashboard.MetricsManager
	QUICManager    *quic.QUICManager
}

// NewGlobalContext создает новый глобальный контекст
func NewGlobalContext() *GlobalContext {
	return &GlobalContext{
		MetricsManager: dashboard.NewMetricsManager(),
		QUICManager:    nil, // Инициализируется при необходимости
	}
}

// CommandHandler базовый интерфейс для обработчиков команд
type CommandHandler interface {
	Run(args []string) error
}

// BaseHandler базовая структура для обработчиков команд
type BaseHandler struct {
	context *GlobalContext
}

// NewBaseHandler создает новый базовый обработчик
func NewBaseHandler(context *GlobalContext) *BaseHandler {
	return &BaseHandler{
		context: context,
	}
}

// ConfigHelper содержит вспомогательные функции для работы с конфигурацией
type ConfigHelper struct{}

// NewConfigHelper создает новый помощник конфигурации
func NewConfigHelper() *ConfigHelper {
	return &ConfigHelper{}
}

// GetString безопасно извлекает строку из конфигурации
func (h *ConfigHelper) GetString(config map[string]interface{}, key string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return ""
}

// GetInt безопасно извлекает int из конфигурации
func (h *ConfigHelper) GetInt(config map[string]interface{}, key string) int {
	if val, ok := config[key].(int); ok {
		return val
	}
	return 0
}

// GetBool безопасно извлекает bool из конфигурации
func (h *ConfigHelper) GetBool(config map[string]interface{}, key string) bool {
	if val, ok := config[key].(bool); ok {
		return val
	}
	return false
}

// CreateTestConfig создает конфигурацию теста из параметров
func (h *ConfigHelper) CreateTestConfig(mode string, config map[string]interface{}) internal.TestConfig {
	return internal.TestConfig{
		Mode:         mode,
		Addr:         h.GetString(config, "addr"),
		CertPath:     h.GetString(config, "cert"),
		KeyPath:      h.GetString(config, "key"),
		Connections:  h.GetInt(config, "connections"),
		Streams:      h.GetInt(config, "streams"),
		PacketSize:   h.GetInt(config, "packetSize"),
		Rate:         h.GetInt(config, "rate"),
		Pattern:      h.GetString(config, "pattern"),
		Prometheus:   h.GetBool(config, "prometheus"),
		AIEnabled:    h.GetBool(config, "aiEnabled"),
		AIServiceURL: h.GetString(config, "aiServiceURL"),
	}
}