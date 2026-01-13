package utils

import (
	"net/url"
)

type LoadBalancer interface {
	SelectServer(servers []*Server) *Server
	Name() string
}

type Request struct {
	Cost float64
}

type RoundRobin struct {
	index uint64 // Uses atomic operations for thread-safety
}

type Server struct {
	ID          int
	Capacity    float64
	Load        float64
	URL         *url.URL
	Conns       int64
	Weight      int64 // Weight for Weighted Round Robin
	CurrentWRR  int64 // Current weight for Weighted Round Robin
	Connections int64 // Number of active connections
}

type WeightedRoundRobin struct {
	// Lock-free: uses atomic operations on Server.CurrentWRR for thread-safety
}

type LeastConnections struct {
	// Least Connections algorithm doesn't require special state
	// On each selection, we find the server with the fewest connections
}
