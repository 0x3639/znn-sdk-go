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
// For better responsiveness, use asynchronous PoW generation:
//
//	resultChan := make(chan *pow.PowResult)
//	cancelChan := make(chan struct{})
//
//	// Start PoW generation in background
//	go pow.GeneratePowAsync(
//	    accountBlock.Address.Bytes(),
//	    accountBlock.Hash.Bytes(),
//	    difficulty,
//	    resultChan,
//	    cancelChan,
//	)
//
//	// Wait for result or cancel
//	select {
//	case result := <-resultChan:
//	    if result.Err != nil {
//	        log.Fatal(result.Err)
//	    }
//	    accountBlock.Nonce = result.Nonce
//	case <-time.After(30 * time.Second):
//	    close(cancelChan)
//	    log.Fatal("PoW generation timeout")
//	}
//
// # Performance Considerations
//
// PoW generation time depends on:
//   - Required difficulty (higher difficulty = longer time)
//   - CPU performance (uses all available cores)
//   - Random luck factor
//
// Typical generation times:
//   - Low difficulty (simple transfers with some plasma): < 1 second
//   - Medium difficulty (contract calls): 1-10 seconds
//   - High difficulty (no plasma, complex operations): 10-60+ seconds
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
