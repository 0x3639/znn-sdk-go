package rpc_client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/0x3639/znn-sdk-go/api"
	"github.com/0x3639/znn-sdk-go/api/embedded"

	"github.com/zenon-network/go-zenon/rpc/server"
)

// ConnectionEstablishedCallback is called when connection is established or re-established
type ConnectionEstablishedCallback func()

// ConnectionLostCallback is called when connection is lost
type ConnectionLostCallback func(err error)

// RpcClient wraps go-zenon's RPC client with connection management
type RpcClient struct {
	// Connection management
	client     *server.Client
	url        string
	status     WebsocketStatus
	statusLock sync.RWMutex

	// Auto-reconnect configuration
	autoReconnect      bool
	reconnectDelay     time.Duration
	maxReconnectDelay  time.Duration
	reconnectAttempts  int
	currentAttempt     int
	stopReconnectChan  chan struct{}
	reconnectCtx       context.Context
	reconnectCtxCancel context.CancelFunc
	reconnectLock      sync.Mutex // Prevents concurrent reconnection attempts

	// Callbacks
	onConnectionEstablished []ConnectionEstablishedCallback
	onConnectionLost        []ConnectionLostCallback
	callbackLock            sync.RWMutex

	// Monitoring
	monitorTicker  *time.Ticker
	monitorCtx     context.Context
	monitorCancel  context.CancelFunc
	healthCheckCmd string

	// API lock protects API field reassignment during reconnection
	apiLock sync.RWMutex

	// Embedded contract APIs
	AcceleratorApi *embedded.AcceleratorApi
	PillarApi      *embedded.PillarApi
	PlasmaApi      *embedded.PlasmaApi
	SentinelApi    *embedded.SentinelApi
	SporkApi       *embedded.SporkApi
	StakeApi       *embedded.StakeApi
	SwapApi        *embedded.SwapApi
	TokenApi       *embedded.TokenApi
	BridgeApi      *embedded.BridgeApi
	LiquidityApi   *embedded.LiquidityApi
	HtlcApi        *embedded.HtlcApi

	// Ledger & Stats APIs
	LedgerApi     *api.LedgerApi
	StatsApi      *api.StatsApi
	SubscriberApi *api.SubscriberApi
}

// ClientOptions configures RpcClient behavior
type ClientOptions struct {
	// AutoReconnect enables automatic reconnection on connection loss
	AutoReconnect bool
	// ReconnectDelay is the initial delay between reconnect attempts (default: 1s)
	ReconnectDelay time.Duration
	// MaxReconnectDelay is the maximum delay between reconnect attempts (default: 30s)
	MaxReconnectDelay time.Duration
	// ReconnectAttempts is the maximum number of reconnect attempts (0 = infinite)
	ReconnectAttempts int
	// HealthCheckInterval is the interval for connection health checks (default: 30s, 0 = disabled)
	HealthCheckInterval time.Duration
	// HealthCheckCommand is the RPC command to use for health checks (default: "ledger.getFrontierMomentum")
	HealthCheckCommand string
}

// DefaultClientOptions returns default client options
func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		AutoReconnect:       true,
		ReconnectDelay:      1 * time.Second,
		MaxReconnectDelay:   30 * time.Second,
		ReconnectAttempts:   0, // infinite
		HealthCheckInterval: 30 * time.Second,
		HealthCheckCommand:  "ledger.getFrontierMomentum",
	}
}

// NewRpcClient creates a new RPC client connected to a Zenon node with default options.
//
// This is the main entry point for the SDK. It establishes a WebSocket connection to the
// specified node URL and initializes all API endpoints (Ledger, Stats, Subscriber, and
// all embedded contract APIs).
//
// Default options include:
//   - Auto-reconnect enabled with exponential backoff
//   - Health checks every 30 seconds
//   - Infinite reconnection attempts
//
// Parameters:
//   - url: WebSocket URL of the Zenon node (e.g., "ws://127.0.0.1:35998")
//
// Returns an initialized RpcClient ready to use, or an error if connection fails.
//
// Example:
//
//	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Stop()
//
//	// Use the client
//	momentum, _ := client.LedgerApi.GetFrontierMomentum()
//	fmt.Printf("Height: %d\n", momentum.Height)
//
// For custom configuration (e.g., disable auto-reconnect, custom health check intervals),
// use NewRpcClientWithOptions instead.
func NewRpcClient(url string) (*RpcClient, error) {
	return NewRpcClientWithOptions(url, DefaultClientOptions())
}

