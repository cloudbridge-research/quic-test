package cmd

import (
	"flag"
	"fmt"
	"quic-test/internal"
)

// UtilsHandler содержит вспомогательные функции CLI
type UtilsHandler struct {
	registry *CommandRegistry
}

// NewUtilsHandler создает новый обработчик утилит
func NewUtilsHandler(registry *CommandRegistry) *UtilsHandler {
	return &UtilsHandler{
		registry: registry,
	}
}

// ShowHelp показывает справку по командам
func (h *UtilsHandler) ShowHelp() {
	// Показываем версию
	version, err := internal.GetVersion()
	if err == nil {
		fmt.Printf("2GC Network Protocol Suite v%s - Комплексное тестирование сетевых протоколов\n", version)
	} else {
		fmt.Println("2GC Network Protocol Suite - Комплексное тестирование сетевых протоколов")
	}
	fmt.Println()
	fmt.Println("Использование:")
	fmt.Println("  quic-test -mode=<режим> [флаги]")
	fmt.Println()
	fmt.Println("Режимы:")
	
	// Получаем все команды из реестра
	commands := h.registry.GetAllCommands()
	for name, cmd := range commands {
		fmt.Printf("  %-10s - %s\n", name, cmd.Description)
	}
	
	fmt.Println()
	fmt.Println("Флаги:")
	flag.PrintDefaults()
}

// ShowVersion показывает версию программы
func (h *UtilsHandler) ShowVersion() {
	version, err := internal.GetVersion()
	if err != nil {
		fmt.Println("Версия неизвестна")
		return
	}
	fmt.Printf("2GC Network Protocol Suite v%s\n", version)
}

// ShowCommandInfo показывает детальную информацию о команде
func (h *UtilsHandler) ShowCommandInfo(commandName string) {
	cmd, exists := h.registry.GetCommand(commandName)
	if !exists {
		fmt.Printf("Команда '%s' не найдена\n", commandName)
		return
	}
	
	fmt.Printf("Команда: %s\n", cmd.Name)
	fmt.Printf("Описание: %s\n", cmd.Description)
	
	// Показываем дополнительную информацию в зависимости от команды
	switch commandName {
	case "server":
		h.showServerInfo()
	case "client":
		h.showClientInfo()
	case "test":
		h.showTestInfo()
	case "dashboard":
		h.showDashboardInfo()
	case "masque":
		h.showMASQUEInfo()
	case "ice":
		h.showICEInfo()
	case "enhanced":
		h.showEnhancedInfo()
	}
}

// showServerInfo показывает информацию о сервере
func (h *UtilsHandler) showServerInfo() {
	fmt.Println("\nВозможности сервера:")
	fmt.Println("  - HTTP/3 поддержка")
	fmt.Println("  - Множественные соединения")
	fmt.Println("  - Prometheus метрики")
	fmt.Println("  - TLS 1.3 шифрование")
	fmt.Println("  - Congestion control алгоритмы")
	
	fmt.Println("\nПоддерживаемые алгоритмы:")
	fmt.Println("  - BBR, BBRv2, BBRv3")
	fmt.Println("  - CUBIC")
	fmt.Println("  - NewReno")
}

// showClientInfo показывает информацию о клиенте
func (h *UtilsHandler) showClientInfo() {
	fmt.Println("\nВозможности клиента:")
	fmt.Println("  - Множественные соединения")
	fmt.Println("  - Параллельные потоки")
	fmt.Println("  - Различные паттерны отправки")
	fmt.Println("  - AI-оптимизированная маршрутизация")
	fmt.Println("  - Prometheus метрики")
	
	fmt.Println("\nПаттерны отправки:")
	fmt.Println("  - burst   - Пакетная отправка")
	fmt.Println("  - steady  - Равномерная отправка")
	fmt.Println("  - random  - Случайная отправка")
}

// showTestInfo показывает информацию о тестировании
func (h *UtilsHandler) showTestInfo() {
	fmt.Println("\nВозможности тестирования:")
	fmt.Println("  - Автоматическое тестирование производительности")
	fmt.Println("  - Измерение латентности и пропускной способности")
	fmt.Println("  - Тестирование различных сценариев нагрузки")
	fmt.Println("  - Сбор детальных метрик")
	
	fmt.Println("\nСценарии тестирования:")
	fmt.Println("  - Базовое подключение")
	fmt.Println("  - Множественные соединения")
	fmt.Println("  - Высокая нагрузка")
	fmt.Println("  - Длительное соединение")
}

