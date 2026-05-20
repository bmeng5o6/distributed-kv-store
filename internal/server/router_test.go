package server

// import (
// 	"fmt"
// 	"testing"
// 	"distributed-kv-store/internal/ring"
// )

// // run: go test ./...

// func TestRouter_KeyRoutesToSameNode(t *testing.T) {
// 	r := ring.NewRing()
// 	r.AddNode("serverA")

// 	node, err := r.GetNode("username")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if node == "" {
// 		t.Errorf("expected node, got empty string")
// 	}
// }
