# Tutorial 02: Wallet Management

Learn how to create, manage, and use wallets with the Zenon Go SDK. This tutorial covers key generation, wallet storage, and security best practices.

## Understanding Zenon Wallets

### Key Concepts

1. **KeyFile**: Encrypted JSON file containing your private keys
2. **Mnemonic**: 24-word seed phrase for wallet recovery
3. **HD Wallet**: Hierarchical Deterministic wallet (multiple addresses from one seed)
4. **KeyPair**: Public/private key pair for a specific address

### Wallet Structure
```
KeyFile
├── Mnemonic (encrypted)
├── Base Address (first derived address)
└── Derivation capability (generate multiple addresses)
```

## Creating a New Wallet

### Method 1: Generate New KeyFile

```go
package main

import (
    "fmt"
    "log"
    "github.com/MoonBaZZe/znn-sdk-go/wallet"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

func main() {
    // Generate a new keystore with random mnemonic
    keyStore, err := wallet.NewKeyStore()
    if err != nil {
        log.Fatal("Failed to create keystore:", err)
    }
    
    // Display the mnemonic (SAVE THIS SECURELY!)
    fmt.Println("IMPORTANT: Save this mnemonic phrase securely!")
    fmt.Println("Mnemonic:", keyStore.Mnemonic)
    fmt.Println()
    
    // Display the base address
    fmt.Println("Base Address:", keyStore.BaseAddress.String())
    
    // Save to file with encryption
    password := "your-secure-password"
    filename := "my-wallet"
    
    err = wallet.WriteKeyFile(keyStore, filename, password)
    if err != nil {
        log.Fatal("Failed to save wallet:", err)
    }
    
    fmt.Printf("Wallet saved to: %s/%s\n", wallet.DefaultWalletDir, filename)
}
```

### Method 2: Import from Mnemonic

```go
package main

import (
    "fmt"
    "log"
    "github.com/MoonBaZZe/znn-sdk-go/wallet"
)

func main() {
    // Example mnemonic (use your own!)
    mnemonic := "your twenty four word mnemonic phrase goes here separated by spaces"
    
    // Create keystore from mnemonic
    keyStore, err := wallet.NewKeyStoreFromMnemonic(mnemonic)
    if err != nil {
        log.Fatal("Invalid mnemonic:", err)
    }
    
    fmt.Println("Wallet imported successfully!")
    fmt.Println("Base Address:", keyStore.BaseAddress.String())
    
    // Save the imported wallet
    err = wallet.WriteKeyFile(keyStore, "imported-wallet", "password123")
    if err != nil {
        log.Fatal("Failed to save wallet:", err)
    }
}
```

## Loading and Using a Wallet

### Basic Wallet Loading

```go
package main

import (
    "fmt"
    "log"
    "github.com/MoonBaZZe/znn-sdk-go/wallet"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

func main() {
    // Read existing keyfile
    keyStore, err := wallet.ReadKeyFile("my-wallet", "your-secure-password", "")
    if err != nil {
        log.Fatal("Failed to read wallet:", err)
    }
    
    fmt.Println("Wallet loaded successfully!")
    fmt.Println("Base Address:", keyStore.BaseAddress.String())
    
    // Derive addresses (HD wallet feature)
    for i := uint32(0); i < 5; i++ {
        _, keyPair, err := keyStore.DeriveForIndexPath(i)
        if err != nil {
            log.Printf("Failed to derive address %d: %v", i, err)
            continue
        }
        fmt.Printf("Address[%d]: %s\n", i, keyPair.Address.String())
    }
}
```

### Connect with Wallet for Transactions

```go
package main

import (
    "fmt"
    "log"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
)

func main() {
    // Initialize with wallet
    z, err := zenon.NewZenon("my-wallet") // Wallet filename
    if err != nil {
        log.Fatal("Failed to create Zenon client:", err)
    }
    
    // Connect and unlock wallet
    password := "your-secure-password"
    nodeURL := "ws://127.0.0.1:35998"
    addressIndex := uint32(0) // Use first derived address
    
    err = z.Start(password, nodeURL, addressIndex)
    if err != nil {
        log.Fatal("Failed to start client:", err)
    }
    defer z.Stop()
    
    // Now you can send transactions
    fmt.Println("Connected with wallet!")
    fmt.Println("Active address:", z.Address().String())
}
```

## Address Management

### Working with Multiple Addresses

