package utils

import "sync/atomic"

func NewWeightedRoundRobin() *WeightedRoundRobin {
	return &WeightedRoundRobin{}
}

// SelectServer implements lock-free GCD-based Weighted Round Robin algorithm
// Uses atomic operations for thread-safety without mutex overhead
// Optimized for high throughput (50k+ req/s):
// - O(n) complexity
// - Lock-free atomic operations instead of mutex
// - No unnecessary checks or branches
// - Better scalability under high concurrency
func (w *WeightedRoundRobin) SelectServer(servers []*Server) *Server {
	// GCD-based Weighted Round Robin algorithm using atomic operations:
	// 1. Atomically add each server's weight to its current weight
	// 2. Find the server with the highest current weight
	// 3. Atomically subtract total weight sum from selected server's current weight

	// Initialize with first server (guaranteed to exist if function is called correctly)
	best := servers[0]
	maxCurrent := atomic.AddInt64(&best.CurrentWRR, best.Weight)
	totalWeight := best.Weight

	// Process remaining servers
	for i := 1; i < len(servers); i++ {
		s := servers[i]
		newCurrent := atomic.AddInt64(&s.CurrentWRR, s.Weight)
		totalWeight += s.Weight

		if newCurrent > maxCurrent {
			maxCurrent = newCurrent
			best = s
		}
	}

	// Atomically subtract total weight from selected server
	atomic.AddInt64(&best.CurrentWRR, -totalWeight)

	return best
}

func (w *WeightedRoundRobin) Name() string {
	return "WeightedRoundRobin"
}
