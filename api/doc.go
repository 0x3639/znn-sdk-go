// Package api provides the core blockchain API interfaces for interacting with the
// Zenon Network ledger, including querying account blocks, momentums, network statistics,
// and managing real-time subscriptions.
//
// The api package contains three primary APIs that are accessed through an RPC client:
//   - LedgerApi: Query and submit account blocks and momentums
//   - StatsApi: Network statistics, node information, and metrics
//   - SubscriberApi: Real-time subscriptions to blockchain events
//
// # LedgerApi
//
// The LedgerApi provides methods for querying the ledger state and submitting transactions:
//
//	// Query account information
//	info, err := client.LedgerApi.GetAccountInfoByAddress(address)
//	fmt.Printf("Balance: %s ZNN\n", info.Znn())
//
//	// Get current blockchain height
//	momentum, err := client.LedgerApi.GetFrontierMomentum()
//	fmt.Printf("Height: %d\n", momentum.Height)
//
//	// Get unreceived blocks
//	blocks, err := client.LedgerApi.GetUnreceivedBlocksByAddress(address, 0, 10)
//
// # Transaction Templates
//
// LedgerApi provides helper methods to create transaction templates:
//
//	// Create a send transaction template
//	template := client.LedgerApi.SendTemplate(
//	    toAddress,
//	    types.ZnnTokenStandard,
//	    amount,
//	    []byte{},
//	)
//
//	// Create a receive transaction template
//	receiveTemplate := client.LedgerApi.ReceiveTemplate(blockHash)
//
// Transaction templates must be signed and enhanced with PoW before publishing.
// See the embedded contract APIs for additional transaction types.
//
// # StatsApi
//
// Query network and node information:
//
//	// Get node OS information
//	osInfo, err := client.StatsApi.OsInfo()
//	fmt.Printf("OS: %s\n", osInfo.OS)
//
//	// Get network information
//	networkInfo, err := client.StatsApi.NetworkInfo()
//	fmt.Printf("Peers: %d\n", networkInfo.NumPeers)
//
// # SubscriberApi
//
// Subscribe to real-time blockchain events:
//
//	ctx := context.Background()
//
//	// Subscribe to new momentums
//	sub, momentumChan, err := client.SubscriberApi.ToMomentums(ctx)
//	defer sub.Unsubscribe()
//	for momentums := range momentumChan {
//	    for _, m := range momentums {
//	        fmt.Printf("New momentum: height=%d\n", m.Height)
//	    }
//	}
//
//	// Subscribe to account blocks for a specific address
//	sub, blockChan, err := client.SubscriberApi.ToAccountBlocksByAddress(ctx, address)
//	defer sub.Unsubscribe()
//	for blocks := range blockChan {
//	    for _, block := range blocks {
//	        fmt.Printf("New block: hash=%s\n", block.Hash)
//	    }
//	}
//
// # Transaction Submission
//
// To submit a transaction:
//  1. Create a transaction template using LedgerApi or embedded contract API
//  2. Autofill transaction parameters (height, previous hash, momentum acknowledgment)
//  3. Generate PoW or use fused plasma
//  4. Sign the transaction with a keypair
//  5. Publish via PublishRawTransaction
//
// For complete transaction examples, see the examples directory.
//
// For more information, see https://pkg.go.dev/github.com/0x3639/znn-sdk-go/api
package api
