package rpc_client

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/0x3639/znn-sdk-go/transport"
)

// swappableCaller is a stable transport.Caller whose underlying normalizing
// caller can be swapped atomically on reconnect.
//
// API objects (LedgerApi, embedded APIs, ...) are created once with a single
// swappableCaller and never reassigned, so callers can read the exported
// RpcClient API fields concurrently with a reconnect without a data race. The
// reconnect only stores a new inner caller here.
type swappableCaller struct {
	current atomic.Pointer[transport.NormalizingCaller]
}

// set installs the caller used for subsequent requests. A nil caller makes
// calls return a normalized "not connected" error.
func (s *swappableCaller) set(caller *transport.NormalizingCaller) {
	s.current.Store(caller)
}

// Call performs a positional JSON-RPC request through the current caller.
func (s *swappableCaller) Call(result interface{}, method string, args ...interface{}) error {
	caller := s.current.Load()
	if caller == nil {
		return transport.NormalizeRPCError(errors.New("RPC client is not connected"), method, args...)
	}
	return caller.Call(result, method, args...)
}

// CallContext performs a positional JSON-RPC request with cancellation through
// the current caller.
func (s *swappableCaller) CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	caller := s.current.Load()
	if caller == nil {
		return transport.NormalizeRPCError(errors.New("RPC client is not connected"), method, args...)
	}
	return caller.CallContext(ctx, result, method, args...)
}
