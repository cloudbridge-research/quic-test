package bbrv3

import (
	"fmt"
	"time"
)

// OnAck handles an ACK event
func (b *BBRv3) OnAck(s Sample) (cwnd int, pacing int64) {
	defer func() {
		if r := recover(); r != nil {
			// Log panic but don't crash - return safe defaults
			b.logQlogEvent("panic", map[string]interface{}{
				"error": fmt.Sprintf("%v", r),
				"state": b.getStateString(),
			})
			// Return last known good values
			cwnd = b.cwnd
			pacing = b.pacingBps
		}
	}()
	
	now := time.Now()
	oldState := b.state
	oldCWND := b.cwnd
	oldPacing := b.pacingBps
	
	// Track recent RTTs for bufferbloat calculation
	if s.RTT > 0 {
		b.recentRTTs[b.recentRTTIdx%len(b.recentRTTs)] = s.RTT
		b.recentRTTIdx++
		b.lastRTT = s.RTT
	}
	
	// Update min RTT
	if s.RTT > 0 && (b.minRTT == 0 || s.RTT < b.minRTT) {
		oldMinRTT := b.minRTT
		b.minRTT = s.RTT
		b.minRTTSince = now
		
		b.logQlogEvent("rtt_update", map[string]interface{}{
			"old_rtt":      float64(oldMinRTT.Nanoseconds()) / 1e6,
			"new_rtt":      float64(s.RTT.Nanoseconds()) / 1e6,
			"min_rtt":      float64(b.minRTT.Nanoseconds()) / 1e6,
			"rtt_variance": 0.0,
			"smoothed_rtt": float64(s.RTT.Nanoseconds()) / 1e6,
		})
	}
	
	// Accumulate round acked bytes
	b.roundAcked += s.RS.BytesAcked
	
	// Update dual-scale bandwidth estimates
	if br := s.RS.BandwidthBps(); br > 0 && !s.RS.IsAppLimited {
		// Fast-scale: immediate samples (max) with decay for stability
		if br > b.bwFast {
			b.bwFast = br
		} else {
			// Slow decay: 99.5% per sample (allows tracking decreases)
			b.bwFast = 0.995 * b.bwFast
		}
		
		// Slow-scale: exponential moving average with adaptive alpha
		if b.bwSlow == 0 {
			b.bwSlow = br
		} else {
			// EMA with adaptive alpha: more responsive when bandwidth is changing
			ratio := b.bwFast / b.bwSlow
			alpha := 0.1 // Default
			if ratio > 1.1 || ratio < 0.9 {
				// Bandwidth is changing, be more responsive
				alpha = 0.15
			}
			b.bwSlow = (1-alpha)*b.bwSlow + alpha*br
		}
		
		// Current bandwidth is max of fast/slow
		b.bw = maxF(b.bwFast, b.bwSlow)
		
		b.logQlogEvent("bandwidth_sample", map[string]interface{}{
			"sample_bandwidth": br,
			"bw_fast":          b.bwFast,
			"bw_slow":          b.bwSlow,
			"bandwidth":        b.bw,
			"bytes_acked":      s.RS.BytesAcked,
			"rtt":              float64(s.RTT.Nanoseconds()) / 1e6,
		})
	}
	
	// Update pacing quantum
	b.updatePacingQuantum()
	
	// State machine
	b.processStateMachine(now)
	
	// Loss threshold check by round
	b.checkLossThreshold(oldCWND)
	
	// Update loss rate EMA for metrics
	b.updateLossRateEMA()
	
	// Track phase durations when state changes
	if oldState != b.state {
		b.trackPhaseTransition(oldState, now)
	}
	
	// Ensure minimum cwnd and set pacing rate
	b.finalizeCwndAndPacing()
	
	// Update metrics
	b.updateHeadroomUsage()
	
	// Log changes
	b.logChanges(oldState, oldCWND, oldPacing)
	
	return b.cwnd, b.pacingBps
}

// processStateMachine handles the BBRv3 state machine logic
func (b *BBRv3) processStateMachine(now time.Time) {
	switch b.state {
	case bbrv3Startup:
		b.handleStartupState(now)
	case bbrv3Drain:
		b.handleDrainState(now)
	case bbrv3ProbeBW:
		b.handleProbeBWState(now)
	case bbrv3ProbeRTT:
		b.handleProbeRTTState(now)
	}
}

