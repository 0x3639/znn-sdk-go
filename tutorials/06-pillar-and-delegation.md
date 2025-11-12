# Tutorial 06: Pillars and Delegation

Learn about Pillars (validators) in Zenon Network, how to register a Pillar, delegate to Pillars, and manage staking operations.

## Understanding Pillars

### What are Pillars?
- Validators that produce momentum blocks
- Require 15,000 ZNN locked + QSR burned
- Earn rewards from block production
- Can share rewards with delegators

### Pillar Economics
```
Registration Cost:
- Lock: 15,000 ZNN (returned when dismantled)
- Burn: 150,000 QSR (Legacy) or 160,000+ QSR (Regular)

Rewards:
- Block rewards for producing momentums
- Delegation rewards from delegators
- Can set reward sharing percentages
```

## Pillar Information

### Get All Pillars

```go
package main

import (
    "fmt"
    "log"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

func listPillars(z *zenon.Zenon) {
    pillars, err := z.Client.PillarApi.GetAll(0, 100)
    if err != nil {
        log.Fatal("Failed to get pillars:", err)
    }
    
    fmt.Printf("Total Pillars: %d\n\n", pillars.Count)
    
    for _, pillar := range pillars.List {
        fmt.Printf("Pillar: %s\n", pillar.Name)
        fmt.Printf("  Rank: %d\n", pillar.Rank)
        fmt.Printf("  Type: %d\n", pillar.Type) // 0=Legacy, 1=Regular
        fmt.Printf("  Owner: %s\n", pillar.OwnerAddress.String())
        fmt.Printf("  Producer: %s\n", pillar.ProducerAddress.String())
        fmt.Printf("  Reward Address: %s\n", pillar.RewardWithdrawAddress.String())
        fmt.Printf("  Weight: %s\n", pillar.Weight.String())
        fmt.Printf("  Produced: %d/%d momentums\n", 
            pillar.ProducedMomentums, 
            pillar.ExpectedMomentums)
        fmt.Printf("  Give Block Reward: %d%%\n", pillar.GiveBlockRewardPercentage)
        fmt.Printf("  Give Delegate Reward: %d%%\n", pillar.GiveDelegateRewardPercentage)
        fmt.Println()
    }
}
```

### Get Specific Pillar

```go
func getPillarByName(z *zenon.Zenon, name string) {
    pillars, err := z.Client.PillarApi.GetByName(name)
    if err != nil {
        log.Fatal("Failed to get pillar:", err)
    }
    
    if pillars == nil {
        fmt.Printf("Pillar '%s' not found\n", name)
        return
    }
    
    fmt.Printf("Pillar Details: %s\n", pillars.Name)
    fmt.Printf("  Status: %s\n", getPillarStatus(pillars))
    fmt.Printf("  Uptime: %.2f%%\n", calculateUptime(pillars))
    fmt.Printf("  Total Delegated: %s ZNN\n", pillars.Weight.String())
}

func getPillarStatus(pillar *embedded.PillarInfo) string {
    if pillar.IsRevocable {
        return "Revocable"
    }
    if pillar.ProducedMomentums == 0 {
        return "Inactive"
    }
    uptime := calculateUptime(pillar)
    if uptime > 90 {
        return "Active"
    }
    return "Low Performance"
}

func calculateUptime(pillar *embedded.PillarInfo) float64 {
    if pillar.ExpectedMomentums == 0 {
        return 0
    }
    return float64(pillar.ProducedMomentums) / float64(pillar.ExpectedMomentums) * 100
}
```

## Registering a Pillar

### Check Requirements