// NewRpcClientWithOptions creates a new RPC client with custom configuration options.
//
// This allows fine-grained control over connection behavior including auto-reconnection,
// health checks, and retry policies.
//
// Parameters:
//   - url: WebSocket URL of the Zenon node (e.g., "ws://127.0.0.1:35998")
//   - opts: ClientOptions struct configuring connection behavior
//
// Available options:
//   - AutoReconnect: Enable automatic reconnection on connection loss (default: true)
//   - ReconnectDelay: Initial delay between reconnect attempts (default: 1s)
//   - MaxReconnectDelay: Maximum delay with exponential backoff (default: 30s)
//   - ReconnectAttempts: Max reconnection attempts, 0 for infinite (default: 0)
//   - HealthCheckInterval: Interval for connection health checks (default: 30s, 0 to disable)
//   - HealthCheckCommand: RPC command for health checks (default: "ledger.getFrontierMomentum")
//
// Returns an initialized RpcClient or an error if the initial connection fails.
//
// Example with custom options:
//
//	opts := rpc_client.ClientOptions{
//	    AutoReconnect:       true,
//	    ReconnectDelay:      2 * time.Second,
//	    MaxReconnectDelay:   60 * time.Second,
//	    ReconnectAttempts:   10,  // Give up after 10 attempts
//	    HealthCheckInterval: 15 * time.Second,
//	}
//	client, err := rpc_client.NewRpcClientWithOptions("ws://127.0.0.1:35998", opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Stop()
//
// The client will validate and normalize the WebSocket URL automatically.
func NewRpcClientWithOptions(url string, opts ClientOptions) (*RpcClient, error) {
	if err := ValidateWsConnectionURL(url); err != nil {
		return nil, fmt.Errorf("invalid WebSocket URL: %w", err)
	}

	normalized, err := NormalizeWsURL(url)
	if err != nil {
		return nil, err
	}

	c := &RpcClient{
		url:                     normalized,
		status:                  Uninitialized,
		autoReconnect:           opts.AutoReconnect,
		reconnectDelay:          opts.ReconnectDelay,
		maxReconnectDelay:       opts.MaxReconnectDelay,
		reconnectAttempts:       opts.ReconnectAttempts,
		stopReconnectChan:       make(chan struct{}),
		onConnectionEstablished: make([]ConnectionEstablishedCallback, 0),
		onConnectionLost:        make([]ConnectionLostCallback, 0),
		healthCheckCmd:          opts.HealthCheckCommand,
	}

	// Connect initially
	if err := c.connect(); err != nil {
		return nil, err
	}

	// Start monitoring if health check is enabled
	if opts.HealthCheckInterval > 0 {
		c.startMonitoring(opts.HealthCheckInterval)
	}

	return c, nil
}

// connect establishes the WebSocket connection and initializes APIs
func (c *RpcClient) connect() error {
	c.setStatus(Connecting)

	client, err := server.Dial(c.url)
	if err != nil {
		c.setStatus(Stopped)
		return fmt.Errorf("failed to connect to %s: %w", c.url, err)
	}

	c.client = client
	c.initializeAPIs()
	c.setStatus(Running)
	c.currentAttempt = 0

	// Trigger connection established callbacks
	c.triggerConnectionEstablished()

	return nil
}

// initializeAPIs creates all API instances with thread-safe locking
func (c *RpcClient) initializeAPIs() {
	c.apiLock.Lock()
	defer c.apiLock.Unlock()

	c.AcceleratorApi = embedded.NewAcceleratorApi(c.client)
	c.BridgeApi = embedded.NewBridgeApi(c.client)
	c.PillarApi = embedded.NewPillarApi(c.client)
	c.PlasmaApi = embedded.NewPlasmaApi(c.client)
	c.SentinelApi = embedded.NewSentinelApi(c.client)
	c.SporkApi = embedded.NewSporkApi(c.client)
	c.StakeApi = embedded.NewStakeApi(c.client)
	c.SwapApi = embedded.NewSwapApi(c.client)
	c.TokenApi = embedded.NewTokenApi(c.client)
	c.LiquidityApi = embedded.NewLiquidityApi(c.client)
	c.HtlcApi = embedded.NewHtlcApi(c.client)
	c.LedgerApi = api.NewLedgerApi(c.client)
	c.StatsApi = api.NewStatsApi(c.client)
	c.SubscriberApi = api.NewSubscriberApi(c.client)
}

