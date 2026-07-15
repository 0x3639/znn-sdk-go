package rpc_client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/0x3639/znn-sdk-go/transport"
	"github.com/gorilla/websocket"
)

func newSubscriptionTestServer(t *testing.T, handle func(*websocket.Conn, transport.Request)) *httptest.Server {
	t.Helper()
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		connection, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer connection.Close()
		var rpcRequest transport.Request
		if err := connection.ReadJSON(&rpcRequest); err != nil {
			return // The RpcClient lifecycle connection sends no request.
		}
		handle(connection, rpcRequest)
	}))
}

func newSubscriptionTestClient(t *testing.T, server *httptest.Server, configure func(*ClientOptions)) *RpcClient {
	t.Helper()
	options := DefaultClientOptions()
	options.HealthCheckInterval = 0
	if configure != nil {
		configure(&options)
	}
	client, err := NewRpcClientWithOptions("ws"+strings.TrimPrefix(server.URL, "http"), options)
	if err != nil {
		t.Fatalf("NewRpcClientWithOptions: %v", err)
	}
	return client
}

func TestSubscribeRejectsInvalidClientAndTransportState(t *testing.T) {
	var nilClient *RpcClient
	if _, err := nilClient.Subscribe(context.Background(), "momentums"); err == nil {
		t.Fatal("nil client accepted a subscription")
	}
	if _, err := (&RpcClient{status: Stopped}).Subscribe(context.Background(), "momentums"); err == nil {
		t.Fatal("stopped client accepted a subscription")
	}
	if _, err := (&RpcClient{status: Running, url: "://bad"}).Subscribe(context.Background(), "momentums"); err == nil {
		t.Fatal("invalid URL accepted a subscription")
	}
	if _, err := (&RpcClient{status: Running, url: "http://127.0.0.1"}).Subscribe(context.Background(), "momentums"); err == nil {
		t.Fatal("HTTP client accepted a subscription")
	}
	var nilSubscription *NormalizedSubscription
	nilSubscription.Unsubscribe()
	closeWebSocket(nil)
}