```go
package main

import (
    "fmt"
    "log"
    "math/big"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

func checkPillarRequirements(z *zenon.Zenon) error {
    // Required amounts
    requiredZNN := big.NewInt(15000 * 1e8)  // 15,000 ZNN
    requiredQSR := big.NewInt(150000 * 1e8) // 150,000 QSR for Legacy
    
    // Get account info
    info, err := z.Client.LedgerApi.GetAccountInfoByAddress(z.Address())
    if err != nil {
        return err
    }
    
    var znnBalance, qsrBalance *big.Int
    
    for _, balance := range info.BalanceInfoList {
        if balance.Token.TokenStandard == types.ZnnTokenStandard {
            znnBalance = balance.Balance
        } else if balance.Token.TokenStandard == types.QsrTokenStandard {
            qsrBalance = balance.Balance
        }
    }
    
    fmt.Println("Pillar Registration Requirements Check:")
    fmt.Printf("  ZNN: %s / %s required", znnBalance.String(), requiredZNN.String())
    if znnBalance.Cmp(requiredZNN) >= 0 {
        fmt.Println(" ✓")
    } else {
        fmt.Println(" ✗")
        return fmt.Errorf("insufficient ZNN")
    }
    
    fmt.Printf("  QSR: %s / %s required", qsrBalance.String(), requiredQSR.String())
    if qsrBalance.Cmp(requiredQSR) >= 0 {
        fmt.Println(" ✓")
    } else {
        fmt.Println(" ✗")
        return fmt.Errorf("insufficient QSR")
    }
    
    // Check QSR deposit
    depositedQsr, err := z.Client.PillarApi.GetDepositedQsr(z.Address())
    if err == nil {
        fmt.Printf("  Deposited QSR: %s\n", depositedQsr.String())
    }
    
    return nil
}
```

### Register Pillar

```go
func registerPillar(z *zenon.Zenon, name string) error {
    // Check if name is available
    existing, _ := z.Client.PillarApi.GetByName(name)
    if existing != nil {
        return fmt.Errorf("pillar name '%s' already taken", name)
    }
    
    // Check requirements
    if err := checkPillarRequirements(z); err != nil {
        return err
    }
    
    fmt.Printf("Registering Pillar: %s\n", name)
    
    // Pillar configuration
    producerAddress := z.Address()           // Block producer address
    rewardAddress := z.Address()             // Where to withdraw rewards
    giveBlockRewardPercentage := uint8(0)    // % of block rewards to share
    giveDelegateRewardPercentage := uint8(10) // % of delegate rewards to share
    
    template := z.Client.PillarApi.Register(
        name,
        producerAddress,
        rewardAddress,
        giveBlockRewardPercentage,
        giveDelegateRewardPercentage,
    )
    
    err := z.Send(template)
    if err != nil {
        return fmt.Errorf("failed to register pillar: %v", err)
    }
    
    fmt.Println("Pillar registration transaction sent!")
    fmt.Println("Note: 15,000 ZNN will be locked and QSR will be burned")
    
    return nil
}
```

### Update Pillar

```go
func updatePillar(z *zenon.Zenon, name string) error {
    // Get current pillar info
    pillar, err := z.Client.PillarApi.GetByName(name)
    if err != nil {
        return err
    }
    
    if pillar.OwnerAddress != z.Address() {
        return fmt.Errorf("not pillar owner")
    }
    
    // New configuration
    newProducerAddress := z.Address()
    newRewardAddress := z.Address()
    newGiveBlockReward := uint8(5)    // Share 5% of block rewards
    newGiveDelegateReward := uint8(15) // Share 15% of delegate rewards
    
    template := z.Client.PillarApi.UpdatePillar(
        name,
        newProducerAddress,
        newRewardAddress,
        newGiveBlockReward,
        newGiveDelegateReward,
    )
    
    return z.Send(template)
}
```

## Delegation

### Delegate to a Pillar

```go
func delegateToPillar(z *zenon.Zenon, pillarName string) error {
    // Check if pillar exists
    pillar, err := z.Client.PillarApi.GetByName(pillarName)
    if err != nil {
        return fmt.Errorf("pillar not found: %v", err)
    }
    
    fmt.Printf("Delegating to Pillar: %s\n", pillarName)
    fmt.Printf("  Pillar gives %d%% delegate rewards\n", 
        pillar.GiveDelegateRewardPercentage)
    
    template := z.Client.PillarApi.Delegate(pillarName)
    
    err = z.Send(template)
    if err != nil {
        return fmt.Errorf("delegation failed: %v", err)
    }
    
    fmt.Println("Delegation successful!")
    return nil
}
```

### Undelegate

```go
func undelegate(z *zenon.Zenon) error {
    // Check current delegation
    delegationInfo, err := z.Client.PillarApi.GetDelegatedPillar(z.Address())
    if err != nil {
        return err
    }
    
    if delegationInfo == nil {
        fmt.Println("Not currently delegating")
        return nil
    }
    
    fmt.Printf("Undelegating from: %s\n", delegationInfo.Name)
    
    template := z.Client.PillarApi.Undelegate()
    
    return z.Send(template)
}
```

### Check Delegation Status

