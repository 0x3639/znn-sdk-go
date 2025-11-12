# Tutorial 04: Sending Transactions

Learn how to send various types of transactions on the Zenon Network, including simple transfers, contract calls, and handling transaction receipts.

## Understanding Zenon Transactions

### Transaction Types
1. **Send Blocks**: Transfer tokens to another address
2. **Receive Blocks**: Accept incoming transfers
3. **Contract Calls**: Interact with embedded contracts

### Transaction Flow
```
Create TX → Auto-fill Parameters → Calculate PoW/Plasma → Sign → Broadcast → Confirm
```

## Basic Token Transfer

### Simple ZNN Transfer

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
    // Initialize with wallet
    z, err := zenon.NewZenon("my-wallet")
    if err != nil {
        log.Fatal(err)
    }
    
    // Connect and unlock wallet
    err = z.Start("password", "ws://127.0.0.1:35998", 0)
    if err != nil {
        log.Fatal(err)
    }
    defer z.Stop()
    
    // Recipient address
    toAddress, _ := types.ParseAddress("z1qr4pexnnfaexqqz8nscjjcsajy5hdqfkgadvwx")
    
    // Amount: 1 ZNN (8 decimals)
    amount := big.NewInt(100000000) // 1 * 10^8
    
    // Create and send transaction
    template := z.Client.LedgerApi.SendTemplate(
        toAddress,
        types.ZnnTokenStandard,
        amount,
        []byte{}, // Optional data
    )
    
    err = z.Send(template)
    if err != nil {
        log.Fatal("Failed to send transaction:", err)
    }
    
    fmt.Println("Transaction sent successfully!")
}
```

### Transfer with Data

```go
func sendWithData(z *zenon.Zenon, to types.Address, amount *big.Int, message string) error {
    // Convert message to bytes
    data := []byte(message)
    
    // Create transaction with data
    template := z.Client.LedgerApi.SendTemplate(
        to,
        types.ZnnTokenStandard,
        amount,
        data,
    )
    
    return z.Send(template)
}
```

## Sending Different Token Types

### Send QSR

```go
func sendQSR(z *zenon.Zenon, to types.Address, amount *big.Int) error {
    template := z.Client.LedgerApi.SendTemplate(
        to,
        types.QsrTokenStandard, // QSR token standard
        amount,
        []byte{},
    )
    
    return z.Send(template)
}
```

### Send Custom ZTS Token

```go
func sendZTSToken(z *zenon.Zenon, to types.Address, tokenStandard types.ZenonTokenStandard, amount *big.Int) error {
    // First, verify token exists
    token, err := z.Client.TokenApi.GetByZts(tokenStandard)
    if err != nil {
        return fmt.Errorf("token not found: %v", err)
    }
    
    fmt.Printf("Sending %s %s\n", amount.String(), token.Symbol)
    
    template := z.Client.LedgerApi.SendTemplate(
        to,
        tokenStandard,
        amount,
        []byte{},
    )
    
    return z.Send(template)
}

// Usage
tokenZTS, _ := types.ParseZTS("zts1qsrxxxxxxxxxxxxxmrhjll")
amount := big.NewInt(1000000) // Adjust for token decimals
sendZTSToken(z, toAddress, tokenZTS, amount)
```

## Receiving Transactions

### Auto-receive Pattern

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
    "github.com/zenon-network/go-zenon/common/types"
)

type AutoReceiver struct {
    client *zenon.Zenon
}

func NewAutoReceiver(walletName, password, nodeURL string) (*AutoReceiver, error) {
    z, err := zenon.NewZenon(walletName)
    if err != nil {
        return nil, err
    }
    
    err = z.Start(password, nodeURL, 0)
    if err != nil {
        return nil, err
    }
    
    return &AutoReceiver{client: z}, nil
}

func (ar *AutoReceiver) Start() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        ar.checkAndReceive()
    }
}

func (ar *AutoReceiver) checkAndReceive() {
    // Get unreceived blocks
    unreceived, err := ar.client.Client.LedgerApi.GetUnreceivedBlocksByAddress(
        ar.client.Address(),
        0, 50, // Page 0, up to 50 blocks
    )
    
    if err != nil {
        log.Printf("Error checking unreceived: %v", err)
        return
    }
    
    if unreceived.Count == 0 {
        return
    }
    
    fmt.Printf("Found %d unreceived transactions\n", unreceived.Count)
    
    // Receive each block
    for _, block := range unreceived.List {
        fmt.Printf("Receiving %s %s from %s\n", 
            block.Amount.String(),
            block.TokenName,
            block.Address.String())
        
        template := ar.client.Client.LedgerApi.ReceiveTemplate(block.Hash)
        
        err := ar.client.Send(template)
        if err != nil {
            log.Printf("Failed to receive %s: %v", block.Hash.String(), err)
            continue
        }
        
        fmt.Printf("Received: %s\n", block.Hash.String())
        
        // Small delay between receives
        time.Sleep(1 * time.Second)
    }
}

func main() {
    receiver, err := NewAutoReceiver("my-wallet", "password", "ws://127.0.0.1:35998")
    if err != nil {
        log.Fatal(err)
    }
    defer receiver.client.Stop()
    
    receiver.Start()
}
```

