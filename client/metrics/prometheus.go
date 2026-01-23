package metrics

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusExporter экспортирует метрики в Prometheus
type PrometheusExporter struct {
	collector *Collector
	registry  *prometheus.Registry
}

// NewPrometheusExporter создает новый Prometheus экспортер
func NewPrometheusExporter(collector *Collector) *PrometheusExporter {
	return &PrometheusExporter{
		collector: collector,
		registry:  prometheus.NewRegistry(),
	}
}

// Start запускает Prometheus сервер
func (pe *PrometheusExporter) Start() {
	// Регистрируем метрики
	success := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_client_success_total",
		Help: "Total successful packets sent",
	}, func() float64 {
		pe.collector.mu.Lock()
		defer pe.collector.mu.Unlock()
		return float64(pe.collector.Success)
	})

	errors := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_client_errors_total",
		Help: "Total errors",
	}, func() float64 {
		pe.collector.mu.Lock()
		defer pe.collector.mu.Unlock()
		return float64(pe.collector.Errors)
	})

	bytesSent := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_client_bytes_sent",
		Help: "Total bytes sent",
	}, func() float64 {
		pe.collector.mu.Lock()
		defer pe.collector.mu.Unlock()
		return float64(pe.collector.BytesSent)
	})

	avgLatency := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_client_avg_latency_ms",
		Help: "Average latency in ms",
	}, func() float64 {
		pe.collector.mu.Lock()
		defer pe.collector.mu.Unlock()
		if len(pe.collector.Latencies) == 0 {
			return 0
		}
		sum := 0.0
		for _, l := range pe.collector.Latencies {
			sum += l
		}
		return sum / float64(len(pe.collector.Latencies))
	})

	throughput := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "quic_client_throughput_kbps",
		Help: "Current throughput in KB/s",
	}, func() float64 {
		pe.collector.mu.Lock()
		defer pe.collector.mu.Unlock()
		uptime := 0.0
		if len(pe.collector.Timestamps) > 0 {
			uptime = time.Since(pe.collector.Timestamps[0]).Seconds()
		}
		if uptime > 0 {
			return float64(pe.collector.BytesSent) / 1024.0 / uptime
		}
		return 0
	})

	prometheus.MustRegister(success, errors, bytesSent, avgLatency, throughput)

	// Создаем отдельный HTTP mux для клиента
	clientMux := http.NewServeMux()
	clientMux.Handle("/metrics", promhttp.Handler())

	fmt.Println("Prometheus endpoint клиента доступен на :2112/metrics")
	if err := http.ListenAndServe(":2112", clientMux); err != nil {
		log.Printf("Failed to start Prometheus server: %v", err)
	}
}