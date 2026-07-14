package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// Caller is the JSON-RPC call surface used by SDK API namespaces.
//
// Implementations must send args as positional parameters and unmarshal a
// successful result into result. When a call fails, callers should return an
// [RPCError] containing method and parameter context.
type Caller interface {
	Call(result interface{}, method string, args ...interface{}) error
}

type contextCaller interface {
	CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error
}

// NormalizingCaller decorates an RPC caller so every failure becomes an
// [RPCError] with complete request context.
type NormalizingCaller struct {
	caller Caller
}

// NewNormalizingCaller creates a caller that preserves normalized error
// details for every request.
//
// Parameters:
//   - caller: Underlying JSON-RPC caller. It must not be nil.
//
// NewNormalizingCaller returns a reusable wrapper. A nil caller is accepted for
// construction but calls will return a normalized configuration error; this is
// useful for template-only API instances that never perform reads.
//
// Example:
//
//	normalized := transport.NewNormalizingCaller(rawClient)
//	err := normalized.Call(&result, "ledger.getFrontierMomentum")
//
// See [NormalizeRPCError] for the error mapping rules.
func NewNormalizingCaller(caller Caller) *NormalizingCaller {
	return &NormalizingCaller{caller: caller}
}

// Call performs a positional JSON-RPC request and normalizes any returned
// error with the method and parameters.
func (c *NormalizingCaller) Call(result interface{}, method string, args ...interface{}) error {
	if c == nil || c.caller == nil {
		return NormalizeRPCError(errors.New("RPC caller is not initialized"), method, args...)
	}
	if err := c.caller.Call(result, method, args...); err != nil {
		return NormalizeRPCError(err, method, args...)
	}
	return nil
}

// CallContext performs a positional JSON-RPC request with cancellation and
// normalizes any returned error with the method and parameters.
func (c *NormalizingCaller) CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	if c == nil || c.caller == nil {
		return NormalizeRPCError(errors.New("RPC caller is not initialized"), method, args...)
	}
	if contextual, ok := c.caller.(contextCaller); ok {
		if err := contextual.CallContext(ctx, result, method, args...); err != nil {
			return NormalizeRPCError(err, method, args...)
		}
		return nil
	}
	select {
	case <-ctx.Done():
		return NormalizeRPCError(ctx.Err(), method, args...)
	default:
		return c.Call(result, method, args...)
	}
}

// RPCError is a normalized JSON-RPC or transport failure.
//
// Code, Message, and Data preserve the node error. Method and Parameters record
// the request that failed. Non-JSON-RPC failures use code -1. Cause retains the
// underlying Go error for errors.Is and errors.As, but is not serialized.
type RPCError struct {
	Code       int           `json:"code"`
	Message    string        `json:"message"`
	Data       interface{}   `json:"data,omitempty"`
	Method     string        `json:"method"`
	Parameters []interface{} `json:"parameters"`
	Cause      error         `json:"-"`
}

// Error returns the normalized RPC error message.
func (e *RPCError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Message
}

// Unwrap returns the underlying transport or JSON-RPC error, when available.
func (e *RPCError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type errorCoder interface {
	ErrorCode() int
}

type dataError interface {
	ErrorData() interface{}
}

// NormalizeRPCError converts err into a complete [RPCError].
//
// Parameters:
//   - err: Node, JSON-RPC, or transport error. Nil returns nil.
//   - method: Wire method associated with the failed request.
//   - parameters: Positional request parameters in their original order.
//
// The returned error preserves ErrorCode and ErrorData interfaces exposed by
// go-zenon. Missing codes use -1 and empty messages use "Unknown error
// occurred". Existing RPCError values are copied so request context can be
// safely attached without mutating shared errors.
//
// Example:
//
//	rpcErr := transport.NormalizeRPCError(err, "ledger.getFrontierMomentum")
//	if rpcErr != nil {
//		log.Printf("RPC %s failed (%d): %s", rpcErr.Method, rpcErr.Code, rpcErr.Message)
//	}
func NormalizeRPCError(err error, method string, parameters ...interface{}) *RPCError {
	if err == nil {
		return nil
	}
	result := &RPCError{Code: -1, Message: err.Error(), Method: method, Parameters: append([]interface{}(nil), parameters...), Cause: err}
	var existing *RPCError
	if errors.As(err, &existing) {
		result.Code = existing.Code
		result.Message = existing.Message
		result.Data = existing.Data
		if result.Method == "" {
			result.Method = existing.Method
		}
		if parameters == nil {
			result.Parameters = append([]interface{}(nil), existing.Parameters...)
		}
	}
	var coded errorCoder
	if errors.As(err, &coded) {
		result.Code = coded.ErrorCode()
	}
	var withData dataError
	if errors.As(err, &withData) {
		result.Data = withData.ErrorData()
	}
	if result.Message == "" {
		result.Message = "Unknown error occurred"
	}
	if result.Parameters == nil {
		result.Parameters = []interface{}{}
	}
	return result
}

// Request is a JSON-RPC 2.0 request with positional parameters.
type Request struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// NewRequest creates a JSON-RPC 2.0 request using positional parameters.
//
// Parameters:
//   - id: Request identifier returned by the server response.
//   - method: Fully qualified wire method.
//   - parameters: Positional method parameters in wire order.
//
// NewRequest always emits a non-nil params array, including for zero-argument
// methods.
func NewRequest(id interface{}, method string, parameters ...interface{}) Request {
	params := append([]interface{}{}, parameters...)
	return Request{JSONRPC: "2.0", ID: id, Method: method, Params: params}
}

// SubscriptionParams builds ledger.subscribe positional parameters.
//
// The topic is always first, followed by its topic-specific arguments.
func SubscriptionParams(topic string, arguments ...interface{}) []interface{} {
	params := make([]interface{}, 1, len(arguments)+1)
	params[0] = topic
	return append(params, arguments...)
}

// SubscriptionEvent is a normalized ledger.subscription notification.
type SubscriptionEvent struct {
	SubscriptionID string        `json:"subscriptionId"`
	Updates        []interface{} `json:"updates"`
}

type subscriptionNotification struct {
	Subscription string            `json:"subscription"`
	Result       []json.RawMessage `json:"result"`
}

// NormalizeSubscriptionNotification validates and normalizes an incoming
// ledger.subscription notification.
//
// Parameters:
//   - method: Notification method; must equal "ledger.subscription".
//   - params: Raw notification params containing subscription and result.
//
// The returned event preserves the opaque subscription ID and decodes each
// update as a JSON value. Malformed methods, IDs, or result arrays return an
// error.
func NormalizeSubscriptionNotification(method string, params json.RawMessage) (SubscriptionEvent, error) {
	if method != "ledger.subscription" {
		return SubscriptionEvent{}, fmt.Errorf("unexpected subscription method %q", method)
	}
	var notification subscriptionNotification
	if err := json.Unmarshal(params, &notification); err != nil {
		return SubscriptionEvent{}, fmt.Errorf("invalid subscription notification: %w", err)
	}
	if notification.Subscription == "" {
		return SubscriptionEvent{}, fmt.Errorf("invalid subscription notification: missing subscription ID")
	}
	event := SubscriptionEvent{SubscriptionID: notification.Subscription, Updates: make([]interface{}, len(notification.Result))}
	for index, raw := range notification.Result {
		if err := json.Unmarshal(raw, &event.Updates[index]); err != nil {
			return SubscriptionEvent{}, fmt.Errorf("invalid subscription update %d: %w", index, err)
		}
	}
	return event, nil
}