### Manual Receive

```go
func receiveSpecificBlock(z *zenon.Zenon, fromBlockHash types.Hash) error {
    // Create receive template
    template := z.Client.LedgerApi.ReceiveTemplate(fromBlockHash)
    
    // Send receive transaction
    return z.Send(template)
}
```

## Transaction Confirmation

### Wait for Confirmation

```go
func waitForConfirmation(z *zenon.Zenon, txHash types.Hash, maxWait time.Duration) error {
    deadline := time.Now().Add(maxWait)
    
    for time.Now().Before(deadline) {
        block, err := z.Client.LedgerApi.GetAccountBlockByHash(txHash)
        if err != nil {
            return err
        }
        
        if block != nil && block.Confirmations > 0 {
            fmt.Printf("Transaction confirmed! Confirmations: %d\n", block.Confirmations)
            return nil
        }
        
        time.Sleep(2 * time.Second)
    }
    
    return fmt.Errorf("transaction not confirmed within %v", maxWait)
}
```

### Track Transaction Status

```go
type TransactionTracker struct {
    client *zenon.Zenon
    pending map[types.Hash]time.Time
}

func NewTransactionTracker(client *zenon.Zenon) *TransactionTracker {
    return &TransactionTracker{
        client: client,
        pending: make(map[types.Hash]time.Time),
    }
}

func (tt *TransactionTracker) SendAndTrack(template *nom.AccountBlock) (types.Hash, error) {
    // Store original hash calculation method
    template.Hash = template.ComputeHash()
    hash := template.Hash
    
    // Send transaction
    err := tt.client.Send(template)
    if err != nil {
        return types.ZeroHash, err
    }
    
    // Track it
    tt.pending[hash] = time.Now()
    
    // Start monitoring
    go tt.monitor(hash)
    
    return hash, nil
}

func (tt *TransactionTracker) monitor(hash types.Hash) {
    for {
        block, err := tt.client.Client.LedgerApi.GetAccountBlockByHash(hash)
        if err != nil {
            log.Printf("Error monitoring %s: %v", hash.String(), err)
            time.Sleep(5 * time.Second)
            continue
        }
        
        if block != nil && block.Confirmations > 0 {
            delete(tt.pending, hash)
            fmt.Printf("✓ Transaction %s confirmed\n", hash.String())
            return
        }
        
        time.Sleep(2 * time.Second)
    }
}
```

## Batch Transactions

### Sequential Batch Send

```go
func batchSend(z *zenon.Zenon, transfers []Transfer) error {
    for i, transfer := range transfers {
        fmt.Printf("Sending transfer %d/%d\n", i+1, len(transfers))
        
        template := z.Client.LedgerApi.SendTemplate(
            transfer.To,
            transfer.Token,
            transfer.Amount,
            []byte{},
        )
        
        err := z.Send(template)
        if err != nil {
            return fmt.Errorf("failed at transfer %d: %v", i+1, err)
        }
        
        // Wait between transactions to avoid nonce issues
        time.Sleep(1 * time.Second)
    }
    
    return nil
}

type Transfer struct {
    To     types.Address
    Token  types.ZenonTokenStandard
    Amount *big.Int
}
```

### Parallel Transactions from Multiple Addresses

