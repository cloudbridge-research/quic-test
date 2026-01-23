package bbrv3

import (
	"time"
)

// maxDur returns the maximum of two durations
func maxDur(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

// maxF returns the maximum of two float64 values
func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// BDP calculates the bandwidth-delay product
func (b *BBRv3) BDP() float64 {
	if b.minRTT <= 0 {
		return float64(b.cwnd)
	}
	return b.bw * b.minRTT.Seconds()
}

// bdp is an alias for BDP (for backward compatibility)
func (b *BBRv3) bdp() float64 {
	return b.BDP()
}

// inflightTarget calculates the target inflight with headroom reserved
// Headroom is RESERVED (not added), so inflight_target = BDP * (1 - headroom_fraction)
func (b *BBRv3) inflightTarget() float64 {
	bdp := b.BDP()
	return bdp * (1.0 - b.params.HeadroomFraction) // e.g. 0.85 * BDP
}

// updatePacingQuantum updates the pacing quantum
// quantum = max(2*MTU, min(64KB, pacing_rate*minRTT/8))
func (b *BBRv3) updatePacingQuantum() {
	if b.minRTT <= 0 || b.bw <= 0 {
		b.sendQuantum = int64(2 * b.mtu)
		return
	}
	
	// Calculate quantum based on pacing rate and RTT
	quantum := float64(b.pacingBps) * b.minRTT.Seconds() / 8.0
	
	// Clamp to reasonable bounds
	minQuantum := int64(2 * b.mtu)
	maxQuantum := int64(64 * 1024) // 64KB
	
	if quantum < float64(minQuantum) {
		b.sendQuantum = minQuantum
	} else if quantum > float64(maxQuantum) {
		b.sendQuantum = maxQuantum
	} else {
		b.sendQuantum = int64(quantum)
	}
}

// NewPacer creates a new pacer (placeholder implementation)
func NewPacer(mtu int) Pacer {
	return &SimplePacer{
		mtu:  mtu,
		rate: int64(mtu * 100), // Initial rate
	}
}

// SimplePacer is a simple implementation of the Pacer interface
type SimplePacer struct {
	mtu  int
	rate int64
}

// SetRate sets the pacing rate
func (p *SimplePacer) SetRate(bps int64) {
	p.rate = bps
}

// GetRate returns the current pacing rate
func (p *SimplePacer) GetRate() int64 {
	return p.rate
}