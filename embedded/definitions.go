package embedded

import "github.com/0x3639/znn-sdk-go/abi"

// =============================================================================
// Contract ABI Definitions (JSON strings)
// =============================================================================

// PlasmaDefinition contains the ABI for the Plasma embedded contract
const PlasmaDefinition = `[
	{"type":"function","name":"Fuse","inputs":[{"name":"address","type":"address"}]},
	{"type":"function","name":"CancelFuse","inputs":[{"name":"id","type":"hash"}]}
]`

// PillarDefinition contains the ABI for the Pillar embedded contract
const PillarDefinition = `[
	{"type":"function","name":"Register","inputs":[{"name":"name","type":"string"},{"name":"producerAddress","type":"address"},{"name":"rewardAddress","type":"address"},{"name":"giveBlockRewardPercentage","type":"uint8"},{"name":"giveDelegateRewardPercentage","type":"uint8"}]},
	{"type":"function","name":"RegisterLegacy","inputs":[{"name":"name","type":"string"},{"name":"producerAddress","type":"address"},{"name":"rewardAddress","type":"address"},{"name":"giveBlockRewardPercentage","type":"uint8"},{"name":"giveDelegateRewardPercentage","type":"uint8"},{"name":"publicKey","type":"string"},{"name":"signature","type":"string"}]},
	{"type":"function","name":"Revoke","inputs":[{"name":"name","type":"string"}]},
	{"type":"function","name":"UpdatePillar","inputs":[{"name":"name","type":"string"},{"name":"producerAddress","type":"address"},{"name":"rewardAddress","type":"address"},{"name":"giveBlockRewardPercentage","type":"uint8"},{"name":"giveDelegateRewardPercentage","type":"uint8"}]},
	{"type":"function","name":"Delegate","inputs":[{"name":"name","type":"string"}]},
	{"type":"function","name":"Undelegate","inputs":[]}
]`

// TokenDefinition contains the ABI for the Token embedded contract
// #nosec G101 -- This is an ABI definition, not hardcoded credentials
const TokenDefinition = `[
	{"type":"function","name":"IssueToken","inputs":[{"name":"tokenName","type":"string"},{"name":"tokenSymbol","type":"string"},{"name":"tokenDomain","type":"string"},{"name":"totalSupply","type":"uint256"},{"name":"maxSupply","type":"uint256"},{"name":"decimals","type":"uint8"},{"name":"isMintable","type":"bool"},{"name":"isBurnable","type":"bool"},{"name":"isUtility","type":"bool"}]},
	{"type":"function","name":"Mint","inputs":[{"name":"tokenStandard","type":"tokenStandard"},{"name":"amount","type":"uint256"},{"name":"receiveAddress","type":"address"}]},
	{"type":"function","name":"Burn","inputs":[]},
	{"type":"function","name":"UpdateToken","inputs":[{"name":"tokenStandard","type":"tokenStandard"},{"name":"owner","type":"address"},{"name":"isMintable","type":"bool"},{"name":"isBurnable","type":"bool"}]}
]`

// SentinelDefinition contains the ABI for the Sentinel embedded contract
const SentinelDefinition = `[
	{"type":"function","name":"Register","inputs":[]},
	{"type":"function","name":"Revoke","inputs":[]}
]`

// SwapDefinition contains the ABI for the Swap embedded contract
const SwapDefinition = `[
	{"type":"function","name":"RetrieveAssets","inputs":[{"name":"publicKey","type":"string"},{"name":"signature","type":"string"}]}
]`

// StakeDefinition contains the ABI for the Stake embedded contract
const StakeDefinition = `[
	{"type":"function","name":"Stake","inputs":[{"name":"durationInSec", "type":"int64"}]},
	{"type":"function","name":"Cancel","inputs":[{"name":"id","type":"hash"}]}
]`

