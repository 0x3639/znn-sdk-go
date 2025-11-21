package wallet

import (
	"errors"
	"testing"
)

// =============================================================================
// NewKeyStoreFromMnemonic Tests
// =============================================================================

func TestNewKeyStoreFromMnemonic_Valid(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	ks, err := NewKeyStoreFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("NewKeyStoreFromMnemonic() error = %v", err)
	}

	if ks.Mnemonic != mnemonic {
		t.Error("Mnemonic not stored correctly")
	}

	if len(ks.Entropy) != 16 {
		t.Errorf("len(Entropy) = %d, want 16", len(ks.Entropy))
	}

	if len(ks.Seed) != 64 {
		t.Errorf("len(Seed) = %d, want 64", len(ks.Seed))
	}
}

func TestNewKeyStoreFromMnemonic_Invalid(t *testing.T) {
	testCases := []string{
		"invalid mnemonic words",
		"abandon abandon abandon",
		"",
	}

	for _, mnemonic := range testCases {
		_, err := NewKeyStoreFromMnemonic(mnemonic)
		if err == nil {
			t.Errorf("NewKeyStoreFromMnemonic(%q) should return error", mnemonic)
		}
	}
}

// =============================================================================
// NewKeyStoreFromSeed Tests
// =============================================================================

func TestNewKeyStoreFromSeed_Valid(t *testing.T) {
	seedHex := "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"[:128]

	ks, err := NewKeyStoreFromSeed(seedHex)
	if err != nil {
		t.Fatalf("NewKeyStoreFromSeed() error = %v", err)
	}

	if len(ks.Seed) != 64 {
		t.Errorf("len(Seed) = %d, want 64", len(ks.Seed))
	}
}

func TestNewKeyStoreFromSeed_Invalid(t *testing.T) {
	testCases := []string{
		"invalid",
		"zz00",
	}

	for _, seedHex := range testCases {
		_, err := NewKeyStoreFromSeed(seedHex)
		if err == nil {
			t.Errorf("NewKeyStoreFromSeed(%q) should return error", seedHex)
		}
	}
}

// =============================================================================
// NewKeyStoreFromEntropy Tests
// =============================================================================

func TestNewKeyStoreFromEntropy_16Bytes(t *testing.T) {
	entropy := make([]byte, 16)

	ks, err := NewKeyStoreFromEntropy(entropy)
	if err != nil {
		t.Fatalf("NewKeyStoreFromEntropy() error = %v", err)
	}

	if len(ks.Entropy) != 16 {
		t.Errorf("len(Entropy) = %d, want 16", len(ks.Entropy))
	}

	if ks.Mnemonic == "" {
		t.Error("Mnemonic should be generated from entropy")
	}
}

func TestNewKeyStoreFromEntropy_32Bytes(t *testing.T) {
	entropy := make([]byte, 32)

	ks, err := NewKeyStoreFromEntropy(entropy)
	if err != nil {
		t.Fatalf("NewKeyStoreFromEntropy() error = %v", err)
	}

	if len(ks.Entropy) != 32 {
		t.Errorf("len(Entropy) = %d, want 32", len(ks.Entropy))
	}
}

func TestNewKeyStoreFromEntropy_InvalidSize(t *testing.T) {
	testCases := []int{8, 15, 17, 31, 33, 64}

	for _, size := range testCases {
		entropy := make([]byte, size)
		_, err := NewKeyStoreFromEntropy(entropy)
		if err == nil {
			t.Errorf("NewKeyStoreFromEntropy() with %d bytes should return error", size)
		}
	}
}

// =============================================================================
// NewKeyStoreRandom Tests
// =============================================================================

func TestNewKeyStoreRandom(t *testing.T) {
	ks, err := NewKeyStoreRandom()
	if err != nil {
		t.Fatalf("NewKeyStoreRandom() error = %v", err)
	}

	if ks.Mnemonic == "" {
		t.Error("Mnemonic should not be empty")
	}

	if len(ks.Entropy) != 32 {
		t.Errorf("len(Entropy) = %d, want 32 (256-bit)", len(ks.Entropy))
	}

	if len(ks.Seed) != 64 {
		t.Errorf("len(Seed) = %d, want 64", len(ks.Seed))
	}
}

