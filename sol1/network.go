package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
)

func InitializeNetwork(numNodes int, startingPort int) *Network {
	network := &Network{
		Nodes: make(map[int]*Node),
	}
	for i := 0; i < N; i++ {
		address := fmt.Sprintf("localhost:%d", startingPort+i)
		listener, err := net.Listen("tcp", address)
		byzantine := false
		if err != nil {
			log.Fatalf("Error starting listener: %v", err)
		}
		fmt.Println("Node", i, "listening on", address)
		if (i%10) == 0 && i != 0 {
			fmt.Println("Node", i, "is Byzantine")
			byzantine = true
		}
		node := &Node{
			ID:             i,
			Address:        address,
			Listener:       listener,
			Peers:          make(map[int]string),
			Blockchain:     make([]*Block, 0),
			IsByzantine:    byzantine,
			BlackList:      make(map[int]bool),
			ReceivedChunks: make(map[int]Chunk),
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

func (n *Node) handleIncomingConnection(conn net.Conn) {
	fmt.Println("Node", n.ID, "handling connection from", conn.RemoteAddr().String())
	defer conn.Close()
	var buf [4096]byte

	for {
		length, err := conn.Read(buf[:])
		if err != nil {
			if err != io.EOF {
				log.Printf("Node %d error reading from connection: %v", n.ID, err)
			}
			break
		}

		// fmt.Printf("Node %d read %d bytes: %s\n", n.ID, length, string(buf[:length])) // Log raw data for debugging

		var message Message
		if err := json.Unmarshal(buf[:length], &message); err != nil {
			log.Printf("Node %d error unmarshaling message: %v", n.ID, err)
			continue
		}
		fmt.Println("Node", n.ID, "received message:", message)

		n.handleMessage(message, conn)
	}
}

func (n *Node) handleMessage(message Message, conn net.Conn) {

	switch message.Type {
	case "request":
		var request ChunkRequest
		contentBytes, err := json.Marshal(message.Content)
		if err != nil {
			log.Printf("Error marshaling content: %v", err)
			return
		}
		if err := json.Unmarshal(contentBytes, &request); err != nil {
			log.Printf("Error unmarshaling chunk request: %v", err)
			return
		}

		fmt.Println("Received chunk request from node", request.NodeID)
		n.processChunkRequest(&request, conn)
	}
}

func (n *Node) readResponse(conn net.Conn) {
	var buf [4096]byte
	var accumulatedData bytes.Buffer

	for {
		length, err := conn.Read(buf[:])
		if err != nil {
			if err != io.EOF {
				log.Printf("Node %d error reading response: %v", n.ID, err)
			}
			break
		}

		// fmt.Printf("Node %d read %d bytes\n", n.ID, length)
		accumulatedData.Write(buf[:length])

		// Check if we have a full JSON object
		data := accumulatedData.Bytes()
		if closingIndex := findClosingBrace(data); closingIndex != -1 {
			fullMessage := data[:closingIndex+1]
			accumulatedData.Next(closingIndex + 1) // Remove processed bytes

			// fmt.Printf("Node %d received full message: %s\n", n.ID, string(fullMessage))

			var message Message
			if err := json.Unmarshal(fullMessage, &message); err != nil {
				log.Printf("Node %d error unmarshaling response: %v", n.ID, err)
				continue
			}
			// fmt.Printf("Node %d received message: %v\n", n.ID, message)

			if message.Type == "response" {
				var response ChunkResponse
				contentBytes, err := json.Marshal(message.Content)
				if err != nil {
					log.Printf("Node %d error marshaling content: %v", n.ID, err)
					continue
				}
				if err := json.Unmarshal(contentBytes, &response); err != nil {
					log.Printf("Node %d error unmarshaling chunk response: %v", n.ID, err)
					continue
				}
				n.handleChunkResponse(&response)
				break
			}
		}
	}
}

// Helper function to find the closing brace of a JSON object
func findClosingBrace(data []byte) int {
	openBraces := 0
	for i, b := range data {
		if b == '{' {
			openBraces++
		} else if b == '}' {
			openBraces--
			if openBraces == 0 {
				return i
			}
		}
	}
	return -1
}
