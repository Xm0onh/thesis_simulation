package main

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/cbergoon/merkletree"
)

func (c Content) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(c.Data)); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func (c Content) Equals(other merkletree.Content) (bool, error) {
	return c.Data == other.(Content).Data, nil
}

func SimulateNetworkDelay() {
	delay := time.Duration(rand.Intn(100)) * time.Millisecond
	time.Sleep(delay)
}

func SimulateBandwidthLimit(chunkSize int) {

	bandwidth := 10 * 1_000_000 / 8
	delay := time.Duration(chunkSize) * time.Second / time.Duration(bandwidth)
	fmt.Println("Simulating bandwidth limit:", delay)
	time.Sleep(delay)
}

func ReadFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}
