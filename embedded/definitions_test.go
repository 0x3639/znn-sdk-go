package embedded

import (
	"testing"
)

// =============================================================================
// ABI Parsing Tests
// =============================================================================

func TestPlasma_ParseSuccess(t *testing.T) {
	if Plasma == nil {
		t.Fatal("Plasma ABI is nil")
	}

	if len(Plasma.Entries) != 2 {
		t.Errorf("Plasma has %d entries, want 2", len(Plasma.Entries))
	}

	// Check Fuse function
	if Plasma.Entries[0].Name != "Fuse" {
		t.Errorf("Plasma.Entries[0].Name = %s, want Fuse", Plasma.Entries[0].Name)
	}

	// Check CancelFuse function
	if Plasma.Entries[1].Name != "CancelFuse" {
		t.Errorf("Plasma.Entries[1].Name = %s, want CancelFuse", Plasma.Entries[1].Name)
	}
}

func TestPillar_ParseSuccess(t *testing.T) {
	if Pillar == nil {
		t.Fatal("Pillar ABI is nil")
	}

	if len(Pillar.Entries) != 6 {
		t.Errorf("Pillar has %d entries, want 6", len(Pillar.Entries))
	}

	expectedNames := []string{"Register", "RegisterLegacy", "Revoke", "UpdatePillar", "Delegate", "Undelegate"}
	for i, expected := range expectedNames {
		if Pillar.Entries[i].Name != expected {
			t.Errorf("Pillar.Entries[%d].Name = %s, want %s", i, Pillar.Entries[i].Name, expected)
		}
	}
}

func TestToken_ParseSuccess(t *testing.T) {
	if Token == nil {
		t.Fatal("Token ABI is nil")
	}

	if len(Token.Entries) != 4 {
		t.Errorf("Token has %d entries, want 4", len(Token.Entries))
	}

	expectedNames := []string{"IssueToken", "Mint", "Burn", "UpdateToken"}
	for i, expected := range expectedNames {
		if Token.Entries[i].Name != expected {
			t.Errorf("Token.Entries[%d].Name = %s, want %s", i, Token.Entries[i].Name, expected)
		}
	}
}

func TestToken_IssueTokenEncoding(t *testing.T) {
	if Token == nil {
		t.Fatal("Token ABI is nil")
	}

	// Test encoding IssueToken function
	args := []interface{}{
		"TestToken", // tokenName
		"TST",       // tokenSymbol
		"test.com",  // tokenDomain
		1000000,     // totalSupply
		1000000,     // maxSupply
		8,           // decimals
		false,       // isMintable
		true,        // isBurnable
		false,       // isUtility
	}

	encoded, err := Token.EncodeFunction("IssueToken", args)
	if err != nil {
		t.Fatalf("Token.EncodeFunction(IssueToken) error = %v", err)
	}

	// Should have signature (4 bytes) + encoded arguments
	if len(encoded) < 4 {
		t.Errorf("encoded length = %d, expected > 4", len(encoded))
	}

	// Verify we can decode it back
	decoded, err := Token.DecodeFunction(encoded)
	if err != nil {
		t.Fatalf("Token.DecodeFunction() error = %v", err)
	}

	if len(decoded) != 9 {
		t.Errorf("decoded length = %d, want 9", len(decoded))
	}
}

func TestSentinel_ParseSuccess(t *testing.T) {
	if Sentinel == nil {
		t.Fatal("Sentinel ABI is nil")
	}

	if len(Sentinel.Entries) != 2 {
		t.Errorf("Sentinel has %d entries, want 2", len(Sentinel.Entries))
	}

	expectedNames := []string{"Register", "Revoke"}
	for i, expected := range expectedNames {
		if Sentinel.Entries[i].Name != expected {
			t.Errorf("Sentinel.Entries[%d].Name = %s, want %s", i, Sentinel.Entries[i].Name, expected)
		}
	}
}

