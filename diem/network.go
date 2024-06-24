package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

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

func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()
	var buf [1024]byte

	length, err := conn.Read(buf[:])

	if err != nil {
		log.Fatalf("Error reading from connection: %v", err)
		return
	}

	var request ChunkRequest
	err = json.Unmarshal(buf[:length], &request)
	if err != nil {
		log.Fatalf("Error unmarshalling request: %v", err)
		return
	}
	fmt.Println("Received chunk request from node", request.NodeID)

	n.processChunkRequest(&request, conn)
}
