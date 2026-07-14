package embedded

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/zenon-network/go-zenon/common/types"
)

// =============================================================================
// Common Types Tests
// =============================================================================

func TestUncollectedReward_UnmarshalJSON(t *testing.T) {
	addr := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	raw := `{"address":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz","znnAmount":"1000000000","qsrAmount":"500000000"}`

	var r UncollectedReward
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if r.Address != addr {
		t.Errorf("Address = %s, want %s", r.Address, addr)
	}
	if r.ZnnAmount == nil || r.ZnnAmount.String() != "1000000000" {
		t.Errorf("ZnnAmount = %v, want 1000000000", r.ZnnAmount)
	}
	if r.QsrAmount == nil || r.QsrAmount.String() != "500000000" {
		t.Errorf("QsrAmount = %v, want 500000000", r.QsrAmount)
	}
}

func TestUncollectedReward_UnmarshalJSON_ZeroAmounts(t *testing.T) {
	raw := `{"address":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz","znnAmount":"0","qsrAmount":"0"}`

	var r UncollectedReward
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if r.ZnnAmount == nil || r.ZnnAmount.Sign() != 0 {
		t.Errorf("ZnnAmount should be 0, got %v", r.ZnnAmount)
	}
	if r.QsrAmount == nil || r.QsrAmount.Sign() != 0 {
		t.Errorf("QsrAmount should be 0, got %v", r.QsrAmount)
	}
}

func TestUncollectedReward_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var r UncollectedReward
	if err := json.Unmarshal([]byte(`not json`), &r); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestRewardHistoryEntry_UnmarshalJSON(t *testing.T) {
	raw := `{"epoch":42,"znnAmount":"2000000000","qsrAmount":"800000000"}`

	var r RewardHistoryEntry
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if r.Epoch != 42 {
		t.Errorf("Epoch = %d, want 42", r.Epoch)
	}
	if r.ZnnAmount == nil || r.ZnnAmount.String() != "2000000000" {
		t.Errorf("ZnnAmount = %v, want 2000000000", r.ZnnAmount)
	}
	if r.QsrAmount == nil || r.QsrAmount.String() != "800000000" {
		t.Errorf("QsrAmount = %v, want 800000000", r.QsrAmount)
	}
}

func TestRewardHistoryEntry_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var r RewardHistoryEntry
	if err := json.Unmarshal([]byte(`{`), &r); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// =============================================================================
// Pillar Types Tests
// =============================================================================

func TestPillarInfo_UnmarshalJSON(t *testing.T) {
	ownerAddr := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	raw := `{
		"name":"TestPillar",
		"rank":1,
		"type":2,
		"ownerAddress":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"producerAddress":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"withdrawAddress":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"giveMomentumRewardPercentage":50,
		"giveDelegateRewardPercentage":50,
		"isRevocable":true,
		"revokeCooldown":100,
		"revokeTimestamp":1234567890,
		"currentStats":{"producedMomentums":10,"expectedMomentums":10},
		"weight":"5000000000"
	}`

	var p PillarInfo
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if p.Name != "TestPillar" {
		t.Errorf("Name = %s, want TestPillar", p.Name)
	}
	if p.Rank != 1 {
		t.Errorf("Rank = %d, want 1", p.Rank)
	}
	if p.Type != RegularPillarType {
		t.Errorf("Type = %d, want %d (RegularPillarType)", p.Type, RegularPillarType)
	}
	if p.OwnerAddress != ownerAddr {
		t.Errorf("OwnerAddress = %s, want %s", p.OwnerAddress, ownerAddr)
	}
	if !p.IsRevocable {
		t.Error("IsRevocable should be true")
	}
	if p.Weight == nil || p.Weight.String() != "5000000000" {
		t.Errorf("Weight = %v, want 5000000000", p.Weight)
	}
	if p.CurrentStats == nil {
		t.Fatal("CurrentStats should not be nil")
	}
	if p.CurrentStats.ProducedMomentums != 10 {
		t.Errorf("CurrentStats.ProducedMomentums = %d, want 10", p.CurrentStats.ProducedMomentums)
	}
}

