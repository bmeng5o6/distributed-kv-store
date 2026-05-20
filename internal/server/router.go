package server

import "distributed-kv-store/internal/ring"

type Router struct {
	self string
	ring *ring.Ring
}

func (r *Router) GetNode(key string) (string, error) {
	return r.ring.GetNode(key)
}
