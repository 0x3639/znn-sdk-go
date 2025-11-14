package pow_test

import (
	"context"
	"fmt"
	"time"

	"github.com/0x3639/znn-sdk-go/pow"
	"github.com/zenon-network/go-zenon/common/types"
)

// Example demonstrates basic PoW generation for a transaction.
func Example() {
	// Sample transaction hash
	dataHash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")

	// Low difficulty for demonstration
	difficulty := uint64(1000000)

	fmt.Println("Generating PoW...")

	// Generate PoW (blocking operation)
	nonce := pow.GeneratePoW(dataHash, difficulty)

	fmt.Printf("PoW generated successfully\n")
	fmt.Printf("Nonce: %s\n", nonce)
	fmt.Println("Ready to publish transaction")

	// Output:
	// Generating PoW...
	// PoW generated successfully
	// Nonce: 0000000000000001
	// Ready to publish transaction
}

// Example_withContext demonstrates PoW generation with cancellation support.
func Example_withContext() {
	dataHash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")
	difficulty := uint64(1000000)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("Generating PoW with 10s timeout...")

	// Generate with context
	nonce, err := pow.GeneratePowWithContext(ctx, dataHash, difficulty)
	if err != nil {
		if err == pow.ErrCancelled {
			fmt.Println("PoW generation cancelled")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	fmt.Println("PoW completed within timeout")
	fmt.Printf("Nonce: %s\n", nonce)
}

// Example_async demonstrates asynchronous PoW generation.
func Example_async() {
	dataHash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")
	difficulty := uint64(1000000)

	ctx := context.Background()

	fmt.Println("Starting async PoW generation...")

	// Generate PoW asynchronously
	resultChan := pow.GeneratePowAsync(ctx, dataHash, difficulty)

	// Can do other work here while PoW generates...
	fmt.Println("Waiting for PoW result...")

	// Wait for result
	select {
	case result := <-resultChan:
		if result.Error != nil {
			fmt.Printf("PoW error: %v\n", result.Error)
			return
		}
		fmt.Println("PoW completed asynchronously")
		fmt.Printf("Nonce: %s\n", result.Nonce)
	case <-time.After(30 * time.Second):
		fmt.Println("PoW timeout")
	}
}

// Example_withCancellation demonstrates canceling PoW generation.
func Example_withCancellation() {
	dataHash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")
	difficulty := uint64(10000000000) // Very high difficulty

	ctx, cancel := context.WithCancel(context.Background())

	fmt.Println("Starting PoW generation...")
	resultChan := pow.GeneratePowAsync(ctx, dataHash, difficulty)

	// Simulate cancellation after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Canceling PoW generation...")
		cancel()
	}()

	// Wait for result or cancellation
	result := <-resultChan
	if result.Error == pow.ErrCancelled {
		fmt.Println("PoW was cancelled as expected")
	} else if result.Error != nil {
		fmt.Printf("Unexpected error: %v\n", result.Error)
	} else {
		fmt.Println("PoW completed (shouldn't happen with high difficulty)")
	}

	// Output:
	// Starting PoW generation...
	// Canceling PoW generation...
	// PoW was cancelled as expected
}

// Example_difficultyComparison demonstrates how difficulty affects generation time.
func Example_difficultyComparison() {
	dataHash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")

	// Low difficulty - fast
	lowDifficulty := uint64(100000)

	start := time.Now()
	pow.GeneratePoW(dataHash, lowDifficulty)
	lowTime := time.Since(start)

	fmt.Printf("Low difficulty (%d): ~instant\n", lowDifficulty)

	// Medium difficulty - slower
	mediumDifficulty := uint64(1000000)

	start = time.Now()
	pow.GeneratePoW(dataHash, mediumDifficulty)
	medTime := time.Since(start)

	fmt.Printf("Medium difficulty (%d): takes longer\n", mediumDifficulty)

	// Higher difficulty takes exponentially longer
	fmt.Println("Difficulty directly impacts generation time")
	fmt.Printf("Low: %v, Medium: %v\n", lowTime, medTime)
}

// Example_zeroDifficulty demonstrates handling zero difficulty.
func Example_zeroDifficulty() {
	dataHash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")

	// Zero difficulty returns immediately
	nonce := pow.GeneratePoW(dataHash, 0)

	fmt.Println("Zero difficulty PoW")
	fmt.Printf("Nonce: %s\n", nonce)
	fmt.Println("Returns instantly")

	// Output:
	// Zero difficulty PoW
	// Nonce: 0000000000000000
	// Returns instantly
}

// Example_bytes demonstrates generating PoW as bytes instead of hex string.
func Example_bytes() {
	dataHash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")
	difficulty := uint64(1000000)

	// Generate as bytes
	nonceBytes := pow.GeneratePowBytes(dataHash, difficulty)

	fmt.Println("PoW generated as bytes")
	fmt.Printf("Nonce length: %d bytes\n", len(nonceBytes))
	fmt.Printf("First byte: %d\n", nonceBytes[0])

	// Useful when transaction requires byte format
	fmt.Println("Ready for transaction")
}