func TestPillarInfo_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var p PillarInfo
	if err := json.Unmarshal([]byte(`{invalid}`), &p); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestPillarEpochHistory_UnmarshalJSON(t *testing.T) {
	raw := `{
		"name":"TestPillar",
		"epoch":5,
		"giveBlockRewardPercentage":50,
		"giveDelegateRewardPercentage":50,
		"producedBlockNum":100,
		"expectedBlockNum":100,
		"weight":"3000000000"
	}`

	var h PillarEpochHistory
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if h.Name != "TestPillar" {
		t.Errorf("Name = %s, want TestPillar", h.Name)
	}
	if h.Epoch != 5 {
		t.Errorf("Epoch = %d, want 5", h.Epoch)
	}
	if h.ProducedBlockNum != 100 {
		t.Errorf("ProducedBlockNum = %d, want 100", h.ProducedBlockNum)
	}
	if h.Weight == nil || h.Weight.String() != "3000000000" {
		t.Errorf("Weight = %v, want 3000000000", h.Weight)
	}
}

func TestPillarEpochHistory_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var h PillarEpochHistory
	if err := json.Unmarshal([]byte(`[bad`), &h); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestDelegationInfo_UnmarshalJSON(t *testing.T) {
	raw := `{"name":"MyPillar","status":1,"weight":"1500000000"}`

	var d DelegationInfo
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if d.Name != "MyPillar" {
		t.Errorf("Name = %s, want MyPillar", d.Name)
	}
	if d.Status != 1 {
		t.Errorf("Status = %d, want 1", d.Status)
	}
	if d.Weight == nil || d.Weight.String() != "1500000000" {
		t.Errorf("Weight = %v, want 1500000000", d.Weight)
	}
}

func TestDelegationInfo_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var d DelegationInfo
	if err := json.Unmarshal([]byte(`}`), &d); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestDelegationInfo_IsPillarActive(t *testing.T) {
	tests := []struct {
		status int32
		want   bool
	}{
		{1, true},
		{0, false},
		{2, false},
		{-1, false},
	}
	for _, tc := range tests {
		d := &DelegationInfo{Status: tc.status}
		got := d.IsPillarActive()
		if got != tc.want {
			t.Errorf("IsPillarActive() with status=%d = %v, want %v", tc.status, got, tc.want)
		}
	}
}

// =============================================================================
// Plasma Types Tests
// =============================================================================

func TestPlasmaInfo_UnmarshalJSON(t *testing.T) {
	raw := `{"currentPlasma":21000,"maxPlasma":42000,"qsrAmount":"1000000000"}`

	var p PlasmaInfo
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if p.CurrentPlasma != 21000 {
		t.Errorf("CurrentPlasma = %d, want 21000", p.CurrentPlasma)
	}
	if p.MaxPlasma != 42000 {
		t.Errorf("MaxPlasma = %d, want 42000", p.MaxPlasma)
	}
	if p.QsrAmount == nil || p.QsrAmount.String() != "1000000000" {
		t.Errorf("QsrAmount = %v, want 1000000000", p.QsrAmount)
	}
}

func TestPlasmaInfo_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var p PlasmaInfo
	if err := json.Unmarshal([]byte(`not json`), &p); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestFusionEntry_UnmarshalJSON(t *testing.T) {
	beneficiary := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	id := types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	raw := `{
		"qsrAmount":"500000000",
		"beneficiary":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"expirationHeight":1000,
		"id":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	}`

	var f FusionEntry
	if err := json.Unmarshal([]byte(raw), &f); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if f.QsrAmount == nil || f.QsrAmount.String() != "500000000" {
		t.Errorf("QsrAmount = %v, want 500000000", f.QsrAmount)
	}
	if f.Beneficiary != beneficiary {
		t.Errorf("Beneficiary = %s, want %s", f.Beneficiary, beneficiary)
	}
	if f.ExpirationHeight != 1000 {
		t.Errorf("ExpirationHeight = %d, want 1000", f.ExpirationHeight)
	}
	if f.Id != id {
		t.Errorf("Id = %s, want %s", f.Id, id)
	}
}

