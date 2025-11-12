# Tutorial 08: Building a Complete DApp

Learn how to build a complete decentralized application on Zenon Network, combining all the concepts from previous tutorials into a real-world project.

## Project Overview: ZenonPay

We'll build **ZenonPay** - a payment and wallet management system that showcases:
- Multi-wallet management
- Token transfers and payments
- Plasma optimization
- Pillar delegation
- Real-time monitoring
- Web interface

### Features
- 💳 Multi-wallet support
- 💸 Token transfers (ZNN, QSR, ZTS)
- ⚡ Automatic Plasma management
- 🏛️ Pillar delegation optimization
- 📊 Real-time balance monitoring
- 🌐 Web dashboard
- 📱 REST API

## Project Structure

```
zenonpay/
├── cmd/
│   ├── server/          # Web server
│   └── cli/             # Command-line tool
├── internal/
│   ├── api/             # REST API handlers
│   ├── service/         # Business logic
│   ├── wallet/          # Wallet management
│   ├── plasma/          # Plasma optimization
│   ├── delegation/      # Pillar delegation
│   └── monitor/         # Real-time monitoring
├── web/
│   ├── static/          # CSS, JS
│   └── templates/       # HTML templates
├── config/
│   └── config.go        # Configuration
└── go.mod
```

## Core Service Layer

### Configuration

```go
// config/config.go
package config

import (
    "encoding/json"
    "os"
)

type Config struct {
    Server   ServerConfig   `json:"server"`
    Zenon    ZenonConfig    `json:"zenon"`
    Wallets  WalletsConfig  `json:"wallets"`
    Features FeatureConfig  `json:"features"`
}

type ServerConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

type ZenonConfig struct {
    NodeURL string `json:"nodeUrl"`
}

type WalletsConfig struct {
    Directory string `json:"directory"`
}

type FeatureConfig struct {
    AutoPlasma     bool `json:"autoPlasma"`
    AutoDelegation bool `json:"autoDelegation"`
    Monitoring     bool `json:"monitoring"`
}

func LoadConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var config Config
    err = json.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }
    
    return &config, nil
}

func (c *Config) Save(filename string) error {
    data, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(filename, data, 0644)
}
```

### Wallet Service

```go
// internal/wallet/service.go
package wallet

import (
    "fmt"
    "math/big"
    "sync"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
    "github.com/MoonBaZZe/znn-sdk-go/wallet"
    "github.com/zenon-network/go-zenon/common/types"
)

type Service struct {
    wallets map[string]*WalletInstance
    mutex   sync.RWMutex
    nodeURL string
}

type WalletInstance struct {
    Name     string
    Client   *zenon.Zenon
    Address  types.Address
    Unlocked bool
}

type Balance struct {
    Token  string   `json:"token"`
    Symbol string   `json:"symbol"`
    Amount *big.Int `json:"amount"`
}

func NewService(nodeURL string) *Service {
    return &Service{
        wallets: make(map[string]*WalletInstance),
        nodeURL: nodeURL,
    }
}

func (s *Service) CreateWallet(name, password string) (*WalletInstance, error) {
    // Generate new wallet
    keyStore, err := wallet.NewKeyStore()
    if err != nil {
        return nil, err
    }
    
    // Save wallet file
    err = wallet.WriteKeyFile(keyStore, name, password)
    if err != nil {
        return nil, err
    }
    
    // Load and connect
    return s.LoadWallet(name, password)
}

func (s *Service) LoadWallet(name, password string) (*WalletInstance, error) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    // Check if already loaded
    if wallet, exists := s.wallets[name]; exists {
        return wallet, nil
    }
    
    // Create Zenon client
    z, err := zenon.NewZenon(name)
    if err != nil {
        return nil, err
    }
    
    // Connect and unlock
    err = z.Start(password, s.nodeURL, 0)
    if err != nil {
        return nil, err
    }
    
    instance := &WalletInstance{
        Name:     name,
        Client:   z,
        Address:  z.Address(),
        Unlocked: true,
    }
    
    s.wallets[name] = instance
    
    return instance, nil
}

func (s *Service) GetWallet(name string) (*WalletInstance, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    wallet, exists := s.wallets[name]
    if !exists {
        return nil, fmt.Errorf("wallet %s not loaded", name)
    }
    
    return wallet, nil
}

func (s *Service) GetBalances(walletName string) ([]Balance, error) {
    wallet, err := s.GetWallet(walletName)
    if err != nil {
        return nil, err
    }
    
    info, err := wallet.Client.Client.LedgerApi.GetAccountInfoByAddress(wallet.Address)
    if err != nil {
        return nil, err
    }
    
    balances := make([]Balance, len(info.BalanceInfoList))
    for i, balance := range info.BalanceInfoList {
        balances[i] = Balance{
            Token:  balance.Token.TokenStandard.String(),
            Symbol: balance.Token.Symbol,
            Amount: balance.Balance,
        }
    }
    
    return balances, nil
}

func (s *Service) SendPayment(walletName string, to types.Address, token types.ZenonTokenStandard, amount *big.Int) error {
    wallet, err := s.GetWallet(walletName)
    if err != nil {
        return err
    }
    
    template := wallet.Client.Client.LedgerApi.SendTemplate(to, token, amount, []byte{})
    
    return wallet.Client.Send(template)
}

func (s *Service) ListWallets() []string {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    names := make([]string, 0, len(s.wallets))
    for name := range s.wallets {
        names = append(names, name)
    }
    
    return names
}

func (s *Service) CloseWallet(name string) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    wallet, exists := s.wallets[name]
    if !exists {
        return fmt.Errorf("wallet not found")
    }
    
    err := wallet.Client.Stop()
    delete(s.wallets, name)
    
    return err
}

func (s *Service) Shutdown() {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    for name, wallet := range s.wallets {
        wallet.Client.Stop()
        delete(s.wallets, name)
    }
}
```

