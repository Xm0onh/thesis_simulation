package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
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

func (n *Node) sendChunkRequest(blockID int) {

	n.Metrics = &SyncMetrics{
		NodeID:           n.ID,
		StartTime:        time.Now(),
		VerificationTime: 0,
		TotalChunks:      0,
		SuccessfulChunks: 0,
		FailedChunks:     0,
	}
	wg := sync.WaitGroup{}

	for i := 1; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			request := &ChunkRequest{
				NodeID:  n.ID,
				BlockID: blockID,
			}
	
			message := &Message{
				From:    n.ID,
				To:      -1,
				Type:    "request",
				Content: request,
			}
	
			requestData, err := json.Marshal(message)
			if err != nil {
				log.Fatalf("Error marshaling request: %v", err)
			}
			selectedPeer := n.Peers[i]
			fmt.Println("Sending request to peer", i)
			conn, err := net.Dial("tcp", selectedPeer)
			if err != nil {
				log.Fatalf("Error connecting to peer: %v", err)
			}
			// defer conn.Close()
			_, err = conn.Write(requestData)
			if err != nil {
				log.Printf("Error writing response to connection: %v", err)
				return
			}

			n.readResponse(conn)
		}(i)

	}
	wg.Wait()

	// Update and display metrics
	n.Metrics.EndTime = time.Now()
	n.Metrics.TotalDuration = n.Metrics.EndTime.Sub(n.Metrics.StartTime)
	fmt.Printf("Sync Metrics for Node %d: %+v\n", n.ID, n.Metrics)
}

func (n *Node) processChunkRequest(request *ChunkRequest, conn net.Conn) {
	block := n.generateBlockForRequest(request.BlockID)
	blockBytes, _ := json.Marshal(block)
	chunks := GenerateCodedChunks(blockBytes)

	response := ChunkResponse{
		NodeID: n.ID,
		Chunk:  &chunks[n.ID],
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

}

func (n *Node) generateBlockForRequest(blockID int) *Block {

	txs := GenerateTransactions(TXN_SIZE)
	block := &Block{
		ID:           blockID,
		PreviousHash: "",
		Transactions: txs,
		Nonce:        0,
		Timestamp:    time.Now().Unix(),
		Hash:         "",
	}
	/// size of the block in byte
	// blockBytes, _ := json.Marshal(block)
	// fmt.Printf("Size of the block: %d bytes\n", len(blockBytes))

	block.Hash = GenerateBlockHash(*block)
	return block
}

func (n *Node) handleChunkResponse(response *ChunkResponse) {

	n.ReceivedChunks[response.NodeID] = *response.Chunk
	fmt.Println("Total chunks received:", len(n.ReceivedChunks))
	if len(n.ReceivedChunks) == K {
		fmt.Println("Enough chunks recieved")
		decodedMessage, err := Decode(n.ReceivedChunks)
		if err != nil {
			log.Fatalf("Error decoding message: %v", err)
		}
		fmt.Println("Decoded message:", len(decodedMessage))
	}
	// n.Metrics.TotalChunks++
	// if n.verifyChunk(response) {
	// 	n.Metrics.SuccessfulChunks++
	// 	n.integrateChunk(response)
	// 	fmt.Println("Chunk integrated successfully.")
	// } else {
	// 	n.Metrics.FailedChunks++
	// 	fmt.Println("Failed to verify chunk.")
	// 	// TODO: Implement retry logic

	// }

}
