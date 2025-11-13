package crypto

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"
)

// =============================================================================
// GetPublicKey Tests
// =============================================================================

func TestGetPublicKey_ValidPrivateKey(t *testing.T) {
	// Generate a valid Ed25519 private key
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// Get public key using our function
	derivedPubKey, err := GetPublicKey(privKey)
	if err != nil {
		t.Fatalf("GetPublicKey() error = %v", err)
	}

	// Should match the generated public key
	if !bytes.Equal(derivedPubKey, pubKey) {
		t.Error("Derived public key does not match generated public key")
	}
}

func TestGetPublicKey_InvalidPrivateKeySize(t *testing.T) {
	testCases := []struct {
		name string
		key  []byte
	}{
		{"Too short", make([]byte, 32)},
		{"Too long", make([]byte, 128)},
		{"Empty", []byte{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GetPublicKey(tc.key)
			if err == nil {
				t.Error("GetPublicKey() should return error for invalid key size")
			}
		})
	}
}

func TestGetPublicKey_OutputSize(t *testing.T) {
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	pubKey, err := GetPublicKey(privKey)
	if err != nil {
		t.Fatalf("GetPublicKey() error = %v", err)
	}

	if len(pubKey) != ed25519.PublicKeySize {
		t.Errorf("len(pubKey) = %d, want %d", len(pubKey), ed25519.PublicKeySize)
	}
}

// =============================================================================
// Sign Tests
// =============================================================================

func TestSign_ValidSignature(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	message := []byte("test message")

	signature, err := Sign(message, privKey)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	// Verify signature using standard library
	valid := ed25519.Verify(pubKey, message, signature)
	if !valid {
		t.Error("Signature should be valid")
	}
}

func TestSign_InvalidPrivateKeySize(t *testing.T) {
	testCases := []struct {
		name string
		key  []byte
	}{
		{"Too short", make([]byte, 32)},
		{"Too long", make([]byte, 128)},
		{"Empty", []byte{}},
	}

	message := []byte("test message")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Sign(message, tc.key)
			if err == nil {
				t.Error("Sign() should return error for invalid key size")
			}
		})
	}
}

func TestSign_EmptyMessage(t *testing.T) {
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	signature, err := Sign([]byte{}, privKey)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if len(signature) != ed25519.SignatureSize {
		t.Errorf("len(signature) = %d, want %d", len(signature), ed25519.SignatureSize)
	}
}

func TestSign_DifferentMessagesDifferentSignatures(t *testing.T) {
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	sig1, err := Sign([]byte("message1"), privKey)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	sig2, err := Sign([]byte("message2"), privKey)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	if bytes.Equal(sig1, sig2) {
		t.Error("Different messages should produce different signatures")
	}
}

func TestSign_SameMessageSameSignature(t *testing.T) {
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	message := []byte("test message")

	sig1, err := Sign(message, privKey)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	sig2, err := Sign(message, privKey)
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
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	message := []byte("test message")
	signature, err := Sign(message, privKey)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	valid, err := Verify(signature, message, pubKey)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if !valid {
		t.Error("Valid signature should verify successfully")
	}
}

func TestVerify_InvalidSignature(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	message := []byte("test message")
	signature, err := Sign(message, privKey)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	// Tamper with signature
	signature[0] ^= 0xFF

	valid, err := Verify(signature, message, pubKey)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if valid {
		t.Error("Tampered signature should not verify")
	}
}

func TestVerify_WrongMessage(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	message := []byte("test message")
	signature, err := Sign(message, privKey)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	// Verify with different message
	wrongMessage := []byte("wrong message")
	valid, err := Verify(signature, wrongMessage, pubKey)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if valid {
		t.Error("Signature should not verify with wrong message")
	}
}

func TestVerify_WrongPublicKey(t *testing.T) {
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	wrongPubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	message := []byte("test message")
	signature, err := Sign(message, privKey)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	valid, err := Verify(signature, message, wrongPubKey)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if valid {
		t.Error("Signature should not verify with wrong public key")
	}
}

func TestVerify_InvalidPublicKeySize(t *testing.T) {
	testCases := []struct {
		name string
		key  []byte
	}{
		{"Too short", make([]byte, 16)},
		{"Too long", make([]byte, 64)},
		{"Empty", []byte{}},
	}

	signature := make([]byte, ed25519.SignatureSize)
	message := []byte("test message")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Verify(signature, message, tc.key)
			if err == nil {
				t.Error("Verify() should return error for invalid public key size")
			}
		})
	}
}

func TestVerify_InvalidSignatureSize(t *testing.T) {
	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	testCases := []struct {
		name      string
		signature []byte
	}{
		{"Too short", make([]byte, 32)},
		{"Too long", make([]byte, 128)},
		{"Empty", []byte{}},
	}

	message := []byte("test message")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Verify(tc.signature, message, pubKey)
			if err == nil {
				t.Error("Verify() should return error for invalid signature size")
			}
		})
	}
}

// =============================================================================
// Digest Tests (SHA3-256)
// =============================================================================

func TestDigest_DefaultSize(t *testing.T) {
	data := []byte("test data")
	hash := Digest(data, 32)

	if len(hash) != 32 {
		t.Errorf("len(hash) = %d, want 32", len(hash))
	}
}

func TestDigest_ZeroSize(t *testing.T) {
	data := []byte("test data")
	hash := Digest(data, 0)

	// Zero should use default (32)
	if len(hash) != 32 {
		t.Errorf("len(hash) = %d, want 32", len(hash))
	}
}

