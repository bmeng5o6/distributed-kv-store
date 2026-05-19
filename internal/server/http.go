package server

import (
	"distributed-kv-store/internal/store"
	"fmt"
	"io"
	"net/http"
)

type Server struct {
	Store *store.Store
}

func NewServer(store *store.Store) *Server {
	return &Server{Store: store}
}

func (s *Server) HandleGet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	val, ok := s.Store.Get(key)
	if !ok {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	fmt.Fprint(w, val)
}

func (s *Server) HandlePut(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
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
	ok := s.Store.Delete(key)
	if !ok {
		http.Error(w, "delete key not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
