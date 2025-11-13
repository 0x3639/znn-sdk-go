package crypto

import (
	"golang.org/x/crypto/argon2"
)

// Argon2Parameters represents the parameters for Argon2 key derivation
type Argon2Parameters struct {
	Memory      uint32 // Memory in KB
	Iterations  uint32 // Number of iterations
	Parallelism uint8  // Degree of parallelism
	SaltLength  uint32 // Length of salt in bytes
	KeyLength   uint32 // Length of derived key in bytes
}

// DefaultArgon2Parameters returns the default parameters used by the Zenon SDK
// These match the parameters used in the Dart SDK:
// - Memory: 64 * 1024 KB (64 MB)
// - Iterations: 1
// - Parallelism: 4
// - KeyLength: 32 bytes
func DefaultArgon2Parameters() Argon2Parameters {
	return Argon2Parameters{
		Memory:      64 * 1024,
		Iterations:  1,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// DeriveKey derives a key from a password using Argon2id
// This is used for wallet encryption/decryption
func DeriveKey(password []byte, salt []byte, params Argon2Parameters) []byte {
	return argon2.IDKey(
		password,
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)
}

// DeriveKeyDefault derives a key using the default parameters
func DeriveKeyDefault(password []byte, salt []byte) []byte {
	params := DefaultArgon2Parameters()
	return DeriveKey(password, salt, params)
}
