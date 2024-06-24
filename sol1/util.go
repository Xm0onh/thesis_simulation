package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

func GenerateTransactions(num int) []Transaction {
	var transactions []Transaction
	for i := 0; i < num; i++ {
		txn := Transaction{
			ID:        strconv.Itoa(i), // Simple incremental IDs
			Content:   "Data for transaction " + strconv.Itoa(i),
			Signature: GenerateSignature("Data for transaction " + strconv.Itoa(i)),
			Timestamp: time.Now().Unix(),
		}
		transactions = append(transactions, txn)
	}
	return transactions
}

func GenerateSignature(data string) string {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GenerateBlockHash(block Block) string {
	hasher := sha256.New()
	hasher.Write([]byte(block.PreviousHash))
	for _, txn := range block.Transactions {
		hasher.Write([]byte(txn.ID + txn.Content + txn.Signature))
	}
	hasher.Write([]byte(strconv.Itoa(block.Nonce)))
	hasher.Write([]byte(strconv.FormatInt(block.Timestamp, 10)))
	return hex.EncodeToString(hasher.Sum(nil))
}