// AcceleratorDefinition contains the ABI for the Accelerator embedded contract
const AcceleratorDefinition = `[
	{"type":"function","name":"CreateProject", "inputs":[{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"url","type":"string"},{"name":"znnFundsNeeded","type":"uint256"},{"name":"qsrFundsNeeded","type":"uint256"}]},
	{"type":"function","name":"AddPhase", "inputs":[{"name":"id","type":"hash"},{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"url","type":"string"},{"name":"znnFundsNeeded","type":"uint256"},{"name":"qsrFundsNeeded","type":"uint256"}]},
	{"type":"function","name":"UpdatePhase", "inputs":[{"name":"id","type":"hash"},{"name":"name","type":"string"},{"name":"description","type":"string"},{"name":"url","type":"string"},{"name":"znnFundsNeeded","type":"uint256"},{"name":"qsrFundsNeeded","type":"uint256"}]},
	{"type":"function","name":"Donate", "inputs":[]},
	{"type":"function","name":"VoteByName","inputs":[{"name":"id","type":"hash"},{"name":"name","type":"string"},{"name":"vote","type":"uint8"}]},
	{"type":"function","name":"VoteByProdAddress","inputs":[{"name":"id","type":"hash"},{"name":"vote","type":"uint8"}]}
]`

// SporkDefinition contains the ABI for the Spork embedded contract
const SporkDefinition = `[
	{"type":"function","name":"CreateSpork","inputs":[{"name":"name","type":"string"},{"name":"description","type":"string"}]},
	{"type":"function","name":"ActivateSpork","inputs":[{"name":"id","type":"hash"}]}
]`

// HtlcDefinition contains the ABI for the HTLC embedded contract
const HtlcDefinition = `[
	{"type":"function","name":"Create", "inputs":[{"name":"hashLocked","type":"address"},{"name":"expirationTime","type":"int64"},{"name":"hashType","type":"uint8"},{"name":"keyMaxSize","type":"uint8"},{"name":"hashLock","type":"bytes"}]},
	{"type":"function","name":"Reclaim","inputs":[{"name":"id","type":"hash"}]},
	{"type":"function","name":"Unlock","inputs":[{"name":"id","type":"hash"},{"name":"preimage","type":"bytes"}]},
	{"type":"function","name":"DenyProxyUnlock","inputs":[]},
	{"type":"function","name":"AllowProxyUnlock","inputs":[]}
]`

