package embedded

import (
	"github.com/zenon-network/go-zenon/common/types"
)

// Spork represents a protocol spork configuration.
//
// Sporks are protocol upgrade mechanisms that allow the network to activate
// new features or changes at a specific block height. Once activated, nodes
// must comply with the new rules after the enforcement height.
//
// Fields:
//   - Id: Unique identifier for this spork
//   - Name: Human-readable name of the spork
//   - Description: Detailed description of what this spork enables
//   - Activated: Whether the spork has been activated
//   - EnforcementHeight: Momentum height when the spork becomes mandatory
//
// Spork Lifecycle:
//  1. Spork is created but not activated
//  2. Spork is activated by governance
//  3. After EnforcementHeight, all nodes must support the new feature
type Spork struct {
	Id                types.Hash `json:"id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	Activated         bool       `json:"activated"`
	EnforcementHeight uint64     `json:"enforcementHeight"`
}

// SporkList represents a paginated list of sporks.
//
// Fields:
//   - Count: Total number of sporks matching the query
//   - List: Slice of Spork entries for the current page
type SporkList struct {
	Count int      `json:"count"`
	List  []*Spork `json:"list"`
}
