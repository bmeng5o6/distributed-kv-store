package ring

import (
	"fmt"
	"testing"
)

// run: go test ./...

func TestRing_GetNodeBasic(t *testing.T) {
	r := NewRing()
	r.AddNode("serverA")
	r.AddNode("serverB")
	r.AddNode("serverC")

	node, err := r.GetNode("username")
	if err != nil {
		t.Fatal(err)
	}

	if node == "" {
		t.Errorf("expected node, got empty string")
	}
}

func TestRing_GetNodeConsistentWithSameKey(t *testing.T) {
	r := NewRing()
	r.AddNode("Server A")
	r.AddNode("Server B")
	r.AddNode("Server C")

	node1, err := r.GetNode("username")
	if err != nil {
		t.Fatal(err)
	}

	node2, err := r.GetNode("username")
	if err != nil {
		t.Fatal(err)
	}

	if node1 != node2 {
		t.Errorf("expected two nodes to equal, got %s then %s", node1, node2)
	}
}

func TestRing_SingleNodeAllHashToSameNode(t *testing.T) {
	r := NewRing()
	r.AddNode("Server A")

	node1, err := r.GetNode("username")
	if err != nil {
		t.Fatal(err)
	}

	node2, err := r.GetNode("password")
	if err != nil {
		t.Fatal(err)
	}

	node3, err := r.GetNode("email")
	if err != nil {
		t.Fatal(err)
	}

	if node1 != node2 || node2 != node3 {
		t.Errorf("expected all nodes to equal, got %s, %s, %s", node1, node2, node3)
	}
}

func TestRing_GetNodeEqualDistribution(t *testing.T) {
	r := NewRing()
	r.AddNode("serverA")
	r.AddNode("serverB")
	r.AddNode("serverC")

	counts := map[string]int{}
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		node, err := r.GetNode(key)
		if err != nil {
			t.Fatal(err)
		}

		counts[node]++
	}

	for node, count := range counts {
		t.Logf("node %s got %d/1000 keys", node, count)
		if count > 500 {
			t.Errorf("node %s got %d/1000 keys, too uneven", node, count)
		}
	}
}

func TestRing_EmptyRing(t *testing.T) {
	r := NewRing()

	_, err := r.GetNode("username")
	if err == nil {
		t.Errorf("expected error on empty ring, got nil")
	}
}