// BridgeDefinition contains the ABI for the Bridge embedded contract
const BridgeDefinition = `[
	{"type":"function","name":"WrapToken","inputs":[{"name":"networkClass","type":"uint32"},{"name":"chainId","type":"uint32"},{"name":"toAddress","type":"string"}]},
	{"type":"function","name":"UpdateWrapRequest","inputs":[{"name":"id","type":"hash"},{"name":"signature","type":"string"}]},
	{"type":"function","name":"SetNetwork","inputs":[{"name":"networkClass","type":"uint32"},{"name":"chainId","type":"uint32"},{"name":"name","type":"string"},{"name":"contractAddress","type":"string"},{"name":"metadata","type":"string"}]},
	{"type":"function","name":"RemoveNetwork","inputs":[{"name":"networkClass","type":"uint32"},{"name":"chainId","type":"uint32"}]},
	{"type":"function","name":"SetTokenPair","inputs":[{"name":"networkClass","type":"uint32"},{"name":"chainId","type":"uint32"},{"name":"tokenStandard","type":"tokenStandard"},{"name":"tokenAddress","type":"string"},{"name":"bridgeable","type":"bool"},{"name":"redeemable","type":"bool"},{"name":"owned","type":"bool"},{"name":"minAmount","type":"uint256"},{"name":"feePercentage","type":"uint32"},{"name":"redeemDelay","type":"uint32"},{"name":"metadata","type":"string"}]},
	{"type":"function","name":"SetNetworkMetadata","inputs":[{"name":"networkClass","type":"uint32"},{"name":"chainId","type":"uint32"},{"name":"metadata","type":"string"}]},
	{"type":"function","name":"RemoveTokenPair","inputs":[{"name":"networkClass","type":"uint32"},{"name":"chainId","type":"uint32"},{"name":"tokenStandard","type":"tokenStandard"},{"name":"tokenAddress","type":"string"}]},
	{"type":"function","name":"Halt","inputs":[{"name":"signature","type":"string"}]},
	{"type":"function","name":"Unhalt","inputs":[]},
	{"type":"function","name":"Emergency","inputs":[]},
	{"type":"function","name":"ChangeTssECDSAPubKey","inputs":[{"name":"pubKey","type":"string"},{"name":"oldPubKeySignature","type":"string"},{"name":"newPubKeySignature","type":"string"}]},
	{"type":"function","name":"ChangeAdministrator","inputs":[{"name":"administrator","type":"address"}]},
	{"type":"function","name":"ProposeAdministrator","inputs":[{"name":"address","type":"address"}]},
	{"type":"function","name":"SetAllowKeyGen","inputs":[{"name":"allowKeyGen","type":"bool"}]},
	{"type":"function","name":"SetRedeemDelay","inputs":[{"name":"redeemDelay","type":"uint64"}]},
	{"type":"function","name":"SetBridgeMetadata","inputs":[{"name":"metadata","type":"string"}]},
	{"type":"function","name":"UnwrapToken","inputs":[{"name":"networkClass","type":"uint32"},{"name":"chainId","type":"uint32"},{"name":"transactionHash","type":"hash"},{"name":"logIndex","type":"uint32"},{"name":"toAddress","type":"address"},{"name":"tokenAddress","type":"string"},{"name":"amount","type":"uint256"},{"name":"signature","type":"string"}]},
	{"type":"function","name":"RevokeUnwrapRequest","inputs":[{"name":"transactionHash","type":"hash"},{"name":"logIndex","type":"uint32"}]},
	{"type":"function","name":"Redeem","inputs":[{"name":"transactionHash","type":"hash"},{"name":"logIndex","type":"uint32"}]},
	{"type":"function","name":"NominateGuardians","inputs":[{"name":"guardians","type":"address[]"}]},
	{"type":"function","name":"SetOrchestratorInfo","inputs":[{"name":"windowSize","type":"uint64"},{"name":"keyGenThreshold","type":"uint32"},{"name":"confirmationsToFinality","type":"uint32"},{"name":"estimatedMomentumTime","type":"uint32"}]}
]`

// LiquidityDefinition contains the ABI for the Liquidity embedded contract
const LiquidityDefinition = `[
	{"type":"function","name":"Update","inputs":[]},
	{"type":"function","name":"Donate","inputs":[]},
	{"type":"function","name":"Fund","inputs":[{"name":"znnReward","type":"uint256"},{"name":"qsrReward","type":"uint256"}]},
	{"type":"function","name":"BurnZnn","inputs":[{"name":"burnAmount","type":"uint256"}]},
	{"type":"function","name":"SetTokenTuple","inputs":[{"name":"tokenStandards","type":"string[]"},{"name":"znnPercentages","type":"uint32[]"},{"name":"qsrPercentages","type":"uint32[]"},{"name":"minAmounts","type":"uint256[]"}]},
	{"type":"function","name":"NominateGuardians","inputs":[{"name":"guardians","type":"address[]"}]},
	{"type":"function","name":"ProposeAdministrator","inputs":[{"name":"address","type":"address"}]},
	{"type":"function","name":"Emergency","inputs":[]},
	{"type":"function","name":"SetIsHalted","inputs":[{"name":"isHalted","type":"bool"}]},
	{"type":"function","name":"LiquidityStake","inputs":[{"name":"durationInSec", "type":"int64"}]},
	{"type":"function","name":"CancelLiquidityStake","inputs":[{"name":"id","type":"hash"}]},
	{"type":"function","name":"UnlockLiquidityStakeEntries","inputs":[]},
	{"type":"function","name":"SetAdditionalReward","inputs":[{"name":"znnReward", "type":"uint256"},{"name":"qsrReward", "type":"uint256"}]},
	{"type":"function","name":"CollectReward","inputs":[]},
	{"type":"function","name":"ChangeAdministrator","inputs":[{"name":"administrator","type":"address"}]}
]`

