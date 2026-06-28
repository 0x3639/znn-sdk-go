package embedded

import (
	"encoding/base64"
	"math/big"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
)

// ProposalPayload is the destination/data pair needed to propose a governance
// action that calls another embedded contract.
//
// Pass its fields straight into GovernanceApi.ProposeAction:
//
//	p := client.GovernanceApi.PayloadSporkActivate(sporkId)
//	block := client.GovernanceApi.ProposeAction(name, desc, url, p.Destination, p.Data)
//
// Fields:
//   - Destination: The embedded contract the action will call when executed
//   - Data: Standard base64-encoded ABI call data for Destination
type ProposalPayload struct {
	Destination types.Address
	Data        string
}

// EncodeProposalPayload converts any transaction template into a
// ProposalPayload by taking its destination address and standard-base64
// encoding its ABI call data.
//
// This is the primitive behind every typed Payload* helper. Use it directly to
// wrap a template that doesn't have a dedicated helper:
//
//	tmpl := client.TokenApi.UpdateToken(zts, owner, isMintable, isBurnable)
//	p := embedded.EncodeProposalPayload(tmpl)
//	block := client.GovernanceApi.ProposeAction(name, desc, url, p.Destination, p.Data)
//
// The node decodes Data with standard base64 (not URL-safe) when executing the
// action, so EncodeProposalPayload uses base64.StdEncoding to match.
func EncodeProposalPayload(block *nom.AccountBlock) ProposalPayload {
	return ProposalPayload{
		Destination: block.ToAddress,
		Data:        base64.StdEncoding.EncodeToString(block.Data),
	}
}

// --- Spork payload helpers ---

// PayloadSporkCreate builds the payload for a governance action that creates a
// new spork.
func (g *GovernanceApi) PayloadSporkCreate(name, description string) ProposalPayload {
	return EncodeProposalPayload(NewSporkApi(g.client).CreateSpork(name, description))
}

// PayloadSporkActivate builds the payload for a governance action that
// activates an existing spork.
func (g *GovernanceApi) PayloadSporkActivate(id types.Hash) ProposalPayload {
	return EncodeProposalPayload(NewSporkApi(g.client).ActivateSpork(id))
}

// --- Bridge payload helpers ---

// PayloadBridgeAddNetwork builds the payload for a governance action that adds
// (or updates) a bridge network.
func (g *GovernanceApi) PayloadBridgeAddNetwork(networkClass, chainId uint32, name, contractAddress, metadata string) ProposalPayload {
	return EncodeProposalPayload(NewBridgeApi(g.client).AddNetwork(networkClass, chainId, name, contractAddress, metadata))
}

// PayloadBridgeRemoveNetwork builds the payload for a governance action that
// removes a bridge network.
func (g *GovernanceApi) PayloadBridgeRemoveNetwork(networkClass, chainId uint32) ProposalPayload {
	return EncodeProposalPayload(NewBridgeApi(g.client).RemoveNetwork(networkClass, chainId))
}

// PayloadBridgeSetTokenPair builds the payload for a governance action that
// sets a bridge token pair.
func (g *GovernanceApi) PayloadBridgeSetTokenPair(networkClass, chainId uint32, tokenStandard types.ZenonTokenStandard, tokenAddress string, bridgeable, redeemable, owned bool, minAmount *big.Int, fee, redeemDelay uint32, metadata string) ProposalPayload {
	return EncodeProposalPayload(NewBridgeApi(g.client).SetTokenPair(networkClass, chainId, tokenStandard, tokenAddress, bridgeable, redeemable, owned, minAmount, fee, redeemDelay, metadata))
}

// PayloadBridgeRemoveTokenPair builds the payload for a governance action that
// removes a bridge token pair.
func (g *GovernanceApi) PayloadBridgeRemoveTokenPair(networkClass, chainId uint32, tokenStandard types.ZenonTokenStandard, tokenAddress string) ProposalPayload {
	return EncodeProposalPayload(NewBridgeApi(g.client).RemoveTokenPair(networkClass, chainId, tokenStandard, tokenAddress))
}

// PayloadBridgeHalt builds the payload for a governance action that halts the
// bridge.
func (g *GovernanceApi) PayloadBridgeHalt(signature string) ProposalPayload {
	return EncodeProposalPayload(NewBridgeApi(g.client).Halt(signature))
}

// PayloadBridgeUnhalt builds the payload for a governance action that unhalts
// the bridge.
func (g *GovernanceApi) PayloadBridgeUnhalt() ProposalPayload {
	return EncodeProposalPayload(NewBridgeApi(g.client).Unhalt())
}

// PayloadBridgeEmergency builds the payload for a governance action that
// triggers the bridge emergency halt.
func (g *GovernanceApi) PayloadBridgeEmergency() ProposalPayload {
	return EncodeProposalPayload(NewBridgeApi(g.client).Emergency())
}

// PayloadBridgeChangeAdministrator builds the payload for a governance action
// that changes the bridge administrator.
func (g *GovernanceApi) PayloadBridgeChangeAdministrator(administrator types.Address) ProposalPayload {
	return EncodeProposalPayload(NewBridgeApi(g.client).ChangeAdministrator(administrator))
}

// PayloadBridgeSetAllowKeygen builds the payload for a governance action that
// toggles the bridge allow-keygen flag.
func (g *GovernanceApi) PayloadBridgeSetAllowKeygen(allowKeygen bool) ProposalPayload {
	return EncodeProposalPayload(NewBridgeApi(g.client).SetAllowKeygen(allowKeygen))
}

