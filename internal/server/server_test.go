package server

import (
	"distributed-kv-store/internal/store"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// run: go test ./...

func TestHandleGet_LocalRead(t *testing.T) {
	f, err := os.CreateTemp("", "wal-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	st, err := store.NewStore(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	srv := NewServer(st, "localhost:8080", []string{})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /keys/{key}", srv.HandleGet)
	mux.HandleFunc("PUT /keys/{key}", srv.HandlePut)
	mux.HandleFunc("DELETE /keys/{key}", srv.HandleDelete)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	req, _ := http.NewRequest("PUT", ts.URL+"/key/username", strings.NewReader("brian"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestHandleGet_Replication(t *testing.T) {
	f1, err := os.CreateTemp("", "wal-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := os.CreateTemp("", "wal-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	st1, err := store.NewStore(f1.Name())
	if err != nil {
		t.Fatal(err)
	}

	st2, err := store.NewStore(f2.Name())
	if err != nil {
		t.Fatal(err)
	}

	addr1 := "localhost:19001"
	addr2 := "localhost:19002"

	srv1 := NewServer(st1, addr1, []string{addr2})
	srv2 := NewServer(st2, addr2, []string{addr1})

	mux1 := http.NewServeMux()
	mux1.HandleFunc("GET /keys/{key}", srv1.HandleGet)
	mux1.HandleFunc("PUT /keys/{key}", srv1.HandlePut)
	mux1.HandleFunc("DELETE /keys/{key}", srv1.HandleDelete)

	mux2 := http.NewServeMux()
	mux2.HandleFunc("GET /keys/{key}", srv2.HandleGet)
	mux2.HandleFunc("PUT /keys/{key}", srv2.HandlePut)
	mux2.HandleFunc("DELETE /keys/{key}", srv2.HandleDelete)

	ts1 := httptest.NewUnstartedServer(mux1)
	ts2 := httptest.NewUnstartedServer(mux2)

	l1, err := net.Listen("tcp", addr1)
	if err != nil {
		t.Fatal(err)
	}
	ts1.Listener = l1
	ts1.Start()
	defer ts1.Close()

	l2, err := net.Listen("tcp", addr2)
	if err != nil {
		t.Fatal(err)
	}
	ts2.Listener = l2
	ts2.Start()
	defer ts2.Close()

	// PUT a value
	req, _ := http.NewRequest("PUT", "http://"+addr1+"/keys/username", strings.NewReader("brian"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	ts1.Close()

	resp, err = http.Get("http://" + addr2 + "/keys/username")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if string(body) != "brian" {
		t.Errorf("expected brian, got %s", string(body))
	}
}
