package main

import (
	"distributed-kv-store/internal/server"
	"distributed-kv-store/internal/store"
	"log"
	"net/http"
)

func main() {
	st, err := store.NewStore("wal.log")
	if err != nil {
		log.Fatal(err)
	}

	srv := server.NewServer(st)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /keys/{key}", srv.HandleGet)
	mux.HandleFunc("PUT /keys/{key}", srv.HandlePut)
	mux.HandleFunc("DELETE /keys/{key}", srv.HandleDelete)

	log.Println("listening on: 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