func TestNewKeyStoreRandom_Unique(t *testing.T) {
	ks1, err := NewKeyStoreRandom()
	if err != nil {
		t.Fatalf("NewKeyStoreRandom() error = %v", err)
	}

	ks2, err := NewKeyStoreRandom()
	if err != nil {
		t.Fatalf("NewKeyStoreRandom() error = %v", err)
	}

	if ks1.Mnemonic == ks2.Mnemonic {
		t.Error("Random keystores should have different mnemonics")
	}
}

// =============================================================================
// GetKeyPair Tests
// =============================================================================

func TestGetKeyPair_Account0(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)

	kp, err := ks.GetKeyPair(0)
	if err != nil {
		t.Fatalf("GetKeyPair(0) error = %v", err)
	}

	if kp == nil {
		t.Error("KeyPair should not be nil")
	}

	// Should be able to get address
	addr, err := kp.GetAddress()
	if err != nil {
		t.Fatalf("GetAddress() error = %v", err)
	}

	if addr == nil {
		t.Error("Address should not be nil")
	}
}

func TestGetKeyPair_MultipleAccounts(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)

	// Derive multiple accounts
	accounts := []int{0, 1, 2, 5, 10}
	addresses := make([]string, len(accounts))

	for i, account := range accounts {
		kp, err := ks.GetKeyPair(account)
		if err != nil {
			t.Fatalf("GetKeyPair(%d) error = %v", account, err)
		}

		addr, err := kp.GetAddress()
		if err != nil {
			t.Fatalf("GetAddress() error = %v", err)
		}

		addresses[i] = addr.String()
	}

	// All addresses should be different
	for i := 0; i < len(addresses); i++ {
		for j := i + 1; j < len(addresses); j++ {
			if addresses[i] == addresses[j] {
				t.Errorf("Accounts %d and %d have same address", accounts[i], accounts[j])
			}
		}
	}
}

func TestGetKeyPair_Deterministic(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	ks1, _ := NewKeyStoreFromMnemonic(mnemonic)
	ks2, _ := NewKeyStoreFromMnemonic(mnemonic)

	kp1, _ := ks1.GetKeyPair(0)
	kp2, _ := ks2.GetKeyPair(0)

	addr1, _ := kp1.GetAddress()
	addr2, _ := kp2.GetAddress()

	if addr1.String() != addr2.String() {
		t.Error("Same mnemonic should produce same addresses")
	}
}

// =============================================================================
// DeriveAddressesByRange Tests
// =============================================================================

func TestDeriveAddressesByRange_ValidRange(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)

	addresses, err := ks.DeriveAddressesByRange(0, 5)
	if err != nil {
		t.Fatalf("DeriveAddressesByRange() error = %v", err)
	}

	if len(addresses) != 5 {
		t.Errorf("len(addresses) = %d, want 5", len(addresses))
	}

	// All should be unique
	for i := 0; i < len(addresses); i++ {
		for j := i + 1; j < len(addresses); j++ {
			if addresses[i].String() == addresses[j].String() {
				t.Errorf("Addresses %d and %d are the same", i, j)
			}
		}
	}
}

func TestDeriveAddressesByRange_EmptyRange(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)

	addresses, err := ks.DeriveAddressesByRange(5, 5)
	if err != nil {
		t.Fatalf("DeriveAddressesByRange() error = %v", err)
	}

	if len(addresses) != 0 {
		t.Errorf("len(addresses) = %d, want 0", len(addresses))
	}
}

func TestDeriveAddressesByRange_InvalidRange(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)

	testCases := []struct {
		left, right int
	}{
		{-1, 5},
		{5, 3},
	}

	for _, tc := range testCases {
		_, err := ks.DeriveAddressesByRange(tc.left, tc.right)
		if err == nil {
			t.Errorf("DeriveAddressesByRange(%d, %d) should return error", tc.left, tc.right)
		}
	}
}

