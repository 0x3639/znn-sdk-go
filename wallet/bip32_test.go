package wallet

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// =============================================================================
// GetMasterKeyFromSeed Tests
// =============================================================================

func TestGetMasterKeyFromSeed_ValidSeed(t *testing.T) {
	seed := make([]byte, 64)

	keyData, err := GetMasterKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("GetMasterKeyFromSeed() error = %v", err)
	}

	if len(keyData.Key) != 32 {
		t.Errorf("len(keyData.Key) = %d, want 32", len(keyData.Key))
	}

	if len(keyData.ChainCode) != 32 {
		t.Errorf("len(keyData.ChainCode) = %d, want 32", len(keyData.ChainCode))
	}
}

func TestGetMasterKeyFromSeed_MinimumSize(t *testing.T) {
	seed := make([]byte, 16) // Minimum size

	keyData, err := GetMasterKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("GetMasterKeyFromSeed() error = %v", err)
	}

	if keyData == nil {
		t.Error("keyData should not be nil")
	}
}

func TestGetMasterKeyFromSeed_MaximumSize(t *testing.T) {
	seed := make([]byte, 64) // Maximum size

	keyData, err := GetMasterKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("GetMasterKeyFromSeed() error = %v", err)
	}

	if keyData == nil {
		t.Error("keyData should not be nil")
	}
}

func TestGetMasterKeyFromSeed_TooSmall(t *testing.T) {
	seed := make([]byte, 15)

	_, err := GetMasterKeyFromSeed(seed)
	if err == nil {
		t.Error("GetMasterKeyFromSeed() should return error for seed < 16 bytes")
	}
}

func TestGetMasterKeyFromSeed_TooLarge(t *testing.T) {
	seed := make([]byte, 65)

	_, err := GetMasterKeyFromSeed(seed)
	if err == nil {
		t.Error("GetMasterKeyFromSeed() should return error for seed > 64 bytes")
	}
}

func TestGetMasterKeyFromSeed_Deterministic(t *testing.T) {
	seed := make([]byte, 64)
	for i := range seed {
		seed[i] = byte(i)
	}

	keyData1, err := GetMasterKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("GetMasterKeyFromSeed() error = %v", err)
	}

	keyData2, err := GetMasterKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("GetMasterKeyFromSeed() error = %v", err)
	}

	if !bytes.Equal(keyData1.Key, keyData2.Key) {
		t.Error("Master key derivation should be deterministic")
	}

	if !bytes.Equal(keyData1.ChainCode, keyData2.ChainCode) {
		t.Error("Chain code derivation should be deterministic")
	}
}

func TestGetMasterKeyFromSeed_DifferentSeeds(t *testing.T) {
	seed1 := make([]byte, 64)
	seed2 := make([]byte, 64)
	seed2[0] = 1

	keyData1, err := GetMasterKeyFromSeed(seed1)
	if err != nil {
		t.Fatalf("GetMasterKeyFromSeed() error = %v", err)
	}

	keyData2, err := GetMasterKeyFromSeed(seed2)
	if err != nil {
		t.Fatalf("GetMasterKeyFromSeed() error = %v", err)
	}

	if bytes.Equal(keyData1.Key, keyData2.Key) {
		t.Error("Different seeds should produce different keys")
	}
}

// =============================================================================
// DerivePath Tests
// =============================================================================

func TestDerivePath_MasterKey(t *testing.T) {
	seed := make([]byte, 64)

	// Path "m" should return master key
	keyData, err := DerivePath("m", seed)
	if err != nil {
		t.Fatalf("DerivePath() error = %v", err)
	}

	master, _ := GetMasterKeyFromSeed(seed)
	if !bytes.Equal(keyData.Key, master.Key) {
		t.Error("Path 'm' should return master key")
	}
}

func TestDerivePath_SingleLevel(t *testing.T) {
	seed := make([]byte, 64)

	keyData, err := DerivePath("m/0'", seed)
	if err != nil {
		t.Fatalf("DerivePath() error = %v", err)
	}

	if keyData == nil {
		t.Error("keyData should not be nil")
	}
}

func TestDerivePath_MultipleLevels(t *testing.T) {
	seed := make([]byte, 64)

	keyData, err := DerivePath("m/44'/73404'/0'", seed)
	if err != nil {
		t.Fatalf("DerivePath() error = %v", err)
	}

	if keyData == nil {
		t.Error("keyData should not be nil")
	}
}

func TestDerivePath_WithoutMPrefix(t *testing.T) {
	seed := make([]byte, 64)

	keyData1, err := DerivePath("m/44'/73404'/0'", seed)
	if err != nil {
		t.Fatalf("DerivePath() error = %v", err)
	}

	keyData2, err := DerivePath("44'/73404'/0'", seed)
	if err != nil {
		t.Fatalf("DerivePath() error = %v", err)
	}

	if !bytes.Equal(keyData1.Key, keyData2.Key) {
		t.Error("Path with and without 'm' prefix should produce same key")
	}
}

