package container

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// ContainerManager управляет Docker контейнерами
type ContainerManager struct {
	mu sync.RWMutex
}

// NewContainerManager создает новый менеджер контейнеров
func NewContainerManager() *ContainerManager {
	return &ContainerManager{}
}

// ContainerStatus представляет статус контейнера
type ContainerStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Health    string    `json:"health"`
	Ports     []string  `json:"ports"`
	Image     string    `json:"image"`
	CreatedAt time.Time `json:"created_at"`
	StartedAt time.Time `json:"started_at"`
}

// ServiceConfig конфигурация для запуска сервиса
type ServiceConfig struct {
	ServiceName string            `json:"service_name"`
	Environment map[string]string `json:"environment"`
	Command     []string          `json:"command"`
}

// GetContainerStatus получает статус контейнера
func (cm *ContainerManager) GetContainerStatus(containerName string) (*ContainerStatus, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Выполняем docker inspect для получения информации о контейнере
	cmd := exec.Command("docker", "inspect", containerName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container %s: %w", containerName, err)
	}

	var containers []struct {
		Name   string `json:"Name"`
		State  struct {
			Status     string    `json:"Status"`
			Health     struct {
				Status string `json:"Status"`
			} `json:"Health"`
			StartedAt time.Time `json:"StartedAt"`
		} `json:"State"`
		Config struct {
			Image string `json:"Image"`
		} `json:"Config"`
		NetworkSettings struct {
			Ports map[string][]struct {
				HostPort string `json:"HostPort"`
			} `json:"Ports"`
		} `json:"NetworkSettings"`
		Created time.Time `json:"Created"`
	}

	if err := json.Unmarshal(output, &containers); err != nil {
		return nil, fmt.Errorf("failed to parse container info: %w", err)
	}

	if len(containers) == 0 {
		return nil, fmt.Errorf("container %s not found", containerName)
	}

	container := containers[0]
	
	// Собираем информацию о портах
	var ports []string
	for containerPort, bindings := range container.NetworkSettings.Ports {
		for _, binding := range bindings {
			if binding.HostPort != "" {
				ports = append(ports, fmt.Sprintf("%s:%s", binding.HostPort, containerPort))
			}
		}
	}

	status := &ContainerStatus{
		Name:      strings.TrimPrefix(container.Name, "/"),
		Status:    container.State.Status,
		Health:    container.State.Health.Status,
		Ports:     ports,
		Image:     container.Config.Image,
		CreatedAt: container.Created,
		StartedAt: container.State.StartedAt,
	}

	return status, nil
}

// StartService запускает сервис через docker-compose
func (cm *ContainerManager) StartService(serviceName string, config *ServiceConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Проверяем, не запущен ли уже сервис
	if cm.isServiceRunning(serviceName) {
		return fmt.Errorf("service %s is already running", serviceName)
	}

	// Формируем команду docker-compose
	args := []string{"up", "-d", serviceName}
	
	// Добавляем переменные окружения если есть
	if config != nil && len(config.Environment) > 0 {
		for key, value := range config.Environment {
			args = append([]string{"-e", fmt.Sprintf("%s=%s", key, value)}, args...)
		}
	}

	cmd := exec.Command("docker-compose", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start service %s: %w\nOutput: %s", serviceName, err, string(output))
	}

	// Ждем немного, чтобы сервис успел запуститься
	time.Sleep(2 * time.Second)

	return nil
}

// StopService останавливает сервис через docker-compose
func (cm *ContainerManager) StopService(serviceName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cmd := exec.Command("docker-compose", "stop", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop service %s: %w\nOutput: %s", serviceName, err, string(output))
	}

	return nil
}

// RestartService перезапускает сервис
func (cm *ContainerManager) RestartService(serviceName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cmd := exec.Command("docker-compose", "restart", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart service %s: %w\nOutput: %s", serviceName, err, string(output))
	}

	return nil
}

// GetServiceLogs получает логи сервиса
func (cm *ContainerManager) GetServiceLogs(serviceName string, lines int) ([]string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	args := []string{"logs"}
	if lines > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", lines))
	}
	args = append(args, serviceName)

	cmd := exec.Command("docker-compose", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for service %s: %w", serviceName, err)
	}

	logs := strings.Split(string(output), "\n")
	// Удаляем пустые строки
	var filteredLogs []string
	for _, log := range logs {
		if strings.TrimSpace(log) != "" {
			filteredLogs = append(filteredLogs, log)
		}
	}

	return filteredLogs, nil
}