```go
package main

import (
    "fmt"
    "log"
    "github.com/MoonBaZZe/znn-sdk-go/wallet"
)

type WalletManager struct {
    keyStore *wallet.KeyStore
    addresses map[uint32]string
}

func NewWalletManager(walletFile, password string) (*WalletManager, error) {
    ks, err := wallet.ReadKeyFile(walletFile, password, "")
    if err != nil {
        return nil, err
    }
    
    return &WalletManager{
        keyStore: ks,
        addresses: make(map[uint32]string),
    }, nil
}

func (wm *WalletManager) GetAddress(index uint32) (string, error) {
    // Check cache first
    if addr, exists := wm.addresses[index]; exists {
        return addr, nil
    }
    
    // Derive new address
    _, keyPair, err := wm.keyStore.DeriveForIndexPath(index)
    if err != nil {
        return "", err
    }
    
    address := keyPair.Address.String()
    wm.addresses[index] = address
    
    return address, nil
}

func (wm *WalletManager) ListAddresses(count uint32) []string {
    addresses := make([]string, 0, count)
    
    for i := uint32(0); i < count; i++ {
        addr, err := wm.GetAddress(i)
        if err != nil {
            log.Printf("Failed to get address %d: %v", i, err)
            continue
        }
        addresses = append(addresses, addr)
    }
    
    return addresses
}

func main() {
    wm, err := NewWalletManager("my-wallet", "password")
    if err != nil {
        log.Fatal("Failed to load wallet:", err)
    }
    
    // List first 10 addresses
    addresses := wm.ListAddresses(10)
    for i, addr := range addresses {
        fmt.Printf("Address %d: %s\n", i, addr)
    }
}
```

## Security Best Practices

### 1. Secure Password Generation

```go
import (
    "crypto/rand"
    "encoding/base64"
)

func generateSecurePassword(length int) string {
    bytes := make([]byte, length)
    _, err := rand.Read(bytes)
    if err != nil {
        panic("Failed to generate random password")
    }
    return base64.URLEncoding.EncodeToString(bytes)[:length]
}

// Usage
password := generateSecurePassword(32)
```

### 2. Secure Storage Pattern

```go
package main

import (
    "fmt"
    "os"
    "path/filepath"
    "github.com/MoonBaZZe/znn-sdk-go/wallet"
)

type SecureWalletStorage struct {
    basePath string
}

func NewSecureWalletStorage() *SecureWalletStorage {
    // Use user's home directory
    homeDir, _ := os.UserHomeDir()
    walletDir := filepath.Join(homeDir, ".zenon", "wallets")
    
    // Create directory with restricted permissions
    os.MkdirAll(walletDir, 0700) // Owner read/write/execute only
    
    return &SecureWalletStorage{
        basePath: walletDir,
    }
}

func (s *SecureWalletStorage) SaveWallet(keyStore *wallet.KeyStore, name, password string) error {
    // Set custom wallet directory
    wallet.DefaultWalletDir = s.basePath
    
    // Save with restrictive permissions
    err := wallet.WriteKeyFile(keyStore, name, password)
    if err != nil {
        return err
    }
    
    // Additional: Set file permissions explicitly
    walletPath := filepath.Join(s.basePath, name)
    os.Chmod(walletPath, 0600) // Owner read/write only
    
    return nil
}

func (s *SecureWalletStorage) LoadWallet(name, password string) (*wallet.KeyStore, error) {
    wallet.DefaultWalletDir = s.basePath
    return wallet.ReadKeyFile(name, password, s.basePath)
}
```

### 3. Memory Security

```go
import "runtime"

func secureCleanup(sensitive *string) {
    if sensitive != nil && *sensitive != "" {
        // Overwrite sensitive data
        for i := range *sensitive {
            (*sensitive)[i] = 0
        }
        *sensitive = ""
        
        // Force garbage collection
        runtime.GC()
    }
}

// Usage
password := "sensitive-password"
defer secureCleanup(&password)
```

## Wallet Backup and Recovery

### Creating Encrypted Backups

```go
package main

import (
    "encoding/json"
    "io/ioutil"
    "time"
    "fmt"
    "github.com/MoonBaZZe/znn-sdk-go/wallet"
)

type WalletBackup struct {
    Timestamp   time.Time `json:"timestamp"`
    Version     string    `json:"version"`
    KeyFileData []byte    `json:"keyfile_data"`
    Checksum    string    `json:"checksum"`
}

func createBackup(walletName, password string) error {
    // Read the wallet file
    walletPath := fmt.Sprintf("%s/%s", wallet.DefaultWalletDir, walletName)
    keyFileData, err := ioutil.ReadFile(walletPath)
    if err != nil {
        return err
    }
    
    // Create backup structure
    backup := WalletBackup{
        Timestamp:   time.Now(),
        Version:     "1.0",
        KeyFileData: keyFileData,
        // Add checksum for integrity
    }
    
    // Save backup
    backupData, err := json.MarshalIndent(backup, "", "  ")
    if err != nil {
        return err
    }
    
    backupName := fmt.Sprintf("backup_%s_%d.json", walletName, time.Now().Unix())
    return ioutil.WriteFile(backupName, backupData, 0600)
}

func restoreBackup(backupFile, newWalletName string) error {
    // Read backup
    backupData, err := ioutil.ReadFile(backupFile)
    if err != nil {
        return err
    }
    
    var backup WalletBackup
    err = json.Unmarshal(backupData, &backup)
    if err != nil {
        return err
    }
    
    // Restore wallet file
    walletPath := fmt.Sprintf("%s/%s", wallet.DefaultWalletDir, newWalletName)
    return ioutil.WriteFile(walletPath, backup.KeyFileData, 0600)
}
```

