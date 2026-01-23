package cmd

import (
	"fmt"
	"os"
)

// CLI основной интерфейс командной строки
type CLI struct {
	registry    *CommandRegistry
	flagsParser *FlagsParser
	utils       *UtilsHandler
}

// NewCLI создает новый интерфейс CLI
func NewCLI() *CLI {
	registry := NewCommandRegistry()
	flagsParser := NewFlagsParser()
	utils := NewUtilsHandler(registry)
	
	return &CLI{
		registry:    registry,
		flagsParser: flagsParser,
		utils:       utils,
	}
}

// Run запускает CLI приложение
func (cli *CLI) Run(args []string) error {
	// Парсим флаги командной строки
	mode, config := cli.flagsParser.ParseFlags()
	
	// Проверяем специальные случаи
	if len(args) > 1 {
		switch args[1] {
		case "--help", "-h", "help":
			cli.utils.ShowHelp()
			return nil
		case "--version", "-v", "version":
			cli.utils.ShowVersion()
			return nil
		case "list":
			cli.utils.ListCommands()
			return nil
		case "examples":
			cli.utils.ShowExamples()
			return nil
		}
		
		// Проверяем, не запрашивается ли информация о конкретной команде
		if len(args) > 2 && args[1] == "info" {
			cli.utils.ShowCommandInfo(args[2])
			return nil
		}
	}
	
	// Валидируем режим
	if !cli.utils.ValidateCommand(mode) {
		fmt.Printf("Неизвестный режим: %s\n", mode)
		fmt.Println("Доступные режимы:")
		cli.utils.ListCommands()
		return fmt.Errorf("неизвестный режим: %s", mode)
	}
	
	// Получаем команду
	command, exists := cli.registry.GetCommand(mode)
	if !exists {
		return fmt.Errorf("команда '%s' не найдена", mode)
	}
	
	// Выполняем команду
	return cli.executeCommand(command, config)
}

// executeCommand выполняет команду с обработкой ошибок
func (cli *CLI) executeCommand(command Command, config map[string]interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Критическая ошибка при выполнении команды '%s': %v\n", command.Name, r)
			os.Exit(1)
		}
	}()
	
	// Выполняем команду
	if err := command.Handler(nil); err != nil {
		fmt.Printf("Ошибка выполнения команды '%s': %v\n", command.Name, err)
		return err
	}
	
	return nil
}

// GetRegistry возвращает реестр команд
func (cli *CLI) GetRegistry() *CommandRegistry {
	return cli.registry
}

// GetFlagsParser возвращает парсер флагов
func (cli *CLI) GetFlagsParser() *FlagsParser {
	return cli.flagsParser
}

// GetUtils возвращает утилиты
func (cli *CLI) GetUtils() *UtilsHandler {
	return cli.utils
}

// RegisterCommand регистрирует новую команду
func (cli *CLI) RegisterCommand(name string, command Command) {
	// Добавляем команду в реестр
	cli.registry.commands[name] = command
}

// UnregisterCommand удаляет команду из реестра
func (cli *CLI) UnregisterCommand(name string) {
	delete(cli.registry.commands, name)
}

// HasCommand проверяет наличие команды
func (cli *CLI) HasCommand(name string) bool {
	_, exists := cli.registry.GetCommand(name)
	return exists
}

// ListAvailableCommands возвращает список доступных команд
func (cli *CLI) ListAvailableCommands() []string {
	return cli.utils.GetCommandNames()
}

// ShowCommandHelp показывает справку по конкретной команде
func (cli *CLI) ShowCommandHelp(commandName string) {
	cli.utils.ShowCommandInfo(commandName)
}

// ValidateAndRun валидирует параметры и запускает команду
func (cli *CLI) ValidateAndRun(mode string, config map[string]interface{}) error {
	// Валидируем флаги
	if err := cli.flagsParser.ValidateFlags(config); err != nil {
		return fmt.Errorf("ошибка валидации флагов: %w", err)
	}
	
	// Проверяем существование команды
	if !cli.HasCommand(mode) {
		return fmt.Errorf("команда '%s' не найдена", mode)
	}
	
	// Получаем и выполняем команду
	command, _ := cli.registry.GetCommand(mode)
	return cli.executeCommand(command, config)
}

// GetCommandInfo возвращает информацию о команде
func (cli *CLI) GetCommandInfo(commandName string) (Command, bool) {
	return cli.registry.GetCommand(commandName)
}

// SetVerbose устанавливает режим подробного вывода
func (cli *CLI) SetVerbose(verbose bool) {
	// TODO: реализовать управление уровнем логирования
	if verbose {
		fmt.Println("Включен подробный режим вывода")
	}
}

// GetVersion возвращает версию CLI
func (cli *CLI) GetVersion() string {
	// Используем внутреннюю функцию получения версии
	// TODO: интеграция с internal.GetVersion()
	return "1.0.0"
}

// Shutdown корректно завершает работу CLI
func (cli *CLI) Shutdown() {
	fmt.Println("Завершение работы CLI...")
	// TODO: добавить cleanup логику если необходимо
}