// Status returns the current WebSocket connection status.
//
// Possible statuses:
//   - Uninitialized: Client created but not yet connected
//   - Connecting: Connection attempt in progress
//   - Running: Successfully connected and operational
//   - Stopped: Connection closed or failed
//
// This method is thread-safe and can be called from any goroutine.
//
// Example:
//
//	status := client.Status()
//	if status == rpc_client.Running {
//	    // Connection is healthy
//	} else {
//	    // Handle connection issue
//	}
func (c *RpcClient) Status() WebsocketStatus {
	c.statusLock.RLock()
	defer c.statusLock.RUnlock()
	return c.status
}

// setStatus updates the connection status
func (c *RpcClient) setStatus(status WebsocketStatus) {
	c.statusLock.Lock()
	defer c.statusLock.Unlock()
	c.status = status
}

// IsClosed returns true if the connection is closed
func (c *RpcClient) IsClosed() bool {
	return c.Status() == Stopped
}

// AddOnConnectionEstablishedCallback registers a callback function that will be called
// when the WebSocket connection is successfully established or re-established.
//
// This is useful for:
//   - Logging connection events
//   - Reinitializing state after reconnection
//   - Resubscribing to blockchain events
//   - Notifying other parts of your application
//
// Multiple callbacks can be registered and will be called in registration order.
// Callbacks are executed in separate goroutines to prevent blocking.
//
// Parameters:
//   - callback: Function to call when connection is established (no parameters)
//
// Example:
//
//	client.AddOnConnectionEstablishedCallback(func() {
//	    fmt.Println("Connected to Zenon node")
//	    // Reinitialize subscriptions or state
//	})
func (c *RpcClient) AddOnConnectionEstablishedCallback(callback ConnectionEstablishedCallback) {
	c.callbackLock.Lock()
	defer c.callbackLock.Unlock()
	c.onConnectionEstablished = append(c.onConnectionEstablished, callback)
}

// AddOnConnectionLostCallback registers a callback function that will be called
// when the WebSocket connection is lost or fails.
//
// This is useful for:
//   - Logging disconnection events
//   - Alerting monitoring systems
//   - Implementing custom reconnection logic
//   - Cleaning up resources or state
//
// Multiple callbacks can be registered and will be called in registration order.
// Callbacks are executed in separate goroutines to prevent blocking.
//
// If auto-reconnect is enabled, the client will attempt to reconnect automatically
// after calling these callbacks.
//
// Parameters:
//   - callback: Function to call when connection is lost (receives error describing the failure)
//
// Example:
//
//	client.AddOnConnectionLostCallback(func(err error) {
//	    log.Printf("Connection lost: %v", err)
//	    // Clean up subscriptions or notify application
//	})
func (c *RpcClient) AddOnConnectionLostCallback(callback ConnectionLostCallback) {
	c.callbackLock.Lock()
	defer c.callbackLock.Unlock()
	c.onConnectionLost = append(c.onConnectionLost, callback)
}

// triggerConnectionEstablished calls all connection established callbacks with panic recovery
func (c *RpcClient) triggerConnectionEstablished() {
	c.callbackLock.RLock()
	callbacks := make([]ConnectionEstablishedCallback, len(c.onConnectionEstablished))
	copy(callbacks, c.onConnectionEstablished)
	c.callbackLock.RUnlock()

	for _, callback := range callbacks {
		go func(cb ConnectionEstablishedCallback) {
			defer func() {
				if r := recover(); r != nil {
					// Log or handle panic in callback
					fmt.Printf("Panic in connection established callback: %v\n", r)
				}
			}()
			cb()
		}(callback)
	}
}

