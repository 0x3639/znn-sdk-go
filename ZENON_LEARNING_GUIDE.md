# Zenon Network Go SDK Learning Guide

Welcome to the comprehensive learning guide for the Zenon Network Go SDK! This guide will teach you about Zenon Network, Go programming patterns, and how to build decentralized applications using this SDK.

## Table of Contents

1. [Introduction to Zenon Network](#introduction-to-zenon-network)
2. [Understanding the SDK Architecture](#understanding-the-sdk-architecture)
3. [Go Programming Fundamentals for SDK](#go-programming-fundamentals-for-sdk)
4. [Getting Started](#getting-started)
5. [Core Concepts](#core-concepts)
6. [Building Applications](#building-applications)
7. [Advanced Topics](#advanced-topics)
8. [Resources](#resources)

## Introduction to Zenon Network

### What is Zenon Network?

Zenon Network, also known as the **Network of Momentum (NoM)**, is a revolutionary Layer 1 blockchain that solves the blockchain trilemma through innovative architecture and economic design.

#### Key Features:
- **Feeless Transactions**: No transaction fees through the Plasma mechanism
- **Dual-Ledger Architecture**: Combines block-lattice and DAG structures
- **Dual-Coin Economy**: ZNN and QSR tokens with complementary roles
- **Decentralized Governance**: Protocol-level treasury and community funding

### Network Architecture

```
┌─────────────────────────────────────┐
│       Network of Momentum           │
│  ┌──────────┐      ┌──────────┐    │
│  │  Block   │      │   DAG    │    │
│  │ Lattice  │ <--> │  Layer   │    │
│  └──────────┘      └──────────┘    │
│                                     │
│  ┌──────────────────────────────┐  │
│  │      Consensus Layer         │  │
│  │  ┌─────┐ ┌─────┐ ┌────────┐ │  │
│  │  │ ZNN │ │ QSR │ │ Plasma │ │  │
│  │  └─────┘ └─────┘ └────────┘ │  │
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘
```

### The Dual-Coin Economy

#### ZNN (Zenon)
- **Purpose**: Network security and governance
- **Use Cases**:
  - Running Pillars (validators)
  - Staking for QSR rewards
  - Delegating to Pillars for voting power
  - Governance participation

#### QSR (Quasar)
- **Purpose**: Network utility and anti-spam
- **Use Cases**:
  - Spawning Pillars and Sentinels
  - Fusing into Plasma for feeless transactions
  - Network resource allocation

#### Plasma
- **What it is**: A third-dimensional asset created by fusing QSR
- **Purpose**: Enables feeless, high-throughput transactions
- **How it works**: Acts as computational "gas" without fees

## Understanding the SDK Architecture

### SDK Components Overview

```go
znn-sdk-go/
├── zenon/          # Core client implementation
│   ├── zenon.go    # Main Zenon client
│   ├── utils.go    # Helper functions
│   └── errors.go   # Error definitions
├── rpc_client/     # RPC communication layer
│   └── client.go   # WebSocket client
├── api/            # API implementations
│   ├── ledger.go   # Blockchain operations
│   ├── stats.go    # Network statistics
│   └── embedded/   # Protocol contracts
│       ├── pillar.go
│       ├── token.go
│       ├── plasma.go
│       └── ...
└── wallet/         # Wallet management
    └── ...         # Key handling
```

### Core Design Patterns

#### 1. Client-Server Pattern
The SDK uses WebSocket connections to communicate with Zenon nodes:
```go
client -> WebSocket -> Zenon Node -> Blockchain
```

#### 2. API Abstraction
Each embedded contract has its own API interface:
```go
Client.PillarApi.Register()
Client.TokenApi.IssueToken()
Client.PlasmaApi.Fuse()
```

#### 3. Transaction Flow
```
Create Transaction -> Auto-fill Parameters -> Calculate PoW -> Sign -> Broadcast
```

## Go Programming Fundamentals for SDK

### Essential Go Concepts

#### 1. Error Handling
Go's explicit error handling is crucial in blockchain operations:
```go
result, err := z.Client.LedgerApi.GetAccountInfo(address)
if err != nil {
    // Always handle errors in blockchain operations
    log.Error("Failed to get account info", err)
    return err
}
```

#### 2. Pointers and References
Understanding when the SDK uses pointers:
```go
// SDK often returns pointers to structs
accountBlock *nom.AccountBlock
// Check for nil before using
if accountBlock != nil {
    // Safe to use
}
```

#### 3. Goroutines and Channels
For asynchronous operations and event handling:
```go
// SDK uses channels for termination signals
stopCh := make(chan os.Signal, 1)
go func() {
    <-stopCh
    z.Stop()
}()
```

#### 4. Interfaces
The SDK heavily uses interfaces for flexibility:
```go
// Each API module implements specific methods
type LedgerApi interface {
    GetAccountBlockByHash(hash types.Hash) (*AccountBlock, error)
    PublishRawTransaction(tx *AccountBlock) error
}
```

### Go SDK Patterns

#### Big Numbers
Blockchain values often exceed standard integer limits:
```go
import "math/big"

// ZNN amounts use big.Int
amount := big.NewInt(1000000000) // 1 ZNN (8 decimals)
```

#### Type Safety
The SDK uses custom types for blockchain entities:
```go
// Not just strings, but typed values
var address types.Address
var hash types.Hash
var tokenStandard types.ZenonTokenStandard
```

## Getting Started

### Prerequisites

1. **Go Installation** (v1.18+)
```bash
go version  # Should show 1.18 or higher
```

2. **Zenon Node**
- Local node recommended: `ws://127.0.0.1:35998`
- Or use a public node (check Zenon community resources)

3. **Development Environment**
```bash
# Clone the SDK
git clone https://github.com/MoonBaZZe/znn-sdk-go
cd znn-sdk-go

# Install dependencies
go mod download
```

### Your First Connection

```go
package main

import (
    "fmt"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

func main() {
    // Create client without wallet (read-only)
    z, err := zenon.NewZenon("")
    if err != nil {
        panic(err)
    }
    
    // Connect to node
    err = z.Start("", "ws://127.0.0.1:35998", 0)
    if err != nil {
        panic(err)
    }
    defer z.Stop()
    
    // Get network info
    info, err := z.Client.StatsApi.NetworkInfo()
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Connected to Zenon Network!\n")
    fmt.Printf("Peers: %d\n", info.NumPeers)
}
```

## Core Concepts

### 1. Account-Blocks
Zenon uses an account-chain model where each account has its own blockchain:
```go
// Each account maintains its own chain
type AccountBlock struct {
    Version      uint64
    ChainId      uint64
    BlockType    uint64
    Hash         Hash
    PreviousHash Hash
    Height       uint64
    // ... more fields
}
```

### 2. Momentum
The network's meta-blockchain that references account-chains:
```go
// Momentums are the "global" blocks
momentum, err := z.Client.LedgerApi.GetFrontierMomentum()
fmt.Printf("Current height: %d\n", momentum.Height)
```

### 3. Embedded Contracts
Protocol-level smart contracts that power core functionality:
- **Pillar Contract**: Validator management
- **Plasma Contract**: Feeless transactions
- **Token Contract**: ZTS token creation
- **Accelerator Contract**: Community funding
- **Stake Contract**: ZNN staking
- **Sentinel Contract**: Network security nodes
- **Swap Contract**: Token swapping
- **Bridge Contract**: Cross-chain operations

### 4. Proof of Work (PoW)
Auto-generated for transactions when Plasma is insufficient:
```go
// SDK automatically calculates PoW when needed
// This happens inside z.Send() transparently
```

## Building Applications

### Application Categories

#### 1. Wallet Applications
- Managing keys and addresses
- Sending/receiving ZNN, QSR, and ZTS tokens
- Staking and delegation

#### 2. DeFi Applications
- Token creation and management
- Liquidity provision
- Automated trading bots

#### 3. Governance Tools
- Voting interfaces
- Proposal management
- Pillar monitoring

#### 4. Infrastructure Services
- Node monitoring
- Network statistics
- Block explorers

### Development Workflow

1. **Setup Development Environment**
```bash
mkdir my-zenon-app
cd my-zenon-app
go mod init my-zenon-app
go get github.com/MoonBaZZe/znn-sdk-go
```

2. **Create Application Structure**
```
my-zenon-app/
├── main.go           # Entry point
├── config/           # Configuration
├── services/         # Business logic
│   ├── wallet.go
│   └── blockchain.go
├── models/           # Data structures
└── utils/            # Helpers
```

3. **Implement Core Features**
- Connection management
- Error handling and retries
- Transaction queuing
- Event monitoring

4. **Testing Strategy**
- Unit tests for business logic
- Integration tests with testnet
- Load testing for production

### Example: Simple Wallet Service

```go
package services

import (
    "math/big"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
    "github.com/zenon-network/go-zenon/common/types"
)

type WalletService struct {
    client *zenon.Zenon
}

func NewWalletService(keyFile, password, nodeURL string) (*WalletService, error) {
    z, err := zenon.NewZenon(keyFile)
    if err != nil {
        return nil, err
    }
    
    if err := z.Start(password, nodeURL, 0); err != nil {
        return nil, err
    }
    
    return &WalletService{client: z}, nil
}

func (w *WalletService) GetBalance(address types.Address) (*big.Int, *big.Int, error) {
    info, err := w.client.Client.LedgerApi.GetAccountInfoByAddress(address)
    if err != nil {
        return nil, nil, err
    }
    
    var znnBalance, qsrBalance *big.Int
    
    for _, balance := range info.BalanceInfoList {
        if balance.Token.TokenStandard == types.ZnnTokenStandard {
            znnBalance = balance.Balance
        } else if balance.Token.TokenStandard == types.QsrTokenStandard {
            qsrBalance = balance.Balance
        }
    }
    
    return znnBalance, qsrBalance, nil
}

func (w *WalletService) SendZNN(to types.Address, amount *big.Int) error {
    template := w.client.Client.LedgerApi.SendTemplate(
        to, 
        types.ZnnTokenStandard, 
        amount, 
        []byte{},
    )
    return w.client.Send(template)
}

func (w *WalletService) Close() error {
    return w.client.Stop()
}
```

## Advanced Topics

### 1. Plasma Management
Understanding and optimizing Plasma usage:
```go
// Check Plasma requirements
required, err := z.Client.PlasmaApi.GetRequiredPoWForAccountBlock(params)

// Fuse QSR for Plasma
err = z.Send(z.Client.PlasmaApi.Fuse(beneficiary, qsrAmount))
```

### 2. Pillar Operations
Running and managing validator nodes:
```go
// Register as Pillar
err = z.Send(z.Client.PillarApi.Register(
    name,
    producerAddress,
    rewardAddress,
    giveBlockRewardPercentage,
    giveDelegateRewardPercentage,
))
```

### 3. Token Creation
Issuing custom ZTS tokens:
```go
// Issue a new token
err = z.Send(z.Client.TokenApi.IssueToken(
    tokenName,
    tokenSymbol,
    tokenDomain,
    totalSupply,
    maxSupply,
    decimals,
    isMintable,
    isBurnable,
    isUtility,
))
```

### 4. Event Subscription
Real-time blockchain monitoring:
```go
// Subscribe to account blocks
subscription, err := z.Client.SubscribeApi.ToAccountBlocksByAddress(address)

// Listen for events
for {
    select {
    case block := <-subscription:
        // Process new account block
    }
}
```

### 5. Performance Optimization

#### Connection Pooling
```go
// Reuse connections for multiple operations
type ConnectionPool struct {
    connections []*zenon.Zenon
    // Implementation...
}
```

#### Batch Operations
```go
// Group related operations
func BatchSend(transactions []*nom.AccountBlock) error {
    for _, tx := range transactions {
        // Process in sequence with proper error handling
    }
}
```

#### Caching Strategies
```go
// Cache frequently accessed data
type Cache struct {
    momentums map[uint64]*Momentum
    accounts  map[string]*AccountInfo
    // Implementation with TTL
}
```

## Best Practices

### 1. Security
- **Never hardcode private keys**
- **Use secure key storage (hardware wallets when possible)**
- **Validate all inputs**
- **Implement rate limiting**
- **Use secure RPC connections (wss://)**

### 2. Error Handling
```go
// Always check for specific error types
if err != nil {
    switch {
    case errors.Is(err, zenon.ErrInsufficientBalance):
        // Handle insufficient funds
    case errors.Is(err, zenon.ErrInsufficientPlasma):
        // Handle Plasma shortage
    default:
        // Generic error handling
    }
}
```

### 3. Testing
```go
// Use testnet for development
const TestnetURL = "ws://testnet.zenon.network:35998"

// Mock interfaces for unit testing
type MockLedgerApi struct {
    // Mock implementation
}
```

### 4. Monitoring
```go
// Implement comprehensive logging
logger.Info("Transaction sent", 
    "hash", tx.Hash,
    "amount", amount,
    "to", toAddress,
)

// Track metrics
metrics.Counter("transactions_sent").Inc()
```

## Common Patterns and Solutions

### Pattern 1: Retry Logic
```go
func RetryOperation(operation func() error, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := operation()
        if err == nil {
            return nil
        }
        time.Sleep(time.Second * time.Duration(i+1))
    }
    return fmt.Errorf("operation failed after %d retries", maxRetries)
}
```

### Pattern 2: Transaction Queue
```go
type TransactionQueue struct {
    queue chan *nom.AccountBlock
    client *zenon.Zenon
}

func (tq *TransactionQueue) Process() {
    for tx := range tq.queue {
        err := tq.client.Send(tx)
        if err != nil {
            // Handle error, possibly retry
        }
    }
}
```

### Pattern 3: Balance Monitoring
```go
func MonitorBalance(address types.Address, threshold *big.Int) {
    ticker := time.NewTicker(60 * time.Second)
    for range ticker.C {
        balance, err := getBalance(address)
        if err == nil && balance.Cmp(threshold) < 0 {
            // Alert: balance below threshold
        }
    }
}
```

## Troubleshooting Guide

### Common Issues and Solutions

#### 1. Connection Issues
```
Error: cannot connect to node
```
**Solution**: Ensure node is running and accessible, check firewall settings

#### 2. Insufficient Plasma
```
Error: not enough plasma
```
**Solution**: Fuse QSR for Plasma or wait for Plasma regeneration

#### 3. Invalid Transaction
```
Error: transaction validation failed
```
**Solution**: Check account balance, nonce, and transaction parameters

#### 4. Key File Issues
```
Error: cannot decrypt keyfile
```
**Solution**: Verify password and keyfile integrity

## Resources

### Official Resources
- [Zenon Network Website](https://zenon.network)
- [Zenon Documentation](https://docs.zenon.network)
- [GitHub Repository](https://github.com/zenon-network)

### Community Resources
- Zenon Forum
- Discord/Telegram Communities
- Developer Documentation

### SDK Specific
- [Go SDK Repository](https://github.com/MoonBaZZe/znn-sdk-go)
- [Example Applications](./examples/)
- [API Reference](https://godoc.org/github.com/MoonBaZZe/znn-sdk-go)

### Learning Path

1. **Week 1**: Understand Zenon Network basics and run examples
2. **Week 2**: Build a simple wallet application
3. **Week 3**: Implement token operations and Plasma management
4. **Week 4**: Create a complete DApp with advanced features

## Next Steps

1. **Explore Tutorials**: Check the `tutorials/` directory for hands-on guides
2. **Run Examples**: Try the example applications in `examples/`
3. **Join Community**: Connect with other developers
4. **Build Something**: Start with a simple project and expand

## Contributing

The Zenon ecosystem welcomes contributions! Whether it's improving documentation, adding examples, or building applications, your input is valuable.

---

*This guide is a living document. As you learn and build with the Zenon Go SDK, consider contributing your insights and examples back to the community.*