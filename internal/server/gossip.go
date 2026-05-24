package server

import (
	"net/http"
	"sync"
	"time"
)

type Gossip struct {
	self  string
	peers []string
	alive map[string]bool
	mu    sync.RWMutex
}

func NewGossip(self string, peers []string) *Gossip {
	g := &Gossip{
		self:  self,
		peers: peers,
		alive: make(map[string]bool),
	}

	for _, peer := range peers {
		g.alive[peer] = true
	}

	return g
}

func (g *Gossip) Start() {
	failures := map[string]int{}

	go func() {
		for {
			for _, peer := range g.peers {
				resp, err := http.Get("http://" + peer + "/health")
				if err != nil {
					failures[peer]++
					if failures[peer] >= 2 {
						g.mu.Lock()
						g.alive[peer] = false
						g.mu.Unlock()
					}
				} else {
					resp.Body.Close()
					failures[peer] = 0
					g.mu.Lock()
					g.alive[peer] = true
					g.mu.Unlock()
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func (g *Gossip) IsAlive(peer string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.alive[peer]
}
