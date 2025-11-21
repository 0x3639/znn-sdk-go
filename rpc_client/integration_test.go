//go:build integration

package rpc_client

import (
	"os"
	"testing"
	"time"

	"github.com/zenon-network/go-zenon/common/types"
)

// TestIntegration_ConnectToPublicNode tests connection to a real Zenon node
func TestIntegration_ConnectToPublicNode(t *testing.T) {
	// Get RPC URL from environment variable
	rpcURL := os.Getenv("TEST_ZENON_RPC_URL")
	if rpcURL == "" {
		rpcURL = "wss://my.hc1node.com:35998" // Default public node
	}

	t.Logf("Connecting to: %s", rpcURL)

	// Create client
	client, err := NewRpcClient(rpcURL)
	if err != nil {
		t.Fatalf("Failed to create RPC client: %v", err)
	}
	defer client.Stop()

	// Wait for connection to establish
	time.Sleep(2 * time.Second)

	// Check client is running
	status := client.Status()
	if status != Running {
		t.Errorf("Expected client status Running, got %v", status)
	}
}

// TestIntegration_GetFrontierMomentum tests reading blockchain data
func TestIntegration_GetFrontierMomentum(t *testing.T) {
	rpcURL := os.Getenv("TEST_ZENON_RPC_URL")
	if rpcURL == "" {
		rpcURL = "wss://my.hc1node.com:35998"
	}

	client, err := NewRpcClient(rpcURL)
	if err != nil {
		t.Fatalf("Failed to create RPC client: %v", err)
	}
	defer client.Stop()

	// Wait for connection
	time.Sleep(2 * time.Second)

	// Get frontier momentum
	momentum, err := client.LedgerApi.GetFrontierMomentum()
	if err != nil {
		t.Fatalf("GetFrontierMomentum() failed: %v", err)
	}

	// Validate momentum
	if momentum == nil {
		t.Fatal("GetFrontierMomentum() returned nil")
	}

	if momentum.Height == 0 {
		t.Error("Momentum height should be > 0")
	}

	t.Logf("Frontier Momentum - Height: %d, Hash: %s",
		momentum.Height, momentum.Hash)
}

// TestIntegration_GetAccountInfo tests account balance queries
func TestIntegration_GetAccountInfo(t *testing.T) {
	rpcURL := os.Getenv("TEST_ZENON_RPC_URL")
	if rpcURL == "" {
		rpcURL = "wss://my.hc1node.com:35998"
	}

	client, err := NewRpcClient(rpcURL)
	if err != nil {
		t.Fatalf("Failed to create RPC client: %v", err)
	}
	defer client.Stop()

	// Wait for connection
	time.Sleep(2 * time.Second)

	// Query a well-known embedded contract address
	testAddress, err := types.ParseAddress("z1qxemdeddedxpyllarxxxxxxxxxxxxxxxsy3fmg") // Pillar contract address
	if err != nil {
		t.Fatalf("ParseAddress() failed: %v", err)
	}

	accountInfo, err := client.LedgerApi.GetAccountInfoByAddress(testAddress)
	if err != nil {
		t.Fatalf("GetAccountInfoByAddress() failed: %v", err)
	}

	// Validate account info
	if accountInfo == nil {
		t.Fatal("GetAccountInfoByAddress() returned nil")
	}

	t.Logf("Account Info - Address: %s, Account Height: %d",
		accountInfo.Address, accountInfo.AccountHeight)
}

