package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"quic-test/client/connection"
	"quic-test/client/metrics"
	"quic-test/client/utils"
	"quic-test/internal"
	"quic-test/internal/ai"
	"quic-test/internal/integration"
	"quic-test/internal/pqc"

	"go.uber.org/zap"
)

// Run запускает клиентский тест
func Run(cfg internal.TestConfig) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\nReceived termination signal, generating report...")
		cancel()
	}()

	// Создаем сборщик метрик
	testMetrics := metrics.NewCollector()
	var wg sync.WaitGroup

	// Запускаем Prometheus экспортер если включен
	if cfg.Prometheus {
		go func() {
			exporter := metrics.NewPrometheusExporter(testMetrics)
			exporter.Start()
		}()
	}

	// Создаем и регистрируем глобальный SimpleIntegration для BBRv3
	var globalSI *integration.SimpleIntegration
	if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
		logger, _ := zap.NewDevelopment()
		globalSI = integration.NewSimpleIntegration(logger, cfg.CongestionControl)
		if err := globalSI.Initialize(); err != nil {
			fmt.Printf("Warning: Failed to initialize global %s integration: %v\n", cfg.CongestionControl, err)
			globalSI = nil
		} else {
			gmc := internal.GetGlobalMetricsCollector()
			gmc.SetExperimentalIntegration(globalSI)
			fmt.Printf("[INFO] Global BBRv3 integration registered in GlobalMetricsCollector\n")
		}
	}

	// AI Prediction Consumer
	if cfg.AIEnabled {
		startAIPredictionConsumer(ctx, cfg, testMetrics)
	}

	startTime := time.Now()
	
	// Time series collector
	go startTimeSeriesCollector(ctx, testMetrics, startTime)

	// Ramp-up/ramp-down сценарий
	var rate int64 = int64(cfg.Rate)
	go startRampUpDown(cfg, &rate)

	// Запускаем соединения
	for c := 0; c < cfg.Connections; c++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()
			runConnection(ctx, cfg, testMetrics, connID, &rate, globalSI)
		}(c)
	}

	// Ждем завершения с таймаутом
	waitForCompletion(ctx, cfg, &wg)

	// Обработка результатов
	processResults(cfg, testMetrics)
}

// runConnection запускает одно QUIC соединение
func runConnection(ctx context.Context, cfg internal.TestConfig, collector *metrics.Collector, connID int, ratePtr *int64, si *integration.SimpleIntegration) {
	if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
		fmt.Printf("[DEBUG] Connection %d: started\n", connID)
	}
	defer func() {
		if cfg.CongestionControl == "bbrv3" || cfg.CongestionControl == "bbrv2" {
			fmt.Printf("[DEBUG] Connection %d: returning\n", connID)
		}
	}()

	// Создаем менеджер соединений
	connManager := connection.NewManager(cfg)
	defer connManager.Close()

	// Устанавливаем соединение
	handshakeStart := time.Now()
	
	// PQC симуляция
	if cfg.PQCEnabled && cfg.PQCAlgorithm != "" {
		simulatePQC(cfg, collector)
	}

	session, err := connManager.Connect(ctx, si)
	if err != nil {
		collector.RecordError("quic_handshake")
		fmt.Printf("Connection %d: failed to connect: %v\n", connID, err)
		return
	}

	handshakeTime := time.Since(handshakeStart)
	collector.RecordHandshake(handshakeTime)

	// Записываем TLS информацию
	state := connManager.GetConnectionState()
	collector.TLSVersion = tlsVersionString(state.TLS.Version)
	collector.CipherSuite = cipherSuiteString(state.TLS.CipherSuite)
	if state.TLS.DidResume {
		collector.SessionResumptionCount++
	}
	if state.Used0RTT {
		collector.ZeroRTTCount++
	} else {
		collector.OneRTTCount++
	}

	// Запускаем стримы
	var streamWG sync.WaitGroup
	for s := 0; s < cfg.Streams; s++ {
		streamWG.Add(1)
		go func(streamID int) {
			defer streamWG.Done()
			handler := connection.NewStreamHandler(cfg, collector, connID, streamID, ratePtr, si)
			if err := handler.HandleStream(ctx, session); err != nil {
				fmt.Printf("Connection %d, Stream %d: error: %v\n", connID, streamID, err)
			}
		}(s)
	}

	// Ждем завершения стримов с таймаутом
	waitForStreams(ctx, cfg, &streamWG, connID)
}