func TestFusionEntry_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var f FusionEntry
	if err := json.Unmarshal([]byte(`{`), &f); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestFusionEntryList_UnmarshalJSON(t *testing.T) {
	raw := `{
		"qsrAmount":"2000000000",
		"count":2,
		"list":[
			{"qsrAmount":"1000000000","beneficiary":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz","expirationHeight":500,"id":"0000000000000000000000000000000000000000000000000000000000000001"},
			{"qsrAmount":"1000000000","beneficiary":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz","expirationHeight":600,"id":"0000000000000000000000000000000000000000000000000000000000000002"}
		]
	}`

	var l FusionEntryList
	if err := json.Unmarshal([]byte(raw), &l); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if l.QsrAmount == nil || l.QsrAmount.String() != "2000000000" {
		t.Errorf("QsrAmount = %v, want 2000000000", l.QsrAmount)
	}
	if l.Count != 2 {
		t.Errorf("Count = %d, want 2", l.Count)
	}
	if len(l.List) != 2 {
		t.Errorf("len(List) = %d, want 2", len(l.List))
	}
}

func TestFusionEntryList_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var l FusionEntryList
	if err := json.Unmarshal([]byte(`{invalid}`), &l); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// =============================================================================
// Stake Types Tests
// =============================================================================

func TestStakeEntry_UnmarshalJSON(t *testing.T) {
	addr := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	id := types.HexToHashPanic("aaaa000000000000000000000000000000000000000000000000000000000001")
	raw := `{
		"amount":"100000000",
		"weightedAmount":"100000000",
		"startTimestamp":1000000,
		"expirationTimestamp":1030000,
		"address":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"id":"aaaa000000000000000000000000000000000000000000000000000000000001"
	}`

	var s StakeEntry
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if s.Amount == nil || s.Amount.String() != "100000000" {
		t.Errorf("Amount = %v, want 100000000", s.Amount)
	}
	if s.WeightedAmount == nil || s.WeightedAmount.String() != "100000000" {
		t.Errorf("WeightedAmount = %v, want 100000000", s.WeightedAmount)
	}
	if s.StartTimestamp != 1000000 {
		t.Errorf("StartTimestamp = %d, want 1000000", s.StartTimestamp)
	}
	if s.ExpirationTimestamp != 1030000 {
		t.Errorf("ExpirationTimestamp = %d, want 1030000", s.ExpirationTimestamp)
	}
	if s.Address != addr {
		t.Errorf("Address = %s, want %s", s.Address, addr)
	}
	if s.Id != id {
		t.Errorf("Id = %s, want %s", s.Id, id)
	}
}

func TestStakeEntry_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var s StakeEntry
	if err := json.Unmarshal([]byte(`{bad}`), &s); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestStakeList_UnmarshalJSON(t *testing.T) {
	raw := `{
		"totalAmount":"300000000",
		"totalWeightedAmount":"900000000",
		"count":3,
		"list":[
			{"amount":"100000000","weightedAmount":"300000000","startTimestamp":1000,"expirationTimestamp":2000,"address":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz","id":"0000000000000000000000000000000000000000000000000000000000000001"}
		]
	}`

	var s StakeList
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if s.TotalAmount == nil || s.TotalAmount.String() != "300000000" {
		t.Errorf("TotalAmount = %v, want 300000000", s.TotalAmount)
	}
	if s.TotalWeightedAmount == nil || s.TotalWeightedAmount.String() != "900000000" {
		t.Errorf("TotalWeightedAmount = %v, want 900000000", s.TotalWeightedAmount)
	}
	if s.Count != 3 {
		t.Errorf("Count = %d, want 3", s.Count)
	}
}

func TestStakeList_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var s StakeList
	if err := json.Unmarshal([]byte(`{bad`), &s); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// =============================================================================
// Liquidity Types Tests
// =============================================================================

func TestTokenTuple_UnmarshalJSON(t *testing.T) {
	raw := `{
		"tokenStandard":"zts1znnxxxxxxxxxxxxx9z4ulx",
		"znnPercentage":5000,
		"qsrPercentage":5000,
		"minAmount":"100000000"
	}`

	var tt TokenTuple
	if err := json.Unmarshal([]byte(raw), &tt); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if tt.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", tt.TokenStandard.String())
	}
	if tt.ZnnPercentage != 5000 {
		t.Errorf("ZnnPercentage = %d, want 5000", tt.ZnnPercentage)
	}
	if tt.QsrPercentage != 5000 {
		t.Errorf("QsrPercentage = %d, want 5000", tt.QsrPercentage)
	}
	if tt.MinAmount == nil || tt.MinAmount.String() != "100000000" {
		t.Errorf("MinAmount = %v, want 100000000", tt.MinAmount)
	}
}

