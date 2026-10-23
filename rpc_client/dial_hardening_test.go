package rpc_client

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/0x3639/znn-sdk-go/transport"
	"github.com/gorilla/websocket"
	znnrpc "github.com/zenon-network/go-zenon/rpc/server"
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
	case event, ok := <-subscription.Events():
		if !ok {
			// Cleanup closed the channel after queuing the terminal error;
			// a closed receive is not a delivered frame.
			if terminalErr := <-subscription.Err(); !errors.Is(terminalErr, websocket.ErrReadLimit) {
				t.Fatalf("terminal error = %v, want websocket.ErrReadLimit", terminalErr)
			}
			return
		}
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

// Codex finding: a limit of math.MaxInt64 overflowed remaining+1 and sliced
// with a negative bound.
func TestLimitedBodyMaxInt64LimitDoesNotPanic(t *testing.T) {
	body := &limitedBody{ReadCloser: io.NopCloser(strings.NewReader("payload")), remaining: math.MaxInt64}
	data, err := io.ReadAll(body)
	if err != nil || string(data) != "payload" {
		t.Fatalf("ReadAll = %q, %v", data, err)
	}
}

type recordingCloser struct {
	io.Reader
	closed bool
}

func (r *recordingCloser) Close() error { r.closed = true; return nil }

// Codex finding: an oversized body must be closed when the limit error is
// returned so the connection is released without waiting for the timeout.
// Both detection paths are covered: a read that crosses the limit, and the
// one-byte probe issued once exactly limit bytes have been delivered.
func TestLimitedBodyClosesUnderlyingOnOversize(t *testing.T) {
	rc := &recordingCloser{Reader: strings.NewReader("123456")}
	body := &limitedBody{ReadCloser: rc, remaining: 5}
	if _, err := io.ReadAll(body); !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("err = %v, want ErrResponseTooLarge", err)
	}
	if !rc.closed {
		t.Fatal("underlying body not closed after ErrResponseTooLarge")
	}

	probe := &recordingCloser{Reader: strings.NewReader("123456")}
	exact := &limitedBody{ReadCloser: probe, remaining: 5}
	buf := make([]byte, 5)
	if n, err := io.ReadFull(exact, buf); err != nil || n != 5 {
		t.Fatalf("ReadFull = %d, %v", n, err)
	}
	if probe.closed {
		t.Fatal("body closed before the limit was exceeded")
	}
	if _, err := exact.Read(buf[:1]); !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("probe err = %v, want ErrResponseTooLarge", err)
	}
	if !probe.closed {
		t.Fatal("underlying body not closed when the probe detected the overflow")
	}
}