// showDashboardInfo показывает информацию о dashboard
func (h *UtilsHandler) showDashboardInfo() {
	fmt.Println("\nВозможности dashboard:")
	fmt.Println("  - Управление QUIC сервером и клиентом")
	fmt.Println("  - Мониторинг метрик в реальном времени")
	fmt.Println("  - Настройка параметров тестирования")
	fmt.Println("  - Визуализация результатов")
	fmt.Println("  - REST API для автоматизации")
	
	fmt.Println("\nAPI endpoints:")
	fmt.Println("  - /api/status, /api/metrics, /api/config")
	fmt.Println("  - /api/server/start, /api/server/stop")
	fmt.Println("  - /api/client/start, /api/client/stop")
	fmt.Println("  - /api/test/start")
}

// showMASQUEInfo показывает информацию о MASQUE
func (h *UtilsHandler) showMASQUEInfo() {
	fmt.Println("\nMASQUE тестирование:")
	fmt.Println("  - CONNECT-UDP проксирование")
	fmt.Println("  - Множественные целевые хосты")
	fmt.Println("  - Измерение производительности прокси")
	fmt.Println("  - Анализ задержек через прокси")
	
	fmt.Println("\nПротоколы:")
	fmt.Println("  - MASQUE CONNECT-UDP")
	fmt.Println("  - HTTP/3 over QUIC")
}

// showICEInfo показывает информацию о ICE
func (h *UtilsHandler) showICEInfo() {
	fmt.Println("\nICE тестирование:")
	fmt.Println("  - STUN binding requests")
	fmt.Println("  - TURN allocation и relay")
	fmt.Println("  - ICE candidate gathering")
	fmt.Println("  - Connectivity checks")
	fmt.Println("  - NAT type detection")
	
	fmt.Println("\nПротоколы:")
	fmt.Println("  - STUN (RFC 5389)")
	fmt.Println("  - TURN (RFC 5766)")
	fmt.Println("  - ICE (RFC 8445)")
}

// showEnhancedInfo показывает информацию о расширенном тестировании
func (h *UtilsHandler) showEnhancedInfo() {
	fmt.Println("\nРасширенное тестирование:")
	fmt.Println("  - QUIC базовое тестирование")
	fmt.Println("  - MASQUE проксирование")
	fmt.Println("  - ICE NAT traversal")
	
	fmt.Println("\nВозможности:")
	fmt.Println("  - Интегрированные метрики")
	fmt.Println("  - Сравнительный анализ")
	fmt.Println("  - Комплексные сценарии")
	fmt.Println("  - Автоматическая диагностика")
}

// ListCommands показывает список всех доступных команд
func (h *UtilsHandler) ListCommands() {
	fmt.Println("Доступные команды:")
	
	commands := h.registry.GetAllCommands()
	for name, cmd := range commands {
		fmt.Printf("  %s - %s\n", name, cmd.Description)
	}
}

// ValidateCommand проверяет существование команды
func (h *UtilsHandler) ValidateCommand(commandName string) bool {
	_, exists := h.registry.GetCommand(commandName)
	return exists
}

// GetCommandNames возвращает список имен всех команд
func (h *UtilsHandler) GetCommandNames() []string {
	commands := h.registry.GetAllCommands()
	names := make([]string, 0, len(commands))
	
	for name := range commands {
		names = append(names, name)
	}
	
	return names
}

// FormatUsage форматирует строку использования для команды
func (h *UtilsHandler) FormatUsage(commandName string) string {
	if !h.ValidateCommand(commandName) {
		return fmt.Sprintf("Команда '%s' не найдена", commandName)
	}
	
	return fmt.Sprintf("quic-test -mode=%s [флаги]", commandName)
}

// ShowExamples показывает примеры использования
func (h *UtilsHandler) ShowExamples() {
	fmt.Println("Примеры использования:")
	fmt.Println()
	
	fmt.Println("Запуск сервера:")
	fmt.Println("  quic-test -mode=server -addr=:8443 -cert=server.crt -key=server.key")
	fmt.Println()
	
	fmt.Println("Запуск клиента:")
	fmt.Println("  quic-test -mode=client -addr=localhost:8443 -connections=5 -streams=10")
	fmt.Println()
	
	fmt.Println("Запуск теста:")
	fmt.Println("  quic-test -mode=test -addr=localhost:8443 -rate=1000 -pattern=random")
	fmt.Println()
	
	fmt.Println("Запуск dashboard:")
	fmt.Println("  quic-test -mode=dashboard")
	fmt.Println()
	
	fmt.Println("MASQUE тестирование:")
	fmt.Println("  quic-test -mode=masque -masque-server=proxy.example.com:8443")
	fmt.Println()
	
	fmt.Println("ICE тестирование:")
	fmt.Println("  quic-test -mode=ice -ice-stun=stun.l.google.com:19302")
	fmt.Println()
	
	fmt.Println("Расширенное тестирование:")
	fmt.Println("  quic-test -mode=enhanced -prometheus")
}