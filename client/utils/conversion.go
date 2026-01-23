package utils

import (
	"math"
	"sort"
	"time"
	"quic-test/client/metrics"
)

// ToMap конвертирует метрики в map для совместимости с SLA проверками
func ToMap(m *metrics.Collector) map[string]interface{} {
	// Вычисляем средние значения
	var avgLatency float64
	if len(m.Latencies) > 0 {
		sum := 0.0
		for _, l := range m.Latencies {
			sum += l
		}
		avgLatency = sum / float64(len(m.Latencies))
	}

	var avgThroughput float64
	if len(m.Throughput) > 0 {
		sum := 0.0
		for _, t := range m.Throughput {
			sum += t
		}
		avgThroughput = sum / float64(len(m.Throughput))
	}

	// Вычисляем RTT процентили из Latencies (в миллисекундах)
	var rttP50, rttP95, rttP99 float64
	if len(m.Latencies) > 0 {
		rttP50, rttP95, rttP99 = calcPercentiles(m.Latencies)
	}

	// Вычисляем jitter (стандартное отклонение)
	jitter := calcJitter(m.Latencies)

	// Вычисляем throughput в Mbps (корректная формула: bytes * 8 / duration_seconds / 1e6)
	var throughputMbps float64
	var minRTT float64
	if len(m.Timestamps) > 0 {
		duration := time.Since(m.Timestamps[0]).Seconds()
		if duration > 0 {
			throughputMbps = (float64(m.BytesSent) * 8) / (duration * 1_000_000) // Bytes to Mbps
		}
		// Находим min RTT из latencies
		if len(m.Latencies) > 0 {
			minRTT = m.Latencies[0]
			for _, l := range m.Latencies {
				if l > 0 && l < minRTT {
					minRTT = l
				}
			}
		}
	}

	// Вычисляем goodput (исключая ретрансмиты)
	var goodputMbps float64
	if len(m.Timestamps) > 0 {
		duration := time.Since(m.Timestamps[0]).Seconds()
		if duration > 0 {
			// Приблизительно: вычитаем ретрансмиты из отправленных байт
			estimatedRetransBytes := int64(m.Retransmits) * 1200 // Примерный размер пакета
			goodputBytes := int64(m.BytesSent) - estimatedRetransBytes
			if goodputBytes < 0 {
				goodputBytes = 0
			}
			goodputMbps = (float64(goodputBytes) * 8) / (duration * 1_000_000)
		}
	}

	// Вычисляем bufferbloat factor: (avg_rtt / min_rtt) - 1
	var bufferbloatFactor float64
	if minRTT > 0 && avgLatency > 0 {
		bufferbloatFactor = (avgLatency / minRTT) - 1.0
		if bufferbloatFactor < 0 {
			bufferbloatFactor = 0
		}
	}

	// Вычисляем Fairness Index (Jain's index) для всех соединений
	var fairnessIndex float64
	if len(m.TimeSeriesThroughput) > 0 {
		var sum, sumSq float64
		for _, tp := range m.TimeSeriesThroughput {
			if tp.Value > 0 {
				sum += tp.Value
				sumSq += tp.Value * tp.Value
			}
		}
		if sum > 0 && sumSq > 0 {
			fairnessIndex = (sum * sum) / (float64(len(m.TimeSeriesThroughput)) * sumSq)
		}
	} else {
		// Если нет time series, используем вариацию latencies как proxy
		if len(m.Latencies) > 0 {
			var sum, sumSq float64
			for _, l := range m.Latencies {
				if l > 0 {
					sum += l
					sumSq += l * l
				}
			}
			if sum > 0 && sumSq > 0 {
				fairnessIndex = (sum * sum) / (float64(len(m.Latencies)) * sumSq)
			}
		}
	}

	// Вычисляем retransmission rate
	var retransmissionRate float64
	if m.Success > 0 {
		retransmissionRate = float64(m.Retransmits) / float64(m.Success)
	}

	result := map[string]interface{}{
		"Success":              m.Success,
		"Errors":               m.Errors,
		"BytesSent":            m.BytesSent,
		"Latencies":            m.Latencies,
		"ThroughputAverage":    avgThroughput,
		"ThroughputMbps":       throughputMbps,
		"GoodputMbps":          goodputMbps,
		"RetransmissionRate":   retransmissionRate,
		"RTTP50Ms":             rttP50,
		"RTTP95Ms":             rttP95,
		"RTTP99Ms":             rttP99,
		"RTTMinMs":             minRTT,
		"RTTAvgMs":             avgLatency,
		"JitterMs":             jitter,
		"PacketLoss":           m.PacketLoss,
		"Retransmits":          m.Retransmits,
		"BufferbloatFactor":    bufferbloatFactor,
		"FairnessIndex":        fairnessIndex,
		"TLSVersion":           m.TLSVersion,
		"CipherSuite":          m.CipherSuite,
		"SessionResumptionCount": m.SessionResumptionCount,
		"ZeroRTTCount":         m.ZeroRTTCount,
		"OneRTTCount":          m.OneRTTCount,
		"HandshakeTime":        avgLatency,
		"KeyUpdateEvents":      m.KeyUpdateEvents,
		"FlowControlEvents":    m.FlowControlEvents,
		"ErrorTypeCounts":      m.ErrorTypeCounts,
		"TimeSeriesLatency":    m.TimeSeriesLatency,
		"TimeSeriesThroughput": m.TimeSeriesThroughput,
		"TimeSeriesPacketLoss": m.TimeSeriesPacketLoss,
		"TimeSeriesRetransmits": m.TimeSeriesRetransmits,
		"TimeSeriesHandshakeTime": m.TimeSeriesHandshakeTime,
		"FECPacketsSent":       m.FECPacketsSent,
		"FECRedundancyBytes":   m.FECRedundancyBytes,
		"FECRepairPacketsSent": m.FECRepairPacketsSent,
		"FECRecovered":         m.FECRecovered,
		"FECRecoveryEvents":    m.FECRecoveryEvents,
		"PQCHandshakeSize":     m.PQCHandshakeSize,
		"PQCHandshakeTime":     m.PQCHandshakeTime,
		"PQCAlgorithm":         m.PQCAlgorithm,
	}

	// Добавляем HDR-метрики если доступны
	if m.HDRMetrics != nil {
		result["HDRLatencyStats"] = m.HDRMetrics.GetLatencyStats()
		result["HDRJitterStats"] = m.HDRMetrics.GetJitterStats()
		result["HDRHandshakeStats"] = m.HDRMetrics.GetHandshakeStats()
		result["HDRThroughputStats"] = m.HDRMetrics.GetThroughputStats()
		result["HDRNetworkStats"] = m.HDRMetrics.GetNetworkStats()
	}

	return result
}

// calcPercentiles вычисляет p50, p95, p99 для латенси
func calcPercentiles(latencies []float64) (p50, p95, p99 float64) {
	if len(latencies) == 0 {
		return 0, 0, 0
	}
	copyLat := make([]float64, len(latencies))
	copy(copyLat, latencies)
	sort.Float64s(copyLat)
	idx := func(p float64) int {
		return int(p*float64(len(copyLat)-1) + 0.5)
	}
	p50 = copyLat[idx(0.50)]
	p95 = copyLat[idx(0.95)]
	p99 = copyLat[idx(0.99)]
	return
}

// calcJitter вычисляет стандартное отклонение латенси (jitter)
func calcJitter(latencies []float64) float64 {
	if len(latencies) == 0 {
		return 0
	}
	mean := 0.0
	for _, l := range latencies {
		mean += l
	}
	mean /= float64(len(latencies))
	var sum float64
	for _, l := range latencies {
		d := l - mean
		sum += d * d
	}
	variance := sum / float64(len(latencies))
	// Извлекаем квадратный корень для получения стандартного отклонения
	jitter := math.Sqrt(variance)
	return jitter
}