// startAIPredictionConsumer запускает AI prediction consumer
func startAIPredictionConsumer(ctx context.Context, cfg internal.TestConfig, collector *metrics.Collector) {
	aiClient := ai.NewPredictionClient(cfg.AIServiceURL)
	fmt.Printf("[INFO] AI Routing enabled. Connecting to %s\n", cfg.AIServiceURL)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Collect current metrics for features
				metricsMap := utils.ToMap(collector)
				rtt := metricsMap["RTTP95Ms"].(float64)
				jitter := metricsMap["JitterMs"].(float64)
				loss := metricsMap["PacketLoss"].(float64)
				throughput := metricsMap["ThroughputMbps"].(float64)

				// Feature vector: [rtt, jitter, loss, throughput]
				features := []float64{rtt, jitter, loss, throughput}

				// Request prediction for current route
				pred, err := aiClient.GetPrediction("route-0", features)
				if err != nil {
					continue
				}

				// Log prediction result
				if pred.ConfidenceScore > 0.8 {
					fmt.Printf("[AI] Prediction: Latency=%.2fms, Jitter=%.2fms (Confidence: %.2f)\n",
						pred.PredictedLatencyMs, pred.PredictedJitterMs, pred.ConfidenceScore)

					if pred.PredictedLatencyMs > 100 {
						fmt.Printf("[AI] High latency predicted! Recommending route switch...\n")
					}
				}
			}
		}
	}()
}

// startTimeSeriesCollector запускает сборщик временных рядов
func startTimeSeriesCollector(ctx context.Context, collector *metrics.Collector, startTime time.Time) {
	var lastCount int
	var lastBytes int
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
			lastCount, lastBytes = collector.UpdateTimeSeries(startTime, lastCount, lastBytes)
			
			// Периодическая отправка метрик в QUIC Bottom
			metricsMap := utils.ToMap(collector)
			metricsMap = internal.EnhanceMetricsMap(metricsMap)
			internal.UpdateBottomMetrics(metricsMap)
		}
	}
}

// startRampUpDown запускает ramp-up/ramp-down сценарий
func startRampUpDown(cfg internal.TestConfig, rate *int64) {
	minRate := int64(1)
	maxRate := int64(cfg.Rate)
	if maxRate < 10 {
		maxRate = 100
	}
	step := (maxRate - minRate) / 10
	if step < 1 {
		step = 1
	}
	
	for {
		// Ramp-up
		for r := minRate; r <= maxRate; r += step {
			atomic.StoreInt64(rate, r)
			time.Sleep(1 * time.Second)
		}
		// Ramp-down
		for r := maxRate; r >= minRate; r -= step {
			atomic.StoreInt64(rate, r)
			time.Sleep(1 * time.Second)
		}
	}
}

// simulatePQC симулирует PQC overhead
func simulatePQC(cfg internal.TestConfig, collector *metrics.Collector) {
	pqcSim := pqc.NewPQCSimulator(cfg.PQCAlgorithm)
	pqcOverhead, pqcSize := pqcSim.SimulateHandshake()

	// Добавляем PQC overhead к handshake времени
	time.Sleep(pqcOverhead)

	collector.PQCHandshakeSize = int64(pqcSize)
	collector.PQCHandshakeTime = float64(pqcOverhead.Nanoseconds()) / 1e6
	collector.PQCAlgorithm = cfg.PQCAlgorithm
}

