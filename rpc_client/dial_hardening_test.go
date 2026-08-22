package rpc_client

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/0x3639/znn-sdk-go/transport"
	"github.com/gorilla/websocket"
)

// CodeRabbit finding (SSRF): a dial policy must be applied to resolved
// destinations for the initial dial and subscription sockets.
func TestRejectNonPublicDestinations(t *testing.T) {
	rejected := []string{
		"127.0.0.1:80", "[::1]:80", "10.1.2.3:35998", "192.168.1.1:443", "172.16.0.1:80",
		"169.254.169.254:80", "[fe80::1]:80", "[fc00::1]:80", "100.64.0.1:80", "0.0.0.0:80",
		"224.0.0.1:80", "255.255.255.255:80",
	}
	for _, address := range rejected {
		if err := RejectNonPublicDestinations("tcp", address); !errors.Is(err, ErrDestinationNotAllowed) {
			t.Errorf("%s: expected ErrDestinationNotAllowed, got %v", address, err)
		}
	}
	for _, address := range []string{"8.8.8.8:443", "[2001:4860:4860::8888]:443", "1.1.1.1:35998"} {
		if err := RejectNonPublicDestinations("tcp", address); err != nil {
			t.Errorf("%s: unexpected rejection %v", address, err)
		}
	}
	if err := RejectNonPublicDestinations("tcp", "example.com:80"); !errors.Is(err, ErrDestinationNotAllowed) {
		t.Errorf("unresolved host name must be rejected, got %v", err)
	}
}

func TestDialPolicyAppliedToInitialDialAndSubscription(t *testing.T) {
	// The policy sees the resolved loopback address even though the URL uses
	// a host name, so DNS cannot bypass it.
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":null}`))
	}))
	defer httpServer.Close()
	_, port, _ := net.SplitHostPort(strings.TrimPrefix(httpServer.URL, "http://"))

	opts := DefaultClientOptions()
	opts.HealthCheckInterval = 0
	opts.AutoReconnect = false
	opts.DialPolicy = RejectNonPublicDestinations
	if _, err := NewRpcClientWithOptions("ws://localhost:"+port, opts); !errors.Is(err, ErrDestinationNotAllowed) {
		t.Errorf("ws: expected ErrDestinationNotAllowed, got %v", err)
	}
	// HTTP transports connect lazily; the policy applies on the first call.
	httpClient, err := NewRpcClientWithOptions("http://localhost:"+port, opts)
	if err != nil {
		t.Fatalf("http NewRpcClientWithOptions: %v", err)
	}
	defer httpClient.Stop()
	if _, err := httpClient.LedgerApi.GetFrontierMomentum(); err == nil || !strings.Contains(err.Error(), ErrDestinationNotAllowed.Error()) {
		t.Errorf("http: expected ErrDestinationNotAllowed, got %v", err)
	}

	// Subscription sockets consult the policy too: allow the lifecycle dial,
	// then deny the subscription dial.
	var calls int
	wsServer := newSubscriptionTestServer(t, func(*websocket.Conn, transport.Request) {})
	defer wsServer.Close()
	client := newSubscriptionTestClient(t, wsServer, func(options *ClientOptions) {
		options.AutoReconnect = false
		options.DialPolicy = func(network, address string) error {
			calls++
			if calls > 1 {
				return ErrDestinationNotAllowed
			}
			return nil
		}
	})
	defer client.Stop()
	if _, err := client.Subscribe(context.Background(), "momentums"); !errors.Is(err, ErrDestinationNotAllowed) {
		t.Fatalf("subscription dial bypassed policy: %v", err)
	}
}

