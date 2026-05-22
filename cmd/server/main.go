package main

import (
	"distributed-kv-store/internal/server"
	"distributed-kv-store/internal/store"
	"flag"
	"log"
	"net/http"
	"strings"
)

func main() {
	port := flag.String("port", "8080", "port to listen on")
	peers := flag.String("peers", "", "comma separated peer addresses")
	host := flag.String("host", "localhost", "hostname of this code")

	flag.Parse()

	st, err := store.NewStore("wal.log")
	if err != nil {
		log.Fatal(err)
	}

	self := *host + ":" + *port
	var peerList []string
	if *peers != "" {
		peerList = strings.Split(*peers, ",")
	}

	srv := server.NewServer(st, self, peerList)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /keys/{key}", srv.HandleGet)
	mux.HandleFunc("PUT /keys/{key}", srv.HandlePut)
	mux.HandleFunc("DELETE /keys/{key}", srv.HandleDelete)
	mux.HandleFunc("GET /health", srv.HandleHealth)

	log.Printf("listening on: %s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, mux))
}
