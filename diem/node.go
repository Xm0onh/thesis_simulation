package main

import (
	"encoding/json"
	"fmt"
	"log"
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
			go n.handleIncomingConnection(conn)
		}
	}()
}

func (n *Node) processChunkRequest(request *ChunkRequest, conn net.Conn) {
	block := n.generateBlockForRequest(request.BlockID)
	txs := block.Transactions
	fmt.Println("num of txs: ", len(txs))
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

		var message = &Message{
			From:    n.ID,
			To:      request.NodeID,
			Type:    "response",
			Content: response,
		}

		responseBytes, err := json.Marshal(message)
		if err != nil {
			fmt.Fprintf(conn, "Error marshaling response: %v", err)
			return
		}
		fmt.Println("Sending response to node", request.NodeID)
		_, err = conn.Write(responseBytes)
		if err != nil {
			log.Printf("Error writing response to connection: %v", err)
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func (n *Node) generateBlockForRequest(blockID int) *Block {

	txs := GenerateTransactions(1000)
	block := &Block{
		ID:           blockID,
		PreviousHash: "",
		Transactions: txs,
		Nonce:        0,
		Timestamp:    time.Now().Unix(),
		Hash:         "",
	}
	/// size of the block in byte
	blockBytes, _ := json.Marshal(block)
	fmt.Printf("Size of the block: %d bytes\n", len(blockBytes))
	block.Hash = GenerateBlockHash(*block)
	return block
}

func (n *Node) sendChunkRequest(blockID int) {
	request := &ChunkRequest{
		NodeID:         n.ID,
		BlockID:        blockID,
		TransactionIDs: []string{},
		RequestSize:    100,
	}

	message := &Message{
		From:    n.ID,
		To:      -1,
		Type:    "request",
		Content: request,
	}
	fmt.Println(request)

	requestData, err := json.Marshal(message)
	if err != nil {
		log.Fatalf("Error marshaling request: %v", err)
	}

	// selectedPeer := n.Peers[rand.Intn(len(n.Peers))]
	selectedPeer := n.Peers[2]
	conn, err := net.Dial("tcp", selectedPeer)
	if err != nil {
		log.Fatalf("Error connecting to peer: %v", err)
	}
	fmt.Println("my address: ", conn.LocalAddr().String())
	// defer conn.Close()
	_, err = conn.Write(requestData)
	if err != nil {
		log.Printf("Error writing response to connection: %v", err)
		return
	}

	n.readResponse(conn)
}

func (n *Node) handleChunkResponse(response *ChunkResponse) {

	if n.verifyChunk(response) {
		n.integrateChunk(response)
		fmt.Println("Chunk integrated successfully.")
	} else {
		fmt.Println("Failed to verify chunk.")
	}

}

func (n *Node) verifyChunk(response *ChunkResponse) bool {
	// TODO
	// Implement verification of chunk proof
	return true
}

func (n *Node) integrateChunk(response *ChunkResponse) {
	// TODO
	// Implement integration of chunk into blockchain
}
