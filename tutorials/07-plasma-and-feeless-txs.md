# Tutorial 07: Plasma and Feeless Transactions

Master the Plasma system in Zenon Network to enable feeless, high-throughput transactions through QSR fusion and efficient resource management.

## Understanding Plasma

### What is Plasma?
- A third-dimensional asset that acts as computational fuel
- Created by fusing QSR tokens or generating Proof-of-Work
- Enables truly feeless transactions
- Regenerates over time when QSR is fused

### Plasma Mechanics
```
Transaction without Plasma → Requires PoW calculation
Transaction with Plasma → Instant and feeless
QSR fusion → Creates Plasma for beneficiary
Time → Plasma regenerates naturally
```

## Checking Plasma Status

### Get Plasma Information

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

func checkPlasma(z *zenon.Zenon, address types.Address) {
    plasmaInfo, err := z.Client.PlasmaApi.Get(address)
    if err != nil {
        log.Fatal("Failed to get plasma info:", err)
    }
    
    fmt.Printf("Plasma Status for %s:\n", address.String())
    fmt.Printf("  Current Plasma: %d\n", plasmaInfo.CurrentPlasma)
    fmt.Printf("  Max Plasma: %d\n", plasmaInfo.MaxPlasma)
    fmt.Printf("  QSR Amount: %s\n", plasmaInfo.QsrAmount.String())
    
    // Calculate usage percentage
    if plasmaInfo.MaxPlasma > 0 {
        percentage := float64(plasmaInfo.CurrentPlasma) / float64(plasmaInfo.MaxPlasma) * 100
        fmt.Printf("  Usage: %.2f%%\n", percentage)
    }
    
    // Show status
    status := getPlasmaStatus(plasmaInfo)
    fmt.Printf("  Status: %s\n", status)
}

func getPlasmaStatus(info *api.PlasmaInfo) string {
    if info.QsrAmount.Cmp(big.NewInt(0)) == 0 {
        return "No QSR fused"
    }
    
    percentage := float64(info.CurrentPlasma) / float64(info.MaxPlasma) * 100
    
    switch {
    case percentage >= 80:
        return "Excellent (High throughput available)"
    case percentage >= 50:
        return "Good (Moderate throughput)"
    case percentage >= 20:
        return "Low (Limited throughput)"
    default:
        return "Critical (PoW will be required)"
    }
}

