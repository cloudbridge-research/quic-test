package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"quic-test/internal"
)

func main() {
	// Парсинг флагов
	addr := flag.String("addr", ":9990", "Адрес для веб-интерфейса")
	flag.Parse()

	fmt.Println("==============================")
	fmt.Println("  2GC CloudBridge Dashboard")
	fmt.Println("==============================")

	// Обработка сигналов для graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("\nReceived termination signal, stopping dashboard...")
		os.Exit(0)
	}()

	startDashboard(*addr)
}

// startDashboard запускает веб-интерфейс
func startDashboard(addr string) {
	fmt.Println("Starting QUIC Testing Dashboard on http://localhost:9990")
	fmt.Println("Open your browser and navigate to http://localhost:9990")
	fmt.Println("Press Ctrl+C to stop the server")
	fmt.Println("Advanced analysis features available at:")
	fmt.Println("   - /api/status - Dashboard status")
	fmt.Println("   - /api/run - Start test")
	fmt.Println("   - /api/stop - Stop test")
	fmt.Println("   - /api/preset - Manage presets")
	fmt.Println("   - /api/report - Get reports")
	fmt.Println("   - /api/metrics - Real-time metrics")
	fmt.Println("   - /api/events - Server-Sent Events")

	// Создаем API и SSE менеджер
	api := internal.NewDashboardAPI()
	sseManager := internal.NewSSEManager()

	// Статические файлы
	http.Handle("/static/", http.StripPrefix("/static/", internal.StaticFileHandler()))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		internal.ServeStatic(w, r)
	})

	// API endpoints
	http.HandleFunc("/api/status", api.StatusHandler)
	http.HandleFunc("/api/run", api.RunTestHandler)
	http.HandleFunc("/api/stop", api.StopTestHandler)
	http.HandleFunc("/api/preset", api.PresetHandler)
	http.HandleFunc("/api/report", api.ReportHandler)
	http.HandleFunc("/api/metrics", api.MetricsHandler)
	
	// SSE endpoint
	http.HandleFunc("/api/events", sseManager.SSEServerHandler)

	// Запускаем сервер
	fmt.Printf("Dashboard started on http://localhost%s\n", addr)
	
	// Запускаем goroutine для периодической отправки метрик
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				// Отправляем текущие метрики через SSE
				state := api.GetState()
				sseManager.BroadcastMetrics(state.Metrics)
				sseManager.BroadcastStatus(state.ServerRunning, state.ClientRunning)
			}
		}
	}()
	
	log.Fatal(http.ListenAndServe(addr, nil))
}
