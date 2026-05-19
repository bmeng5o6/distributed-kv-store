package ring

import (
	"fmt"
	"hash/fnv"
	"sort"
)

type Ring struct {
	positions []uint32
	nodes     map[uint32]string
}

func NewRing() *Ring {
	return &Ring{
		nodes: make(map[uint32]string),
	}
}

func hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (r *Ring) AddNode(name string) {
	// 50 virtual nodes to improve balance
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("%s-%d", name, i)
		index := hash(key)
		r.nodes[index] = name

		// Find where to put index in r.positions
		j := sort.Search(len(r.positions), func(j int) bool {
			return r.positions[j] >= index
		})

		r.positions = append(r.positions, 0)
		copy(r.positions[j+1:], r.positions[j:])
		r.positions[j] = index
	}
}

func (r *Ring) GetNode(key string) (string, error) {
	if len(r.positions) == 0 {
		return "", fmt.Errorf("ring is empty")
	}

	index := hash(key)
	i := sort.Search(len(r.positions), func(i int) bool {
		return r.positions[i] >= index
	})

	if i == len(r.positions) {
		i = 0
	}

	return r.nodes[r.positions[i]], nil
}