## Testing Wallet Operations

```go
package main

import (
    "testing"
    "os"
    "github.com/MoonBaZZe/znn-sdk-go/wallet"
)

func TestWalletCreation(t *testing.T) {
    // Create new wallet
    ks, err := wallet.NewKeyStore()
    if err != nil {
        t.Fatal("Failed to create wallet:", err)
    }
    
    // Verify mnemonic
    if len(ks.Mnemonic) == 0 {
        t.Fatal("Empty mnemonic")
    }
    
    // Verify base address
    if ks.BaseAddress.String() == "" {
        t.Fatal("Invalid base address")
    }
    
    t.Log("Wallet creation test passed")
}

func TestWalletSaveLoad(t *testing.T) {
    // Create wallet
    ks, _ := wallet.NewKeyStore()
    originalAddress := ks.BaseAddress.String()
    
    // Save wallet
    testName := "test-wallet"
    testPassword := "test-password-123"
    
    err := wallet.WriteKeyFile(ks, testName, testPassword)
    if err != nil {
        t.Fatal("Failed to save wallet:", err)
    }
    
    // Load wallet
    loadedKs, err := wallet.ReadKeyFile(testName, testPassword, "")
    if err != nil {
        t.Fatal("Failed to load wallet:", err)
    }
    
    // Verify addresses match
    if loadedKs.BaseAddress.String() != originalAddress {
        t.Fatal("Address mismatch after load")
    }
    
    // Cleanup
    os.Remove(wallet.DefaultWalletDir + "/" + testName)
    
    t.Log("Save/Load test passed")
}

func TestAddressDerivation(t *testing.T) {
    ks, _ := wallet.NewKeyStore()
    
    addresses := make(map[string]bool)
    
    // Derive multiple addresses
    for i := uint32(0); i < 10; i++ {
        _, kp, err := ks.DeriveForIndexPath(i)
        if err != nil {
            t.Fatal("Failed to derive address:", err)
        }
        
        addr := kp.Address.String()
        
        // Check uniqueness
        if addresses[addr] {
            t.Fatal("Duplicate address derived")
        }
        addresses[addr] = true
    }
    
    t.Log("Address derivation test passed")
}
```

## Common Wallet Patterns

### 1. Wallet Pool for Multiple Operations
```go
type WalletPool struct {
    wallets map[string]*zenon.Zenon
}

func (wp *WalletPool) AddWallet(name, password, nodeURL string) error {
    z, err := zenon.NewZenon(name)
    if err != nil {
        return err
    }
    
    err = z.Start(password, nodeURL, 0)
    if err != nil {
        return err
    }
    
    wp.wallets[name] = z
    return nil
}
```

### 2. Automatic Wallet Selection
```go
func selectWalletWithBalance(wallets []*zenon.Zenon, minBalance *big.Int) *zenon.Zenon {
    for _, w := range wallets {
        balance, _ := getBalance(w.Address())
        if balance.Cmp(minBalance) >= 0 {
            return w
        }
    }
    return nil
}
```

## Troubleshooting

### Common Issues

1. **"Cannot decrypt keyfile"**
   - Wrong password
   - Corrupted keyfile
   - Wrong file path

2. **"Invalid mnemonic"**
   - Check word count (should be 24)
   - Verify word spelling
   - Ensure proper spacing

3. **"Wallet not found"**
   - Check DefaultWalletDir path
   - Verify file permissions
   - Ensure file exists

## Exercise

1. Create a wallet manager that:
   - Generates new wallets
   - Lists all wallet files
   - Shows balances for each wallet
   - Exports/imports mnemonics securely

2. Implement a secure wallet service with:
   - Password strength validation
   - Encrypted backup system
   - Multi-signature support simulation

## Summary

You've learned:
- ✅ Creating and importing wallets
- ✅ Managing multiple addresses (HD wallet)
- ✅ Secure storage practices
- ✅ Backup and recovery procedures
- ✅ Integration with Zenon client

Next: [03-reading-blockchain-data.md](./03-reading-blockchain-data.md)