// isServiceRunning проверяет, запущен ли сервис
func (cm *ContainerManager) isServiceRunning(serviceName string) bool {
	cmd := exec.Command("docker-compose", "ps", "-q", serviceName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		return false
	}

	// Проверяем статус контейнера
	cmd = exec.Command("docker", "inspect", "--format", "{{.State.Running}}", containerID)
	output, err = cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) == "true"
}

// GetAllServicesStatus получает статус всех сервисов
func (cm *ContainerManager) GetAllServicesStatus() (map[string]*ContainerStatus, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	services := []string{"quic-server", "quic-client", "dashboard", "prometheus", "grafana"}
	statuses := make(map[string]*ContainerStatus)

	for _, service := range services {
		// Получаем имя контейнера для сервиса
		containerName := cm.getContainerNameForService(service)
		
		status, err := cm.GetContainerStatus(containerName)
		if err != nil {
			// Если контейнер не найден, создаем статус "не запущен"
			statuses[service] = &ContainerStatus{
				Name:   containerName,
				Status: "not_running",
				Health: "unknown",
			}
		} else {
			statuses[service] = status
		}
	}

	return statuses, nil
}

// getContainerNameForService возвращает имя контейнера для сервиса
func (cm *ContainerManager) getContainerNameForService(serviceName string) string {
	containerNames := map[string]string{
		"quic-server": "2gc-network-server",
		"quic-client": "2gc-network-client",
		"dashboard":   "2gc-network-dashboard",
		"prometheus":  "2gc-network-prometheus",
		"grafana":     "2gc-network-grafana",
	}

	if containerName, exists := containerNames[serviceName]; exists {
		return containerName
	}

	return serviceName
}

// ExecuteCommand выполняет команду в контейнере
func (cm *ContainerManager) ExecuteCommand(containerName string, command []string) (string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	args := append([]string{"exec", containerName}, command...)
	cmd := exec.Command("docker", args...)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute command in container %s: %w\nOutput: %s", containerName, err, string(output))
	}

	return string(output), nil
}

// GetMetricsFromContainer получает метрики из контейнера
func (cm *ContainerManager) GetMetricsFromContainer(containerName string, metricsPort string) (map[string]interface{}, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Выполняем curl внутри контейнера для получения метрик
	command := []string{"wget", "-qO-", fmt.Sprintf("http://localhost:%s/metrics", metricsPort)}
	
	output, err := cm.ExecuteCommand(containerName, command)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics from container %s: %w", containerName, err)
	}

	// Парсим Prometheus метрики (базовый парсинг)
	metrics := make(map[string]interface{})
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Простой парсинг метрик формата "metric_name value"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			metrics[parts[0]] = parts[1]
		}
	}

	return metrics, nil
}

// ScaleService изменяет количество реплик сервиса
func (cm *ContainerManager) ScaleService(serviceName string, replicas int) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cmd := exec.Command("docker-compose", "up", "-d", "--scale", fmt.Sprintf("%s=%d", serviceName, replicas))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to scale service %s to %d replicas: %w\nOutput: %s", serviceName, replicas, err, string(output))
	}

	return nil
}

// GetDockerComposeStatus получает статус всех сервисов docker-compose
func (cm *ContainerManager) GetDockerComposeStatus() (map[string]interface{}, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	cmd := exec.Command("docker-compose", "ps", "--format", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get docker-compose status: %w", err)
	}

	var services []map[string]interface{}
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		var service map[string]interface{}
		if err := json.Unmarshal([]byte(line), &service); err == nil {
			services = append(services, service)
		}
	}

	return map[string]interface{}{
		"services": services,
		"count":    len(services),
	}, nil
}

// WaitForService ждет, пока сервис не станет доступным
func (cm *ContainerManager) WaitForService(serviceName string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for service %s to become ready", serviceName)
		case <-ticker.C:
			if cm.isServiceRunning(serviceName) {
				return nil
			}
		}
	}
}