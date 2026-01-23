package bbrv3

import (
	"sync"
	"time"
)

// bbrv3State represents the current state of the BBRv3 algorithm
type bbrv3State int

const (
	bbrv3Startup bbrv3State = iota
	bbrv3Drain
	bbrv3ProbeBW
	bbrv3ProbeRTT
)

// BBRv3Parameters contains BBRv3 algorithm parameters
type BBRv3Parameters struct {
	// Loss threshold (2% according to draft)
	LossThreshold float64
	
	// β factor for cwnd reduction (0.7)
	Beta float64
	
	// Headroom as fraction of BDP (0.15)
	HeadroomFraction float64
	
	// Pacing gains
	StartupPacingGain float64 // 2.77
	DrainPacingGain  float64 // 0.35
	
	// ProbeRTT duration (200ms)
	ProbeRTTDuration time.Duration
}

// DefaultBBRv3Parameters returns default BBRv3 parameters per draft
func DefaultBBRv3Parameters() BBRv3Parameters {
	return BBRv3Parameters{
		LossThreshold:     0.02,  // 2%
		Beta:              0.7,   // cwnd reduction factor
		HeadroomFraction:  0.15,  // 15% headroom
		StartupPacingGain: 2.77,  // Startup pacing gain
		DrainPacingGain:   0.35,  // Drain pacing gain
		ProbeRTTDuration:  200 * time.Millisecond,
	}
}

// OptimizedBBRv3Parameters returns optimized parameters for better performance
// These parameters are tuned for inter-regional networks (RTT > 80ms)
func OptimizedBBRv3Parameters() BBRv3Parameters {
	return BBRv3Parameters{
		LossThreshold:     0.018, // 1.8% (slightly more sensitive for faster reaction)
		Beta:              0.75,  // 0.75 (less aggressive cwnd reduction)
		HeadroomFraction:  0.12,  // 12% headroom (slightly less conservative)
		StartupPacingGain: 2.89,  // Higher startup gain for faster ramp-up
		DrainPacingGain:   0.32,  // Slightly faster drain
		ProbeRTTDuration:  180 * time.Millisecond, // Shorter ProbeRTT for better throughput
	}
}

// BBRv3Metrics contains BBRv3-specific metrics for visualization
type BBRv3Metrics struct {
	Phase          string  `json:"phase"`            // Startup, Drain, ProbeBW, ProbeRTT
	BandwidthFast  float64 `json:"bw_fast"`         // Fast-scale bandwidth estimate (bps)
	BandwidthSlow  float64 `json:"bw_slow"`         // Slow-scale bandwidth estimate (bps)
	Bandwidth      float64 `json:"bandwidth"`        // Current bandwidth estimate (bps)
	MinRTT         float64 `json:"min_rtt_ms"`      // Minimum RTT in milliseconds
	CWND           int     `json:"cwnd"`            // Congestion window
	PacingRate     int64   `json:"pacing_rate"`     // Pacing rate (bps)
	LossRate       float64 `json:"loss_rate"`       // Packet loss rate (0-1)
	LossRateRound  float64 `json:"loss_rate_round"` // Loss rate in current round
	LossRateEMA    float64 `json:"loss_rate_ema"`   // Exponential moving average loss rate
	LossThreshold  float64 `json:"loss_threshold"`  // Loss threshold parameter
	BDP            float64 `json:"bdp"`             // Bandwidth-delay product
	InflightTarget float64 `json:"inflight_target"` // Target inflight bytes
	CurrentInflight int64  `json:"current_inflight"` // Current inflight bytes
	SendQuantum    int64   `json:"send_quantum"`    // Pacing quantum
	PacingQuantum  int64   `json:"pacing_quantum"`  // Alias for SendQuantum
	Bufferbloat    float64 `json:"bufferbloat_ms"`  // Estimated bufferbloat in ms
	HeadroomUsage  float64 `json:"headroom_usage"`  // Headroom usage (0-1)
	RecoveryTimeMs float64 `json:"recovery_time_ms"` // Time to recover from loss
	
	// Calculated metrics
	PacingGain             float64 `json:"pacing_gain"`              // Current pacing gain
	CWNDGain               float64 `json:"cwnd_gain"`                // Current CWND gain
	ProbeRTTMinMs          float64 `json:"probe_rtt_min_ms"`         // ProbeRTT minimum RTT
	BufferbloatFactor      float64 `json:"bufferbloat_factor"`       // Bufferbloat factor
	StabilityIndex         float64 `json:"stability_index"`          // Stability index
	LossRecoveryEfficiency float64 `json:"loss_recovery_efficiency"` // Loss recovery efficiency
	
	// Phase durations for analysis
	StartupDuration  time.Duration `json:"startup_duration_ms"`
	DrainDuration    time.Duration `json:"drain_duration_ms"`
	ProbeBWDuration  time.Duration `json:"probe_bw_duration_ms"`
	ProbeRTTDuration time.Duration `json:"probe_rtt_duration_ms"`
	
	// Additional metrics for detailed analysis
	PhaseDurationMs map[string]float64 `json:"phase_duration_ms,omitempty"`
}

// BBRv3 implements the BBRv3 congestion control algorithm
type BBRv3 struct {
	state       bbrv3State
	mtu         int
	cwnd        int
	pacingBps   int64
	minRTT      time.Duration
	minRTTSince time.Time
	bwFast      float64 // Fast-scale bandwidth estimate
	bwSlow      float64 // Slow-scale bandwidth estimate
	bw          float64 // Current bandwidth estimate (max of fast/slow)
	cycleIdx    int
	lastStateTs time.Time
	pacer       Pacer
	params      BBRv3Parameters
	
	// Loss tracking by round (not sliding window)
	roundAcked      int64 // Bytes acked in current round
	roundLost       int64 // Bytes lost in current round
	roundStartTime  time.Time
	
	// Loss rate tracking (for metrics)
	packetsSent      int64
	packetsLost      int64
	lossRateEMA      float64 // Exponential moving average
	
	// Headroom tracking
	currentInflight int64
	
	// Pacing quantum
	sendQuantum int64
	
	// Phase timing tracking
	phaseStartTimes map[bbrv3State]time.Time
	phaseDurations  map[string]time.Duration
	
	// RTT tracking for bufferbloat calculation
	recentRTTs []time.Duration // Last N RTT samples
	recentRTTIdx int
	
	// Recovery tracking
	lastLossTime     time.Time
	lastRecoveryTime time.Time
	recoveredPackets int64
	
	// Throughput delta tracking for stability
	lastThroughput float64
	lastRTT        time.Duration
	
	// Metrics for visualization
	metrics    BBRv3Metrics
	metricsMux sync.Mutex // Protects metrics from concurrent access

	// Qlog callback
	qlogCallback func(eventType string, data map[string]interface{})
}

// Sample represents a congestion control sample (imported from parent package)
type Sample struct {
	RS   RateSample    // from rate sampler
	RTT  time.Duration
	Loss bool
}

// RateSample represents bandwidth measurement data
type RateSample struct {
	BytesAcked   int64
	IsAppLimited bool
	Interval     time.Duration
}

// BandwidthBps calculates bandwidth in bits per second
func (rs *RateSample) BandwidthBps() float64 {
	if rs.Interval <= 0 {
		return 0
	}
	return float64(rs.BytesAcked*8) / rs.Interval.Seconds()
}

// Pacer interface for pacing packets
type Pacer interface {
	SetRate(bps int64)
	GetRate() int64
}