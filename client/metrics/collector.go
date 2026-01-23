package metrics

import (
	"sync"
	"time"
	"quic-test/internal/metrics"
)

// TimePoint представляет точку во времени для метрик
type TimePoint struct {
	Time  float64 `json:"Time"`  // seconds since start
	Value float64 `json:"Value"`
}

// TUIMetric представляет метрику для TUI дашборда
type TUIMetric struct {
	LatencyMs float64 `json:"latency_ms"`
	Code      int     `json:"code"`
	CPU       float64 `json:"cpu"`
	RTTMs     float64 `json:"rtt_ms"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// Collector хранит и собирает метрики теста
type Collector struct {
	mu         sync.Mutex
	Success    int
	Errors     int
	BytesSent  int
	Latencies  []float64
	Timestamps []time.Time
	Throughput []float64
	
	// Time series for latency and throughput
	TimeSeriesLatency    []TimePoint
	TimeSeriesThroughput []TimePoint

	// Advanced QUIC/TLS metrics
	PacketLoss             float64
	Retransmits            int
	HandshakeTimes         []float64
	TLSVersion             string
	CipherSuite            string
	SessionResumptionCount int
	ZeroRTTCount           int
	OneRTTCount            int
	OutOfOrderCount        int
	FlowControlEvents      int
	KeyUpdateEvents        int
	ErrorTypeCounts        map[string]int
	
	// Time series for new metrics
	TimeSeriesPacketLoss    []TimePoint
	TimeSeriesRetransmits   []TimePoint
	TimeSeriesHandshakeTime []TimePoint
	
	// HDR Histograms for precise metrics
	HDRMetrics *metrics.HDRMetrics
	
	// FEC Metrics
	FECPacketsSent    int64   `json:"fec_packets_sent"`
	FECRedundancyBytes int64   `json:"fec_redundancy_bytes"`
	FECRepairPacketsSent int64 `json:"fec_repair_sent"`
	FECRecovered       int64   `json:"fec_recovered"`
	FECRecoveryEvents  int64   `json:"fec_recovery_events"`
	FECUseCXX          bool    `json:"fec_use_cxx"`
	
	// PQC Metrics
	PQCHandshakeSize int64   `json:"pqc_handshake_size"`
	PQCHandshakeTime float64 `json:"pqc_handshake_time_ms"`
	PQCAlgorithm     string  `json:"pqc_algorithm"`
}

// NewCollector создает новый сборщик метрик
func NewCollector() *Collector {
	return &Collector{
		HDRMetrics: metrics.NewHDRMetrics(),
		ErrorTypeCounts: make(map[string]int),
	}
}

// RecordSuccess записывает успешную операцию
func (c *Collector) RecordSuccess(bytes int, latency float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.Success++
	c.BytesSent += bytes
	c.Latencies = append(c.Latencies, latency)
	c.Timestamps = append(c.Timestamps, time.Now())
	
	if c.HDRMetrics != nil {
		c.HDRMetrics.RecordLatency(time.Duration(latency * float64(time.Millisecond)))
		c.HDRMetrics.AddBytesSent(int64(bytes))
		c.HDRMetrics.IncrementPacketsSent()
	}
}

// RecordError записывает ошибку
func (c *Collector) RecordError(errorType string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.Errors++
	if c.ErrorTypeCounts == nil {
		c.ErrorTypeCounts = make(map[string]int)
	}
	c.ErrorTypeCounts[errorType]++
}

// RecordHandshake записывает время handshake
func (c *Collector) RecordHandshake(duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	handshakeMs := float64(duration.Nanoseconds()) / 1e6
	c.HandshakeTimes = append(c.HandshakeTimes, handshakeMs)
	c.TimeSeriesHandshakeTime = append(c.TimeSeriesHandshakeTime, 
		TimePoint{Time: time.Since(time.Now()).Seconds(), Value: handshakeMs})
	
	if c.HDRMetrics != nil {
		c.HDRMetrics.RecordHandshakeTime(duration)
	}
}

// UpdateTimeSeries обновляет временные ряды метрик
func (c *Collector) UpdateTimeSeries(startTime time.Time, lastCount, lastBytes int) (int, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	now := time.Since(startTime).Seconds()
	
	// Обновляем latency time series
	lat := 0.0
	if len(c.Latencies) > lastCount {
		sum := 0.0
		for _, l := range c.Latencies[lastCount:] {
			sum += l
		}
		lat = sum / float64(len(c.Latencies[lastCount:]))
	}
	c.TimeSeriesLatency = append(c.TimeSeriesLatency, TimePoint{Time: now, Value: lat})
	
	// Обновляем throughput time series
	bytesNow := c.BytesSent
	throughput := float64(bytesNow-lastBytes) / 1024.0
	c.TimeSeriesThroughput = append(c.TimeSeriesThroughput, TimePoint{Time: now, Value: throughput})
	
	return len(c.Latencies), bytesNow
}

// GetStats возвращает текущую статистику
func (c *Collector) GetStats() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	stats := map[string]interface{}{
		"success":     c.Success,
		"errors":      c.Errors,
		"bytes_sent":  c.BytesSent,
		"latencies":   len(c.Latencies),
		"retransmits": c.Retransmits,
		"packet_loss": c.PacketLoss,
	}
	
	if c.HDRMetrics != nil {
		stats["hdr_latency"] = c.HDRMetrics.GetLatencyStats()
		stats["hdr_jitter"] = c.HDRMetrics.GetJitterStats()
		stats["hdr_handshake"] = c.HDRMetrics.GetHandshakeStats()
		stats["hdr_throughput"] = c.HDRMetrics.GetThroughputStats()
		stats["hdr_network"] = c.HDRMetrics.GetNetworkStats()
	}
	
	return stats
}