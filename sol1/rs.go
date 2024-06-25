package main

import (
	"bytes"
	"fmt"

	"github.com/klauspost/reedsolomon"
	"github.com/tendermint/tendermint/crypto/merkle"
)

// GenerateCodedChunks generates n coded chunks from the original data
func GenerateCodedChunks(data []byte) []Chunk {
	enc, err := reedsolomon.New(K, N-K)
	if err != nil {
		panic(err)
	}

	// Split data into k data shards
	shards, err := enc.Split(data)
	if err != nil {
		panic(err)
	}

	// fmt.Println("Size of data in bytes: ", len(shards[0]))
	// Encode parity shards

	err = enc.Encode(shards)
	if err != nil {
		panic(err)
	}
	// Convert shards to chunks
	chunks := make([]Chunk, N)
	for i := 0; i < N; i++ {
		chunks[i] = Chunk{Data: shards[i], Proof: merkle.Proof{}}
	}

	return chunks
}

func Decode(chunks map[int]Chunk) (string, error) {
	enc, err := reedsolomon.New(K, N-K)
	if err != nil {
		return "", err
	}

	// Prepare shards
	shards := make([][]byte, N)
	for i := 0; i < N; i++ {
		if chunk, ok := chunks[i]; ok {
			shards[i] = chunk.Data
		} else {
			shards[i] = nil
		}
	}

	// Reconstruct the original data
	err = enc.Reconstruct(shards)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = enc.Join(&buf, shards, len(shards[0])*K)
	if err != nil {
		return "", fmt.Errorf("failed to join shards: %v", err)
	}

	// Convert buffer to string and trim null padding
	decodedData := buf.Bytes()
	trimmedData := bytes.TrimRight(decodedData, "\x00")

	return string(trimmedData), nil
}
