package main

import (
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"

	"github.com/tendermint/tendermint/crypto/merkle"
)

// Number of Servers + Number of Coded chunks
var n = 1000

// Number of Data chunks
var k = 660

type Chunk struct {
	Data  []byte
	Proof merkle.Proof
}

type Server struct {
	ID         int
	Chunks     []Chunk
	Commitment []byte
	Proofs     []*merkle.Proof
}

type Client struct {
	VerifiedChunks map[int]Chunk
	Blacklist      []int
}

type Content struct {
	Data string
}

var p = new(big.Int)

func init() {
	primeStr := "340282366920938463463374607431768211507"
	p.SetString(primeStr, 10)
	rand.Seed(time.Now().UnixNano())
}

func (client *Client) RequestChunk(server *Server, chunkIndex int) {
	SimulateNetworkDelay()

	chunk := server.RespondToRequest(chunkIndex)
	SimulateBandwidthLimit(len(chunk.Data))

	if VerifyChunk(client, server.Commitment, chunk, &server.Chunks[chunkIndex].Proof) {
		if client.VerifiedChunks == nil {
			client.VerifiedChunks = make(map[int]Chunk)
		}
		client.VerifiedChunks[chunkIndex] = chunk
	} else {
		client.Blacklist = append(client.Blacklist, server.ID)
	}
}

func (server *Server) RespondToRequest(chunkIndex int) Chunk {
	SimulateNetworkDelay()

	return server.Chunks[chunkIndex]
}

func Servers(data []byte) []Server {
	servers := make([]Server, n)
	for i := 0; i < n; i++ {
		servers[i] = Server{
			ID:     i + 1,
			Chunks: GenerateCodedChunks(data, n, k),
		}
	}
	return servers
}

func SimulateExperiment(filename string) {
	data, err := ReadFile(filename)
	if err != nil {
		panic(err)
	}

	// Setup servers and clients for Solution #1
	servers := Servers(data)
	fmt.Println("Servers generated:", len(servers))
	// Compute commitments for servers
	var wg sync.WaitGroup
	for i := range len(servers) {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			commitment, proofs := CreateVectorCommitment(servers[i].Chunks)
			servers[i].Commitment = commitment
			for j := range servers[i].Chunks {
				servers[i].Chunks[j].Proof = *proofs[j]
			}
		}(i)
	}
	wg.Wait()
	client1 := Client{}
	startTime := time.Now()
	for i := range len(servers) {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			client1.RequestChunk(&servers[i], i)
		}(i)
	}
	wg.Wait()

	fmt.Println("Client 1 verified chunks:", len(client1.VerifiedChunks))
	if len(client1.VerifiedChunks) >= k {
		originalFile, _ := Decode(client1.VerifiedChunks, n, k)
		if string(originalFile) == string(data) {
			fmt.Println("Time taken to decode the file (Solution #1):", time.Since(startTime))
		} else {
			fmt.Println("Decoded file is different from the original file (Solution #1)")
		}
	} else {
		fmt.Println("Not enough verified chunks to decode the file (Solution #1)")
	}

	fmt.Printf("Client 1 verified %d chunks and blacklisted %d servers (Solution #1)\n",
		len(client1.VerifiedChunks), len(client1.Blacklist))
}

func main() {
	SimulateExperiment("eth_transactions.json")
}
