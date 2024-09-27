package main

import (
	"time"

	"github.com/tendermint/tendermint/crypto/merkle"
)

// ###################################

func CreateVectorCommitment(chunks []Chunk) ([]byte, []*merkle.Proof) {
	leaves := make([][]byte, len(chunks))
	for i, chunk := range chunks {
		leaves[i] = chunk.Data
	}

	rootHash, proofs := merkle.ProofsFromByteSlices(leaves)
	return rootHash, proofs
}

func VerifyChunk(rootHash []byte, chunk Chunk, proof *merkle.Proof, node *Node) bool {
	// Size of proof and commitment in byte
	// proofSize := reflect.TypeOf(*proof).Size()
	// commitmentSize := reflect.TypeOf(rootHash).Size()
	// fmt.Println("Size of proof and commitment in byte: ", proofSize, commitmentSize)

	// Capture the time for verification
	start := time.Now()
	err := proof.Verify(rootHash, chunk.Data)
	// fmt.Println("Time taken for verification: ")
	node.Metrics.VerificationTime += time.Since(start)
	return err == nil
}
