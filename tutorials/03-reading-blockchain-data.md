# Tutorial 03: Reading Blockchain Data

Learn how to query and read data from the Zenon blockchain, including account information, transactions, momentum blocks, and embedded contract states.

## Overview of Zenon Data Structure

### Key Components
1. **Account-Chains**: Each address has its own blockchain
2. **Momentums**: Meta-blocks that reference account-chains
3. **Account Blocks**: Individual transactions in account-chains
4. **Embedded Contracts**: Protocol-level smart contracts

## Getting Account Information

### Basic Account Query

```go
package main

import (
    "fmt"
    "log"
    "math/big"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
    "github.com/zenon-network/go-zenon/common/types"
)

func main() {
    // Connect without wallet (read-only)
    z, err := zenon.NewZenon("")
    if err != nil {
        log.Fatal(err)
    }
    
    err = z.Start("", "ws://127.0.0.1:35998", 0)
    if err != nil {
        log.Fatal(err)
    }
    defer z.Stop()
    
    // Example address
    address, _ := types.ParseAddress("z1qxemdeddedxpyllarxxxxxxxxxxxxxxxsy3fmg")
    
    // Get account info
    accountInfo, err := z.Client.LedgerApi.GetAccountInfoByAddress(address)
    if err != nil {
        log.Fatal("Failed to get account info:", err)
    }
    
    fmt.Printf("Account: %s\n", address.String())
    fmt.Printf("Account Height: %d\n", accountInfo.AccountHeight)
    
    // Display balances
    fmt.Println("\nBalances:")
    for _, balance := range accountInfo.BalanceInfoList {
        amount := new(big.Float).SetInt(balance.Balance)
        decimals := new(big.Float).SetInt(big.NewInt(1e8)) // 8 decimals
        displayAmount := new(big.Float).Quo(amount, decimals)
        
        fmt.Printf("- %s: %s (%s raw)\n", 
            balance.Token.Symbol, 
            displayAmount.String(),
            balance.Balance.String())
    }
}
```

### Detailed Account Information

```go
package main

import (
    "fmt"
    "log"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
    "github.com/zenon-network/go-zenon/common/types"
)

type AccountExplorer struct {
    client *zenon.Zenon
}

func NewAccountExplorer(nodeURL string) (*AccountExplorer, error) {
    z, err := zenon.NewZenon("")
    if err != nil {
        return nil, err
    }
    
    err = z.Start("", nodeURL, 0)
    if err != nil {
        return nil, err
    }
    
    return &AccountExplorer{client: z}, nil
}

func (ae *AccountExplorer) GetFullAccountDetails(address types.Address) error {
    // Get basic account info
    info, err := ae.client.Client.LedgerApi.GetAccountInfoByAddress(address)
    if err != nil {
        return err
    }
    
    fmt.Printf("=== Account Details for %s ===\n", address.String())
    fmt.Printf("Account Height: %d\n", info.AccountHeight)
    
    // Get all token balances
    if len(info.BalanceInfoList) > 0 {
        fmt.Println("\nToken Balances:")
        for _, balance := range info.BalanceInfoList {
            fmt.Printf("  %s (%s): %s\n",
                balance.Token.Symbol,
                balance.Token.TokenStandard.String(),
                balance.Balance.String())
        }
    }
    
    // Get frontier (latest) block
    frontier, err := ae.client.Client.LedgerApi.GetFrontierAccountBlock(address)
    if err == nil && frontier != nil {
        fmt.Println("\nLatest Account Block:")
        fmt.Printf("  Hash: %s\n", frontier.Hash.String())
        fmt.Printf("  Height: %d\n", frontier.Height)
        fmt.Printf("  Momentum Ack: %d\n", frontier.MomentumAcknowledged.Height)
    }
    
    // Check for unreceived transactions
    unreceived, err := ae.client.Client.LedgerApi.GetUnreceivedBlocksByAddress(address, 0, 10)
    if err == nil && unreceived.Count > 0 {
        fmt.Printf("\nUnreceived Transactions: %d\n", unreceived.Count)
        for _, block := range unreceived.List {
            fmt.Printf("  From: %s, Amount: %s %s\n",
                block.Address.String(),
                block.Amount.String(),
                block.TokenName)
        }
    }
    
    return nil
}

func main() {
    explorer, err := NewAccountExplorer("ws://127.0.0.1:35998")
    if err != nil {
        log.Fatal(err)
    }
    defer explorer.client.Stop()
    
    // Check Pillar contract address
    pillarAddress, _ := types.ParseAddress("z1qxemdeddedxpyllarxxxxxxxxxxxxxxxsy3fmg")
    explorer.GetFullAccountDetails(pillarAddress)
}
```

