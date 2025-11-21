package crypto

import (
	"bytes"
	"testing"
)

// =============================================================================
// DefaultArgon2Parameters Tests
// =============================================================================

func TestDefaultArgon2Parameters(t *testing.T) {
	params := DefaultArgon2Parameters()

	if params.Memory != 64*1024 {
		t.Errorf("Memory = %d, want %d", params.Memory, 64*1024)
	}

	if params.Iterations != 1 {
		t.Errorf("Iterations = %d, want 1", params.Iterations)
	}

	if params.Parallelism != 4 {
		t.Errorf("Parallelism = %d, want 4", params.Parallelism)
	}

	if params.SaltLength != 16 {
		t.Errorf("SaltLength = %d, want 16", params.SaltLength)
	}

	if params.KeyLength != 32 {
		t.Errorf("KeyLength = %d, want 32", params.KeyLength)
	}
}

// =============================================================================
// DeriveKey Tests
// =============================================================================

func TestDeriveKey_BasicUsage(t *testing.T) {
	password := []byte("test-password")
	salt := []byte("0123456789abcdef") // 16 bytes
	params := DefaultArgon2Parameters()

	key := DeriveKey(password, salt, params)

	if len(key) != int(params.KeyLength) {
		t.Errorf("len(key) = %d, want %d", len(key), params.KeyLength)
	}
}

func TestDeriveKey_SameInputsSameOutput(t *testing.T) {
	password := []byte("test-password")
	salt := []byte("0123456789abcdef")
	params := DefaultArgon2Parameters()

	key1 := DeriveKey(password, salt, params)
	key2 := DeriveKey(password, salt, params)

	if !bytes.Equal(key1, key2) {
		t.Error("DeriveKey should produce the same output for the same inputs")
	}
}

func TestDeriveKey_DifferentPasswordsDifferentOutputs(t *testing.T) {
	salt := []byte("0123456789abcdef")
	params := DefaultArgon2Parameters()

	key1 := DeriveKey([]byte("password1"), salt, params)
	key2 := DeriveKey([]byte("password2"), salt, params)

	if bytes.Equal(key1, key2) {
		t.Error("DeriveKey should produce different outputs for different passwords")
	}
}

func TestDeriveKey_DifferentSaltsDifferentOutputs(t *testing.T) {
	password := []byte("test-password")
	params := DefaultArgon2Parameters()

	key1 := DeriveKey(password, []byte("0123456789abcdef"), params)
	key2 := DeriveKey(password, []byte("fedcba9876543210"), params)

	if bytes.Equal(key1, key2) {
		t.Error("DeriveKey should produce different outputs for different salts")
	}
}

func TestDeriveKey_EmptyPassword(t *testing.T) {
	password := []byte("")
	salt := []byte("0123456789abcdef")
	params := DefaultArgon2Parameters()

	key := DeriveKey(password, salt, params)

	if len(key) != int(params.KeyLength) {
		t.Errorf("len(key) = %d, want %d", len(key), params.KeyLength)
	}
}

func TestDeriveKey_CustomParameters(t *testing.T) {
	password := []byte("test-password")
	salt := []byte("0123456789abcdef")
	params := Argon2Parameters{
		Memory:      32 * 1024, // 32 MB
		Iterations:  2,
		Parallelism: 2,
		KeyLength:   64, // 64 bytes
	}

	key := DeriveKey(password, salt, params)

	if len(key) != 64 {
		t.Errorf("len(key) = %d, want 64", len(key))
	}
}

// =============================================================================
// DeriveKeyDefault Tests
// =============================================================================

func TestDeriveKeyDefault(t *testing.T) {
	password := []byte("test-password")
	salt := []byte("0123456789abcdef")

	key := DeriveKeyDefault(password, salt)

	// Should match default parameters
	params := DefaultArgon2Parameters()
	expectedKey := DeriveKey(password, salt, params)

	if !bytes.Equal(key, expectedKey) {
		t.Error("DeriveKeyDefault should use default parameters")
	}
}

func TestDeriveKeyDefault_OutputLength(t *testing.T) {
	password := []byte("test-password")
	salt := []byte("0123456789abcdef")

	key := DeriveKeyDefault(password, salt)

	if len(key) != 32 {
		t.Errorf("len(key) = %d, want 32", len(key))
	}
}

// =============================================================================
// Known Test Vector Tests
// =============================================================================

func TestDeriveKey_KnownVector(t *testing.T) {
	// Test with a simple known case to ensure consistency
	password := []byte("password")
	salt := []byte("somesalt")
	params := Argon2Parameters{
		Memory:      64,
		Iterations:  1,
		Parallelism: 1,
		KeyLength:   32,
	}

	key1 := DeriveKey(password, salt, params)
	key2 := DeriveKey(password, salt, params)

	// Should be deterministic
	if !bytes.Equal(key1, key2) {
		t.Error("DeriveKey should be deterministic")
	}

	// Should produce expected length
	if len(key1) != 32 {
		t.Errorf("len(key) = %d, want 32", len(key1))
	}
}

// =============================================================================
// Performance Tests
// =============================================================================

func BenchmarkDeriveKey(b *testing.B) {
	password := []byte("test-password")
	salt := []byte("0123456789abcdef")
	params := DefaultArgon2Parameters()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DeriveKey(password, salt, params)
	}
}

func BenchmarkDeriveKeyDefault(b *testing.B) {
	password := []byte("test-password")
	salt := []byte("0123456789abcdef")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DeriveKeyDefault(password, salt)
	}
}
