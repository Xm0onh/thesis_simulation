package main

import (
	"bytes"
	"fmt"

	"github.com/klauspost/reedsolomon"
)

// GenerateCodedChunks generates n coded chunks from the original data
func GenerateCodedChunks(data []byte, n, k int) []Chunk {
	enc, err := reedsolomon.New(k, n-k)
	if err != nil {
		panic(err)
	}

	// Split data into k data shards
	shards, err := enc.Split(data)
	if err != nil {
		panic(err)
	}

	// Encode parity shards
	err = enc.Encode(shards)
	if err != nil {
		panic(err)
	}

	// Convert shards to chunks
	chunks := make([]Chunk, n)
	for i := 0; i < n; i++ {
		chunks[i] = Chunk{Data: shards[i]}
	}

	return chunks
}
func Decode(chunks map[int]Chunk, n, k int) (string, error) {
	enc, err := reedsolomon.New(k, n-k)
	if err != nil {
		return "", err
	}

	// Prepare shards
	shards := make([][]byte, n)
	for i := 0; i < n; i++ {
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
	err = enc.Join(&buf, shards, len(shards[0])*k)
	if err != nil {
		return "", fmt.Errorf("failed to join shards: %v", err)
	}

	// Convert buffer to string and trim null padding
	decodedData := buf.Bytes()
	trimmedData := bytes.TrimRight(decodedData, "\x00")

	return string(trimmedData), nil
}