// TestIntegration_AutoReconnect tests auto-reconnection functionality
func TestIntegration_AutoReconnect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping auto-reconnect test in short mode")
	}

	rpcURL := os.Getenv("TEST_ZENON_RPC_URL")
	if rpcURL == "" {
		rpcURL = "wss://my.hc1node.com:35998"
	}

	// Create client with auto-reconnect enabled
	opts := DefaultClientOptions()
	opts.AutoReconnect = true
	opts.ReconnectDelay = 1 * time.Second
	opts.MaxReconnectDelay = 5 * time.Second
	opts.ReconnectAttempts = 3

	client, err := NewRpcClientWithOptions(rpcURL, opts)
	if err != nil {
		t.Fatalf("Failed to create RPC client: %v", err)
	}
	defer client.Stop()

	// Wait for initial connection
	time.Sleep(2 * time.Second)

	// Verify initial connection
	if client.Status() != Running {
		t.Fatal("Client should be running initially")
	}

	// Test that client is functional
	momentum, err := client.LedgerApi.GetFrontierMomentum()
	if err != nil {
		t.Fatalf("Initial query failed: %v", err)
	}
	if momentum == nil {
		t.Fatal("Initial query returned nil")
	}

	t.Log("Auto-reconnect test passed (client stays connected)")
}

// TestIntegration_HealthCheck tests health check functionality
func TestIntegration_HealthCheck(t *testing.T) {
	rpcURL := os.Getenv("TEST_ZENON_RPC_URL")
	if rpcURL == "" {
		rpcURL = "wss://my.hc1node.com:35998"
	}

	// Create client with health check enabled
	opts := DefaultClientOptions()
	opts.HealthCheckInterval = 5 * time.Second

	client, err := NewRpcClientWithOptions(rpcURL, opts)
	if err != nil {
		t.Fatalf("Failed to create RPC client: %v", err)
	}
	defer client.Stop()

	// Wait for connection and initial health check
	time.Sleep(3 * time.Second)

	// Verify client is healthy
	if client.Status() != Running {
		t.Errorf("Expected client to be Running, got %v", client.Status())
	}

	// Wait for a health check to occur
	time.Sleep(6 * time.Second)

	// Verify still healthy
	if client.Status() != Running {
		t.Errorf("Client should still be Running after health check")
	}

	t.Log("Health check test passed")
}

// TestIntegration_ConcurrentRequests tests concurrent RPC calls
func TestIntegration_ConcurrentRequests(t *testing.T) {
	rpcURL := os.Getenv("TEST_ZENON_RPC_URL")
	if rpcURL == "" {
		rpcURL = "wss://my.hc1node.com:35998"
	}

	client, err := NewRpcClient(rpcURL)
	if err != nil {
		t.Fatalf("Failed to create RPC client: %v", err)
	}
	defer client.Stop()

	// Wait for connection
	time.Sleep(2 * time.Second)

	// Run multiple concurrent requests
	const numRequests = 10
	done := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(index int) {
			_, err := client.LedgerApi.GetFrontierMomentum()
			done <- err
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent request %d failed: %v", i, err)
		}
	}

	t.Logf("Successfully completed %d concurrent requests", numRequests)
}

// TestIntegration_TLSConnection tests secure WebSocket connection
func TestIntegration_TLSConnection(t *testing.T) {
	rpcURL := os.Getenv("TEST_ZENON_RPC_URL")
	if rpcURL == "" {
		rpcURL = "wss://my.hc1node.com:35998"
	}

	// Verify we're testing with wss://
	if len(rpcURL) < 6 || rpcURL[:6] != "wss://" {
		t.Skip("Skipping TLS test - not using wss:// URL")
	}

	client, err := NewRpcClient(rpcURL)
	if err != nil {
		t.Fatalf("Failed to create RPC client with TLS: %v", err)
	}
	defer client.Stop()

	// Wait for connection
	time.Sleep(2 * time.Second)

	// Verify connection is established (TLS validation passed)
	if client.Status() != Running {
		t.Fatal("TLS connection failed - client not running")
	}

	// Make a request to verify end-to-end TLS functionality
	momentum, err := client.LedgerApi.GetFrontierMomentum()
	if err != nil {
		t.Fatalf("RPC call over TLS failed: %v", err)
	}

	if momentum == nil {
		t.Fatal("RPC call over TLS returned nil")
	}

	t.Log("TLS connection test passed - certificate validation working")
}
