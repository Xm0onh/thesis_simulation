package main

import (
	"crypto/sha256"
	"encoding/hex"
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

func (n *Node) sendChunkRequest(blockID int) {

	n.Metrics = &SyncMetrics{
		NodeID:           n.ID,
		StartTime:        time.Now(),
		VerificationTime: 0,
		TotalChunks:      0,
		SuccessfulChunks: 0,
		FailedChunks:     0,
	}

	request := &ChunkRequest{
		NodeID:         n.ID,
		BlockID:        blockID,
		TransactionIDs: []string{},
		RequestSize:    CHUNK_SIZE,
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
	// Update and display metrics
	n.Metrics.EndTime = time.Now()
	n.Metrics.TotalDuration = n.Metrics.EndTime.Sub(n.Metrics.StartTime)
	fmt.Printf("Sync Metrics for Node %d: %+v\n", n.ID, n.Metrics)
}

func (n *Node) processChunkRequest(request *ChunkRequest, conn net.Conn) {
	block := n.generateBlockForRequest(request.BlockID)
	txs := block.Transactions
	fmt.Println("num of txs: ", len(txs))
	if n.IsByzantine {
		// DOING NOTHING
	} else {
		for i := 0; i < len(block.Transactions); i += request.RequestSize {
			chunk := txs[i:min(i+request.RequestSize, len(txs))]
			proof, _ := generateProof(chunk)

			// Serialize and send the response
			response := ChunkResponse{
				NodeID:       n.ID,
				BlockID:      block.ID,
				Transactions: txs[i:min(i+request.RequestSize, len(txs))],
				Proof:        proof.LeftSiblings,
				Success:      true,
			}

			var message = &Message{
				From:    n.ID,
				To:      request.NodeID,
				Type:    "response",
				Content: response,
			}
			if (i + request.RequestSize) >= len(txs) {
				message.Type = "last_response"
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
			time.Sleep(1 * time.Second)
			time.Sleep(NETWORK_DELAY)
		}
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
	blockBytes, _ := json.Marshal(block)
	fmt.Printf("Size of the block: %d bytes\n", len(blockBytes))
	block.Hash = GenerateBlockHash(*block)
	return block
}

func generateProof(transactions []Transaction) (TransactionAccumulatorRangeProof, []TransactionInfo) {
	var leftSiblings []string
	var transactionInfos []TransactionInfo

	for _, txn := range transactions {
		hash := sha256.Sum256([]byte(txn.ID))
		leftSiblings = append(leftSiblings, hex.EncodeToString(hash[:]))
		transactionInfos = append(transactionInfos, TransactionInfo{TransactionHash: hex.EncodeToString(hash[:])})
	}

	return TransactionAccumulatorRangeProof{LeftSiblings: leftSiblings}, transactionInfos
}

func verifyProof(transactions []Transaction, proof TransactionAccumulatorRangeProof) bool {
	for i, txn := range transactions {
		hash := sha256.Sum256([]byte(txn.ID))
		if proof.LeftSiblings[i] != hex.EncodeToString(hash[:]) {
			return false
		}
	}
	return true
}

func (n *Node) handleChunkResponse(response *ChunkResponse) {
	n.Metrics.TotalChunks++
	if n.verifyChunk(response) {
		n.Metrics.SuccessfulChunks++
		n.integrateChunk(response)
		fmt.Println("Chunk integrated successfully.")
	} else {
		n.Metrics.FailedChunks++
		fmt.Println("Failed to verify chunk.")
		// TODO: Implement retry logic

	}

}

func (n *Node) verifyChunk(response *ChunkResponse) bool {
	startTime := time.Now()
	proof := TransactionAccumulatorRangeProof{LeftSiblings: response.Proof}
	n.Metrics.VerificationTime += time.Since(startTime)
	return verifyProof(response.Transactions, proof)
}

func (n *Node) integrateChunk(response *ChunkResponse) {
	// TODO
	// Implement integration of chunk into blockchain
}