func TestTokenTuple_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var tt TokenTuple
	if err := json.Unmarshal([]byte(`{invalid}`), &tt); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestLiquidityInfo_UnmarshalJSON(t *testing.T) {
	admin := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	raw := `{
		"administrator":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"isHalted":false,
		"znnReward":"1000000000",
		"qsrReward":"500000000",
		"tokenTuples":[
			{"tokenStandard":"zts1znnxxxxxxxxxxxxx9z4ulx","znnPercentage":10000,"qsrPercentage":0,"minAmount":"100000000"}
		]
	}`

	var l LiquidityInfo
	if err := json.Unmarshal([]byte(raw), &l); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if l.Administrator != admin {
		t.Errorf("Administrator = %s, want %s", l.Administrator, admin)
	}
	if l.IsHalted {
		t.Error("IsHalted should be false")
	}
	if l.ZnnReward == nil || l.ZnnReward.String() != "1000000000" {
		t.Errorf("ZnnReward = %v, want 1000000000", l.ZnnReward)
	}
	if l.QsrReward == nil || l.QsrReward.String() != "500000000" {
		t.Errorf("QsrReward = %v, want 500000000", l.QsrReward)
	}
	if len(l.TokenTuples) != 1 {
		t.Errorf("len(TokenTuples) = %d, want 1", len(l.TokenTuples))
	}
}

func TestLiquidityInfo_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var l LiquidityInfo
	if err := json.Unmarshal([]byte(`[not json]`), &l); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestLiquidityStakeEntry_UnmarshalJSON(t *testing.T) {
	addr := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	id := types.HexToHashPanic("bbbb000000000000000000000000000000000000000000000000000000000001")
	raw := `{
		"amount":"200000000",
		"tokenStandard":"zts1znnxxxxxxxxxxxxx9z4ulx",
		"weightedAmount":"200000000",
		"startTime":1000000,
		"revokeTime":0,
		"expirationTime":1030000,
		"stakeAddress":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"id":"bbbb000000000000000000000000000000000000000000000000000000000001"
	}`

	var l LiquidityStakeEntry
	if err := json.Unmarshal([]byte(raw), &l); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if l.Amount == nil || l.Amount.String() != "200000000" {
		t.Errorf("Amount = %v, want 200000000", l.Amount)
	}
	if l.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", l.TokenStandard.String())
	}
	if l.StartTime != 1000000 {
		t.Errorf("StartTime = %d, want 1000000", l.StartTime)
	}
	if l.RevokeTime != 0 {
		t.Errorf("RevokeTime = %d, want 0", l.RevokeTime)
	}
	if l.StakeAddress != addr {
		t.Errorf("StakeAddress = %s, want %s", l.StakeAddress, addr)
	}
	if l.Id != id {
		t.Errorf("Id = %s, want %s", l.Id, id)
	}
}

func TestLiquidityStakeEntry_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var l LiquidityStakeEntry
	if err := json.Unmarshal([]byte(`{bad}`), &l); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestLiquidityStakeList_UnmarshalJSON(t *testing.T) {
	raw := `{
		"totalAmount":"400000000",
		"totalWeightedAmount":"400000000",
		"count":2,
		"list":[]
	}`

	var l LiquidityStakeList
	if err := json.Unmarshal([]byte(raw), &l); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if l.TotalAmount == nil || l.TotalAmount.String() != "400000000" {
		t.Errorf("TotalAmount = %v, want 400000000", l.TotalAmount)
	}
	if l.TotalWeightedAmount == nil || l.TotalWeightedAmount.String() != "400000000" {
		t.Errorf("TotalWeightedAmount = %v, want 400000000", l.TotalWeightedAmount)
	}
	if l.Count != 2 {
		t.Errorf("Count = %d, want 2", l.Count)
	}
}

