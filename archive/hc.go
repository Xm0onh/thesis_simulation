package main

import (
	"crypto/sha256"
	"math/big"
)

func HomomorphicCommitment(chunks []Chunk, p *big.Int) [][]byte {
	commitments := make([][]byte, len(chunks))

	for i, chunk := range chunks {
		hash := sha256.Sum256(chunk.Data)
		commitments[i] = hash[:]
	}

	return commitments
}

func VerifyHomomorphicCommitment(commitments [][]byte, chunk Chunk) bool {
	return true

}
