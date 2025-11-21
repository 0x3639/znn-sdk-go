// Package pow provides proof-of-work generation functionality for Zenon Network transactions.
//
// In the Zenon Network, transactions can be feeless by either:
//  1. Fusing QSR to generate plasma (see PlasmaApi in api/embedded)
//  2. Generating computational proof-of-work
//
// This package implements the PoW algorithm required for transactions that don't have
// sufficient plasma. The difficulty of the PoW depends on the transaction type and current
// network conditions.
//
// # Basic Usage
//
// Generate PoW for a transaction:
//
//	// Check required difficulty
//	difficulty, err := client.PlasmaApi.GetRequiredPoWForAccountBlock(accountBlock)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Generate PoW (blocking)
//	nonce, err := pow.GeneratePoW(
//	    accountBlock.Address.Bytes(),
//	    accountBlock.Hash.Bytes(),
//	    difficulty,
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Set nonce on account block
//	accountBlock.Nonce = nonce
//
// # Asynchronous PoW Generation
//
// For better responsiveness, use asynchronous PoW generation with context:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	// Start PoW generation in background
//	resultChan := pow.GeneratePowAsync(ctx, dataHash, difficulty)
//
//	// Wait for result
//	result := <-resultChan
//	if result.Error != nil {
//	    log.Fatal(result.Error)
//	}
//	accountBlock.Nonce = result.Nonce
//
// # Worker Pool and Concurrency Control
//
// To prevent CPU exhaustion when multiple transactions are submitted concurrently,
// the async PoW functions use a worker pool that limits concurrent operations to 8
// by default (matching the Dart SDK's parallelism level).
//
// When more than 8 PoW requests are made simultaneously:
//   - First 8 requests start immediately
//   - Additional requests queue and wait for a worker to become available
//   - Context cancellation is respected while queued
//
// Configure the worker pool before generating PoW:
//
//	// Limit to 4 workers (for low-end hardware)
//	pow.SetMaxPoWWorkers(4)
//
//	// Or use environment variable: POW_MAX_WORKERS=4
//
// Example with multiple concurrent transactions:
//
//	results := make([]<-chan pow.PowResult, 20)
//	for i := 0; i < 20; i++ {
//	    results[i] = pow.GeneratePowAsync(ctx, hashes[i], difficulty)
//	}
//	// First 8 start immediately, remaining 12 queue
//	for i := 0; i < 20; i++ {
//	    result := <-results[i]
//	    if result.Error != nil {
//	        log.Printf("PoW %d failed: %v", i, result.Error)
//	        continue
//	    }
//	    // Process successful result
//	}
//
// # Performance Considerations
//
// PoW generation time depends on:
//   - Required difficulty (higher difficulty = longer time)
//   - CPU performance (each PoW computation runs single-threaded)
//   - Random luck factor
//   - Worker pool availability (queuing when >8 concurrent requests)
//
// Typical generation times:
//   - Low difficulty (simple transfers with some plasma): < 1 second
//   - Medium difficulty (contract calls): 1-10 seconds
//   - High difficulty (no plasma, complex operations): 10-60+ seconds
//
// Note: Each individual PoW computation runs on a single CPU core. The worker pool
// controls how many PoW computations can run simultaneously, not parallel computation
// within a single PoW operation.
//
// # Plasma vs PoW
//
// For frequent transactions, fusing QSR for plasma is more efficient than
// repeatedly generating PoW. Consider:
//   - Fuse QSR once → many feeless transactions
//   - Generate PoW → computational cost per transaction
//
// Check plasma availability before generating PoW:
//
//	plasmaInfo, err := client.PlasmaApi.Get(address)
//	if plasmaInfo.CurrentPlasma < requiredPlasma {
//	    // Need to generate PoW or fuse more QSR
//	}
//
// For more information, see https://pkg.go.dev/github.com/0x3639/znn-sdk-go/pow
package pow
