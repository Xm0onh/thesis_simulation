package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type CodedChunk struct {
	ChunkID   int
	OrgChunks []int
	Data      string
}

type Block struct {
	Header string
	Data   string
	Size   int // Total size of the block (header + data)
}

type Node struct {
	ID      int
	Blocks  map[int]Block
	Chunk   CodedChunk
	Mutex   sync.Mutex
	Peers   []*Node
	Network *Network
	F       int
}

// Network simulates network conditions
type Network struct {
	Delay     time.Duration
	Bandwidth int // in Mbps
}

// Constants representing block header and body sizes in bytes
const (
	blockHSize               = 508 // Block header size in bytes
	CommitmentSize           = 24
	blockHSizeWithCommitment = blockHSize + CommitmentSize
	proofSize                = 64
	blockBSize               = 1000 * 1024 // Block body size in bytes (1000 KB)
	numberOfChunks           = 10
	chunkSize                = blockBSize / numberOfChunks
)

// downloadBlockHeader simulates downloading the block header from peers
func (node *Node) downloadBlockHeader(blockID int) (string, error) {
	node.simulateNetworkConditions(blockHSizeWithCommitment)

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

func (node *Node) downloadChunk(blockID int) (CodedChunk, error) {
	fmt.Println("Downloading chunk ", blockID)
	node.simulateNetworkConditions(chunkSize + proofSize)

	for _, peer := range node.Peers {
		peer.Mutex.Lock()
		chunk := peer.Chunk
		peer.Mutex.Unlock()
		return chunk, nil
	}

	return CodedChunk{}, fmt.Errorf("chunk not found")
}

// simulateNetworkConditions simulates network delay and bandwidth constraints
func (node *Node) simulateNetworkConditions(dataSizeBytes int) {
	time.Sleep(node.Network.Delay)

	// Convert bandwidth from Mbps to KBps (1 Mbps = 125 KBps)
	bandwidthKBps := node.Network.Bandwidth * 125
	dataSizeKB := dataSizeBytes / 1024
	transferTime := time.Duration(dataSizeKB*1000/bandwidthKBps) * time.Millisecond
	time.Sleep(transferTime)
}

// verifyBlock simulates the time taken to verify a block
func (node *Node) verifyBlock() bool {
	verifyTime := time.Duration(rand.Intn(100)+50) * time.Millisecond // Random verification time between 50-150 ms
	time.Sleep(verifyTime)
	return true
}

func simulateBlockRecovery(node *Node, missingBlocks []int) {
	if 3*node.F+1 > len(node.Peers) {
		fmt.Printf("Not enough nodes to tolerate %d faulty nodes\n", node.F)
		return
	}

	start := time.Now()

	for _, blockID := range missingBlocks {
		// Step 1: Download block header
		header, err := node.downloadBlockHeader(blockID)
		if err != nil {
			fmt.Printf("Failed to download block header: %v\n", err)
			return
		}

		// Step 2: Download the full block
		codedChunk, err := node.downloadChunk(blockID)
		if err != nil {
			fmt.Printf("Failed to download block: %v\n", err)
			return
		}

		// Step 3: Verify the block from at least 2f+1 nodes
		verificationCount := 0
		for _, peer := range node.Peers {
			if peer.verifyBlock() {
				verificationCount++
				if verificationCount >= 2*node.F+1 {
					break
				}
			}
		}

		if verificationCount >= 2*node.F+1 {
			// Store the recovered block
			node.Mutex.Lock()
			node.Chunk = codedChunk
			node.Mutex.Unlock()
			fmt.Printf("Recovered and verified Chunk %d: header = %s, data = %s\n", blockID, header, codedChunk.Data)
		} else {
			fmt.Printf("Failed to verify block %d from 2f+1 nodes\n", blockID)
		}
	}

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
			ID:      i,
			Blocks:  make(map[int]Block),
			Network: network,
			F:       f,
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
			nodes[i].Chunk = CodedChunk{
				ChunkID:   blockID,
				OrgChunks: []int{blockID},
				Data:      fmt.Sprintf("Data %d", blockID),
			}
		}
	}
}

func connectNodes(nodes []*Node) {
	for i := 0; i < len(nodes); i++ {
		for j := 0; j < len(nodes); j++ {
			if i != j {
				nodes[i].Peers = append(nodes[i].Peers, nodes[j])
			}
		}
	}
}

func main() {
	network := initializeNetwork(50*time.Millisecond, 10) // 50 ms network delay, 10 Mbps bandwidth
	numNodes := 10
	f := 1

	nodes := initializeNodes(numNodes, f, network)
	populateBlocks(nodes, 5)
	connectNodes(nodes)

	simulateBlockRecovery(nodes[0], []int{1, 2, 3})
}