```go
func parallelSend(wallets []*zenon.Zenon, transfers []Transfer) {
    var wg sync.WaitGroup
    
    for i, wallet := range wallets {
        if i >= len(transfers) {
            break
        }
        
        wg.Add(1)
        go func(w *zenon.Zenon, t Transfer) {
            defer wg.Done()
            
            template := w.Client.LedgerApi.SendTemplate(
                t.To,
                t.Token,
                t.Amount,
                []byte{},
            )
            
            err := w.Send(template)
            if err != nil {
                log.Printf("Failed: %v", err)
            } else {
                fmt.Printf("Sent from %s\n", w.Address().String())
            }
        }(wallet, transfers[i])
    }
    
    wg.Wait()
}
```

## Error Handling

### Common Transaction Errors

```go
func handleTransactionError(err error) {
    if err == nil {
        return
    }
    
    switch {
    case strings.Contains(err.Error(), "insufficient balance"):
        fmt.Println("Error: Not enough funds")
        
    case strings.Contains(err.Error(), "insufficient plasma"):
        fmt.Println("Error: Need more Plasma (fuse QSR or wait)")
        
    case strings.Contains(err.Error(), "invalid address"):
        fmt.Println("Error: Invalid recipient address")
        
    case strings.Contains(err.Error(), "nonce"):
        fmt.Println("Error: Nonce issue - wait and retry")
        
    default:
        fmt.Printf("Transaction error: %v\n", err)
    }
}
```

### Retry Logic

```go
func sendWithRetry(z *zenon.Zenon, template *nom.AccountBlock, maxRetries int) error {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        err := z.Send(template)
        if err == nil {
            return nil // Success
        }
        
        lastErr = err
        
        // Check if error is retryable
        if strings.Contains(err.Error(), "nonce") {
            time.Sleep(time.Duration(i+1) * 2 * time.Second)
            continue
        }
        
        // Non-retryable error
        return err
    }
    
    return fmt.Errorf("failed after %d retries: %v", maxRetries, lastErr)
}
```

## Gas-Free Transactions with Plasma

```go
func checkPlasmaBeforeSend(z *zenon.Zenon, to types.Address, amount *big.Int) error {
    // Check Plasma availability
    plasmaInfo, err := z.Client.PlasmaApi.Get(z.Address())
    if err != nil {
        return err
    }
    
    fmt.Printf("Current Plasma: %d/%d\n", plasmaInfo.CurrentPlasma, plasmaInfo.MaxPlasma)
    
    // Estimate required Plasma
    template := z.Client.LedgerApi.SendTemplate(to, types.ZnnTokenStandard, amount, []byte{})
    
    required, err := z.Client.PlasmaApi.GetRequiredPoWForAccountBlock(embedded.GetRequiredParam{
        SelfAddr:  z.Address(),
        BlockType: template.BlockType,
        ToAddr:    &template.ToAddress,
        Data:      template.Data,
    })
    
    if err != nil {
        return err
    }
    
    if required.RequiredDifficulty > 0 {
        fmt.Printf("Warning: Insufficient Plasma. PoW will be generated.\n")
        fmt.Printf("Required difficulty: %d\n", required.RequiredDifficulty)
    } else {
        fmt.Printf("Sufficient Plasma available for feeless transaction\n")
    }
    
    // Send transaction (SDK handles PoW automatically)
    return z.Send(template)
}
```

## Complete Transaction Example

