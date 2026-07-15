package embedded

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
)

type embeddedRecordingCaller struct {
	method string
	args   []interface{}
	err    error
}

func (c *embeddedRecordingCaller) Call(_ interface{}, method string, args ...interface{}) error {
	c.method = method
	c.args = append([]interface{}(nil), args...)
	return c.err
}

func TestEmbeddedReadMethodsUseCanonicalWireNames(t *testing.T) {
	caller := new(embeddedRecordingCaller)
	address := types.ParseAddressPanic("z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f")
	hash := types.HexToHashPanic("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")

	accelerator := NewAcceleratorApi(caller)
	bridge := NewBridgeApi(caller)
	htlc := NewHtlcApi(caller)
	liquidity := NewLiquidityApi(caller)
	pillar := NewPillarApi(caller)
	plasma := NewPlasmaApi(caller)
	sentinel := NewSentinelApi(caller)
	spork := NewSporkApi(caller)
	stake := NewStakeApi(caller)
	swap := NewSwapApi(caller)
	token := NewTokenApi(caller)

	tests := []struct {
		method string
		call   func() error
	}{
		{"embedded.accelerator.getAll", func() error { _, err := accelerator.GetAll(1, 2); return err }},
		{"embedded.accelerator.getProjectById", func() error { _, err := accelerator.GetProjectById(hash); return err }},
		{"embedded.accelerator.getPhaseById", func() error { _, err := accelerator.GetPhaseById(hash); return err }},
		{"embedded.accelerator.getVoteBreakdown", func() error { _, err := accelerator.GetVoteBreakdown(hash); return err }},
		{"embedded.accelerator.getPillarVotes", func() error { _, err := accelerator.GetPillarVotes("pillar", []types.Hash{hash}); return err }},

		{"embedded.bridge.getBridgeInfo", func() error { _, err := bridge.GetBridgeInfo(); return err }},
		{"embedded.bridge.getOrchestratorInfo", func() error { _, err := bridge.GetOrchestratorInfo(); return err }},
		{"embedded.bridge.getTimeChallengesInfo", func() error { _, err := bridge.GetTimeChallengesInfo(); return err }},
		{"embedded.bridge.getSecurityInfo", func() error { _, err := bridge.GetSecurityInfo(); return err }},
		{"embedded.bridge.getNetworkInfo", func() error { _, err := bridge.GetNetworkInfo(1, 2); return err }},
		{"embedded.bridge.getWrapTokenRequestById", func() error { _, err := bridge.GetWrapTokenRequestById(hash); return err }},
		{"embedded.bridge.getAllWrapTokenRequests", func() error { _, err := bridge.GetAllWrapTokenRequests(1, 2); return err }},
		{"embedded.bridge.getAllWrapTokenRequestsByToAddress", func() error { _, err := bridge.GetAllWrapTokenRequestsByToAddress("0x1", 1, 2); return err }},
		{"embedded.bridge.getAllWrapTokenRequestsByToAddressNetworkClassAndChainId", func() error {
			_, err := bridge.GetAllWrapTokenRequestsByToAddressNetworkClassAndChainId("0x1", 1, 2, 3, 4)
			return err
		}},
		{"embedded.bridge.getAllNetworks", func() error { _, err := bridge.GetAllNetworks(1, 2); return err }},
		{"embedded.bridge.getAllUnsignedWrapTokenRequests", func() error { _, err := bridge.GetAllUnsignedWrapTokenRequests(1, 2); return err }},
		{"embedded.bridge.getUnwrapTokenRequestByHashAndLog", func() error { _, err := bridge.GetUnwrapTokenRequestByHashAndLog(hash, 3); return err }},
		{"embedded.bridge.getAllUnwrapTokenRequests", func() error { _, err := bridge.GetAllUnwrapTokenRequests(1, 2); return err }},
		{"embedded.bridge.getAllUnwrapTokenRequestsByToAddress", func() error { _, err := bridge.GetAllUnwrapTokenRequestsByToAddress("0x1", 1, 2); return err }},
		{"embedded.bridge.getFeeTokenPair", func() error { _, err := bridge.GetFeeTokenPair(types.ZnnTokenStandard); return err }},

		{"embedded.htlc.getById", func() error { _, err := htlc.GetById(hash); return err }},
		{"embedded.htlc.getProxyUnlockStatus", func() error { _, err := htlc.GetProxyUnlockStatus(address); return err }},

		{"embedded.liquidity.getUncollectedReward", func() error { _, err := liquidity.GetUncollectedReward(address); return err }},
		{"embedded.liquidity.getFrontierRewardByPage", func() error { _, err := liquidity.GetFrontierRewardByPage(address, 1, 2); return err }},
		{"embedded.liquidity.getLiquidityInfo", func() error { _, err := liquidity.GetLiquidityInfo(); return err }},
		{"embedded.liquidity.getSecurityInfo", func() error { _, err := liquidity.GetSecurityInfo(); return err }},
		{"embedded.liquidity.getLiquidityStakeEntriesByAddress", func() error { _, err := liquidity.GetLiquidityStakeEntriesByAddress(address, 1, 2); return err }},
		{"embedded.liquidity.getTimeChallengesInfo", func() error { _, err := liquidity.GetTimeChallengesInfo(); return err }},

		{"embedded.pillar.getDepositedQsr", func() error { _, err := pillar.GetDepositedQsr(address); return err }},
		{"embedded.pillar.getQsrRegistrationCost", func() error { _, err := pillar.GetQsrRegistrationCost(); return err }},
		{"embedded.pillar.getUncollectedReward", func() error { _, err := pillar.GetUncollectedReward(address); return err }},
		{"embedded.pillar.getFrontierRewardByPage", func() error { _, err := pillar.GetFrontierRewardByPage(address, 1, 2); return err }},
		{"embedded.pillar.getAll", func() error { _, err := pillar.GetAll(1, 2); return err }},
		{"embedded.pillar.getByOwner", func() error { _, err := pillar.GetByOwner(address); return err }},
		{"embedded.pillar.getByName", func() error { _, err := pillar.GetByName("pillar"); return err }},
		{"embedded.pillar.checkNameAvailability", func() error { _, err := pillar.CheckNameAvailability("pillar"); return err }},
		{"embedded.pillar.getDelegatedPillar", func() error { _, err := pillar.GetDelegatedPillar(address); return err }},
		{"embedded.pillar.getPillarEpochHistory", func() error { _, err := pillar.GetPillarEpochHistory("pillar", 1, 2); return err }},
		{"embedded.pillar.getPillarsHistoryByEpoch", func() error { _, err := pillar.GetPillarsHistoryByEpoch(1, 2, 3); return err }},

		{"embedded.plasma.get", func() error { _, err := plasma.Get(address); return err }},
		{"embedded.plasma.getEntriesByAddress", func() error { _, err := plasma.GetEntriesByAddress(address, 1, 2); return err }},
		{"embedded.plasma.getRequiredPoWForAccountBlock", func() error {
			_, err := plasma.GetRequiredPoWForAccountBlock(GetRequiredParam{Address: address})
			return err
		}},

		{"embedded.sentinel.getByOwner", func() error { _, err := sentinel.GetByOwner(address); return err }},
		{"embedded.sentinel.getAllActive", func() error { _, err := sentinel.GetAllActive(1, 2); return err }},
		{"embedded.sentinel.getDepositedQsr", func() error { _, err := sentinel.GetDepositedQsr(address); return err }},
		{"embedded.sentinel.getUncollectedReward", func() error { _, err := sentinel.GetUncollectedReward(address); return err }},
		{"embedded.sentinel.getFrontierRewardByPage", func() error { _, err := sentinel.GetFrontierRewardByPage(address, 1, 2); return err }},

		{"embedded.spork.getAll", func() error { _, err := spork.GetAll(1, 2); return err }},
		{"embedded.stake.getUncollectedReward", func() error { _, err := stake.GetUncollectedReward(address); return err }},
		{"embedded.stake.getFrontierRewardByPage", func() error { _, err := stake.GetFrontierRewardByPage(address, 1, 2); return err }},
		{"embedded.stake.getEntriesByAddress", func() error { _, err := stake.GetEntriesByAddress(address, 1, 2); return err }},

		{"embedded.swap.getAssetsByKeyIdHash", func() error { _, err := swap.GetAssetsByKeyIdHash(hash); return err }},
		{"embedded.swap.getAssets", func() error { _, err := swap.GetAssets(); return err }},
		{"embedded.swap.getLegacyPillars", func() error { _, err := swap.GetLegacyPillars(); return err }},

		{"embedded.token.getAll", func() error { _, err := token.GetAll(1, 2); return err }},
		{"embedded.token.getByOwner", func() error { _, err := token.GetByOwner(address, 1, 2); return err }},
		{"embedded.token.getByZts", func() error { _, err := token.GetByZts(types.ZnnTokenStandard); return err }},
	}

	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			caller.method = ""
			if err := test.call(); err != nil {
				t.Fatalf("call error = %v", err)
			}
			if caller.method != test.method {
				t.Fatalf("wire method = %q, want %q", caller.method, test.method)
			}
			wantErr := errors.New("injected RPC failure")
			caller.err = wantErr
			if err := test.call(); !errors.Is(err, wantErr) {
				t.Fatalf("injected error = %v, want %v", err, wantErr)
			}
			caller.err = nil
		})
	}
}