## Reading Account Blocks (Transactions)

### Get Transaction History

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
    "github.com/zenon-network/go-zenon/common/types"
)

func getTransactionHistory(z *zenon.Zenon, address types.Address, pageIndex, pageSize uint32) {
    // Get account blocks by height
    blocks, err := z.Client.LedgerApi.GetAccountBlocksByHeight(address, pageIndex, pageSize)
    if err != nil {
        log.Fatal("Failed to get blocks:", err)
    }
    
    fmt.Printf("Transaction History for %s\n", address.String())
    fmt.Printf("Total transactions: %d\n\n", blocks.Count)
    
    for _, block := range blocks.List {
        fmt.Printf("Block #%d\n", block.Height)
        fmt.Printf("  Hash: %s\n", block.Hash.String())
        fmt.Printf("  Type: %s\n", getBlockType(block.BlockType))
        
        if block.BlockType == 2 || block.BlockType == 3 { // Send
            fmt.Printf("  To: %s\n", block.ToAddress.String())
            fmt.Printf("  Amount: %s %s\n", block.Amount.String(), block.TokenName)
        } else if block.BlockType == 4 || block.BlockType == 5 { // Receive
            fmt.Printf("  From Hash: %s\n", block.FromBlockHash.String())
        }
        
        if len(block.Data) > 0 {
            fmt.Printf("  Data: %x\n", block.Data)
        }
        
        fmt.Printf("  Timestamp: %s\n", time.Unix(block.Timestamp.Unix(), 0))
        fmt.Println()
    }
}

func getBlockType(blockType uint64) string {
    switch blockType {
    case 1:
        return "Genesis"
    case 2:
        return "Send"
    case 3:
        return "Contract Send"
    case 4:
        return "Receive"
    case 5:
        return "Contract Receive"
    default:
        return fmt.Sprintf("Unknown (%d)", blockType)
    }
}

