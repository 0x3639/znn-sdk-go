package embedded

import (
	"math/big"

	"github.com/0x3639/znn-sdk-go/internal/rpcvalidation"
	"github.com/0x3639/znn-sdk-go/transport"
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/constants"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

type AcceleratorApi struct {
	client transport.Caller
}

func NewAcceleratorApi(client transport.Caller) *AcceleratorApi {
	return &AcceleratorApi{
		client: client,
	}
}

func (aa *AcceleratorApi) GetAll(pageIndex, pageSize uint32) (*ProjectList, error) {
	if err := rpcvalidation.ValidateLimit("embedded.accelerator.getAll", "pageSize", uint64(pageSize), rpcvalidation.MaxPageSize); err != nil {
		return nil, err
	}
	ans := new(ProjectList)
	if err := aa.client.Call(ans, "embedded.accelerator.getAll", pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (aa *AcceleratorApi) GetProjectById(id types.Hash) (*Project, error) {
	ans := new(Project)
	if err := aa.client.Call(ans, "embedded.accelerator.getProjectById", id.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

func (aa *AcceleratorApi) GetPhaseById(id types.Hash) (*Phase, error) {
	ans := new(Phase)
	if err := aa.client.Call(ans, "embedded.accelerator.getPhaseById", id.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

func (aa *AcceleratorApi) GetVoteBreakdown(id types.Hash) (*VoteBreakdown, error) {
	ans := new(VoteBreakdown)
	if err := aa.client.Call(ans, "embedded.accelerator.getVoteBreakdown", id.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

func (aa *AcceleratorApi) GetPillarVotes(name string, hashes []types.Hash) ([]*definition.PillarVote, error) {
	var ans []*definition.PillarVote
	if err := aa.client.Call(&ans, "embedded.accelerator.getPillarVotes", name, hashes); err != nil {
		return nil, err
	}
	return ans, nil
}

// CreateProject creates a transaction template to submit a new Accelerator-Z project proposal.
//
// Accelerator-Z is Zenon's decentralized funding mechanism for ecosystem development.
// Projects submit proposals requesting ZNN/QSR funding, which Pillars vote on.
//
// Requirements:
//   - Cost: 1 ZNN (project creation fee, non-refundable)
//   - Project must include clear deliverables and milestones
//   - Funding delivered in phases as milestones are completed
//
// Parameters:
//   - name: Project name (3-50 characters)
//   - description: Detailed project description
//   - url: Project website or documentation URL
//   - znnFundsNeeded: Total ZNN requested
//   - qsrFundsNeeded: Total QSR requested
//
// Returns an unsigned AccountBlock template ready for processing.
//
// Example:
//
//	znnNeeded := big.NewInt(5000 * 100000000) // 5,000 ZNN
//	qsrNeeded := big.NewInt(50000 * 100000000) // 50,000 QSR
//
//	template := client.AcceleratorApi.CreateProject(
//	    "My Zenon Project",
//	    "Building a new tool for the ecosystem...",
//	    "https://myproject.com",
//	    znnNeeded,
//	    qsrNeeded,
//	)
//
// Note: After submission, Pillars vote on the project. If approved, add phases with milestones.
func (aa *AcceleratorApi) CreateProject(name, description, url string, znnFundsNeeded, qsrFundsNeeded *big.Int) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.AcceleratorContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        constants.ProjectCreationAmount,
		Data: definition.ABIAccelerator.PackMethodPanic(
			definition.CreateProjectMethodName,
			name,
			description,
			url,
			znnFundsNeeded,
			qsrFundsNeeded,
		),
	}
}

func (aa *AcceleratorApi) AddPhase(id types.Hash, name, description, url string, znnFundsNeeded, qsrFundsNeeded *big.Int) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.AcceleratorContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIAccelerator.PackMethodPanic(
			definition.AddPhaseMethodName,
			id,
			name,
			description,
			url,
			znnFundsNeeded,
			qsrFundsNeeded,
		),
	}
}

func (aa *AcceleratorApi) UpdatePhase(id types.Hash, name, description, url string, znnFundsNeeded, qsrFundsNeeded *big.Int) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.AcceleratorContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIAccelerator.PackMethodPanic(
			definition.UpdatePhaseMethodName,
			id,
			name,
			description,
			url,
			znnFundsNeeded,
			qsrFundsNeeded,
		),
	}
}

func (aa *AcceleratorApi) Donate(amount *big.Int, tokenStandard types.ZenonTokenStandard) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.AcceleratorContract,
		TokenStandard: tokenStandard,
		Amount:        amount,
		Data:          definition.ABIAccelerator.PackMethodPanic(definition.DonateMethodName),
	}
}

// Accelerator-Z vote choices. These alias go-zenon's on-chain definitions
// (github.com/zenon-network/go-zenon/vm/embedded/definition) so callers never
// have to hardcode the magic numbers: the contract tallies a vote of 0 as
// "yes", 1 as "no", and 2 as "abstain". Pass one of these as the vote argument
// to VoteByName / VoteByProducerAddress.
const (
	VoteYes     uint8 = definition.VoteYes     // 0 — approve
	VoteNo      uint8 = definition.VoteNo      // 1 — reject
	VoteAbstain uint8 = definition.VoteAbstain // 2 — abstain
)

// VoteByName creates a transaction template for a Pillar to vote on a project/phase.
//
// Only Pillar operators can vote on Accelerator proposals. Votes determine whether
// projects receive funding and whether phases are approved for payment.
//
// Vote options (see the VoteYes/VoteNo/VoteAbstain constants):
//   - 0: Yes (approve)
//   - 1: No (reject)
//   - 2: Abstain
//
// Parameters:
//   - id: Project or phase ID to vote on
//   - pillarName: Name of the voting Pillar
//   - vote: Vote choice (0=yes, 1=no, 2=abstain)
//
// Returns an unsigned AccountBlock template ready for processing.
//
// Example:
//
//	projectId := types.HexToHashPanic("0x123...")
//	template := client.AcceleratorApi.VoteByName(projectId, "MyPillar", embedded.VoteYes)
//
// Note: Only Pillar owners can call this. Voting period has time limits.
func (aa *AcceleratorApi) VoteByName(id types.Hash, pillarName string, vote uint8) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.AcceleratorContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIAccelerator.PackMethodPanic(
			definition.VoteByNameMethodName,
			id,
			pillarName,
			vote,
		),
	}
}

// VoteByProducerAddress creates a transaction template for a Pillar to vote on a
// project/phase using the calling account's producer address (rather than the
// Pillar name). Only Pillar operators can vote.
//
// vote takes the same values as VoteByName: 0=yes, 1=no, 2=abstain (see the
// VoteYes/VoteNo/VoteAbstain constants).
func (aa *AcceleratorApi) VoteByProducerAddress(id types.Hash, vote uint8) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.AcceleratorContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIAccelerator.PackMethodPanic(
			definition.VoteByProdAddressMethodName,
			id,
			vote,
		),
	}
}