```go
package main

import (
    "fmt"
    "log"
    "math/big"
    "time"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
    "github.com/zenon-network/go-zenon/common/types"
)

type TransactionManager struct {
    client *zenon.Zenon
}

func NewTransactionManager(wallet, password, nodeURL string) (*TransactionManager, error) {
    z, err := zenon.NewZenon(wallet)
    if err != nil {
        return nil, err
    }
    
    err = z.Start(password, nodeURL, 0)
    if err != nil {
        return nil, err
    }
    
    return &TransactionManager{client: z}, nil
}

func (tm *TransactionManager) SendTransaction(to types.Address, token types.ZenonTokenStandard, amount *big.Int) error {
    // Pre-flight checks
    balance, err := tm.getBalance(token)
    if err != nil {
        return err
    }
    
    if balance.Cmp(amount) < 0 {
        return fmt.Errorf("insufficient balance: have %s, need %s", balance.String(), amount.String())
    }
    
    // Create transaction
    fmt.Printf("Sending %s to %s\n", amount.String(), to.String())
    
    template := tm.client.Client.LedgerApi.SendTemplate(to, token, amount, []byte{})
    
    // Record pre-send height
    preHeight := tm.getAccountHeight()
    
    // Send transaction
    err = tm.client.Send(template)
    if err != nil {
        return fmt.Errorf("send failed: %v", err)
    }
    
    // Wait for new block
    newHash, err := tm.waitForNewBlock(preHeight, 30*time.Second)
    if err != nil {
        return err
    }
    
    fmt.Printf("Transaction successful! Hash: %s\n", newHash.String())
    
    // Wait for confirmation
    return tm.waitForConfirmation(newHash, 60*time.Second)
}

func (tm *TransactionManager) getBalance(token types.ZenonTokenStandard) (*big.Int, error) {
    info, err := tm.client.Client.LedgerApi.GetAccountInfoByAddress(tm.client.Address())
    if err != nil {
        return nil, err
    }
    
    for _, balance := range info.BalanceInfoList {
        if balance.Token.TokenStandard == token {
            return balance.Balance, nil
        }
    }
    
    return big.NewInt(0), nil
}

func (tm *TransactionManager) getAccountHeight() uint64 {
    info, err := tm.client.Client.LedgerApi.GetAccountInfoByAddress(tm.client.Address())
    if err != nil {
        return 0
    }
    return info.AccountHeight
}

func (tm *TransactionManager) waitForNewBlock(oldHeight uint64, timeout time.Duration) (types.Hash, error) {
    deadline := time.Now().Add(timeout)
    
    for time.Now().Before(deadline) {
        newHeight := tm.getAccountHeight()
        if newHeight > oldHeight {
            // Get the new block
            blocks, err := tm.client.Client.LedgerApi.GetAccountBlocksByHeight(
                tm.client.Address(),
                0, 1,
            )
            if err == nil && len(blocks.List) > 0 {
                return blocks.List[0].Hash, nil
            }
        }
        
        time.Sleep(1 * time.Second)
    }
    
    return types.ZeroHash, fmt.Errorf("no new block after %v", timeout)
}

func (tm *TransactionManager) waitForConfirmation(hash types.Hash, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    
    for time.Now().Before(deadline) {
        block, err := tm.client.Client.LedgerApi.GetAccountBlockByHash(hash)
        if err != nil {
            return err
        }
        
        if block != nil && block.Confirmations > 0 {
            fmt.Printf("✓ Confirmed with %d confirmations\n", block.Confirmations)
            return nil
        }
        
        time.Sleep(2 * time.Second)
    }
    
    return fmt.Errorf("not confirmed after %v", timeout)
}

func (tm *TransactionManager) Close() error {
    return tm.client.Stop()
}

func main() {
    tm, err := NewTransactionManager("my-wallet", "password", "ws://127.0.0.1:35998")
    if err != nil {
        log.Fatal(err)
    }
    defer tm.Close()
    
    // Send 0.1 ZNN
    to, _ := types.ParseAddress("z1qr4pexnnfaexqqz8nscjjcsajy5hdqfkgadvwx")
    amount := big.NewInt(10000000) // 0.1 ZNN
    
    err = tm.SendTransaction(to, types.ZnnTokenStandard, amount)
    if err != nil {
        log.Fatal("Transaction failed:", err)
    }
}
```

## Testing Transactions

```go
func TestSendTransaction(t *testing.T) {
    // Use testnet
    z, _ := zenon.NewZenon("test-wallet")
    z.Start("password", "ws://testnet.zenon.network:35998", 0)
    defer z.Stop()
    
    // Test small amount
    testAmount := big.NewInt(1) // Smallest unit
    testAddress, _ := types.ParseAddress("z1qr4pexnnfaexqqz8nscjjcsajy5hdqfkgadvwx")
    
    template := z.Client.LedgerApi.SendTemplate(
        testAddress,
        types.ZnnTokenStandard,
        testAmount,
        []byte("test"),
    )
    
    err := z.Send(template)
    if err != nil {
        t.Fatalf("Failed to send: %v", err)
    }
    
    t.Log("Transaction sent successfully")
}
```

## Summary

You've learned:
- ✅ Sending basic token transfers
- ✅ Receiving transactions automatically
- ✅ Handling different token types
- ✅ Transaction confirmation tracking
- ✅ Batch and parallel transactions
- ✅ Error handling and retry logic
- ✅ Plasma and PoW management

Next: [05-working-with-tokens.md](./05-working-with-tokens.md)