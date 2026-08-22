package rpc_client

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/0x3639/znn-sdk-go/api"
	"github.com/0x3639/znn-sdk-go/api/embedded"
	"github.com/0x3639/znn-sdk-go/transport"

	"github.com/zenon-network/go-zenon/rpc/server"
)

// ConnectionEstablishedCallback is called when connection is established or re-established
type ConnectionEstablishedCallback func()

// ConnectionLostCallback is called when connection is lost
type ConnectionLostCallback func(err error)

// RpcClient wraps go-zenon's RPC client with connection management
type RpcClient struct {
	// Connection management. lifecycleLock guards the mutable connection
	// lifecycle fields below (client, caller, currentAttempt, reconnectCtx,
	// reconnectCtxCancel, stopped), which are accessed from the constructor,
	// the reconnect goroutine, the health monitor, and Stop/Restart.
	lifecycleLock sync.Mutex
	client        *server.Client
	caller        *transport.NormalizingCaller
	stopped       bool // latched by Stop(); prevents reconnect resurrection
	url           string
	status        WebsocketStatus
	statusLock    sync.RWMutex

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

	// Normalized subscriptions created through Subscribe.
	subscriptionLock sync.Mutex
	subscriptions    map[*NormalizedSubscription]struct{}

	// Transport hardening (see ClientOptions).
	dialPolicy                  DialPolicy
	maxHTTPResponseBytes        int64
	httpTimeout                 time.Duration
	maxSubscriptionMessageBytes int64

	// Monitoring
	monitorTicker  *time.Ticker
	monitorCtx     context.Context
	monitorCancel  context.CancelFunc
	healthCheckCmd string

	// apiCaller is the stable caller shared by all exported API objects. The
	// API objects are created once (apiInitOnce) and never reassigned; a
	// reconnect only swaps this caller's inner target, so callers can read the
	// exported API fields concurrently with a reconnect without a data race.
	apiCaller   *swappableCaller
	apiInitOnce sync.Once

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

	// DialPolicy authorizes every resolved destination the client connects to
	// (initial dial, reconnects, subscription sockets, HTTP redirects). nil
	// allows all destinations, which is appropriate for operator-configured
	// node URLs. Set RejectNonPublicDestinations when the URL comes from an
	// untrusted source (CWE-918).
	DialPolicy DialPolicy

	// MaxHTTPResponseBytes bounds each HTTP(S) JSON-RPC response body before
	// it is decoded or buffered (default: DefaultMaxHTTPResponseBytes; <= 0
	// uses the default).
	MaxHTTPResponseBytes int64

	// HTTPTimeout is the end-to-end timeout for one HTTP JSON-RPC call
	// (default: DefaultHTTPTimeout; <= 0 uses the default).
	HTTPTimeout time.Duration

	// MaxSubscriptionMessageBytes bounds each frame read on the dedicated
	// subscription WebSocket created by Subscribe (default:
	// DefaultMaxSubscriptionMessageBytes; <= 0 uses the default).
	MaxSubscriptionMessageBytes int64
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

		MaxHTTPResponseBytes:        DefaultMaxHTTPResponseBytes,
		HTTPTimeout:                 DefaultHTTPTimeout,
		MaxSubscriptionMessageBytes: DefaultMaxSubscriptionMessageBytes,
	}
}

