# Tutorial 05: Working with Tokens

Learn how to create, manage, and interact with ZTS (Zenon Token Standard) tokens on the Zenon Network.

## Understanding ZTS Tokens

### What are ZTS Tokens?
- Native tokens on Zenon Network
- Inherit properties of ZNN/QSR: security, censorship-resistance, feeless transfers
- Can be mintable, burnable, and have utility properties
- Created through the Token embedded contract

### Token Properties
```go
type Token struct {
    Name          string
    Symbol        string
    Domain        string
    TotalSupply   *big.Int
    MaxSupply     *big.Int
    Decimals      uint8
    Owner         types.Address
    TokenStandard types.ZenonTokenStandard
    IsMintable    bool
    IsBurnable    bool
    IsUtility     bool
}
```

## Creating a New Token

### Basic Token Creation

```go
package main

import (
    "fmt"
    "log"
    "math/big"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
    "github.com/zenon-network/go-zenon/common/types"
)

func createToken(z *zenon.Zenon) error {
    // Token parameters
    name := "My Token"
    symbol := "MTK"
    domain := "mytoken.com"
    decimals := uint8(8)
    
    // Supply configuration
    totalSupply := big.NewInt(1000000 * 1e8)  // 1 million tokens
    maxSupply := big.NewInt(10000000 * 1e8)   // 10 million max
    
    // Token properties
    isMintable := true  // Can mint more tokens later
    isBurnable := true  // Holders can burn tokens
    isUtility := false  // Not a utility token
    
    fmt.Println("Creating token:", name)
    
    // Create token (costs ZNN + QSR)
    template := z.Client.TokenApi.IssueToken(
        name,
        symbol,
        domain,
        totalSupply,
        maxSupply,
        decimals,
        isMintable,
        isBurnable,
        isUtility,
    )
    
    err := z.Send(template)
    if err != nil {
        return fmt.Errorf("failed to create token: %v", err)
    }
    
    fmt.Println("Token creation transaction sent!")
    fmt.Println("Note: Token creation costs 1 ZNN")
    
    return nil
}

func main() {
    z, err := zenon.NewZenon("my-wallet")
    if err != nil {
        log.Fatal(err)
    }
    
    err = z.Start("password", "ws://127.0.0.1:35998", 0)
    if err != nil {
        log.Fatal(err)
    }
    defer z.Stop()
    
    // Check balance first (need 1 ZNN for token creation)
    info, _ := z.Client.LedgerApi.GetAccountInfoByAddress(z.Address())
    fmt.Printf("Current balance: %v\n", info.BalanceInfoList)
    
    err = createToken(z)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Advanced Token Creation

```go
package main