```go
func checkDelegation(z *zenon.Zenon, address types.Address) {
    delegationInfo, err := z.Client.PillarApi.GetDelegatedPillar(address)
    if err != nil {
        log.Printf("Error checking delegation: %v", err)
        return
    }
    
    if delegationInfo == nil {
        fmt.Printf("%s is not delegating\n", address.String())
        return
    }
    
    fmt.Printf("Delegation Info for %s:\n", address.String())
    fmt.Printf("  Delegated to: %s\n", delegationInfo.Name)
    fmt.Printf("  Weight: %s\n", delegationInfo.Weight.String())
    
    // Get balance to show delegation weight
    info, err := z.Client.LedgerApi.GetAccountInfoByAddress(address)
    if err == nil {
        for _, balance := range info.BalanceInfoList {
            if balance.Token.TokenStandard == types.ZnnTokenStandard {
                fmt.Printf("  ZNN Balance (delegation weight): %s\n", balance.Balance.String())
                break
            }
        }
    }
}
```

## Staking

### Stake ZNN for QSR

```go
func stakeZNN(z *zenon.Zenon, amount *big.Int, durationInMonths int) error {
    // Duration in seconds (approximate months)
    duration := int64(durationInMonths * 30 * 24 * 60 * 60)
    
    fmt.Printf("Staking %s ZNN for %d months\n", amount.String(), durationInMonths)
    
    template := z.Client.StakeApi.Stake(duration, amount)
    
    err := z.Send(template)
    if err != nil {
        return fmt.Errorf("staking failed: %v", err)
    }
    
    fmt.Println("Staking successful!")
    fmt.Printf("You will receive QSR rewards over %d months\n", durationInMonths)
    
    return nil
}
```

### Cancel Stake

```go
func cancelStake(z *zenon.Zenon, stakeHash types.Hash) error {
    // Get stake info first
    stakes, err := z.Client.StakeApi.GetStakesByAddress(z.Address(), 0, 10)
    if err != nil {
        return err
    }
    
    var targetStake *embedded.StakeInfo
    for _, stake := range stakes.List {
        if stake.Hash == stakeHash {
            targetStake = stake
            break
        }
    }
    
    if targetStake == nil {
        return fmt.Errorf("stake not found")
    }
    
    fmt.Printf("Canceling stake of %s ZNN\n", targetStake.Amount.String())
    
    template := z.Client.StakeApi.Cancel(stakeHash)
    
    return z.Send(template)
}
```

### List Stakes

```go
func listStakes(z *zenon.Zenon, address types.Address) {
    stakes, err := z.Client.StakeApi.GetStakesByAddress(address, 0, 100)
    if err != nil {
        log.Fatal("Failed to get stakes:", err)
    }
    
    if stakes.Count == 0 {
        fmt.Println("No active stakes")
        return
    }
    
    fmt.Printf("Active Stakes: %d\n", stakes.Count)
    
    for _, stake := range stakes.List {
        fmt.Printf("\nStake Hash: %s\n", stake.Hash.String())
        fmt.Printf("  Amount: %s ZNN\n", stake.Amount.String())
        fmt.Printf("  Duration: %d seconds\n", stake.Duration)
        fmt.Printf("  Start Time: %s\n", time.Unix(stake.StartTime, 0))
        fmt.Printf("  Expiration: %s\n", time.Unix(stake.ExpirationTime, 0))
        
        if time.Now().Unix() > stake.ExpirationTime {
            fmt.Println("  Status: Ready to collect")
        } else {
            remaining := time.Until(time.Unix(stake.ExpirationTime, 0))
            fmt.Printf("  Status: Active (%.1f days remaining)\n", remaining.Hours()/24)
        }
    }
}
```

## Reward Management

### Check Uncollected Rewards

