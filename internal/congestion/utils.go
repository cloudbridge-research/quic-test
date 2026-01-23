package congestion

import (
	"math"
	"sort"
	"time"
)

// CalculateRTTPercentiles calculates RTT percentiles from samples
func CalculateRTTPercentiles(samples []time.Duration) (p50, p95, p99 time.Duration) {
	if len(samples) == 0 {
		return 0, 0, 0
	}
	
	// Convert to float64 for easier calculation
	floatSamples := make([]float64, len(samples))
	for i, sample := range samples {
		floatSamples[i] = float64(sample.Nanoseconds())
	}
	
	// Sort samples
	sort.Float64s(floatSamples)
	
	// Calculate percentiles
	p50f := percentile(floatSamples, 0.50)
	p95f := percentile(floatSamples, 0.95)
	p99f := percentile(floatSamples, 0.99)
	
	return time.Duration(p50f), time.Duration(p95f), time.Duration(p99f)
}

// CalculateJitter calculates jitter (standard deviation) from RTT samples
func CalculateJitter(samples []time.Duration) time.Duration {
	if len(samples) < 2 {
		return 0
	}
	
	// Calculate mean
	var sum float64
	for _, sample := range samples {
		sum += float64(sample.Nanoseconds())
	}
	mean := sum / float64(len(samples))
	
	// Calculate variance
	var variance float64
	for _, sample := range samples {
		diff := float64(sample.Nanoseconds()) - mean
		variance += diff * diff
	}
	variance /= float64(len(samples) - 1)
	
	// Return standard deviation as jitter
	return time.Duration(math.Sqrt(variance))
}

// JainFairnessIndex calculates Jain's Fairness Index from throughput values
func JainFairnessIndex(throughputs []float64) float64 {
	if len(throughputs) == 0 {
		return 1.0 // Perfect fairness for empty set
	}
	
	var sum, sumSquares float64
	for _, throughput := range throughputs {
		sum += throughput
		sumSquares += throughput * throughput
	}
	
	if sumSquares == 0 {
		return 1.0 // All throughputs are zero
	}
	
	n := float64(len(throughputs))
	return (sum * sum) / (n * sumSquares)
}

// percentile calculates the given percentile from sorted samples
func percentile(sortedSamples []float64, p float64) float64 {
	if len(sortedSamples) == 0 {
		return 0
	}
	
	if len(sortedSamples) == 1 {
		return sortedSamples[0]
	}
	
	// Calculate index
	index := p * float64(len(sortedSamples)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	
	if lower == upper {
		return sortedSamples[lower]
	}
	
	// Linear interpolation
	weight := index - float64(lower)
	return sortedSamples[lower]*(1-weight) + sortedSamples[upper]*weight
}