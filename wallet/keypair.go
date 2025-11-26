package wallet

import (
	"crypto/ed25519"
	"runtime"

	"github.com/0x3639/znn-sdk-go/crypto"
	"github.com/zenon-network/go-zenon/common/types"
)

// zeroBytes securely zeros a byte slice in a way that cannot be optimized away
// by the compiler. This is critical for clearing sensitive data like private keys
// from memory.
//
// The runtime.KeepAlive call ensures the slice remains reachable until after
// the zeroing operation completes, preventing the compiler from optimizing
// away the writes as "dead stores".
func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
	// Ensure the slice is kept alive until zeroing completes.
	// This prevents the compiler from optimizing away the zeroing loop
	// since it sees the slice is "used" after the writes.
	runtime.KeepAlive(b)
}

// KeyPair represents an Ed25519 key pair with address
type KeyPair struct {
	privateKey []byte
	publicKey  []byte
	address    *types.Address
}

// NewKeyPair creates a new KeyPair from a private key
// The public key and address will be derived lazily
func NewKeyPair(privateKey []byte) *KeyPair {
	return &KeyPair{
		privateKey: privateKey,
	}
}

// NewKeyPairFromSeed creates a KeyPair from a 32-byte seed
func NewKeyPairFromSeed(seed []byte) (*KeyPair, error) {
	if len(seed) != 32 {
		return nil, ErrInvalidPrivateKey
	}

	privateKey := ed25519.NewKeyFromSeed(seed)
	return NewKeyPair(privateKey), nil
}

// GetPrivateKey returns the private key bytes
func (kp *KeyPair) GetPrivateKey() []byte {
	return kp.privateKey
}

// GetPublicKey returns the public key, deriving it if necessary
func (kp *KeyPair) GetPublicKey() ([]byte, error) {
	if kp.publicKey == nil {
		pubKey, err := crypto.GetPublicKey(kp.privateKey)
		if err != nil {
			return nil, err
		}
		kp.publicKey = pubKey
	}
	return kp.publicKey, nil
}

// GetAddress returns the Zenon address, deriving it if necessary
func (kp *KeyPair) GetAddress() (*types.Address, error) {
	if kp.address == nil {
		pubKey, err := kp.GetPublicKey()
		if err != nil {
			return nil, err
		}

		// Derive address from public key
		addr := types.PubKeyToAddress(pubKey)
		kp.address = &addr
	}
	return kp.address, nil
}

// Sign signs a message with the private key
func (kp *KeyPair) Sign(message []byte) ([]byte, error) {
	return crypto.Sign(message, kp.privateKey)
}

// Verify verifies a signature against a message using this keypair's public key
func (kp *KeyPair) Verify(signature []byte, message []byte) (bool, error) {
	pubKey, err := kp.GetPublicKey()
	if err != nil {
		return false, err
	}

	return crypto.Verify(signature, message, pubKey)
}

// Destroy securely zeros out the private key from memory
// This method should be called when the KeyPair is no longer needed
// to prevent the private key from lingering in memory.
//
// IMPORTANT: After calling Destroy(), the KeyPair should not be used
// for any operations. Attempting to use it will result in undefined behavior.
//
// Example:
//
//	kp, _ := NewKeyPairFromSeed(seed)
//	defer kp.Destroy()  // Ensure cleanup even if function panics
//	// ... use keypair for signing ...
func (kp *KeyPair) Destroy() {
	// Zero out private key bytes using secure zeroing
	if kp.privateKey != nil {
		zeroBytes(kp.privateKey)
		kp.privateKey = nil
	}

	// Zero out public key bytes (defense in depth)
	if kp.publicKey != nil {
		zeroBytes(kp.publicKey)
		kp.publicKey = nil
	}

	// Clear address reference
	kp.address = nil
}

// GeneratePublicKey is a static method that generates a public key from a private key
func GeneratePublicKey(privateKey []byte) ([]byte, error) {
	return crypto.GetPublicKey(privateKey)
}
