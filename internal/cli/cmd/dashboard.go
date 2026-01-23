package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"go.uber.org/zap"
	"quic-test/internal/quic"
	"quic-test/internal/dashboard"
)

// DashboardHandler обрабатывает команду dashboard
type DashboardHandler struct {
	*BaseHandler
	logger        *zap.Logger
	quicAPI       *dashboard.QUICDashboardAPI
	educationalAPI *dashboard.EducationalAPI
}

// NewDashboardHandler создает новый обработчик команды dashboard
func NewDashboardHandler() *DashboardHandler {
	logger := CreateLogger()
	return &DashboardHandler{
		BaseHandler:    NewBaseHandler(NewGlobalContext()),
		logger:         logger,
		quicAPI:        dashboard.NewQUICDashboardAPI(logger),
		educationalAPI: dashboard.NewEducationalAPI(logger),
	}
}

// Run выполняет команду dashboard
func (h *DashboardHandler) Run(args []string) error {
	fmt.Println("Starting QUIC Testing Dashboard on http://localhost:9990")
	fmt.Println("Open your browser and navigate to http://localhost:9990")
	fmt.Println("Press Ctrl+C to stop the server")
	fmt.Println("Dashboard features:")
	fmt.Println("   - Start/Stop QUIC Server and Client")
	fmt.Println("   - Real-time metrics monitoring")
	fmt.Println("   - Test configuration and management")
	fmt.Println("   - API endpoints for automation")

	// Инициализируем QUIC менеджер
	if err := h.initializeQUICManager(); err != nil {
		return fmt.Errorf("failed to initialize QUIC manager: %w", err)
	}

	// Запускаем сбор метрик
	h.quicAPI.StartMetricsCollection()
	defer h.quicAPI.StopMetricsCollection()

	// Регистрируем обработчики
	h.registerHandlers()

	// Запускаем HTTP сервер
	return http.ListenAndServe(":9990", nil)
}

// initializeQUICManager инициализирует QUIC менеджер
func (h *DashboardHandler) initializeQUICManager() error {
	quicConfig := &quic.QUICManagerConfig{
		ServerAddr:     ":9001", // Уникальный порт для QUIC
		MaxConnections: 10,
		MaxStreams:     100,
		ConnectTimeout: 30 * time.Second,
		IdleTimeout:    60 * time.Second,
	}
	
	h.context.QUICManager = quic.NewQUICManager(h.logger, quicConfig)
	return nil
}

// registerHandlers регистрирует HTTP обработчики
func (h *DashboardHandler) registerHandlers() {
	// Статические файлы и основная страница
	http.HandleFunc("/", h.handleStatic)
	
	// Регистрируем QUIC API маршруты
	h.quicAPI.RegisterRoutes(http.DefaultServeMux)
	
	// Регистрируем образовательные API маршруты
	h.educationalAPI.RegisterRoutes(http.DefaultServeMux)
	
	// Дополнительные API endpoints для совместимости
	http.HandleFunc("/api/status", h.handleLegacyStatus)
	http.HandleFunc("/api/metrics", h.handleLegacyMetrics)
	http.HandleFunc("/api/config", h.handleLegacyConfig)
}

// handleStatic обрабатывает статические файлы
func (h *DashboardHandler) handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		// Возвращаем внешний HTML файл dashboard
		http.ServeFile(w, r, "web/templates/dashboard.html")
	} else if strings.HasPrefix(r.URL.Path, "/static/") {
		// Обрабатываем статические файлы
		http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))).ServeHTTP(w, r)
	} else {
		// Для остальных путей пытаемся найти файл
		http.ServeFile(w, r, r.URL.Path[1:])
	}
}

// handleLegacyStatus обрабатывает legacy запрос статуса
func (h *DashboardHandler) handleLegacyStatus(w http.ResponseWriter, r *http.Request) {
	processManager := h.quicAPI.GetProcessManager()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"server": map[string]interface{}{
			"running": processManager.IsServerRunning(),
		},
		"client": map[string]interface{}{
			"running": processManager.IsClientRunning(),
		},
		"last_update": time.Now(),
	})
}

// handleLegacyMetrics обрабатывает legacy запрос метрик
func (h *DashboardHandler) handleLegacyMetrics(w http.ResponseWriter, r *http.Request) {
	processManager := h.quicAPI.GetProcessManager()
	metrics := processManager.GetMetrics()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleLegacyConfig обрабатывает legacy запрос конфигурации
func (h *DashboardHandler) handleLegacyConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	config := map[string]interface{}{
		"server": map[string]interface{}{
			"addr": ":9000",
			"cert": "server.crt",
			"key":  "server.key",
		},
		"client": map[string]interface{}{
			"addr":        "localhost:9000",
			"connections": 1,
			"streams":     1,
			"packetSize":  1200,
			"rate":        100,
			"pattern":     "random",
		},
	}
	
	json.NewEncoder(w).Encode(config)
}

