package utils

import "sync/atomic"

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{index: 0}
}

func (rr *RoundRobin) SelectServer(servers []*Server) *Server {
	if len(servers) == 0 {
		return nil
	}
	// Thread-safe index increment using atomic operations
	i := atomic.AddUint64(&rr.index, 1)
	return servers[i%uint64(len(servers))]
}

func (rr *RoundRobin) Name() string {
	return "RoundRobin"
}