### Plasma Service

```go
// internal/plasma/service.go
package plasma

import (
    "fmt"
    "log"
    "math/big"
    "time"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
    "github.com/zenon-network/go-zenon/common/types"
)

type Service struct {
    autoFusion      bool
    minPlasmaRatio  float64
    defaultFusion   *big.Int
    monitorInterval time.Duration
}

type PlasmaStatus struct {
    CurrentPlasma uint64  `json:"currentPlasma"`
    MaxPlasma     uint64  `json:"maxPlasma"`
    QSRFused      *big.Int `json:"qsrFused"`
    Ratio         float64 `json:"ratio"`
    Status        string  `json:"status"`
}

func NewService() *Service {
    return &Service{
        autoFusion:      true,
        minPlasmaRatio:  0.2,
        defaultFusion:   big.NewInt(10 * 1e8), // 10 QSR
        monitorInterval: 30 * time.Second,
    }
}

func (s *Service) GetPlasmaStatus(client *zenon.Zenon) (*PlasmaStatus, error) {
    info, err := client.Client.PlasmaApi.Get(client.Address())
    if err != nil {
        return nil, err
    }
    
    var ratio float64
    if info.MaxPlasma > 0 {
        ratio = float64(info.CurrentPlasma) / float64(info.MaxPlasma)
    }
    
    status := s.getStatusString(ratio)
    
    return &PlasmaStatus{
        CurrentPlasma: info.CurrentPlasma,
        MaxPlasma:     info.MaxPlasma,
        QSRFused:      info.QsrAmount,
        Ratio:         ratio,
        Status:        status,
    }, nil
}

func (s *Service) getStatusString(ratio float64) string {
    switch {
    case ratio >= 0.8:
        return "Excellent"
    case ratio >= 0.5:
        return "Good"
    case ratio >= 0.2:
        return "Low"
    default:
        return "Critical"
    }
}

func (s *Service) OptimizePlasma(client *zenon.Zenon) error {
    status, err := s.GetPlasmaStatus(client)
    if err != nil {
        return err
    }
    
    if status.Ratio < s.minPlasmaRatio {
        log.Printf("Plasma low (%.2f%%), auto-fusing QSR", status.Ratio)
        return s.FuseQSR(client, s.defaultFusion)
    }
    
    return nil
}

func (s *Service) FuseQSR(client *zenon.Zenon, amount *big.Int) error {
    // Check QSR balance
    balance, err := s.getQSRBalance(client)
    if err != nil {
        return err
    }
    
    if balance.Cmp(amount) < 0 {
        return fmt.Errorf("insufficient QSR: have %s, need %s",
            balance.String(), amount.String())
    }
    
    template := client.Client.PlasmaApi.Fuse(client.Address(), amount)
    
    return client.Send(template)
}

func (s *Service) getQSRBalance(client *zenon.Zenon) (*big.Int, error) {
    info, err := client.Client.LedgerApi.GetAccountInfoByAddress(client.Address())
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

func (s *Service) StartAutoOptimization(clients map[string]*zenon.Zenon) {
    if !s.autoFusion {
        return
    }
    
    ticker := time.NewTicker(s.monitorInterval)
    go func() {
        defer ticker.Stop()
        
        for range ticker.C {
            for name, client := range clients {
                err := s.OptimizePlasma(client)
                if err != nil {
                    log.Printf("Auto-plasma failed for %s: %v", name, err)
                }
            }
        }
    }()
}
```

### Delegation Service