// PayloadBridgeSetOrchestratorInfo builds the payload for a governance action
// that updates the bridge orchestrator parameters.
func (g *GovernanceApi) PayloadBridgeSetOrchestratorInfo(windowSize uint64, keyGenThreshold, confirmationsToFinality, estimatedMomentumTime uint32) ProposalPayload {
	return EncodeProposalPayload(NewBridgeApi(g.client).SetOrchestratorInfo(windowSize, keyGenThreshold, confirmationsToFinality, estimatedMomentumTime))
}

// PayloadBridgeSetMetadata builds the payload for a governance action that sets
// the bridge metadata.
func (g *GovernanceApi) PayloadBridgeSetMetadata(metadata string) ProposalPayload {
	return EncodeProposalPayload(NewBridgeApi(g.client).SetBridgeMetadata(metadata))
}

// PayloadBridgeSetNetworkMetadata builds the payload for a governance action
// that sets a bridge network's metadata.
func (g *GovernanceApi) PayloadBridgeSetNetworkMetadata(networkClass, chainId uint32, metadata string) ProposalPayload {
	return EncodeProposalPayload(NewBridgeApi(g.client).SetNetworkMetadata(networkClass, chainId, metadata))
}

// PayloadBridgeRevokeUnwrapRequest builds the payload for a governance action
// that revokes an unwrap request.
func (g *GovernanceApi) PayloadBridgeRevokeUnwrapRequest(transactionHash types.Hash, logIndex uint32) ProposalPayload {
	return EncodeProposalPayload(NewBridgeApi(g.client).RevokeUnwrapRequest(transactionHash, logIndex))
}

// PayloadBridgeNominateGuardians builds the payload for a governance action
// that nominates bridge guardians.
func (g *GovernanceApi) PayloadBridgeNominateGuardians(guardians []types.Address) ProposalPayload {
	return EncodeProposalPayload(NewBridgeApi(g.client).NominateGuardians(guardians))
}

// --- Liquidity payload helpers ---

// PayloadLiquidityFund builds the payload for a governance action that funds
// the liquidity reward pool.
func (g *GovernanceApi) PayloadLiquidityFund(znnReward, qsrReward *big.Int) ProposalPayload {
	return EncodeProposalPayload(NewLiquidityApi(g.client).Fund(znnReward, qsrReward))
}

// PayloadLiquidityBurnZnn builds the payload for a governance action that burns
// ZNN held by the liquidity contract.
func (g *GovernanceApi) PayloadLiquidityBurnZnn(burnAmount *big.Int) ProposalPayload {
	return EncodeProposalPayload(NewLiquidityApi(g.client).BurnZnn(burnAmount))
}

// PayloadLiquiditySetTokenTuple builds the payload for a governance action that
// sets the liquidity token tuple configuration.
func (g *GovernanceApi) PayloadLiquiditySetTokenTuple(tokenStandards []string, znnPercentages, qsrPercentages []uint32, minAmounts []*big.Int) ProposalPayload {
	return EncodeProposalPayload(NewLiquidityApi(g.client).SetTokenTupleMethod(tokenStandards, znnPercentages, qsrPercentages, minAmounts))
}

// PayloadLiquiditySetIsHalted builds the payload for a governance action that
// toggles the liquidity halt flag.
func (g *GovernanceApi) PayloadLiquiditySetIsHalted(value bool) ProposalPayload {
	return EncodeProposalPayload(NewLiquidityApi(g.client).SetIsHalted(value))
}

// PayloadLiquidityUnlockStakeEntries builds the payload for a governance action
// that unlocks liquidity stake entries for a token standard.
func (g *GovernanceApi) PayloadLiquidityUnlockStakeEntries(zts types.ZenonTokenStandard) ProposalPayload {
	return EncodeProposalPayload(NewLiquidityApi(g.client).UnlockLiquidityStakeEntries(zts))
}

// PayloadLiquiditySetAdditionalReward builds the payload for a governance
// action that sets the liquidity additional reward.
func (g *GovernanceApi) PayloadLiquiditySetAdditionalReward(znnReward, qsrAmount *big.Int) ProposalPayload {
	return EncodeProposalPayload(NewLiquidityApi(g.client).SetAdditionalReward(znnReward, qsrAmount))
}

// PayloadLiquidityChangeAdministrator builds the payload for a governance
// action that changes the liquidity administrator.
func (g *GovernanceApi) PayloadLiquidityChangeAdministrator(administrator types.Address) ProposalPayload {
	return EncodeProposalPayload(NewLiquidityApi(g.client).ChangeAdministrator(administrator))
}

// PayloadLiquidityNominateGuardians builds the payload for a governance action
// that nominates liquidity guardians.
func (g *GovernanceApi) PayloadLiquidityNominateGuardians(guardians []types.Address) ProposalPayload {
	return EncodeProposalPayload(NewLiquidityApi(g.client).NominateGuardians(guardians))
}

// PayloadLiquidityEmergency builds the payload for a governance action that
// triggers the liquidity emergency halt.
func (g *GovernanceApi) PayloadLiquidityEmergency() ProposalPayload {
	return EncodeProposalPayload(NewLiquidityApi(g.client).Emergency())
}