```go
func checkRewards(z *zenon.Zenon, address types.Address) {
    // Pillar rewards
    pillarRewards, err := z.Client.PillarApi.GetUncollectedReward(address)
    if err == nil && (pillarRewards.ZnnAmount.Cmp(big.NewInt(0)) > 0 || 
                      pillarRewards.QsrAmount.Cmp(big.NewInt(0)) > 0) {
        fmt.Println("Pillar Rewards:")
        fmt.Printf("  ZNN: %s\n", pillarRewards.ZnnAmount.String())
        fmt.Printf("  QSR: %s\n", pillarRewards.QsrAmount.String())
    }
    
    // Staking rewards
    stakeRewards, err := z.Client.StakeApi.GetUncollectedReward(address)
    if err == nil && stakeRewards.QsrAmount.Cmp(big.NewInt(0)) > 0 {
        fmt.Println("Staking Rewards:")
        fmt.Printf("  QSR: %s\n", stakeRewards.QsrAmount.String())
    }
    
    // Sentinel rewards
    sentinelRewards, err := z.Client.SentinelApi.GetUncollectedReward(address)
    if err == nil && (sentinelRewards.ZnnAmount.Cmp(big.NewInt(0)) > 0 || 
                       sentinelRewards.QsrAmount.Cmp(big.NewInt(0)) > 0) {
        fmt.Println("Sentinel Rewards:")
        fmt.Printf("  ZNN: %s\n", sentinelRewards.ZnnAmount.String())
        fmt.Printf("  QSR: %s\n", sentinelRewards.QsrAmount.String())
    }
}
```

### Collect Rewards

```go
func collectAllRewards(z *zenon.Zenon) error {
    fmt.Println("Collecting all available rewards...")
    
    // Collect Pillar rewards
    pillarRewards, _ := z.Client.PillarApi.GetUncollectedReward(z.Address())
    if pillarRewards != nil && (pillarRewards.ZnnAmount.Cmp(big.NewInt(0)) > 0 || 
                                 pillarRewards.QsrAmount.Cmp(big.NewInt(0)) > 0) {
        template := z.Client.PillarApi.CollectReward()
        if err := z.Send(template); err != nil {
            log.Printf("Failed to collect pillar rewards: %v", err)
        } else {
            fmt.Println("✓ Collected pillar rewards")
        }
    }
    
    // Collect Staking rewards
    stakeRewards, _ := z.Client.StakeApi.GetUncollectedReward(z.Address())
    if stakeRewards != nil && stakeRewards.QsrAmount.Cmp(big.NewInt(0)) > 0 {
        template := z.Client.StakeApi.CollectReward()
        if err := z.Send(template); err != nil {
            log.Printf("Failed to collect staking rewards: %v", err)
        } else {
            fmt.Println("✓ Collected staking rewards")
        }
    }
    
    return nil
}
```

## Pillar Analytics

### Pillar Performance Tracker

```go
type PillarAnalytics struct {
    client *zenon.Zenon
}

func NewPillarAnalytics(nodeURL string) (*PillarAnalytics, error) {
    z, err := zenon.NewZenon("")
    if err != nil {
        return nil, err
    }
    
    err = z.Start("", nodeURL, 0)
    if err != nil {
        return nil, err
    }
    
    return &PillarAnalytics{client: z}, nil
}

func (pa *PillarAnalytics) GetTopPillars(count int) {
    pillars, err := pa.client.Client.PillarApi.GetAll(0, uint32(count))
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Top %d Pillars by Weight:\n", count)
    for i, pillar := range pillars.List {
        uptime := calculateUptime(pillar)
        fmt.Printf("%d. %s\n", i+1, pillar.Name)
        fmt.Printf("   Weight: %s\n", pillar.Weight.String())
        fmt.Printf("   Uptime: %.2f%%\n", uptime)
        fmt.Printf("   Rewards: Block %d%%, Delegate %d%%\n",
            pillar.GiveBlockRewardPercentage,
            pillar.GiveDelegateRewardPercentage)
    }
}

func (pa *PillarAnalytics) FindBestPillarForDelegation() {
    pillars, err := pa.client.Client.PillarApi.GetAll(0, 100)
    if err != nil {
        log.Fatal(err)
    }
    
    type pillarScore struct {
        pillar *embedded.PillarInfo
        score  float64
    }
    
    scores := make([]pillarScore, 0)
    
    for _, pillar := range pillars.List {
        uptime := calculateUptime(pillar)
        
        // Skip low-performance pillars
        if uptime < 50 {
            continue
        }
        
        // Calculate score based on uptime and reward sharing
        score := uptime * 0.7 + // 70% weight on uptime
                float64(pillar.GiveDelegateRewardPercentage) * 0.3 // 30% on rewards
        
        scores = append(scores, pillarScore{pillar, score})
    }
    
    // Sort by score
    sort.Slice(scores, func(i, j int) bool {
        return scores[i].score > scores[j].score
    })
    
    fmt.Println("Best Pillars for Delegation:")
    for i := 0; i < 5 && i < len(scores); i++ {
        p := scores[i].pillar
        fmt.Printf("%d. %s (Score: %.2f)\n", i+1, p.Name, scores[i].score)
        fmt.Printf("   Uptime: %.2f%%, Delegate Reward: %d%%\n",
            calculateUptime(p),
            p.GiveDelegateRewardPercentage)
    }
}
```

