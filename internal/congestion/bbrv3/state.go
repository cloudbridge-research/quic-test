package bbrv3

import (
	"time"
)

// NewBBRv3 creates a new BBRv3 congestion controller
func NewBBRv3(mtu int, initialCWND int) *BBRv3 {
	if mtu <= 0 {
		mtu = 1460
	}
	if initialCWND <= 0 {
		initialCWND = 32 * mtu
	}

	// Use optimized parameters for better performance
	params := OptimizedBBRv3Parameters()
	now := time.Now()
	
	b := &BBRv3{
		state:         bbrv3Startup,
		mtu:           mtu,
		cwnd:          initialCWND,
		lastStateTs:   now,
		pacer:         NewPacer(mtu),
		params:        params,
		roundStartTime: now,
		sendQuantum:   int64(2 * mtu), // Initial quantum
		phaseStartTimes: make(map[bbrv3State]time.Time),
		phaseDurations:  make(map[string]time.Duration),
		recentRTTs:      make([]time.Duration, 100), // Last 100 RTT samples
	}
	
	b.phaseStartTimes[bbrv3Startup] = now
	b.updateMetrics()
	return b
}

// Name returns the algorithm name
func (b *BBRv3) Name() string {
	return "bbrv3"
}

// GetState returns the current BBRv3 state
func (b *BBRv3) GetState() bbrv3State {
	return b.state
}

// GetStateInt returns the current BBRv3 state as integer
func (b *BBRv3) GetStateInt() int {
	return int(b.state)
}

// GetCWND returns the current congestion window
func (b *BBRv3) GetCWND() int {
	return b.cwnd
}

// GetPacingRate returns the current pacing rate
func (b *BBRv3) GetPacingRate() int64 {
	return b.pacingBps
}

// GetBandwidth returns the current bandwidth estimate
func (b *BBRv3) GetBandwidth() float64 {
	return b.bw
}

// GetBandwidthFast returns the fast-scale bandwidth estimate
func (b *BBRv3) GetBandwidthFast() float64 {
	return b.bwFast
}

// GetBandwidthSlow returns the slow-scale bandwidth estimate
func (b *BBRv3) GetBandwidthSlow() float64 {
	return b.bwSlow
}

// GetMinRTT returns the minimum RTT
func (b *BBRv3) GetMinRTT() time.Duration {
	return b.minRTT
}

// GetLossRate returns the current packet loss rate (EMA)
func (b *BBRv3) GetLossRate() float64 {
	return b.lossRateEMA
}

// getStateString returns the string representation of the current state
func (b *BBRv3) getStateString() string {
	return b.getStateStringFromState(b.state)
}

// getStateStringFromState returns the string representation of a state
func (b *BBRv3) getStateStringFromState(state bbrv3State) string {
	switch state {
	case bbrv3Startup:
		return "Startup"
	case bbrv3Drain:
		return "Drain"
	case bbrv3ProbeBW:
		return "ProbeBW"
	case bbrv3ProbeRTT:
		return "ProbeRTT"
	default:
		return "Unknown"
	}
}

// resetRound resets the loss tracking round
func (b *BBRv3) resetRound() {
	b.roundAcked = 0
	b.roundLost = 0
	b.roundStartTime = time.Now()
}

// roundTotal returns total bytes in current round
func (b *BBRv3) roundTotal() int64 {
	return b.roundAcked + b.roundLost
}

// fullPipeDetected checks if full pipe bandwidth is detected
func (b *BBRv3) fullPipeDetected() bool {
	// Simple heuristic: if bandwidth hasn't increased significantly
	// in last 2 seconds, consider pipe full
	return time.Since(b.lastStateTs) > 2*time.Second && b.bw > 0
}

// OnPacketSent tracks packet sending for loss rate calculation
func (b *BBRv3) OnPacketSent() {
	b.packetsSent++
	b.currentInflight++
}

// OnPacketAcked tracks packet acknowledgment
func (b *BBRv3) OnPacketAcked() {
	if b.currentInflight > 0 {
		b.currentInflight--
	}
}

// SetQlogCallback sets the callback for qlog events
func (b *BBRv3) SetQlogCallback(callback func(eventType string, data map[string]interface{})) {
	b.qlogCallback = callback
}

// logQlogEvent logs an event to qlog
func (b *BBRv3) logQlogEvent(eventType string, data map[string]interface{}) {
	if b.qlogCallback != nil {
		b.qlogCallback(eventType, data)
	}
}