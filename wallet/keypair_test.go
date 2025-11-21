package wallet

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

// =============================================================================
// NewKeyPair Tests
// =============================================================================

func TestNewKeyPair(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	kp := NewKeyPair(privateKey)

	if kp == nil {
		t.Fatal("NewKeyPair() should not return nil")
	}

	if !bytes.Equal(kp.privateKey, privateKey) {
		t.Error("Private key not stored correctly")
	}
}

func TestNewKeyPair_LazyDerivation(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	kp := NewKeyPair(privateKey)

	// Initially, public key and address should be nil (lazy)
	if kp.publicKey != nil {
		t.Error("Public key should be nil initially")
	}

	if kp.address != nil {
		t.Error("Address should be nil initially")
	}
}

// =============================================================================
// NewKeyPairFromSeed Tests
// =============================================================================

func TestNewKeyPairFromSeed_ValidSeed(t *testing.T) {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}

	kp, err := NewKeyPairFromSeed(seed)
	if err != nil {
		t.Fatalf("NewKeyPairFromSeed() error = %v", err)
	}

	if kp == nil {
		t.Fatal("KeyPair should not be nil")
	}

	// Private key should be 64 bytes (Ed25519 format)
	if len(kp.privateKey) != 64 {
		t.Errorf("len(privateKey) = %d, want 64", len(kp.privateKey))
	}
}

func TestNewKeyPairFromSeed_InvalidSeed(t *testing.T) {
	testCases := []int{16, 31, 33, 64}

	for _, size := range testCases {
		seed := make([]byte, size)
		_, err := NewKeyPairFromSeed(seed)
		if err == nil {
			t.Errorf("NewKeyPairFromSeed() with %d bytes should return error", size)
		}
	}
}

func TestNewKeyPairFromSeed_Deterministic(t *testing.T) {
	seed := make([]byte, 32)

	kp1, err := NewKeyPairFromSeed(seed)
	if err != nil {
		t.Fatalf("NewKeyPairFromSeed() error = %v", err)
	}

	kp2, err := NewKeyPairFromSeed(seed)
	if err != nil {
		t.Fatalf("NewKeyPairFromSeed() error = %v", err)
	}

	if !bytes.Equal(kp1.privateKey, kp2.privateKey) {
		t.Error("Same seed should produce same private key")
	}
}

// =============================================================================
// GetPrivateKey Tests
// =============================================================================

func TestGetPrivateKey(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	kp := NewKeyPair(privateKey)
	retrieved := kp.GetPrivateKey()

	if !bytes.Equal(retrieved, privateKey) {
		t.Error("GetPrivateKey() should return the correct private key")
	}
}

// =============================================================================
// GetPublicKey Tests
// =============================================================================

func TestGetPublicKey_DerivesCorrectly(t *testing.T) {
	pubKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	kp := NewKeyPair(privateKey)
	derived, err := kp.GetPublicKey()
	if err != nil {
		t.Fatalf("GetPublicKey() error = %v", err)
	}

	if !bytes.Equal(derived, pubKey) {
		t.Error("GetPublicKey() should derive correct public key")
	}
}

func TestGetPublicKey_Caches(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	kp := NewKeyPair(privateKey)

	pubKey1, err := kp.GetPublicKey()
	if err != nil {
		t.Fatalf("GetPublicKey() error = %v", err)
	}

	// Second call should return cached value
	pubKey2, err := kp.GetPublicKey()
	if err != nil {
		t.Fatalf("GetPublicKey() error = %v", err)
	}

	if !bytes.Equal(pubKey1, pubKey2) {
		t.Error("GetPublicKey() should return cached value")
	}

	// Check that it's actually cached
	if kp.publicKey == nil {
		t.Error("Public key should be cached after first call")
	}
}

func TestGetPublicKey_InvalidPrivateKey(t *testing.T) {
	// Invalid private key (wrong size)
	kp := &KeyPair{
		privateKey: make([]byte, 32),
	}

	_, err := kp.GetPublicKey()
	if err == nil {
		t.Error("GetPublicKey() should return error for invalid private key")
	}
}

// =============================================================================
// GetAddress Tests
// =============================================================================

func TestGetAddress_DerivesCorrectly(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	kp := NewKeyPair(privateKey)
	addr, err := kp.GetAddress()
	if err != nil {
		t.Fatalf("GetAddress() error = %v", err)
	}

	if addr == nil {
		t.Fatal("Address should not be nil")
	}

	// Address should be valid Zenon address format
	addrStr := addr.String()
	if len(addrStr) == 0 {
		t.Error("Address string should not be empty")
	}
}