```go
// internal/delegation/service.go
package delegation

import (
    "fmt"
    "log"
    "time"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
    "github.com/zenon-network/go-zenon/rpc/api/embedded"
)

type Service struct {
    autoOptimize     bool
    optimizeInterval time.Duration
    minUptime        float64
}

type DelegationInfo struct {
    PillarName       string  `json:"pillarName"`
    Weight           string  `json:"weight"`
    RewardPercentage uint8   `json:"rewardPercentage"`
    Uptime           float64 `json:"uptime"`
}

func NewService() *Service {
    return &Service{
        autoOptimize:     true,
        optimizeInterval: 24 * time.Hour,
        minUptime:        80.0,
    }
}

func (s *Service) GetDelegationStatus(client *zenon.Zenon) (*DelegationInfo, error) {
    delegation, err := client.Client.PillarApi.GetDelegatedPillar(client.Address())
    if err != nil {
        return nil, err
    }
    
    if delegation == nil {
        return nil, fmt.Errorf("not delegating")
    }
    
    uptime := s.calculateUptime(delegation)
    
    return &DelegationInfo{
        PillarName:       delegation.Name,
        Weight:           delegation.Weight.String(),
        RewardPercentage: delegation.GiveDelegateRewardPercentage,
        Uptime:           uptime,
    }, nil
}

func (s *Service) calculateUptime(pillar *embedded.PillarInfo) float64 {
    if pillar.ExpectedMomentums == 0 {
        return 0
    }
    return float64(pillar.ProducedMomentums) / float64(pillar.ExpectedMomentums) * 100
}

func (s *Service) FindBestPillar(client *zenon.Zenon) (*embedded.PillarInfo, error) {
    pillars, err := client.Client.PillarApi.GetAll(0, 50)
    if err != nil {
        return nil, err
    }
    
    var best *embedded.PillarInfo
    bestScore := 0.0
    
    for _, pillar := range pillars.List {
        uptime := s.calculateUptime(pillar)
        
        if uptime < s.minUptime {
            continue
        }
        
        // Score based on uptime and reward sharing
        score := uptime*0.7 + float64(pillar.GiveDelegateRewardPercentage)*0.3
        
        if score > bestScore {
            best = pillar
            bestScore = score
        }
    }
    
    return best, nil
}

func (s *Service) OptimizeDelegation(client *zenon.Zenon) error {
    // Get current delegation
    current, err := s.GetDelegationStatus(client)
    if err != nil && err.Error() != "not delegating" {
        return err
    }
    
    // Find best pillar
    best, err := s.FindBestPillar(client)
    if err != nil {
        return err
    }
    
    if best == nil {
        return fmt.Errorf("no suitable pillar found")
    }
    
    // Check if we should switch
    if current != nil && current.PillarName == best.Name {
        log.Printf("Already delegating to best pillar: %s", best.Name)
        return nil
    }
    
    // Undelegate if currently delegating
    if current != nil {
        log.Printf("Undelegating from %s", current.PillarName)
        template := client.Client.PillarApi.Undelegate()
        err = client.Send(template)
        if err != nil {
            return err
        }
        
        time.Sleep(10 * time.Second) // Wait for confirmation
    }
    
    // Delegate to best pillar
    log.Printf("Delegating to %s (uptime: %.2f%%, rewards: %d%%)",
        best.Name,
        s.calculateUptime(best),
        best.GiveDelegateRewardPercentage)
    
    template := client.Client.PillarApi.Delegate(best.Name)
    
    return client.Send(template)
}

func (s *Service) StartAutoOptimization(clients map[string]*zenon.Zenon) {
    if !s.autoOptimize {
        return
    }
    
    ticker := time.NewTicker(s.optimizeInterval)
    go func() {
        defer ticker.Stop()
        
        for range ticker.C {
            for name, client := range clients {
                err := s.OptimizeDelegation(client)
                if err != nil {
                    log.Printf("Auto-delegation failed for %s: %v", name, err)
                }
            }
        }
    }()
}
```

### Monitor Service