// =============================================================================
// FindAddress Tests
// =============================================================================

func TestFindAddress_Found(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)

	// Get address at index 3
	kp, _ := ks.GetKeyPair(3)
	targetAddr, _ := kp.GetAddress()

	// Find it
	result, err := ks.FindAddress(*targetAddr, 10)
	if err != nil {
		t.Fatalf("FindAddress() error = %v", err)
	}

	if result == nil {
		t.Fatal("FindResponse should not be nil")
	}

	if result.Index != 3 {
		t.Errorf("Index = %d, want 3", result.Index)
	}

	if result.KeyPair == nil {
		t.Error("KeyPair should not be nil")
	}
}

func TestFindAddress_NotFound(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)

	// Get address from a different keystore
	otherKs, _ := NewKeyStoreRandom()
	otherKp, _ := otherKs.GetKeyPair(0)
	otherAddr, _ := otherKp.GetAddress()

	// Try to find it (should not be found)
	_, err := ks.FindAddress(*otherAddr, 10)
	if !errors.Is(err, ErrAddressNotFound) {
		t.Errorf("FindAddress() error = %v, want ErrAddressNotFound", err)
	}
}

func TestFindAddress_DefaultMaxIndex(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)

	kp, _ := ks.GetKeyPair(0)
	addr, _ := kp.GetAddress()

	// Should use DefaultMaxIndex when maxAccounts <= 0
	result, err := ks.FindAddress(*addr, 0)
	if err != nil {
		t.Fatalf("FindAddress() error = %v", err)
	}

	if result.Index != 0 {
		t.Errorf("Index = %d, want 0", result.Index)
	}
}

// =============================================================================
// GetBaseAddress Tests
// =============================================================================

func TestGetBaseAddress(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)

	baseAddr, err := ks.GetBaseAddress()
	if err != nil {
		t.Fatalf("GetBaseAddress() error = %v", err)
	}

	if baseAddr == nil {
		t.Fatal("Base address should not be nil")
	}

	// Should match account 0
	kp0, _ := ks.GetKeyPair(0)
	addr0, _ := kp0.GetAddress()

	if baseAddr.String() != addr0.String() {
		t.Error("Base address should match account 0")
	}
}

// =============================================================================
// ToEncryptedFile / FromEncryptedFile Tests
// =============================================================================

func TestToEncryptedFile_Basic(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)

	password := "password123"
	ef, err := ks.ToEncryptedFile(password, nil)
	if err != nil {
		t.Fatalf("ToEncryptedFile() error = %v", err)
	}

	if ef == nil {
		t.Fatal("EncryptedFile should not be nil")
	}

	if ef.Crypto == nil {
		t.Error("Crypto params should not be nil")
	}
}

func TestToEncryptedFile_WithMetadata(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)

	metadata := map[string]interface{}{
		"name": "my-wallet",
	}

	password := "password123"
	ef, err := ks.ToEncryptedFile(password, metadata)
	if err != nil {
		t.Fatalf("ToEncryptedFile() error = %v", err)
	}

	if ef.Metadata["name"] != "my-wallet" {
		t.Error("Metadata not preserved")
	}

	// Should auto-add base address
	if ef.Metadata[BaseAddressKey] == nil {
		t.Error("Base address should be auto-added to metadata")
	}

	// Should auto-add wallet type
	if ef.Metadata[WalletTypeKey] != KeyStoreWalletType {
		t.Error("Wallet type should be auto-added to metadata")
	}
}

func TestFromEncryptedFile_Mnemonic(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks1, _ := NewKeyStoreFromMnemonic(mnemonic)

	password := "password123"
	ef, _ := ks1.ToEncryptedFile(password, nil)

	ks2, err := FromEncryptedFile(ef, password)
	if err != nil {
		t.Fatalf("FromEncryptedFile() error = %v", err)
	}

	if ks2.Mnemonic != mnemonic {
		t.Error("Mnemonic not restored correctly")
	}

	// Should derive same addresses
	addr1, _ := ks1.GetBaseAddress()
	addr2, _ := ks2.GetBaseAddress()

	if addr1.String() != addr2.String() {
		t.Error("Restored keystore should produce same addresses")
	}
}

