package server

import (
	"sync"
	"time"

	"quic-test/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// AdvancedPrometheusExporter provides advanced Prometheus metrics for the server
type AdvancedPrometheusExporter struct {
	// Basic metrics
	metrics *metrics.PrometheusMetrics

	// Additional server metrics
	serverMetrics *ServerMetrics

	// Request type counters
	requestTypeCounters *prometheus.CounterVec

	// Request processing histograms
	requestProcessingHistograms *prometheus.HistogramVec

	// Connection metrics
	connectionMetrics *prometheus.GaugeVec

	// Stream metrics
	streamMetrics *prometheus.GaugeVec

	// Data processing metrics
	dataProcessingMetrics *prometheus.CounterVec

	mu sync.RWMutex
}

// ServerMetrics contains server metrics
type ServerMetrics struct {
	ServerAddr         string
	MaxConnections     int
	CurrentConnections int
	CurrentStreams     int
	StartTime          time.Time
	LastUpdate         time.Time
	Uptime             time.Duration
}

// NewAdvancedPrometheusExporter creates a new metrics exporter for the server
func NewAdvancedPrometheusExporter(serverAddr string) *AdvancedPrometheusExporter {
	return &AdvancedPrometheusExporter{
		metrics: metrics.NewPrometheusMetrics(prometheus.DefaultRegisterer),
		serverMetrics: &ServerMetrics{
			ServerAddr: serverAddr,
			StartTime:  time.Now(),
		},
		requestTypeCounters: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "quic_server_request_type_total",
			Help: "Total requests by type",
		}, []string{"request_type", "connection_id", "stream_id", "result"}),
		requestProcessingHistograms: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "quic_server_request_processing_duration_seconds",
			Help:    "Request processing duration",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		}, []string{"request_type", "connection_id", "result"}),
		connectionMetrics: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "quic_server_connection_info",
			Help: "Server connection information",
		}, []string{"connection_id", "remote_addr", "tls_version", "cipher_suite", "state"}),
		streamMetrics: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "quic_server_stream_info",
			Help: "Server stream information",
		}, []string{"stream_id", "connection_id", "stream_type", "state", "direction"}),
		dataProcessingMetrics: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "quic_server_data_processing_total",
			Help: "Data processing metrics",
		}, []string{"operation", "connection_id", "stream_id", "data_type"}),
	}
}

// UpdateServerInfo updates server information
func (ape *AdvancedPrometheusExporter) UpdateServerInfo(maxConnections int) {
	ape.mu.Lock()
	defer ape.mu.Unlock()

	ape.serverMetrics.MaxConnections = maxConnections
	ape.serverMetrics.LastUpdate = time.Now()
	ape.serverMetrics.Uptime = time.Since(ape.serverMetrics.StartTime)
}

// RecordRequestProcessing records request processing
func (ape *AdvancedPrometheusExporter) RecordRequestProcessing(requestType, connectionID string, duration time.Duration, result string) {
	// Record in basic metrics
		ape.metrics.RecordScenarioDuration(duration)

	// Record in server-specific metrics
	ape.requestTypeCounters.WithLabelValues(requestType, connectionID, "", result).Inc()
	ape.requestProcessingHistograms.WithLabelValues(requestType, connectionID, result).Observe(duration.Seconds())
}

// RecordConnectionInfo records connection information
func (ape *AdvancedPrometheusExporter) RecordConnectionInfo(connectionID, remoteAddr, tlsVersion, cipherSuite, state string) {
	ape.connectionMetrics.WithLabelValues(connectionID, remoteAddr, tlsVersion, cipherSuite, state).Set(1)
}

// RecordStreamInfo records stream information
func (ape *AdvancedPrometheusExporter) RecordStreamInfo(streamID, connectionID, streamType, state, direction string) {
	ape.streamMetrics.WithLabelValues(streamID, connectionID, streamType, state, direction).Set(1)
}

// RecordDataProcessing records data processing
func (ape *AdvancedPrometheusExporter) RecordDataProcessing(operation, connectionID, streamID, dataType string, bytes int64) {
	ape.dataProcessingMetrics.WithLabelValues(operation, connectionID, streamID, dataType).Add(float64(bytes))
}

// RecordLatency records latency
func (ape *AdvancedPrometheusExporter) RecordLatency(latency time.Duration) {
	ape.metrics.RecordLatency(latency)
}

// RecordJitter records jitter
func (ape *AdvancedPrometheusExporter) RecordJitter(jitter time.Duration) {
	ape.metrics.RecordJitter(jitter)
}

// RecordThroughput records throughput
func (ape *AdvancedPrometheusExporter) RecordThroughput(throughput float64) {
	ape.metrics.RecordThroughput(int64(throughput))
}

// RecordHandshakeTime records handshake time
func (ape *AdvancedPrometheusExporter) RecordHandshakeTime(duration time.Duration) {
	ape.metrics.RecordHandshakeTime(duration)
}

// RecordRTT records RTT
func (ape *AdvancedPrometheusExporter) RecordRTT(rtt time.Duration) {
	ape.metrics.RecordRTT(rtt)
}

