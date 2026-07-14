package transport

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

type codedDataError struct{}

func (codedDataError) Error() string          { return "node unavailable" }
func (codedDataError) ErrorCode() int         { return -32000 }
func (codedDataError) ErrorData() interface{} { return map[string]interface{}{"retry": true} }

func TestNormalizeRPCErrorPreservesContext(t *testing.T) {
	err := NormalizeRPCError(codedDataError{}, "ledger.getFrontierMomentum")
	if err.Code != -32000 || err.Message != "node unavailable" || err.Method != "ledger.getFrontierMomentum" {
		t.Fatalf("NormalizeRPCError() = %+v", err)
	}
	if !reflect.DeepEqual(err.Data, map[string]interface{}{"retry": true}) {
		t.Fatalf("Data = %#v", err.Data)
	}
	if err.Parameters == nil || len(err.Parameters) != 0 {
		t.Fatalf("Parameters = %#v, want empty array", err.Parameters)
	}
	var target codedDataError
	if !errors.As(err, &target) {
		t.Fatal("normalized error does not unwrap to its cause")
	}
}

func TestRequestAndSubscriptionNormalization(t *testing.T) {
	request := NewRequest(7, "ledger.getAccountBlocksByPage", "address", 0, 25)
	if request.JSONRPC != "2.0" || request.ID != 7 || request.Method != "ledger.getAccountBlocksByPage" ||
		!reflect.DeepEqual(request.Params, []interface{}{"address", 0, 25}) {
		t.Fatalf("NewRequest() = %+v", request)
	}
	params := SubscriptionParams("accountBlocksByAddress", "address")
	if !reflect.DeepEqual(params, []interface{}{"accountBlocksByAddress", "address"}) {
		t.Fatalf("SubscriptionParams() = %#v", params)
	}
	event, err := NormalizeSubscriptionNotification("ledger.subscription", json.RawMessage(`{"subscription":"sub-1","result":[{"height":42}]}`))
	if err != nil {
		t.Fatalf("NormalizeSubscriptionNotification() error = %v", err)
	}
	if event.SubscriptionID != "sub-1" || len(event.Updates) != 1 {
		t.Fatalf("event = %+v", event)
	}
	update := event.Updates[0].(map[string]interface{})
	if update["height"] != float64(42) {
		t.Fatalf("update = %#v", update)
	}
}

func TestNormalizeSubscriptionNotificationRejectsInvalidInput(t *testing.T) {
	for _, test := range []struct {
		method string
		params string
	}{
		{"other.subscription", `{"subscription":"sub-1","result":[]}`},
		{"ledger.subscription", `{"subscription":"","result":[]}`},
		{"ledger.subscription", `{"subscription":"sub-1","result":{}}`},
	} {
		if _, err := NormalizeSubscriptionNotification(test.method, json.RawMessage(test.params)); err == nil {
			t.Errorf("NormalizeSubscriptionNotification(%q, %s) accepted invalid input", test.method, test.params)
		}
	}
}
