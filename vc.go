package main

import (
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

func VerifyChunk(client *Client, commitment []byte, chunk Chunk, proof *merkle.Proof) bool {
	err := proof.Verify(commitment, chunk.Data)
	return err == nil
}
