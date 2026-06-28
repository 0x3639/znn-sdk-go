package embedded

import (
	"encoding/base64"

	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/constants"
)

// Governance reuses the package-level vote constants VoteYes, VoteNo, and
// VoteAbstain (declared in accelerator.go, aliasing go-zenon's definition.Vote*
// values 0/1/2). Pass one of those as the vote argument to VoteByName /
// VoteByProducerAddress.

// Governance action types and statuses. These mirror go-zenon's
// vm/constants values, which are declared as package variables (not untyped
// constants), so they are re-exported here as variables.
//
// The node assigns the type automatically when an action is proposed: actions
// targeting the Spork contract are Type1 (stricter, shorter first round),
// everything else is Type2. The type selects the per-round thresholds and
// voting periods.
var (
	// Type1Action denotes a Spork-targeting governance action.
	Type1Action = constants.Type1Action
	// Type2Action denotes a non-Spork governance action.
	Type2Action = constants.Type2Action

	// ActionStatusVoting indicates the action is open for voting.
	ActionStatusVoting = constants.ActionStatusVoting
	// ActionStatusApproved indicates the action was approved and executed.
	ActionStatusApproved = constants.ActionStatusApproved
	// ActionStatusRejected indicates the action was rejected by voting.
	ActionStatusRejected = constants.ActionStatusRejected
	// ActionStatusNoDecision indicates the action expired across all rounds
	// without reaching a decision.
	ActionStatusNoDecision = constants.ActionStatusNoDecision
)

// Action represents a single governance action returned by the
// embedded.governance RPC namespace.
//
// The JSON field names are PascalCase because the node serializes the action
// using Go's default field names (it embeds the on-chain action variable and
// augments it with the computed voting fields below).
//
// Fields:
//   - Id: Hash of the originating ProposeAction send block; the action's identifier
//   - Owner: Address that proposed the action
//   - Name: Short action name
//   - Description: Human-readable description
//   - Url: Reference URL with additional detail
//   - Destination: Embedded contract the action calls when executed
//   - Data: Base64-encoded ABI call data for Destination (see DecodedData)
//   - CreationTimestamp: Unix timestamp when the action was proposed
//   - Type: Action type (Type1Action or Type2Action)
//   - Round: Current 0-indexed voting round
//   - CurrentVoteId: Vote-breakdown id for the current round
//   - RoundStartTimestamp: Unix timestamp when the current round started
//   - Status: Current status (ActionStatus* constant)
//   - Executed: Whether the underlying contract call has been executed
//   - Expired: Whether the current round's voting period has elapsed
//   - ActivePillarThreshold: Quorum (percent of active pillars) for this round
//   - DirectionalThreshold: Yes-share (percent) required to approve this round
//   - VotingPeriod: Length of the current round's voting period, in seconds
//   - Votes: Current round's vote tally
type Action struct {
	Id                    types.Hash     `json:"Id"`
	Owner                 types.Address  `json:"Owner"`
	Name                  string         `json:"Name"`
	Description           string         `json:"Description"`
	Url                   string         `json:"Url"`
	Destination           types.Address  `json:"Destination"`
	Data                  string         `json:"Data"`
	CreationTimestamp     int64          `json:"CreationTimestamp"`
	Type                  uint8          `json:"Type"`
	Round                 uint8          `json:"Round"`
	CurrentVoteId         types.Hash     `json:"CurrentVoteId"`
	RoundStartTimestamp   int64          `json:"RoundStartTimestamp"`
	Status                uint8          `json:"Status"`
	Executed              bool           `json:"Executed"`
	Expired               bool           `json:"Expired"`
	ActivePillarThreshold uint32         `json:"ActivePillarThreshold"`
	DirectionalThreshold  uint32         `json:"DirectionalThreshold"`
	VotingPeriod          int64          `json:"VotingPeriod"`
	Votes                 *VoteBreakdown `json:"Votes"`
}

// DecodedData returns the raw ABI call bytes carried by the action, decoding
// the base64-encoded Data field. These are the bytes the governance contract
// will use as the block data when it calls Destination on execution.
//
// Returns an error if Data is not valid standard base64.
func (a *Action) DecodedData() ([]byte, error) {
	return base64.StdEncoding.DecodeString(a.Data)
}

// ActionList is a paginated list of governance actions, as returned by
// GetAllActions.
//
// Fields:
//   - Count: Total number of actions stored by the node (not just this page)
//   - List: Actions for the requested page
type ActionList struct {
	Count int       `json:"count"`
	List  []*Action `json:"list"`
}
