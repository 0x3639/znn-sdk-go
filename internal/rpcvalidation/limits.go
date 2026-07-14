package rpcvalidation

import "fmt"

const (
	// MaxPageSize is the canonical maximum page size or count for ordinary
	// Zenon JSON-RPC list endpoints.
	MaxPageSize uint64 = 1024

	// MemoryPoolPageSize is the canonical maximum for unconfirmed, unreceived,
	// and liquidity-stake memory-pool style endpoints.
	MemoryPoolPageSize uint64 = 50
)

// ValidateLimit checks one page-size or count argument against its endpoint maximum.
//
// Parameters:
//   - method: Canonical JSON-RPC method name used to identify the invalid request.
//   - parameter: Argument name, normally "pageSize" or "count".
//   - value: Caller-supplied unsigned value.
//   - maximum: Inclusive endpoint maximum from the stable SDK specification.
//
// ValidateLimit returns nil when value is at most maximum. Otherwise it returns
// an error containing the method, argument, supplied value, and maximum. API
// methods call this helper before invoking their transport, so an invalid
// request cannot reach a node.
//
// Example:
//
//	err := rpcvalidation.ValidateLimit("ledger.getMomentumsByPage", "pageSize", 1025, rpcvalidation.MaxPageSize)
//	if err != nil {
//		// Reject the request without making an RPC call.
//	}
func ValidateLimit(method, parameter string, value, maximum uint64) error {
	if value <= maximum {
		return nil
	}
	return fmt.Errorf("%s: %s %d exceeds maximum %d", method, parameter, value, maximum)
}