func TestDigest_CustomSize(t *testing.T) {
	data := []byte("test data")
	hash := Digest(data, 64)

	if len(hash) != 64 {
		t.Errorf("len(hash) = %d, want 64", len(hash))
	}
}

func TestDigest_EmptyData(t *testing.T) {
	hash := Digest([]byte{}, 32)

	if len(hash) != 32 {
		t.Errorf("len(hash) = %d, want 32", len(hash))
	}

	// Empty data should produce a specific hash (not all zeros)
	allZeros := make([]byte, 32)
	if bytes.Equal(hash, allZeros) {
		t.Error("Hash of empty data should not be all zeros")
	}
}

func TestDigest_Deterministic(t *testing.T) {
	data := []byte("test data")

	hash1 := Digest(data, 32)
	hash2 := Digest(data, 32)

	if !bytes.Equal(hash1, hash2) {
		t.Error("Digest should be deterministic")
	}
}

func TestDigest_DifferentDataDifferentHash(t *testing.T) {
	hash1 := Digest([]byte("data1"), 32)
	hash2 := Digest([]byte("data2"), 32)

	if bytes.Equal(hash1, hash2) {
		t.Error("Different data should produce different hashes")
	}
}

func TestDigest_KnownVector(t *testing.T) {
	// Test with known SHA3-256 test vector
	// "abc" -> BA7816BF8F01CFEA414140DE5DAE2223B00361A396177A9CB410FF61F20015AD (SHA-256, not SHA3)
	// SHA3-256("") = A7FFC6F8BF1ED76651C14756A061D662F580FF4DE43B49FA82D80A4B80F8434A
	data := []byte("")
	hash := Digest(data, 32)

	expectedHex := "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"
	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		t.Fatalf("Failed to decode expected hash: %v", err)
	}

	if !bytes.Equal(hash, expected) {
		t.Errorf("Digest() = %x, want %x", hash, expected)
	}
}

// =============================================================================
// DigestDefault Tests
// =============================================================================

func TestDigestDefault(t *testing.T) {
	data := []byte("test data")

	hash := DigestDefault(data)

	if len(hash) != 32 {
		t.Errorf("len(hash) = %d, want 32", len(hash))
	}
}

func TestDigestDefault_MatchesDigest32(t *testing.T) {
	data := []byte("test data")

	hash1 := DigestDefault(data)
	hash2 := Digest(data, 32)

	if !bytes.Equal(hash1, hash2) {
		t.Error("DigestDefault should match Digest with size 32")
	}
}

// =============================================================================
// SHA256Bytes Tests
// =============================================================================

func TestSHA256Bytes(t *testing.T) {
	data := []byte("test data")
	hash := SHA256Bytes(data)

	if len(hash) != 32 {
		t.Errorf("len(hash) = %d, want 32", len(hash))
	}
}

func TestSHA256Bytes_EmptyData(t *testing.T) {
	hash := SHA256Bytes([]byte{})

	if len(hash) != 32 {
		t.Errorf("len(hash) = %d, want 32", len(hash))
	}

	// SHA-256("") = E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855
	expectedHex := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		t.Fatalf("Failed to decode expected hash: %v", err)
	}

	if !bytes.Equal(hash, expected) {
		t.Errorf("SHA256Bytes() = %x, want %x", hash, expected)
	}
}

func TestSHA256Bytes_Deterministic(t *testing.T) {
	data := []byte("test data")

	hash1 := SHA256Bytes(data)
	hash2 := SHA256Bytes(data)

	if !bytes.Equal(hash1, hash2) {
		t.Error("SHA256Bytes should be deterministic")
	}
}

func TestSHA256Bytes_DifferentDataDifferentHash(t *testing.T) {
	hash1 := SHA256Bytes([]byte("data1"))
	hash2 := SHA256Bytes([]byte("data2"))

	if bytes.Equal(hash1, hash2) {
		t.Error("Different data should produce different hashes")
	}
}

// =============================================================================
// Round Trip Tests
// =============================================================================

func TestSignVerifyRoundTrip(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	testMessages := [][]byte{
		[]byte("test message"),
		[]byte(""),
		[]byte("a"),
		[]byte("The quick brown fox jumps over the lazy dog"),
		make([]byte, 1000), // Large message
	}

	for _, message := range testMessages {
		signature, err := Sign(message, privKey)
		if err != nil {
			t.Fatalf("Sign() error = %v", err)
		}

		valid, err := Verify(signature, message, pubKey)
		if err != nil {
			t.Fatalf("Verify() error = %v", err)
		}

		if !valid {
			t.Errorf("Round trip failed for message of length %d", len(message))
		}
	}
}

// =============================================================================
// Performance Benchmarks
// =============================================================================

func BenchmarkGetPublicKey(b *testing.B) {
	_, privKey, _ := ed25519.GenerateKey(rand.Reader)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetPublicKey(privKey)
	}
}

func BenchmarkSign(b *testing.B) {
	_, privKey, _ := ed25519.GenerateKey(rand.Reader)
	message := []byte("test message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sign(message, privKey)
	}
}

func BenchmarkVerify(b *testing.B) {
	pubKey, privKey, _ := ed25519.GenerateKey(rand.Reader)
	message := []byte("test message")
	signature, _ := Sign(message, privKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Verify(signature, message, pubKey)
	}
}

func BenchmarkDigest(b *testing.B) {
	data := []byte("test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Digest(data, 32)
	}
}

func BenchmarkSHA256Bytes(b *testing.B) {
	data := []byte("test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SHA256Bytes(data)
	}
}
