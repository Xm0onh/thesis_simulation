package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// Parameters for robust soliton distribution
type RobustSolitonParams struct {
	K     int     // Number of input symbols
	c     float64 // Failure probability parameter
	delta float64 // Failure probability parameter
}

// Generate the ideal soliton distribution
func IdealSolitonDistribution(K int) []float64 {
	rho := make([]float64, K+1)
	rho[1] = 1.0 / float64(K)
	for d := 2; d <= K; d++ {
		rho[d] = 1.0 / (float64(d) * float64(d-1))
	}
	return rho
}

// Generate the robust soliton distribution
func RobustSolitonDistribution(params RobustSolitonParams) []float64 {
	K := params.K
	c := params.c
	delta := params.delta

	rho := IdealSolitonDistribution(K)

	R := c * math.Log(float64(K)/delta) * math.Sqrt(float64(K))
	tau := make([]float64, K+1)
	for d := 1; d <= K; d++ {
		if d < int(math.Floor(float64(K)/R)) {
			tau[d] = R / (float64(d) * float64(K))
		} else if d == int(math.Floor(float64(K)/R)) {
			tau[d] = R * math.Log(R/delta) / float64(K)
		} else {
			tau[d] = 0
		}
	}

	Z := 0.0
	for d := 1; d <= K; d++ {
		Z += rho[d] + tau[d]
	}

	robust := make([]float64, K+1)
	for d := 1; d <= K; d++ {
		robust[d] = (rho[d] + tau[d]) / Z
	}
	return robust
}

// Select a degree d according to the robust soliton distribution
func SampleDegree(robust []float64) int {
	cumulative := make([]float64, len(robust))
	cumulative[0] = 0.0
	for i := 1; i < len(robust); i++ {
		cumulative[i] = cumulative[i-1] + robust[i]
	}
	r := rand.Float64()
	for i := 1; i < len(cumulative); i++ {
		if r < cumulative[i] {
			return i
		}
	}
	return len(robust) - 1
}

// Encoding function over Zp
func Encode(message []int, numEncodedSymbols int, p int, robust []float64) []EncodedSymbol {
	K := len(message)
	encodedSymbols := make([]EncodedSymbol, numEncodedSymbols)
	for i := 0; i < numEncodedSymbols; i++ {
		d := SampleDegree(robust)
		positions := rand.Perm(K)[:d]
		symbol := 0
		for _, pos := range positions {
			symbol = (symbol + message[pos]) % p
		}
		encodedSymbols[i] = EncodedSymbol{
			Value:     symbol,
			Positions: positions,
		}
	}
	return encodedSymbols
}

// EncodedSymbol represents an encoded symbol with its associated positions
type EncodedSymbol struct {
	Value     int
	Positions []int
}

// Peeling decoding over Zp
func Decode(encodedSymbols []EncodedSymbol, K int, p int) ([]int, bool) {
	// Initialize the set of unrecovered positions
	unrecovered := make(map[int]bool)
	for i := 0; i < K; i++ {
		unrecovered[i] = true
	}

	// Make a copy of encodedSymbols to modify during decoding
	esCopy := make([]EncodedSymbol, len(encodedSymbols))
	copy(esCopy, encodedSymbols)

	// Initialize queue with encoded symbols of degree one
	queue := []int{}
	for i, es := range esCopy {
		if len(es.Positions) == 1 {
			queue = append(queue, i)
		}
	}

	recovered := make([]int, K)

	for len(queue) > 0 {
		esIndex := queue[0]
		queue = queue[1:]
		es := esCopy[esIndex]

		// Check if Positions is not empty
		if len(es.Positions) == 0 {
			continue
		}

		pos := es.Positions[0]

		if !unrecovered[pos] {
			continue
		}

		value := es.Value
		recovered[pos] = value
		unrecovered[pos] = false

		// Update other encoded symbols
		for i, otherEs := range esCopy {
			if i == esIndex {
				continue
			}
			if contains(otherEs.Positions, pos) {
				// Subtract the recovered value
				esCopy[i].Value = (otherEs.Value - value + p) % p
				// Remove the position from Positions
				esCopy[i].Positions = removePosition(otherEs.Positions, pos)
				if len(esCopy[i].Positions) == 1 {
					queue = append(queue, i)
				}
			}
		}
	}

	success := true
	for _, unrecovered := range unrecovered {
		if unrecovered {
			success = false
			break
		}
	}

	if success {
		return recovered, true
	} else {
		return nil, false
	}
}

func contains(positions []int, pos int) bool {
	for _, p := range positions {
		if p == pos {
			return true
		}
	}
	return false
}

func removePosition(positions []int, pos int) []int {
	index := -1
	for i, p := range positions {
		if p == pos {
			index = i
			break
		}
	}
	if index != -1 {
		positions = append(positions[:index], positions[index+1:]...)
	}
	return positions
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Parameters
	p := 11                 // Prime number for Zp
	K := 10                 // Number of input symbols
	numEncodedSymbols := 25 // Increased number of encoded symbols

	// Message symbols (random integers in Zp)
	message := make([]int, K)
	for i := 0; i < K; i++ {
		message[i] = rand.Intn(p)
	}
	fmt.Println("Original message:", message)

	// Robust Soliton Distribution
	params := RobustSolitonParams{
		K:     K,
		c:     0.1,
		delta: 0.5,
	}
	robust := RobustSolitonDistribution(params)

	// Encoding
	encodedSymbols := Encode(message, numEncodedSymbols, p, robust)
	fmt.Println("Encoded Symbols:")
	for _, es := range encodedSymbols {
		sort.Ints(es.Positions)
		fmt.Printf("Value: %d, Positions: %v\n", es.Value, es.Positions)
	}

	// Decoding
	recoveredMessage, success := Decode(encodedSymbols, K, p)
	if success {
		fmt.Println("Recovered message:", recoveredMessage)
	} else {
		fmt.Println("Decoding failed. Not enough symbols or insufficient degrees.")
	}
}
