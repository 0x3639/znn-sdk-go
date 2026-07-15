package rpc_client

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/0x3639/znn-sdk-go/transport"
)

type lifecycleCaller struct {
	err   error
	calls atomic.Int32
	seen  chan struct{}
}

func (c *lifecycleCaller) Call(_ interface{}, _ string, _ ...interface{}) error {
	c.calls.Add(1)
	if c.seen != nil {
		select {
		case c.seen <- struct{}{}:
		default:
		}
	}
	return c.err
}

func TestPerformHealthCheckSuccessFailureAndClosed(t *testing.T) {
	raw := &lifecycleCaller{seen: make(chan struct{}, 1)}
	client := &RpcClient{
		status:                  Running,
		caller:                  transport.NewNormalizingCaller(raw),
		healthCheckCmd:          "health.check",
		stopReconnectChan:       make(chan struct{}, 1),
		subscriptions:           make(map[*NormalizedSubscription]struct{}),
		onConnectionEstablished: make([]ConnectionEstablishedCallback, 0),
		onConnectionLost:        make([]ConnectionLostCallback, 0),
	}
	client.performHealthCheck()
	if raw.calls.Load() != 1 || client.Status() != Running {
		t.Fatalf("success calls/status = %d/%v", raw.calls.Load(), client.Status())
	}

	lost := make(chan error, 1)
	client.AddOnConnectionLostCallback(func(err error) { lost <- err })
	raw.err = errors.New("offline")
	client.performHealthCheck()
	if client.Status() != Stopped {
		t.Fatalf("failure status = %v, want Stopped", client.Status())
	}
	select {
	case err := <-lost:
		if err == nil || !strings.Contains(err.Error(), "health check failed") {
			t.Fatalf("connection-lost error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("connection-lost callback was not invoked")
	}

	before := raw.calls.Load()
	client.performHealthCheck()
	if raw.calls.Load() != before {
		t.Fatal("closed client performed another health check")
	}
}

func TestStartMonitoringTicksAndStops(t *testing.T) {
	raw := &lifecycleCaller{seen: make(chan struct{}, 1)}
	client := &RpcClient{
		status:            Running,
		caller:            transport.NewNormalizingCaller(raw),
		healthCheckCmd:    "health.check",
		stopReconnectChan: make(chan struct{}, 1),
		subscriptions:     make(map[*NormalizedSubscription]struct{}),
	}
	client.startMonitoring(time.Millisecond)
	select {
	case <-raw.seen:
	case <-time.After(time.Second):
		t.Fatal("monitor did not perform a health check")
	}
	client.Stop()
	if client.monitorCancel == nil || client.monitorTicker == nil {
		t.Fatal("monitor lifecycle fields were not initialized")
	}
}

func TestHandleConnectionLossReconnects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	options := DefaultClientOptions()
	options.HealthCheckInterval = 0
	options.ReconnectDelay = time.Millisecond
	options.MaxReconnectDelay = 2 * time.Millisecond
	options.ReconnectAttempts = 2
	client, err := NewRpcClientWithOptions(server.URL, options)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Stop()

	lost := make(chan error, 1)
	client.AddOnConnectionLostCallback(func(err error) { lost <- err })
	client.handleConnectionLoss(errors.New("disconnected"))
	select {
	case <-lost:
	case <-time.After(time.Second):
		t.Fatal("connection-lost callback was not invoked")
	}
	deadline := time.Now().Add(time.Second)
	for client.Status() != Running && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if client.Status() != Running {
		t.Fatalf("reconnect status = %v, want Running", client.Status())
	}
	for {
		if client.reconnectLock.TryLock() {
			client.reconnectLock.Unlock()
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("reconnect goroutine did not finish")
		}
		time.Sleep(time.Millisecond)
	}
	client.setStatus(Stopped)
	client.handleConnectionLoss(errors.New("already stopped check"))
}

func TestStartReconnectStopsAfterAttemptLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	server.Close()
	client := &RpcClient{
		url:                     url,
		status:                  Stopped,
		reconnectDelay:          time.Millisecond,
		maxReconnectDelay:       time.Millisecond,
		reconnectAttempts:       1,
		stopReconnectChan:       make(chan struct{}, 1),
		subscriptions:           make(map[*NormalizedSubscription]struct{}),
		onConnectionEstablished: make([]ConnectionEstablishedCallback, 0),
		onConnectionLost:        make([]ConnectionLostCallback, 0),
	}
	client.startReconnect()
	if client.currentAttempt != 1 || client.Status() != Stopped {
		t.Fatalf("attempts/status = %d/%v", client.currentAttempt, client.Status())
	}
}

func TestRestartReconnectsStoppedClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	options := DefaultClientOptions()
	options.AutoReconnect = false
	options.HealthCheckInterval = 0
	client, err := NewRpcClientWithOptions(server.URL, options)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Stop()
	if err := client.Restart(); err != nil {
		t.Fatalf("Restart: %v", err)
	}
	if client.Status() != Running || client.LedgerApi == nil || client.PlasmaApi == nil {
		t.Fatalf("restarted client = status %v ledger %v plasma %v", client.Status(), client.LedgerApi, client.PlasmaApi)
	}
}