func TestLiquidityStakeList_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var l LiquidityStakeList
	if err := json.Unmarshal([]byte(`bad`), &l); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// =============================================================================
// Accelerator Types Tests
// =============================================================================

func TestPhaseInfo_UnmarshalJSON(t *testing.T) {
	id := types.HexToHashPanic("cccc000000000000000000000000000000000000000000000000000000000001")
	projectID := types.HexToHashPanic("dddd000000000000000000000000000000000000000000000000000000000002")
	raw := `{
		"id":"cccc000000000000000000000000000000000000000000000000000000000001",
		"projectID":"dddd000000000000000000000000000000000000000000000000000000000002",
		"name":"Phase 1",
		"description":"First phase",
		"url":"https://example.com",
		"znnFundsNeeded":"500000000",
		"qsrFundsNeeded":"200000000",
		"creationTimestamp":1000000,
		"acceptedTimestamp":0,
		"status":1
	}`

	var p PhaseInfo
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if p.Id != id {
		t.Errorf("Id = %s, want %s", p.Id, id)
	}
	if p.ProjectID != projectID {
		t.Errorf("ProjectID = %s, want %s", p.ProjectID, projectID)
	}
	if p.Name != "Phase 1" {
		t.Errorf("Name = %s, want Phase 1", p.Name)
	}
	if p.ZnnFundsNeeded == nil || p.ZnnFundsNeeded.String() != "500000000" {
		t.Errorf("ZnnFundsNeeded = %v, want 500000000", p.ZnnFundsNeeded)
	}
	if p.QsrFundsNeeded == nil || p.QsrFundsNeeded.String() != "200000000" {
		t.Errorf("QsrFundsNeeded = %v, want 200000000", p.QsrFundsNeeded)
	}
	if p.Status != 1 {
		t.Errorf("Status = %d, want 1", p.Status)
	}
}

func TestPhaseInfo_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var p PhaseInfo
	if err := json.Unmarshal([]byte(`{bad}`), &p); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestProject_UnmarshalJSON(t *testing.T) {
	id := types.HexToHashPanic("eeee000000000000000000000000000000000000000000000000000000000001")
	owner := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	raw := `{
		"id":"eeee000000000000000000000000000000000000000000000000000000000001",
		"owner":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"name":"My Project",
		"description":"A test project",
		"url":"https://example.com",
		"znnFundsNeeded":"1000000000",
		"qsrFundsNeeded":"500000000",
		"creationTimestamp":1000000,
		"lastUpdateTimestamp":1000100,
		"status":0,
		"phaseIds":[],
		"votes":{"id":"eeee000000000000000000000000000000000000000000000000000000000001","total":5,"yes":4,"no":1},
		"phases":[]
	}`

	var p Project
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if p.Id != id {
		t.Errorf("Id = %s, want %s", p.Id, id)
	}
	if p.Owner != owner {
		t.Errorf("Owner = %s, want %s", p.Owner, owner)
	}
	if p.Name != "My Project" {
		t.Errorf("Name = %s, want My Project", p.Name)
	}
	if p.ZnnFundsNeeded == nil || p.ZnnFundsNeeded.String() != "1000000000" {
		t.Errorf("ZnnFundsNeeded = %v, want 1000000000", p.ZnnFundsNeeded)
	}
	if p.QsrFundsNeeded == nil || p.QsrFundsNeeded.String() != "500000000" {
		t.Errorf("QsrFundsNeeded = %v, want 500000000", p.QsrFundsNeeded)
	}
	if p.Votes == nil {
		t.Fatal("Votes should not be nil")
	}
	if p.Votes.Total != 5 || p.Votes.Yes != 4 || p.Votes.No != 1 {
		t.Errorf("Votes = %+v, want total=5 yes=4 no=1", p.Votes)
	}
}

func TestProject_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var p Project
	if err := json.Unmarshal([]byte(`{invalid}`), &p); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// =============================================================================
// Bridge Types Tests
// =============================================================================