func TestEmbeddedReadMethodsPropagateErrors(t *testing.T) {
	want := errors.New("rpc unavailable")
	api := NewTokenApi(&embeddedRecordingCaller{err: want})
	if _, err := api.GetByZts(types.ZnnTokenStandard); !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

func TestEmbeddedTransactionBuildersProduceSendBlocks(t *testing.T) {
	address := types.ParseAddressPanic("z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f")
	hash := types.HexToHashPanic("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	amount := big.NewInt(100)

	accelerator := NewAcceleratorApi(nil)
	bridge := NewBridgeApi(nil)
	htlc := NewHtlcApi(nil)
	liquidity := NewLiquidityApi(nil)
	pillar := NewPillarApi(nil)
	plasma := NewPlasmaApi(nil)
	sentinel := NewSentinelApi(nil)
	spork := NewSporkApi(nil)
	stake := NewStakeApi(nil)
	swap := NewSwapApi(nil)
	token := NewTokenApi(nil)

	tests := []struct {
		name     string
		contract types.Address
		build    func() *nom.AccountBlock
	}{
		{"accelerator/create", types.AcceleratorContract, func() *nom.AccountBlock {
			return accelerator.CreateProject("name", "description", "https://example.com", amount, amount)
		}},
		{"accelerator/add-phase", types.AcceleratorContract, func() *nom.AccountBlock {
			return accelerator.AddPhase(hash, "name", "description", "https://example.com", amount, amount)
		}},
		{"accelerator/update-phase", types.AcceleratorContract, func() *nom.AccountBlock {
			return accelerator.UpdatePhase(hash, "name", "description", "https://example.com", amount, amount)
		}},
		{"accelerator/donate", types.AcceleratorContract, func() *nom.AccountBlock { return accelerator.Donate(amount, types.QsrTokenStandard) }},
		{"accelerator/vote-name", types.AcceleratorContract, func() *nom.AccountBlock { return accelerator.VoteByName(hash, "pillar", VoteYes) }},
		{"accelerator/vote-producer", types.AcceleratorContract, func() *nom.AccountBlock { return accelerator.VoteByProducerAddress(hash, VoteNo) }},

		{"bridge/wrap", types.BridgeContract, func() *nom.AccountBlock { return bridge.WrapToken(1, 2, "0x1", amount, types.ZnnTokenStandard) }},
		{"bridge/update-wrap", types.BridgeContract, func() *nom.AccountBlock { return bridge.UpdateWrapRequest(hash, "signature") }},
		{"bridge/unwrap", types.BridgeContract, func() *nom.AccountBlock {
			return bridge.UnwrapToken(1, 2, "0x2", hash, 3, amount, address, "signature")
		}},
		{"bridge/redeem", types.BridgeContract, func() *nom.AccountBlock { return bridge.Redeem(hash, 3) }},
		{"bridge/halt", types.BridgeContract, func() *nom.AccountBlock { return bridge.Halt("signature") }},
		{"bridge/emergency", types.BridgeContract, bridge.Emergency},
		{"bridge/unhalt", types.BridgeContract, bridge.Unhalt},
		{"bridge/keygen", types.BridgeContract, func() *nom.AccountBlock { return bridge.SetAllowKeygen(true) }},
		{"bridge/change-tss", types.BridgeContract, func() *nom.AccountBlock { return bridge.ChangeTssECDSAPubKey("key", "sig", "new-sig") }},
		{"bridge/change-admin", types.BridgeContract, func() *nom.AccountBlock { return bridge.ChangeAdministrator(address) }},
		{"bridge/add-network", types.BridgeContract, func() *nom.AccountBlock { return bridge.AddNetwork(1, 2, "name", "0x1", "metadata") }},
		{"bridge/remove-network", types.BridgeContract, func() *nom.AccountBlock { return bridge.RemoveNetwork(1, 2) }},
		{"bridge/set-token-pair", types.BridgeContract, func() *nom.AccountBlock {
			return bridge.SetTokenPair(1, 2, types.ZnnTokenStandard, "0x1", true, true, false, amount, 1, 2, "metadata")
		}},
		{"bridge/remove-token-pair", types.BridgeContract, func() *nom.AccountBlock { return bridge.RemoveTokenPair(1, 2, types.ZnnTokenStandard, "0x1") }},
		{"bridge/network-metadata", types.BridgeContract, func() *nom.AccountBlock { return bridge.SetNetworkMetadata(1, 2, "metadata") }},
		{"bridge/orchestrator", types.BridgeContract, func() *nom.AccountBlock { return bridge.SetOrchestratorInfo(1, 2, 3, 4) }},
		{"bridge/guardians", types.BridgeContract, func() *nom.AccountBlock { return bridge.NominateGuardians([]types.Address{address}) }},
		{"bridge/metadata", types.BridgeContract, func() *nom.AccountBlock { return bridge.SetBridgeMetadata("metadata") }},
		{"bridge/propose-admin", types.BridgeContract, func() *nom.AccountBlock { return bridge.ProposeAdministrator(address) }},
		{"bridge/revoke-unwrap", types.BridgeContract, func() *nom.AccountBlock { return bridge.RevokeUnwrapRequest(hash, 3) }},

		{"htlc/create", types.HtlcContract, func() *nom.AccountBlock {
			return htlc.Create(types.ZnnTokenStandard, amount, address, 1000, 0, 32, make([]byte, 32))
		}},
		{"htlc/reclaim", types.HtlcContract, func() *nom.AccountBlock { return htlc.Reclaim(hash) }},
		{"htlc/unlock", types.HtlcContract, func() *nom.AccountBlock { return htlc.Unlock(hash, []byte("key")) }},
		{"htlc/deny", types.HtlcContract, htlc.DenyProxyUnlock},
		{"htlc/allow", types.HtlcContract, htlc.AllowProxyUnlock},

		{"liquidity/set-tuple", types.LiquidityContract, func() *nom.AccountBlock {
			return liquidity.SetTokenTupleMethod([]string{types.ZnnTokenStandard.String()}, []uint32{1}, []uint32{2}, []*big.Int{amount})
		}},
		{"liquidity/stake", types.LiquidityContract, func() *nom.AccountBlock { return liquidity.LiquidityStake(2592000, amount, types.ZnnTokenStandard) }},
		{"liquidity/halt", types.LiquidityContract, func() *nom.AccountBlock { return liquidity.SetIsHalted(true) }},
		{"liquidity/collect", types.LiquidityContract, liquidity.CollectReward},
		{"liquidity/cancel", types.LiquidityContract, func() *nom.AccountBlock { return liquidity.CancelLiquidity(hash) }},
		{"liquidity/unlock", types.LiquidityContract, func() *nom.AccountBlock { return liquidity.UnlockLiquidityStakeEntries(types.ZnnTokenStandard) }},
		{"liquidity/reward", types.LiquidityContract, func() *nom.AccountBlock { return liquidity.SetAdditionalReward(amount, amount) }},
		{"liquidity/guardians", types.LiquidityContract, func() *nom.AccountBlock { return liquidity.NominateGuardians([]types.Address{address}) }},
		{"liquidity/propose-admin", types.LiquidityContract, func() *nom.AccountBlock { return liquidity.ProposeAdministrator(address) }},
		{"liquidity/change-admin", types.LiquidityContract, func() *nom.AccountBlock { return liquidity.ChangeAdministrator(address) }},
		{"liquidity/emergency", types.LiquidityContract, liquidity.Emergency},

		{"pillar/register", types.PillarContract, func() *nom.AccountBlock { return pillar.Register("pillar", address, address, 50, 50) }},
		{"pillar/update", types.PillarContract, func() *nom.AccountBlock { return pillar.UpdatePillar("pillar", address, address, 50, 50) }},
		{"pillar/revoke", types.PillarContract, func() *nom.AccountBlock { return pillar.Revoke("pillar") }},
		{"pillar/legacy", types.PillarContract, func() *nom.AccountBlock {
			return pillar.RegisterLegacy("pillar", address, address, 50, 50, "key", "signature")
		}},
		{"pillar/delegate", types.PillarContract, func() *nom.AccountBlock { return pillar.Delegate("pillar") }},
		{"pillar/undelegate", types.PillarContract, pillar.Undelegate},
		{"pillar/deposit", types.PillarContract, func() *nom.AccountBlock { return pillar.DepositQsr(amount) }},
		{"pillar/withdraw", types.PillarContract, pillar.WithdrawQsr},
		{"pillar/collect", types.PillarContract, pillar.CollectReward},

		{"plasma/fuse", types.PlasmaContract, func() *nom.AccountBlock { return plasma.Fuse(address, amount) }},
		{"plasma/cancel", types.PlasmaContract, func() *nom.AccountBlock { return plasma.Cancel(hash) }},
		{"sentinel/register", types.SentinelContract, sentinel.Register},
		{"sentinel/revoke", types.SentinelContract, sentinel.Revoke},
		{"sentinel/deposit", types.SentinelContract, func() *nom.AccountBlock { return sentinel.DepositQsr(amount) }},
		{"sentinel/withdraw", types.SentinelContract, sentinel.WithdrawQsr},
		{"sentinel/collect", types.SentinelContract, sentinel.CollectReward},
		{"spork/create", types.SporkContract, func() *nom.AccountBlock { return spork.CreateSpork("name", "description") }},
		{"spork/activate", types.SporkContract, func() *nom.AccountBlock { return spork.ActivateSpork(hash) }},
		{"stake/stake", types.StakeContract, func() *nom.AccountBlock { return stake.Stake(2592000, amount) }},
		{"stake/cancel", types.StakeContract, func() *nom.AccountBlock { return stake.Cancel(hash) }},
		{"stake/collect", types.StakeContract, stake.CollectReward},
		{"swap/retrieve", types.SwapContract, func() *nom.AccountBlock { return swap.RetrieveAssets("key", "signature") }},
		{"token/issue", types.TokenContract, func() *nom.AccountBlock {
			return token.IssueToken("Token", "TOK", "example.com", amount, amount, 8, true, true, false)
		}},
		{"token/mint", types.TokenContract, func() *nom.AccountBlock { return token.Mint(types.ZnnTokenStandard, amount, address) }},
		{"token/burn", types.TokenContract, func() *nom.AccountBlock { return token.Burn(types.ZnnTokenStandard, amount) }},
		{"token/update", types.TokenContract, func() *nom.AccountBlock { return token.UpdateToken(types.ZnnTokenStandard, address, false, false) }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block := test.build()
			if block == nil || block.BlockType != nom.BlockTypeUserSend {
				t.Fatalf("block = %+v, want user-send template", block)
			}
			if block.ToAddress != test.contract {
				t.Fatalf("ToAddress = %s, want %s", block.ToAddress, test.contract)
			}
			if block.Amount == nil || len(block.Data) == 0 {
				t.Fatalf("template lacks amount or ABI data: %+v", block)
			}
		})
	}
}

func TestEmbeddedCustomJSONTypesDecodeStringAmounts(t *testing.T) {
	typesUnderTest := []json.Unmarshaler{
		new(UncollectedReward), new(RewardHistoryEntry), new(PillarInfo),
		new(PillarEpochHistory), new(DelegationInfo), new(StakeEntry),
		new(StakeList), new(PlasmaInfo), new(FusionEntry), new(FusionEntryList),
		new(Token), new(TokenTuple), new(LiquidityInfo), new(LiquidityStakeEntry),
		new(LiquidityStakeList), new(TokenPair), new(WrapTokenRequest),
		new(UnwrapTokenRequest), new(ZtsFeesInfo), new(PhaseInfo), new(Project),
		new(SwapAssetEntry), new(SwapAssetEntrySimple), new(HtlcInfo),
	}

	for _, value := range typesUnderTest {
		t.Run(fmt.Sprintf("%T", value), func(t *testing.T) {
			if err := value.UnmarshalJSON([]byte(`{}`)); err != nil {
				t.Fatalf("empty object: %v", err)
			}
			if err := value.UnmarshalJSON([]byte(`{`)); err == nil {
				t.Fatal("malformed JSON was accepted")
			}
		})
	}
}

func TestEmbeddedTypeHelpersAndHTLCBase64Validation(t *testing.T) {
	if (&DelegationInfo{Status: 1}).IsPillarActive() != true || (&DelegationInfo{Status: 0}).IsPillarActive() != false {
		t.Fatal("IsPillarActive returned an unexpected value")
	}
	if (&SwapAssetEntry{Qsr: big.NewInt(1), Znn: big.NewInt(0)}).HasBalance() != true ||
		(&SwapAssetEntry{Qsr: big.NewInt(0), Znn: big.NewInt(0)}).HasBalance() != false {
		t.Fatal("HasBalance returned an unexpected value")
	}
	if err := json.Unmarshal([]byte(`{"hashLock":"%%%"}`), new(HtlcInfo)); err == nil {
		t.Fatal("invalid base64 hash lock was accepted")
	}
	if got := NewPlasmaApi(nil).GetPlasmaByQsr(nil); got.Sign() != 0 {
		t.Fatalf("nil QSR plasma = %s, want 0", got)
	}
	if got := NewPlasmaApi(nil).GetPlasmaByQsr(big.NewInt(2)); got.Cmp(big.NewInt(4200)) != 0 {
		t.Fatalf("plasma = %s, want 4200", got)
	}
}