func TestSwap_ParseSuccess(t *testing.T) {
	if Swap == nil {
		t.Fatal("Swap ABI is nil")
	}

	if len(Swap.Entries) != 1 {
		t.Errorf("Swap has %d entries, want 1", len(Swap.Entries))
	}

	if Swap.Entries[0].Name != "RetrieveAssets" {
		t.Errorf("Swap.Entries[0].Name = %s, want RetrieveAssets", Swap.Entries[0].Name)
	}
}

func TestStake_ParseSuccess(t *testing.T) {
	if Stake == nil {
		t.Fatal("Stake ABI is nil")
	}

	if len(Stake.Entries) != 2 {
		t.Errorf("Stake has %d entries, want 2", len(Stake.Entries))
	}

	expectedNames := []string{"Stake", "Cancel"}
	for i, expected := range expectedNames {
		if Stake.Entries[i].Name != expected {
			t.Errorf("Stake.Entries[%d].Name = %s, want %s", i, Stake.Entries[i].Name, expected)
		}
	}
}

func TestAccelerator_ParseSuccess(t *testing.T) {
	if Accelerator == nil {
		t.Fatal("Accelerator ABI is nil")
	}

	if len(Accelerator.Entries) != 6 {
		t.Errorf("Accelerator has %d entries, want 6", len(Accelerator.Entries))
	}

	expectedNames := []string{"CreateProject", "AddPhase", "UpdatePhase", "Donate", "VoteByName", "VoteByProdAddress"}
	for i, expected := range expectedNames {
		if Accelerator.Entries[i].Name != expected {
			t.Errorf("Accelerator.Entries[%d].Name = %s, want %s", i, Accelerator.Entries[i].Name, expected)
		}
	}
}

func TestSpork_ParseSuccess(t *testing.T) {
	if Spork == nil {
		t.Fatal("Spork ABI is nil")
	}

	if len(Spork.Entries) != 2 {
		t.Errorf("Spork has %d entries, want 2", len(Spork.Entries))
	}

	expectedNames := []string{"CreateSpork", "ActivateSpork"}
	for i, expected := range expectedNames {
		if Spork.Entries[i].Name != expected {
			t.Errorf("Spork.Entries[%d].Name = %s, want %s", i, Spork.Entries[i].Name, expected)
		}
	}
}

func TestHtlc_ParseSuccess(t *testing.T) {
	if Htlc == nil {
		t.Fatal("Htlc ABI is nil")
	}

	if len(Htlc.Entries) != 5 {
		t.Errorf("Htlc has %d entries, want 5", len(Htlc.Entries))
	}

	expectedNames := []string{"Create", "Reclaim", "Unlock", "DenyProxyUnlock", "AllowProxyUnlock"}
	for i, expected := range expectedNames {
		if Htlc.Entries[i].Name != expected {
			t.Errorf("Htlc.Entries[%d].Name = %s, want %s", i, Htlc.Entries[i].Name, expected)
		}
	}
}

func TestBridge_ParseSuccess(t *testing.T) {
	if Bridge == nil {
		t.Fatal("Bridge ABI is nil")
	}

	// Bridge has many functions (20+)
	if len(Bridge.Entries) < 20 {
		t.Errorf("Bridge has %d entries, expected at least 20", len(Bridge.Entries))
	}

	// Check first few function names
	expectedNames := []string{"WrapToken", "UpdateWrapRequest", "SetNetwork", "RemoveNetwork"}
	for i, expected := range expectedNames {
		if i >= len(Bridge.Entries) {
			t.Fatalf("Bridge has only %d entries, expected at least %d", len(Bridge.Entries), i+1)
		}
		if Bridge.Entries[i].Name != expected {
			t.Errorf("Bridge.Entries[%d].Name = %s, want %s", i, Bridge.Entries[i].Name, expected)
		}
	}
}

