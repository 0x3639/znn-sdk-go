package embedded

import (
	"math/big"
	"testing"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
)

// =============================================================================
// Constructor Tests
// =============================================================================

func TestNewHtlcApi(t *testing.T) {
	api := NewHtlcApi(nil)
	if api == nil {
		t.Fatal("NewHtlcApi() should not return nil")
	}

	if api.client != nil {
		t.Error("NewHtlcApi(nil) should have nil client")
	}
}

// =============================================================================
// Create Tests
// =============================================================================

func TestHtlcApi_Create(t *testing.T) {
	api := NewHtlcApi(nil)

	hashLocked := types.Address{}
	copy(hashLocked[:], []byte("test_address_1234567890"))

	hashLock := make([]byte, 32)
	for i := range hashLock {
		hashLock[i] = byte(i)
	}

	amount := big.NewInt(1000000000) // 10 ZNN
	expirationTime := int64(1700000000)
	hashType := uint8(0) // SHA3-256
	keyMaxSize := uint8(32)

	block := api.Create(
		types.ZnnTokenStandard,
		amount,
		hashLocked,
		expirationTime,
		hashType,
		keyMaxSize,
		hashLock,
	)

	if block == nil {
		t.Fatal("Create() should not return nil")
	}

	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}

	if block.ToAddress != types.HtlcContract {
		t.Errorf("ToAddress = %s, want HtlcContract", block.ToAddress.String())
	}

	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard.String())
	}

	if block.Amount.Cmp(amount) != 0 {
		t.Errorf("Amount = %s, want %s", block.Amount.String(), amount.String())
	}

	if len(block.Data) == 0 {
		t.Error("Data should not be empty")
	}
}

func TestHtlcApi_Create_QSR(t *testing.T) {
	api := NewHtlcApi(nil)

	hashLocked := types.Address{}
	hashLock := make([]byte, 32)
	amount := big.NewInt(5000000000) // 50 QSR

	block := api.Create(
		types.QsrTokenStandard,
		amount,
		hashLocked,
		1700000000,
		0,
		32,
		hashLock,
	)

	if block.TokenStandard != types.QsrTokenStandard {
		t.Errorf("TokenStandard = %s, want QSR", block.TokenStandard.String())
	}

	if block.Amount.Cmp(amount) != 0 {
		t.Errorf("Amount = %s, want %s", block.Amount.String(), amount.String())
	}
}

func TestHtlcApi_Create_SHA256HashType(t *testing.T) {
	api := NewHtlcApi(nil)

	hashLocked := types.Address{}
	hashLock := make([]byte, 32)

	// Test with SHA-256 hash type
	block := api.Create(
		types.ZnnTokenStandard,
		big.NewInt(1000000000),
		hashLocked,
		1700000000,
		1, // SHA-256
		32,
		hashLock,
	)

	if block == nil {
		t.Fatal("Create() with SHA-256 should not return nil")
	}

	if len(block.Data) == 0 {
		t.Error("Data should not be empty")
	}
}

// =============================================================================
// Reclaim Tests
// =============================================================================

func TestHtlcApi_Reclaim(t *testing.T) {
	api := NewHtlcApi(nil)

	htlcId := types.Hash{}
	copy(htlcId[:], []byte("test_htlc_id_1234567890123"))

	block := api.Reclaim(htlcId)

	if block == nil {
		t.Fatal("Reclaim() should not return nil")
	}

	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}

	if block.ToAddress != types.HtlcContract {
		t.Errorf("ToAddress = %s, want HtlcContract", block.ToAddress.String())
	}

	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard.String())
	}

	if block.Amount.Sign() != 0 {
		t.Errorf("Amount = %s, want 0", block.Amount.String())
	}

	if len(block.Data) == 0 {
		t.Error("Data should not be empty")
	}
}

func TestHtlcApi_Reclaim_ZeroHash(t *testing.T) {
	api := NewHtlcApi(nil)

	block := api.Reclaim(types.ZeroHash)

	if block == nil {
		t.Fatal("Reclaim() with zero hash should not return nil")
	}

	if len(block.Data) == 0 {
		t.Error("Data should not be empty")
	}
}

// =============================================================================
// Unlock Tests
// =============================================================================

