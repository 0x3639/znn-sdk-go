package wallet

import (
	"encoding/hex"
	"strings"
	"testing"
)

// =============================================================================
// GenerateMnemonic Tests
// =============================================================================

func TestGenerateMnemonic_128Bits(t *testing.T) {
	mnemonic, err := GenerateMnemonic(128)
	if err != nil {
		t.Fatalf("GenerateMnemonic(128) error = %v", err)
	}

	words := strings.Split(mnemonic, " ")
	if len(words) != 12 {
		t.Errorf("len(words) = %d, want 12", len(words))
	}

	if !ValidateMnemonicString(mnemonic) {
		t.Error("Generated mnemonic should be valid")
	}
}

func TestGenerateMnemonic_256Bits(t *testing.T) {
	mnemonic, err := GenerateMnemonic(256)
	if err != nil {
		t.Fatalf("GenerateMnemonic(256) error = %v", err)
	}

	words := strings.Split(mnemonic, " ")
	if len(words) != 24 {
		t.Errorf("len(words) = %d, want 24", len(words))
	}

	if !ValidateMnemonicString(mnemonic) {
		t.Error("Generated mnemonic should be valid")
	}
}

func TestGenerateMnemonic_InvalidStrength(t *testing.T) {
	testCases := []int{
		64,  // Too small
		100, // Not multiple of 32
		512, // Too large
	}

	for _, strength := range testCases {
		_, err := GenerateMnemonic(strength)
		if err == nil {
			t.Errorf("GenerateMnemonic(%d) should return error", strength)
		}
	}
}

func TestGenerateMnemonic_Unique(t *testing.T) {
	mnemonic1, err := GenerateMnemonic(128)
	if err != nil {
		t.Fatalf("GenerateMnemonic(128) error = %v", err)
	}

	mnemonic2, err := GenerateMnemonic(128)
	if err != nil {
		t.Fatalf("GenerateMnemonic(128) error = %v", err)
	}

	if mnemonic1 == mnemonic2 {
		t.Error("Generated mnemonics should be unique")
	}
}

// =============================================================================
// ValidateMnemonic Tests
// =============================================================================

