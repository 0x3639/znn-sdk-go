package rpc_client_test

import (
	"fmt"
	"log"
	"time"

	"github.com/0x3639/znn-sdk-go/rpc_client"
)

// Example demonstrates basic RPC client connection to a Zenon node.
func Example() {
	// Connect to local Zenon node with default options
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Check connection status
	if client.Status() == rpc_client.Running {
		fmt.Println("Connected to Zenon node")
	}

	// Query blockchain data
	momentum, err := client.LedgerApi.GetFrontierMomentum()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Current height: %d\n", momentum.Height)
	// Output example (actual output will vary):
	// Connected to Zenon node
	// Current height: 1234567
}

// Example_withCallbacks demonstrates connection lifecycle management with callbacks.
func Example_withCallbacks() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Register callback for connection established
	client.AddOnConnectionEstablishedCallback(func() {
		fmt.Println("Connection established")
		// Reinitialize state, resubscribe to events, etc.
	})

	// Register callback for connection lost
	client.AddOnConnectionLostCallback(func(err error) {
		fmt.Printf("Connection lost: %v\n", err)
		// Clean up resources, notify application, etc.
	})

	// Use the client normally
	momentum, _ := client.LedgerApi.GetFrontierMomentum()
	fmt.Printf("Height: %d\n", momentum.Height)
}

// Example_customOptions demonstrates creating a client with custom configuration.
func Example_customOptions() {
	// Configure custom options
	opts := rpc_client.ClientOptions{
		AutoReconnect:       true,
		ReconnectDelay:      2 * time.Second,
		MaxReconnectDelay:   60 * time.Second,
		ReconnectAttempts:   10, // Give up after 10 attempts
		HealthCheckInterval: 15 * time.Second,
	}

	client, err := rpc_client.NewRpcClientWithOptions("ws://127.0.0.1:35998", opts)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Client will auto-reconnect with exponential backoff up to 10 attempts
	// Health checks run every 15 seconds
	fmt.Println("Client configured with custom reconnection policy")

	// Use client normally
	info, _ := client.StatsApi.OsInfo()
	fmt.Printf("Node OS: %s\n", info.Os)
}

// Example_checkStatus demonstrates monitoring connection status.
func Example_checkStatus() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Check status before making requests
	status := client.Status()
	switch status {
	case rpc_client.Running:
		fmt.Println("Client is connected and ready")
		// Safe to make requests
	case rpc_client.Connecting:
		fmt.Println("Client is connecting...")
	case rpc_client.Stopped:
		fmt.Println("Client is stopped")
	case rpc_client.Uninitialized:
		fmt.Println("Client is not initialized")
	}

	// Make request if connected
	if client.Status() == rpc_client.Running {
		momentum, _ := client.LedgerApi.GetFrontierMomentum()
		fmt.Printf("Current momentum height: %d\n", momentum.Height)
	}
}

// Example_multipleAPIs demonstrates using different API endpoints.
func Example_multipleAPIs() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Use Ledger API
	momentum, _ := client.LedgerApi.GetFrontierMomentum()
	fmt.Printf("Momentum height: %d\n", momentum.Height)

	// Use Stats API
	networkInfo, _ := client.StatsApi.NetworkInfo()
	fmt.Printf("Network peers: %d\n", networkInfo.NumPeers)

	// Use embedded contract APIs
	pillars, _ := client.PillarApi.GetAll(0, 5)
	fmt.Printf("Total pillars: %d\n", pillars.Count)

	// All APIs share the same WebSocket connection
	fmt.Println("All APIs use single connection")
}
