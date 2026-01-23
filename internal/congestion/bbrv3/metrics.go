package bbrv3

import (
	"time"
)

// GetMetrics returns BBRv3-specific metrics for visualization
func (b *BBRv3) GetMetrics() BBRv3Metrics {
	b.metricsMux.Lock()
	defer b.metricsMux.Unlock()

	// Create a deep copy of metrics to avoid concurrent map access
	metricsCopy := b.metrics
	if b.metrics.PhaseDurationMs != nil {
		metricsCopy.PhaseDurationMs = make(map[string]float64)
		for k, v := range b.metrics.PhaseDurationMs {
			metricsCopy.PhaseDurationMs[k] = v
		}
	}
	return metricsCopy
}

// updateMetrics updates the metrics structure
func (b *BBRv3) updateMetrics() {
	b.metricsMux.Lock()
	defer b.metricsMux.Unlock()

	b.metrics.Phase = b.getStateString()
	b.metrics.BandwidthFast = b.bwFast
	b.metrics.BandwidthSlow = b.bwSlow
	b.metrics.Bandwidth = b.bw
	b.metrics.SendQuantum = b.sendQuantum
	b.metrics.PacingQuantum = b.sendQuantum // Alias
	b.metrics.InflightTarget = b.inflightTarget()
	b.metrics.LossRate = b.lossRateEMA
	b.metrics.LossRateEMA = b.lossRateEMA
	b.metrics.LossThreshold = b.params.LossThreshold
	b.metrics.CWND = b.cwnd
	b.metrics.PacingRate = b.pacingBps
	b.metrics.BDP = b.BDP()
	b.metrics.CurrentInflight = b.currentInflight
	
	if b.minRTT > 0 {
		b.metrics.MinRTT = float64(b.minRTT.Nanoseconds()) / 1e6
	}

	// Loss rate per round
	if b.roundTotal() > 0 {
		b.metrics.LossRateRound = float64(b.roundLost) / float64(b.roundTotal())
	}
	
	// Calculated metrics
	b.metrics.PacingGain = b.CalculatePacingGain()
	b.metrics.CWNDGain = b.CalculateCWNDGain()
	
	// Bufferbloat calculation
	if b.minRTT > 0 && len(b.recentRTTs) > 0 {
		var sum time.Duration
		count := 0
		for _, rtt := range b.recentRTTs {
			if rtt > 0 {
				sum += rtt
				count++
			}
		}
		if count > 0 {
			avgRTT := sum / time.Duration(count)
			b.metrics.Bufferbloat = b.calculateBufferbloat(avgRTT)
			b.metrics.BufferbloatFactor = b.CalculateBufferbloatFactor(avgRTT)
		}
	}
	
	// ProbeRTT minimum (update if in ProbeRTT state)
	if b.state == 3 && b.minRTT > 0 { // bbrv3ProbeRTT = 3
		b.metrics.ProbeRTTMinMs = float64(b.minRTT.Nanoseconds()) / 1e6
	}
	
	// Stability index (placeholder)
	b.metrics.StabilityIndex = 1.0
	
	// Loss recovery efficiency
	if b.packetsLost > 0 {
		b.metrics.LossRecoveryEfficiency = CalculateLossRecoveryEfficiency(b.recoveredPackets, b.packetsLost)
	} else {
		b.metrics.LossRecoveryEfficiency = 1.0
	}
	
	// Phase durations (convert to ms)
	b.updatePhaseDurations()
	
	// Recovery time
	b.updateRecoveryMetrics()
}

// calculateBufferbloat estimates bufferbloat based on RTT variance
func (b *BBRv3) calculateBufferbloat(avgRTT time.Duration) float64 {
	if b.minRTT <= 0 {
		return 0.0
	}
	
	// Bufferbloat is the excess delay beyond minimum RTT
	excess := avgRTT - b.minRTT
	if excess <= 0 {
		return 0.0
	}
	
	return float64(excess.Nanoseconds()) / 1e6 // Convert to milliseconds
}