// handleStartupState processes the Startup state
func (b *BBRv3) handleStartupState(now time.Time) {
	// Aggressive growth with Startup pacing gain 2.77
	b.cwnd += max(1, int(b.roundAcked))
	b.pacingBps = int64(b.params.StartupPacingGain * b.bw)
	
	// Transition to Drain after stable bandwidth or timeout
	startupTimeout := 2 * time.Second
	if b.fullPipeDetected() {
		startupTimeout = 1 * time.Second
	}
	if now.Sub(b.lastStateTs) > startupTimeout || b.fullPipeDetected() {
		b.state = bbrv3Drain
		b.lastStateTs = now
		b.logQlogEvent("state_transition", map[string]interface{}{
			"from":   "Startup",
			"to":     "Drain",
			"reason": "timeout_or_full_pipe",
		})
	}
}

// handleDrainState processes the Drain state
func (b *BBRv3) handleDrainState(now time.Time) {
	// Drain excess queued data with Drain pacing gain 0.35
	inflightTarget := b.inflightTarget()
	b.cwnd = int(inflightTarget)
	b.pacingBps = int64(b.params.DrainPacingGain * b.bw)
	
	// Transition to ProbeBW after draining
	probePeriod := maxDur(200*time.Millisecond, 2*b.minRTT)
	if now.Sub(b.lastStateTs) > probePeriod {
		b.state = bbrv3ProbeBW
		b.cycleIdx = 0
		b.lastStateTs = now
		b.logQlogEvent("state_transition", map[string]interface{}{
			"from":   "Drain",
			"to":     "ProbeBW",
			"reason": "timeout",
		})
	}
}

// handleProbeBWState processes the ProbeBW state
func (b *BBRv3) handleProbeBWState(now time.Time) {
	// ProbeBW cycle with adaptive gains
	gains := []float64{1.25, 1.0, 0.75, 1.0}
	g := gains[b.cycleIdx%len(gains)]
	
	// Adaptive gain adjustment based on loss rate
	if b.metrics.LossRate < 0.01 && b.roundTotal() > 0 {
		if g == 1.25 {
			g = 1.28 // Slightly more aggressive probe up
		}
	}
	
	// Inflight target with headroom reserved
	inflightTarget := g * b.inflightTarget()
	b.cwnd = max(int(inflightTarget), 4*b.mtu)
	b.pacingBps = int64(g * b.bw)
	
	// ProbeBW period depends on RTT
	probePeriod := maxDur(200*time.Millisecond, 2*b.minRTT)
	if now.Sub(b.lastStateTs) > probePeriod {
		b.cycleIdx++
		b.lastStateTs = now
	}
	
	// Check for ProbeRTT condition
	if b.minRTT > 0 && now.Sub(b.minRTTSince) > 10*time.Second {
		b.state = bbrv3ProbeRTT
		b.lastStateTs = now
		b.logQlogEvent("state_transition", map[string]interface{}{
			"from":   "ProbeBW",
			"to":     "ProbeRTT",
			"reason": "min_rtt_stale",
		})
	}
}

// handleProbeRTTState processes the ProbeRTT state
func (b *BBRv3) handleProbeRTTState(now time.Time) {
	// ProbeRTT with reduced cwnd
	target := 0.5 * b.BDP()
	b.cwnd = max(int(target), 4*b.mtu)
	b.pacingBps = int64(0.5 * b.bw)
	
	// Return to ProbeBW after ProbeRTTDuration
	if now.Sub(b.lastStateTs) > b.params.ProbeRTTDuration {
		b.minRTTSince = now
		b.state = bbrv3ProbeBW
		b.cycleIdx = 0
		b.lastStateTs = now
		b.logQlogEvent("state_transition", map[string]interface{}{
			"from":   "ProbeRTT",
			"to":     "ProbeBW",
			"reason": "probe_rtt_complete",
		})
	}
}

// checkLossThreshold checks if loss threshold is exceeded and reduces cwnd
func (b *BBRv3) checkLossThreshold(oldCWND int) {
	if b.roundTotal() > 0 {
		lossRateRound := float64(b.roundLost) / float64(b.roundTotal())
		if lossRateRound > b.params.LossThreshold {
			b.cwnd = max(int(b.params.Beta*float64(b.cwnd)), 2*b.mtu)
			b.resetRound()
			b.logQlogEvent("loss_threshold_exceeded", map[string]interface{}{
				"loss_rate_round": lossRateRound,
				"threshold":       b.params.LossThreshold,
				"round_acked":     b.roundAcked,
				"round_lost":      b.roundLost,
				"old_cwnd":        oldCWND,
				"new_cwnd":        b.cwnd,
				"beta":            b.params.Beta,
			})
		}
	}
}

// updateLossRateEMA updates the exponential moving average of loss rate
func (b *BBRv3) updateLossRateEMA() {
	if b.packetsSent > 0 {
		currentLossRate := float64(b.packetsLost) / float64(b.packetsSent)
		if b.lossRateEMA == 0 {
			b.lossRateEMA = currentLossRate
		} else {
			b.lossRateEMA = 0.875*b.lossRateEMA + 0.125*currentLossRate
		}
	}
}

