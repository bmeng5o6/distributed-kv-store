package server

import (
	"distributed-kv-store/internal/ring"
	"distributed-kv-store/internal/store"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Server struct {
	Store  *store.Store
	Router *Router
	Gossip *Gossip
}

func NewServer(store *store.Store, self string, peers []string) *Server {
	r := ring.NewRing()
	r.AddNode(self)
	for _, peer := range peers {
		r.AddNode(peer)
	}

	g := NewGossip(self, peers)
	g.Start()

	return &Server{Store: store, Router: &Router{self: self, ring: r}, Gossip: g}
}

func (s *Server) forward(w http.ResponseWriter, r *http.Request, owner string) {
	url := "http://" + owner + r.URL.Path
	req, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		http.Error(w, "forward error", http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "forward error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func (s *Server) replicate(method, key, value, addr string) error {
	url := "http://" + addr + "/keys/" + key
	req, err := http.NewRequest(method, url, strings.NewReader(value))
	if err != nil {
		return err
	}
	req.Header.Set("X-Replication", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (s *Server) tryForward(r *http.Request, node string) (*http.Response, error) {
	url := "http://" + node + r.URL.Path
	req, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

func (s *Server) HandleGet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	nodes, err := s.Router.GetNodes(key, 2)
	if err != nil {
		http.Error(w, "ring error", http.StatusInternalServerError)
		return
	}

	// Walks through nodes that own key in order.
	for _, node := range nodes {
		if node == s.Router.self {
			val, ok := s.Store.Get(key)
			if !ok {
				http.Error(w, "key not found", http.StatusNotFound)
				return
			}
			fmt.Fprint(w, val)
			return
		}

		if !s.Gossip.IsAlive(node) {
			log.Printf("node %s known dead, skipping", node)
			continue
		}

		resp, err := s.tryForward(r, node)
		if err != nil {
			log.Printf("node %s unreachable, trying next", node)
			continue
		}

		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		return
	}

	http.Error(w, "all nodes unreachable", http.StatusServiceUnavailable)
}

func (s *Server) HandlePut(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if r.Header.Get("X-Replication") == "true" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "body read error", http.StatusBadRequest)
			return
		}

		err = s.Store.Put(key, string(body))
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		return
	}

	nodes, err := s.Router.GetNodes(key, 2)
	if err != nil {
		http.Error(w, "ring error", http.StatusInternalServerError)
		return
	}

	if nodes[0] != s.Router.self {
		if s.Gossip.IsAlive(nodes[0]) {
			s.forward(w, r, nodes[0])
			return
		}

		for _, replica := range nodes[1:] {
			if s.Gossip.IsAlive(replica) {
				s.forward(w, r, replica)
				return
			}
		}

		http.Error(w, "all nodes unreachable", http.StatusServiceUnavailable)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "body read error", http.StatusBadRequest)
		return
	}

	val := string(body)

	err = s.Store.Put(key, val)
	if err != nil {
		http.Error(w, "value internal error", http.StatusInternalServerError)
		return
	}

	for _, replica := range nodes[1:] {
		err = s.replicate("PUT", key, val, replica)
		if err != nil {
			log.Printf("replication to %s failed: %v", replica, err)
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) HandleDelete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if r.Header.Get("X-Replication") == "true" {
		ok := s.Store.Delete(key)
		if !ok {
			http.Error(w, "key not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	nodes, err := s.Router.GetNodes(key, 2)
	if err != nil {
		http.Error(w, "ring error", http.StatusInternalServerError)
		return
	}

	if nodes[0] != s.Router.self {
		if s.Gossip.IsAlive(nodes[0]) {
			s.forward(w, r, nodes[0])
			return
		}

		for _, replica := range nodes[1:] {
			if s.Gossip.IsAlive(replica) {
				s.forward(w, r, replica)
				return
			}
		}

		http.Error(w, "all nodes unreachable", http.StatusServiceUnavailable)
		return
	}

	ok := s.Store.Delete(key)
	if !ok {
		http.Error(w, "delete key not found", http.StatusNotFound)
		return
	}

	if len(nodes) > 1 {
		for _, replica := range nodes[1:] {
			err := s.replicate("DELETE", key, "", replica)
			if err != nil {
				log.Printf("delete replication to %s failed: %v", replica, err)
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
