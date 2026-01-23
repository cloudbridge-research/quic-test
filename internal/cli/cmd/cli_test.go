package cmd

import (
	"testing"
)

func TestNewCLI(t *testing.T) {
	cli := NewCLI()
	if cli == nil {
		t.Fatal("NewCLI() returned nil")
	}
	
	if cli.registry == nil {
		t.Fatal("CLI registry is nil")
	}
	
	if cli.flagsParser == nil {
		t.Fatal("CLI flagsParser is nil")
	}
	
	if cli.utils == nil {
		t.Fatal("CLI utils is nil")
	}
}

func TestCommandRegistry(t *testing.T) {
	registry := NewCommandRegistry()
	if registry == nil {
		t.Fatal("NewCommandRegistry() returned nil")
	}
	
	// Проверяем, что все основные команды зарегистрированы
	expectedCommands := []string{"server", "client", "test", "dashboard", "masque", "ice", "enhanced"}
	
	for _, cmdName := range expectedCommands {
		cmd, exists := registry.GetCommand(cmdName)
		if !exists {
			t.Errorf("Command '%s' not found in registry", cmdName)
		}
		
		if cmd.Name != cmdName {
			t.Errorf("Command name mismatch: expected '%s', got '%s'", cmdName, cmd.Name)
		}
		
		if cmd.Handler == nil {
			t.Errorf("Command '%s' has nil handler", cmdName)
		}
	}
}

func TestFlagsParser(t *testing.T) {
	parser := NewFlagsParser()
	if parser == nil {
		t.Fatal("NewFlagsParser() returned nil")
	}
	
	// Проверяем определения флагов
	definitions := parser.GetFlagDefinitions()
	if len(definitions) == 0 {
		t.Fatal("No flag definitions found")
	}
	
	// Проверяем наличие основных флагов
	expectedFlags := []string{"mode", "addr", "connections", "streams"}
	flagNames := make(map[string]bool)
	
	for _, def := range definitions {
		flagNames[def.Name] = true
	}
	
	for _, flagName := range expectedFlags {
		if !flagNames[flagName] {
			t.Errorf("Expected flag '%s' not found in definitions", flagName)
		}
	}
}

func TestConfigHelper(t *testing.T) {
	helper := NewConfigHelper()
	if helper == nil {
		t.Fatal("NewConfigHelper() returned nil")
	}
	
	// Тестируем извлечение значений
	config := map[string]interface{}{
		"stringValue": "test",
		"intValue":    42,
		"boolValue":   true,
	}
	
	if helper.GetString(config, "stringValue") != "test" {
		t.Error("GetString failed")
	}
	
	if helper.GetInt(config, "intValue") != 42 {
		t.Error("GetInt failed")
	}
	
	if !helper.GetBool(config, "boolValue") {
		t.Error("GetBool failed")
	}
	
	// Тестируем значения по умолчанию
	if helper.GetString(config, "nonexistent") != "" {
		t.Error("GetString should return empty string for nonexistent key")
	}
	
	if helper.GetInt(config, "nonexistent") != 0 {
		t.Error("GetInt should return 0 for nonexistent key")
	}
	
	if helper.GetBool(config, "nonexistent") {
		t.Error("GetBool should return false for nonexistent key")
	}
}

func TestCommandHandlers(t *testing.T) {
	// Тестируем создание обработчиков команд
	serverHandler := NewServerHandler()
	if serverHandler == nil {
		t.Fatal("NewServerHandler() returned nil")
	}
	
	clientHandler := NewClientHandler()
	if clientHandler == nil {
		t.Fatal("NewClientHandler() returned nil")
	}
	
	testHandler := NewTestHandler()
	if testHandler == nil {
		t.Fatal("NewTestHandler() returned nil")
	}
	
	dashboardHandler := NewDashboardHandler()
	if dashboardHandler == nil {
		t.Fatal("NewDashboardHandler() returned nil")
	}
	
	specialHandler := NewSpecialHandler()
	if specialHandler == nil {
		t.Fatal("NewSpecialHandler() returned nil")
	}
}

func TestUtilsHandler(t *testing.T) {
	registry := NewCommandRegistry()
	utils := NewUtilsHandler(registry)
	if utils == nil {
		t.Fatal("NewUtilsHandler() returned nil")
	}
	
	// Тестируем валидацию команд
	if !utils.ValidateCommand("server") {
		t.Error("ValidateCommand should return true for 'server'")
	}
	
	if utils.ValidateCommand("nonexistent") {
		t.Error("ValidateCommand should return false for nonexistent command")
	}
	
	// Тестируем получение имен команд
	names := utils.GetCommandNames()
	if len(names) == 0 {
		t.Error("GetCommandNames should return non-empty slice")
	}
}

func TestGlobalContext(t *testing.T) {
	context := NewGlobalContext()
	if context == nil {
		t.Fatal("NewGlobalContext() returned nil")
	}
	
	if context.MetricsManager == nil {
		t.Fatal("GlobalContext MetricsManager is nil")
	}
	
	// QUICManager может быть nil при инициализации
	// Это нормально, он инициализируется при необходимости
}