func main() {
    z, _ := zenon.NewZenon("")
    z.Start("", "ws://127.0.0.1:35998", 0)
    defer z.Stop()
    
    // Check your own address
    address, _ := types.ParseAddress("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
    checkPlasma(z, address)
}
```

### Monitor Plasma Regeneration

```go
func monitorPlasmaRegen(z *zenon.Zenon, address types.Address) {
    fmt.Printf("Monitoring Plasma regeneration for %s\n", address.String())
    
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    lastPlasma := uint64(0)
    
    for range ticker.C {
        info, err := z.Client.PlasmaApi.Get(address)
        if err != nil {
            log.Printf("Error checking plasma: %v", err)
            continue
        }
        
        if info.CurrentPlasma != lastPlasma {
            change := int64(info.CurrentPlasma) - int64(lastPlasma)
            fmt.Printf("[%s] Plasma: %d/%d (%+d)\n",
                time.Now().Format("15:04:05"),
                info.CurrentPlasma,
                info.MaxPlasma,
                change)
            
            lastPlasma = info.CurrentPlasma
        }
    }
}
```

## Fusing QSR for Plasma

### Simple Fusion

```go
func fuseQSRForPlasma(z *zenon.Zenon, beneficiary types.Address, qsrAmount *big.Int) error {
    fmt.Printf("Fusing %s QSR for %s\n", qsrAmount.String(), beneficiary.String())
    
    // Check QSR balance first
    balance, err := getQSRBalance(z, z.Address())
    if err != nil {
        return err
    }
    
    if balance.Cmp(qsrAmount) < 0 {
        return fmt.Errorf("insufficient QSR: have %s, need %s",
            balance.String(), qsrAmount.String())
    }
    
    template := z.Client.PlasmaApi.Fuse(beneficiary, qsrAmount)
    
    err = z.Send(template)
    if err != nil {
        return fmt.Errorf("fusion failed: %v", err)
    }
    
    fmt.Println("QSR fusion successful!")
    return nil
}

func getQSRBalance(z *zenon.Zenon, address types.Address) (*big.Int, error) {
    info, err := z.Client.LedgerApi.GetAccountInfoByAddress(address)
    if err != nil {
        return nil, err
    }
    
    for _, balance := range info.BalanceInfoList {
        if balance.Token.TokenStandard == types.QsrTokenStandard {
            return balance.Balance, nil
        }
    }
    
    return big.NewInt(0), nil
}
```

### Optimal Fusion Calculator

```go
package main

import (
    "fmt"
    "math/big"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

type PlasmaManager struct {
    client *zenon.Zenon
}

func NewPlasmaManager(wallet, password, nodeURL string) (*PlasmaManager, error) {
    z, err := zenon.NewZenon(wallet)
    if err != nil {
        return nil, err
    }
    
    err = z.Start(password, nodeURL, 0)
    if err != nil {
        return nil, err
    }
    
    return &PlasmaManager{client: z}, nil
}

func (pm *PlasmaManager) CalculateOptimalFusion(address types.Address, targetTxPerHour int) (*big.Int, error) {
    // Get current plasma info
    info, err := pm.client.Client.PlasmaApi.Get(address)
    if err != nil {
        return nil, err
    }
    
    // Estimate plasma needed per transaction (approximate)
    plasmaPerTx := uint64(21000) // Base transaction cost
    
    // Calculate plasma needed per hour
    plasmaNeededPerHour := uint64(targetTxPerHour) * plasmaPerTx
    
    // Account for regeneration (approximate 1000 plasma per 10 seconds)
    regenPerHour := uint64(360000) // 360 units per hour
    
    // Net plasma needed
    netNeeded := plasmaNeededPerHour
    if plasmaNeededPerHour > regenPerHour {
        netNeeded = plasmaNeededPerHour - regenPerHour
    } else {
        netNeeded = 0
    }
    
    // Current available plasma
    availablePlasma := info.MaxPlasma
    
    if availablePlasma >= netNeeded {
        fmt.Printf("Current plasma sufficient for %d tx/hour\n", targetTxPerHour)
        return big.NewInt(0), nil
    }
    
    // Calculate additional QSR needed
    additionalPlasmaNeeded := netNeeded - availablePlasma
    
    // Approximate: 1 QSR ≈ 1000 plasma units (varies by network state)
    qsrNeeded := new(big.Int).SetUint64(additionalPlasmaNeeded)
    qsrNeeded.Mul(qsrNeeded, big.NewInt(1e8)) // Convert to QSR units
    qsrNeeded.Div(qsrNeeded, big.NewInt(1000)) // Plasma ratio
    
    fmt.Printf("Recommended fusion: %s QSR for %d tx/hour\n", 
        qsrNeeded.String(), targetTxPerHour)
    
    return qsrNeeded, nil
}

func (pm *PlasmaManager) FuseForTarget(targetTxPerHour int) error {
    address := pm.client.Address()
    
    qsrNeeded, err := pm.CalculateOptimalFusion(address, targetTxPerHour)
    if err != nil {
        return err
    }
    
    if qsrNeeded.Cmp(big.NewInt(0)) == 0 {
        fmt.Println("No additional fusion needed")
        return nil
    }
    
    // Check QSR balance
    balance, err := getQSRBalance(pm.client, address)
    if err != nil {
        return err
    }
    
    if balance.Cmp(qsrNeeded) < 0 {
        return fmt.Errorf("insufficient QSR: need %s, have %s",
            qsrNeeded.String(), balance.String())
    }
    
    // Perform fusion
    return fuseQSRForPlasma(pm.client, address, qsrNeeded)
}
```

## Transaction Plasma Management

### Pre-transaction Plasma Check

```go
func checkPlasmaBeforeTransaction(z *zenon.Zenon, to types.Address, amount *big.Int, data []byte) error {
    // Get required PoW for the transaction
    params := embedded.GetRequiredParam{
        SelfAddr:  z.Address(),
        BlockType: nom.BlockTypeUserSend,
        ToAddr:    &to,
        Data:      data,
    }
    
    required, err := z.Client.PlasmaApi.GetRequiredPoWForAccountBlock(params)
    if err != nil {
        return err
    }
    
    fmt.Printf("Transaction Requirements:\n")
    fmt.Printf("  Available Plasma: %d\n", required.AvailablePlasma)
    fmt.Printf("  Base Plasma: %d\n", required.BasePlasma)
    fmt.Printf("  Required Difficulty: %d\n", required.RequiredDifficulty)
    
    if required.RequiredDifficulty > 0 {
        fmt.Printf("⚠️  PoW required (difficulty: %d)\n", required.RequiredDifficulty)
        fmt.Println("Consider fusing more QSR for feeless transactions")
        
        // Estimate PoW time
        estimatedTime := estimatePoWTime(required.RequiredDifficulty)
        fmt.Printf("  Estimated PoW time: %v\n", estimatedTime)
    } else {
        fmt.Println("✅ Sufficient Plasma for feeless transaction")
    }
    
    return nil
}

func estimatePoWTime(difficulty uint64) time.Duration {
    // Very rough estimation based on average system
    // Actual time depends on CPU performance
    baseTime := time.Millisecond * 100
    multiplier := time.Duration(difficulty / 1000000)
    
    if multiplier < 1 {
        multiplier = 1
    }
    
    return baseTime * multiplier
}
```

### Automatic Plasma Management

```go
type AutoPlasmaManager struct {
    client         *zenon.Zenon
    minPlasmaRatio float64
    targetFusion   *big.Int
}

func NewAutoPlasmaManager(client *zenon.Zenon) *AutoPlasmaManager {
    return &AutoPlasmaManager{
        client:         client,
        minPlasmaRatio: 0.2, // Maintain 20% plasma
        targetFusion:   big.NewInt(10 * 1e8), // 10 QSR default
    }
}

func (apm *AutoPlasmaManager) EnsurePlasma() error {
    info, err := apm.client.Client.PlasmaApi.Get(apm.client.Address())
    if err != nil {
        return err
    }
    
    if info.MaxPlasma == 0 {
        fmt.Println("No QSR fused, performing initial fusion...")
        return apm.performFusion(apm.targetFusion)
    }
    
    currentRatio := float64(info.CurrentPlasma) / float64(info.MaxPlasma)
    
    if currentRatio < apm.minPlasmaRatio {
        fmt.Printf("Plasma low (%.2f%%), auto-fusing QSR...\n", currentRatio*100)
        return apm.performFusion(apm.targetFusion)
    }
    
    return nil
}

func (apm *AutoPlasmaManager) performFusion(amount *big.Int) error {
    balance, err := getQSRBalance(apm.client, apm.client.Address())
    if err != nil {
        return err
    }
    
    if balance.Cmp(amount) < 0 {
        return fmt.Errorf("insufficient QSR for auto-fusion")
    }
    
    return fuseQSRForPlasma(apm.client, apm.client.Address(), amount)
}

func (apm *AutoPlasmaManager) SendWithAutoPlasma(to types.Address, token types.ZenonTokenStandard, amount *big.Int) error {
    // Ensure sufficient plasma
    if err := apm.EnsurePlasma(); err != nil {
        log.Printf("Auto-plasma failed: %v", err)
        // Continue anyway - transaction will use PoW
    }
    
    // Send transaction
    template := apm.client.Client.LedgerApi.SendTemplate(to, token, amount, []byte{})
    return apm.client.Send(template)
}
```

## Fusion Management

### List Fusion Entries

```go
func listFusions(z *zenon.Zenon, address types.Address) {
    fusions, err := z.Client.PlasmaApi.GetFusionsByAddress(address, 0, 50)
    if err != nil {
        log.Fatal("Failed to get fusions:", err)
    }
    
    if fusions.Count == 0 {
        fmt.Println("No fusion entries found")
        return
    }
    
    fmt.Printf("Fusion Entries: %d\n", fusions.Count)
    
    totalFused := big.NewInt(0)
    
    for _, fusion := range fusions.List {
        fmt.Printf("\nFusion Hash: %s\n", fusion.Hash.String())
        fmt.Printf("  Beneficiary: %s\n", fusion.Beneficiary.String())
        fmt.Printf("  Amount: %s QSR\n", fusion.QsrAmount.String())
        fmt.Printf("  Height: %d\n", fusion.Height)
        fmt.Printf("  Expiration: %d\n", fusion.ExpirationHeight)
        
        // Check if ready to cancel
        currentHeight := getCurrentMomentumHeight(z)
        if currentHeight >= fusion.ExpirationHeight {
            fmt.Println("  Status: Ready to cancel (QSR can be reclaimed)")
        } else {
            remaining := fusion.ExpirationHeight - currentHeight
            fmt.Printf("  Status: Active (%d momentums remaining)\n", remaining)
        }
        
        totalFused.Add(totalFused, fusion.QsrAmount)
    }
    
    fmt.Printf("\nTotal QSR Fused: %s\n", totalFused.String())
}

func getCurrentMomentumHeight(z *zenon.Zenon) uint64 {
    momentum, err := z.Client.LedgerApi.GetFrontierMomentum()
    if err != nil {
        return 0
    }
    return momentum.Height
}
```

### Cancel Fusion

```go
func cancelFusion(z *zenon.Zenon, fusionHash types.Hash) error {
    // Get fusion info
    fusions, err := z.Client.PlasmaApi.GetFusionsByAddress(z.Address(), 0, 100)
    if err != nil {
        return err
    }
    
    var targetFusion *embedded.FusionInfo
    for _, fusion := range fusions.List {
        if fusion.Hash == fusionHash {
            targetFusion = fusion
            break
        }
    }
    
    if targetFusion == nil {
        return fmt.Errorf("fusion not found")
    }
    
    // Check if cancellable
    currentHeight := getCurrentMomentumHeight(z)
    if currentHeight < targetFusion.ExpirationHeight {
        return fmt.Errorf("fusion not yet expired (height %d < %d)",
            currentHeight, targetFusion.ExpirationHeight)
    }
    
    fmt.Printf("Canceling fusion: %s QSR\n", targetFusion.QsrAmount.String())
    
    template := z.Client.PlasmaApi.Cancel(fusionHash)
    
    return z.Send(template)
}
```

## High-Throughput Transaction Patterns

### Burst Transaction Handler

```go
type BurstTransactionHandler struct {
    client        *zenon.Zenon
    plasmaManager *AutoPlasmaManager
    maxConcurrent int
}

func NewBurstTransactionHandler(client *zenon.Zenon, maxConcurrent int) *BurstTransactionHandler {
    return &BurstTransactionHandler{
        client:        client,
        plasmaManager: NewAutoPlasmaManager(client),
        maxConcurrent: maxConcurrent,
    }
}

func (bth *BurstTransactionHandler) SendBurst(transfers []Transfer) error {
    // Pre-check plasma for burst
    totalTxs := len(transfers)
    fmt.Printf("Preparing burst of %d transactions\n", totalTxs)
    
    // Ensure we have enough plasma for the entire burst
    info, err := bth.client.Client.PlasmaApi.Get(bth.client.Address())
    if err != nil {
        return err
    }
    
    estimatedPlasmaNeeded := uint64(totalTxs) * 21000 // Rough estimate
    
    if info.CurrentPlasma < estimatedPlasmaNeeded {
        fmt.Printf("Insufficient plasma for burst, fusing additional QSR...\n")
        additionalQSR := big.NewInt(int64(estimatedPlasmaNeeded / 1000 * 1e8))
        
        err = bth.plasmaManager.performFusion(additionalQSR)
        if err != nil {
            log.Printf("Could not fuse additional QSR: %v", err)
        }
        
        time.Sleep(5 * time.Second) // Wait for fusion to process
    }
    
    // Execute burst with concurrency control
    semaphore := make(chan struct{}, bth.maxConcurrent)
    var wg sync.WaitGroup
    
    for i, transfer := range transfers {
        wg.Add(1)
        go func(index int, t Transfer) {
            defer wg.Done()
            
            semaphore <- struct{}{} // Acquire
            defer func() { <-semaphore }() // Release
            
            template := bth.client.Client.LedgerApi.SendTemplate(
                t.To, t.Token, t.Amount, []byte{})
            
            err := bth.client.Send(template)
            if err != nil {
                log.Printf("TX %d failed: %v", index+1, err)
            } else {
                fmt.Printf("TX %d sent successfully\n", index+1)
            }
        }(i, transfer)
    }
    
    wg.Wait()
    fmt.Println("Burst transaction completed")
    
    return nil
}

type Transfer struct {
    To     types.Address
    Token  types.ZenonTokenStandard
    Amount *big.Int
}
```

### Plasma-Optimized Service

```go
type PlasmaOptimizedService struct {
    client    *zenon.Zenon
    scheduler *TransactionScheduler
}

type TransactionScheduler struct {
    queue          chan Transaction
    plasmaCheckInterval time.Duration
}

type Transaction struct {
    To     types.Address
    Token  types.ZenonTokenStandard
    Amount *big.Int
    Data   []byte
    Result chan error
}

func NewPlasmaOptimizedService(client *zenon.Zenon) *PlasmaOptimizedService {
    scheduler := &TransactionScheduler{
        queue:          make(chan Transaction, 100),
        plasmaCheckInterval: 5 * time.Second,
    }
    
    service := &PlasmaOptimizedService{
        client:    client,
        scheduler: scheduler,
    }
    
    go service.processQueue()
    
    return service
}

func (pos *PlasmaOptimizedService) QueueTransaction(to types.Address, token types.ZenonTokenStandard, amount *big.Int) error {
    tx := Transaction{
        To:     to,
        Token:  token,
        Amount: amount,
        Result: make(chan error, 1),
    }
    
    select {
    case pos.scheduler.queue <- tx:
        return <-tx.Result
    default:
        return fmt.Errorf("transaction queue full")
    }
}

func (pos *PlasmaOptimizedService) processQueue() {
    for tx := range pos.scheduler.queue {
        // Wait for sufficient plasma
        pos.waitForPlasma()
        
        // Execute transaction
        template := pos.client.Client.LedgerApi.SendTemplate(
            tx.To, tx.Token, tx.Amount, tx.Data)
        
        err := pos.client.Send(template)
        tx.Result <- err
        
        // Small delay to prevent overwhelming the network
        time.Sleep(100 * time.Millisecond)
    }
}

func (pos *PlasmaOptimizedService) waitForPlasma() {
    for {
        info, err := pos.client.Client.PlasmaApi.Get(pos.client.Address())
        if err != nil {
            time.Sleep(1 * time.Second)
            continue
        }
        
        // Wait until we have at least 50% plasma
        if info.MaxPlasma == 0 || float64(info.CurrentPlasma)/float64(info.MaxPlasma) < 0.5 {
            time.Sleep(pos.scheduler.plasmaCheckInterval)
            continue
        }
        
        break // Sufficient plasma available
    }
}
```

## Plasma Analytics

### Plasma Usage Tracker

```go
type PlasmaAnalytics struct {
    client         *zenon.Zenon
    usageHistory   []PlasmaSnapshot
    historyMutex   sync.Mutex
}

type PlasmaSnapshot struct {
    Timestamp     time.Time
    CurrentPlasma uint64
    MaxPlasma     uint64
    Usage         float64
}

func NewPlasmaAnalytics(client *zenon.Zenon) *PlasmaAnalytics {
    pa := &PlasmaAnalytics{
        client:       client,
        usageHistory: make([]PlasmaSnapshot, 0),
    }
    
    go pa.collectMetrics()
    
    return pa
}

func (pa *PlasmaAnalytics) collectMetrics() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        info, err := pa.client.Client.PlasmaApi.Get(pa.client.Address())
        if err != nil {
            continue
        }
        
        var usage float64
        if info.MaxPlasma > 0 {
            usage = float64(info.CurrentPlasma) / float64(info.MaxPlasma)
        }
        
        snapshot := PlasmaSnapshot{
            Timestamp:     time.Now(),
            CurrentPlasma: info.CurrentPlasma,
            MaxPlasma:     info.MaxPlasma,
            Usage:         usage,
        }
        
        pa.historyMutex.Lock()
        pa.usageHistory = append(pa.usageHistory, snapshot)
        
        // Keep only last 24 hours
        cutoff := time.Now().Add(-24 * time.Hour)
        for i, snap := range pa.usageHistory {
            if snap.Timestamp.After(cutoff) {
                pa.usageHistory = pa.usageHistory[i:]
                break
            }
        }
        
        pa.historyMutex.Unlock()
    }
}

