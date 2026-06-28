package embedded

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

func TestGovernanceApi_ProposeAction(t *testing.T) {
	api := NewGovernanceApi(nil)
	dest := types.SporkContract
	const name = "Activate governance"
	const desc = "Enable the governance ratchet"
	const url = "https://forum.zenon.org/x"
	const data = "aGVsbG8=" // base64("hello")

	block := api.ProposeAction(name, desc, url, dest, data)
	if block == nil {
		t.Fatal("ProposeAction returned nil")
	}
	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}
	if block.ToAddress != types.GovernanceContract {
		t.Errorf("ToAddress = %s, want GovernanceContract", block.ToAddress.String())
	}
	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard.String())
	}
	// ProposeAction must cost exactly 1 ZNN (8 decimals -> 100000000 base units).
	if block.Amount == nil || block.Amount.String() != "100000000" {
		t.Errorf("Amount = %v, want 100000000 (1 ZNN)", block.Amount)
	}
	expected := definition.ABIGovernance.PackMethodPanic(definition.ProposeActionMethodName, name, desc, url, dest, data)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch\n  got:  %x\n  want: %x", block.Data, expected)
	}
}

func TestGovernanceApi_ExecuteAction(t *testing.T) {
	api := NewGovernanceApi(nil)
	id := types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	block := api.ExecuteAction(id)
	if block == nil {
		t.Fatal("ExecuteAction returned nil")
	}
	if block.ToAddress != types.GovernanceContract {
		t.Errorf("ToAddress = %s, want GovernanceContract", block.ToAddress.String())
	}
	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard.String())
	}
	if block.Amount == nil || block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}
	expected := definition.ABIGovernance.PackMethodPanic(definition.ExecuteActionMethodName, id)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch\n  got:  %x\n  want: %x", block.Data, expected)
	}
}

func TestGovernanceApi_VoteByName(t *testing.T) {
	api := NewGovernanceApi(nil)
	id := types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	const pillar = "MyPillar"

	block := api.VoteByName(id, pillar, VoteYes)
	if block.ToAddress != types.GovernanceContract {
		t.Errorf("ToAddress = %s, want GovernanceContract", block.ToAddress.String())
	}
	if block.Amount == nil || block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}
	expected := definition.ABIGovernance.PackMethodPanic(definition.VoteByNameMethodName, id, pillar, VoteYes)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch\n  got:  %x\n  want: %x", block.Data, expected)
	}
}

func TestGovernanceApi_VoteByProducerAddress(t *testing.T) {
	api := NewGovernanceApi(nil)
	id := types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	block := api.VoteByProducerAddress(id, VoteNo)
	if block.ToAddress != types.GovernanceContract {
		t.Errorf("ToAddress = %s, want GovernanceContract", block.ToAddress.String())
	}
	expected := definition.ABIGovernance.PackMethodPanic(definition.VoteByProdAddressMethodName, id, VoteNo)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch\n  got:  %x\n  want: %x", block.Data, expected)
	}
}

// TestGovernanceVoteConstants pins the vote enum to its on-chain values.
func TestGovernanceVoteConstants(t *testing.T) {
	if VoteYes != 0 {
		t.Errorf("VoteYes = %d, want 0", VoteYes)
	}
	if VoteNo != 1 {
		t.Errorf("VoteNo = %d, want 1", VoteNo)
	}
	if VoteAbstain != 2 {
		t.Errorf("VoteAbstain = %d, want 2", VoteAbstain)
	}
}

// TestGovernanceActionConstants pins the action type/status values that
// go-syrius relies on when rendering action state.
func TestGovernanceActionConstants(t *testing.T) {
	if Type1Action != 1 || Type2Action != 2 {
		t.Errorf("action types = (%d,%d), want (1,2)", Type1Action, Type2Action)
	}
	if ActionStatusVoting != 0 || ActionStatusApproved != 1 ||
		ActionStatusRejected != 2 || ActionStatusNoDecision != 3 {
		t.Errorf("action statuses = (%d,%d,%d,%d), want (0,1,2,3)",
			ActionStatusVoting, ActionStatusApproved, ActionStatusRejected, ActionStatusNoDecision)
	}
}

// TestAction_Unmarshal verifies the SDK Action type deserializes the
// PascalCase JSON shape produced by embedded.governance, including the nested
// lowercase VoteBreakdown and the base64 Data field.
func TestAction_Unmarshal(t *testing.T) {
	const hashHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	raw := []byte(`{
		"Id":"` + hashHex + `",
		"Owner":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"Name":"Activate spork",
		"Description":"desc",
		"Url":"https://x",
		"Destination":"z1qxemdeddedxsp0rkxxxxxxxxxxxxxxxx956u48",
		"Data":"aGVsbG8=",
		"CreationTimestamp":1700000000,
		"Type":1,
		"Round":0,
		"CurrentVoteId":"` + hashHex + `",
		"RoundStartTimestamp":1700000000,
		"Status":0,
		"Executed":false,
		"Expired":false,
		"ActivePillarThreshold":66,
		"DirectionalThreshold":50,
		"VotingPeriod":3600,
		"Votes":{"id":"` + hashHex + `","total":3,"yes":2,"no":1}
	}`)

	var a Action
	if err := json.Unmarshal(raw, &a); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if a.Id.String() != hashHex {
		t.Errorf("Id = %s, want %s", a.Id.String(), hashHex)
	}
	if a.Name != "Activate spork" {
		t.Errorf("Name = %q", a.Name)
	}
	if a.Type != Type1Action {
		t.Errorf("Type = %d, want %d", a.Type, Type1Action)
	}
	if a.ActivePillarThreshold != 66 || a.DirectionalThreshold != 50 || a.VotingPeriod != 3600 {
		t.Errorf("thresholds/period = (%d,%d,%d)", a.ActivePillarThreshold, a.DirectionalThreshold, a.VotingPeriod)
	}
	if a.Votes == nil || a.Votes.Total != 3 || a.Votes.Yes != 2 || a.Votes.No != 1 {
		t.Errorf("Votes = %+v, want total=3 yes=2 no=1", a.Votes)
	}
	decoded, err := a.DecodedData()
	if err != nil {
		t.Fatalf("DecodedData failed: %v", err)
	}
	if string(decoded) != "hello" {
		t.Errorf("DecodedData = %q, want %q", decoded, "hello")
	}
}