func TestTokenPair_UnmarshalJSON(t *testing.T) {
	raw := `{
		"tokenStandard":"zts1znnxxxxxxxxxxxxx9z4ulx",
		"tokenAddress":"0x1234567890abcdef",
		"bridgeable":true,
		"redeemable":true,
		"owned":false,
		"minAmount":"100000000",
		"feePercentage":100,
		"redeemDelay":20,
		"metadata":"{}"
	}`

	var tp TokenPair
	if err := json.Unmarshal([]byte(raw), &tp); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if tp.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", tp.TokenStandard.String())
	}
	if !tp.Bridgeable {
		t.Error("Bridgeable should be true")
	}
	if tp.MinAmount == nil || tp.MinAmount.String() != "100000000" {
		t.Errorf("MinAmount = %v, want 100000000", tp.MinAmount)
	}
	if tp.FeePercentage != 100 {
		t.Errorf("FeePercentage = %d, want 100", tp.FeePercentage)
	}
}

func TestTokenPair_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var tp TokenPair
	if err := json.Unmarshal([]byte(`{bad`), &tp); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestWrapTokenRequest_UnmarshalJSON(t *testing.T) {
	id := types.HexToHashPanic("ffff000000000000000000000000000000000000000000000000000000000001")
	raw := `{
		"networkClass":2,
		"chainId":1,
		"id":"ffff000000000000000000000000000000000000000000000000000000000001",
		"toAddress":"0xabcdef",
		"tokenStandard":"zts1znnxxxxxxxxxxxxx9z4ulx",
		"tokenAddress":"0x1234",
		"amount":"1000000000",
		"fee":"1000000",
		"signature":"sig123",
		"creationMomentumHeight":100,
		"confirmationsToFinality":5
	}`

	var w WrapTokenRequest
	if err := json.Unmarshal([]byte(raw), &w); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if w.NetworkClass != 2 {
		t.Errorf("NetworkClass = %d, want 2", w.NetworkClass)
	}
	if w.ChainId != 1 {
		t.Errorf("ChainId = %d, want 1", w.ChainId)
	}
	if w.Id != id {
		t.Errorf("Id = %s, want %s", w.Id, id)
	}
	if w.Amount == nil || w.Amount.String() != "1000000000" {
		t.Errorf("Amount = %v, want 1000000000", w.Amount)
	}
	if w.Fee == nil || w.Fee.String() != "1000000" {
		t.Errorf("Fee = %v, want 1000000", w.Fee)
	}
	if w.Signature != "sig123" {
		t.Errorf("Signature = %s, want sig123", w.Signature)
	}
	if w.ConfirmationsToFinality != 5 {
		t.Errorf("ConfirmationsToFinality = %d, want 5", w.ConfirmationsToFinality)
	}
}

func TestWrapTokenRequest_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var w WrapTokenRequest
	if err := json.Unmarshal([]byte(`bad`), &w); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestUnwrapTokenRequest_UnmarshalJSON(t *testing.T) {
	txHash := types.HexToHashPanic("1111000000000000000000000000000000000000000000000000000000000001")
	toAddr := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	raw := `{
		"registrationMomentumHeight":500,
		"networkClass":2,
		"chainId":1,
		"transactionHash":"1111000000000000000000000000000000000000000000000000000000000001",
		"logIndex":3,
		"toAddress":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"tokenAddress":"0x1234",
		"tokenStandard":"zts1znnxxxxxxxxxxxxx9z4ulx",
		"amount":"750000000",
		"signature":"sig456",
		"redeemed":0,
		"revoked":0,
		"redeemableIn":10
	}`

	var u UnwrapTokenRequest
	if err := json.Unmarshal([]byte(raw), &u); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if u.RegistrationMomentumHeight != 500 {
		t.Errorf("RegistrationMomentumHeight = %d, want 500", u.RegistrationMomentumHeight)
	}
	if u.TransactionHash != txHash {
		t.Errorf("TransactionHash = %s, want %s", u.TransactionHash, txHash)
	}
	if u.LogIndex != 3 {
		t.Errorf("LogIndex = %d, want 3", u.LogIndex)
	}
	if u.ToAddress != toAddr {
		t.Errorf("ToAddress = %s, want %s", u.ToAddress, toAddr)
	}
	if u.Amount == nil || u.Amount.String() != "750000000" {
		t.Errorf("Amount = %v, want 750000000", u.Amount)
	}
	if u.RedeemableIn != 10 {
		t.Errorf("RedeemableIn = %d, want 10", u.RedeemableIn)
	}
}