// NewRpcClient creates a new RPC client connected to a Zenon node with default options.
//
// This is the main entry point for the SDK. It establishes an HTTP or WebSocket connection to the
// specified node URL and initializes all API endpoints (Ledger, Stats, Subscriber, and
// all embedded contract APIs).
//
// Default options include:
//   - Auto-reconnect enabled with exponential backoff
//   - Health checks every 30 seconds
//   - Infinite reconnection attempts
//
// Parameters:
//   - url: HTTP(S) or WebSocket URL of the Zenon node (for example,
//     "http://127.0.0.1:35997" or "ws://127.0.0.1:35998")
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
//   - url: HTTP(S) or WebSocket URL of the Zenon node (for example,
//     "https://node.example" or "wss://node.example")
//   - opts: ClientOptions struct configuring connection behavior
//
// Available options:
//   - AutoReconnect: Enable automatic reconnection on connection loss (default: true)
//   - ReconnectDelay: Initial delay between reconnect attempts (default: 1s)
//   - MaxReconnectDelay: Maximum delay with exponential backoff (default: 30s)
//   - ReconnectAttempts: Max reconnection attempts, 0 for infinite (default: 0)
//   - HealthCheckInterval: Interval for connection health checks (default: 30s, 0 to disable)
//   - HealthCheckCommand: RPC command for health checks (default: "ledger.getFrontierMomentum")
//   - DialPolicy: Authorizes resolved destinations; nil allows all (see RejectNonPublicDestinations)
//   - MaxHTTPResponseBytes, HTTPTimeout: Bounds for HTTP(S) JSON-RPC responses
//   - MaxSubscriptionMessageBytes: Bound for subscription WebSocket frames
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
// The client validates and normalizes the transport URL automatically. HTTP
// clients support request/response RPC calls; subscriptions require ws or wss.
func NewRpcClientWithOptions(url string, opts ClientOptions) (*RpcClient, error) {
	if err := ValidateConnectionURL(url); err != nil {
		return nil, fmt.Errorf("invalid RPC URL: %w", err)
	}

	normalized, err := NormalizeConnectionURL(url)
	if err != nil {
		return nil, err
	}

	c := &RpcClient{
		url:               normalized,
		status:            Uninitialized,
		autoReconnect:     opts.AutoReconnect,
		reconnectDelay:    opts.ReconnectDelay,
		maxReconnectDelay: opts.MaxReconnectDelay,
		reconnectAttempts: opts.ReconnectAttempts,
		// Buffered so Stop()'s signal is never dropped when the reconnect
		// goroutine is not currently parked in a receive.
		stopReconnectChan:       make(chan struct{}, 1),
		onConnectionEstablished: make([]ConnectionEstablishedCallback, 0),
		onConnectionLost:        make([]ConnectionLostCallback, 0),
		healthCheckCmd:          opts.HealthCheckCommand,
		subscriptions:           make(map[*NormalizedSubscription]struct{}),

		dialPolicy:                  opts.DialPolicy,
		maxHTTPResponseBytes:        opts.MaxHTTPResponseBytes,
		httpTimeout:                 opts.HTTPTimeout,
		maxSubscriptionMessageBytes: opts.MaxSubscriptionMessageBytes,
	}
	if c.maxHTTPResponseBytes <= 0 {
		c.maxHTTPResponseBytes = DefaultMaxHTTPResponseBytes
	}
	if c.httpTimeout <= 0 {
		c.httpTimeout = DefaultHTTPTimeout
	}
	if c.maxSubscriptionMessageBytes <= 0 {
		c.maxSubscriptionMessageBytes = DefaultMaxSubscriptionMessageBytes
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

// dial opens the go-zenon JSON-RPC transport for c.url, routing every TCP
// connection through the dial policy and, for HTTP(S), bounding response
// bodies and call duration.
func (c *RpcClient) dial() (*server.Client, error) {
	parsed, err := url.Parse(c.url)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		// Note: HTTP transports connect lazily, so the dial policy is applied
		// on the first call rather than here.
		return server.DialHTTPWithClient(c.url, newHTTPClient(c.dialPolicy, c.maxHTTPResponseBytes, c.httpTimeout))
	case "ws", "wss":
		ctx, cancel := context.WithTimeout(context.Background(), c.httpTimeout)
		defer cancel()
		dialer, guard := newWebsocketDialer(c.dialPolicy)
		client, err := server.DialWebsocketWithDialer(ctx, c.url, "", dialer)
		return client, wrapDialError(err, guard)
	default:
		return nil, fmt.Errorf("unsupported scheme %q", parsed.Scheme)
	}
}

// connect establishes the selected JSON-RPC transport and points the stable
// API objects at it.
func (c *RpcClient) connect() error {
	// Create the exported API objects exactly once. They are never reassigned,
	// so a concurrent reader of client.LedgerApi (etc.) never races a reconnect.
	c.ensureAPIsInitialized()

	c.setStatus(Connecting)

	client, err := c.dial()
	if err != nil {
		c.setStatus(Stopped)
		return fmt.Errorf("failed to connect to %s: %w", c.url, err)
	}

	// server.Dial can return after Stop() latched the client. Publish the new
	// connection, point the API objects at it, and transition to Running all
	// while holding lifecycleLock and re-checking the stopped latch, so a late
	// reconnect can never resurrect a stopped client. Abandon the connection if
	// it lost the race.
	caller := transport.NewNormalizingCaller(client)
	c.lifecycleLock.Lock()
	if c.stopped {
		c.lifecycleLock.Unlock()
		client.Close()
		c.setStatus(Stopped)
		return fmt.Errorf("rpc client stopped during connect")
	}
	c.client = client
	c.caller = caller
	c.currentAttempt = 0
	c.apiCaller.set(caller)
	c.SubscriberApi.SetClient(client)
	c.setStatus(Running)
	c.lifecycleLock.Unlock()

	// Trigger connection established callbacks
	c.triggerConnectionEstablished()

	return nil
}

// ensureAPIsInitialized creates the stable API objects on first use. The
// objects share c.apiCaller and are never reassigned afterwards.
func (c *RpcClient) ensureAPIsInitialized() {
	c.apiInitOnce.Do(func() {
		c.apiCaller = &swappableCaller{}
		c.AcceleratorApi = embedded.NewAcceleratorApi(c.apiCaller)
		c.BridgeApi = embedded.NewBridgeApi(c.apiCaller)
		c.PillarApi = embedded.NewPillarApi(c.apiCaller)
		c.PlasmaApi = embedded.NewPlasmaApi(c.apiCaller)
		c.SentinelApi = embedded.NewSentinelApi(c.apiCaller)
		c.SporkApi = embedded.NewSporkApi(c.apiCaller)
		c.StakeApi = embedded.NewStakeApi(c.apiCaller)
		c.SwapApi = embedded.NewSwapApi(c.apiCaller)
		c.TokenApi = embedded.NewTokenApi(c.apiCaller)
		c.LiquidityApi = embedded.NewLiquidityApi(c.apiCaller)
		c.HtlcApi = embedded.NewHtlcApi(c.apiCaller)
		c.LedgerApi = api.NewLedgerApi(c.apiCaller)
		c.StatsApi = api.NewStatsApi(c.apiCaller)
		c.SubscriberApi = api.NewSubscriberApi(nil)
	})
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

	c.lifecycleLock.Lock()
	caller := c.caller
	c.lifecycleLock.Unlock()
	if caller == nil {
		return
	}

	var result interface{}
	err := caller.CallContext(ctx, &result, c.healthCheckCmd)
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
	c.lifecycleLock.Lock()
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
	stopped := c.stopped
	c.lifecycleLock.Unlock()

	// Trigger connection lost callbacks
	c.triggerConnectionLost(err)

	// Start reconnection if enabled and the client was not intentionally stopped
	if c.autoReconnect && !stopped {
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

	c.lifecycleLock.Lock()
	if c.stopped {
		c.lifecycleLock.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.reconnectCtx = ctx
	c.reconnectCtxCancel = cancel
	c.currentAttempt = 0
	c.lifecycleLock.Unlock()
	defer cancel()

	delay := c.reconnectDelay

	for {
		// Stop() latches c.stopped; check it directly so the loop terminates
		// even if the channel/context signals were missed.
		if c.isStopped() {
			return
		}
		select {
		case <-c.stopReconnectChan:
			return
		case <-ctx.Done():
			return
		default:
		}

		// Check if we've exceeded max attempts
		c.lifecycleLock.Lock()
		if c.reconnectAttempts > 0 && c.currentAttempt >= c.reconnectAttempts {
			c.lifecycleLock.Unlock()
			return
		}
		c.currentAttempt++
		c.lifecycleLock.Unlock()

		// Attempt to reconnect
		if err := c.connect(); err == nil {
			// Successfully reconnected
			return
		}

		// Wait before next attempt with exponential backoff, while remaining
		// responsive to intentional shutdown.
		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return
		case <-c.stopReconnectChan:
			timer.Stop()
			return
		}
		delay *= 2
		if delay > c.maxReconnectDelay {
			delay = c.maxReconnectDelay
		}
	}
}

// isStopped reports whether Stop() has latched the client.
func (c *RpcClient) isStopped() bool {
	c.lifecycleLock.Lock()
	defer c.lifecycleLock.Unlock()
	return c.stopped
}

// Restart manually triggers a reconnection.
//
// Restart is the only sanctioned way to reuse a client after Stop(): it clears
// the stopped latch that Stop() set before re-establishing the connection.
func (c *RpcClient) Restart() error {
	c.Stop()
	time.Sleep(100 * time.Millisecond) // Brief delay

	// Clear the stopped latch so connect() will publish the new connection, and
	// drain any stop signal Stop() left in the buffered channel. Otherwise the
	// next connection loss would start startReconnect(), immediately consume the
	// stale signal, and exit without reconnecting.
	c.lifecycleLock.Lock()
	c.stopped = false
	select {
	case <-c.stopReconnectChan:
	default:
	}
	c.lifecycleLock.Unlock()

	return c.connect()
}

// Stop gracefully shuts down the RPC client, closing its HTTP or WebSocket transport
// and stopping all background tasks.
//
// This method:
//   - Closes the underlying RPC transport
//   - Closes normalized subscriptions and their dedicated WebSocket connections
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

	// Latch stopped and tear down the connection atomically, so an in-flight
	// reconnect's connect() observes the latch and abandons its connection
	// instead of resurrecting the client (findings #32, #33).
	c.lifecycleLock.Lock()
	c.stopped = true
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
	c.caller = nil
	// Point the stable API objects at nothing so post-Stop calls fail cleanly.
	if c.apiCaller != nil {
		c.apiCaller.set(nil)
	}
	if c.SubscriberApi != nil {
		c.SubscriberApi.SetClient(nil)
	}
	reconnectCancel := c.reconnectCtxCancel
	c.lifecycleLock.Unlock()

	c.closeNormalizedSubscriptions()

	// Stop monitoring
	if c.monitorCancel != nil {
		c.monitorCancel()
	}
	if c.monitorTicker != nil {
		c.monitorTicker.Stop()
	}

	// Stop reconnection
	if reconnectCancel != nil {
		reconnectCancel()
	}
	select {
	case c.stopReconnectChan <- struct{}{}:
	default:
	}

	// Release connection lifecycle callbacks on intentional disconnect.
	c.callbackLock.Lock()
	c.onConnectionEstablished = nil
	c.onConnectionLost = nil
	c.callbackLock.Unlock()
}
