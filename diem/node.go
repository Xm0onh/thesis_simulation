package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

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

func (n *Node) processChunkRequest(request *ChunkRequest, conn net.Conn) {
	block := n.generateBlockForRequest(request.BlockID)
	txs := block.Transactions
	for i := 0; i < len(block.Transactions); i += request.RequestSize {
		// TODO
		// Implement proof generation for chunk of transactions

		// Serialize and send the response
		response := ChunkResponse{
			NodeID:       n.ID,
			BlockID:      block.ID,
			Transactions: txs[i:min(i+request.RequestSize, len(txs))],
			Proof:        []byte{}, // TODO: Implement proof generation
			Success:      true,
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			fmt.Fprintf(conn, "Error marshaling response: %v", err)
			return
		}
		conn.Write(responseBytes)
		time.Sleep(100 * time.Millisecond)
	}
}

func (n *Node) generateBlockForRequest(blockID int) *Block {

	txs := GenerateTransactions(100)
	block := &Block{
		ID:           blockID,
		PreviousHash: "",
		Transactions: txs,
		Nonce:        0,
		Timestamp:    time.Now().Unix(),
		Hash:         "",
	}
	block.Hash = GenerateBlockHash(*block)
	return block
}

func (n *Node) sendChunkRequest(blockID int) {
	request := &ChunkRequest{
		NodeID:         n.ID,
		BlockID:        blockID,
		TransactionIDs: []string{},
		RequestSize:    10,
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("Error marshaling request: %v", err)
	}

	selectedPeer := n.Peers[rand.Intn(len(n.Peers))]

	conn, err := net.Dial("tcp", selectedPeer)
	if err != nil {
		log.Fatalf("Error connecting to peer: %v", err)
	}
	defer conn.Close()
	conn.Write(requestData)
}
