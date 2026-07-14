package embedded

import (
	"reflect"
	"testing"

	"github.com/0x3639/znn-sdk-go/abi"
)

// TestConcreteCatalogsMatchStableSpec pins all 84 embedded functions from the
// stable Zenon SDK specification. Keeping the complete inventory here prevents
// shared Common entries from masking omissions in a concrete contract catalog.
func TestConcreteCatalogsMatchStableSpec(t *testing.T) {
	t.Parallel()
	catalogs := []struct {
		name string
		abi  *abi.Abi
		want []string
	}{
		{"Accelerator", Accelerator, []string{
			"Update()", "CreateProject(string,string,string,uint256,uint256)",
			"AddPhase(hash,string,string,string,uint256,uint256)",
			"UpdatePhase(hash,string,string,string,uint256,uint256)", "Donate()",
			"VoteByName(hash,string,uint8)", "VoteByProdAddress(hash,uint8)",
		}},
		{"Bridge", Bridge, []string{
			"WrapToken(uint32,uint32,string)", "UpdateWrapRequest(hash,string)",
			"SetNetwork(uint32,uint32,string,string,string)", "RemoveNetwork(uint32,uint32)",
			"SetTokenPair(uint32,uint32,tokenStandard,string,bool,bool,bool,uint256,uint32,uint32,string)",
			"SetNetworkMetadata(uint32,uint32,string)",
			"RemoveTokenPair(uint32,uint32,tokenStandard,string)", "Halt(string)", "Unhalt()",
			"Emergency()", "ChangeTssECDSAPubKey(string,string,string)",
			"ChangeAdministrator(address)", "ProposeAdministrator(address)", "SetAllowKeyGen(bool)",
			"SetRedeemDelay(uint64)", "SetBridgeMetadata(string)",
			"UnwrapToken(uint32,uint32,hash,uint32,address,string,uint256,string)",
			"RevokeUnwrapRequest(hash,uint32)", "Redeem(hash,uint32)", "NominateGuardians(address[])",
			"SetOrchestratorInfo(uint64,uint32,uint32,uint32)",
		}},
		{"Common", Common, []string{
			"DepositQsr()", "WithdrawQsr()", "CollectReward()", "Update()", "Donate()",
			"VoteByName(hash,string,uint8)", "VoteByProdAddress(hash,uint8)",
		}},
		{"Htlc", Htlc, []string{
			"Create(address,int64,uint8,uint8,bytes)", "Reclaim(hash)", "Unlock(hash,bytes)",
			"DenyProxyUnlock()", "AllowProxyUnlock()",
		}},
		{"Liquidity", Liquidity, []string{
			"Update()", "Donate()", "Fund(uint256,uint256)", "BurnZnn(uint256)",
			"SetTokenTuple(string[],uint32[],uint32[],uint256[])", "NominateGuardians(address[])",
			"ProposeAdministrator(address)", "Emergency()", "SetIsHalted(bool)", "LiquidityStake(int64)",
			"CancelLiquidityStake(hash)", "UnlockLiquidityStakeEntries()",
			"SetAdditionalReward(uint256,uint256)", "CollectReward()", "ChangeAdministrator(address)",
		}},
		{"Pillar", Pillar, []string{
			"Update()", "Register(string,address,address,uint8,uint8)",
			"RegisterLegacy(string,address,address,uint8,uint8,string,string)",
			"UpdatePillar(string,address,address,uint8,uint8)", "DepositQsr()", "WithdrawQsr()",
			"Revoke(string)", "Delegate(string)", "Undelegate()", "CollectReward()",
		}},
		{"Plasma", Plasma, []string{"Fuse(address)", "CancelFuse(hash)"}},
		{"Sentinel", Sentinel, []string{
			"DepositQsr()", "WithdrawQsr()", "Register()", "Revoke()", "Update()", "CollectReward()",
		}},
		{"Spork", Spork, []string{"CreateSpork(string,string)", "ActivateSpork(hash)"}},
		{"Stake", Stake, []string{"Stake(int64)", "Cancel(hash)", "CollectReward()", "Update()"}},
		{"Swap", Swap, []string{"RetrieveAssets(string,string)"}},
		{"Token", Token, []string{
			"IssueToken(string,string,string,uint256,uint256,uint8,bool,bool,bool)",
			"Mint(tokenStandard,uint256,address)", "Burn()", "UpdateToken(tokenStandard,address,bool,bool)",
		}},
	}

	total := 0
	for _, catalog := range catalogs {
		t.Run(catalog.name, func(t *testing.T) {
			if catalog.abi == nil {
				t.Fatal("catalog is nil")
			}
			got := make([]string, len(catalog.abi.Entries))
			for index := range catalog.abi.Entries {
				got[index] = catalog.abi.Entries[index].FormatSignature()
			}
			if !reflect.DeepEqual(got, catalog.want) {
				t.Fatalf("catalog signatures mismatch\n got: %v\nwant: %v", got, catalog.want)
			}
		})
		total += len(catalog.want)
	}
	if total != 84 {
		t.Fatalf("stable ABI inventory contains %d functions, want 84", total)
	}
}
