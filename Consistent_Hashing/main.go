package main

import (
	"fmt"
	"hash/crc32"
	"slices"
	"strconv"
)

type HashRing struct {
	replicas int
	ring     map[uint32]string
	keys     []uint32
}

func main() {
	keys := generateKeys(100_000)

	// ---------- CONSISTENT HASHING ----------
	ring := NewHashRing(100)
	ring.AddNode("A")
	ring.AddNode("B")
	ring.AddNode("C")
	ring.AddNode("D")

	oldConsistent := make(map[string]string)
	for _, k := range keys {
		oldConsistent[k] = ring.GetNode(k)
	}

	// add one node
	ring.AddNode("E")

	movedConsistent := 0
	for _, k := range keys {
		if oldConsistent[k] != ring.GetNode(k) {
			movedConsistent++
		}
	}

	consistentPct := float64(movedConsistent) / float64(len(keys)) * 100

	// ---------- MODULO HASHING ----------
	nodes := []string{"A", "B", "C", "D"}

	oldModulo := make(map[string]string)
	for _, k := range keys {
		oldModulo[k] = getNodeModulo(k, nodes)
	}

	nodes = append(nodes, "E")

	movedModulo := 0
	for _, k := range keys {
		if oldModulo[k] != getNodeModulo(k, nodes) {
			movedModulo++
		}
	}

	moduloPct := float64(movedModulo) / float64(len(keys)) * 100

	// ---------- RESULTS ----------
	fmt.Printf("CONSISTENT HASHING movement: %.2f %%\n", consistentPct)
	fmt.Printf("MODULO HASHING movement:     %.2f %%\n", moduloPct)
}

// ---------------- CONSISTENT HASHING ----------------

func NewHashRing(replicas int) *HashRing {
	if replicas <= 0 {
		panic("replicas must be > 0")
	}
	return &HashRing{
		replicas: replicas,
		ring:     make(map[uint32]string),
		keys:     make([]uint32, 0),
	}
}

func (h *HashRing) AddNode(node string) {
	for i := 0; i < h.replicas; i++ {
		vnode := node + "#" + strconv.Itoa(i)
		hash := crc32.ChecksumIEEE([]byte(vnode))
		h.ring[hash] = node
		h.keys = append(h.keys, hash)
	}
	slices.Sort(h.keys)
}

func (h *HashRing) GetNode(key string) string {
	if len(h.keys) == 0 {
		panic("hash ring is empty")
	}

	hash := crc32.ChecksumIEEE([]byte(key))
	idx, _ := slices.BinarySearch(h.keys, hash)

	if idx == len(h.keys) {
		idx = 0
	}
	return h.ring[h.keys[idx]]
}

// ---------------- MODULO HASHING ----------------

func getNodeModulo(key string, nodes []string) string {
	hash := crc32.ChecksumIEEE([]byte(key))
	return nodes[int(hash)%len(nodes)]
}

// ---------------- UTIL ----------------

func generateKeys(n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = "key" + strconv.Itoa(i)
	}
	return keys
}
