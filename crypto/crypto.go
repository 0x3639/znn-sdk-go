package crypto

import (
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/sha3"
)

// GetPublicKey derives the Ed25519 public key from a private key
func GetPublicKey(privateKey []byte) ([]byte, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d", ed25519.PrivateKeySize, len(privateKey))
	}

	privKey := ed25519.PrivateKey(privateKey)
	pubKey, ok := privKey.Public().(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to derive public key: type assertion failed")
	}

	return []byte(pubKey), nil
}

// Sign creates an Ed25519 signature of a message using a private key
func Sign(message []byte, privateKey []byte) ([]byte, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d", ed25519.PrivateKeySize, len(privateKey))
	}

	privKey := ed25519.PrivateKey(privateKey)
	signature := ed25519.Sign(privKey, message)

	return signature, nil
}

// Verify verifies an Ed25519 signature
func Verify(signature []byte, message []byte, publicKey []byte) (bool, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return false, fmt.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(publicKey))
	}

	if len(signature) != ed25519.SignatureSize {
		return false, fmt.Errorf("invalid signature size: expected %d, got %d", ed25519.SignatureSize, len(signature))
	}

	pubKey := ed25519.PublicKey(publicKey)
	return ed25519.Verify(pubKey, message, signature), nil
}

// Digest computes the SHA3-256 hash of data
// The digestSize parameter allows customization of output length (default: 32 bytes)
func Digest(data []byte, digestSize int) []byte {
	if digestSize == 0 || digestSize == 32 {
		// Standard SHA3-256
		hash := sha3.Sum256(data)
		return hash[:]
	}

	// Use SHAKE256 for custom output sizes
	hasher := sha3.NewShake256()
	// #nosec G104 -- shake.Write never returns an error
	hasher.Write(data) //nolint:errcheck
	result := make([]byte, digestSize)
	// #nosec G104 -- shake.Read always succeeds with sufficient buffer
	hasher.Read(result) //nolint:errcheck
	return result
}

// DigestDefault computes SHA3-256 hash with default 32-byte output
func DigestDefault(data []byte) []byte {
	return Digest(data, 32)
}

// SHA256Bytes computes the SHA-256 hash of data
func SHA256Bytes(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