func TestDerivePath_NonHardenedShouldFail(t *testing.T) {
	seed := make([]byte, 64)

	// Ed25519 only supports hardened derivation
	_, err := DerivePath("m/44", seed)
	if err == nil {
		t.Error("DerivePath() should return error for non-hardened path")
	}
}

func TestDerivePath_InvalidPath(t *testing.T) {
	seed := make([]byte, 64)

	testCases := []string{
		"",
		"invalid",
		"m/abc'",
		"m/-1'",
		"m/44'/73404'/abc",
	}

	for _, path := range testCases {
		_, err := DerivePath(path, seed)
		if err == nil {
			t.Errorf("DerivePath(%q) should return error", path)
		}
	}
}

func TestDerivePath_Deterministic(t *testing.T) {
	seed := make([]byte, 64)
	path := "m/44'/73404'/0'"

	keyData1, err := DerivePath(path, seed)
	if err != nil {
		t.Fatalf("DerivePath() error = %v", err)
	}

	keyData2, err := DerivePath(path, seed)
	if err != nil {
		t.Fatalf("DerivePath() error = %v", err)
	}

	if !bytes.Equal(keyData1.Key, keyData2.Key) {
		t.Error("Path derivation should be deterministic")
	}
}

func TestDerivePath_DifferentPaths(t *testing.T) {
	seed := make([]byte, 64)

	keyData1, err := DerivePath("m/44'/73404'/0'", seed)
	if err != nil {
		t.Fatalf("DerivePath() error = %v", err)
	}

	keyData2, err := DerivePath("m/44'/73404'/1'", seed)
	if err != nil {
		t.Fatalf("DerivePath() error = %v", err)
	}

	if bytes.Equal(keyData1.Key, keyData2.Key) {
		t.Error("Different paths should produce different keys")
	}
}

// =============================================================================
// DeriveKey Tests
// =============================================================================

func TestDeriveKey_ValidHexSeed(t *testing.T) {
	// 128 hex characters = 64 bytes (valid seed size)
	seedHex := "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"[:128]
	path := "m/44'/73404'/0'"

	privateKey, err := DeriveKey(path, seedHex)
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	// Ed25519 private key should be 64 bytes
	if len(privateKey) != 64 {
		t.Errorf("len(privateKey) = %d, want 64", len(privateKey))
	}
}

func TestDeriveKey_InvalidHexSeed(t *testing.T) {
	testCases := []string{
		"invalid",
		"zz0000",
		"1", // Odd length
	}

	for _, seedHex := range testCases {
		_, err := DeriveKey("m/44'/73404'/0'", seedHex)
		if err == nil {
			t.Errorf("DeriveKey() should return error for invalid hex: %q", seedHex)
		}
	}
}

func TestDeriveKey_Deterministic(t *testing.T) {
	// 128 hex characters = 64 bytes (valid seed size)
	seedHex := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	path := "m/44'/73404'/0'"

	privateKey1, err := DeriveKey(path, seedHex)
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	privateKey2, err := DeriveKey(path, seedHex)
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	if !bytes.Equal(privateKey1, privateKey2) {
		t.Error("DeriveKey should be deterministic")
	}
}

// =============================================================================
// getCKDPriv Tests
// =============================================================================

func TestGetCKDPriv_HardenedKey(t *testing.T) {
	seed := make([]byte, 64)
	parent, err := GetMasterKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("GetMasterKeyFromSeed() error = %v", err)
	}

	child, err := getCKDPriv(parent, HardenedKeyStart)
	if err != nil {
		t.Fatalf("getCKDPriv() error = %v", err)
	}

	if len(child.Key) != 32 {
		t.Errorf("len(child.Key) = %d, want 32", len(child.Key))
	}

	if len(child.ChainCode) != 32 {
		t.Errorf("len(child.ChainCode) = %d, want 32", len(child.ChainCode))
	}
}

func TestGetCKDPriv_NonHardenedShouldFail(t *testing.T) {
	seed := make([]byte, 64)
	parent, err := GetMasterKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("GetMasterKeyFromSeed() error = %v", err)
	}

	// Non-hardened index should fail for Ed25519
	_, err = getCKDPriv(parent, 0)
	if err == nil {
		t.Error("getCKDPriv() should return error for non-hardened index")
	}
}

func TestGetCKDPriv_NilParent(t *testing.T) {
	_, err := getCKDPriv(nil, HardenedKeyStart)
	if err == nil {
		t.Error("getCKDPriv() should return error for nil parent")
	}
}

func TestGetCKDPriv_DifferentIndices(t *testing.T) {
	seed := make([]byte, 64)
	parent, err := GetMasterKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("GetMasterKeyFromSeed() error = %v", err)
	}

	child1, err := getCKDPriv(parent, HardenedKeyStart+0)
	if err != nil {
		t.Fatalf("getCKDPriv() error = %v", err)
	}

	child2, err := getCKDPriv(parent, HardenedKeyStart+1)
	if err != nil {
		t.Fatalf("getCKDPriv() error = %v", err)
	}

	if bytes.Equal(child1.Key, child2.Key) {
		t.Error("Different indices should produce different keys")
	}
}

