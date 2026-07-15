package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/0x3639/znn-sdk-go/transport"
	"github.com/gorilla/websocket"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/server"
)

func TestSubscriberMethodsUseCanonicalTopics(t *testing.T) {
	requests := make(chan transport.Request, 8)
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	httpServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		connection, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer connection.Close()
		for {
			var rpcRequest transport.Request
			if err := connection.ReadJSON(&rpcRequest); err != nil {
				return
			}
			requests <- rpcRequest
			result := interface{}("subscription-id")
			if rpcRequest.Method == "ledger.unsubscribe" {
				result = true
			}
			if err := connection.WriteJSON(map[string]interface{}{
				"jsonrpc": "2.0", "id": rpcRequest.ID, "result": result,
			}); err != nil {
				return
			}
		}
	}))
	defer httpServer.Close()

	raw, err := server.Dial("ws" + strings.TrimPrefix(httpServer.URL, "http"))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer raw.Close()
	subscriber := NewSubscriberApi(raw)
	address := types.ParseAddressPanic("z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f")

	tests := []struct {
		name   string
		topic  string
		params []interface{}
		call   func() (*server.ClientSubscription, error)
	}{
		{
			name: "momentums", topic: "momentums", params: []interface{}{"momentums"},
			call: func() (*server.ClientSubscription, error) {
				subscription, channel, err := subscriber.ToMomentums(context.Background())
				if channel == nil {
					t.Error("momentum channel is nil")
				}
				return subscription, err
			},
		},
		{
			name: "all-account-blocks", topic: "allAccountBlocks", params: []interface{}{"allAccountBlocks"},
			call: func() (*server.ClientSubscription, error) {
				subscription, channel, err := subscriber.ToAllAccountBlocks(context.Background())
				if channel == nil {
					t.Error("account-block channel is nil")
				}
				return subscription, err
			},
		},
		{
			name: "account-by-address", topic: "accountBlocksByAddress", params: []interface{}{"accountBlocksByAddress", address.String()},
			call: func() (*server.ClientSubscription, error) {
				subscription, channel, err := subscriber.ToAccountBlocksByAddress(context.Background(), address)
				if channel == nil {
					t.Error("address account-block channel is nil")
				}
				return subscription, err
			},
		},
		{
			name: "unreceived-by-address", topic: "unreceivedAccountBlocksByAddress", params: []interface{}{"unreceivedAccountBlocksByAddress", address.String()},
			call: func() (*server.ClientSubscription, error) {
				subscription, channel, err := subscriber.ToUnreceivedAccountBlocksByAddress(context.Background(), address)
				if channel == nil {
					t.Error("unreceived channel is nil")
				}
				return subscription, err
			},
		},
	}

	var subscriptions []*server.ClientSubscription
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			subscription, err := test.call()
			if err != nil {
				t.Fatalf("subscribe error = %v", err)
			}
			if subscription == nil {
				t.Fatal("subscription is nil")
			}
			subscriptions = append(subscriptions, subscription)
			request := <-requests
			if request.Method != "ledger.subscribe" {
				t.Fatalf("method = %q", request.Method)
			}
			encoded, _ := json.Marshal(request.Params)
			want, _ := json.Marshal(test.params)
			if string(encoded) != string(want) {
				t.Fatalf("params = %s, want %s", encoded, want)
			}
		})
	}

	manager := NewSubscriptionManager()
	manager.Add(subscriptions[0])
	manager.AddMultiple(subscriptions[1:]...)
	if manager.Count() != len(subscriptions) || manager.IsEmpty() {
		t.Fatalf("managed subscriptions = %d", manager.Count())
	}
	if !manager.Remove(subscriptions[0]) {
		t.Fatal("Remove did not find the managed subscription")
	}
	<-requests // ledger.unsubscribe from Remove
	if manager.Remove(subscriptions[0]) {
		t.Fatal("Remove found an already-removed subscription")
	}
	manager.UnsubscribeAll()
	for range subscriptions[1:] {
		<-requests // ledger.unsubscribe from UnsubscribeAll
	}
	if !manager.IsEmpty() {
		t.Fatal("manager is not empty after UnsubscribeAll")
	}
}

func TestSubscriberMethodsReturnSubscriptionErrors(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	client := NewSubscriberApi(new(server.Client))
	if subscription, channel, err := client.ToMomentums(canceled); err == nil || subscription != nil || channel != nil {
		t.Fatalf("ToMomentums = %v, %v, %v", subscription, channel, err)
	}
}