func TestValidateMnemonic_Valid12Words(t *testing.T) {
	// Valid 12-word mnemonic
	words := []string{"abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "about"}

	if !ValidateMnemonic(words) {
		t.Error("ValidateMnemonic() should return true for valid mnemonic")
	}
}

func TestValidateMnemonic_Valid24Words(t *testing.T) {
	// Valid 24-word mnemonic
	words := []string{
		"abandon", "abandon", "abandon", "abandon", "abandon", "abandon",
		"abandon", "abandon", "abandon", "abandon", "abandon", "abandon",
		"abandon", "abandon", "abandon", "abandon", "abandon", "abandon",
		"abandon", "abandon", "abandon", "abandon", "abandon", "art",
	}

	if !ValidateMnemonic(words) {
		t.Error("ValidateMnemonic() should return true for valid mnemonic")
	}
}

func TestValidateMnemonic_InvalidChecksum(t *testing.T) {
	// Invalid checksum (last word wrong)
	words := []string{"abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon"}

	if ValidateMnemonic(words) {
		t.Error("ValidateMnemonic() should return false for invalid checksum")
	}
}

func TestValidateMnemonic_WrongWordCount(t *testing.T) {
	testCases := [][]string{
		{"abandon", "abandon", "abandon"},                       // Too few
		{"abandon", "abandon", "abandon", "abandon", "abandon"}, // Wrong count
		strings.Split(strings.Repeat("abandon ", 50), " ")[:50], // Too many
	}

	for _, words := range testCases {
		if ValidateMnemonic(words) {
			t.Errorf("ValidateMnemonic() should return false for %d words", len(words))
		}
	}
}

func TestValidateMnemonic_InvalidWord(t *testing.T) {
	words := []string{"invalid", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "about"}

	if ValidateMnemonic(words) {
		t.Error("ValidateMnemonic() should return false for invalid word")
	}
}

func TestValidateMnemonic_Empty(t *testing.T) {
	if ValidateMnemonic([]string{}) {
		t.Error("ValidateMnemonic() should return false for empty slice")
	}
}

// =============================================================================
// ValidateMnemonicString Tests
// =============================================================================

func TestValidateMnemonicString_Valid(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	if !ValidateMnemonicString(mnemonic) {
		t.Error("ValidateMnemonicString() should return true for valid mnemonic")
	}
}

func TestValidateMnemonicString_Invalid(t *testing.T) {
	testCases := []string{
		"invalid words here",
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon", // Wrong checksum
		"",
		"abandon", // Too short
	}

	for _, mnemonic := range testCases {
		if ValidateMnemonicString(mnemonic) {
			t.Errorf("ValidateMnemonicString(%q) should return false", mnemonic)
		}
	}
}

// =============================================================================
// IsValidWord Tests
// =============================================================================

func TestIsValidWord_Valid(t *testing.T) {
	validWords := []string{
		"abandon",
		"ability",
		"able",
		"about",
		"above",
		"absent",
		"zoo",
	}

	for _, word := range validWords {
		if !IsValidWord(word) {
			t.Errorf("IsValidWord(%q) should return true", word)
		}
	}
}

func TestIsValidWord_Invalid(t *testing.T) {
	invalidWords := []string{
		"invalid",
		"notaword",
		"",
		"Abandon", // Capital letters
		"12345",
	}

	for _, word := range invalidWords {
		if IsValidWord(word) {
			t.Errorf("IsValidWord(%q) should return false", word)
		}
	}
}

// =============================================================================
// MnemonicToEntropy Tests
// =============================================================================

func TestMnemonicToEntropy_Valid(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	entropy, err := MnemonicToEntropy(mnemonic)
	if err != nil {
		t.Fatalf("MnemonicToEntropy() error = %v", err)
	}

	// 12 words = 128 bits = 16 bytes
	if len(entropy) != 16 {
		t.Errorf("len(entropy) = %d, want 16", len(entropy))
	}

	// Known entropy for this mnemonic
	expectedHex := "00000000000000000000000000000000"
	if hex.EncodeToString(entropy) != expectedHex {
		t.Errorf("entropy = %x, want %s", entropy, expectedHex)
	}
}

func TestMnemonicToEntropy_24Words(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"

	entropy, err := MnemonicToEntropy(mnemonic)
	if err != nil {
		t.Fatalf("MnemonicToEntropy() error = %v", err)
	}

	// 24 words = 256 bits = 32 bytes
	if len(entropy) != 32 {
		t.Errorf("len(entropy) = %d, want 32", len(entropy))
	}
}

func TestMnemonicToEntropy_Invalid(t *testing.T) {
	invalidMnemonics := []string{
		"invalid words here",
		"abandon abandon abandon", // Too short
		"",
	}

	for _, mnemonic := range invalidMnemonics {
		_, err := MnemonicToEntropy(mnemonic)
		if err == nil {
			t.Errorf("MnemonicToEntropy(%q) should return error", mnemonic)
		}
	}
}

// =============================================================================
// EntropyToMnemonic Tests
// =============================================================================

func TestEntropyToMnemonic_16Bytes(t *testing.T) {
	entropy := make([]byte, 16) // All zeros

	mnemonic, err := EntropyToMnemonic(entropy)
	if err != nil {
		t.Fatalf("EntropyToMnemonic() error = %v", err)
	}

	expected := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	if mnemonic != expected {
		t.Errorf("mnemonic = %q, want %q", mnemonic, expected)
	}
}

func TestEntropyToMnemonic_32Bytes(t *testing.T) {
	entropy := make([]byte, 32) // All zeros

	mnemonic, err := EntropyToMnemonic(entropy)
	if err != nil {
		t.Fatalf("EntropyToMnemonic() error = %v", err)
	}

	words := strings.Split(mnemonic, " ")
	if len(words) != 24 {
		t.Errorf("len(words) = %d, want 24", len(words))
	}
}

func TestEntropyToMnemonic_InvalidLength(t *testing.T) {
	testCases := []int{
		8,  // Too short
		15, // Not valid length
		17, // Not valid length
		33, // Too long
	}

	for _, length := range testCases {
		entropy := make([]byte, length)
		_, err := EntropyToMnemonic(entropy)
		if err == nil {
			t.Errorf("EntropyToMnemonic() with %d bytes should return error", length)
		}
	}
}

// =============================================================================
// MnemonicToSeed Tests
// =============================================================================

func TestMnemonicToSeed_NoPassphrase(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	seed := MnemonicToSeed(mnemonic, "")

	if len(seed) != 64 {
		t.Errorf("len(seed) = %d, want 64", len(seed))
	}

	// Known seed for this mnemonic with empty passphrase
	expectedHex := "5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d48b2d2ce9e38e4"
	actualHex := hex.EncodeToString(seed)
	if actualHex != expectedHex {
		t.Errorf("seed = %s, want %s", actualHex, expectedHex)
	}
}

func TestMnemonicToSeed_WithPassphrase(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	passphrase := "TREZOR"

	seed := MnemonicToSeed(mnemonic, passphrase)

	if len(seed) != 64 {
		t.Errorf("len(seed) = %d, want 64", len(seed))
	}

	// Seed should be different from no passphrase
	seedNoPassphrase := MnemonicToSeed(mnemonic, "")
	if hex.EncodeToString(seed) == hex.EncodeToString(seedNoPassphrase) {
		t.Error("Seed with passphrase should differ from seed without passphrase")
	}
}

func TestMnemonicToSeed_DifferentMnemonics(t *testing.T) {
	mnemonic1 := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	mnemonic2 := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"

	seed1 := MnemonicToSeed(mnemonic1, "")
	seed2 := MnemonicToSeed(mnemonic2, "")

	if hex.EncodeToString(seed1) == hex.EncodeToString(seed2) {
		t.Error("Different mnemonics should produce different seeds")
	}
}

func TestMnemonicToSeed_Deterministic(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	seed1 := MnemonicToSeed(mnemonic, "")
	seed2 := MnemonicToSeed(mnemonic, "")

	if hex.EncodeToString(seed1) != hex.EncodeToString(seed2) {
		t.Error("MnemonicToSeed should be deterministic")
	}
}

// =============================================================================
// Round Trip Tests
// =============================================================================

func TestMnemonicRoundTrip(t *testing.T) {
	// Generate -> Entropy -> Mnemonic -> Seed
	mnemonic1, err := GenerateMnemonic(128)
	if err != nil {
		t.Fatalf("GenerateMnemonic() error = %v", err)
	}

	// Convert to entropy
	entropy, err := MnemonicToEntropy(mnemonic1)
	if err != nil {
		t.Fatalf("MnemonicToEntropy() error = %v", err)
	}

	// Convert back to mnemonic
	mnemonic2, err := EntropyToMnemonic(entropy)
	if err != nil {
		t.Fatalf("EntropyToMnemonic() error = %v", err)
	}

	if mnemonic1 != mnemonic2 {
		t.Error("Round trip failed: mnemonics don't match")
	}

	// Should be valid
	if !ValidateMnemonicString(mnemonic2) {
		t.Error("Round trip mnemonic should be valid")
	}
}

// =============================================================================
// Performance Benchmarks
// =============================================================================

func BenchmarkGenerateMnemonic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateMnemonic(128)
	}
}

func BenchmarkValidateMnemonic(b *testing.B) {
	words := []string{"abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "abandon", "about"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateMnemonic(words)
	}
}

func BenchmarkMnemonicToSeed(b *testing.B) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MnemonicToSeed(mnemonic, "")
	}
}
