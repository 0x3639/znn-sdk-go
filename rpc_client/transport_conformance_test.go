package rpc_client

import (
	"context"
	"encoding/json"
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

func TestHTTPTransportLifecycleAndNormalizedError(t *testing.T) {
	requests := make(chan transport.Request, 2)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		var rpcRequest transport.Request
		if err := json.NewDecoder(request.Body).Decode(&rpcRequest); err != nil {
			t.Errorf("Decode() error = %v", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		requests <- rpcRequest
		writer.Header().Set("Content-Type", "application/json")
		if rpcRequest.Method == "test.error" {
			_ = json.NewEncoder(writer).Encode(map[string]interface{}{
				"jsonrpc": "2.0", "id": rpcRequest.ID,
				"error": map[string]interface{}{"code": -32000, "message": "node unavailable", "data": map[string]interface{}{"retry": true}},
			})
			return
		}
		_ = json.NewEncoder(writer).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": rpcRequest.ID, "result": map[string]interface{}{"ok": true}})
	}))
	defer server.Close()

	options := DefaultClientOptions()
	options.HealthCheckInterval = 0
	client, err := NewRpcClientWithOptions(server.URL, options)
	if err != nil {
		t.Fatalf("NewRpcClientWithOptions() error = %v", err)
	}
	defer client.Stop()

	var result map[string]interface{}
	if callErr := client.caller.Call(&result, "test.read", "address", 0, 25); callErr != nil {
		t.Fatalf("Call(test.read) error = %v", callErr)
	}
	if result["ok"] != true {
		t.Fatalf("result = %#v", result)
	}
	readRequest := <-requests
	if readRequest.JSONRPC != "2.0" || !reflectParameters(readRequest.Params, []interface{}{"address", float64(0), float64(25)}) {
		t.Fatalf("HTTP request = %+v", readRequest)
	}

	err = client.caller.Call(&result, "test.error", "parameter")
	var rpcErr *transport.RPCError
	if !errors.As(err, &rpcErr) {
		t.Fatalf("Call(test.error) error = %T %v, want *transport.RPCError", err, err)
	}
	if rpcErr.Code != -32000 || rpcErr.Message != "node unavailable" || rpcErr.Method != "test.error" ||
		len(rpcErr.Parameters) != 1 || rpcErr.Parameters[0] != "parameter" {
		t.Fatalf("normalized error = %+v", rpcErr)
	}
	if retry := rpcErr.Data.(map[string]interface{})["retry"]; retry != true {
		t.Fatalf("error data = %#v", rpcErr.Data)
	}
}

func TestNormalizedSubscriptionReconnectsAndResubscribes(t *testing.T) {
	var subscriptions atomic.Int32
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		connection, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer connection.Close()
		var rpcRequest transport.Request
		if err := connection.ReadJSON(&rpcRequest); err != nil {
			return // RpcClient's underlying lifecycle connection sends no request.
		}
		attempt := subscriptions.Add(1)
		if rpcRequest.Method != "ledger.subscribe" || !reflectParameters(rpcRequest.Params, []interface{}{"accountBlocksByAddress", "address"}) {
			t.Errorf("subscription request = %+v", rpcRequest)
			return
		}
		subscriptionID := "sub-1"
		if attempt > 1 {
			subscriptionID = "sub-2"
		}
		_ = connection.WriteJSON(map[string]interface{}{"jsonrpc": "2.0", "id": rpcRequest.ID, "result": subscriptionID})
		height := 41 + attempt - 1
		_ = connection.WriteJSON(map[string]interface{}{
			"jsonrpc": "2.0", "method": "ledger.subscription",
			"params": map[string]interface{}{"subscription": subscriptionID, "result": []interface{}{map[string]interface{}{"height": height}}},
		})
		if attempt == 1 {
			return // Force reconnect after the first update.
		}
		_, _, _ = connection.ReadMessage() // Wait for client cleanup.
	}))
	defer server.Close()

	options := DefaultClientOptions()
	options.HealthCheckInterval = 0
	options.ReconnectDelay = 10 * time.Millisecond
	options.MaxReconnectDelay = 20 * time.Millisecond
	options.ReconnectAttempts = 3
	client, err := NewRpcClientWithOptions("ws"+strings.TrimPrefix(server.URL, "http"), options)
	if err != nil {
		t.Fatalf("NewRpcClientWithOptions() error = %v", err)
	}
	defer client.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	subscription, err := client.Subscribe(ctx, "accountBlocksByAddress", "address")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}
	first := <-subscription.Events()
	second := <-subscription.Events()
	if first.SubscriptionID != "sub-1" || updateHeight(first) != 41 {
		t.Fatalf("first event = %+v", first)
	}
	if second.SubscriptionID != "sub-2" || updateHeight(second) != 42 {
		t.Fatalf("reconnected event = %+v", second)
	}
	if subscription.ID() != "sub-2" || subscriptions.Load() != 2 {
		t.Fatalf("subscription ID/count = %q/%d", subscription.ID(), subscriptions.Load())
	}
	subscription.Unsubscribe()
	if _, ok := <-subscription.Events(); ok {
		t.Fatal("Events channel remains open after Unsubscribe")
	}
}

func reflectParameters(actual, expected []interface{}) bool {
	if len(actual) != len(expected) {
		return false
	}
	for index := range actual {
		if actual[index] != expected[index] {
			return false
		}
	}
	return true
}

func updateHeight(event transport.SubscriptionEvent) int32 {
	value := event.Updates[0].(map[string]interface{})["height"].(float64)
	return int32(value)
}
