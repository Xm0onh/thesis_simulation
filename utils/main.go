package main

import (
	crand "crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
)

// Configuration holds the parameters for generating big integers
type Configuration struct {
	Number     int    // Number of big integers to generate (N)
	BitSize    int    // Bit size of each big integer
	OutputFile string // Name of the output JSON file
}

func main() {
	// Parse command-line flags
	config := parseFlags()

	// Validate input
	if config.Number <= 0 {
		log.Fatalf("Invalid number of elements (N): %d. Must be a positive integer.", config.Number)
	}
	if config.BitSize <= 0 {
		log.Fatalf("Invalid bit size: %d. Must be a positive integer.", config.BitSize)
	}

	// Generate big integers
	bigInts, err := generateBigIntegers(config.Number, config.BitSize)
	if err != nil {
		log.Fatalf("Error generating big integers: %v", err)
	}

	// Marshal big integers to JSON
	jsonData, err := json.MarshalIndent(bigInts, "", "    ")
	if err != nil {
		log.Fatalf("Error marshaling big integers to JSON: %v", err)
	}

	// Write JSON data to file
	err = writeJSONToFile(config.OutputFile, jsonData)
	if err != nil {
		log.Fatalf("Error writing JSON to file '%s': %v", config.OutputFile, err)
	}

	fmt.Printf("Successfully generated %d big integers and saved to '%s'.\n", config.Number, config.OutputFile)
}

// parseFlags parses and returns the command-line flags
func parseFlags() Configuration {
	numberPtr := flag.Int("n", 0, "Number of big integers to generate (required, must be >0)")
	bitSizePtr := flag.Int("bits", 256, "Bit size of each big integer (default: 256)")
	outputPtr := flag.String("o", "bigints.json", "Output JSON file name (default: bigints.json)")

	flag.Parse()

	// Validate that -n flag is provided by checking if it is greater than 0
	// Since the default is 0, users must provide a positive integer for -n
	if *numberPtr <= 0 {
		fmt.Println("Usage: generate_bigints -n <NUMBER_OF_ELEMENTS> [OPTIONS]")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	return Configuration{
		Number:     *numberPtr,
		BitSize:    *bitSizePtr,
		OutputFile: *outputPtr,
	}
}

// generateBigIntegers generates N big integers each with the specified bit size
func generateBigIntegers(N int, bitSize int) ([]string, error) {
	bigInts := make([]string, 0, N)
	for i := 0; i < N; i++ {
		// Generate a random big integer in [0, 2^bitSize)
		bInt, err := crand.Int(crand.Reader, new(big.Int).Lsh(big.NewInt(1), uint(bitSize)))
		if err != nil {
			return nil, fmt.Errorf("failed to generate big integer %d: %v", i+1, err)
		}
		bigInts = append(bigInts, bInt.String())
	}
	return bigInts, nil
}

// writeJSONToFile writes the JSON data to the specified file
func writeJSONToFile(filename string, data []byte) error {
	// Create or truncate the file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create or truncate file '%s': %v", filename, err)
	}
	defer file.Close()

	// Write data to file
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to file '%s': %v", filename, err)
	}

	return nil
}
