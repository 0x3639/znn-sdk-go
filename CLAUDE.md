# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is the Zenon Go SDK implementation, a client library for interacting with Zenon Network nodes. It follows the official Dart SDK structure and is compatible with go-zenon nodes running on the Zenon Network.

## Key Architecture

### Core Components

1. **Zenon Client (`zenon/zenon.go`)** - Main SDK entry point. Manages wallet lifecycle, keypairs, and coordinates with the RPC client. Provides the `Send()` method which handles transaction signing, PoW generation, and submission.

2. **RPC Client (`rpc_client/client.go`)** - Manages WebSocket connection to Zenon nodes and instantiates all API endpoints. All API objects share the same underlying WebSocket client.

3. **API Layer (`api/`)** - Core blockchain APIs:
   - `LedgerApi` - Read ledger state, query account blocks/momentums, and provides `SendTemplate()`/`ReceiveTemplate()` helpers
   - `StatsApi` - Network statistics and metrics
   - `SubscriberApi` - WebSocket subscriptions for real-time events

4. **Embedded Contract APIs (`api/embedded/`)** - Smart contract interaction APIs for protocol-level operations:
   - `PillarApi`, `SentinelApi` - Network infrastructure management
   - `TokenApi` - Token issuance, minting, burning
   - `PlasmaApi` - Plasma fusion for feeless transactions
   - `StakeApi` - Staking operations
   - `AcceleratorApi`, `BridgeApi`, `LiquidityApi`, `SwapApi`, `SporkApi` - Protocol features

5. **Wallet (`wallet/`)** - Keyfile storage and cryptographic operations. Wallet directory determined by `DefaultWalletDir` (platform-specific via `go-zenon/node` package).

### Connection Model

The SDK connects to Zenon nodes via WebSocket (default: `ws://127.0.0.1:35998`).
- **Read operations**: Initialize RpcClient - no wallet required
- **Write operations**: Require wallet/keypair for signing transactions

### Transaction Flow

All transactions follow this pattern:
1. Create transaction template using API method (e.g., `client.LedgerApi.SendTemplate()` or `client.TokenApi.IssueToken()`)
2. Sign and send transaction:
   - Autofill transaction parameters (height, previous hash, momentum acknowledgment)
   - Query required PoW difficulty via `PlasmaApi`
   - Generate PoW nonce if needed (or use fused plasma)
   - Sign transaction with keypair
   - Publish to node via `LedgerApi.PublishRawTransaction()`

## Development Commands

### Build and Dependencies
```bash
# Download dependencies
go mod download

# Tidy dependencies (after adding/removing imports)
go mod tidy

# Build entire SDK
go build ./...

# Build specific package
go build ./zenon
go build ./api/embedded
```

### Code Quality
```bash
# Format all Go code
go fmt ./...

# Run static analysis
go vet ./...

# Run tests (currently no test files exist in this repo)
go test ./...

# Run a single test
go test ./path/to/package -run TestName
```

### Running Examples

Examples demonstrate SDK usage patterns. All require a running Zenon node at `ws://127.0.0.1:35998`.

```bash
# Simple client connection example
go run examples/simple_client/main.go

# Full RPC operations example
go run examples/rpc/main.go

# Wallet management example
go run examples/wallet/main.go

# Real-time subscription example
go run examples/subscribe/main.go

# Genesis generation script
go run examples/scripts/generate_genesis.go
```

## Common Development Patterns

### Initializing the SDK

**Connect to node (read-only):**
```go
import "github.com/0x3639/znn-sdk-go/rpc_client"

client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
if err != nil {
    return err
}
defer client.Stop()

// Query data
momentum, err := client.LedgerApi.GetFrontierMomentum()
```

**With wallet (for transactions):**
```go
import (
    "github.com/0x3639/znn-sdk-go/rpc_client"
    "github.com/0x3639/znn-sdk-go/wallet"
)

// Initialize client
client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
if err != nil {
    return err
}
defer client.Stop()

// Load wallet
manager, err := wallet.NewKeyStoreManager("./wallets")
if err != nil {
    return err
}
keystore, err := manager.ReadKeyStore("password", "my-wallet")
if err != nil {
    return err
}

// Get keypair for signing
keypair, err := keystore.GetKeyPair(0)
if err != nil {
    return err
}

// Now you can sign transactions
```

### Transaction Pattern

All embedded contract methods return `*nom.AccountBlock` templates:

```go
// 1. Create template
template := client.TokenApi.IssueToken(
    name, symbol, domain,
    totalSupply, maxSupply, decimals,
    isMintable, isBurnable, isUtility,
)

// 2. Sign and send transaction
// (Implementation depends on your wallet setup)
// - Autofill parameters
// - Generate PoW if needed
// - Sign with keypair
// - Publish via client.LedgerApi.PublishRawTransaction()
```

For basic transfers, use `LedgerApi.SendTemplate()`:
```go
template := client.LedgerApi.SendTemplate(
    toAddress,
    types.ZnnTokenStandard,  // or types.QsrTokenStandard
    amount,
    []byte{},  // optional data
)
// Then sign and send as above
```

### Reading Blockchain Data

Query methods return data directly:

```go
// Get account info
info, err := client.LedgerApi.GetAccountInfoByAddress(address)

// Get frontier momentum
momentum, err := client.LedgerApi.GetFrontierMomentum()

// Get token info
token, err := client.TokenApi.GetByZts(zts)

// Get unreceived blocks
blocks, err := client.LedgerApi.GetUnreceivedBlocksByAddress(address, 0, 25)
```

### Wallet Management

```go
import "github.com/0x3639/znn-sdk-go/wallet"

// Create wallet manager
manager, err := wallet.NewKeyStoreManager("./wallets")
if err != nil {
    return err
}

// Create new keystore
keystore, err := manager.CreateNew("password", "my-wallet")
if err != nil {
    return err
}

// Load existing keystore
keystore, err := manager.ReadKeyStore("password", "my-wallet")
if err != nil {
    return err
}

fmt.Println("Base address:", keystore.GetBaseAddress())

// Derive keypair at index
keypair, err := keystore.GetKeyPair(0)
if err != nil {
    return err
}
fmt.Println("Address:", keypair.GetAddress())
```

## Important Implementation Details

### Dependencies
- Uses `github.com/zenon-network/go-zenon` for core types, chain structures, and RPC client
- Wallet files stored in platform-specific directory via `go-zenon/node.DefaultDataDir()`
- Keyfiles named by base address unless custom name provided

### Transaction Mechanics
- `zenon/utils.go` contains transaction preparation logic:
  - `autofillTransactionParameters()` - Sets height, previous hash, momentum acknowledgment
  - `checkAndSetFields()` - Validates transaction fields
  - `SetDifficulty()` - Queries required PoW and generates nonce
  - `setHashAndSignature()` - Finalizes transaction
- PoW generation happens synchronously and can take time for high difficulty
- Fused plasma reduces/eliminates PoW requirement

### API Structure
- All embedded APIs return `*nom.AccountBlock` for contract calls
- Template methods package ABI-encoded data for smart contracts
- `LedgerApi.PublishRawTransaction()` submits to node
- Error handling: RPC calls return errors, transaction submission may succeed but fail on-chain

### Type Conversions
Common helper functions:
- `types.ParseAddressPanic()` - String to Address
- `types.HexToHashPanic()` - Hex string to Hash
- `types.ParseZTS()` - String to ZenonTokenStandard
- `big.NewInt(amount * constants.Decimals)` - Convert to base units (8 decimals)