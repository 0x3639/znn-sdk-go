package rpc_client

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// ClientOptions Tests
// =============================================================================

func TestDefaultClientOptions(t *testing.T) {
	opts := DefaultClientOptions()

	if !opts.AutoReconnect {
		t.Error("AutoReconnect should be true by default")
	}

	if opts.ReconnectDelay != 1*time.Second {
		t.Errorf("ReconnectDelay = %v, want 1s", opts.ReconnectDelay)
	}

	if opts.MaxReconnectDelay != 30*time.Second {
		t.Errorf("MaxReconnectDelay = %v, want 30s", opts.MaxReconnectDelay)
	}

	if opts.ReconnectAttempts != 0 {
		t.Errorf("ReconnectAttempts = %d, want 0 (infinite)", opts.ReconnectAttempts)
	}

	if opts.HealthCheckInterval != 30*time.Second {
		t.Errorf("HealthCheckInterval = %v, want 30s", opts.HealthCheckInterval)
	}

	if opts.HealthCheckCommand != "ledger.getFrontierMomentum" {
		t.Errorf("HealthCheckCommand = %s, want 'ledger.getFrontierMomentum'", opts.HealthCheckCommand)
	}
}

func TestClientOptions_CustomValues(t *testing.T) {
	opts := ClientOptions{
		AutoReconnect:       false,
		ReconnectDelay:      2 * time.Second,
		MaxReconnectDelay:   60 * time.Second,
		ReconnectAttempts:   5,
		HealthCheckInterval: 10 * time.Second,
		HealthCheckCommand:  "custom.healthCheck",
	}

	if opts.AutoReconnect {
		t.Error("AutoReconnect should be false")
	}

	if opts.ReconnectDelay != 2*time.Second {
		t.Errorf("ReconnectDelay = %v, want 2s", opts.ReconnectDelay)
	}

	if opts.MaxReconnectDelay != 60*time.Second {
		t.Errorf("MaxReconnectDelay = %v, want 60s", opts.MaxReconnectDelay)
	}

	if opts.ReconnectAttempts != 5 {
		t.Errorf("ReconnectAttempts = %d, want 5", opts.ReconnectAttempts)
	}
}

// =============================================================================
// Connection Status Tests
// =============================================================================

func TestRpcClient_Status(t *testing.T) {
	client := &RpcClient{
		status: Uninitialized,
	}

	if client.Status() != Uninitialized {
		t.Errorf("Status() = %v, want Uninitialized", client.Status())
	}

	client.setStatus(Running)
	if client.Status() != Running {
		t.Errorf("Status() = %v, want Running", client.Status())
	}

	client.setStatus(Stopped)
	if client.Status() != Stopped {
		t.Errorf("Status() = %v, want Stopped", client.Status())
	}
}

func TestRpcClient_IsClosed(t *testing.T) {
	client := &RpcClient{
		status: Running,
	}

	if client.IsClosed() {
		t.Error("IsClosed() = true, want false for Running status")
	}

	client.setStatus(Stopped)
	if !client.IsClosed() {
		t.Error("IsClosed() = false, want true for Stopped status")
	}
}

// =============================================================================
// Callback Tests
// =============================================================================

func TestRpcClient_AddOnConnectionEstablishedCallback(t *testing.T) {
	client := &RpcClient{
		onConnectionEstablished: make([]ConnectionEstablishedCallback, 0),
	}

	called := false
	callback := func() {
		called = true
	}

	client.AddOnConnectionEstablishedCallback(callback)

	if len(client.onConnectionEstablished) != 1 {
		t.Errorf("Expected 1 callback, got %d", len(client.onConnectionEstablished))
	}

	// Test callback is callable
	client.onConnectionEstablished[0]()
	if !called {
		t.Error("Callback was not called")
	}
}

func TestRpcClient_AddOnConnectionLostCallback(t *testing.T) {
	client := &RpcClient{
		onConnectionLost: make([]ConnectionLostCallback, 0),
	}

	var receivedErr error
	callback := func(err error) {
		receivedErr = err
	}

	client.AddOnConnectionLostCallback(callback)

	if len(client.onConnectionLost) != 1 {
		t.Errorf("Expected 1 callback, got %d", len(client.onConnectionLost))
	}

	// Test callback is callable
	testErr := &ConnectionError{Message: "test error"}
	client.onConnectionLost[0](testErr)
	if receivedErr == nil {
		t.Error("Callback did not receive error")
	}
	if !errors.Is(receivedErr, testErr) {
		t.Errorf("Expected error %v, got %v", testErr, receivedErr)
	}
}

func TestRpcClient_TriggerConnectionEstablished(t *testing.T) {
	client := &RpcClient{
		onConnectionEstablished: make([]ConnectionEstablishedCallback, 0),
	}

	callCount := 0
	var mu sync.Mutex

	// Add multiple callbacks
	for i := 0; i < 3; i++ {
		client.AddOnConnectionEstablishedCallback(func() {
			mu.Lock()
			callCount++
			mu.Unlock()
		})
	}

	client.triggerConnectionEstablished()

	// Wait for goroutines to complete
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if callCount != 3 {
		t.Errorf("Expected 3 callbacks to be called, got %d", callCount)
	}
}

func TestRpcClient_TriggerConnectionLost(t *testing.T) {
	client := &RpcClient{
		onConnectionLost: make([]ConnectionLostCallback, 0),
	}

	callCount := 0
	var mu sync.Mutex

	// Add multiple callbacks
	for i := 0; i < 3; i++ {
		client.AddOnConnectionLostCallback(func(err error) {
			mu.Lock()
			callCount++
			mu.Unlock()
		})
	}

	testErr := &ConnectionError{Message: "test"}
	client.triggerConnectionLost(testErr)

	// Wait for goroutines to complete
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if callCount != 3 {
		t.Errorf("Expected 3 callbacks to be called, got %d", callCount)
	}
}

