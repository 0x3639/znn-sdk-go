# Tutorial 01: Setup and Connection

This tutorial will guide you through setting up your development environment and making your first connection to the Zenon Network.

## Prerequisites

### 1. Install Go
Ensure you have Go 1.18 or higher installed:
```bash
go version
```

If not installed, download from [https://golang.org/dl/](https://golang.org/dl/)

### 2. Setup Zenon Node
You have two options:

#### Option A: Run a Local Node (Recommended)
1. Download the latest Zenon node from [GitHub Releases](https://github.com/zenon-network/go-zenon/releases)
2. Extract and run:
```bash
./znnd
```
3. Node will be available at `ws://127.0.0.1:35998`

#### Option B: Use a Public Node
Check Zenon community resources for public node URLs. Note: Local node is more reliable and faster.

## Project Setup

### Step 1: Create Your Project
```bash
mkdir zenon-tutorial
cd zenon-tutorial
go mod init zenon-tutorial
```

### Step 2: Install the SDK
```bash
go get github.com/MoonBaZZe/znn-sdk-go
```

### Step 3: Verify Installation
Create `test.go`:
```go
package main

import (
    "fmt"
    _ "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

func main() {
    fmt.Println("SDK imported successfully!")
}
```

Run it:
```bash
go run test.go
```

## Your First Connection

### Basic Connection (Read-Only)

Create `connect.go`:
```go
package main

import (
    "fmt"
    "log"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

func main() {
    // Create a client without a wallet (read-only mode)
    z, err := zenon.NewZenon("")
    if err != nil {
        log.Fatal("Failed to create Zenon client:", err)
    }
    
    // Connect to the node
    // Parameters: password (empty for no wallet), URL, key index
    err = z.Start("", "ws://127.0.0.1:35998", 0)
    if err != nil {
        log.Fatal("Failed to connect:", err)
    }
    
    // Always clean up
    defer func() {
        if err := z.Stop(); err != nil {
            log.Println("Error stopping client:", err)
        }
    }()
    
    fmt.Println("Successfully connected to Zenon Network!")
    
    // Test the connection
    networkInfo, err := z.Client.StatsApi.NetworkInfo()
    if err != nil {
        log.Fatal("Failed to get network info:", err)
    }
    
    fmt.Printf("Network Info:\n")
    fmt.Printf("- Peers: %d\n", networkInfo.NumPeers)
    fmt.Printf("- Public Key: %s\n", networkInfo.Self.PublicKey)
}
```

Run it:
```bash
go run connect.go
```

Expected output:
```
Successfully connected to Zenon Network!
Network Info:
- Peers: 8
- Public Key: [your_node_public_key]
```

## Understanding the Connection

### Connection Parameters

```go
z.Start(password, url, keyIndex)
```

- **password**: Decrypt wallet (empty if no wallet)
- **url**: WebSocket URL of the node
- **keyIndex**: Derivation index for HD wallet (usually 0)

### Connection States

1. **Created**: Client instantiated but not connected
2. **Connected**: WebSocket connection established
3. **Authenticated**: Wallet decrypted (if provided)
4. **Stopped**: Connection closed

## Advanced Connection Example

Create `advanced_connect.go`:
```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

func main() {
    // Connection with retry logic
    var z *zenon.Zenon
    var err error
    
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        z, err = connectWithTimeout("ws://127.0.0.1:35998", 5*time.Second)
        if err == nil {
            break
        }
        
        log.Printf("Connection attempt %d failed: %v", i+1, err)
        if i < maxRetries-1 {
            time.Sleep(2 * time.Second)
        }
    }
    
    if err != nil {
        log.Fatal("Failed to connect after retries:", err)
    }
    
    defer z.Stop()
    
    // Get chain identifier
    chainId, err := z.Client.LedgerApi.GetFrontierMomentum()
    if err != nil {
        log.Fatal("Failed to get chain info:", err)
    }
    
    fmt.Printf("Connected to chain with height: %d\n", chainId.Height)
    
    // Get sync status
    syncInfo, err := z.Client.StatsApi.SyncInfo()
    if err != nil {
        log.Fatal("Failed to get sync info:", err)
    }
    
    fmt.Printf("Sync Status:\n")
    fmt.Printf("- Current Height: %d\n", syncInfo.CurrentHeight) 
    fmt.Printf("- Target Height: %d\n", syncInfo.TargetHeight)
    
    if syncInfo.CurrentHeight == syncInfo.TargetHeight {
        fmt.Println("✓ Node is fully synced")
    } else {
        progress := float64(syncInfo.CurrentHeight) / float64(syncInfo.TargetHeight) * 100
        fmt.Printf("⟳ Syncing... %.2f%%\n", progress)
    }
}

func connectWithTimeout(url string, timeout time.Duration) (*zenon.Zenon, error) {
    z, err := zenon.NewZenon("")
    if err != nil {
        return nil, err
    }
    
    // Create a channel for the connection result
    done := make(chan error, 1)
    
    go func() {
        done <- z.Start("", url, 0)
    }()
    
    // Wait for connection or timeout
    select {
    case err := <-done:
        if err != nil {
            return nil, err
        }
        return z, nil
    case <-time.After(timeout):
        z.Stop()
        return nil, fmt.Errorf("connection timeout after %v", timeout)
    }
}
```

## Connection Best Practices

### 1. Always Handle Cleanup
```go
defer z.Stop()
```

### 2. Check Node Sync Status
Before performing operations, ensure the node is synced:
```go
syncInfo, _ := z.Client.StatsApi.SyncInfo()
if syncInfo.CurrentHeight < syncInfo.TargetHeight {
    // Node is still syncing
}
```

### 3. Implement Reconnection Logic
```go
func maintainConnection(url string) {
    for {
        z, err := zenon.NewZenon("")
        if err == nil {
            err = z.Start("", url, 0)
            if err == nil {
                // Connected successfully
                // Do work...
                z.Stop()
            }
        }
        
        // Wait before reconnecting
        time.Sleep(5 * time.Second)
    }
}
```

### 4. Use Environment Variables
```go
import "os"

nodeURL := os.Getenv("ZENON_NODE_URL")
if nodeURL == "" {
    nodeURL = "ws://127.0.0.1:35998" // Default
}
```

## Testing Your Connection

Create `connection_test.go`:
```go
package main

import (
    "testing"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

func TestConnection(t *testing.T) {
    z, err := zenon.NewZenon("")
    if err != nil {
        t.Fatal("Failed to create client:", err)
    }
    
    err = z.Start("", "ws://127.0.0.1:35998", 0)
    if err != nil {
        t.Fatal("Failed to connect:", err)
    }
    defer z.Stop()
    
    // Test API call
    _, err = z.Client.StatsApi.NetworkInfo()
    if err != nil {
        t.Fatal("Failed to get network info:", err)
    }
    
    t.Log("Connection test passed")
}

func TestInvalidConnection(t *testing.T) {
    z, err := zenon.NewZenon("")
    if err != nil {
        t.Fatal("Failed to create client:", err)
    }
    
    // Try invalid URL
    err = z.Start("", "ws://invalid:9999", 0)
    if err == nil {
        t.Fatal("Expected error for invalid URL")
    }
    
    t.Log("Invalid connection handled correctly")
}
```

Run tests:
```bash
go test -v
```

## Troubleshooting

### Common Issues

#### 1. Connection Refused
```
Error: dial tcp 127.0.0.1:35998: connect: connection refused
```
**Solution**: Ensure your Zenon node is running

#### 2. Wrong Protocol
```
Error: websocket: bad handshake
```
**Solution**: Use `ws://` not `http://`

#### 3. Node Not Synced
Check sync status and wait for full sync before operations

#### 4. Timeout Issues
Increase timeout or check network connectivity

## What's Next?

Now that you can connect to the Zenon Network, you're ready to:
- Learn wallet management (Tutorial 02)
- Read blockchain data (Tutorial 03)
- Send transactions (Tutorial 04)

## Exercise

Try modifying the connection example to:
1. Connect to a different node
2. Display more network statistics
3. Implement automatic reconnection on disconnect
4. Create a connection pool for multiple simultaneous connections

## Summary

You've learned:
- ✅ How to set up a Zenon development environment
- ✅ Making connections to Zenon nodes
- ✅ Handling connection errors and retries
- ✅ Checking node sync status
- ✅ Best practices for connection management

Next tutorial: [02-wallet-management.md](./02-wallet-management.md)