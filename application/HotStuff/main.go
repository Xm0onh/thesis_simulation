package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Block struct {
	Header string
	Data   string
	Size   int // Total size of the block (header + data)
}

type Node struct {
	ID            int
	Blocks        map[int]Block
	Mutex         sync.Mutex
	Peers         []*Node
	Network       *Network
	F             int
	BandwidthUsed int // Track bandwidth usage
	Recovery      map[*Node]int
}

// Network simulates network conditions
type Network struct {
	Delay     time.Duration
	Bandwidth int // in Mbps
}

const (
	blockHSize           = 80              // Block header size in bytes
	blockBSize           = 1 * 1024 * 1024 // Block body size in bytes (1 MB)
	laggingNodeBandwidth = 3               // Bandwidth of the lagging node in number of blocks per second
)

// downloadBlockHeader simulates downloading the block header from peers
func (node *Node) downloadBlockHeader(blockID int) (string, error) {
	node.simulateNetworkConditions(blockHSize)

	for _, peer := range node.Peers {
		peer.Mutex.Lock()
		block, exists := peer.Blocks[blockID]
		peer.Mutex.Unlock()

		if exists {
			return block.Header, nil
		}
	}

	return "", fmt.Errorf("block header not found")
}

func (node *Node) downloadBlock(blockID int, fromNode *Node) (Block, error) {
	dataSize := blockHSize + blockBSize

	// Simulate the upload time from the peer node
	fromNode.simulateNetworkConditions(dataSize)
	fromNode.BandwidthUsed += dataSize

	fromNode.Mutex.Lock()
	block, exists := fromNode.Blocks[blockID]
	fromNode.Mutex.Unlock()

	if exists {
		return block, nil
	}

	return Block{}, fmt.Errorf("block not found")
}

func (node *Node) simulateNetworkConditions(dataSizeBytes int) {
	time.Sleep(node.Network.Delay)

	// Convert bandwidth from Mbps to KBps (1 Mbps = 125 KBps)
	bandwidthKBps := node.Network.Bandwidth * 125
	dataSizeKB := dataSizeBytes / 1024
	transferTime := time.Duration(dataSizeKB*1000/bandwidthKBps) * time.Millisecond
	time.Sleep(transferTime)
}

func (node *Node) verifyBlock() bool {
	verifyTime := time.Duration(rand.Intn(100)+50) * time.Millisecond // Random verification time between 50-150 ms
	time.Sleep(verifyTime)
	return true
}

func simulateBlockRecovery(node *Node, missingBlocks []int, numNodes int, numMissingBlocks int) {
	if 3*node.F+1 > len(node.Peers) {
		fmt.Printf("Not enough nodes to tolerate %d faulty nodes\n", node.F)
		return
	}

	start := time.Now()
	var wg sync.WaitGroup
	sem := make(chan struct{}, laggingNodeBandwidth) // Semaphore to limit concurrent downloads

	for i, blockID := range missingBlocks {
		sem <- struct{}{} // Acquire a slot
		wg.Add(1)
		peerIndex := i % (numNodes - 1)

		go func(blockID int) {
			defer wg.Done()
			defer func() { <-sem }() // Release the slot

			// Step 1: Download block header
			_, err := node.downloadBlockHeader(blockID)
			if err != nil {
				fmt.Printf("Failed to download block header: %v\n", err)
				return
			}

			// Step 2: Download the full block from a random peer
			block, err := node.downloadBlock(blockID, node.Peers[peerIndex])
			if err != nil {
				fmt.Printf("Failed to download block: %v\n", err)
				return
			}

			// Step 3: Verify the block from at least 2f+1 nodes
			verificationCount := 0
			for _, peer := range node.Peers {
				if peer.verifyBlock() {
					verificationCount++
					if verificationCount == numMissingBlocks {
						break
					}
				}
			}

			if verificationCount >= 2*node.F+1 {
				// Store the recovered block
				node.Mutex.Lock()
				node.Blocks[blockID] = block
				node.Mutex.Unlock()
			}
		}(blockID)
	}

	wg.Wait()

	duration := time.Since(start)
	fmt.Printf("Recovered %d blocks in %v\n", len(missingBlocks), duration)
}

func initializeNetwork(delay time.Duration, bandwidth int) *Network {
	return &Network{
		Delay:     delay,
		Bandwidth: bandwidth,
	}
}

func initializeNodes(numNodes, f int, network *Network) []*Node {
	nodes := make([]*Node, numNodes)
	for i := 0; i < numNodes; i++ {
		nodes[i] = &Node{
			ID:            i,
			Blocks:        make(map[int]Block),
			Network:       network,
			F:             f,
			BandwidthUsed: 0,
			Recovery:      make(map[*Node]int), // Initialize the Recovery map
		}
	}
	return nodes
}

func populateBlocks(nodes []*Node, numBlocks int) {
	for i := 0; i < len(nodes); i++ {
		for j := 0; j < numBlocks; j++ {
			blockID := j + 1
			nodes[i].Blocks[blockID] = Block{
				Header: fmt.Sprintf("Header %d", blockID),
				Data:   fmt.Sprintf("Data %d", blockID),
				Size:   blockHSize + blockBSize,
			}
		}
	}
}

func connectNodes(nodes []*Node) {
	for i := 0; i < len(nodes); i++ {
		for j := 0; j < len(nodes); j++ {
			if i != j {
				nodes[i].Peers = append(nodes[i].Peers, nodes[j])
				nodes[i].Recovery[nodes[j]] = 0
			}
		}
	}
}

func missingBlocks(n int) []int {
	blocks := make([]int, n)
	for i := 0; i < n; i++ {
		blocks[i] = rand.Intn(100) + 1
	}
	return blocks
}
func main() {
	nodeCounts := []int{10, 25, 35, 50, 60, 80, 100}

	for _, numNodes := range nodeCounts {
		var fValues []int
		switch {
		case numNodes <= 10:
			fValues = []int{0, 1, 2}
		case numNodes <= 25:
			fValues = []int{0, 1, 5, 7}
		case numNodes <= 50:
			fValues = []int{0, 1, 5, 10, 15}
		case numNodes <= 100:
			fValues = []int{0, 1, 10, 20, 30}
		}

		for _, f := range fValues {
			if 3*f+1 > numNodes {
				continue
			}

			for numMissingBlocks := 5; numMissingBlocks < numNodes; numMissingBlocks += 5 {
				experiment(numNodes, f, numMissingBlocks)
			}
		}
	}
}

func experiment(numNodes, f, numMissingBlocks int) {
	fmt.Printf("Running experiment with %d nodes, f = %d, %d missing blocks\n", numNodes, f, numMissingBlocks)
	network := initializeNetwork(50*time.Millisecond, 10) // 50 ms network delay, 10 Mbps bandwidth

	nodes := initializeNodes(numNodes, f, network)
	populateBlocks(nodes, 100)
	connectNodes(nodes)

	simulateBlockRecovery(nodes[0], missingBlocks(numMissingBlocks), numNodes, numMissingBlocks)
}