// =============================================================================
// URL Validation Tests
// =============================================================================

func TestNewRpcClient_InvalidURL(t *testing.T) {
	invalidURLs := []string{
		"",
		"http://localhost:35998",
		"not-a-url",
		"ws://localhost:99999",
	}

	for _, url := range invalidURLs {
		_, err := NewRpcClient(url)
		if err == nil {
			t.Errorf("NewRpcClient(%q) should return error for invalid URL", url)
		}
	}
}

func TestNewRpcClientWithOptions_InvalidURL(t *testing.T) {
	opts := DefaultClientOptions()

	_, err := NewRpcClientWithOptions("invalid-url", opts)
	if err == nil {
		t.Error("NewRpcClientWithOptions with invalid URL should return error")
	}
}

// =============================================================================
// Stop Tests
// =============================================================================

func TestRpcClient_Stop(t *testing.T) {
	client := &RpcClient{
		status:            Running,
		monitorTicker:     time.NewTicker(1 * time.Second),
		stopReconnectChan: make(chan struct{}, 1),
	}

	// Create contexts for monitoring and reconnection
	client.monitorCtx, client.monitorCancel = nil, func() {}
	client.reconnectCtx, client.reconnectCtxCancel = nil, func() {}

	client.Stop()

	if client.Status() != Stopped {
		t.Errorf("Status after Stop() = %v, want Stopped", client.Status())
	}

	if client.client != nil {
		t.Error("client should be nil after Stop()")
	}
}

// =============================================================================
// Helper Types for Tests
// =============================================================================

type ConnectionError struct {
	Message string
}

func (e *ConnectionError) Error() string {
	return e.Message
}

// =============================================================================
// Integration-style Tests (without actual network connection)
// =============================================================================

func TestRpcClient_StatusTransitions(t *testing.T) {
	client := &RpcClient{
		status: Uninitialized,
	}

	// Test status progression
	statuses := []WebsocketStatus{
		Connecting,
		Running,
		Stopped,
	}

	for _, status := range statuses {
		client.setStatus(status)
		if client.Status() != status {
			t.Errorf("After setStatus(%v), Status() = %v", status, client.Status())
		}
	}
}

func TestRpcClient_MultipleCallbacks(t *testing.T) {
	client := &RpcClient{
		onConnectionEstablished: make([]ConnectionEstablishedCallback, 0),
		onConnectionLost:        make([]ConnectionLostCallback, 0),
	}

	// Register multiple connection established callbacks
	for i := 0; i < 5; i++ {
		client.AddOnConnectionEstablishedCallback(func() {})
	}

	if len(client.onConnectionEstablished) != 5 {
		t.Errorf("Expected 5 connection established callbacks, got %d", len(client.onConnectionEstablished))
	}

	// Register multiple connection lost callbacks
	for i := 0; i < 5; i++ {
		client.AddOnConnectionLostCallback(func(err error) {})
	}

	if len(client.onConnectionLost) != 5 {
		t.Errorf("Expected 5 connection lost callbacks, got %d", len(client.onConnectionLost))
	}
}

func TestRpcClient_CallbackConcurrency(t *testing.T) {
	client := &RpcClient{
		onConnectionEstablished: make([]ConnectionEstablishedCallback, 0),
	}

	var wg sync.WaitGroup
	callCount := 0
	var mu sync.Mutex

	// Add callbacks concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.AddOnConnectionEstablishedCallback(func() {
				mu.Lock()
				callCount++
				mu.Unlock()
			})
		}()
	}

	wg.Wait()

	if len(client.onConnectionEstablished) != 10 {
		t.Errorf("Expected 10 callbacks, got %d", len(client.onConnectionEstablished))
	}

	// Trigger callbacks
	client.triggerConnectionEstablished()
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if callCount != 10 {
		t.Errorf("Expected 10 callback invocations, got %d", callCount)
	}
}

// =============================================================================
// Options Validation Tests
// =============================================================================

func TestClientOptions_ZeroValues(t *testing.T) {
	opts := ClientOptions{
		AutoReconnect:       false,
		ReconnectDelay:      0,
		MaxReconnectDelay:   0,
		ReconnectAttempts:   0,
		HealthCheckInterval: 0,
		HealthCheckCommand:  "",
	}

	// Should not panic with zero values
	if opts.AutoReconnect {
		t.Error("AutoReconnect should be false")
	}

	if opts.ReconnectDelay != 0 {
		t.Error("ReconnectDelay should be 0")
	}

	if opts.HealthCheckInterval != 0 {
		t.Error("HealthCheckInterval should be 0 (disabled)")
	}
}

func TestClientOptions_ExtremeValues(t *testing.T) {
	opts := ClientOptions{
		AutoReconnect:       true,
		ReconnectDelay:      1 * time.Millisecond,
		MaxReconnectDelay:   1 * time.Hour,
		ReconnectAttempts:   1000000,
		HealthCheckInterval: 1 * time.Millisecond,
		HealthCheckCommand:  "test.command",
	}

	if opts.ReconnectDelay != 1*time.Millisecond {
		t.Error("Should accept very short reconnect delay")
	}

	if opts.MaxReconnectDelay != 1*time.Hour {
		t.Error("Should accept very long max delay")
	}

	if opts.ReconnectAttempts != 1000000 {
		t.Error("Should accept large reconnect attempt count")
	}
}
