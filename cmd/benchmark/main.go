package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

var nodes = []string{
	"localhost:8080",
	"localhost:8081",
	"localhost:8082",
}

func main() {
	benchmarkWrites()
	benchmarkReads()
	benchmarkFailover()
}

func printPercentiles(latencies []time.Duration) {
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	p50 := latencies[len(latencies)*50/100]
	p95 := latencies[len(latencies)*95/100]
	p99 := latencies[len(latencies)*99/100]

	fmt.Printf("P50: %v\n", p50)
	fmt.Printf("P95: %v\n", p95)
	fmt.Printf("P99: %v\n", p99)
}

func benchmarkWrites() {
	fmt.Println("Benchmark Writes")
	latencies := []time.Duration{}

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		start := time.Now()

		req, _ := http.NewRequest("PUT", "http://"+nodes[0]+"/keys/"+key, strings.NewReader("value"))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("write error: %v\n", err)
			continue
		}
		resp.Body.Close()

		latencies = append(latencies, time.Since(start))
	}

	printPercentiles(latencies)
}

func benchmarkReads() {
	fmt.Println("Benchmark Reads")
	latencies := []time.Duration{}

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		start := time.Now()

		req, _ := http.NewRequest("GET", "http://"+nodes[0]+"/keys/"+key, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("write error: %v\n", err)
			continue
		}
		resp.Body.Close()

		latencies = append(latencies, time.Since(start))
	}

	printPercentiles(latencies)
}

func benchmarkFailover() {
	fmt.Println("Kill node1 now, monitoring for failures...")
	var client = &http.Client{Timeout: 500 * time.Millisecond}

	var failStart time.Time
	inFailure := false
	recovered := false

	for i := 0; i < 5000; i++ {
		key := fmt.Sprintf("failover-key%d", i)
		req, _ := http.NewRequest("PUT", "http://"+nodes[1]+"/keys/"+key, strings.NewReader("value"))
		resp, err := client.Do(req)

		failed := err != nil || (resp != nil && resp.StatusCode == 502)
		if resp != nil {
			resp.Body.Close()
		}

		if failed && !inFailure {
			failStart = time.Now()
			inFailure = true
			fmt.Println("failure detected")
		}

		if !failed && inFailure && !recovered {
			fmt.Printf("Recovered after: %v\n", time.Since(failStart))
			recovered = true
		}

		time.Sleep(50 * time.Millisecond)
	}

	if !recovered && inFailure {
		fmt.Println("No recovery detected during test window")
	}

	if !inFailure {
		fmt.Println("No failure detected")
	}
}
