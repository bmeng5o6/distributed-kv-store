package server

import "distributed-kv-store/internal/ring"

type Router struct {
	self string
	ring *ring.Ring
}

func (r *Router) GetNodes(key string, count int) ([]string, error) {
	return r.ring.GetNodes(key, count)
}
