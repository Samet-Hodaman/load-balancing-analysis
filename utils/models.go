package utils

type LoadBalancer interface {
	SelectServer(servers []*Server) *Server
	Name() string
}

type Request struct {
	Cost float64
}

type RoundRobin struct {
	index int
}

type Server struct {
	ID       int
	Capacity float64
	Load     float64
}

type WeightedRoundRobin struct {
	current int
}