func main() {
    z, _ := zenon.NewZenon("")
    z.Start("", "ws://127.0.0.1:35998", 0)
    defer z.Stop()
    
    // Example: Get transaction history
    address, _ := types.ParseAddress("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
    getTransactionHistory(z, address, 0, 10) // First 10 transactions
}
```

### Get Specific Transaction by Hash

```go
func getTransactionByHash(z *zenon.Zenon, hash types.Hash) {
    block, err := z.Client.LedgerApi.GetAccountBlockByHash(hash)
    if err != nil {
        log.Fatal("Failed to get block:", err)
    }
    
    if block == nil {
        fmt.Println("Transaction not found")
        return
    }
    
    fmt.Printf("Transaction Details:\n")
    fmt.Printf("  Hash: %s\n", block.Hash.String())
    fmt.Printf("  Height: %d\n", block.Height)
    fmt.Printf("  Address: %s\n", block.Address.String())
    fmt.Printf("  To Address: %s\n", block.ToAddress.String())
    fmt.Printf("  Amount: %s\n", block.Amount.String())
    fmt.Printf("  Token: %s\n", block.TokenName)
    fmt.Printf("  Confirmations: %d\n", block.Confirmations)
}
```

## Reading Momentum Blocks

### Get Latest Momentum

```go
package main

import (
    "fmt"
    "log"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

func main() {
    z, _ := zenon.NewZenon("")
    z.Start("", "ws://127.0.0.1:35998", 0)
    defer z.Stop()
    
    // Get frontier (latest) momentum
    momentum, err := z.Client.LedgerApi.GetFrontierMomentum()
    if err != nil {
        log.Fatal("Failed to get momentum:", err)
    }
    
    fmt.Printf("Latest Momentum:\n")
    fmt.Printf("  Height: %d\n", momentum.Height)
    fmt.Printf("  Hash: %s\n", momentum.Hash.String())
    fmt.Printf("  Previous Hash: %s\n", momentum.PreviousHash.String())
    fmt.Printf("  Timestamp: %d\n", momentum.Timestamp.Unix())
    fmt.Printf("  Producer: %s\n", momentum.Producer.String())
}
```

### Get Momentum Range

```go
func getMomentumRange(z *zenon.Zenon, startHeight, count uint64) {
    momentums, err := z.Client.LedgerApi.GetMomentumsByHeight(startHeight, count)
    if err != nil {
        log.Fatal("Failed to get momentums:", err)
    }
    
    fmt.Printf("Momentums from height %d to %d:\n", startHeight, startHeight+count-1)
    for _, m := range momentums.List {
        fmt.Printf("  Height %d: %s (Producer: %s)\n",
            m.Height,
            m.Hash.String(),
            m.Producer.String())
    }
}
```

### Get Detailed Momentum Information

```go
func getDetailedMomentum(z *zenon.Zenon, height uint64) {
    detailed, err := z.Client.LedgerApi.GetDetailedMomentumsByHeight(height, 1)
    if err != nil {
        log.Fatal("Failed to get detailed momentum:", err)
    }
    
    if len(detailed.List) == 0 {
        fmt.Println("No momentum found at height", height)
        return
    }
    
    m := detailed.List[0]
    fmt.Printf("Detailed Momentum at height %d:\n", height)
    fmt.Printf("  Hash: %s\n", m.Momentum.Hash.String())
    fmt.Printf("  Version: %d\n", m.Momentum.Version)
    fmt.Printf("  Chain Identifier: %d\n", m.Momentum.ChainIdentifier)
    fmt.Printf("  Producer: %s\n", m.Producer.String())
    fmt.Printf("  Account Blocks: %d\n", len(m.AccountBlocks))
    
    // Show account blocks in this momentum
    for _, ab := range m.AccountBlocks {
        fmt.Printf("    - %s: %s\n", ab.Address.String(), ab.Hash.String())
    }
}
```

## Reading Embedded Contract Data

### Pillar Information

```go
func getPillarInfo(z *zenon.Zenon) {
    // Get all pillars
    pillars, err := z.Client.PillarApi.GetAll(0, 100)
    if err != nil {
        log.Fatal("Failed to get pillars:", err)
    }
    
    fmt.Printf("Total Pillars: %d\n\n", pillars.Count)
    
    for _, pillar := range pillars.List {
        fmt.Printf("Pillar: %s\n", pillar.Name)
        fmt.Printf("  Owner: %s\n", pillar.OwnerAddress.String())
        fmt.Printf("  Producer: %s\n", pillar.ProducerAddress.String())
        fmt.Printf("  Weight: %s\n", pillar.Weight.String())
        fmt.Printf("  Produced Momentums: %d\n", pillar.ProducedMomentums)
        fmt.Printf("  Expected Momentums: %d\n", pillar.ExpectedMomentums)
        fmt.Printf("  Give Block %d%%, Delegate %d%%\n",
            pillar.GiveBlockRewardPercentage,
            pillar.GiveDelegateRewardPercentage)
        fmt.Println()
    }
}
```

### Token Information

```go
func getTokenInfo(z *zenon.Zenon) {
    // Get all tokens
    tokens, err := z.Client.TokenApi.GetAll(0, 50)
    if err != nil {
        log.Fatal("Failed to get tokens:", err)
    }
    
    fmt.Printf("Total Tokens: %d\n\n", tokens.Count)
    
    for _, token := range tokens.List {
        fmt.Printf("Token: %s (%s)\n", token.Name, token.Symbol)
        fmt.Printf("  Standard: %s\n", token.TokenStandard.String())
        fmt.Printf("  Owner: %s\n", token.Owner.String())
        fmt.Printf("  Decimals: %d\n", token.Decimals)
        fmt.Printf("  Total Supply: %s\n", token.TotalSupply.String())
        fmt.Printf("  Max Supply: %s\n", token.MaxSupply.String())
        fmt.Printf("  Mintable: %t, Burnable: %t\n", token.IsMintable, token.IsBurnable)
        fmt.Println()
    }
}

// Get specific token
func getSpecificToken(z *zenon.Zenon, tokenStandard types.ZenonTokenStandard) {
    token, err := z.Client.TokenApi.GetByZts(tokenStandard)
    if err != nil {
        log.Fatal("Failed to get token:", err)
    }
    
    fmt.Printf("Token Details for %s:\n", tokenStandard.String())
    fmt.Printf("  Name: %s\n", token.Name)
    fmt.Printf("  Symbol: %s\n", token.Symbol)
    fmt.Printf("  Domain: %s\n", token.Domain)
    fmt.Printf("  Total Supply: %s\n", token.TotalSupply.String())
}
```

### Plasma Information

```go
func getPlasmaInfo(z *zenon.Zenon, address types.Address) {
    // Get plasma info for address
    plasmaInfo, err := z.Client.PlasmaApi.Get(address)
    if err != nil {
        log.Fatal("Failed to get plasma info:", err)
    }
    
    fmt.Printf("Plasma Info for %s:\n", address.String())
    fmt.Printf("  Current Plasma: %d\n", plasmaInfo.CurrentPlasma)
    fmt.Printf("  Max Plasma: %d\n", plasmaInfo.MaxPlasma)
    fmt.Printf("  QSR Staked: %s\n", plasmaInfo.QsrAmount.String())
    
    // Get fusion entries
    fusions, err := z.Client.PlasmaApi.GetFusionsByAddress(address, 0, 10)
    if err == nil {
        fmt.Printf("\nFusion History:\n")
        for _, fusion := range fusions.List {
            fmt.Printf("  Beneficiary: %s, Amount: %s QSR\n",
                fusion.Beneficiary.String(),
                fusion.QsrAmount.String())
        }
    }
}
```

## Network Statistics

```go
func getNetworkStats(z *zenon.Zenon) {
    // Network info
    netInfo, err := z.Client.StatsApi.NetworkInfo()
    if err != nil {
        log.Fatal("Failed to get network info:", err)
    }
    
    fmt.Printf("Network Statistics:\n")
    fmt.Printf("  Peers: %d\n", netInfo.NumPeers)
    fmt.Printf("  Public Key: %s\n", netInfo.Self.PublicKey)
    
    // Sync info
    syncInfo, err := z.Client.StatsApi.SyncInfo()
    if err == nil {
        fmt.Printf("\nSync Status:\n")
        fmt.Printf("  Current Height: %d\n", syncInfo.CurrentHeight)
        fmt.Printf("  Target Height: %d\n", syncInfo.TargetHeight)
        
        if syncInfo.CurrentHeight == syncInfo.TargetHeight {
            fmt.Println("  Status: Fully Synced ✓")
        } else {
            progress := float64(syncInfo.CurrentHeight) / float64(syncInfo.TargetHeight) * 100
            fmt.Printf("  Status: Syncing... %.2f%%\n", progress)
        }
    }
    
    // Process info
    processInfo, err := z.Client.StatsApi.ProcessInfo()
    if err == nil {
        fmt.Printf("\nNode Process Info:\n")
        fmt.Printf("  Version: %s\n", processInfo.Version)
        fmt.Printf("  Commit: %s\n", processInfo.Commit)
    }
}
```

## Advanced Query Patterns

### Pagination Helper

```go
type PaginationHelper struct {
    PageSize uint32
}

func NewPaginationHelper(pageSize uint32) *PaginationHelper {
    return &PaginationHelper{PageSize: pageSize}
}

func (p *PaginationHelper) GetAllAccountBlocks(z *zenon.Zenon, address types.Address) ([]*api.AccountBlock, error) {
    var allBlocks []*api.AccountBlock
    pageIndex := uint32(0)
    
    for {
        blocks, err := z.Client.LedgerApi.GetAccountBlocksByHeight(address, pageIndex, p.PageSize)
        if err != nil {
            return nil, err
        }
        
        allBlocks = append(allBlocks, blocks.List...)
        
        if uint32(len(blocks.List)) < p.PageSize {
            break // No more pages
        }
        
        pageIndex++
    }
    
    return allBlocks, nil
}
```

### Caching Layer

```go
import (
    "sync"
    "time"
)

type CachedReader struct {
    client    *zenon.Zenon
    cache     map[string]interface{}
    cacheLock sync.RWMutex
    ttl       time.Duration
}

func NewCachedReader(client *zenon.Zenon, ttl time.Duration) *CachedReader {
    return &CachedReader{
        client: client,
        cache:  make(map[string]interface{}),
        ttl:    ttl,
    }
}

func (cr *CachedReader) GetAccountInfo(address types.Address) (*api.AccountInfo, error) {
    key := "account:" + address.String()
    
    // Check cache
    cr.cacheLock.RLock()
    if cached, exists := cr.cache[key]; exists {
        cr.cacheLock.RUnlock()
        return cached.(*api.AccountInfo), nil
    }
    cr.cacheLock.RUnlock()
    
    // Fetch from chain
    info, err := cr.client.Client.LedgerApi.GetAccountInfoByAddress(address)
    if err != nil {
        return nil, err
    }
    
    // Update cache
    cr.cacheLock.Lock()
    cr.cache[key] = info
    cr.cacheLock.Unlock()
    
    // Auto-expire
    go func() {
        time.Sleep(cr.ttl)
        cr.cacheLock.Lock()
        delete(cr.cache, key)
        cr.cacheLock.Unlock()
    }()
    
    return info, nil
}
```

## Real-time Data Monitoring

```go
func monitorAccount(z *zenon.Zenon, address types.Address) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    var lastHeight uint64
    
    for range ticker.C {
        info, err := z.Client.LedgerApi.GetAccountInfoByAddress(address)
        if err != nil {
            log.Printf("Error checking account: %v", err)
            continue
        }
        
        if info.AccountHeight > lastHeight {
            fmt.Printf("New activity on %s! Height: %d -> %d\n",
                address.String(),
                lastHeight,
                info.AccountHeight)
            
            // Get the new blocks
            blocks, _ := z.Client.LedgerApi.GetAccountBlocksByHeight(
                address,
                0,
                uint32(info.AccountHeight-lastHeight))
            
            for _, block := range blocks.List {
                fmt.Printf("  New block: %s\n", block.Hash.String())
            }
            
            lastHeight = info.AccountHeight
        }
    }
}
```

## Exercise

Build a blockchain explorer that:
1. Displays account information and balances
2. Shows transaction history with pagination
3. Monitors momentum production
4. Tracks token statistics
5. Provides real-time updates

## Summary

You've learned:
- ✅ Reading account information and balances
- ✅ Querying transaction history
- ✅ Accessing momentum blocks
- ✅ Reading embedded contract states
- ✅ Implementing pagination and caching
- ✅ Real-time monitoring patterns

Next: [04-sending-transactions.md](./04-sending-transactions.md)