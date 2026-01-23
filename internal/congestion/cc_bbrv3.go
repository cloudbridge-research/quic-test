package congestion

import (
	"quic-test/internal/congestion/bbrv3"
)

// BBRv3 is an alias for the modular BBRv3 implementation
type BBRv3 = bbrv3.BBRv3

// BBRv3Metrics is an alias for the modular BBRv3 metrics
type BBRv3Metrics = bbrv3.BBRv3Metrics

// BBRv3Parameters is an alias for the modular BBRv3 parameters
type BBRv3Parameters = bbrv3.BBRv3Parameters

// NewBBRv3 creates a new BBRv3 congestion controller
func NewBBRv3(mtu int, initialCWND int) *BBRv3 {
	return bbrv3.NewBBRv3(mtu, initialCWND)
}

// DefaultBBRv3Parameters returns default BBRv3 parameters
func DefaultBBRv3Parameters() BBRv3Parameters {
	return bbrv3.DefaultBBRv3Parameters()
}

// OptimizedBBRv3Parameters returns optimized BBRv3 parameters
func OptimizedBBRv3Parameters() BBRv3Parameters {
	return bbrv3.OptimizedBBRv3Parameters()
}

// GetBBRv3AlgorithmInfo returns information about the BBRv3 algorithm
func GetBBRv3AlgorithmInfo() map[string]interface{} {
	return bbrv3.GetAlgorithmInfo()
}