```go
// internal/monitor/service.go
package monitor

import (
    "log"
    "math/big"
    "sync"
    "time"
    "github.com/MoonBaZZe/znn-sdk-go/zenon"
    "github.com/zenon-network/go-zenon/common/types"
)

type Service struct {
    clients    map[string]*zenon.Zenon
    snapshots  map[string][]AccountSnapshot
    mutex      sync.RWMutex
    interval   time.Duration
    maxHistory int
}

type AccountSnapshot struct {
    Timestamp time.Time          `json:"timestamp"`
    Address   string             `json:"address"`
    Balances  map[string]*big.Int `json:"balances"`
    Height    uint64             `json:"height"`
}

type BalanceChange struct {
    Address   string    `json:"address"`
    Token     string    `json:"token"`
    Previous  *big.Int  `json:"previous"`
    Current   *big.Int  `json:"current"`
    Change    *big.Int  `json:"change"`
    Timestamp time.Time `json:"timestamp"`
}

func NewService() *Service {
    return &Service{
        clients:    make(map[string]*zenon.Zenon),
        snapshots:  make(map[string][]AccountSnapshot),
        interval:   30 * time.Second,
        maxHistory: 1000,
    }
}

func (s *Service) RegisterWallet(name string, client *zenon.Zenon) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    s.clients[name] = client
    s.snapshots[name] = make([]AccountSnapshot, 0)
}

func (s *Service) UnregisterWallet(name string) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    delete(s.clients, name)
    delete(s.snapshots, name)
}

func (s *Service) Start() {
    ticker := time.NewTicker(s.interval)
    go func() {
        defer ticker.Stop()
        
        for range ticker.C {
            s.collectSnapshots()
        }
    }()
    
    log.Println("Monitor service started")
}

func (s *Service) collectSnapshots() {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    for name, client := range s.clients {
        snapshot := s.createSnapshot(client)
        if snapshot != nil {
            s.snapshots[name] = append(s.snapshots[name], *snapshot)
            
            // Trim history
            if len(s.snapshots[name]) > s.maxHistory {
                s.snapshots[name] = s.snapshots[name][1:]
            }
            
            // Check for changes
            s.checkForChanges(name)
        }
    }
}

func (s *Service) createSnapshot(client *zenon.Zenon) *AccountSnapshot {
    info, err := client.Client.LedgerApi.GetAccountInfoByAddress(client.Address())
    if err != nil {
        log.Printf("Failed to get account info: %v", err)
        return nil
    }
    
    balances := make(map[string]*big.Int)
    for _, balance := range info.BalanceInfoList {
        balances[balance.Token.Symbol] = balance.Balance
    }
    
    return &AccountSnapshot{
        Timestamp: time.Now(),
        Address:   client.Address().String(),
        Balances:  balances,
        Height:    info.AccountHeight,
    }
}

func (s *Service) checkForChanges(walletName string) {
    snapshots := s.snapshots[walletName]
    if len(snapshots) < 2 {
        return
    }
    
    current := snapshots[len(snapshots)-1]
    previous := snapshots[len(snapshots)-2]
    
    // Check for balance changes
    for token, currentBalance := range current.Balances {
        previousBalance, exists := previous.Balances[token]
        if !exists {
            previousBalance = big.NewInt(0)
        }
        
        if currentBalance.Cmp(previousBalance) != 0 {
            change := new(big.Int).Sub(currentBalance, previousBalance)
            
            s.logBalanceChange(BalanceChange{
                Address:   current.Address,
                Token:     token,
                Previous:  previousBalance,
                Current:   currentBalance,
                Change:    change,
                Timestamp: current.Timestamp,
            })
        }
    }
    
    // Check for height changes (new transactions)
    if current.Height > previous.Height {
        log.Printf("New transaction detected for %s: height %d -> %d",
            walletName, previous.Height, current.Height)
    }
}

func (s *Service) logBalanceChange(change BalanceChange) {
    direction := "received"
    if change.Change.Sign() < 0 {
        direction = "sent"
    }
    
    log.Printf("Balance change: %s %s %s %s",
        change.Address,
        direction,
        change.Change.String(),
        change.Token)
}

func (s *Service) GetHistory(walletName string, hours int) []AccountSnapshot {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    snapshots, exists := s.snapshots[walletName]
    if !exists {
        return nil
    }
    
    cutoff := time.Now().Add(time.Duration(-hours) * time.Hour)
    
    result := make([]AccountSnapshot, 0)
    for _, snapshot := range snapshots {
        if snapshot.Timestamp.After(cutoff) {
            result = append(result, snapshot)
        }
    }
    
    return result
}
```

## REST API Layer

### API Handlers