func TestHtlcApi_Unlock(t *testing.T) {
	api := NewHtlcApi(nil)

	htlcId := types.Hash{}
	copy(htlcId[:], []byte("test_htlc_id_1234567890123"))

	preimage := []byte("secret_preimage_key")

	block := api.Unlock(htlcId, preimage)

	if block == nil {
		t.Fatal("Unlock() should not return nil")
	}

	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}

	if block.ToAddress != types.HtlcContract {
		t.Errorf("ToAddress = %s, want HtlcContract", block.ToAddress.String())
	}

	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard.String())
	}

	if block.Amount.Sign() != 0 {
		t.Errorf("Amount = %s, want 0", block.Amount.String())
	}

	if len(block.Data) == 0 {
		t.Error("Data should not be empty")
	}
}

func TestHtlcApi_Unlock_EmptyPreimage(t *testing.T) {
	api := NewHtlcApi(nil)

	htlcId := types.Hash{}
	preimage := []byte{}

	block := api.Unlock(htlcId, preimage)

	if block == nil {
		t.Fatal("Unlock() with empty preimage should not return nil")
	}

	if len(block.Data) == 0 {
		t.Error("Data should not be empty")
	}
}

func TestHtlcApi_Unlock_LargePreimage(t *testing.T) {
	api := NewHtlcApi(nil)

	htlcId := types.Hash{}
	preimage := make([]byte, 255) // Maximum size
	for i := range preimage {
		preimage[i] = byte(i)
	}

	block := api.Unlock(htlcId, preimage)

	if block == nil {
		t.Fatal("Unlock() with large preimage should not return nil")
	}

	if len(block.Data) == 0 {
		t.Error("Data should not be empty")
	}
}

// =============================================================================
// DenyProxyUnlock Tests
// =============================================================================

func TestHtlcApi_DenyProxyUnlock(t *testing.T) {
	api := NewHtlcApi(nil)

	block := api.DenyProxyUnlock()

	if block == nil {
		t.Fatal("DenyProxyUnlock() should not return nil")
	}

	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}

	if block.ToAddress != types.HtlcContract {
		t.Errorf("ToAddress = %s, want HtlcContract", block.ToAddress.String())
	}

	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard.String())
	}

	if block.Amount.Sign() != 0 {
		t.Errorf("Amount = %s, want 0", block.Amount.String())
	}

	if len(block.Data) == 0 {
		t.Error("Data should not be empty")
	}
}

// =============================================================================
// AllowProxyUnlock Tests
// =============================================================================

func TestHtlcApi_AllowProxyUnlock(t *testing.T) {
	api := NewHtlcApi(nil)

	block := api.AllowProxyUnlock()

	if block == nil {
		t.Fatal("AllowProxyUnlock() should not return nil")
	}

	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}

	if block.ToAddress != types.HtlcContract {
		t.Errorf("ToAddress = %s, want HtlcContract", block.ToAddress.String())
	}

	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard.String())
	}

	if block.Amount.Sign() != 0 {
		t.Errorf("Amount = %s, want 0", block.Amount.String())
	}

	if len(block.Data) == 0 {
		t.Error("Data should not be empty")
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestHtlcApi_FullWorkflow(t *testing.T) {
	api := NewHtlcApi(nil)

	// 1. Create HTLC
	hashLocked := types.Address{}
	hashLock := make([]byte, 32)
	amount := big.NewInt(1000000000)

	createBlock := api.Create(
		types.ZnnTokenStandard,
		amount,
		hashLocked,
		1700000000,
		0,
		32,
		hashLock,
	)

	if createBlock == nil {
		t.Fatal("Create() failed")
	}

	// 2. Unlock with preimage
	htlcId := types.Hash{}
	preimage := []byte("secret_key")

	unlockBlock := api.Unlock(htlcId, preimage)

	if unlockBlock == nil {
		t.Fatal("Unlock() failed")
	}

	// 3. Reclaim if expired
	reclaimBlock := api.Reclaim(htlcId)

	if reclaimBlock == nil {
		t.Fatal("Reclaim() failed")
	}

	// 4. Proxy unlock management
	denyBlock := api.DenyProxyUnlock()
	if denyBlock == nil {
		t.Fatal("DenyProxyUnlock() failed")
	}

	allowBlock := api.AllowProxyUnlock()
	if allowBlock == nil {
		t.Fatal("AllowProxyUnlock() failed")
	}
}
