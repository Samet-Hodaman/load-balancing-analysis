package utils

import "sync/atomic"

// NewLeastConnections creates a new LeastConnections load balancer
func NewLeastConnections() *LeastConnections {
	return &LeastConnections{}
}

// SelectServer selects the server with the fewest active connections
func (lc *LeastConnections) SelectServer(servers []*Server) *Server {
	if len(servers) == 0 {
		return nil
	}

	var best *Server
	min := int64(^uint64(0) >> 1)

	for _, s := range servers {
		c := atomic.LoadInt64(&s.Conns)
		if c < min {
			min = c
			best = s
		}
	}

	return best
}

// Name returns the name of the algorithm
func (lc *LeastConnections) Name() string {
	return "LeastConnections"
}
