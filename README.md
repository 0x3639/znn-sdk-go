# Zenon Go SDK

[![Go Report Card](https://goreportcard.com/badge/github.com/MoonBaZZe/znn-sdk-go)](https://goreportcard.com/report/github.com/MoonBaZZe/znn-sdk-go)
[![GoDoc](https://godoc.org/github.com/MoonBaZZe/znn-sdk-go?status.svg)](https://godoc.org/github.com/MoonBaZZe/znn-sdk-go)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](#contributing)
[![GitHub license](https://img.shields.io/github/license/MoonBaZZe/znn-sdk-go)](LICENSE)

A comprehensive Go SDK for interacting with the Zenon Network. Features a complete implementation of ABI encoding/decoding, embedded contract APIs, wallet management, PoW generation, and an enhanced WebSocket client with auto-reconnect.

Follows the [official Dart SDK](https://github.com/zenon-network/znn_sdk_dart) structure. Tested with Go v1.18+. Compatible with [go-zenon](https://github.com/zenon-network/go-zenon).

## Features

- **Complete ABI Implementation** - Full Solidity ABI encoding/decoding for all Zenon types
- **Embedded Contract APIs** - Type-safe interfaces for all protocol contracts
- **Wallet Management** - BIP39/BIP44 HD wallets with keystore encryption
- **PoW Generation** - Pure Go implementation of Zenon's proof-of-work algorithm
- **Enhanced WebSocket Client** - Auto-reconnect, health monitoring, event callbacks
- **Comprehensive Testing** - 497+ unit tests covering all modules
- **Type Safety** - Leverages Go's type system for compile-time safety

## Installation

```bash
go get github.com/MoonBaZZe/znn-sdk-go
```

**Requirements:**
- Go 1.18 or higher
- Access to a Zenon node (local or remote)

## Quick Start

### Connect to Node (Read-Only)

```go
package main

import (
    "fmt"
    "github.com/MoonBaZZe/znn-sdk-go/rpc_client"
)

func main() {
    // Connect to local node
    client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
    if err != nil {
        panic(err)
    }
    defer client.Stop()

    // Query frontier momentum
    momentum, err := client.LedgerApi.GetFrontierMomentum()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Current height: %d\n", momentum.Height)
}
```

### Connect with Custom Options

```go
opts := rpc_client.ClientOptions{
    AutoReconnect:       true,
    ReconnectDelay:      2 * time.Second,
    MaxReconnectDelay:   60 * time.Second,
    ReconnectAttempts:   10,
    HealthCheckInterval: 15 * time.Second,
}

client, err := rpc_client.NewRpcClientWithOptions("ws://127.0.0.1:35998", opts)
```

### Event Callbacks

```go
client.AddOnConnectionEstablishedCallback(func() {
    fmt.Println("Connected to node!")
})

client.AddOnConnectionLostCallback(func(err error) {
    fmt.Printf("Connection lost: %v\n", err)
})
```

## Wallet Management

### Create New Wallet

```go
import "github.com/MoonBaZZe/znn-sdk-go/wallet"

// Create new keystore with random mnemonic
manager, err := wallet.NewKeyStoreManager("./wallets")
if err != nil {
    panic(err)
}
keystore, err := manager.CreateNew("password123", "my-wallet")
if err != nil {
    panic(err)
}

fmt.Println("Base address:", keystore.GetBaseAddress())
```

### Import from Mnemonic

```go
mnemonic := "route become dream access impulse price inform obtain engage ski believe awful..."
keystore, err := manager.CreateFromMnemonic(mnemonic, "password123", "imported-wallet")
```

### Derive Keypairs

```go
// Get keypair at account index 0
keypair, err := keystore.GetKeyPair(0)
if err != nil {
    panic(err)
}

fmt.Println("Address:", keypair.GetAddress())
fmt.Println("Public key:", hex.EncodeToString(keypair.GetPublicKey()))

// Sign message
signature := keypair.Sign([]byte("Hello Zenon"))
```

### Load Existing Wallet

```go
keystore, err := manager.ReadKeyStore("password123", "my-wallet")
if err != nil {
    panic(err)
}
```

## Embedded Contract APIs

All embedded contracts are accessible via the RPC client:

```go
client, _ := rpc_client.NewRpcClient("ws://127.0.0.1:35998")

// Access contract APIs
client.PillarApi       // Pillar management
client.SentinelApi     // Sentinel operations
client.TokenApi        // Token issuance/management
client.PlasmaApi       // Plasma fusion
client.StakeApi        // Staking operations
client.AcceleratorApi  // Accelerator Z projects
client.SwapApi         // Legacy swap contract
client.BridgeApi       // Bridge operations
client.LiquidityApi    // Liquidity management
client.HtlcApi         // HTLC atomic swaps
client.SporkApi        // Protocol upgrades
```

### Token Operations

```go
// Issue new token
template := client.TokenApi.IssueToken(
    "MyToken",                      // name
    "MTK",                          // symbol
    "mytoken.com",                  // domain
    big.NewInt(1000000 * 1e8),     // totalSupply
    big.NewInt(10000000 * 1e8),    // maxSupply
    8,                              // decimals
    true,                           // isMintable
    true,                           // isBurnable
    false,                          // isUtility
)

// Mint tokens
template := client.TokenApi.Mint(tokenZTS, amount, beneficiary)

// Burn tokens
template := client.TokenApi.Burn(tokenZTS, amount)
```

### Pillar Operations

```go
// Register pillar
template := client.PillarApi.Register(
    "MyPillar",                     // name
    producerAddress,                // block producer address
    rewardAddress,                  // reward recipient
    0,                              // giveBlockRewardPercentage
    50,                             // giveDelegateRewardPercentage
)

// Delegate to pillar
template := client.PillarApi.Delegate("MyPillar")

// Undelegate
template := client.PillarApi.Undelegate()
```

### Plasma Fusion

```go
// Fuse QSR for plasma
template := client.PlasmaApi.Fuse(
    beneficiaryAddress,
    big.NewInt(100 * 1e8), // 100 QSR
)

// Cancel plasma fusion
template := client.PlasmaApi.CancelFuse(fusionId)
```

### Staking

```go
// Stake ZNN
template := client.StakeApi.Stake(
    6 * 30 * 24 * 60 * 60, // 6 months in seconds
    big.NewInt(5000 * 1e8), // 5000 ZNN
)

// Cancel stake
template := client.StakeApi.Cancel(stakeHash)
```

### HTLC (Atomic Swaps)

```go
import "crypto/sha256"

// Create HTLC
preimage := []byte("secret")
hash := sha256.Sum256(preimage)

template := client.HtlcApi.Create(
    types.ZnnTokenStandard,         // token
    big.NewInt(100 * 1e8),         // amount
    hashLockedAddress,              // recipient
    time.Now().Add(24*time.Hour).Unix(), // expiration
    1,                              // SHA-256 hash type
    32,                             // max preimage size
    hash[:],                        // hash lock
)

// Unlock HTLC with preimage
template := client.HtlcApi.Unlock(htlcId, preimage)

// Reclaim expired HTLC
template := client.HtlcApi.Reclaim(htlcId)
```

### Bridge Operations

```go
// Wrap tokens
template := client.BridgeApi.WrapToken(
    networkClass,
    chainId,
    toAddress,
    types.ZnnTokenStandard,
    big.NewInt(100 * 1e8),
)

// Unwrap tokens
template := client.BridgeApi.UnwrapToken(
    types.ZnnTokenStandard,
    big.NewInt(100 * 1e8),
    signature,
)
```

## Sending Transactions

Transactions require a wallet/keypair:

```go
// For contract calls, first get the template
template := client.TokenApi.IssueToken(...)

// Then sign and send (requires wallet integration)
// Implementation depends on your wallet setup
```

## ABI Encoding/Decoding

The SDK includes a complete ABI implementation:

```go
import "github.com/MoonBaZZe/znn-sdk-go/abi"
import "github.com/MoonBaZZe/znn-sdk-go/embedded"

// Encode function call
data, err := embedded.Token.EncodeFunction("IssueToken", []interface{}{
    name,
    symbol,
    domain,
    totalSupply,
    maxSupply,
    decimals,
    isMintable,
    isBurnable,
    isUtility,
})

// Decode function call
functionName, args, err := embedded.Token.DecodeFunction(data)
```

## PoW Generation

Generate proof-of-work for transactions:

```go
import "github.com/MoonBaZZe/znn-sdk-go/pow"

// Generate PoW
hash := types.HexToHashPanic("...")
difficulty := uint64(80000)
nonce := pow.GeneratePoW(hash, difficulty)

// Verify PoW
valid := pow.CheckPoW(hash, nonce, difficulty)
```

## Connection Management

The enhanced WebSocket client provides:

### Status Monitoring

```go
status := client.Status() // Uninitialized, Connecting, Running, Stopped
if client.IsClosed() {
    // Handle disconnection
}
```

### Manual Restart

```go
if err := client.Restart(); err != nil {
    fmt.Printf("Reconnection failed: %v\n", err)
}
```

### Graceful Shutdown

```go
client.Stop() // Stops monitoring, reconnection, and closes connection
```

## Architecture

### Core Modules

- **`abi/`** - Complete Solidity ABI implementation
- **`embedded/`** - Contract definitions and constants
- **`api/`** - Core blockchain APIs (Ledger, Stats, Subscriber)
- **`api/embedded/`** - Embedded contract APIs
- **`wallet/`** - HD wallet management (BIP39/BIP44)
- **`crypto/`** - Cryptographic primitives (Ed25519, SHA3, Argon2)
- **`pow/`** - Proof-of-work generation
- **`rpc_client/`** - Enhanced WebSocket client
- **`utils/`** - Common utilities (bytes, amounts, blocks)

### Type System

The SDK uses go-zenon's type system:

```go
import "github.com/zenon-network/go-zenon/common/types"

types.Address           // Zenon address
types.Hash              // 32-byte hash
types.ZenonTokenStandard // Token standard (ZTS)
types.ZnnTokenStandard  // ZNN token
types.QsrTokenStandard  // QSR token
```

## Testing

The SDK includes comprehensive test coverage:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./abi/...
go test ./wallet/...
go test ./rpc_client/...
```

**Test Statistics:**
- 497+ unit tests
- Covers all public APIs
- 80%+ test coverage

## Examples

See the `examples/` directory for working examples:

### Basic Client Example

Connect to a node and query data:

```bash
cd examples/basic_client
go run main.go
```

This example demonstrates:
- Connecting to a Zenon node
- Registering connection callbacks
- Querying frontier momentum
- Querying network info

### Wallet Management Example

Create and manage wallets:

```bash
cd examples/wallet_management
go run main.go
```

This example demonstrates:
- Creating new wallets
- Deriving keypairs (BIP44)
- Signing and verifying messages
- Loading existing wallets

## Development

### Build

```bash
go build ./...
```

### Format

```bash
go fmt ./...
```

### Lint

```bash
go vet ./...
```

## Documentation

- **[CLAUDE.md](CLAUDE.md)** - Detailed SDK architecture and development guide
- **[roadmap.md](roadmap.md)** - Implementation roadmap and progress
- **[GoDoc](https://godoc.org/github.com/MoonBaZZe/znn-sdk-go)** - API reference

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Based on the [Dart SDK](https://github.com/zenon-network/znn_sdk_dart) structure
- Compatible with [go-zenon](https://github.com/zenon-network/go-zenon)
- Built for the Zenon Network community

## Support

- **Issues**: [GitHub Issues](https://github.com/MoonBaZZe/znn-sdk-go/issues)
- **Community**: [Zenon Network](https://zenon.network)