// A gzip-encoded body is limited after transparent decompression, so a small
// compressed payload cannot expand past MaxHTTPResponseBytes.
func TestHTTPGzipExpandedBodyBounded(t *testing.T) {
	var compressed bytes.Buffer
	zw := gzip.NewWriter(&compressed)
	_, _ = zw.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"pad":"` + strings.Repeat("x", 64<<10) + `"}}`))
	_ = zw.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		_, _ = w.Write(compressed.Bytes())
	}))
	defer server.Close()
	if compressed.Len() > 1024 {
		t.Fatalf("compressed payload is %d bytes; test needs it under the limit", compressed.Len())
	}
	opts := DefaultClientOptions()
	opts.HealthCheckInterval = 0
	opts.AutoReconnect = false
	opts.MaxHTTPResponseBytes = 1024
	client, err := NewRpcClientWithOptions(server.URL, opts)
	if err != nil {
		t.Fatalf("NewRpcClientWithOptions: %v", err)
	}
	defer client.Stop()
	if _, err := client.LedgerApi.GetFrontierMomentum(); !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge after gzip expansion, got %v", err)
	}
}

// Codex finding: go-zenon buffers non-2xx bodies and discards the read
// error, so an oversized chunked error response surfaced as a generic HTTP
// error. The size-limit failure must be preserved.
func TestHTTPOversizedErrorBodyReportsLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		flusher, _ := w.(http.Flusher)
		for i := 0; i < 8; i++ { // chunked: no Content-Length to reject up front
			_, _ = w.Write([]byte(strings.Repeat("e", 512)))
			if flusher != nil {
				flusher.Flush()
			}
		}
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
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}
}

// Codex finding: buffering non-2xx bodies must not delay redirects, which
// net/http follows without waiting for the body (it drains at most a few KiB
// for connection reuse). A 302 is used because
// go-zenon's POST requests set no GetBody, so net/http never follows 307/308
// for them; 301/302/303 are converted to GET and followed.
func TestHTTPRedirectNotDelayedByBodyBuffering(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":null}`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/ok")
		w.Header().Set("Content-Length", "4096")
		w.WriteHeader(http.StatusFound)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		// In-limit body, delivered only after the client's HTTPTimeout.
		time.Sleep(1500 * time.Millisecond)
		_, _ = w.Write([]byte(strings.Repeat("r", 4096)))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	opts := DefaultClientOptions()
	opts.HealthCheckInterval = 0
	opts.AutoReconnect = false
	opts.HTTPTimeout = 500 * time.Millisecond
	client, err := NewRpcClientWithOptions(server.URL, opts)
	if err != nil {
		t.Fatalf("NewRpcClientWithOptions: %v", err)
	}
	defer client.Stop()
	started := time.Now()
	if _, err := client.LedgerApi.GetFrontierMomentum(); err != nil {
		t.Fatalf("redirect was not followed promptly: %v", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("redirect took %v; the redirect body must not be read before following", elapsed)
	}
}

// An in-limit terminal error body must still reach go-zenon's HTTPError
// with its status and bytes intact.
func TestHTTPInLimitErrorBodyPreserved(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("boom"))
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
	var httpErr znnrpc.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("err = %T %v, want znnrpc.HTTPError", err, err)
	}
	if httpErr.StatusCode != http.StatusBadGateway || string(httpErr.Body) != "boom" {
		t.Fatalf("HTTPError = %d %q, want 502 \"boom\"", httpErr.StatusCode, httpErr.Body)
	}
}

// Codex finding (round 3): 3xx responses that net/http will not follow (307
// and 308 for go-zenon's POSTs, which set no GetBody; 300 always) are
// terminal errors and must keep the size-limit failure like any other.
func TestHTTPOversizedUnfollowedRedirectReportsLimit(t *testing.T) {
	for _, status := range []int{http.StatusMultipleChoices, http.StatusTemporaryRedirect, http.StatusPermanentRedirect} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", "/elsewhere")
				w.WriteHeader(status)
				flusher, _ := w.(http.Flusher)
				for i := 0; i < 8; i++ { // chunked: no Content-Length to reject up front
					_, _ = w.Write([]byte(strings.Repeat("e", 512)))
					if flusher != nil {
						flusher.Flush()
					}
				}
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
			if !errors.Is(err, ErrResponseTooLarge) {
				t.Fatalf("expected ErrResponseTooLarge, got %v", err)
			}
		})
	}
}

// Codex finding (round 4): net/http decides 307/308 replayability from the
// ORIGINAL request, so after a 302 turns the POST into a bodiless GET, a
// 307/308 on the next hop is still returned unfollowed. The size-limit error
// must survive that chain too.
func TestHTTPOversizedRedirectAfterChainReportsLimit(t *testing.T) {
	for _, status := range []int{http.StatusTemporaryRedirect, http.StatusPermanentRedirect} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/hop", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", "/final")
				w.WriteHeader(status)
				flusher, _ := w.(http.Flusher)
				for i := 0; i < 8; i++ {
					_, _ = w.Write([]byte(strings.Repeat("e", 512)))
					if flusher != nil {
						flusher.Flush()
					}
				}
			})
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/hop", http.StatusFound)
			})
			server := httptest.NewServer(mux)
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
			if !errors.Is(err, ErrResponseTooLarge) {
				t.Fatalf("expected ErrResponseTooLarge, got %v", err)
			}
		})
	}
}