// triggerConnectionLost calls all connection lost callbacks with panic recovery
func (c *RpcClient) triggerConnectionLost(err error) {
	c.callbackLock.RLock()
	callbacks := make([]ConnectionLostCallback, len(c.onConnectionLost))
	copy(callbacks, c.onConnectionLost)
	c.callbackLock.RUnlock()

	for _, callback := range callbacks {
		go func(cb ConnectionLostCallback, e error) {
			defer func() {
				if r := recover(); r != nil {
					// Log or handle panic in callback
					fmt.Printf("Panic in connection lost callback: %v\n", r)
				}
			}()
			cb(e)
		}(callback, err)
	}
}

// startMonitoring starts the connection health check monitor
func (c *RpcClient) startMonitoring(interval time.Duration) {
	c.monitorCtx, c.monitorCancel = context.WithCancel(context.Background())
	c.monitorTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-c.monitorTicker.C:
				c.performHealthCheck()
			case <-c.monitorCtx.Done():
				return
			}
		}
	}()
}

// performHealthCheck checks if the connection is healthy
func (c *RpcClient) performHealthCheck() {
	if c.IsClosed() {
		return
	}

	// Try a simple RPC call
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result interface{}
	err := c.client.CallContext(ctx, &result, c.healthCheckCmd)
	if err != nil {
		// Connection appears to be lost
		c.handleConnectionLoss(fmt.Errorf("health check failed: %w", err))
	}
}

// handleConnectionLoss handles a detected connection loss
func (c *RpcClient) handleConnectionLoss(err error) {
	if c.IsClosed() {
		return
	}

	c.setStatus(Stopped)

	// Close the old client
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}

	// Trigger connection lost callbacks
	c.triggerConnectionLost(err)

	// Start reconnection if enabled
	if c.autoReconnect {
		go c.startReconnect()
	}
}

// startReconnect attempts to reconnect with exponential backoff
// Uses reconnectLock to prevent concurrent reconnection attempts
func (c *RpcClient) startReconnect() {
	// Try to acquire lock; if already reconnecting, return
	if !c.reconnectLock.TryLock() {
		return
	}
	defer c.reconnectLock.Unlock()

	c.reconnectCtx, c.reconnectCtxCancel = context.WithCancel(context.Background())
	defer c.reconnectCtxCancel()

	delay := c.reconnectDelay
	c.currentAttempt = 0

	for {
		select {
		case <-c.stopReconnectChan:
			return
		case <-c.reconnectCtx.Done():
			return
		default:
		}

		// Check if we've exceeded max attempts
		if c.reconnectAttempts > 0 && c.currentAttempt >= c.reconnectAttempts {
			return
		}

		c.currentAttempt++

		// Attempt to reconnect
		if err := c.connect(); err == nil {
			// Successfully reconnected
			return
		}

		// Wait before next attempt with exponential backoff
		time.Sleep(delay)
		delay *= 2
		if delay > c.maxReconnectDelay {
			delay = c.maxReconnectDelay
		}
	}
}

// Restart manually triggers a reconnection
func (c *RpcClient) Restart() error {
	c.Stop()
	time.Sleep(100 * time.Millisecond) // Brief delay
	return c.connect()
}

// Stop gracefully shuts down the RPC client, closing the WebSocket connection
// and stopping all background tasks.
//
// This method:
//   - Closes the WebSocket connection
//   - Stops health check monitoring
//   - Cancels any ongoing reconnection attempts
//   - Cleans up all resources
//
// After calling Stop(), the client cannot be reused. Create a new client if you need
// to reconnect.
//
// This method is idempotent - calling it multiple times is safe.
// It's recommended to use defer for proper cleanup:
//
//	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Stop()
//
//	// Use client...
//
// Note: This method does not trigger connection lost callbacks since it's an
// intentional shutdown rather than a connection failure.
func (c *RpcClient) Stop() {
	c.setStatus(Stopped)

	// Stop monitoring
	if c.monitorCancel != nil {
		c.monitorCancel()
	}
	if c.monitorTicker != nil {
		c.monitorTicker.Stop()
	}

	// Stop reconnection
	if c.reconnectCtxCancel != nil {
		c.reconnectCtxCancel()
	}
	select {
	case c.stopReconnectChan <- struct{}{}:
	default:
	}

	// Close client
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
}