func TestUnwrapTokenRequest_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var u UnwrapTokenRequest
	if err := json.Unmarshal([]byte(`{bad}`), &u); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// =============================================================================
// Swap Types Tests
// =============================================================================

func TestSwapAssetEntry_UnmarshalJSON(t *testing.T) {
	keyIdHash := types.HexToHashPanic("2222000000000000000000000000000000000000000000000000000000000001")
	raw := `{
		"keyIdHash":"2222000000000000000000000000000000000000000000000000000000000001",
		"qsr":"300000000",
		"znn":"150000000"
	}`

	var s SwapAssetEntry
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if s.KeyIdHash != keyIdHash {
		t.Errorf("KeyIdHash = %s, want %s", s.KeyIdHash, keyIdHash)
	}
	if s.Qsr == nil || s.Qsr.String() != "300000000" {
		t.Errorf("Qsr = %v, want 300000000", s.Qsr)
	}
	if s.Znn == nil || s.Znn.String() != "150000000" {
		t.Errorf("Znn = %v, want 150000000", s.Znn)
	}
}

func TestSwapAssetEntry_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var s SwapAssetEntry
	if err := json.Unmarshal([]byte(`{bad}`), &s); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestSwapAssetEntry_HasBalance(t *testing.T) {
	tests := []struct {
		name string
		qsr  *big.Int
		znn  *big.Int
		want bool
	}{
		{"both positive", big.NewInt(100), big.NewInt(100), true},
		{"only qsr", big.NewInt(100), big.NewInt(0), true},
		{"only znn", big.NewInt(0), big.NewInt(100), true},
		{"both zero", big.NewInt(0), big.NewInt(0), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &SwapAssetEntry{Qsr: tc.qsr, Znn: tc.znn}
			got := s.HasBalance()
			if got != tc.want {
				t.Errorf("HasBalance() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSwapAssetEntrySimple_UnmarshalJSON(t *testing.T) {
	raw := `{"qsr":"400000000","znn":"200000000"}`

	var s SwapAssetEntrySimple
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if s.Qsr == nil || s.Qsr.String() != "400000000" {
		t.Errorf("Qsr = %v, want 400000000", s.Qsr)
	}
	if s.Znn == nil || s.Znn.String() != "200000000" {
		t.Errorf("Znn = %v, want 200000000", s.Znn)
	}
}

func TestSwapAssetEntrySimple_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var s SwapAssetEntrySimple
	if err := json.Unmarshal([]byte(`{bad}`), &s); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// =============================================================================
// HTLC Types Tests
// =============================================================================

func TestHtlcInfo_UnmarshalJSON(t *testing.T) {
	id := types.HexToHashPanic("3333000000000000000000000000000000000000000000000000000000000001")
	timeLocked := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	hashLocked := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	// base64 of "hello" = "aGVsbG8="
	raw := `{
		"id":"3333000000000000000000000000000000000000000000000000000000000001",
		"timeLocked":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"hashLocked":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"tokenStandard":"zts1znnxxxxxxxxxxxxx9z4ulx",
		"amount":"50000000",
		"expirationTime":9999999,
		"hashType":0,
		"keyMaxSize":32,
		"hashLock":"aGVsbG8="
	}`

	var h HtlcInfo
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if h.Id != id {
		t.Errorf("Id = %s, want %s", h.Id, id)
	}
	if h.TimeLocked != timeLocked {
		t.Errorf("TimeLocked = %s, want %s", h.TimeLocked, timeLocked)
	}
	if h.HashLocked != hashLocked {
		t.Errorf("HashLocked = %s, want %s", h.HashLocked, hashLocked)
	}
	if h.Amount == nil || h.Amount.String() != "50000000" {
		t.Errorf("Amount = %v, want 50000000", h.Amount)
	}
	if h.ExpirationTime != 9999999 {
		t.Errorf("ExpirationTime = %d, want 9999999", h.ExpirationTime)
	}
	if string(h.HashLock) != "hello" {
		t.Errorf("HashLock = %s, want hello", string(h.HashLock))
	}
}

func TestHtlcInfo_UnmarshalJSON_EmptyHashLock(t *testing.T) {
	raw := `{
		"id":"3333000000000000000000000000000000000000000000000000000000000001",
		"timeLocked":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"hashLocked":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"tokenStandard":"zts1znnxxxxxxxxxxxxx9z4ulx",
		"amount":"50000000",
		"expirationTime":9999999,
		"hashType":0,
		"keyMaxSize":32,
		"hashLock":""
	}`

	var h HtlcInfo
	if err := json.Unmarshal([]byte(raw), &h); err != nil {
		t.Fatalf("UnmarshalJSON with empty hashLock failed: %v", err)
	}
	if h.HashLock != nil {
		t.Errorf("HashLock should be nil for empty string, got %v", h.HashLock)
	}
}

func TestHtlcInfo_UnmarshalJSON_InvalidBase64(t *testing.T) {
	raw := `{
		"id":"3333000000000000000000000000000000000000000000000000000000000001",
		"timeLocked":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"hashLocked":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"tokenStandard":"zts1znnxxxxxxxxxxxxx9z4ulx",
		"amount":"50000000",
		"expirationTime":9999999,
		"hashType":0,
		"keyMaxSize":32,
		"hashLock":"not-valid-base64!!!"
	}`

	var h HtlcInfo
	if err := json.Unmarshal([]byte(raw), &h); err == nil {
		t.Error("Expected error for invalid base64 hashLock")
	}
}

func TestHtlcInfo_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var h HtlcInfo
	if err := json.Unmarshal([]byte(`{bad}`), &h); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// =============================================================================
// Token Types Tests
// =============================================================================

func TestToken_UnmarshalJSON(t *testing.T) {
	owner := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	raw := `{
		"name":"Zenon",
		"symbol":"ZNN",
		"domain":"zenon.network",
		"totalSupply":"10000000000000",
		"decimals":8,
		"owner":"z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz",
		"tokenStandard":"zts1znnxxxxxxxxxxxxx9z4ulx",
		"maxSupply":"10000000000000",
		"isBurnable":true,
		"isMintable":false,
		"isUtility":true
	}`

	var tok Token
	if err := json.Unmarshal([]byte(raw), &tok); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if tok.Name != "Zenon" {
		t.Errorf("Name = %s, want Zenon", tok.Name)
	}
	if tok.Symbol != "ZNN" {
		t.Errorf("Symbol = %s, want ZNN", tok.Symbol)
	}
	if tok.Domain != "zenon.network" {
		t.Errorf("Domain = %s, want zenon.network", tok.Domain)
	}
	if tok.TotalSupply == nil || tok.TotalSupply.String() != "10000000000000" {
		t.Errorf("TotalSupply = %v, want 10000000000000", tok.TotalSupply)
	}
	if tok.Decimals != 8 {
		t.Errorf("Decimals = %d, want 8", tok.Decimals)
	}
	if tok.Owner != owner {
		t.Errorf("Owner = %s, want %s", tok.Owner, owner)
	}
	if tok.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", tok.TokenStandard.String())
	}
	if tok.MaxSupply == nil || tok.MaxSupply.String() != "10000000000000" {
		t.Errorf("MaxSupply = %v, want 10000000000000", tok.MaxSupply)
	}
	if !tok.IsBurnable {
		t.Error("IsBurnable should be true")
	}
	if tok.IsMintable {
		t.Error("IsMintable should be false")
	}
	if !tok.IsUtility {
		t.Error("IsUtility should be true")
	}
}

func TestToken_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var tok Token
	if err := json.Unmarshal([]byte(`{bad}`), &tok); err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// =============================================================================
// Pillar Type Constants Tests
// =============================================================================

func TestPillarTypeConstants(t *testing.T) {
	if UnknownPillarType != 0 {
		t.Errorf("UnknownPillarType = %d, want 0", UnknownPillarType)
	}
	if LegacyPillarType != 1 {
		t.Errorf("LegacyPillarType = %d, want 1", LegacyPillarType)
	}
	if RegularPillarType != 2 {
		t.Errorf("RegularPillarType = %d, want 2", RegularPillarType)
	}
}
