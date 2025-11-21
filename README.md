# Zenon Go SDK

[![SDK Status](https://img.shields.io/badge/SDK%20Status-98%25%20Complete-brightgreen)](SDK_STATUS.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/0x3639/znn-sdk-go)](https://goreportcard.com/report/github.com/0x3639/znn-sdk-go)
[![GoDoc](https://godoc.org/github.com/0x3639/znn-sdk-go?status.svg)](https://godoc.org/github.com/0x3639/znn-sdk-go)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](#contributing)
[![GitHub license](https://img.shields.io/github/license/0x3639/znn-sdk-go)](LICENSE)

A comprehensive Go SDK for interacting with the Zenon Network. Features a complete implementation of ABI encoding/decoding, embedded contract APIs, wallet management, PoW generation, and an enhanced WebSocket client with auto-reconnect.

Follows the [official Dart SDK](https://github.com/zenon-network/znn_sdk_dart) structure. Tested with Go v1.18+. Compatible with [go-zenon](https://github.com/zenon-network/go-zenon).

## Features

- **Complete ABI Implementation** - Full Solidity ABI encoding/decoding for all Zenon types
- **Embedded Contract APIs** - Type-safe interfaces for all protocol contracts
- **Wallet Management** - BIP39/BIP44 HD wallets with keystore encryption
- **PoW Generation** - Pure Go implementation of Zenon's proof-of-work algorithm
- **Enhanced WebSocket Client** - Auto-reconnect, health monitoring, event callbacks
- **Comprehensive Documentation** - 96+ Example functions demonstrating all SDK capabilities
- **Type Safety** - Leverages Go's type system for compile-time safety

## SDK Status

**Implementation Status:** 98% Complete | [View Detailed Status](SDK_STATUS.md)

This SDK is production-ready with comprehensive test coverage, security audits, and CI/CD pipeline. See [SDK_STATUS.md](SDK_STATUS.md) for detailed feature-by-feature implementation status including:

- ✅ All 11 embedded contract APIs (100% coverage)
- ✅ 568+ unit tests with excellent coverage
- ✅ 89+ runnable example functions
- ✅ Security audits (BIP39, memory protection)
- ✅ Production CI/CD with 7 automated jobs
- ✅ Multi-platform support (Linux, macOS, Windows)

## Installation

```bash
go get github.com/0x3639/znn-sdk-go
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
    "log"
    "github.com/0x3639/znn-sdk-go/rpc_client"
)

func main() {
    // Connect to local node
    client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    // Query frontier momentum
    momentum, err := client.LedgerApi.GetFrontierMomentum()
    if err != nil {
        log.Fatal(err)
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

## API Overview

All embedded contracts are accessible via the RPC client:

```go
client, _ := rpc_client.NewRpcClient("ws://127.0.0.1:35998")

// Core APIs
client.LedgerApi       // Account blocks, momentums, transactions
client.StatsApi        // Network statistics
client.SubscriberApi   // Real-time event subscriptions

// Embedded Contract APIs
client.PillarApi       // Pillar registration and delegation
client.SentinelApi     // Sentinel node operations
client.TokenApi        // ZTS token issuance and management
client.PlasmaApi       // Plasma fusion for feeless transactions
client.StakeApi        // ZNN staking for rewards
client.AcceleratorApi  // Accelerator-Z ecosystem funding
client.SwapApi         // Legacy asset migration
client.BridgeApi       // Cross-chain bridge operations
client.LiquidityApi    // Liquidity pool management
client.HtlcApi         // Hash Time Locked Contracts (atomic swaps)
client.SporkApi        // Protocol activation mechanism
```

## Examples

The SDK includes 96+ runnable Example functions demonstrating all APIs. View examples:
- **[On pkg.go.dev](https://pkg.go.dev/github.com/0x3639/znn-sdk-go)** - Browse all examples with documentation
- **Run locally**: `go test -run Example` in any package directory

### Example: Query Account Balance

```go
import (
    "fmt"
    "log"
    "github.com/0x3639/znn-sdk-go/rpc_client"
    "github.com/zenon-network/go-zenon/common/types"
)

func Example_queryAccountBalance() {
    client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    address := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")
    info, err := client.LedgerApi.GetAccountInfoByAddress(address)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Address: %s\n", info.Address)
    for token, balance := range info.BalanceInfoMap {
        fmt.Printf("  %s: %s\n", token, balance.Balance)
    }
}
```

### Example: Send Transaction

```go
func Example_sendTransaction() {
    client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    toAddress := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")
    amount := big.NewInt(100 * 100000000) // 100 ZNN

    template := client.LedgerApi.SendTemplate(
        toAddress,
        types.ZnnTokenStandard,
        amount,
        []byte{}, // optional data
    )

    // Template must be:
    // 1. Autofilled (height, previous hash, momentum acknowledgment)
    // 2. Enhanced with PoW or use fused plasma
    // 3. Signed with keypair
    // 4. Published via client.LedgerApi.PublishRawTransaction()

    fmt.Println("Transaction template created")
    fmt.Printf("Amount: %s ZNN\n", amount)
}
```

### Example: Issue Token

```go
func Example_issueToken() {
    client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    template := client.TokenApi.IssueToken(
        "MyToken",                     // name
        "MTK",                         // symbol
        "mytoken.com",                 // domain
        big.NewInt(1000000 * 1e8),    // totalSupply
        big.NewInt(10000000 * 1e8),   // maxSupply
        8,                             // decimals
        true,                          // isMintable
        true,                          // isBurnable
        false,                         // isUtility
    )

    fmt.Println("Token issuance template created")
    fmt.Println("Cost: 1 ZNN (burned as protocol fee)")
}
```

### Example: Stake ZNN

```go
func Example_stakeZNN() {
    client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    // Stake for 12 months (highest rewards)
    duration := int64(31536000) // 365 days in seconds
    amount := big.NewInt(5000 * 100000000) // 5000 ZNN

    template := client.StakeApi.Stake(duration, amount)

    fmt.Println("Staking template created")
    fmt.Printf("Duration: 12 months\n")
    fmt.Printf("Amount: %s ZNN\n", amount)
}
```

### Example: Delegate to Pillar

```go
func Example_delegateToPillar() {
    client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    pillarName := "MyFavoritePillar"
    template := client.PillarApi.Delegate(pillarName)

    fmt.Printf("Delegation template created for Pillar: %s\n", pillarName)
    fmt.Println("Delegation becomes active after 2 momentum confirmations")
}
```

### Example: Subscribe to Momentums

```go
func Example_subscribeMomentums() {
    ctx := context.Background()
    client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    // Subscribe to momentum events
    sub, momentumChan, err := client.SubscriberApi.ToMomentums(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer sub.Unsubscribe()

    fmt.Println("Monitoring new momentums...")

    timeout := time.After(30 * time.Second)
    for {
        select {
        case momentums := <-momentumChan:
            for _, m := range momentums {
                fmt.Printf("New Momentum - Height: %d, Hash: %s\n",
                    m.Height, m.Hash.String()[:16]+"...")
            }
        case <-timeout:
            fmt.Println("Monitoring complete")
            return
        }
    }
}
```

## Wallet Management

### Create New Wallet

```go
import "github.com/0x3639/znn-sdk-go/wallet"

// Create new keystore with random mnemonic
manager, err := wallet.NewKeyStoreManager("./wallets")
if err != nil {
    log.Fatal(err)
}
keystore, err := manager.CreateNew("password123", "my-wallet")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Base address:", keystore.GetBaseAddress())
fmt.Println("Mnemonic:", keystore.GetMnemonic())
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
    log.Fatal(err)
}

fmt.Println("Address:", keypair.GetAddress())
fmt.Println("Public key:", hex.EncodeToString(keypair.GetPublicKey()))

// Sign message
signature := keypair.Sign([]byte("Hello Zenon"))

// Verify signature
valid := keypair.Verify(signature, []byte("Hello Zenon"))
```

### Load Existing Wallet

```go
keystore, err := manager.ReadKeyStore("password123", "my-wallet")
if err != nil {
    log.Fatal(err)
}
```

## Transaction Flow

All Zenon transactions follow this pattern:

1. **Create Template** - Use API methods to create transaction templates
2. **Autofill** - Set height, previous hash, momentum acknowledgment
3. **PoW/Plasma** - Generate PoW nonce or use fused plasma
4. **Sign** - Sign transaction with keypair
5. **Publish** - Submit via `LedgerApi.PublishRawTransaction()`

```go
// 1. Create template
template := client.TokenApi.IssueToken(...)

// 2. Autofill transaction parameters
// (Implementation in zenon/utils.go)

// 3. Generate PoW or use plasma
difficulty := 80000
nonce := pow.GeneratePoW(template.Hash, difficulty)

// 4. Sign transaction
signature := keypair.Sign(template.Hash.Bytes())

// 5. Publish
err := client.LedgerApi.PublishRawTransaction(template)
```

## PoW Generation

Generate proof-of-work for transactions:

### Synchronous PoW

```go
import "github.com/0x3639/znn-sdk-go/pow"

// Generate PoW (blocking)
hash := types.HexToHashPanic("...")
difficulty := uint64(80000)
nonce := pow.GeneratePoW(hash, difficulty)

// Verify PoW
valid := pow.CheckPoW(hash, nonce, difficulty)
```

### Asynchronous PoW (Recommended)

```go
import (
    "context"
    "time"
    "github.com/0x3639/znn-sdk-go/pow"
)

// Generate PoW asynchronously with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resultChan := pow.GeneratePowAsync(ctx, hash, difficulty)
result := <-resultChan

if result.Error != nil {
    if result.Error == pow.ErrCancelled {
        fmt.Println("PoW was cancelled or timed out")
    } else {
        fmt.Printf("PoW generation failed: %v\n", result.Error)
    }
    return
}

fmt.Println("PoW nonce:", result.Nonce)
```

### Context-Based Cancellation

```go
// User-cancellable PoW generation
ctx, cancel := context.WithCancel(context.Background())

// Start PoW in background
resultChan := pow.GeneratePowAsync(ctx, hash, difficulty)

// Simulate user clicking "Cancel" button
go func() {
    time.Sleep(5 * time.Second)
    cancel() // This stops the PoW generation
}()

result := <-resultChan
if result.Error == pow.ErrCancelled {
    fmt.Println("User cancelled PoW generation")
}
```

## Real-Time Subscriptions

Monitor blockchain events in real-time:

```go
// Subscribe to momentums
sub, momentumChan, err := client.SubscriberApi.ToMomentums(ctx)
defer sub.Unsubscribe()

// Subscribe to all account blocks
sub, blockChan, err := client.SubscriberApi.ToAllAccountBlocks(ctx)

// Subscribe to specific address
sub, blockChan, err := client.SubscriberApi.ToAccountBlocksByAddress(ctx, address)

// Subscribe to unreceived blocks (for auto-receive)
sub, blockChan, err := client.SubscriberApi.ToUnreceivedAccountBlocksByAddress(ctx, address)
```

## Connection Management

The enhanced WebSocket client provides robust connection handling:

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

- **`rpc_client/`** - Enhanced WebSocket client with auto-reconnect
- **`api/`** - Core blockchain APIs (Ledger, Stats, Subscriber)
- **`api/embedded/`** - Embedded contract APIs (11 contracts)
- **`wallet/`** - HD wallet management (BIP39/BIP44)
- **`pow/`** - Proof-of-work generation (sync and async)
- **`crypto/`** - Cryptographic primitives (Ed25519, SHA3, Argon2)
- **`abi/`** - Complete Solidity ABI implementation
- **`embedded/`** - Contract definitions and constants
- **`utils/`** - Common utilities (bytes, amounts, blocks)

### Type System

The SDK uses go-zenon's type system:

```go
import "github.com/zenon-network/go-zenon/common/types"

types.Address           // Zenon address (z1...)
types.Hash              // 32-byte hash
types.ZenonTokenStandard // Token standard (ZTS)
types.ZnnTokenStandard  // ZNN token constant
types.QsrTokenStandard  // QSR token constant
```

### Embedded Contracts

Zenon's protocol is powered by 11 embedded smart contracts:

| Contract | Purpose | Key Functions |
|----------|---------|---------------|
| **Pillar** | Consensus nodes | Register, Delegate, Undelegate |
| **Sentinel** | Infrastructure nodes | Register, Revoke, Collect rewards |
| **Token** | ZTS tokens | Issue, Mint, Burn, Update |
| **Plasma** | Feeless transactions | Fuse, Cancel fusion |
| **Stake** | Staking rewards | Stake, Cancel, Collect |
| **Accelerator** | Ecosystem funding | Create project, Vote, Add phase |
| **Bridge** | Cross-chain | Wrap, Unwrap, Set metadata |
| **Liquidity** | Liquidity pools | Add liquidity, Withdraw, Swap |
| **HTLC** | Atomic swaps | Create, Unlock, Reclaim |
| **Swap** | Legacy migration | Query swap assets |
| **Spork** | Protocol upgrades | Query active sporks |

## Token Standards (ZTS)

Zenon uses the ZTS (Zenon Token Standard) for all tokens:

- **ZNN**: Native coin, used for gas and staking
- **QSR**: Quasar token, used for plasma fusion
- **Custom ZTS**: User-issued tokens via TokenApi

All amounts use 8 decimals (base units):
```go
1 ZNN = 100000000 base units
amount := big.NewInt(100 * 100000000) // 100 ZNN
```

## Common Patterns

### Check Account Balance

```go
info, err := client.LedgerApi.GetAccountInfoByAddress(address)
for zts, balance := range info.BalanceInfoMap {
    fmt.Printf("%s: %s\n", zts, balance.Balance)
}
```

### List Unreceived Transactions

```go
blocks, err := client.LedgerApi.GetUnreceivedBlocksByAddress(address, 0, 25)
for _, block := range blocks.List {
    fmt.Printf("Unreceived: %s ZNN from %s\n", block.Amount, block.Address)
}
```

### Auto-Receive Payments

```go
sub, blockChan, err := client.SubscriberApi.ToUnreceivedAccountBlocksByAddress(ctx, myAddress)
defer sub.Unsubscribe()

for blocks := range blockChan {
    for _, block := range blocks {
        // Create receive block
        receiveTemplate := client.LedgerApi.ReceiveTemplate(block.Hash)
        // Sign and publish...
    }
}
```

### Query Pillar Rewards

```go
pillar, err := client.PillarApi.GetByName("MyPillar")
uncollected, err := client.PillarApi.GetUncollectedReward(pillar.StakeAddress)
fmt.Printf("Uncollected rewards: %s ZNN, %s QSR\n",
    uncollected.Znn, uncollected.Qsr)
```

### Fuse Plasma for Feeless Transactions

```go
// Fuse 100 QSR for beneficiary address
amount := big.NewInt(100 * 100000000)
template := client.PlasmaApi.Fuse(beneficiaryAddress, amount)

// Check plasma
plasma, err := client.PlasmaApi.Get(beneficiaryAddress)
fmt.Printf("Current plasma: %d\n", plasma.CurrentPlasma)
```

## Documentation

- **[pkg.go.dev](https://pkg.go.dev/github.com/0x3639/znn-sdk-go)** - Complete API reference with 96+ examples
- **[CLAUDE.md](CLAUDE.md)** - Detailed architecture and development guide

### Running Example Functions

All Example functions are runnable tests:

```bash
# List all examples
go test -list Example ./...

# Run all examples in a package
go test -run Example ./api/embedded

# Run specific example
go test -run Example_issueToken ./api/embedded

# Run all examples across SDK
go test -run Example ./...
```

Examples include:
- **RPC Client** (11 examples): Connection, options, callbacks, monitoring
- **Wallet** (10 examples): Creation, import, keypair derivation, signing
- **Ledger API** (13 examples): Balance, blocks, send, receive, account info
- **Token API** (11 examples): Issue, mint, burn, transfer ownership
- **Staking API** (11 examples): Stake, cancel, collect rewards
- **Pillar API** (9 examples): List, register, delegate, rewards
- **Sentinel API** (9 examples): Register, revoke, deposit QSR
- **Plasma API** (10 examples): Fuse, cancel, query plasma
- **Subscription API** (8 examples): Momentums, account blocks, payment gateway
- **Accelerator API** (6 examples): Projects, voting, donations
- **Plus**: HTLC, Bridge, Liquidity examples

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

### Test

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run examples only
go test -run Example ./...
```

## Troubleshooting

### Connection Issues

**Problem**: Cannot connect to node
```
Solution: Ensure Zenon node is running at ws://127.0.0.1:35998
Check with: curl http://127.0.0.1:35997
```

**Problem**: Connection drops frequently
```go
// Solution: Enable auto-reconnect
opts := rpc_client.ClientOptions{
    AutoReconnect: true,
    ReconnectAttempts: 10,
}
client, err := rpc_client.NewRpcClientWithOptions(url, opts)
```

### Wallet Issues

**Problem**: Cannot load wallet
```
Solution: Verify wallet file exists and password is correct
Check location: wallet.DefaultWalletDir()
```

**Problem**: Invalid address derivation
```
Solution: Ensure correct BIP44 path index
First address: keystore.GetKeyPair(0)
```

### Transaction Issues

**Problem**: Transaction fails with "insufficient plasma"
```go
// Solution 1: Fuse QSR for plasma
template := client.PlasmaApi.Fuse(myAddress, qsrAmount)

// Solution 2: Generate PoW
nonce := pow.GeneratePoW(hash, difficulty)
```

**Problem**: Transaction not confirmed
```
Solution: Wait 2 momentums (~20 seconds) for confirmation
Use SubscriberApi.ToMomentums(ctx) to monitor confirmations
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Add tests for new functionality
4. Ensure all tests pass (`go test ./...`)
5. Format code (`go fmt ./...`)
6. Run linter (`go vet ./...`)
7. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Development Methodology

This SDK was developed using AI-assisted development with rigorous quality assurance:

### Development Process
1. **Initial Implementation** - Code generated using Claude (Anthropic) in an iterative process
2. **Port from Official SDK** - Systematically ported from the [official Dart SDK](https://github.com/zenon-network/znn_sdk_dart) to maintain API compatibility
3. **Security Audit** - Comprehensive security review performed by Grok 4.1 (xAI)
4. **Iterative Refinement** - Issues identified by audit addressed through multiple Claude iterations
5. **Comprehensive Testing** - 568+ unit tests, integration tests, and fuzz tests to verify correctness
6. **Manual Verification** - Real-world testing against live Zenon nodes to ensure full functionality

### Quality Assurance
- ✅ AI-generated code with human oversight
- ✅ Independent AI security audit (Grok 4.1)
- ✅ Multi-platform CI/CD testing (Linux, macOS, Windows)
- ✅ Static analysis (gosec, staticcheck, govulncheck)
- ✅ Production testing on live network
- ✅ Comprehensive documentation and examples

This AI-assisted approach enabled rapid development while maintaining high code quality through automated testing, security scanning, and iterative refinement.

## Attribution

This SDK is based on the original Go SDK created by MoonBaZZe:

- **Original Repository**: https://github.com/MoonBaZZe/znn-sdk-go
- **Original Author**: MoonBaZZe (2022)
- **License**: MIT License

This repository is an independently maintained and enhanced version with additional features including:
- Enhanced security features and comprehensive security audits
- 96+ documented example functions
- Robust WebSocket client with auto-reconnect and context support
- Advanced subscription management with best practices
- Comprehensive CI/CD pipeline and testing
- Active maintenance and community support

## Acknowledgments

- Based on the [Dart SDK](https://github.com/zenon-network/znn_sdk_dart) structure
- Compatible with [go-zenon](https://github.com/zenon-network/go-zenon)
- Built for the Zenon Network community

## Support

- **Issues**: [GitHub Issues](https://github.com/0x3639/znn-sdk-go/issues)
- **Community**: [Zenon Network](https://zenon.network)
- **Documentation**: [pkg.go.dev](https://pkg.go.dev/github.com/0x3639/znn-sdk-go)