func TestGetAddress_Caches(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	kp := NewKeyPair(privateKey)

	addr1, err := kp.GetAddress()
	if err != nil {
		t.Fatalf("GetAddress() error = %v", err)
	}

	addr2, err := kp.GetAddress()
	if err != nil {
		t.Fatalf("GetAddress() error = %v", err)
	}

	if addr1.String() != addr2.String() {
		t.Error("GetAddress() should return cached value")
	}

	// Check that it's actually cached
	if kp.address == nil {
		t.Error("Address should be cached after first call")
	}
}

func TestGetAddress_Deterministic(t *testing.T) {
	seed := make([]byte, 32)

	kp1, _ := NewKeyPairFromSeed(seed)
	addr1, err := kp1.GetAddress()
	if err != nil {
		t.Fatalf("GetAddress() error = %v", err)
	}

	kp2, _ := NewKeyPairFromSeed(seed)
	addr2, err := kp2.GetAddress()
	if err != nil {
		t.Fatalf("GetAddress() error = %v", err)
	}

	if addr1.String() != addr2.String() {
		t.Error("Same seed should produce same address")
	}
}

// =============================================================================
// Sign Tests
// =============================================================================

func TestSign_ValidSignature(t *testing.T) {
	pubKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	kp := NewKeyPair(privateKey)
	message := []byte("test message")

	signature, err := kp.Sign(message)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	// Verify using standard library
	valid := ed25519.Verify(pubKey, message, signature)
	if !valid {
		t.Error("Signature should be valid")
	}
}

func TestSign_DifferentMessages(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	kp := NewKeyPair(privateKey)

	sig1, err := kp.Sign([]byte("message1"))
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	sig2, err := kp.Sign([]byte("message2"))
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if bytes.Equal(sig1, sig2) {
		t.Error("Different messages should produce different signatures")
	}
}

func TestSign_Deterministic(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	kp := NewKeyPair(privateKey)
	message := []byte("test message")

	sig1, err := kp.Sign(message)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	sig2, err := kp.Sign(message)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if !bytes.Equal(sig1, sig2) {
		t.Error("Same message should produce same signature")
	}
}

// =============================================================================
// Verify Tests
// =============================================================================

func TestVerify_ValidSignature(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	kp := NewKeyPair(privateKey)
	message := []byte("test message")

	signature, err := kp.Sign(message)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	valid, err := kp.Verify(signature, message)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if !valid {
		t.Error("Valid signature should verify successfully")
	}
}

func TestVerify_InvalidSignature(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	kp := NewKeyPair(privateKey)
	message := []byte("test message")

	signature, err := kp.Sign(message)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	// Tamper with signature
	signature[0] ^= 0xFF

	valid, err := kp.Verify(signature, message)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if valid {
		t.Error("Tampered signature should not verify")
	}
}

func TestVerify_WrongMessage(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	kp := NewKeyPair(privateKey)
	message := []byte("test message")

	signature, err := kp.Sign(message)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	// Verify with different message
	wrongMessage := []byte("wrong message")
	valid, err := kp.Verify(signature, wrongMessage)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if valid {
		t.Error("Signature should not verify with wrong message")
	}
}

// =============================================================================
// GeneratePublicKey Static Method Tests
// =============================================================================

func TestGeneratePublicKey_Valid(t *testing.T) {
	pubKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	derived, err := GeneratePublicKey(privateKey)
	if err != nil {
		t.Fatalf("GeneratePublicKey() error = %v", err)
	}

	if !bytes.Equal(derived, pubKey) {
		t.Error("GeneratePublicKey() should derive correct public key")
	}
}

