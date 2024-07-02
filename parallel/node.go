package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/exp/rand"
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
		if (i%COUNTER == 0) && (i != 0) {
			time.Sleep(1 * time.Second)
		}
		go func(i int) {
			defer wg.Done()
			request := &ChunkRequest{
				NodeID:  n.ID,
				BlockID: blockID,
				ChunkID: i,
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
			readDeadline := time.Now().Add(20 * time.Second)
			conn.SetReadDeadline(readDeadline)
			if !n.readResponse(conn, i) {
				fmt.Println("Failed to read response from peer", i)
				fmt.Println("Blacklisting peer", n.BlackList)
				conn.Close()
				for {
					index := n.selectRandomPeer()
					conn, err = net.Dial("tcp", n.Peers[index])
					if err != nil {
						log.Fatalf("Error connecting to peer: %v", err)
					}
					_, err = conn.Write(requestData)
					if err != nil {
						log.Printf("Error writing response to connection: %v", err)
						return
					}
					if n.readResponse(conn, i) {
						break
					} else {
						continue
					}
				}

			}
		}(i)

	}
	wg.Wait()

	// Update and display metrics

}

func (n *Node) processChunkRequest(request *ChunkRequest, conn net.Conn) {
	block := n.generateBlockForRequest(request.BlockID)
	blockBytes, _ := json.Marshal(block)
	fmt.Println("Size of block in bytes: ", len(blockBytes))
	chunks := GenerateDataChunks(blockBytes)
	rootHash, proofs := CreateVectorCommitment(chunks)
	if n.IsByzantine {
		// DOINT NOTHING
	} else {
		for i := range chunks {
			chunks[i].Proof = *proofs[i]
		}
		fmt.Println("Sending chunk id", request.ChunkID, "to node", request.NodeID)
		response := ChunkResponse{
			NodeID:     n.ID,
			Chunk:      &chunks[request.ChunkID],
			ChunkID:    request.ChunkID,
			Commitment: rootHash,
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
		time.Sleep(NETWORK_DELAY)
		// Upload bandwidth limitation:
		bandwidthLimit := len(responseBytes) / UPLOAD_BANDWIDTH
		fmt.Println("Bandwidth limit: ", bandwidthLimit)
		// time.Sleep(time.Duration(bandwidthLimit) * time.Second)
		_, err = conn.Write(responseBytes)
		if err != nil {
			log.Printf("Error writing response to connection: %v", err)
			return
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
	// blockBytes, _ := json.Marshal(block)
	// fmt.Printf("Size of the block: %d bytes\n", len(blockBytes))

	block.Hash = GenerateBlockHash(*block)
	return block
}

func (n *Node) handleChunkResponse(response *ChunkResponse) {

	n.Metrics.TotalChunks++
	// fmt.Println("response commtiment: ", response.Commitment, "proof", response.Chunk.Proof)
	if VerifyChunk(response.Commitment, *response.Chunk, &response.Chunk.Proof, n) {
		n.Metrics.SuccessfulChunks++
		n.ReceivedChunks[response.ChunkID] = *response.Chunk
		fmt.Println("Chunk integrated successfully.")
		fmt.Println("Number of received chunks ", len(n.ReceivedChunks))
	} else {
		n.Metrics.FailedChunks++
		fmt.Println("Failed to verify chunk.")
	}
	if len(n.ReceivedChunks) == (N - 1) {
		fmt.Println("Enough chunks recieved, size of chunk is", len(n.ReceivedChunks[0].Data))
		decodedMessage, err := Decode(n.ReceivedChunks)
		if err != nil {
			log.Fatalf("Error decoding message: %v", err)
		}
		fmt.Println("Decoded message:", len(decodedMessage))
		n.Metrics.EndTime = time.Now()
		n.Metrics.TotalDuration = n.Metrics.EndTime.Sub(n.Metrics.StartTime)
		fmt.Printf("Sync Metrics for Node %d: %+v\n", n.ID, n.Metrics)
		panic("Sync complete")
		// size of message in byte
	}
}

func (n *Node) selectRandomPeer() int {
	availablePeers := []int{}
	for peerID := range n.Peers {
		if !n.BlackList[peerID] && peerID != n.ID {
			availablePeers = append(availablePeers, peerID)
		}
	}
	fmt.Println("Blacklisted peers: ", n.BlackList)

	if len(availablePeers) == 0 {
		return -1
	}

	rand.Seed(uint64(time.Now().UnixNano()))
	return availablePeers[rand.Intn(len(availablePeers))]
}
