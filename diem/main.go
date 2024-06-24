package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

type Transaction struct {
	ID        string // Unique identifier for the transaction
	Content   string // Content or data of the transaction
	Signature string // Cryptographic signature to verify the transaction's authenticity
	Timestamp int64  // Unix timestamp for when the transaction was created
}

type Block struct {
	ID           string        // Unique identifier for the block
	Transactions []Transaction // List of transactions included in the block
	Hash         string        // Cryptographic hash of the block's contents
	PreviousHash string        // Hash of the previous block to ensure integrity in the chain
	Nonce        int           // Nonce used for mining/block creation
	Timestamp    int64         // Unix timestamp for when the block was created
}

type Node struct {
	ID            int    // Unique identifier for the node
	Address       string // TCP address for the node
	Listener      net.Listener
	Blockchain    []*Block       // Dynamic array of blocks representing the node's current blockchain
	Network       *Network       // Reference to the network for communications
	IsByzantine   bool           // Indicates whether the node exhibits Byzantine behavior
	Peers         map[int]string // Map of peer nodes for direct referencing and messaging
	BlockHeight   int            // Current height of the blockchain this node maintains
	ConsensusRole string         // Role of the node in the consensus process, e.g., proposer, validator
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
	NodeID         int      // ID of the requesting node
	BlockID        string   // Identifier of the block from which data is needed
	TransactionIDs []string // Specific transaction IDs requested, if not all transactions are needed
	RequestSize    int      // Number of transactions requested in each chunk
}

type ChunkResponse struct {
	NodeID       int           // ID of the responding node
	BlockID      string        // Identifier of the block from which the chunk is derived
	Transactions []Transaction // The transactions that are included in the chunk
	Proof        []byte        // Cryptographic proof validating the transactions
	Success      bool          // Indicates if the response is successfully processed or not
	ErrorMessage string        // In case of an error, contains the error message
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
}

func InitializeNetwork(numNodes int, startingPort int) *Network {
	network := &Network{
		Nodes: make(map[int]*Node),
	}
	for i := 0; i < numNodes; i++ {
		address := fmt.Sprintf("localhost:%d", startingPort+i)
		listener, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatalf("Error starting listener: %v", err)
		}
		fmt.Println("Node", i, "listening on", address)
		node := &Node{
			ID:          i,
			Address:     address,
			Listener:    listener,
			Peers:       make(map[int]string),
			Blockchain:  make([]*Block, 0),
			IsByzantine: false,
		}
		network.Nodes[i] = node
	}

	for id, node := range network.Nodes {
		for peerID, peerNode := range network.Nodes {
			if id != peerID {
				node.Peers[peerID] = peerNode.Address
			}
		}
	}
	return network
}

func (n *Node) Start() {
	go func() {
		for {
			conn, err := n.Listener.Accept()
			if err != nil {
				log.Fatalf("Error accepting connection: %v", err)
				continue
			}
			go n.handleConnection(conn)
		}
	}()
}

func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()
}

func (n *Node) sendData(peerID int, data []byte) error {
	conn, err := net.Dial("tcp", n.Peers[peerID])
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	network := InitializeNetwork(5, 8000)
	for _, node := range network.Nodes {
		node.Start()
	}
	select {}
}
