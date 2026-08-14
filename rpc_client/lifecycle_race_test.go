package rpc_client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/0x3639/znn-sdk-go/transport"
	"github.com/gorilla/websocket"
)

// Finding #32: Stop() must be final. While the auto-reconnect loop is blocked
// inside server.Dial (delayed handshake), Stop() runs to completion; the
// pending connect() must not then publish the new connection and flip the
// client back to Running.
func TestStopDuringInFlightReconnectDoesNotResurrectClient(t *testing.T) {
	var slowHandshake atomic.Bool
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if slowHandshake.Load() {
			time.Sleep(500 * time.Millisecond)
		}
		connection, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer connection.Close()
		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()

	options := DefaultClientOptions()
	options.HealthCheckInterval = 0
	options.ReconnectDelay = time.Millisecond
	options.MaxReconnectDelay = time.Millisecond
	client, err := NewRpcClientWithOptions("ws"+strings.TrimPrefix(server.URL, "http"), options)
	if err != nil {
		t.Fatalf("NewRpcClientWithOptions: %v", err)
	}
	defer client.Stop()
	if client.Status() != Running {
		t.Fatalf("initial status = %v, want Running", client.Status())
	}

	slowHandshake.Store(true)
	client.handleConnectionLoss(errors.New("simulated connection loss"))

	time.Sleep(100 * time.Millisecond)
	client.Stop()
	if client.Status() != Stopped {
		t.Fatalf("status immediately after Stop() = %v, want Stopped", client.Status())
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if got := client.Status(); got != Stopped {
			t.Fatalf("client resurrected after Stop(): status = %v", got)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// Finding #33: connection lifecycle fields are mutated from multiple
// goroutines. Exercise Restart()/Stop()/Status() concurrently under -race.
func TestRpcClientLifecycleFieldsRaceFree(t *testing.T) {
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		connection, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer connection.Close()
		time.Sleep(20 * time.Millisecond)
	}))
	defer server.Close()

	options := DefaultClientOptions()
	options.HealthCheckInterval = 5 * time.Millisecond
	options.ReconnectDelay = time.Millisecond
	options.MaxReconnectDelay = time.Millisecond
	client, err := NewRpcClientWithOptions("ws"+strings.TrimPrefix(server.URL, "http"), options)
	if err != nil {
		t.Fatalf("NewRpcClientWithOptions: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = client.Status()
				client.handleConnectionLoss(errors.New("churn"))
				time.Sleep(time.Millisecond)
			}
		}()
	}
	wg.Wait()
	client.Stop()
}

// Review finding #rev4: exported API fields (LedgerApi, ...) were reassigned on
// every reconnect while callers read them without a lock. They must be stable
// (write-once) so concurrent reads during reconnect are race-free.
func TestRpcClientAPIFieldsStableDuringReconnect(t *testing.T) {
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		connection, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer connection.Close()
		time.Sleep(20 * time.Millisecond)
	}))
	defer server.Close()

	options := DefaultClientOptions()
	options.HealthCheckInterval = 0
	options.ReconnectDelay = time.Millisecond
	options.MaxReconnectDelay = time.Millisecond
	client, err := NewRpcClientWithOptions("ws"+strings.TrimPrefix(server.URL, "http"), options)
	if err != nil {
		t.Fatalf("NewRpcClientWithOptions: %v", err)
	}
	defer client.Stop()

	ledger := client.LedgerApi

	var wg sync.WaitGroup
	stop := make(chan struct{})
	// Reader goroutines read the exported field concurrently.
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					if client.LedgerApi != ledger {
						t.Errorf("LedgerApi field was reassigned")
						return
					}
				}
			}
		}()
	}
	// Force repeated reconnects concurrently.
	for i := 0; i < 20; i++ {
		client.handleConnectionLoss(errors.New("churn"))
		time.Sleep(time.Millisecond)
	}
	close(stop)
	wg.Wait()
}