func TestGeneratePublicKey_Invalid(t *testing.T) {
	// Invalid private key
	_, err := GeneratePublicKey(make([]byte, 32))
	if err == nil {
		t.Error("GeneratePublicKey() should return error for invalid key")
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestKeyPair_FullWorkflow(t *testing.T) {
	// Create from seed
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}

	kp, err := NewKeyPairFromSeed(seed)
	if err != nil {
		t.Fatalf("NewKeyPairFromSeed() error = %v", err)
	}

	// Get public key
	pubKey, err := kp.GetPublicKey()
	if err != nil {
		t.Fatalf("GetPublicKey() error = %v", err)
	}

	if len(pubKey) != 32 {
		t.Errorf("len(pubKey) = %d, want 32", len(pubKey))
	}

	// Get address
	addr, err := kp.GetAddress()
	if err != nil {
		t.Fatalf("GetAddress() error = %v", err)
	}

	if addr == nil {
		t.Error("Address should not be nil")
	}

	// Sign message
	message := []byte("test message")
	signature, err := kp.Sign(message)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if len(signature) != 64 {
		t.Errorf("len(signature) = %d, want 64", len(signature))
	}

	// Verify signature
	valid, err := kp.Verify(signature, message)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if !valid {
		t.Error("Signature should be valid")
	}
}

func TestKeyPair_WithBIP32Derivation(t *testing.T) {
	// Full integration: Mnemonic -> Seed -> BIP32 -> KeyPair
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed := MnemonicToSeed(mnemonic, "")

	// Derive key using BIP32
	keyData, err := DerivePath("m/44'/73404'/0'", seed)
	if err != nil {
		t.Fatalf("DerivePath() error = %v", err)
	}

	// Create KeyPair from derived key
	kp, err := NewKeyPairFromSeed(keyData.Key)
	if err != nil {
		t.Fatalf("NewKeyPairFromSeed() error = %v", err)
	}

	// Get address
	addr, err := kp.GetAddress()
	if err != nil {
		t.Fatalf("GetAddress() error = %v", err)
	}

	// Address should be deterministic for this mnemonic
	if addr == nil {
		t.Error("Address should not be nil")
	}

	// Derive second account
	keyData2, err := DerivePath("m/44'/73404'/1'", seed)
	if err != nil {
		t.Fatalf("DerivePath() error = %v", err)
	}

	kp2, err := NewKeyPairFromSeed(keyData2.Key)
	if err != nil {
		t.Fatalf("NewKeyPairFromSeed() error = %v", err)
	}

	addr2, err := kp2.GetAddress()
	if err != nil {
		t.Fatalf("GetAddress() error = %v", err)
	}

	// Different accounts should have different addresses
	if addr.String() == addr2.String() {
		t.Error("Different accounts should have different addresses")
	}
}

// =============================================================================
// Performance Benchmarks
// =============================================================================

func BenchmarkNewKeyPair(b *testing.B) {
	_, privateKey, _ := ed25519.GenerateKey(rand.Reader)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewKeyPair(privateKey)
	}
}

func BenchmarkGetPublicKey(b *testing.B) {
	_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	kp := NewKeyPair(privateKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kp.GetPublicKey()
	}
}

func BenchmarkSign(b *testing.B) {
	_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	kp := NewKeyPair(privateKey)
	message := []byte("test message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kp.Sign(message)
	}
}

func BenchmarkVerify(b *testing.B) {
	_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	kp := NewKeyPair(privateKey)
	message := []byte("test message")
	signature, _ := kp.Sign(message)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kp.Verify(signature, message)
	}
}

// =============================================================================
// Destroy Tests (Security)
// =============================================================================

func TestDestroy(t *testing.T) {
	// Generate a keypair
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Create copy of original private key for comparison
	originalPrivateKey := make([]byte, len(privateKey))
	copy(originalPrivateKey, privateKey)

	kp := NewKeyPair(privateKey)

	// Derive public key and address before destruction
	publicKey, err := kp.GetPublicKey()
	if err != nil {
		t.Fatalf("Failed to get public key: %v", err)
	}
	originalPublicKey := make([]byte, len(publicKey))
	copy(originalPublicKey, publicKey)

	_, err = kp.GetAddress()
	if err != nil {
		t.Fatalf("Failed to get address: %v", err)
	}

	// Destroy the keypair
	kp.Destroy()

	// Verify private key is zeroed
	if kp.privateKey != nil {
		t.Error("Private key reference should be nil after Destroy()")
	}

	// Verify public key is zeroed
	if kp.publicKey != nil {
		t.Error("Public key reference should be nil after Destroy()")
	}

	// Verify address is cleared
	if kp.address != nil {
		t.Error("Address reference should be nil after Destroy()")
	}

	// Verify the original byte slice was zeroed
	allZeros := true
	for _, b := range privateKey {
		if b != 0 {
			allZeros = false
			break
		}
	}
	if !allZeros {
		t.Error("Original private key bytes should be zeroed after Destroy()")
	}
}

func TestDestroy_CanCallMultipleTimes(t *testing.T) {
	// Generate a keypair
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	kp := NewKeyPair(privateKey)

	// Calling Destroy() multiple times should not panic
	kp.Destroy()
	kp.Destroy() // Second call should be safe
	kp.Destroy() // Third call should be safe
}

func TestDestroy_PreventMemoryLeaks(t *testing.T) {
	// This test demonstrates the recommended defer pattern
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Simulate a function that uses keypair and ensures cleanup
	func() {
		kp := NewKeyPair(privateKey)
		defer kp.Destroy() // Ensure cleanup even on panic

		// Use the keypair
		message := []byte("test message")
		_, err := kp.Sign(message)
		if err != nil {
			t.Errorf("Sign() failed: %v", err)
		}

		// When function exits, Destroy() is called automatically
	}()

	// After function exits, verify private key was zeroed
	allZeros := true
	for _, b := range privateKey {
		if b != 0 {
			allZeros = false
			break
		}
	}
	if !allZeros {
		t.Error("Private key should be zeroed after defer Destroy()")
	}
}
