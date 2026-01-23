package cli

import (
	"fmt"
	"quic-test/internal/cli/cmd"
)

// Глобальный CLI интерфейс
var globalCLI *cmd.CLI

// init инициализирует CLI при загрузке пакета
func init() {
	globalCLI = cmd.NewCLI()
}

// Command представляет команду CLI (для обратной совместимости)
type Command = cmd.Command

// Commands содержит все доступные команды (для обратной совместимости)
var Commands map[string]Command

// init инициализирует карту команд для обратной совместимости
func init() {
	registry := globalCLI.GetRegistry()
	Commands = registry.GetAllCommands()
}

// ParseFlags парсит флаги командной строки (для обратной совместимости)
func ParseFlags() (string, map[string]interface{}) {
	parser := globalCLI.GetFlagsParser()
	return parser.ParseFlags()
}

// CreateLogger создает логгер (для обратной совместимости)
func CreateLogger() interface{} {
	return cmd.CreateLogger()
}

// ShowHelp показывает справку (для обратной совместимости)
func ShowHelp() {
	utils := globalCLI.GetUtils()
	utils.ShowHelp()
}

// RunCLI запускает CLI приложение
func RunCLI(args []string) error {
	return globalCLI.Run(args)
}

// GetCLI возвращает глобальный CLI интерфейс
func GetCLI() *cmd.CLI {
	return globalCLI
}

// RegisterCommand регистрирует новую команду
func RegisterCommand(name string, command Command) {
	globalCLI.RegisterCommand(name, command)
	// Обновляем карту для обратной совместимости
	Commands[name] = command
}

// GetCommand возвращает команду по имени
func GetCommand(name string) (Command, bool) {
	return globalCLI.GetCommandInfo(name)
}

// ValidateCommand проверяет существование команды
func ValidateCommand(name string) bool {
	return globalCLI.HasCommand(name)
}

// ListCommands возвращает список всех команд
func ListCommands() []string {
	return globalCLI.ListAvailableCommands()
}

// ExecuteCommand выполняет команду по имени
func ExecuteCommand(name string, args []string) error {
	command, exists := globalCLI.GetCommandInfo(name)
	if !exists {
		return fmt.Errorf("команда '%s' не найдена", name)
	}
	
	return command.Handler(args)
}

// SetVerbose устанавливает режим подробного вывода
func SetVerbose(verbose bool) {
	globalCLI.SetVerbose(verbose)
}

// GetVersion возвращает версию CLI
func GetVersion() string {
	return globalCLI.GetVersion()
}

// Shutdown корректно завершает работу CLI
func Shutdown() {
	globalCLI.Shutdown()
}

// Вспомогательные функции для обратной совместимости
func getString(config map[string]interface{}, key string) string {
	helper := cmd.NewConfigHelper()
	return helper.GetString(config, key)
}

func getInt(config map[string]interface{}, key string) int {
	helper := cmd.NewConfigHelper()
	return helper.GetInt(config, key)
}

func getBool(config map[string]interface{}, key string) bool {
	helper := cmd.NewConfigHelper()
	return helper.GetBool(config, key)
}

// Функции-обработчики команд для обратной совместимости
func runServer(args []string) error {
	handler := cmd.NewServerHandler()
	return handler.Run(args)
}

func runClient(args []string) error {
	handler := cmd.NewClientHandler()
	return handler.Run(args)
}

func runTest(args []string) error {
	handler := cmd.NewTestHandler()
	return handler.Run(args)
}

func runDashboard(args []string) error {
	handler := cmd.NewDashboardHandler()
	return handler.Run(args)
}

func runMASQUE(args []string) error {
	handler := cmd.NewSpecialHandler()
	return handler.RunMASQUE(args)
}

func runICE(args []string) error {
	handler := cmd.NewSpecialHandler()
	return handler.RunICE(args)
}

func runEnhanced(args []string) error {
	handler := cmd.NewSpecialHandler()
	return handler.RunEnhanced(args)
}