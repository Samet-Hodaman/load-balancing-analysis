package utils

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{index: 0}
}

func (rr *RoundRobin) SelectServer(servers []*Server) *Server {
	server := servers[rr.index]
	rr.index = (rr.index + 1) % len(servers)
	return server
}

func (rr *RoundRobin) Name() string {
	return "RoundRobin"
}