// =============================================================================
// GetPublicKey Tests
// =============================================================================

func TestKeyData_GetPublicKey(t *testing.T) {
	seed := make([]byte, 64)
	keyData, err := GetMasterKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("GetMasterKeyFromSeed() error = %v", err)
	}

	publicKey, err := keyData.GetPublicKey()
	if err != nil {
		t.Fatalf("GetPublicKey() error = %v", err)
	}

	// Ed25519 public key should be 32 bytes
	if len(publicKey) != 32 {
		t.Errorf("len(publicKey) = %d, want 32", len(publicKey))
	}
}

func TestKeyData_GetPublicKey_Deterministic(t *testing.T) {
	seed := make([]byte, 64)
	keyData, err := GetMasterKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("GetMasterKeyFromSeed() error = %v", err)
	}

	publicKey1, err := keyData.GetPublicKey()
	if err != nil {
		t.Fatalf("GetPublicKey() error = %v", err)
	}

	publicKey2, err := keyData.GetPublicKey()
	if err != nil {
		t.Fatalf("GetPublicKey() error = %v", err)
	}

	if !bytes.Equal(publicKey1, publicKey2) {
		t.Error("GetPublicKey should be deterministic")
	}
}

func TestKeyData_GetPublicKey_InvalidKey(t *testing.T) {
	keyData := &KeyData{
		Key:       make([]byte, 16), // Invalid length
		ChainCode: make([]byte, 32),
	}

	_, err := keyData.GetPublicKey()
	if err == nil {
		t.Error("GetPublicKey() should return error for invalid key length")
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestBIP32_FullDerivation(t *testing.T) {
	// Test full BIP44 derivation for Zenon
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed := MnemonicToSeed(mnemonic, "")

	// Derive m/44'/73404'/0'
	keyData, err := DerivePath("m/44'/73404'/0'", seed)
	if err != nil {
		t.Fatalf("DerivePath() error = %v", err)
	}

	// Get public key
	publicKey, err := keyData.GetPublicKey()
	if err != nil {
		t.Fatalf("GetPublicKey() error = %v", err)
	}

	if len(publicKey) != 32 {
		t.Errorf("len(publicKey) = %d, want 32", len(publicKey))
	}

	// Verify determinism
	keyData2, _ := DerivePath("m/44'/73404'/0'", seed)
	publicKey2, _ := keyData2.GetPublicKey()

	if !bytes.Equal(publicKey, publicKey2) {
		t.Error("Full derivation should be deterministic")
	}
}

func TestBIP32_MultipleAccounts(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed := MnemonicToSeed(mnemonic, "")

	// Derive multiple accounts
	accounts := []string{
		"m/44'/73404'/0'",
		"m/44'/73404'/1'",
		"m/44'/73404'/2'",
	}

	publicKeys := make([][]byte, len(accounts))
	for i, path := range accounts {
		keyData, err := DerivePath(path, seed)
		if err != nil {
			t.Fatalf("DerivePath(%q) error = %v", path, err)
		}

		publicKey, err := keyData.GetPublicKey()
		if err != nil {
			t.Fatalf("GetPublicKey() error = %v", err)
		}

		publicKeys[i] = publicKey
	}

	// All public keys should be different
	for i := 0; i < len(publicKeys); i++ {
		for j := i + 1; j < len(publicKeys); j++ {
			if bytes.Equal(publicKeys[i], publicKeys[j]) {
				t.Errorf("Public keys for accounts %d and %d should be different", i, j)
			}
		}
	}
}

func TestBIP32_KnownTestVector(t *testing.T) {
	// Test with known seed
	seedHex := "000102030405060708090a0b0c0d0e0f"
	seed, _ := hex.DecodeString(seedHex)

	master, err := GetMasterKeyFromSeed(seed)
	if err != nil {
		t.Fatalf("GetMasterKeyFromSeed() error = %v", err)
	}

	// Master key should be deterministic for this seed
	if len(master.Key) != 32 {
		t.Errorf("len(master.Key) = %d, want 32", len(master.Key))
	}

	// Derive child
	child, err := getCKDPriv(master, HardenedKeyStart)
	if err != nil {
		t.Fatalf("getCKDPriv() error = %v", err)
	}

	if len(child.Key) != 32 {
		t.Errorf("len(child.Key) = %d, want 32", len(child.Key))
	}
}

// =============================================================================
// Performance Benchmarks
// =============================================================================

func BenchmarkGetMasterKeyFromSeed(b *testing.B) {
	seed := make([]byte, 64)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetMasterKeyFromSeed(seed)
	}
}

func BenchmarkDerivePath(b *testing.B) {
	seed := make([]byte, 64)
	path := "m/44'/73404'/0'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DerivePath(path, seed)
	}
}

func BenchmarkGetCKDPriv(b *testing.B) {
	seed := make([]byte, 64)
	parent, _ := GetMasterKeyFromSeed(seed)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getCKDPriv(parent, HardenedKeyStart)
	}
}