func TestLiquidity_ParseSuccess(t *testing.T) {
	if Liquidity == nil {
		t.Fatal("Liquidity ABI is nil")
	}

	if len(Liquidity.Entries) != 15 {
		t.Errorf("Liquidity has %d entries, want 15", len(Liquidity.Entries))
	}

	// Check first few function names
	expectedNames := []string{"Update", "Donate", "Fund", "BurnZnn"}
	for i, expected := range expectedNames {
		if Liquidity.Entries[i].Name != expected {
			t.Errorf("Liquidity.Entries[%d].Name = %s, want %s", i, Liquidity.Entries[i].Name, expected)
		}
	}
}

func TestCommon_ParseSuccess(t *testing.T) {
	if Common == nil {
		t.Fatal("Common ABI is nil")
	}

	if len(Common.Entries) != 7 {
		t.Errorf("Common has %d entries, want 7", len(Common.Entries))
	}

	expectedNames := []string{"DepositQsr", "WithdrawQsr", "CollectReward", "Update", "Donate", "VoteByName", "VoteByProdAddress"}
	for i, expected := range expectedNames {
		if Common.Entries[i].Name != expected {
			t.Errorf("Common.Entries[%d].Name = %s, want %s", i, Common.Entries[i].Name, expected)
		}
	}
}

// =============================================================================
// Function Encoding Tests
// =============================================================================

func TestPlasma_FuseEncoding(t *testing.T) {
	args := []interface{}{
		"z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7",
	}

	encoded, err := Plasma.EncodeFunction("Fuse", args)
	if err != nil {
		t.Fatalf("Plasma.EncodeFunction(Fuse) error = %v", err)
	}

	if len(encoded) < 4 {
		t.Errorf("encoded length = %d, expected >= 4", len(encoded))
	}

	// Verify we can decode it back
	decoded, err := Plasma.DecodeFunction(encoded)
	if err != nil {
		t.Fatalf("Plasma.DecodeFunction() error = %v", err)
	}

	if len(decoded) != 1 {
		t.Errorf("decoded length = %d, want 1", len(decoded))
	}
}

func TestStake_StakeEncoding(t *testing.T) {
	args := []interface{}{
		int64(2592000), // 30 days in seconds
	}

	encoded, err := Stake.EncodeFunction("Stake", args)
	if err != nil {
		t.Fatalf("Stake.EncodeFunction(Stake) error = %v", err)
	}

	if len(encoded) < 4 {
		t.Errorf("encoded length = %d, expected >= 4", len(encoded))
	}

	// Verify we can decode it back
	decoded, err := Stake.DecodeFunction(encoded)
	if err != nil {
		t.Fatalf("Stake.DecodeFunction() error = %v", err)
	}

	if len(decoded) != 1 {
		t.Errorf("decoded length = %d, want 1", len(decoded))
	}
}

func TestPillar_DelegateEncoding(t *testing.T) {
	args := []interface{}{
		"MyPillar",
	}

	encoded, err := Pillar.EncodeFunction("Delegate", args)
	if err != nil {
		t.Fatalf("Pillar.EncodeFunction(Delegate) error = %v", err)
	}

	if len(encoded) < 4 {
		t.Errorf("encoded length = %d, expected >= 4", len(encoded))
	}

	// Verify we can decode it back
	decoded, err := Pillar.DecodeFunction(encoded)
	if err != nil {
		t.Fatalf("Pillar.DecodeFunction() error = %v", err)
	}

	if len(decoded) != 1 {
		t.Errorf("decoded length = %d, want 1", len(decoded))
	}
}

func TestSentinel_RegisterEncoding(t *testing.T) {
	// Register has no inputs
	args := []interface{}{}

	encoded, err := Sentinel.EncodeFunction("Register", args)
	if err != nil {
		t.Fatalf("Sentinel.EncodeFunction(Register) error = %v", err)
	}

	// Should just be the 4-byte signature
	if len(encoded) != 4 {
		t.Errorf("encoded length = %d, want 4", len(encoded))
	}

	// Verify we can decode it back
	decoded, err := Sentinel.DecodeFunction(encoded)
	if err != nil {
		t.Fatalf("Sentinel.DecodeFunction() error = %v", err)
	}

	if len(decoded) != 0 {
		t.Errorf("decoded length = %d, want 0", len(decoded))
	}
}