// updatePhaseDurations updates phase duration metrics
func (b *BBRv3) updatePhaseDurations() {
	// Convert phase durations to milliseconds
	for phase, duration := range b.phaseDurations {
		switch phase {
		case "Startup":
			b.metrics.StartupDuration = duration
		case "Drain":
			b.metrics.DrainDuration = duration
		case "ProbeBW":
			b.metrics.ProbeBWDuration = duration
		case "ProbeRTT":
			b.metrics.ProbeRTTDuration = duration
		}
	}
	
	// Add current phase duration
	if startTime, ok := b.phaseStartTimes[b.state]; ok {
		currentDuration := time.Since(startTime)
		switch b.getStateString() {
		case "Startup":
			b.metrics.StartupDuration = currentDuration
		case "Drain":
			b.metrics.DrainDuration = currentDuration
		case "ProbeBW":
			b.metrics.ProbeBWDuration = currentDuration
		case "ProbeRTT":
			b.metrics.ProbeRTTDuration = currentDuration
		}
	}
}

// updateRecoveryMetrics updates loss recovery metrics
func (b *BBRv3) updateRecoveryMetrics() {
	// Recovery time calculation
	if !b.lastLossTime.IsZero() && !b.lastRecoveryTime.IsZero() && b.lastRecoveryTime.After(b.lastLossTime) {
		recoveryDuration := b.lastRecoveryTime.Sub(b.lastLossTime)
		b.metrics.RecoveryTimeMs = float64(recoveryDuration.Nanoseconds()) / 1e6
	}
}

// CalculatePacingGain calculates the current pacing gain
func (b *BBRv3) CalculatePacingGain() float64 {
	if b.bw <= 0 {
		return 1.0
	}
	
	return float64(b.pacingBps) / b.bw
}

// CalculateCWNDGain calculates the current CWND gain
func (b *BBRv3) CalculateCWNDGain() float64 {
	bdp := b.BDP()
	if bdp <= 0 {
		return 1.0
	}
	
	return float64(b.cwnd) / bdp
}

// CalculateBufferbloatFactor calculates bufferbloat factor
func (b *BBRv3) CalculateBufferbloatFactor(avgRTT time.Duration) float64 {
	if b.minRTT <= 0 {
		return 1.0
	}
	
	return float64(avgRTT) / float64(b.minRTT)
}

// CalculateLossRecoveryEfficiency calculates loss recovery efficiency
func CalculateLossRecoveryEfficiency(recoveredPackets, lostPackets int64) float64 {
	if lostPackets <= 0 {
		return 1.0
	}
	
	efficiency := float64(recoveredPackets) / float64(lostPackets)
	if efficiency > 1.0 {
		efficiency = 1.0
	}
	
	return efficiency
}

// GetDetailedMetrics returns detailed metrics for analysis
func (b *BBRv3) GetDetailedMetrics() map[string]interface{} {
	b.metricsMux.Lock()
	defer b.metricsMux.Unlock()
	
	return map[string]interface{}{
		"algorithm":           "BBRv3",
		"state":              b.getStateString(),
		"bandwidth_bps":      b.bw,
		"bandwidth_fast_bps": b.bwFast,
		"bandwidth_slow_bps": b.bwSlow,
		"min_rtt_ms":         float64(b.minRTT.Nanoseconds()) / 1e6,
		"cwnd_bytes":         b.cwnd,
		"pacing_rate_bps":    b.pacingBps,
		"bdp_bytes":          b.BDP(),
		"inflight_target":    b.inflightTarget(),
		"current_inflight":   b.currentInflight,
		"loss_rate_ema":      b.lossRateEMA,
		"loss_rate_round":    b.metrics.LossRateRound,
		"loss_threshold":     b.params.LossThreshold,
		"beta":               b.params.Beta,
		"headroom_fraction":  b.params.HeadroomFraction,
		"send_quantum":       b.sendQuantum,
		"packets_sent":       b.packetsSent,
		"packets_lost":       b.packetsLost,
		"round_acked":        b.roundAcked,
		"round_lost":         b.roundLost,
		"cycle_index":        b.cycleIdx,
		"phase_durations": map[string]float64{
			"startup_ms":    float64(b.metrics.StartupDuration.Nanoseconds()) / 1e6,
			"drain_ms":      float64(b.metrics.DrainDuration.Nanoseconds()) / 1e6,
			"probe_bw_ms":   float64(b.metrics.ProbeBWDuration.Nanoseconds()) / 1e6,
			"probe_rtt_ms":  float64(b.metrics.ProbeRTTDuration.Nanoseconds()) / 1e6,
		},
		"bufferbloat_ms":     b.metrics.Bufferbloat,
		"recovery_time_ms":   b.metrics.RecoveryTimeMs,
		"pacing_gain":        b.CalculatePacingGain(),
		"cwnd_gain":          b.CalculateCWNDGain(),
	}
}