// CommonDefinition contains common ABI methods used across multiple contracts
const CommonDefinition = `[
	{"type":"function","name":"DepositQsr","inputs":[]},
	{"type":"function","name":"WithdrawQsr","inputs":[]},
	{"type":"function","name":"CollectReward","inputs":[]},
	{"type":"function","name":"Update","inputs":[]},
	{"type":"function","name":"Donate","inputs":[]},
	{"type":"function","name":"VoteByName","inputs":[{"name":"id","type":"hash"},{"name":"name","type":"string"},{"name":"vote","type":"uint8"}]},
	{"type":"function","name":"VoteByProdAddress","inputs":[{"name":"id","type":"hash"},{"name":"vote","type":"uint8"}]}
]`

// =============================================================================
// Parsed ABI Objects
// =============================================================================

var (
	// Plasma is the parsed ABI for the Plasma embedded contract
	Plasma *abi.Abi

	// Pillar is the parsed ABI for the Pillar embedded contract
	Pillar *abi.Abi

	// Token is the parsed ABI for the Token embedded contract
	Token *abi.Abi

	// Sentinel is the parsed ABI for the Sentinel embedded contract
	Sentinel *abi.Abi

	// Swap is the parsed ABI for the Swap embedded contract
	Swap *abi.Abi

	// Stake is the parsed ABI for the Stake embedded contract
	Stake *abi.Abi

	// Accelerator is the parsed ABI for the Accelerator embedded contract
	Accelerator *abi.Abi

	// Spork is the parsed ABI for the Spork embedded contract
	Spork *abi.Abi

	// Htlc is the parsed ABI for the HTLC embedded contract
	Htlc *abi.Abi

	// Bridge is the parsed ABI for the Bridge embedded contract
	Bridge *abi.Abi

	// Liquidity is the parsed ABI for the Liquidity embedded contract
	Liquidity *abi.Abi

	// Common is the parsed ABI for common embedded contract methods
	Common *abi.Abi
)

// init parses all ABI definitions at package initialization
func init() {
	var err error

	Plasma, err = abi.FromJson(PlasmaDefinition)
	if err != nil {
		panic("failed to parse PlasmaDefinition: " + err.Error())
	}

	Pillar, err = abi.FromJson(PillarDefinition)
	if err != nil {
		panic("failed to parse PillarDefinition: " + err.Error())
	}

	Token, err = abi.FromJson(TokenDefinition)
	if err != nil {
		panic("failed to parse TokenDefinition: " + err.Error())
	}

	Sentinel, err = abi.FromJson(SentinelDefinition)
	if err != nil {
		panic("failed to parse SentinelDefinition: " + err.Error())
	}

	Swap, err = abi.FromJson(SwapDefinition)
	if err != nil {
		panic("failed to parse SwapDefinition: " + err.Error())
	}

	Stake, err = abi.FromJson(StakeDefinition)
	if err != nil {
		panic("failed to parse StakeDefinition: " + err.Error())
	}

	Accelerator, err = abi.FromJson(AcceleratorDefinition)
	if err != nil {
		panic("failed to parse AcceleratorDefinition: " + err.Error())
	}

	Spork, err = abi.FromJson(SporkDefinition)
	if err != nil {
		panic("failed to parse SporkDefinition: " + err.Error())
	}

	Htlc, err = abi.FromJson(HtlcDefinition)
	if err != nil {
		panic("failed to parse HtlcDefinition: " + err.Error())
	}

	Bridge, err = abi.FromJson(BridgeDefinition)
	if err != nil {
		panic("failed to parse BridgeDefinition: " + err.Error())
	}

	Liquidity, err = abi.FromJson(LiquidityDefinition)
	if err != nil {
		panic("failed to parse LiquidityDefinition: " + err.Error())
	}

	Common, err = abi.FromJson(CommonDefinition)
	if err != nil {
		panic("failed to parse CommonDefinition: " + err.Error())
	}
}
