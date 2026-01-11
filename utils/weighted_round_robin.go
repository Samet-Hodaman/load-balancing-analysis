package utils

func NewWeightedRoundRobin() *WeightedRoundRobin {
	return &WeightedRoundRobin{current: 0}
}

func (w *WeightedRoundRobin) SelectServer(servers []*Server) *Server {
	best := servers[0]
	bestScore := best.Load / best.Capacity

	for _, s := range servers {
		score := s.Load / s.Capacity
		if score < bestScore {
			best = s
			bestScore = score
		}
	}
	return best
}

func (w *WeightedRoundRobin) Name() string {
	return "WeightedRoundRobin"
}