// IncrementConnections increments connection counter
func (ape *AdvancedPrometheusExporter) IncrementConnections() {
	ape.metrics.IncrementConnections()
	ape.mu.Lock()
	ape.serverMetrics.CurrentConnections++
	ape.mu.Unlock()
}

// DecrementConnections decrements connection counter
func (ape *AdvancedPrometheusExporter) DecrementConnections() {
	ape.metrics.DecrementConnections()
	ape.mu.Lock()
	ape.serverMetrics.CurrentConnections--
	ape.mu.Unlock()
}

// IncrementStreams increments stream counter
func (ape *AdvancedPrometheusExporter) IncrementStreams() {
	ape.metrics.IncrementStreams()
	ape.mu.Lock()
	ape.serverMetrics.CurrentStreams++
	ape.mu.Unlock()
}

// DecrementStreams decrements stream counter
func (ape *AdvancedPrometheusExporter) DecrementStreams() {
	ape.metrics.DecrementStreams()
	ape.mu.Lock()
	ape.serverMetrics.CurrentStreams--
	ape.mu.Unlock()
}

// AddBytesSent adds sent bytes
func (ape *AdvancedPrometheusExporter) AddBytesSent(bytes int64) {
	ape.metrics.AddBytesSent(bytes)
}

// AddBytesReceived adds received bytes
func (ape *AdvancedPrometheusExporter) AddBytesReceived(bytes int64) {
	ape.metrics.AddBytesReceived(bytes)
}

// IncrementErrors increments error counter
func (ape *AdvancedPrometheusExporter) IncrementErrors() {
	ape.metrics.IncrementErrors()
}

// IncrementRetransmits increments retransmission counter
func (ape *AdvancedPrometheusExporter) IncrementRetransmits() {
	ape.metrics.IncrementRetransmits()
}

// IncrementHandshakes increments handshake counter
func (ape *AdvancedPrometheusExporter) IncrementHandshakes() {
	ape.metrics.IncrementHandshakes()
}

// IncrementZeroRTT increments 0-RTT counter
func (ape *AdvancedPrometheusExporter) IncrementZeroRTT() {
	ape.metrics.IncrementZeroRTT()
}

// IncrementOneRTT increments 1-RTT counter
func (ape *AdvancedPrometheusExporter) IncrementOneRTT() {
	ape.metrics.IncrementOneRTT()
}

// IncrementSessionResumptions increments session resumption counter
func (ape *AdvancedPrometheusExporter) IncrementSessionResumptions() {
	ape.metrics.IncrementSessionResumptions()
}

// SetCurrentThroughput sets current throughput
func (ape *AdvancedPrometheusExporter) SetCurrentThroughput(throughput float64) {
	ape.metrics.SetCurrentThroughput(int64(throughput))
}

// SetCurrentLatency sets current latency
func (ape *AdvancedPrometheusExporter) SetCurrentLatency(latency time.Duration) {
	ape.metrics.SetCurrentLatency(latency)
}

// SetPacketLossRate sets packet loss rate
func (ape *AdvancedPrometheusExporter) SetPacketLossRate(rate float64) {
	ape.metrics.SetPacketLossRate(rate)
}

// SetConnectionDuration sets connection duration
func (ape *AdvancedPrometheusExporter) SetConnectionDuration(duration time.Duration) {
	ape.metrics.SetConnectionDuration(duration)
}

// RecordScenarioEvent records scenario event
func (ape *AdvancedPrometheusExporter) RecordScenarioEvent(scenario, connectionID, streamID, result string) {
	ape.metrics.RecordScenarioEvent(scenario)
}

// RecordErrorEvent records error event
func (ape *AdvancedPrometheusExporter) RecordErrorEvent(errorType, connectionID, streamID, severity string) {
	ape.metrics.RecordErrorEvent(errorType)
}

// RecordProtocolEvent records protocol event
func (ape *AdvancedPrometheusExporter) RecordProtocolEvent(eventType, connectionID, tlsVersion, cipherSuite string) {
	ape.metrics.RecordProtocolEvent(eventType)
}

// RecordNetworkLatency records network latency by profile
func (ape *AdvancedPrometheusExporter) RecordNetworkLatency(networkProfile, connectionID, region string, latency time.Duration) {
	ape.metrics.RecordNetworkLatency(latency)
}

// GetServerMetrics returns current server metrics
func (ape *AdvancedPrometheusExporter) GetServerMetrics() *ServerMetrics {
	ape.mu.RLock()
	defer ape.mu.RUnlock()

	return &ServerMetrics{
		ServerAddr:         ape.serverMetrics.ServerAddr,
		MaxConnections:     ape.serverMetrics.MaxConnections,
		CurrentConnections: ape.serverMetrics.CurrentConnections,
		CurrentStreams:     ape.serverMetrics.CurrentStreams,
		StartTime:          ape.serverMetrics.StartTime,
		LastUpdate:         ape.serverMetrics.LastUpdate,
		Uptime:             time.Since(ape.serverMetrics.StartTime),
	}
}