// Review finding #rev2: Restart calls Stop(), which deposits a value into the
// buffered stopReconnectChan. If Restart does not drain it, the next
// connection loss starts startReconnect(), immediately consumes the stale
// signal, and exits without reconnecting. After a Restart, a later connection
// loss must still trigger a successful reconnect.
func TestReconnectWorksAfterRestart(t *testing.T) {
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		connection, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer connection.Close()
		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()

	options := DefaultClientOptions()
	options.HealthCheckInterval = 0
	options.ReconnectDelay = time.Millisecond
	options.MaxReconnectDelay = time.Millisecond
	client, err := NewRpcClientWithOptions("ws"+strings.TrimPrefix(server.URL, "http"), options)
	if err != nil {
		t.Fatalf("NewRpcClientWithOptions: %v", err)
	}
	defer client.Stop()

	if err := client.Restart(); err != nil {
		t.Fatalf("Restart: %v", err)
	}
	if client.Status() != Running {
		t.Fatalf("status after Restart = %v, want Running", client.Status())
	}

	// A connection loss after Restart must reconnect (not be swallowed by a
	// stale stop signal).
	client.handleConnectionLoss(errors.New("post-restart loss"))

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if client.Status() == Running {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("client did not reconnect after Restart + connection loss: status = %v", client.Status())
}

// Finding #22: reconnect racing Unsubscribe must not orphan the fresh socket.
func TestUnsubscribeDuringPendingReconnectHandshakeTerminates(t *testing.T) {
	var requests atomic.Int32
	resume := make(chan struct{})
	reconnectStarted := make(chan struct{}, 1)
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		connection, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer connection.Close()
		var rpcRequest transport.Request
		if err := connection.ReadJSON(&rpcRequest); err != nil {
			return
		}
		if requests.Add(1) > 1 {
			select {
			case reconnectStarted <- struct{}{}:
			default:
			}
			select {
			case <-resume:
			case <-time.After(5 * time.Second):
			}
			_ = connection.WriteJSON(map[string]interface{}{
				"jsonrpc": "2.0", "id": rpcRequest.ID, "result": "sub-unsub-race-2",
			})
			select {
			case <-resume:
			case <-time.After(5 * time.Second):
			}
			return
		}
		_ = connection.WriteJSON(map[string]interface{}{
			"jsonrpc": "2.0", "id": rpcRequest.ID, "result": "sub-unsub-race",
		})
	}))
	defer server.Close()
	client := newSubscriptionTestClient(t, server, func(options *ClientOptions) {
		options.ReconnectDelay = time.Millisecond
		options.MaxReconnectDelay = time.Millisecond
	})
	defer client.Stop()
	defer close(resume)

	subscription, err := client.Subscribe(context.Background(), "momentums")
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	select {
	case <-reconnectStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("subscription never attempted a reconnect")
	}
	time.Sleep(20 * time.Millisecond)

	unsubscribed := make(chan struct{})
	go func() {
		defer close(unsubscribed)
		subscription.Unsubscribe()
	}()
	select {
	case <-unsubscribed:
	case <-time.After(2 * time.Second):
		t.Fatal("Unsubscribe never returned after racing a pending reconnect handshake")
	}
}

// Finding #23: Stop() during the Subscribe handshake must not leak an
// unterminated subscription whose Events() channel never closes.
func TestSubscribeStoppedDuringHandshakeDoesNotLeak(t *testing.T) {
	var requests atomic.Int32
	release := make(chan struct{})
	subscribeReceived := make(chan struct{}, 1)
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		connection, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer connection.Close()
		var rpcRequest transport.Request
		if err := connection.ReadJSON(&rpcRequest); err != nil {
			return
		}
		if requests.Add(1) >= 1 {
			select {
			case subscribeReceived <- struct{}{}:
			default:
			}
			<-release // hold the handshake response until the test says go
			_ = connection.WriteJSON(map[string]interface{}{
				"jsonrpc": "2.0", "id": rpcRequest.ID, "result": "sub-stop-race",
			})
			time.Sleep(20 * time.Millisecond)
		}
	}))
	defer server.Close()

	client := newSubscriptionTestClient(t, server, func(options *ClientOptions) {
		options.AutoReconnect = false
	})

	type result struct {
		sub *NormalizedSubscription
		err error
	}
	done := make(chan result, 1)
	go func() {
		sub, err := client.Subscribe(context.Background(), "momentums")
		done <- result{sub, err}
	}()

	select {
	case <-subscribeReceived:
	case <-time.After(2 * time.Second):
		t.Fatal("server never received the subscribe request")
	}
	// Stop while the handshake response is still held.
	client.Stop()
	close(release)

	select {
	case res := <-done:
		// Either an error, or a subscription whose Events() closes promptly.
		if res.err == nil {
			select {
			case <-res.sub.Events():
			case <-time.After(2 * time.Second):
				t.Fatal("Subscribe succeeded during Stop() but Events() never closed (leaked subscription)")
			}
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Subscribe never returned after Stop() during handshake")
	}
}
