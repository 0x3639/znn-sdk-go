package transport

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

type stubCaller struct {
	err   error
	calls int
}

func (c *stubCaller) Call(_ interface{}, _ string, _ ...interface{}) error {
	c.calls++
	return c.err
}

type stubContextCaller struct {
	stubCaller
	contextErr   error
	contextCalls int
}

func (c *stubContextCaller) CallContext(_ context.Context, _ interface{}, _ string, _ ...interface{}) error {
	c.contextCalls++
	return c.contextErr
}

type emptyMessageError struct{}

func (emptyMessageError) Error() string { return "" }

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

func TestNormalizingCallerCallPaths(t *testing.T) {
	var nilWrapper *NormalizingCaller
	if err := nilWrapper.Call(nil, "missing", 1); err == nil {
		t.Fatal("nil wrapper call succeeded")
	}
	if err := NewNormalizingCaller(nil).Call(nil, "missing", 1); err == nil {
		t.Fatal("nil underlying caller succeeded")
	}

	raw := new(stubCaller)
	wrapper := NewNormalizingCaller(raw)
	if err := wrapper.Call(nil, "ok", 1); err != nil {
		t.Fatalf("success error = %v", err)
	}
	raw.err = errors.New("boom")
	err := wrapper.Call(nil, "failure", 2)
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Method != "failure" || !reflect.DeepEqual(rpcErr.Parameters, []interface{}{2}) {
		t.Fatalf("normalized error = %+v", err)
	}
}

func TestNormalizingCallerContextPaths(t *testing.T) {
	if err := (*NormalizingCaller)(nil).CallContext(context.Background(), nil, "missing"); err == nil {
		t.Fatal("nil context wrapper call succeeded")
	}

	contextual := new(stubContextCaller)
	wrapper := NewNormalizingCaller(contextual)
	if err := wrapper.CallContext(context.Background(), nil, "ok"); err != nil {
		t.Fatalf("context success error = %v", err)
	}
	if contextual.contextCalls != 1 || contextual.calls != 0 {
		t.Fatalf("context calls = %d, ordinary calls = %d", contextual.contextCalls, contextual.calls)
	}
	contextual.contextErr = context.DeadlineExceeded
	if err := wrapper.CallContext(context.Background(), nil, "failure"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("context error = %v", err)
	}

	fallback := new(stubCaller)
	wrapper = NewNormalizingCaller(fallback)
	if err := wrapper.CallContext(context.Background(), nil, "fallback"); err != nil || fallback.calls != 1 {
		t.Fatalf("fallback error = %v, calls = %d", err, fallback.calls)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := wrapper.CallContext(canceled, nil, "canceled"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled error = %v", err)
	}
}

func TestRPCErrorNilAndExistingErrorBehavior(t *testing.T) {
	var nilErr *RPCError
	if got := nilErr.Error(); got != "<nil>" {
		t.Fatalf("nil Error() = %q", got)
	}
	if nilErr.Unwrap() != nil {
		t.Fatal("nil Unwrap() returned a cause")
	}
	if NormalizeRPCError(nil, "method") != nil {
		t.Fatal("nil input returned a normalized error")
	}

	cause := errors.New("cause")
	existing := &RPCError{Code: 17, Message: "existing", Data: "data", Method: "old", Parameters: []interface{}{1}, Cause: cause}
	copy := NormalizeRPCError(existing, "")
	if copy == existing || copy.Code != 17 || copy.Method != "old" ||
		!reflect.DeepEqual(copy.Parameters, []interface{}{1}) || !errors.Is(copy, existing) {
		t.Fatalf("copied error = %+v", copy)
	}

	empty := NormalizeRPCError(emptyMessageError{}, "empty")
	if empty.Message != "Unknown error occurred" || empty.Code != -1 {
		t.Fatalf("empty-message normalization = %+v", empty)
	}
}

func TestNormalizeSubscriptionNotificationMalformedJSON(t *testing.T) {
	if _, err := NormalizeSubscriptionNotification("ledger.subscription", json.RawMessage(`{`)); err == nil {
		t.Fatal("malformed notification was accepted")
	}
}
