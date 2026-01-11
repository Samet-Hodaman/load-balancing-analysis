package utils

func AverageLatency(latencies []float64) float64 {
	var sum float64
	for _, l := range latencies {
		sum += l
	}
	return sum / float64(len(latencies))
}
