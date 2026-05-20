package server

import (
	"distributed-kv-store/internal/ring"
	"distributed-kv-store/internal/store"
	"fmt"
	"io"
	"net/http"
)

type Server struct {
	Store  *store.Store
	Router *Router
}

func NewServer(store *store.Store, self string, peers []string) *Server {
	r := ring.NewRing()
	r.AddNode(self)
	for _, peer := range peers {
		r.AddNode(peer)
	}

	return &Server{Store: store, Router: &Router{self: self, ring: r}}
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

func (s *Server) HandleGet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	owner, err := s.Router.GetNode(key)
	if err != nil {
		http.Error(w, "ring error", http.StatusInternalServerError)
		return
	}

	if owner != s.Router.self {
		s.forward(w, r, owner)
		return
	}

	val, ok := s.Store.Get(key)
	if !ok {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	fmt.Fprint(w, val)
}

func (s *Server) HandlePut(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	owner, err := s.Router.GetNode(key)
	if err != nil {
		http.Error(w, "ring error", http.StatusInternalServerError)
		return
	}

	if owner != s.Router.self {
		s.forward(w, r, owner)
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

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) HandleDelete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	owner, err := s.Router.GetNode(key)
	if err != nil {
		http.Error(w, "ring error", http.StatusInternalServerError)
		return
	}

	if owner != s.Router.self {
		s.forward(w, r, owner)
		return
	}

	ok := s.Store.Delete(key)
	if !ok {
		http.Error(w, "delete key not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