## Sentinel Management

### Register Sentinel

```go
func registerSentinel(z *zenon.Zenon) error {
    // Check QSR deposit (need 5000 QSR)
    requiredQSR := big.NewInt(5000 * 1e8)
    
    // First deposit QSR
    fmt.Printf("Depositing %s QSR for Sentinel\n", requiredQSR.String())
    
    template := z.Client.SentinelApi.DepositQsr(requiredQSR)
    err := z.Send(template)
    if err != nil {
        return fmt.Errorf("failed to deposit QSR: %v", err)
    }
    
    time.Sleep(10 * time.Second) // Wait for confirmation
    
    // Register Sentinel
    fmt.Println("Registering Sentinel...")
    
    template = z.Client.SentinelApi.Register()
    err = z.Send(template)
    if err != nil {
        return fmt.Errorf("failed to register sentinel: %v", err)
    }
    
    fmt.Println("Sentinel registered successfully!")
    return nil
}
```

## Complete Example: Delegation Manager

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

type DelegationManager struct {
    client *zenon.Zenon
}

func NewDelegationManager(wallet, password, nodeURL string) (*DelegationManager, error) {
    z, err := zenon.NewZenon(wallet)
    if err != nil {
        return nil, err
    }
    
    err = z.Start(password, nodeURL, 0)
    if err != nil {
        return nil, err
    }
    
    return &DelegationManager{client: z}, nil
}

func (dm *DelegationManager) OptimizeDelegation() error {
    // Check current delegation
    current, _ := dm.client.Client.PillarApi.GetDelegatedPillar(dm.client.Address())
    
    if current != nil {
        fmt.Printf("Currently delegating to: %s\n", current.Name)
        
        // Check performance
        uptime := calculateUptime(current)
        if uptime < 80 {
            fmt.Printf("Warning: Pillar uptime is low (%.2f%%)\n", uptime)
            fmt.Println("Consider switching to a better performing pillar")
        }
    }
    
    // Find best pillar
    best := dm.findBestPillar()
    if best == nil {
        return fmt.Errorf("no suitable pillar found")
    }
    
    if current != nil && current.Name == best.Name {
        fmt.Println("Already delegating to the best pillar")
        return nil
    }
    
    // Switch delegation
    if current != nil {
        fmt.Println("Undelegating from current pillar...")
        template := dm.client.Client.PillarApi.Undelegate()
        if err := dm.client.Send(template); err != nil {
            return err
        }
        time.Sleep(10 * time.Second)
    }
    
    fmt.Printf("Delegating to: %s\n", best.Name)
    template := dm.client.Client.PillarApi.Delegate(best.Name)
    return dm.client.Send(template)
}

func (dm *DelegationManager) findBestPillar() *embedded.PillarInfo {
    pillars, err := dm.client.Client.PillarApi.GetAll(0, 50)
    if err != nil {
        return nil
    }
    
    var best *embedded.PillarInfo
    bestScore := 0.0
    
    for _, pillar := range pillars.List {
        uptime := calculateUptime(pillar)
        if uptime < 80 {
            continue // Skip low uptime pillars
        }
        
        score := uptime + float64(pillar.GiveDelegateRewardPercentage)
        
        if score > bestScore {
            best = pillar
            bestScore = score
        }
    }
    
    return best
}

func (dm *DelegationManager) MonitorRewards() {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        // Check rewards
        rewards, err := dm.client.Client.PillarApi.GetUncollectedReward(dm.client.Address())
        if err != nil {
            continue
        }
        
        if rewards.ZnnAmount.Cmp(big.NewInt(1e8)) > 0 { // More than 1 ZNN
            fmt.Println("Collecting delegation rewards...")
            template := dm.client.Client.PillarApi.CollectReward()
            dm.client.Send(template)
        }
    }
}
```

## Summary

You've learned:
- ✅ Understanding Pillars and their role in Zenon
- ✅ Registering and managing Pillars
- ✅ Delegation strategies and optimization
- ✅ Staking ZNN for QSR rewards
- ✅ Collecting and managing rewards
- ✅ Sentinel registration and management
- ✅ Analytics for finding best Pillars

Next: [07-plasma-and-feeless-txs.md](./07-plasma-and-feeless-txs.md)