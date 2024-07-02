package main

import (
	"fmt"
	"net"
	"time"

	"github.com/tendermint/tendermint/crypto/merkle"
)

// Assuming upload bandwidth is 10 Mbps - download bandwidth is 109 Mbps
// K = N - f | f = 10% of N
const (
	TXN_SIZE         = 1_000_000
	N                = 100      // size of each coded chunk is TXN_SIZE/K !!!
	BUFFER_SIZE      = 65536    // 2^16
	BANDWIDTH        = 12500000 // 10 Megabit per sec = 1.25 * 10^6 bytes per second
	UPLOAD_BANDWIDTH = 1250000
	// 8765437
	NETWORK_DELAY = 300 * time.Millisecond
	COUNTER       = 7 // Bandwdith / (F/N) -- F is the size of the file in bytes
)

var F = make(map[int]bool)

type Transaction struct {
	ID        string // Unique identifier for the transaction
	Content   string // Content or data of the transaction
	Signature string // Cryptographic signature to verify the transaction's authenticity
	Timestamp int64  // Unix timestamp for when the transaction was created
}

type Block struct {
	ID           int           // Unique identifier for the block
	Transactions []Transaction // List of transactions included in the block
	Hash         string        // Cryptographic hash of the block's contents
	PreviousHash string        // Hash of the previous block to ensure integrity in the chain
	Nonce        int           // Nonce used for mining/block creation
	Timestamp    int64         // Unix timestamp for when the block was created
}

type Node struct {
	ID             int            // Unique identifier for the node
	Address        string         // TCP address for the node
	Listener       net.Listener   // Listener for incoming connections
	Blockchain     []*Block       // Dynamic array of blocks representing the node's current blockchain
	Network        *Network       // Reference to the network for communications
	IsByzantine    bool           // Indicates whether the node exhibits Byzantine behavior
	Peers          map[int]string // Map of peer nodes for direct referencing and messaging
	BlockHeight    int            // Current height of the blockchain this node maintains
	ConsensusRole  string         // Role of the node in the consensus process, e.g., proposer, validator
	Metrics        *SyncMetrics   // Metrics for tracking synchronization performance
	BlackList      map[int]bool   // List of nodes to ignore during synchronization
	ReceivedChunks map[int]Chunk  // Map of received chunks indexed by their block ID

}

type Network struct {
	Nodes    map[int]*Node         // Map of all nodes indexed by their ID for quick access
	Latency  map[int]map[int]int   // Matrix to simulate network latency between nodes
	Channels map[int]chan *Message // Channels for node-to-node communication, mapped by node ID
}

type Message struct {
	From    int
	To      int
	Type    string
	Content interface{}
}

type ChunkRequest struct {
	NodeID  int // ID of the requesting node
	BlockID int // Identifier of the block from which data is needed
	ChunkID int // Identifier of the chunk within the block
}

type ChunkResponse struct {
	NodeID     int    // ID of the responding node
	Chunk      *Chunk // Data chunk with proof
	ChunkID    int    // Identifier of the chunk
	Commitment []byte // Vector commitment for the chunks
}

type Chunk struct {
	Data  []byte
	Proof merkle.Proof
}

type SyncMetrics struct {
	NodeID            int           // ID of the node for which metrics are being tracked
	StartTime         time.Time     // Time when the first chunk request was sent
	EndTime           time.Time     // Time when the last chunk was successfully verified and integrated
	TotalTransactions int           // Total number of transactions received
	TotalChunks       int           // Total number of chunks received
	SuccessfulChunks  int           // Number of successfully verified chunks
	FailedChunks      int           // Number of chunks that failed verification
	TotalDuration     time.Duration // Total time taken for the synchronization process
	VerificationTime  time.Duration // Time taken to verify all chunks
}

func main() {
	// Size of a single transaction in bytes]
	fmt.Printf("Size of a single transaction: %d bytes\n", SizeOfOneTransaction())
	// Size of the entire file in bytes
	fmt.Printf("Size of the entire file: %d bytes\n", SizeOfTheFile())
	// Maximum size of a chunk respected to the bandwidth
	fmt.Printf("Maximum size of a chunk: %d txs\n", TXN_SIZE/(SizeOfTheFile()/UPLOAD_BANDWIDTH))
	// Size of each coded chunk in bytes
	fmt.Printf("Size of each coded chunk: %d bytes\n", SizeOfTheFile()/N)
	// Maximum number of coded chunk respected to the bandwidth
	fmt.Printf("Maximum number of coded chunks: %d\n", BANDWIDTH/(SizeOfTheFile()/N))
	time.Sleep(10 * time.Second)
	faultyNodes := []int{1, 2, 3, 4, 5, 6, 7, 8}
	InitializeAdversary(faultyNodes)

	network := InitializeNetwork(N, 8000)
	for _, node := range network.Nodes {
		go node.Start()
	}
	time.Sleep(5 * time.Second)
	fallingBehindNode := network.Nodes[0]
	fallingBehindNode.sendChunkRequest(0)
	select {}
}