// CodeRabbit finding: HTTP response bodies must be bounded before decoding.
func TestHTTPResponseBodyBounded(t *testing.T) {
	// A syntactically valid but oversized result; decoding must fail on the
	// size limit before the decoder ever sees the whole body.
	big := `{"jsonrpc":"2.0","id":1,"result":{"pad":"` + strings.Repeat("x", 4096) + `"}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(big))
	}))
	defer server.Close()

	opts := DefaultClientOptions()
	opts.HealthCheckInterval = 0
	opts.AutoReconnect = false
	opts.MaxHTTPResponseBytes = 1024
	client, err := NewRpcClientWithOptions(server.URL, opts)
	if err != nil {
		t.Fatalf("NewRpcClientWithOptions: %v", err)
	}
	defer client.Stop()
	_, err = client.LedgerApi.GetFrontierMomentum()
	if !errors.Is(err, ErrResponseTooLarge) && (err == nil || !strings.Contains(err.Error(), ErrResponseTooLarge.Error())) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}

	// An in-limit body still succeeds.
	opts.MaxHTTPResponseBytes = int64(len(big)) + 10
	client2, err := NewRpcClientWithOptions(server.URL, opts)
	if err != nil {
		t.Fatal(err)
	}
	defer client2.Stop()
	if _, err := client2.LedgerApi.GetFrontierMomentum(); err != nil && strings.Contains(err.Error(), ErrResponseTooLarge.Error()) {
		t.Fatalf("in-limit call hit the size limit: %v", err)
	}
}

// CodeRabbit finding: the subscription WebSocket must apply a read limit.
func TestSubscriptionReadLimitEnforced(t *testing.T) {
	server := newSubscriptionTestServer(t, func(connection *websocket.Conn, request transport.Request) {
		_ = connection.WriteJSON(map[string]interface{}{
			"jsonrpc": "2.0", "id": request.ID, "result": "sub-big",
		})
		_ = connection.WriteJSON(map[string]interface{}{
			"jsonrpc": "2.0", "method": "ledger.subscription",
			"params": map[string]interface{}{"subscription": "sub-big", "result": []interface{}{strings.Repeat("y", 8192)}},
		})
		time.Sleep(200 * time.Millisecond)
	})
	defer server.Close()
	client := newSubscriptionTestClient(t, server, func(options *ClientOptions) {
		options.AutoReconnect = false
		options.MaxSubscriptionMessageBytes = 1024
	})
	defer client.Stop()
	subscription, err := client.Subscribe(context.Background(), "momentums")
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	select {
	case terminalErr := <-subscription.Err():
		if !errors.Is(terminalErr, websocket.ErrReadLimit) {
			t.Fatalf("terminal error = %v, want websocket.ErrReadLimit", terminalErr)
		}
	case event := <-subscription.Events():
		t.Fatalf("oversized frame delivered: %+v", event)
	case <-time.After(2 * time.Second):
		t.Fatal("oversized frame did not terminate the subscription")
	}
}

// Codex finding: an exact-limit body must read to EOF, not ErrResponseTooLarge.
func TestLimitedBodyExactLimitReachesEOF(t *testing.T) {
	body := &limitedBody{ReadCloser: io.NopCloser(strings.NewReader("12345")), remaining: 5}
	data, err := io.ReadAll(body)
	if err != nil || string(data) != "12345" {
		t.Fatalf("ReadAll = %q, %v", data, err)
	}
	over := &limitedBody{ReadCloser: io.NopCloser(strings.NewReader("123456")), remaining: 5}
	if _, err := io.ReadAll(over); !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("over-limit err = %v", err)
	}
}

// Codex finding: proxies must be disabled while a DialPolicy is active so the
// policy cannot be bypassed by routing through a proxy.
func TestDialPolicyDisablesProxy(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:9")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:9")
	if proxyFor(RejectNonPublicDestinations) != nil {
		t.Fatal("proxy resolver returned while a policy is active")
	}
	if proxyFor(nil) == nil {
		t.Fatal("proxy resolver disabled without a policy")
	}
	dialer, _ := newWebsocketDialer(RejectNonPublicDestinations)
	if dialer.Proxy != nil {
		t.Fatal("websocket dialer still has a proxy with a policy")
	}
}
