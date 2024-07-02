package main

import (
	"bytes"
	"fmt"

	"github.com/tendermint/tendermint/crypto/merkle"
)

func GenerateDataChunks(data []byte) []Chunk {
	chunkSize := (len(data) + N - 1) / N
	chunks := make([]Chunk, N)

	for i := 0; i < N; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks[i] = Chunk{Data: data[start:end], Proof: merkle.Proof{}}
	}

	return chunks
}

// Decode simply reassembles the chunks into the original data
func Decode(chunks map[int]Chunk) (string, error) {
	var buf bytes.Buffer

	for i := 1; i < N; i++ {
		chunk, exists := chunks[i]
		if !exists {
			return "", fmt.Errorf("missing chunk %d", i)
		}
		buf.Write(chunk.Data)
	}

	decodedData := buf.Bytes()
	return string(decodedData), nil
}