```go
// internal/api/handlers.go
package api

import (
    "encoding/json"
    "math/big"
    "net/http"
    "strconv"
    "github.com/gorilla/mux"
    "github.com/zenon-network/go-zenon/common/types"
    "zenonpay/internal/wallet"
    "zenonpay/internal/plasma"
    "zenonpay/internal/delegation"
    "zenonpay/internal/monitor"
)

type Server struct {
    walletService     *wallet.Service
    plasmaService     *plasma.Service
    delegationService *delegation.Service
    monitorService    *monitor.Service
}

type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func NewServer(ws *wallet.Service, ps *plasma.Service, ds *delegation.Service, ms *monitor.Service) *Server {
    return &Server{
        walletService:     ws,
        plasmaService:     ps,
        delegationService: ds,
        monitorService:    ms,
    }
}

func (s *Server) SetupRoutes() *mux.Router {
    router := mux.NewRouter()
    
    // Wallet endpoints
    router.HandleFunc("/api/wallets", s.listWallets).Methods("GET")
    router.HandleFunc("/api/wallets", s.createWallet).Methods("POST")
    router.HandleFunc("/api/wallets/{name}/balances", s.getBalances).Methods("GET")
    router.HandleFunc("/api/wallets/{name}/send", s.sendPayment).Methods("POST")
    router.HandleFunc("/api/wallets/{name}/plasma", s.getPlasmaStatus).Methods("GET")
    router.HandleFunc("/api/wallets/{name}/delegation", s.getDelegationStatus).Methods("GET")
    
    // Service endpoints
    router.HandleFunc("/api/optimize/plasma/{name}", s.optimizePlasma).Methods("POST")
    router.HandleFunc("/api/optimize/delegation/{name}", s.optimizeDelegation).Methods("POST")
    router.HandleFunc("/api/monitor/{name}", s.getMonitorHistory).Methods("GET")
    
    // Static files
    router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static/")))
    
    return router
}

func (s *Server) listWallets(w http.ResponseWriter, r *http.Request) {
    wallets := s.walletService.ListWallets()
    s.respond(w, Response{Success: true, Data: wallets})
}

func (s *Server) createWallet(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name     string `json:"name"`
        Password string `json:"password"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.respondError(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    wallet, err := s.walletService.CreateWallet(req.Name, req.Password)
    if err != nil {
        s.respondError(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Register with monitor
    s.monitorService.RegisterWallet(req.Name, wallet.Client)
    
    s.respond(w, Response{
        Success: true,
        Data: map[string]interface{}{
            "name":    wallet.Name,
            "address": wallet.Address.String(),
        },
    })
}

func (s *Server) getBalances(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    name := vars["name"]
    
    balances, err := s.walletService.GetBalances(name)
    if err != nil {
        s.respondError(w, err.Error(), http.StatusNotFound)
        return
    }
    
    s.respond(w, Response{Success: true, Data: balances})
}

func (s *Server) sendPayment(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    name := vars["name"]
    
    var req struct {
        To     string `json:"to"`
        Token  string `json:"token"`
        Amount string `json:"amount"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.respondError(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Parse parameters
    toAddr, err := types.ParseAddress(req.To)
    if err != nil {
        s.respondError(w, "Invalid address", http.StatusBadRequest)
        return
    }
    
    tokenStd, err := types.ParseZTS(req.Token)
    if err != nil {
        s.respondError(w, "Invalid token", http.StatusBadRequest)
        return
    }
    
    amount, ok := new(big.Int).SetString(req.Amount, 10)
    if !ok {
        s.respondError(w, "Invalid amount", http.StatusBadRequest)
        return
    }
    
    // Send payment
    err = s.walletService.SendPayment(name, toAddr, tokenStd, amount)
    if err != nil {
        s.respondError(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    s.respond(w, Response{Success: true, Data: "Payment sent"})
}

func (s *Server) getPlasmaStatus(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    name := vars["name"]
    
    wallet, err := s.walletService.GetWallet(name)
    if err != nil {
        s.respondError(w, err.Error(), http.StatusNotFound)
        return
    }
    
    status, err := s.plasmaService.GetPlasmaStatus(wallet.Client)
    if err != nil {
        s.respondError(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    s.respond(w, Response{Success: true, Data: status})
}

func (s *Server) optimizePlasma(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    name := vars["name"]
    
    wallet, err := s.walletService.GetWallet(name)
    if err != nil {
        s.respondError(w, err.Error(), http.StatusNotFound)
        return
    }
    
    err = s.plasmaService.OptimizePlasma(wallet.Client)
    if err != nil {
        s.respondError(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    s.respond(w, Response{Success: true, Data: "Plasma optimized"})
}

func (s *Server) getDelegationStatus(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    name := vars["name"]
    
    wallet, err := s.walletService.GetWallet(name)
    if err != nil {
        s.respondError(w, err.Error(), http.StatusNotFound)
        return
    }
    
    status, err := s.delegationService.GetDelegationStatus(wallet.Client)
    if err != nil {
        s.respondError(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    s.respond(w, Response{Success: true, Data: status})
}

func (s *Server) optimizeDelegation(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    name := vars["name"]
    
    wallet, err := s.walletService.GetWallet(name)
    if err != nil {
        s.respondError(w, err.Error(), http.StatusNotFound)
        return
    }
    
    err = s.delegationService.OptimizeDelegation(wallet.Client)
    if err != nil {
        s.respondError(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    s.respond(w, Response{Success: true, Data: "Delegation optimized"})
}

func (s *Server) getMonitorHistory(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    name := vars["name"]
    
    hoursStr := r.URL.Query().Get("hours")
    hours := 24 // default
    if hoursStr != "" {
        if h, err := strconv.Atoi(hoursStr); err == nil {
            hours = h
        }
    }
    
    history := s.monitorService.GetHistory(name, hours)
    s.respond(w, Response{Success: true, Data: history})
}

func (s *Server) respond(w http.ResponseWriter, response Response) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (s *Server) respondError(w http.ResponseWriter, message string, code int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(Response{
        Success: false,
        Error:   message,
    })
}
```

## Web Interface

### HTML Template

```html
<!-- web/templates/dashboard.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ZenonPay Dashboard</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <style>
        .wallet-card { margin-bottom: 20px; }
        .balance-item { padding: 10px; border-left: 3px solid #007bff; margin: 5px 0; }
        .plasma-excellent { border-left-color: #28a745; }
        .plasma-good { border-left-color: #ffc107; }
        .plasma-low { border-left-color: #fd7e14; }
        .plasma-critical { border-left-color: #dc3545; }
        .status-badge { margin-left: 10px; }
    </style>
</head>
<body>
    <nav class="navbar navbar-dark bg-dark">
        <div class="container">
            <span class="navbar-brand">⚡ ZenonPay Dashboard</span>
            <button class="btn btn-outline-light" onclick="loadWallets()">Refresh</button>
        </div>
    </nav>
    
    <div class="container mt-4">
        <div class="row">
            <div class="col-md-4">
                <div class="card">
                    <div class="card-header">
                        <h5>Create Wallet</h5>
                    </div>
                    <div class="card-body">
                        <form id="createWalletForm">
                            <div class="mb-3">
                                <input type="text" class="form-control" id="walletName" placeholder="Wallet Name" required>
                            </div>
                            <div class="mb-3">
                                <input type="password" class="form-control" id="walletPassword" placeholder="Password" required>
                            </div>
                            <button type="submit" class="btn btn-primary">Create Wallet</button>
                        </form>
                    </div>
                </div>
                
                <div class="card mt-3">
                    <div class="card-header">
                        <h5>Send Payment</h5>
                    </div>
                    <div class="card-body">
                        <form id="sendPaymentForm">
                            <div class="mb-3">
                                <select class="form-control" id="fromWallet" required>
                                    <option value="">Select Wallet</option>
                                </select>
                            </div>
                            <div class="mb-3">
                                <input type="text" class="form-control" id="toAddress" placeholder="Recipient Address" required>
                            </div>
                            <div class="mb-3">
                                <select class="form-control" id="tokenType" required>
                                    <option value="zts1znnxxxxxxxxxxxxx9z4ulx">ZNN</option>
                                    <option value="zts1qsrxxxxxxxxxxxxxmrhjll">QSR</option>
                                </select>
                            </div>
                            <div class="mb-3">
                                <input type="text" class="form-control" id="amount" placeholder="Amount" required>
                            </div>
                            <button type="submit" class="btn btn-success">Send Payment</button>
                        </form>
                    </div>
                </div>
            </div>
            
            <div class="col-md-8">
                <div id="walletsContainer">
                    <div class="text-center">
                        <button class="btn btn-outline-primary" onclick="loadWallets()">Load Wallets</button>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        // API functions
        async function apiCall(url, options = {}) {
            try {
                const response = await fetch(url, {
                    headers: { 'Content-Type': 'application/json' },
                    ...options
                });
                return await response.json();
            } catch (error) {
                console.error('API call failed:', error);
                return { success: false, error: error.message };
            }
        }

        // Load and display wallets
        async function loadWallets() {
            const result = await apiCall('/api/wallets');
            if (!result.success) {
                alert('Failed to load wallets: ' + result.error);
                return;
            }

            const container = document.getElementById('walletsContainer');
            container.innerHTML = '';

            if (result.data.length === 0) {
                container.innerHTML = '<div class="alert alert-info">No wallets found. Create one to get started!</div>';
                return;
            }

            // Update wallet selector
            const selector = document.getElementById('fromWallet');
            selector.innerHTML = '<option value="">Select Wallet</option>';
            
            for (const walletName of result.data) {
                // Add to selector
                const option = document.createElement('option');
                option.value = walletName;
                option.textContent = walletName;
                selector.appendChild(option);
                
                // Create wallet card
                const card = await createWalletCard(walletName);
                container.appendChild(card);
            }
        }

        async function createWalletCard(walletName) {
            const card = document.createElement('div');
            card.className = 'card wallet-card';
            
            // Get wallet data
            const [balances, plasma, delegation] = await Promise.all([
                apiCall(`/api/wallets/${walletName}/balances`),
                apiCall(`/api/wallets/${walletName}/plasma`),
                apiCall(`/api/wallets/${walletName}/delegation`)
            ]);

            let balanceHtml = '';
            if (balances.success) {
                balanceHtml = balances.data.map(b => 
                    `<div class="balance-item">
                        <strong>${b.symbol}:</strong> ${formatAmount(b.amount)}
                    </div>`
                ).join('');
            }

            let plasmaHtml = '';
            if (plasma.success) {
                const statusClass = `plasma-${plasma.data.status.toLowerCase()}`;
                plasmaHtml = `
                    <div class="balance-item ${statusClass}">
                        <strong>Plasma:</strong> ${plasma.data.currentPlasma}/${plasma.data.maxPlasma}
                        <span class="badge bg-secondary status-badge">${plasma.data.status}</span>
                        <button class="btn btn-sm btn-outline-primary ms-2" onclick="optimizePlasma('${walletName}')">Optimize</button>
                    </div>
                `;
            }

            let delegationHtml = '';
            if (delegation.success) {
                delegationHtml = `
                    <div class="balance-item">
                        <strong>Delegated to:</strong> ${delegation.data.pillarName}
                        <span class="badge bg-info status-badge">${delegation.data.uptime.toFixed(1)}% uptime</span>
                        <button class="btn btn-sm btn-outline-primary ms-2" onclick="optimizeDelegation('${walletName}')">Optimize</button>
                    </div>
                `;
            }

            card.innerHTML = `
                <div class="card-header">
                    <h6>${walletName}</h6>
                </div>
                <div class="card-body">
                    ${balanceHtml}
                    ${plasmaHtml}
                    ${delegationHtml}
                </div>
            `;

            return card;
        }

        function formatAmount(amount) {
            // Simple formatting - convert from smallest unit
            const val = parseFloat(amount) / 100000000;
            return val.toFixed(8);
        }

        // Form handlers
        document.getElementById('createWalletForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const name = document.getElementById('walletName').value;
            const password = document.getElementById('walletPassword').value;
            
            const result = await apiCall('/api/wallets', {
                method: 'POST',
                body: JSON.stringify({ name, password })
            });
            
            if (result.success) {
                alert('Wallet created successfully!');
                document.getElementById('createWalletForm').reset();
                loadWallets();
            } else {
                alert('Failed to create wallet: ' + result.error);
            }
        });

        document.getElementById('sendPaymentForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const wallet = document.getElementById('fromWallet').value;
            const to = document.getElementById('toAddress').value;
            const token = document.getElementById('tokenType').value;
            const amount = document.getElementById('amount').value;
            
            // Convert amount to smallest unit
            const amountWei = (parseFloat(amount) * 100000000).toString();
            
            const result = await apiCall(`/api/wallets/${wallet}/send`, {
                method: 'POST',
                body: JSON.stringify({ to, token, amount: amountWei })
            });
            
            if (result.success) {
                alert('Payment sent successfully!');
                document.getElementById('sendPaymentForm').reset();
                setTimeout(loadWallets, 2000); // Refresh after delay
            } else {
                alert('Failed to send payment: ' + result.error);
            }
        });

        async function optimizePlasma(walletName) {
            const result = await apiCall(`/api/optimize/plasma/${walletName}`, {
                method: 'POST'
            });
            
            if (result.success) {
                alert('Plasma optimization started!');
                setTimeout(loadWallets, 3000);
            } else {
                alert('Failed to optimize plasma: ' + result.error);
            }
        }

        async function optimizeDelegation(walletName) {
            const result = await apiCall(`/api/optimize/delegation/${walletName}`, {
                method: 'POST'
            });
            
            if (result.success) {
                alert('Delegation optimization started!');
                setTimeout(loadWallets, 5000);
            } else {
                alert('Failed to optimize delegation: ' + result.error);
            }
        }

        // Auto-refresh
        setInterval(loadWallets, 30000); // Refresh every 30 seconds
        
        // Initial load
        loadWallets();
    </script>
</body>
</html>
```

## Main Application

### Server Entry Point

```go
// cmd/server/main.go
package main

import (
    "fmt"
    "log"
    "net/http"
    "zenonpay/config"
    "zenonpay/internal/api"
    "zenonpay/internal/wallet"
    "zenonpay/internal/plasma"
    "zenonpay/internal/delegation"
    "zenonpay/internal/monitor"
)

func main() {
    // Load configuration
    cfg, err := config.LoadConfig("config.json")
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // Initialize services
    walletService := wallet.NewService(cfg.Zenon.NodeURL)
    plasmaService := plasma.NewService()
    delegationService := delegation.NewService()
    monitorService := monitor.NewService()
    
    // Start monitoring
    monitorService.Start()
    
    // Setup API server
    apiServer := api.NewServer(walletService, plasmaService, delegationService, monitorService)
    router := apiServer.SetupRoutes()
    
    // Start HTTP server
    addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
    log.Printf("ZenonPay server starting on %s", addr)
    
    if err := http.ListenAndServe(addr, router); err != nil {
        log.Fatal("Server failed:", err)
    }
}
```

### CLI Tool

```go
// cmd/cli/main.go
package main

import (
    "flag"
    "fmt"
    "log"
    "math/big"
    "os"
    "zenonpay/internal/wallet"
    "github.com/zenon-network/go-zenon/common/types"
)

func main() {
    var (
        nodeURL = flag.String("node", "ws://127.0.0.1:35998", "Zenon node URL")
        action  = flag.String("action", "", "Action: create, list, send, balance")
        name    = flag.String("name", "", "Wallet name")
        password = flag.String("password", "", "Wallet password")
        to      = flag.String("to", "", "Recipient address")
        token   = flag.String("token", "zts1znnxxxxxxxxxxxxx9z4ulx", "Token standard")
        amount  = flag.String("amount", "", "Amount to send")
    )
    flag.Parse()

    service := wallet.NewService(*nodeURL)
    defer service.Shutdown()

    switch *action {
    case "create":
        if *name == "" || *password == "" {
            log.Fatal("Name and password required for create")
        }
        
        wallet, err := service.CreateWallet(*name, *password)
        if err != nil {
            log.Fatal("Failed to create wallet:", err)
        }
        
        fmt.Printf("Wallet created: %s\n", wallet.Address.String())

    case "list":
        wallets := service.ListWallets()
        if len(wallets) == 0 {
            fmt.Println("No wallets loaded")
            return
        }
        
        fmt.Println("Loaded wallets:")
        for _, name := range wallets {
            fmt.Printf("- %s\n", name)
        }

    case "balance":
        if *name == "" {
            log.Fatal("Wallet name required")
        }
        
        balances, err := service.GetBalances(*name)
        if err != nil {
            log.Fatal("Failed to get balances:", err)
        }
        
        fmt.Printf("Balances for %s:\n", *name)
        for _, balance := range balances {
            fmt.Printf("- %s: %s\n", balance.Symbol, balance.Amount.String())
        }

    case "send":
        if *name == "" || *to == "" || *amount == "" {
            log.Fatal("Name, to address, and amount required for send")
        }
        
        toAddr, err := types.ParseAddress(*to)
        if err != nil {
            log.Fatal("Invalid address:", err)
        }
        
        tokenStd, err := types.ParseZTS(*token)
        if err != nil {
            log.Fatal("Invalid token:", err)
        }
        
        amt, ok := new(big.Int).SetString(*amount, 10)
        if !ok {
            log.Fatal("Invalid amount")
        }
        
        err = service.SendPayment(*name, toAddr, tokenStd, amt)
        if err != nil {
            log.Fatal("Failed to send payment:", err)
        }
        
        fmt.Println("Payment sent successfully!")

    default:
        fmt.Println("Available actions: create, list, balance, send")
        flag.Usage()
        os.Exit(1)
    }
}
```

## Configuration

### Sample Config

```json
{
  "server": {
    "host": "localhost",
    "port": 8080
  },
  "zenon": {
    "nodeUrl": "ws://127.0.0.1:35998"
  },
  "wallets": {
    "directory": "./wallets"
  },
  "features": {
    "autoPlasma": true,
    "autoDelegation": true,
    "monitoring": true
  }
}
```

## Deployment

### Docker Configuration

```dockerfile
# Dockerfile
FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o zenonpay ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/zenonpay .
COPY --from=builder /app/web ./web
COPY --from=builder /app/config.json .

EXPOSE 8080
CMD ["./zenonpay"]
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  zenonpay:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./wallets:/root/wallets
      - ./config.json:/root/config.json
    restart: unless-stopped
```

## Testing

### Integration Tests

```go
// tests/integration_test.go
package tests

import (
    "testing"
    "time"
    "zenonpay/internal/wallet"
    "zenonpay/internal/plasma"
    "zenonpay/internal/delegation"
)

func TestFullWorkflow(t *testing.T) {
    // Setup services
    walletService := wallet.NewService("ws://testnet:35998")
    plasmaService := plasma.NewService()
    delegationService := delegation.NewService()
    
    // Create test wallet
    wallet, err := walletService.CreateWallet("test-wallet", "password123")
    if err != nil {
        t.Fatal("Failed to create wallet:", err)
    }
    
    // Test plasma optimization
    err = plasmaService.OptimizePlasma(wallet.Client)
    if err != nil {
        t.Log("Plasma optimization failed (expected for new wallet):", err)
    }
    
    // Test delegation optimization
    err = delegationService.OptimizeDelegation(wallet.Client)
    if err != nil {
        t.Log("Delegation optimization failed:", err)
    }
    
    // Cleanup
    walletService.CloseWallet("test-wallet")
}
```

## Summary

You've built a complete DApp that demonstrates:
- ✅ Multi-wallet management with secure storage
- ✅ Automatic Plasma optimization for feeless transactions
- ✅ Intelligent Pillar delegation with performance tracking
- ✅ Real-time balance and transaction monitoring
- ✅ RESTful API with comprehensive endpoints
- ✅ Modern web interface with real-time updates
- ✅ Modular architecture with separation of concerns
- ✅ CLI tools for power users
- ✅ Docker deployment configuration

This project showcases advanced Zenon Network development patterns and serves as a foundation for building sophisticated DApps on the Network of Momentum.

## Next Steps

- Add multi-signature wallet support
- Implement token swap functionality
- Create mobile app interface
- Add advanced analytics and reporting
- Integrate with hardware wallets
- Build automated trading strategies
- Add support for embedded contracts