import (
    "fmt"
    "math/big"
    "time"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

type TokenCreator struct {
    client *zenon.Zenon
}

func NewTokenCreator(wallet, password, nodeURL string) (*TokenCreator, error) {
    z, err := zenon.NewZenon(wallet)
    if err != nil {
        return nil, err
    }
    
    err = z.Start(password, nodeURL, 0)
    if err != nil {
        return nil, err
    }
    
    return &TokenCreator{client: z}, nil
}

func (tc *TokenCreator) CreateTokenWithValidation(config TokenConfig) (types.ZenonTokenStandard, error) {
    // Validate parameters
    if err := config.Validate(); err != nil {
        return types.ZenonTokenStandard{}, err
    }
    
    // Check issuer balance (need 1 ZNN)
    balance, err := tc.getZNNBalance()
    if err != nil {
        return types.ZenonTokenStandard{}, err
    }
    
    requiredZNN := big.NewInt(1e8) // 1 ZNN
    if balance.Cmp(requiredZNN) < 0 {
        return types.ZenonTokenStandard{}, fmt.Errorf("insufficient ZNN: need 1 ZNN for token creation")
    }
    
    // Create token
    fmt.Printf("Creating token: %s (%s)\n", config.Name, config.Symbol)
    
    template := tc.client.Client.TokenApi.IssueToken(
        config.Name,
        config.Symbol,
        config.Domain,
        config.TotalSupply,
        config.MaxSupply,
        config.Decimals,
        config.IsMintable,
        config.IsBurnable,
        config.IsUtility,
    )
    
    // Get account height before sending
    preHeight := tc.getAccountHeight()
    
    err = tc.client.Send(template)
    if err != nil {
        return types.ZenonTokenStandard{}, err
    }
    
    // Wait for token to be created
    tokenStandard, err := tc.waitForTokenCreation(preHeight, config.Symbol)
    if err != nil {
        return types.ZenonTokenStandard{}, err
    }
    
    fmt.Printf("Token created successfully! ZTS: %s\n", tokenStandard.String())
    
    return tokenStandard, nil
}

func (tc *TokenCreator) waitForTokenCreation(preHeight uint64, symbol string) (types.ZenonTokenStandard, error) {
    deadline := time.Now().Add(2 * time.Minute)
    
    for time.Now().Before(deadline) {
        // Check if new block created
        currentHeight := tc.getAccountHeight()
        if currentHeight > preHeight {
            // Search for the new token
            tokens, err := tc.client.Client.TokenApi.GetByOwner(tc.client.Address(), 0, 10)
            if err == nil {
                for _, token := range tokens.List {
                    if token.Symbol == symbol {
                        return token.TokenStandard, nil
                    }
                }
            }
        }
        
        time.Sleep(2 * time.Second)
    }
    
    return types.ZenonTokenStandard{}, fmt.Errorf("token creation timeout")
}

func (tc *TokenCreator) getZNNBalance() (*big.Int, error) {
    info, err := tc.client.Client.LedgerApi.GetAccountInfoByAddress(tc.client.Address())
    if err != nil {
        return nil, err
    }
    
    for _, balance := range info.BalanceInfoList {
        if balance.Token.TokenStandard == types.ZnnTokenStandard {
            return balance.Balance, nil
        }
    }
    
    return big.NewInt(0), nil
}

func (tc *TokenCreator) getAccountHeight() uint64 {
    info, _ := tc.client.Client.LedgerApi.GetAccountInfoByAddress(tc.client.Address())
    return info.AccountHeight
}

type TokenConfig struct {
    Name        string
    Symbol      string
    Domain      string
    TotalSupply *big.Int
    MaxSupply   *big.Int
    Decimals    uint8
    IsMintable  bool
    IsBurnable  bool
    IsUtility   bool
}

func (tc *TokenConfig) Validate() error {
    if len(tc.Name) == 0 || len(tc.Name) > 40 {
        return fmt.Errorf("token name must be 1-40 characters")
    }
    
    if len(tc.Symbol) == 0 || len(tc.Symbol) > 10 {
        return fmt.Errorf("token symbol must be 1-10 characters")
    }
    
    if len(tc.Domain) > 128 {
        return fmt.Errorf("domain must be <= 128 characters")
    }
    
    if tc.Decimals > 18 {
        return fmt.Errorf("decimals must be <= 18")
    }
    
    if tc.TotalSupply.Cmp(tc.MaxSupply) > 0 {
        return fmt.Errorf("total supply cannot exceed max supply")
    }
    
    return nil
}
```

## Managing Existing Tokens

### Minting Tokens

```go
func mintTokens(z *zenon.Zenon, tokenStandard types.ZenonTokenStandard, amount *big.Int, recipient types.Address) error {
    // Check if we own the token
    token, err := z.Client.TokenApi.GetByZts(tokenStandard)
    if err != nil {
        return fmt.Errorf("token not found: %v", err)
    }
    
    if token.Owner != z.Address() {
        return fmt.Errorf("not token owner")
    }
    
    if !token.IsMintable {
        return fmt.Errorf("token is not mintable")
    }
    
    // Check max supply limit
    newTotal := new(big.Int).Add(token.TotalSupply, amount)
    if newTotal.Cmp(token.MaxSupply) > 0 {
        return fmt.Errorf("would exceed max supply")
    }
    
    fmt.Printf("Minting %s %s to %s\n", amount.String(), token.Symbol, recipient.String())
    
    template := z.Client.TokenApi.Mint(tokenStandard, amount, recipient)
    
    return z.Send(template)
}
```

### Burning Tokens

```go
func burnTokens(z *zenon.Zenon, tokenStandard types.ZenonTokenStandard, amount *big.Int) error {
    // Check token properties
    token, err := z.Client.TokenApi.GetByZts(tokenStandard)
    if err != nil {
        return fmt.Errorf("token not found: %v", err)
    }
    
    if !token.IsBurnable {
        return fmt.Errorf("token is not burnable")
    }
    
    // Check balance
    balance, err := getTokenBalance(z, z.Address(), tokenStandard)
    if err != nil {
        return err
    }
    
    if balance.Cmp(amount) < 0 {
        return fmt.Errorf("insufficient balance")
    }
    
    fmt.Printf("Burning %s %s\n", amount.String(), token.Symbol)
    
    template := z.Client.TokenApi.Burn(tokenStandard, amount)
    
    return z.Send(template)
}

func getTokenBalance(z *zenon.Zenon, address types.Address, tokenStandard types.ZenonTokenStandard) (*big.Int, error) {
    info, err := z.Client.LedgerApi.GetAccountInfoByAddress(address)
    if err != nil {
        return nil, err
    }
    
    for _, balance := range info.BalanceInfoList {
        if balance.Token.TokenStandard == tokenStandard {
            return balance.Balance, nil
        }
    }
    
    return big.NewInt(0), nil
}
```

### Updating Token Properties

```go
func updateToken(z *zenon.Zenon, tokenStandard types.ZenonTokenStandard, newOwner types.Address, isMintable, isBurnable bool) error {
    // Verify ownership
    token, err := z.Client.TokenApi.GetByZts(tokenStandard)
    if err != nil {
        return err
    }
    
    if token.Owner != z.Address() {
        return fmt.Errorf("not token owner")
    }
    
    fmt.Printf("Updating token %s\n", token.Symbol)
    
    template := z.Client.TokenApi.UpdateToken(
        tokenStandard,
        newOwner,
        isMintable,
        isBurnable,
    )
    
    return z.Send(template)
}
```

## Token Explorer

### List All Tokens

```go
package main

import (
    "fmt"
    "log"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

func exploreTokens(z *zenon.Zenon) {
    pageIndex := uint32(0)
    pageSize := uint32(50)
    
    for {
        tokens, err := z.Client.TokenApi.GetAll(pageIndex, pageSize)
        if err != nil {
            log.Printf("Error fetching tokens: %v", err)
            break
        }
        
        fmt.Printf("Page %d - Tokens %d-%d of %d\n", 
            pageIndex+1,
            pageIndex*pageSize+1,
            pageIndex*pageSize+uint32(len(tokens.List)),
            tokens.Count)
        
        for _, token := range tokens.List {
            fmt.Printf("\n%s (%s)\n", token.Name, token.Symbol)
            fmt.Printf("  ZTS: %s\n", token.TokenStandard.String())
            fmt.Printf("  Owner: %s\n", token.Owner.String())
            fmt.Printf("  Supply: %s / %s\n", token.TotalSupply.String(), token.MaxSupply.String())
            fmt.Printf("  Decimals: %d\n", token.Decimals)
            fmt.Printf("  Properties: Mintable=%t, Burnable=%t, Utility=%t\n",
                token.IsMintable, token.IsBurnable, token.IsUtility)
            
            if token.Domain != "" {
                fmt.Printf("  Domain: %s\n", token.Domain)
            }
        }
        
        if uint32(len(tokens.List)) < pageSize {
            break // Last page
        }
        
        pageIndex++
    }
}
```

### Token Statistics

```go
type TokenStats struct {
    TotalTokens    int
    MintableTokens int
    BurnableTokens int
    UtilityTokens  int
    TotalHolders   map[string]int
}

func gatherTokenStats(z *zenon.Zenon) (*TokenStats, error) {
    stats := &TokenStats{
        TotalHolders: make(map[string]int),
    }
    
    pageIndex := uint32(0)
    pageSize := uint32(100)
    
    for {
        tokens, err := z.Client.TokenApi.GetAll(pageIndex, pageSize)
        if err != nil {
            return nil, err
        }
        
        for _, token := range tokens.List {
            stats.TotalTokens++
            
            if token.IsMintable {
                stats.MintableTokens++
            }
            if token.IsBurnable {
                stats.BurnableTokens++
            }
            if token.IsUtility {
                stats.UtilityTokens++
            }
        }
        
        if uint32(len(tokens.List)) < pageSize {
            break
        }
        pageIndex++
    }
    
    return stats, nil
}
```

## Token Transfer Service

```go
package main

import (
    "fmt"
    "math/big"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
    "github.com/zenon-network/go-zenon/common/types"
)

type TokenService struct {
    client *zenon.Zenon
}

func NewTokenService(wallet, password, nodeURL string) (*TokenService, error) {
    z, err := zenon.NewZenon(wallet)
    if err != nil {
        return nil, err
    }
    
    err = z.Start(password, nodeURL, 0)
    if err != nil {
        return nil, err
    }
    
    return &TokenService{client: z}, nil
}

func (ts *TokenService) TransferToken(tokenSymbol string, to types.Address, amount *big.Int) error {
    // Find token by symbol
    token, err := ts.findTokenBySymbol(tokenSymbol)
    if err != nil {
        return err
    }
    
    // Check balance
    balance, err := ts.getBalance(token.TokenStandard)
    if err != nil {
        return err
    }
    
    if balance.Cmp(amount) < 0 {
        return fmt.Errorf("insufficient %s balance: have %s, need %s",
            token.Symbol, balance.String(), amount.String())
    }
    
    // Send tokens
    fmt.Printf("Sending %s %s to %s\n", amount.String(), token.Symbol, to.String())
    
    template := ts.client.Client.LedgerApi.SendTemplate(
        to,
        token.TokenStandard,
        amount,
        []byte{},
    )
    
    return ts.client.Send(template)
}

func (ts *TokenService) findTokenBySymbol(symbol string) (*api.Token, error) {
    pageIndex := uint32(0)
    pageSize := uint32(100)
    
    for {
        tokens, err := ts.client.Client.TokenApi.GetAll(pageIndex, pageSize)
        if err != nil {
            return nil, err
        }
        
        for _, token := range tokens.List {
            if token.Symbol == symbol {
                return token, nil
            }
        }
        
        if uint32(len(tokens.List)) < pageSize {
            break
        }
        pageIndex++
    }
    
    return nil, fmt.Errorf("token %s not found", symbol)
}

func (ts *TokenService) getBalance(tokenStandard types.ZenonTokenStandard) (*big.Int, error) {
    info, err := ts.client.Client.LedgerApi.GetAccountInfoByAddress(ts.client.Address())
    if err != nil {
        return nil, err
    }
    
    for _, balance := range info.BalanceInfoList {
        if balance.Token.TokenStandard == tokenStandard {
            return balance.Balance, nil
        }
    }
    
    return big.NewInt(0), nil
}

func (ts *TokenService) ListMyTokens() error {
    info, err := ts.client.Client.LedgerApi.GetAccountInfoByAddress(ts.client.Address())
    if err != nil {
        return err
    }
    
    fmt.Println("Your token balances:")
    for _, balance := range info.BalanceInfoList {
        if balance.Balance.Cmp(big.NewInt(0)) > 0 {
            fmt.Printf("  %s: %s\n", balance.Token.Symbol, balance.Balance.String())
        }
    }
    
    return nil
}

func (ts *TokenService) GetOwnedTokens() ([]*api.Token, error) {
    tokens, err := ts.client.Client.TokenApi.GetByOwner(ts.client.Address(), 0, 100)
    if err != nil {
        return nil, err
    }
    
    return tokens.List, nil
}
```

## Token Utility Functions

### Format Token Amount

```go
func formatTokenAmount(amount *big.Int, decimals uint8) string {
    divisor := new(big.Float).SetInt(new(big.Int).Exp(
        big.NewInt(10),
        big.NewInt(int64(decimals)),
        nil,
    ))
    
    result := new(big.Float).SetInt(amount)
    result.Quo(result, divisor)
    
    return result.Text('f', int(decimals))
}

// Usage
formatted := formatTokenAmount(big.NewInt(123456789), 8)
fmt.Println(formatted) // "1.23456789"
```

### Parse Token Amount

```go
func parseTokenAmount(amountStr string, decimals uint8) (*big.Int, error) {
    amount, ok := new(big.Float).SetString(amountStr)
    if !ok {
        return nil, fmt.Errorf("invalid amount format")
    }
    
    multiplier := new(big.Float).SetInt(new(big.Int).Exp(
        big.NewInt(10),
        big.NewInt(int64(decimals)),
        nil,
    ))
    
    amount.Mul(amount, multiplier)
    
    result, _ := amount.Int(nil)
    return result, nil
}

// Usage
amount, _ := parseTokenAmount("1.5", 8)
fmt.Println(amount) // 150000000
```

## Testing Token Operations

```go
import "testing"

func TestTokenCreation(t *testing.T) {
    ts, err := NewTokenService("test-wallet", "password", "ws://testnet:35998")
    if err != nil {
        t.Fatal(err)
    }
    defer ts.client.Stop()
    
    config := TokenConfig{
        Name:        "Test Token",
        Symbol:      "TEST",
        Domain:      "test.com",
        TotalSupply: big.NewInt(1000 * 1e8),
        MaxSupply:   big.NewInt(10000 * 1e8),
        Decimals:    8,
        IsMintable:  true,
        IsBurnable:  true,
        IsUtility:   false,
    }
    
    err = config.Validate()
    if err != nil {
        t.Fatal("Invalid config:", err)
    }
    
    t.Log("Token configuration valid")
}

func TestTokenTransfer(t *testing.T) {
    // Test amount parsing
    amount, err := parseTokenAmount("10.5", 8)
    if err != nil {
        t.Fatal(err)
    }
    
    expected := big.NewInt(1050000000)
    if amount.Cmp(expected) != 0 {
        t.Fatalf("Expected %s, got %s", expected.String(), amount.String())
    }
    
    // Test formatting
    formatted := formatTokenAmount(amount, 8)
    if formatted != "10.50000000" {
        t.Fatalf("Expected 10.50000000, got %s", formatted)
    }
}
```

## Common Token Patterns

### Airdrop Tokens

```go
func airdropTokens(z *zenon.Zenon, tokenStandard types.ZenonTokenStandard, recipients []types.Address, amountEach *big.Int) error {
    for i, recipient := range recipients {
        fmt.Printf("Airdropping to %d/%d: %s\n", i+1, len(recipients), recipient.String())
        
        template := z.Client.LedgerApi.SendTemplate(
            recipient,
            tokenStandard,
            amountEach,
            []byte("Airdrop"),
        )
        
        err := z.Send(template)
        if err != nil {
            log.Printf("Failed to airdrop to %s: %v", recipient.String(), err)
            continue
        }
        
        time.Sleep(1 * time.Second) // Rate limiting
    }
    
    return nil
}
```

### Token Swap

```go
func swapTokens(z *zenon.Zenon, fromToken, toToken types.ZenonTokenStandard, amount *big.Int) error {
    // This would interact with a DEX or swap contract
    // Example placeholder for swap logic
    
    fmt.Printf("Swapping %s for %s\n", fromToken.String(), toToken.String())
    
    // In reality, this would call a swap contract
    // For now, just demonstrate the pattern
    
    return fmt.Errorf("swap functionality not implemented")
}
```

## Summary

You've learned:
- ✅ Creating new ZTS tokens with custom properties
- ✅ Minting and burning tokens
- ✅ Updating token ownership and properties
- ✅ Exploring and analyzing tokens
- ✅ Building token transfer services
- ✅ Formatting and parsing token amounts
- ✅ Common token patterns (airdrops, etc.)

Next: [06-pillar-and-delegation.md](./06-pillar-and-delegation.md)