func TestSubscribeSurfacesHandshakeFailures(t *testing.T) {
	tests := []struct {
		name   string
		handle func(*websocket.Conn, transport.Request)
		check  func(*testing.T, error)
	}{
		{
			name: "read failure",
			handle: func(*websocket.Conn, transport.Request) {
				// Returning closes the socket before a response is written.
			},
			check: func(t *testing.T, err error) {
				if err == nil || !strings.Contains(err.Error(), "failed to read subscription response") {
					t.Fatalf("error = %v", err)
				}
			},
		},
		{
			name: "RPC error",
			handle: func(connection *websocket.Conn, request transport.Request) {
				_ = connection.WriteJSON(map[string]interface{}{
					"jsonrpc": "2.0", "id": request.ID,
					"error": map[string]interface{}{"code": -32000, "message": "denied", "data": "detail"},
				})
			},
			check: func(t *testing.T, err error) {
				var rpcErr *transport.RPCError
				if !errors.As(err, &rpcErr) || rpcErr.Code != -32000 || rpcErr.Message != "denied" || rpcErr.Data != "detail" {
					t.Fatalf("error = %#v", err)
				}
			},
		},
		{
			name: "empty RPC message",
			handle: func(connection *websocket.Conn, request transport.Request) {
				_ = connection.WriteJSON(map[string]interface{}{
					"jsonrpc": "2.0", "id": request.ID,
					"error": map[string]interface{}{"code": -32001, "message": ""},
				})
			},
			check: func(t *testing.T, err error) {
				var rpcErr *transport.RPCError
				if !errors.As(err, &rpcErr) || rpcErr.Message != "Unknown error occurred" {
					t.Fatalf("error = %#v", err)
				}
			},
		},
		{
			name: "missing ID",
			handle: func(connection *websocket.Conn, request transport.Request) {
				_ = connection.WriteJSON(map[string]interface{}{
					"jsonrpc": "2.0", "id": request.ID, "result": "",
				})
			},
			check: func(t *testing.T, err error) {
				if err == nil || !strings.Contains(err.Error(), "missing an ID") {
					t.Fatalf("error = %v", err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newSubscriptionTestServer(t, test.handle)
			defer server.Close()
			client := newSubscriptionTestClient(t, server, func(options *ClientOptions) { options.AutoReconnect = false })
			defer client.Stop()
			_, err := client.Subscribe(context.Background(), "momentums")
			test.check(t, err)
		})
	}
}

func TestSubscribeWithNilContextAndMalformedNotification(t *testing.T) {
	server := newSubscriptionTestServer(t, func(connection *websocket.Conn, request transport.Request) {
		_ = connection.WriteJSON(map[string]interface{}{
			"jsonrpc": "2.0", "id": request.ID, "result": "sub-invalid",
		})
		_ = connection.WriteJSON(map[string]interface{}{
			"jsonrpc": "2.0", "method": "unexpected.notification",
			"params": map[string]interface{}{"subscription": "sub-invalid", "result": []interface{}{}},
		})
	})
	defer server.Close()
	client := newSubscriptionTestClient(t, server, func(options *ClientOptions) { options.AutoReconnect = false })
	defer client.Stop()
	subscription, err := client.Subscribe(nil, "momentums")
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	select {
	case terminalErr := <-subscription.Err():
		if terminalErr == nil || !strings.Contains(terminalErr.Error(), "unexpected subscription method") {
			t.Fatalf("terminal error = %v", terminalErr)
		}
	case <-time.After(time.Second):
		t.Fatal("malformed notification did not terminate the subscription")
	}
	if _, ok := <-subscription.Events(); ok {
		t.Fatal("events channel remains open after terminal error")
	}
}

func TestSubscriptionSocketLossWithoutReconnectIsTerminal(t *testing.T) {
	server := newSubscriptionTestServer(t, func(connection *websocket.Conn, request transport.Request) {
		_ = connection.WriteJSON(map[string]interface{}{
			"jsonrpc": "2.0", "id": request.ID, "result": "sub-close",
		})
	})
	defer server.Close()
	client := newSubscriptionTestClient(t, server, func(options *ClientOptions) { options.AutoReconnect = false })
	defer client.Stop()
	subscription, err := client.Subscribe(context.Background(), "momentums")
	if err != nil {
		t.Fatal(err)
	}
	select {
	case terminalErr := <-subscription.Err():
		if terminalErr == nil || !strings.Contains(terminalErr.Error(), "connection lost") {
			t.Fatalf("terminal error = %v", terminalErr)
		}
	case <-time.After(time.Second):
		t.Fatal("socket loss did not terminate the subscription")
	}
}

func TestSubscriptionReconnectExhaustion(t *testing.T) {
	var requestCount atomic.Int32
	server := newSubscriptionTestServer(t, func(connection *websocket.Conn, request transport.Request) {
		if requestCount.Add(1) == 1 {
			_ = connection.WriteJSON(map[string]interface{}{
				"jsonrpc": "2.0", "id": request.ID, "result": "sub-first",
			})
		}
		// Initial and retry connections both close immediately after this handler.
	})
	defer server.Close()
	client := newSubscriptionTestClient(t, server, func(options *ClientOptions) {
		options.ReconnectDelay = time.Millisecond
		options.MaxReconnectDelay = time.Millisecond
		options.ReconnectAttempts = 1
	})
	defer client.Stop()
	subscription, err := client.Subscribe(context.Background(), "momentums")
	if err != nil {
		t.Fatal(err)
	}
	select {
	case terminalErr := <-subscription.Err():
		if terminalErr == nil || !strings.Contains(terminalErr.Error(), "reconnect failed after 1 attempts") {
			t.Fatalf("terminal error = %v", terminalErr)
		}
	case <-time.After(time.Second):
		t.Fatal("reconnect exhaustion did not terminate the subscription")
	}
}

func TestReconnectCancellationAndErrorBuffering(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	subscription := &NormalizedSubscription{
		client: &RpcClient{reconnectDelay: 0, maxReconnectDelay: 0, reconnectAttempts: 1},
		ctx:    canceled,
		errors: make(chan error, 1),
	}
	if connection, ok := subscription.reconnect(); ok || connection != nil {
		t.Fatalf("canceled reconnect = %v, %v", connection, ok)
	}

	first := errors.New("first")
	subscription.finishWithError(first)
	subscription.finishWithError(errors.New("second"))
	if got := <-subscription.errors; !errors.Is(got, first) {
		t.Fatalf("buffered error = %v", got)
	}
}
