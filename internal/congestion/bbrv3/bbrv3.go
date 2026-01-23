// Package bbrv3 implements the BBRv3 congestion control algorithm
// Based on draft-ietf-ccwg-bbr-04 specification
//
// Key improvements over BBRv2:
// - Dual-scale bandwidth model (fast/slow)
// - Loss threshold = 2%
// - β = 0.7 (cwnd reduction factor)
// - Headroom = 0.15 BDP
// - Adaptive pacing gain (Startup 2.77, Drain 0.35)
// - ProbeRTTDuration = 200ms
package bbrv3

import "time"

// CongestionController interface that BBRv3 implements
type CongestionController interface {
	Name() string
	OnAck(s Sample) (cwnd int, pacing int64)
	OnLoss() (cwnd int, pacing int64)
	OnPacketSent()
	OnPacketAcked()
	GetCWND() int
	GetPacingRate() int64
	GetBandwidth() float64
	GetMinRTT() time.Duration
	GetLossRate() float64
	GetMetrics() BBRv3Metrics
	SetQlogCallback(callback func(eventType string, data map[string]interface{}))
}

// Ensure BBRv3 implements CongestionController
var _ CongestionController = (*BBRv3)(nil)

// NewBBRv3CongestionController creates a new BBRv3 congestion controller
// This is the main entry point for creating BBRv3 instances
func NewBBRv3CongestionController(mtu int, initialCWND int) CongestionController {
	return NewBBRv3(mtu, initialCWND)
}

// GetAlgorithmInfo returns information about the BBRv3 algorithm
func GetAlgorithmInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "BBRv3",
		"version":     "draft-ietf-ccwg-bbr-04",
		"description": "BBRv3 congestion control with dual-scale bandwidth estimation",
		"features": []string{
			"Dual-scale bandwidth model",
			"Adaptive pacing gains",
			"Loss threshold based recovery",
			"Headroom reservation",
			"ProbeRTT for minimum RTT tracking",
		},
		"parameters": map[string]interface{}{
			"loss_threshold":      0.02,  // 2%
			"beta":               0.7,   // cwnd reduction factor
			"headroom_fraction":  0.15,  // 15% headroom
			"startup_pacing_gain": 2.77, // Startup pacing gain
			"drain_pacing_gain":  0.35,  // Drain pacing gain
			"probe_rtt_duration": "200ms",
		},
	}
}