// handleStatus обрабатывает запрос статуса
func (h *DashboardHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if h.context.QUICManager != nil {
		json.NewEncoder(w).Encode(h.context.QUICManager.GetStatus())
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"server": map[string]interface{}{
				"running": false,
			},
			"client": map[string]interface{}{
				"running": false,
			},
		})
	}
}

// handleMetrics обрабатывает запрос метрик
func (h *DashboardHandler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	h.context.MetricsManager.UpdateMetrics()
	json.NewEncoder(w).Encode(h.context.MetricsManager.GetMetrics())
}

// handleConfig обрабатывает запрос конфигурации
func (h *DashboardHandler) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	config := map[string]interface{}{
		"server": map[string]interface{}{
			"addr": ":9001",
			"cert": "server.crt",
			"key":  "server.key",
		},
		"client": map[string]interface{}{
			"addr":        "localhost:9001",
			"connections": 1,
			"streams":     1,
			"packetSize":  1200,
			"rate":        100,
			"pattern":     "random",
		},
	}
	
	json.NewEncoder(w).Encode(config)
}

// handleServerStart обрабатывает запуск сервера
func (h *DashboardHandler) handleServerStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if h.context.QUICManager != nil {
		err := h.context.QUICManager.StartServer()
		if err != nil {
			h.sendJSONResponse(w, map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		
		h.sendJSONResponse(w, map[string]interface{}{
			"status":  "started",
			"message": "QUIC server started on port 9001",
		})
	} else {
		h.sendJSONResponse(w, map[string]interface{}{
			"status":  "error",
			"message": "QUIC manager not initialized",
		})
	}
}

// handleServerStop обрабатывает остановку сервера
func (h *DashboardHandler) handleServerStop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if h.context.QUICManager != nil {
		err := h.context.QUICManager.StopServer()
		if err != nil {
			h.sendJSONResponse(w, map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		
		h.sendJSONResponse(w, map[string]interface{}{
			"status":  "stopped",
			"message": "QUIC server stopped",
		})
	} else {
		h.sendJSONResponse(w, map[string]interface{}{
			"status":  "error",
			"message": "QUIC manager not initialized",
		})
	}
}

// handleClientStart обрабатывает запуск клиента
func (h *DashboardHandler) handleClientStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if h.context.QUICManager != nil {
		err := h.context.QUICManager.StartClient()
		if err != nil {
			h.sendJSONResponse(w, map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		
		h.sendJSONResponse(w, map[string]interface{}{
			"status":  "started",
			"message": "QUIC client connected to localhost:9001",
		})
	} else {
		h.sendJSONResponse(w, map[string]interface{}{
			"status":  "error",
			"message": "QUIC manager not initialized",
		})
	}
}

// handleClientStop обрабатывает остановку клиента
func (h *DashboardHandler) handleClientStop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if h.context.QUICManager != nil {
		err := h.context.QUICManager.StopClient()
		if err != nil {
			h.sendJSONResponse(w, map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		
		h.sendJSONResponse(w, map[string]interface{}{
			"status":  "stopped",
			"message": "QUIC client disconnected",
		})
	} else {
		h.sendJSONResponse(w, map[string]interface{}{
			"status":  "error",
			"message": "QUIC manager not initialized",
		})
	}
}

// handleTestStart обрабатывает запуск теста
func (h *DashboardHandler) handleTestStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if h.context.QUICManager == nil {
		h.sendJSONResponse(w, map[string]interface{}{
			"status":  "error",
			"message": "QUIC manager not initialized",
		})
		return
	}
	
	// Парсим параметры теста
	var testParams struct {
		PacketSize  int `json:"packet_size"`
		PacketCount int `json:"packet_count"`
		Duration    int `json:"duration"` // в секундах
	}
	
	if err := json.NewDecoder(r.Body).Decode(&testParams); err != nil {
		h.sendJSONResponse(w, map[string]interface{}{
			"status":  "error",
			"message": "Invalid test parameters",
		})
		return
	}
	
	// Запускаем тест
	// TODO: Реализовать запуск теста с параметрами
	h.sendJSONResponse(w, map[string]interface{}{
		"status":  "started",
		"message": "Test started with specified parameters",
		"params":  testParams,
	})
}

// sendJSONResponse отправляет JSON ответ
func (h *DashboardHandler) sendJSONResponse(w http.ResponseWriter, data interface{}) {
	json.NewEncoder(w).Encode(data)
}

// GetDashboardInfo возвращает информацию о dashboard
func (h *DashboardHandler) GetDashboardInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "QUIC Testing Dashboard",
		"description": "Веб-интерфейс для управления и мониторинга QUIC тестов",
		"features": []string{
			"Управление QUIC сервером и клиентом",
			"Мониторинг метрик в реальном времени",
			"Настройка параметров тестирования",
			"Визуализация результатов",
			"REST API для автоматизации",
		},
		"endpoints": []string{
			"/api/status",
			"/api/metrics",
			"/api/config",
			"/api/server/start",
			"/api/server/stop",
			"/api/client/start",
			"/api/client/stop",
			"/api/test/start",
		},
	}
}

// CreateLogger создает логгер
func CreateLogger() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	return logger
}