package cmd

import (
	"flag"
	"fmt"
	"os"
	"quic-test/internal"
)

// FlagsParser отвечает за парсинг флагов командной строки
type FlagsParser struct{}

// NewFlagsParser создает новый парсер флагов
func NewFlagsParser() *FlagsParser {
	return &FlagsParser{}
}

// ParseFlags парсит флаги командной строки
func (p *FlagsParser) ParseFlags() (string, map[string]interface{}) {
	mode := flag.String("mode", "server", "Режим работы: server, client, test, dashboard, masque, ice, enhanced")
	version := flag.Bool("version", false, "Показать версию программы")

	// Общие флаги
	addr := flag.String("addr", "localhost:8443", "Адрес сервера")
	cert := flag.String("cert", "server.crt", "Путь к сертификату")
	key := flag.String("key", "server.key", "Путь к приватному ключу")
	prometheus := flag.Bool("prometheus", false, "Включить Prometheus метрики")

	// Флаги для клиента
	connections := flag.Int("connections", 1, "Количество соединений")
	streams := flag.Int("streams", 1, "Количество потоков")
	packetSize := flag.Int("packet-size", 1024, "Размер пакета")
	rate := flag.Int("rate", 100, "Скорость отправки (пакетов/сек)")
	pattern := flag.String("pattern", "burst", "Паттерн отправки: burst, steady, random")

	// Флаги для MASQUE
	masqueServer := flag.String("masque-server", "localhost:8443", "MASQUE сервер для тестирования")
	masqueTargets := flag.String("masque-targets", "8.8.8.8:53,1.1.1.1:53", "Целевые хосты для CONNECT-UDP (через запятую)")

	// Флаги для ICE
	iceStunServers := flag.String("ice-stun", "stun.l.google.com:19302,stun1.l.google.com:19302", "STUN серверы (через запятую)")
	iceTurnServers := flag.String("ice-turn", "", "TURN серверы (через запятую)")
	iceTurnUser := flag.String("ice-turn-user", "", "TURN username")
	iceTurnPass := flag.String("ice-turn-pass", "", "TURN password")

	// Флаги для AI
	aiEnabled := flag.Bool("ai-enabled", false, "Включить AI-маршрутизацию")
	aiServiceURL := flag.String("ai-service-url", "http://localhost:5000", "URL сервиса прогнозирования")

	flag.Parse()

	// Обработка флага --version
	if *version {
		internal.PrintVersion()
		os.Exit(0)
	}

	config := map[string]interface{}{
		"addr":           *addr,
		"cert":           *cert,
		"key":            *key,
		"prometheus":     *prometheus,
		"connections":    *connections,
		"streams":        *streams,
		"packetSize":     *packetSize,
		"rate":           *rate,
		"pattern":        *pattern,
		"masqueServer":   *masqueServer,
		"masqueTargets":  *masqueTargets,
		"iceStunServers": *iceStunServers,
		"iceTurnServers": *iceTurnServers,
		"iceTurnUser":    *iceTurnUser,
		"iceTurnPass":    *iceTurnPass,
		"aiEnabled":      *aiEnabled,
		"aiServiceURL":   *aiServiceURL,
	}

	return *mode, config
}

// FlagDefinition представляет определение флага
type FlagDefinition struct {
	Name         string
	DefaultValue interface{}
	Description  string
	Category     string
}

// GetFlagDefinitions возвращает все определения флагов
func (p *FlagsParser) GetFlagDefinitions() []FlagDefinition {
	return []FlagDefinition{
		// Общие флаги
		{Name: "mode", DefaultValue: "server", Description: "Режим работы", Category: "general"},
		{Name: "version", DefaultValue: false, Description: "Показать версию", Category: "general"},
		{Name: "addr", DefaultValue: "localhost:8443", Description: "Адрес сервера", Category: "general"},
		{Name: "cert", DefaultValue: "server.crt", Description: "Путь к сертификату", Category: "general"},
		{Name: "key", DefaultValue: "server.key", Description: "Путь к приватному ключу", Category: "general"},
		{Name: "prometheus", DefaultValue: false, Description: "Включить Prometheus метрики", Category: "general"},

		// Флаги клиента
		{Name: "connections", DefaultValue: 1, Description: "Количество соединений", Category: "client"},
		{Name: "streams", DefaultValue: 1, Description: "Количество потоков", Category: "client"},
		{Name: "packet-size", DefaultValue: 1024, Description: "Размер пакета", Category: "client"},
		{Name: "rate", DefaultValue: 100, Description: "Скорость отправки (пакетов/сек)", Category: "client"},
		{Name: "pattern", DefaultValue: "burst", Description: "Паттерн отправки", Category: "client"},

		// Флаги MASQUE
		{Name: "masque-server", DefaultValue: "localhost:8443", Description: "MASQUE сервер", Category: "masque"},
		{Name: "masque-targets", DefaultValue: "8.8.8.8:53,1.1.1.1:53", Description: "Целевые хосты", Category: "masque"},

		// Флаги ICE
		{Name: "ice-stun", DefaultValue: "stun.l.google.com:19302", Description: "STUN серверы", Category: "ice"},
		{Name: "ice-turn", DefaultValue: "", Description: "TURN серверы", Category: "ice"},
		{Name: "ice-turn-user", DefaultValue: "", Description: "TURN username", Category: "ice"},
		{Name: "ice-turn-pass", DefaultValue: "", Description: "TURN password", Category: "ice"},

		// Флаги AI
		{Name: "ai-enabled", DefaultValue: false, Description: "Включить AI-маршрутизацию", Category: "ai"},
		{Name: "ai-service-url", DefaultValue: "http://localhost:5000", Description: "URL сервиса прогнозирования", Category: "ai"},
	}
}

// ValidateFlags проверяет корректность флагов
func (p *FlagsParser) ValidateFlags(config map[string]interface{}) error {
	// Проверка обязательных параметров
	if addr, ok := config["addr"].(string); !ok || addr == "" {
		return fmt.Errorf("адрес сервера не может быть пустым")
	}

	// Проверка числовых параметров
	if connections, ok := config["connections"].(int); ok && connections <= 0 {
		return fmt.Errorf("количество соединений должно быть больше 0")
	}

	if streams, ok := config["streams"].(int); ok && streams <= 0 {
		return fmt.Errorf("количество потоков должно быть больше 0")
	}

	if packetSize, ok := config["packetSize"].(int); ok && (packetSize <= 0 || packetSize > 65535) {
		return fmt.Errorf("размер пакета должен быть от 1 до 65535 байт")
	}

	if rate, ok := config["rate"].(int); ok && rate <= 0 {
		return fmt.Errorf("скорость отправки должна быть больше 0")
	}

	return nil
}