// waitForStreams ждет завершения стримов с таймаутом
func waitForStreams(ctx context.Context, cfg internal.TestConfig, wg *sync.WaitGroup, connID int) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	streamTimeout := cfg.Duration + 10*time.Second
	if cfg.Duration == 0 {
		streamTimeout = 70 * time.Second
	}

	select {
	case <-done:
		// Все стримы завершились
	case <-ctx.Done():
		// Контекст отменен
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			fmt.Printf("[WARNING] Connection %d: Some streams didn't finish after context cancel\n", connID)
		}
	case <-time.After(streamTimeout):
		// Таймаут
		fmt.Printf("[WARNING] Connection %d streams timeout after %v\n", connID, streamTimeout)
		select {
		case <-done:
		case <-time.After(1 * time.Second):
		}
	}
}

// waitForCompletion ждет завершения всех соединений
func waitForCompletion(ctx context.Context, cfg internal.TestConfig, wg *sync.WaitGroup) {
	if cfg.Duration > 0 {
		timer := time.NewTimer(cfg.Duration)
		go func() {
			<-timer.C
			fmt.Println("\nTest completed by timer, generating report...")
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	timeout := cfg.Duration + 10*time.Second
	if cfg.Duration == 0 {
		timeout = 120 * time.Second
	}

	select {
	case <-done:
		// Все соединения завершились
	case <-time.After(timeout):
		fmt.Printf("\n⚠️  Таймаут ожидания завершения (%v). Завершаем принудительно...\n", timeout)
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			fmt.Println("⚠️  Некоторые горутины не завершились, продолжаем...")
		}
	}
}

// processResults обрабатывает результаты теста
func processResults(cfg internal.TestConfig, collector *metrics.Collector) {
	fmt.Printf("\nTest completed. Processing results...\n")

	// Отправляем метрики в QUIC Bottom
	metricsMap := utils.ToMap(collector)
	metricsMap = internal.EnhanceMetricsMap(metricsMap)

	// Базовый вывод для контроля
	if bbrv3Metrics, ok := metricsMap["BBRv3Metrics"].(map[string]interface{}); ok {
		fmt.Printf("BBRv3 Phase: %v, BW: %.2f Mbps\n",
			bbrv3Metrics["phase"],
			bbrv3Metrics["bw"].(float64)/1_000_000)
	}

	internal.UpdateBottomMetrics(metricsMap)

	// Не сохраняем отчет - только выводим основные метрики
	fmt.Printf("Success: %d, Errors: %d, Bytes sent: %d\n", 
		collector.Success, collector.Errors, collector.BytesSent)
	
	if len(collector.Latencies) > 0 {
		sum := 0.0
		for _, l := range collector.Latencies {
			sum += l
		}
		avgLatency := sum / float64(len(collector.Latencies))
		fmt.Printf("Average latency: %.2f ms\n", avgLatency)
	}

	// Проверяем SLA если настроено
	if cfg.SlaRttP95 > 0 || cfg.SlaLoss > 0 || cfg.SlaThroughput > 0 || cfg.SlaErrors > 0 {
		internal.ExitWithSLA(cfg, metricsMap)
	}
}

// Вспомогательные функции для TLSVersion/CipherSuite
func tlsVersionString(v uint16) string {
	switch v {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:
		return "TLS 1.2"
	default:
		return fmt.Sprintf("0x%x", v)
	}
}

func cipherSuiteString(cs uint16) string {
	switch cs {
	case tls.TLS_AES_128_GCM_SHA256:
		return "TLS_AES_128_GCM_SHA256"
	case tls.TLS_AES_256_GCM_SHA384:
		return "TLS_AES_256_GCM_SHA384"
	case tls.TLS_CHACHA20_POLY1305_SHA256:
		return "TLS_CHACHA20_POLY1305_SHA256"
	default:
		return fmt.Sprintf("0x%x", cs)
	}
}