func (pa *PlasmaAnalytics) GetUsageReport() {
    pa.historyMutex.Lock()
    defer pa.historyMutex.Unlock()
    
    if len(pa.usageHistory) == 0 {
        fmt.Println("No usage data available")
        return
    }
    
    latest := pa.usageHistory[len(pa.usageHistory)-1]
    
    fmt.Printf("Plasma Usage Report:\n")
    fmt.Printf("  Current: %d/%d (%.2f%%)\n",
        latest.CurrentPlasma, latest.MaxPlasma, latest.Usage*100)
    
    // Calculate averages
    totalUsage := 0.0
    minUsage := 1.0
    maxUsage := 0.0
    
    for _, snap := range pa.usageHistory {
        totalUsage += snap.Usage
        if snap.Usage < minUsage {
            minUsage = snap.Usage
        }
        if snap.Usage > maxUsage {
            maxUsage = snap.Usage
        }
    }
    
    avgUsage := totalUsage / float64(len(pa.usageHistory))
    
    fmt.Printf("  24h Average: %.2f%%\n", avgUsage*100)
    fmt.Printf("  24h Range: %.2f%% - %.2f%%\n", minUsage*100, maxUsage*100)
    
    // Recommendations
    if avgUsage < 0.3 {
        fmt.Println("  💡 Consider reducing QSR fusion for efficiency")
    } else if avgUsage > 0.8 {
        fmt.Println("  ⚠️  Consider fusing more QSR for better performance")
    }
}
```

## Summary

You've learned:
- ✅ Understanding Plasma mechanics and benefits
- ✅ Monitoring Plasma status and regeneration
- ✅ Optimal QSR fusion strategies
- ✅ Pre-transaction Plasma checks
- ✅ Automatic Plasma management
- ✅ High-throughput transaction handling
- ✅ Fusion entry management
- ✅ Plasma analytics and optimization

Next: [08-building-a-dapp.md](./08-building-a-dapp.md)