// trackPhaseTransition tracks phase durations when state changes
func (b *BBRv3) trackPhaseTransition(oldState bbrv3State, now time.Time) {
	b.metricsMux.Lock()
	defer b.metricsMux.Unlock()
	
	// Record duration of previous phase
	if startTime, ok := b.phaseStartTimes[oldState]; ok {
		duration := now.Sub(startTime)
		b.phaseDurations[b.getStateStringFromState(oldState)] = duration
	}
	// Start timing new phase
	b.phaseStartTimes[b.state] = now
	
	// Update recovery time if transitioning from recovery state
	if oldState == bbrv3ProbeRTT || (oldState == bbrv3Drain && b.state == bbrv3ProbeBW) {
		if !b.lastRecoveryTime.IsZero() && b.lastLossTime.After(b.lastRecoveryTime) {
			b.lastRecoveryTime = now
		}
	}
}

// finalizeCwndAndPacing ensures minimum cwnd and sets pacing rate
func (b *BBRv3) finalizeCwndAndPacing() {
	// Ensure minimum cwnd
	if b.cwnd < 2*b.mtu {
		b.cwnd = 2 * b.mtu
	}
	
	// Set pacing rate
	if b.pacingBps <= 0 && b.minRTT > 0 {
		b.pacingBps = int64(float64(b.cwnd) / b.minRTT.Seconds())
	}
	b.pacer.SetRate(b.pacingBps)
}

// updateHeadroomUsage updates the headroom usage metric
func (b *BBRv3) updateHeadroomUsage() {
	bdp := b.BDP()
	if bdp > 0 {
		inflightTarget := b.inflightTarget()
		if b.currentInflight > 0 {
			headroomSize := bdp - inflightTarget
			if headroomSize > 0 {
				headroomUsage := (float64(b.currentInflight) - inflightTarget) / headroomSize
				if headroomUsage > 1.0 {
					headroomUsage = 1.0
				} else if headroomUsage < 0.0 {
					headroomUsage = 0.0
				}
				b.metrics.HeadroomUsage = headroomUsage
			}
		}
	}
}

// logChanges logs state, cwnd, and pacing changes
func (b *BBRv3) logChanges(oldState bbrv3State, oldCWND int, oldPacing int64) {
	// Log state change
	if oldState != b.state {
		b.updateMetrics()
		b.logQlogEvent("state_change", map[string]interface{}{
			"old_state":        b.getStateStringFromState(oldState),
			"new_state":        b.getStateString(),
			"reason":           "state_machine",
			"bandwidth":        b.bw,
			"bw_fast":          b.bwFast,
			"bw_slow":          b.bwSlow,
			"min_rtt":          float64(b.minRTT.Nanoseconds()) / 1e6,
			"cwnd":             b.cwnd,
			"pacing_rate":      b.pacingBps,
			"loss_rate_ema":    b.lossRateEMA,
		})
	}
	
	// Log CWND update
	if oldCWND != b.cwnd {
		b.logQlogEvent("cwnd_update", map[string]interface{}{
			"old_cwnd":         oldCWND,
			"new_cwnd":         b.cwnd,
			"change":           b.cwnd - oldCWND,
			"reason":           "ack_processing",
			"bandwidth":        b.bw,
			"inflight_target":  b.inflightTarget(),
			"headroom_usage":   b.metrics.HeadroomUsage,
		})
	}
	
	// Log pacing update
	if oldPacing != b.pacingBps {
		b.logQlogEvent("pacing_update", map[string]interface{}{
			"old_rate":  oldPacing,
			"new_rate":  b.pacingBps,
			"bandwidth": b.bw,
			"min_rtt":   float64(b.minRTT.Nanoseconds()) / 1e6,
		})
	}
}

// OnLoss handles a loss event
func (b *BBRv3) OnLoss() (cwnd int, pacing int64) {
	b.packetsLost++
	b.packetsSent++
	
	// Track loss time for recovery metrics
	b.lastLossTime = time.Now()
	
	// Accumulate lost bytes in current round
	b.roundLost += int64(b.mtu)
	
	b.updateMetrics()
	b.logQlogEvent("packet_loss", map[string]interface{}{
		"packets_lost":   b.packetsLost,
		"packets_sent":   b.packetsSent,
		"loss_rate_ema":  b.lossRateEMA,
		"round_lost":     b.roundLost,
		"round_acked":    b.roundAcked,
		"loss_threshold": b.params.LossThreshold,
		"beta":           b.params.Beta,
	})
	
	return b.cwnd, b.pacingBps
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}