func TestFromEncryptedFile_WrongPassword(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)

	ef, _ := ks.ToEncryptedFile("password123", nil)

	_, err := FromEncryptedFile(ef, "wrongpassword")
	if !errors.Is(err, ErrIncorrectPassword) {
		t.Errorf("FromEncryptedFile() error = %v, want ErrIncorrectPassword", err)
	}
}

// =============================================================================
// Round Trip Tests
// =============================================================================

func TestKeyStore_FullRoundTrip(t *testing.T) {
	// Create from mnemonic
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks1, err := NewKeyStoreFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("NewKeyStoreFromMnemonic() error = %v", err)
	}

	// Derive some addresses
	addresses1, err := ks1.DeriveAddressesByRange(0, 5)
	if err != nil {
		t.Fatalf("DeriveAddressesByRange() error = %v", err)
	}

	// Encrypt to file
	password := "password123"
	metadata := map[string]interface{}{
		"name": "test-wallet",
	}

	ef, err := ks1.ToEncryptedFile(password, metadata)
	if err != nil {
		t.Fatalf("ToEncryptedFile() error = %v", err)
	}

	// Serialize to JSON
	jsonData, err := ef.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Deserialize from JSON
	ef2, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	// Decrypt to keystore
	ks2, err := FromEncryptedFile(ef2, password)
	if err != nil {
		t.Fatalf("FromEncryptedFile() error = %v", err)
	}

	// Derive same addresses
	addresses2, err := ks2.DeriveAddressesByRange(0, 5)
	if err != nil {
		t.Fatalf("DeriveAddressesByRange() error = %v", err)
	}

	// Verify addresses match
	if len(addresses1) != len(addresses2) {
		t.Fatalf("Address count mismatch: %d != %d", len(addresses1), len(addresses2))
	}

	for i := range addresses1 {
		if addresses1[i].String() != addresses2[i].String() {
			t.Errorf("Address %d mismatch", i)
		}
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestSerializeKeyStoreData(t *testing.T) {
	data := map[string]interface{}{
		"mnemonic": "test mnemonic",
		"entropy":  "0123456789abcdef",
	}

	result, err := serializeKeyStoreData(data)
	if err != nil {
		t.Fatalf("serializeKeyStoreData() error = %v", err)
	}

	if len(result) == 0 {
		t.Error("Serialized data should not be empty")
	}
}

func TestDeserializeKeyStoreData(t *testing.T) {
	jsonStr := `{"mnemonic":"test mnemonic","entropy":"0123456789abcdef"}`

	result, err := deserializeKeyStoreData([]byte(jsonStr))
	if err != nil {
		t.Fatalf("deserializeKeyStoreData() error = %v", err)
	}

	if result["mnemonic"] != "test mnemonic" {
		t.Error("Mnemonic not deserialized correctly")
	}

	if result["entropy"] != "0123456789abcdef" {
		t.Error("Entropy not deserialized correctly")
	}
}

// =============================================================================
// Performance Benchmarks
// =============================================================================

func BenchmarkGetKeyPair(b *testing.B) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ks.GetKeyPair(0)
	}
}

func BenchmarkDeriveAddressesByRange(b *testing.B) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ks.DeriveAddressesByRange(0, 10)
	}
}

func BenchmarkToEncryptedFile(b *testing.B) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)
	password := "password123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ks.ToEncryptedFile(password, nil)
	}
}

func BenchmarkFromEncryptedFile(b *testing.B) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	ks, _ := NewKeyStoreFromMnemonic(mnemonic)
	password := "password123"
	ef, _ := ks.ToEncryptedFile(password, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FromEncryptedFile(ef, password)
	}
}
