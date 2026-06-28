package embedded

import (
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/server"
	"github.com/zenon-network/go-zenon/vm/constants"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

// GovernanceApi exposes the on-chain governance contract introduced by the
// governance-ratchet spork. The governance contract lets pillars propose
// protocol actions (typically privileged calls into other embedded contracts
// such as Spork, Bridge, or Liquidity), vote on them across escalating rounds,
// and execute approved actions.
//
// Construct it through rpc_client.RpcClient.GovernanceApi rather than directly:
//
//	client, _ := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
//	actions, _ := client.GovernanceApi.GetAllActions(0, 25)
//
// All read methods call the node's "embedded.governance.*" RPC namespace.
// All transaction builders return unsigned *nom.AccountBlock templates that
// must be signed and published (e.g. via zenon.Zenon.Send).
type GovernanceApi struct {
	client *server.Client
}

// NewGovernanceApi creates a GovernanceApi bound to the given RPC client.
//
// Most callers obtain a ready-to-use instance via RpcClient.GovernanceApi and
// do not call this directly.
func NewGovernanceApi(client *server.Client) *GovernanceApi {
	return &GovernanceApi{
		client: client,
	}
}

// GetAllActions returns a paginated list of governance actions.
//
// Actions are returned newest-to-oldest as stored by the node. Each Action
// includes its current voting round, status, the per-round thresholds and
// voting period derived from the action's type, and the current VoteBreakdown.
//
// Parameters:
//   - pageIndex: Page number (0-indexed)
//   - pageSize: Number of actions per page (must not exceed the node's
//     RpcMaxPageSize, currently 50)
//
// Returns the populated ActionList or an error if the RPC call fails.
//
// Example:
//
//	list, err := client.GovernanceApi.GetAllActions(0, 25)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, a := range list.List {
//	    fmt.Printf("%s: status=%d votes=%d/%d\n", a.Name, a.Status, a.Votes.Yes, a.Votes.Total)
//	}
func (g *GovernanceApi) GetAllActions(pageIndex, pageSize uint32) (*ActionList, error) {
	ans := new(ActionList)
	if err := g.client.Call(ans, "embedded.governance.getAllActions", pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

// GetActionById returns a single governance action by its hash identifier.
//
// The id is the hash of the originating ProposeAction send block, as reported
// in Action.Id by GetAllActions.
//
// Parameters:
//   - id: Hash identifier of the action
//
// Returns the Action or an error if it does not exist or the RPC call fails.
//
// Example:
//
//	id := types.HexToHashPanic("0x...")
//	action, err := client.GovernanceApi.GetActionById(id)
func (g *GovernanceApi) GetActionById(id types.Hash) (*Action, error) {
	ans := new(Action)
	if err := g.client.Call(ans, "embedded.governance.getActionById", id.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

// Contract call templates

// ProposeAction creates a transaction template that proposes a new governance
// action.
//
// A governance action wraps a privileged contract call: when the action is
// approved by pillar voting and executed, the governance contract issues a
// contract-send to destination carrying the decoded data as its block data.
//
// The data argument is the base64-encoded ABI call data for the destination
// contract (standard base64, as required by the node). Rather than packing this
// by hand, build it with the payload helpers (for example
// PayloadSporkActivate, PayloadBridgeSetTokenPair) which return a
// ProposalPayload containing both the Destination and the encoded Data.
//
// Requirements:
//   - Cost: exactly 1 ZNN (constants.ProjectCreationAmount), non-refundable
//   - Token: ZNN
//
// Parameters:
//   - name: Short action name
//   - description: Human-readable description of the action
//   - url: Reference URL with additional detail
//   - destination: Embedded contract the action will call when executed
//   - data: Base64-encoded ABI call data for destination (see ProposalPayload)
//
// Returns an unsigned AccountBlock template ready for signing and publishing.
//
// Example:
//
//	payload := client.GovernanceApi.PayloadSporkActivate(sporkId)
//	template := client.GovernanceApi.ProposeAction(
//	    "Activate governance spork",
//	    "Enable the governance ratchet on testnet",
//	    "https://forum.zenon.org/...",
//	    payload.Destination,
//	    payload.Data,
//	)
func (g *GovernanceApi) ProposeAction(name, description, url string, destination types.Address, data string) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.GovernanceContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        constants.ProjectCreationAmount,
		Data: definition.ABIGovernance.PackMethodPanic(
			definition.ProposeActionMethodName,
			name,
			description,
			url,
			destination,
			data,
		),
	}
}

// ExecuteAction creates a transaction template that advances or finalizes a
// governance action.
//
// ExecuteAction is permissionless and idempotent: anyone may call it to push an
// action through its lifecycle. Depending on the current vote tally and timing
// it tallies votes, advances the action to the next voting round, marks it
// rejected/approved/no-decision, or — when approved — triggers the underlying
// contract call on the action's destination.
//
// Parameters:
//   - id: Hash identifier of the action to execute
//
// Returns an unsigned AccountBlock template (0 ZNN) ready for signing and
// publishing.
func (g *GovernanceApi) ExecuteAction(id types.Hash) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.GovernanceContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIGovernance.PackMethodPanic(
			definition.ExecuteActionMethodName,
			id,
		),
	}
}

// VoteByName creates a transaction template for a pillar to vote on a
// governance action, identifying the pillar by name.
//
// Only pillar operators may vote, and the vote is recorded against the action's
// current voting round.
//
// Vote options (see the exported constants):
//   - VoteYes (0): approve
//   - VoteNo (1): reject
//   - VoteAbstain (2): abstain
//
// Parameters:
//   - id: Hash identifier of the action
//   - pillarName: Name of the voting pillar
//   - vote: Vote choice (VoteYes/VoteNo/VoteAbstain)
//
// Returns an unsigned AccountBlock template (0 ZNN) ready for signing and
// publishing.
func (g *GovernanceApi) VoteByName(id types.Hash, pillarName string, vote uint8) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.GovernanceContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIGovernance.PackMethodPanic(
			definition.VoteByNameMethodName,
			id,
			pillarName,
			vote,
		),
	}
}

// VoteByProducerAddress creates a transaction template for a pillar to vote on
// a governance action, identifying the pillar by the calling producer address.
//
// This is the address-based counterpart to VoteByName: the node resolves the
// pillar from the producing address of the signing key, so no pillar name is
// supplied.
//
// Vote options:
//   - VoteYes (0): approve
//   - VoteNo (1): reject
//   - VoteAbstain (2): abstain
//
// Parameters:
//   - id: Hash identifier of the action
//   - vote: Vote choice (VoteYes/VoteNo/VoteAbstain)
//
// Returns an unsigned AccountBlock template (0 ZNN) ready for signing and
// publishing.
func (g *GovernanceApi) VoteByProducerAddress(id types.Hash, vote uint8) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.GovernanceContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIGovernance.PackMethodPanic(
			definition.VoteByProdAddressMethodName,
			id,
			vote,
		),
	}
}
