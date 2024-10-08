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
	ID            int
	Blocks        map[int]Block
	Chunks        map[int]CodedChunk
	Mutex         sync.Mutex
	Peers         []*Node
	Network       *Network
	F             int
	BandwidthUsed int // Track bandwidth usage
}

// Network simulates network conditions
type Network struct {
	Delay     time.Duration
	Bandwidth int // in Mbps
}

// Constants representing block header and body sizes in bytes
const (
	blockHSize               = 80 // Block header size in bytes
	CommitmentSize           = 24
	blockHSizeWithCommitment = blockHSize + CommitmentSize
	proofSize                = 64
	blockBSize               = 1 * 1024 * 1024 // Block body size in bytes (1 MB)
	numberOfChunks           = 10
	chunkSize                = blockBSize / numberOfChunks
	laggingNodeBandwidth     = 3 * numberOfChunks // Bandwidth of the lagging node in number of chunks per second
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

func (node *Node) downloadChunk(blockID, chunkID int, fromNode *Node) (CodedChunk, error) {
	dataSize := chunkSize + proofSize

	// Simulate the upload time from the peer node
	fromNode.simulateNetworkConditions(dataSize)
	fromNode.BandwidthUsed += dataSize

	fromNode.Mutex.Lock()
	_, exists := fromNode.Blocks[blockID]
	fromNode.Mutex.Unlock()

	if exists {
		return CodedChunk{
			ChunkID:   chunkID,
			OrgChunks: []int{chunkID},
			Data:      fmt.Sprintf("Chunk data %d from block %d", chunkID, blockID),
		}, nil
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

func simulateBlockRecovery(node *Node, missingBlocks []int, numNodes int, numMissingBlocks int) {
	if 3*node.F+1 > len(node.Peers) {
		fmt.Printf("Not enough nodes to tolerate %d faulty nodes\n", node.F)
		return
	}

	start := time.Now()
	var wg sync.WaitGroup
	sem := make(chan struct{}, laggingNodeBandwidth)
	totalChunks := 0

	for _, blockID := range missingBlocks {
		for chunkID := 0; chunkID < numberOfChunks; chunkID++ {
			sem <- struct{}{}
			wg.Add(1)
			peerIndex := (blockID + chunkID) % (numNodes - 1)

			go func(blockID, chunkID int) {
				defer wg.Done()
				defer func() { <-sem }() // Release the slot

				// Step 1: Download block header
				_, err := node.downloadBlockHeader(blockID)
				if err != nil {
					fmt.Printf("Failed to download block header: %v\n", err)
					return
				}

				// Step 2: Download the chunk from a peer
				codedChunk, err := node.downloadChunk(blockID, chunkID, node.Peers[peerIndex])
				if err != nil {
					fmt.Printf("Failed to download chunk: %v\n", err)
					return
				}

				// Step 3: Verify the chunk from at least 2f+1 nodes
				verificationCount := 0
				for _, peer := range node.Peers {
					if peer.verifyBlock() {
						verificationCount++
						if verificationCount > node.F+1 {
							break
						}
					}
				}

				if verificationCount >= node.F+1 {
					// Store the recovered chunk
					node.Mutex.Lock()
					node.Chunks[chunkID] = codedChunk
					node.Mutex.Unlock()
					// fmt.Printf("Recovered and verified Chunk %d: header = %s, data = %s\n", blockID, header, codedChunk.Data)
				}

				node.Mutex.Lock()
				totalChunks++
				node.Mutex.Unlock()
			}(blockID, chunkID)
		}
	}

	wg.Wait()

	duration := time.Since(start)
	fmt.Printf("Recovered %d blocks in %v\n", len(missingBlocks), duration)
	fmt.Printf("Recovered %d chunks for %d blocks in %v\n", totalChunks, len(missingBlocks), duration)
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
			Chunks:        make(map[int]CodedChunk),
			Network:       network,
			F:             f,
			BandwidthUsed: 0,
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
			for k := 0; k < numberOfChunks; k++ {
				nodes[i].Chunks[k] = CodedChunk{
					ChunkID:   k,
					OrgChunks: []int{k},
					Data:      fmt.Sprintf("Chunk data %d", k),
				}
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

func missingBlocks(n int) []int {
	blocks := make([]int, n)
	for i := 0; i < n; i++ {
		blocks[i] = rand.Intn(100) + 1
	}
	return blocks
}

func main() {
	nodeCounts := []int{100}
	for _, numNodes := range nodeCounts {
		var fValues []int
		switch {
		case numNodes <= 100:
			fValues = []int{0, 10, 20, 30}
		}

		for _, f := range fValues {
			if 3*f+1 > numNodes {
				continue
			}

			for numMissingBlocks := 5; numMissingBlocks < numNodes; numMissingBlocks += 20 {
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
