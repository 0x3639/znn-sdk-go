package pow_test

import (
	"context"
	"errors"
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
		if errors.Is(err, pow.ErrCancelled) {
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
	difficulty := uint64(100000000) // Very high difficulty

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
	if errors.Is(result.Error, pow.ErrCancelled) {
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

// Example_multipleTransactions demonstrates handling multiple concurrent transactions
// with automatic worker pool management to prevent CPU exhaustion.
func Example_multipleTransactions() {
	// Configure worker pool (optional - defaults to 8)
	pow.SetMaxPoWWorkers(4) // Limit to 4 concurrent PoW operations

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Simulate multiple transactions requiring PoW
	numTransactions := 10
	difficulty := uint64(500000)

	fmt.Printf("Processing %d transactions with max %d concurrent PoW...\n",
		numTransactions, pow.GetMaxPoWWorkers())

	results := make([]<-chan pow.PowResult, numTransactions)

	// Launch all PoW operations
	// First 4 start immediately, remaining 6 queue automatically
	for i := 0; i < numTransactions; i++ {
		hash := types.Hash{}
		copy(hash[:], []byte(fmt.Sprintf("transaction_%d", i)))
		results[i] = pow.GeneratePowAsync(ctx, hash, difficulty)
	}

	// Collect results as they complete
	successCount := 0
	for i := 0; i < numTransactions; i++ {
		result := <-results[i]
		if result.Error != nil {
			fmt.Printf("Transaction %d failed: %v\n", i, result.Error)
			continue
		}
		successCount++
	}

	fmt.Printf("Completed %d/%d transactions successfully\n", successCount, numTransactions)
	fmt.Println("Worker pool prevented CPU exhaustion")

	// Example output:
	// Processing 10 transactions with max 4 concurrent PoW...
	// Completed 10/10 transactions successfully
	// Worker pool prevented CPU exhaustion
}

// Example_workerPoolConfiguration demonstrates different worker pool configurations
// for various hardware scenarios.
func Example_workerPoolConfiguration() {
	// Scenario 1: Low-end hardware (2-4 cores)
	pow.SetMaxPoWWorkers(2)
	fmt.Printf("Low-end config: %d workers\n", pow.GetMaxPoWWorkers())

	// Scenario 2: Standard desktop (4-8 cores)
	pow.SetMaxPoWWorkers(4)
	fmt.Printf("Standard config: %d workers\n", pow.GetMaxPoWWorkers())

	// Scenario 3: High-performance server (16+ cores)
	pow.SetMaxPoWWorkers(8)
	fmt.Printf("Server config: %d workers\n", pow.GetMaxPoWWorkers())

	// Can also use environment variable:
	// POW_MAX_WORKERS=16 go run main.go

	fmt.Println("Choose based on available CPU cores")

	// Output:
	// Low-end config: 2 workers
	// Standard config: 4 workers
	// Server config: 8 workers
	// Choose based on available CPU cores
}

// Example_batchProcessing demonstrates efficient batch processing of multiple
// PoW operations with automatic queuing and resource management.
func Example_batchProcessing() {
	// Use default worker pool (8 concurrent operations)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Batch of transactions (e.g., from an exchange or service)
	batchSize := 20
	difficulty := uint64(750000)

	fmt.Printf("Batch processing %d PoW operations...\n", batchSize)
	fmt.Printf("Max concurrent: %d (remaining queue automatically)\n", pow.GetMaxPoWWorkers())

	results := make([]<-chan pow.PowResult, batchSize)

	startTime := time.Now()

	// Submit all operations at once
	for i := 0; i < batchSize; i++ {
		hash := types.Hash{}
		copy(hash[:], []byte(fmt.Sprintf("batch_tx_%d", i)))
		results[i] = pow.GeneratePowAsync(ctx, hash, difficulty)
	}

	// Process results as they complete
	completed := 0
	failed := 0

	for i := 0; i < batchSize; i++ {
		result := <-results[i]
		if result.Error != nil {
			failed++
		} else {
			completed++
		}
	}

	elapsed := time.Since(startTime)

	fmt.Printf("Completed: %d, Failed: %d\n", completed, failed)
	fmt.Printf("Total time: %v\n", elapsed.Round(time.Millisecond))
	fmt.Println("Worker pool managed resources efficiently")
}
