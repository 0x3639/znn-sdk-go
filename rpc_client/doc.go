// Package rpc_client provides HTTP and WebSocket JSON-RPC client functionality for
// connecting to Zenon Network nodes. This is the main entry point for the SDK,
// managing connections and instantiating all API endpoints.
//
// HTTP and HTTPS support ledger, stats, embedded reads, and transaction publication.
// WebSocket and secure WebSocket transports support the same requests plus real-time
// subscriptions. API errors are normalized as transport.RPCError values containing
// node error data and the original method and positional parameters.
//
// # Basic Usage
//
// Connect to a local node:
//
//	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Stop()
//
//	// Query blockchain data
//	momentum, err := client.LedgerApi.GetFrontierMomentum()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Current height: %d\n", momentum.Height)
//
// HTTP read/write lifecycle:
//
//	httpClient, err := rpc_client.NewRpcClient("http://127.0.0.1:35997")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer httpClient.Stop()
//
// # Connection Management
//
// The client supports callbacks for connection lifecycle events:
//
//	client.AddOnConnectionEstablishedCallback(func() {
//	    fmt.Println("Connected to node")
//	})
//
//	client.AddOnConnectionLostCallback(func(err error) {
//	    fmt.Printf("Connection lost: %v\n", err)
//	})
//
// For normalized updates with automatic reconnection and resubscription, use
// [RpcClient.Subscribe]. Calling [RpcClient.Stop] closes these subscription sockets,
// closes their channels, and clears registered lifecycle callbacks.
//
// # Available APIs
//
// Once connected, the client provides access to:
//   - LedgerApi: Query account blocks, momentums, and submit transactions
//   - StatsApi: Network statistics and node information
//   - SubscriberApi: Real-time subscriptions to blockchain events
//   - Embedded contract APIs: Plasma, Pillar, Token, Sentinel, Stake, and more
//
// # Connection Options
//
// For advanced configuration, use NewRpcClientWithOptions:
//
//	options := rpc_client.ClientOptions{
//	    AutoReconnect:    true,
//	    ReconnectAttempts: 5,
//	}
//	client, err := rpc_client.NewRpcClientWithOptions("ws://127.0.0.1:35998", options)
//
// # Read vs Write Operations
//
// Read-only operations (queries) only require a connected client. Write operations
// (transactions) require a wallet and keypair for signing. See the wallet package
// for wallet management.
//
// For more examples, see https://pkg.go.dev/github.com/0x3639/znn-sdk-go/rpc_client
package rpc_client
