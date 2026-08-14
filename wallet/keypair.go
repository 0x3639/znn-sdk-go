package wallet

import (
	"crypto/ed25519"
	"runtime"
	"sync"

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

// KeyPair represents an Ed25519 key pair with address.
//
// A KeyPair is safe for concurrent use: the lazy derivation of the public
// key and address, signing, and Destroy are all serialized by an internal
// mutex.
type KeyPair struct {
	mu         sync.Mutex
	privateKey []byte
	publicKey  []byte
	address    *types.Address
}

// NewKeyPair creates a new KeyPair from a private key.
// The public key and address will be derived lazily.
//
// The private key bytes are copied, so the caller may reuse or zero its own
// buffer without affecting the KeyPair's internal state.
func NewKeyPair(privateKey []byte) *KeyPair {
	return &KeyPair{
		privateKey: append([]byte(nil), privateKey...),
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

// GetPrivateKey returns a copy of the private key bytes.
//
// A copy is returned so a concurrent Destroy cannot zero the caller's slice
// and the caller cannot mutate the KeyPair's internal state.
func (kp *KeyPair) GetPrivateKey() []byte {
	kp.mu.Lock()
	defer kp.mu.Unlock()
	return append([]byte(nil), kp.privateKey...)
}

// GetPublicKey returns a copy of the public key, deriving it if necessary.
func (kp *KeyPair) GetPublicKey() ([]byte, error) {
	kp.mu.Lock()
	defer kp.mu.Unlock()
	pubKey, err := kp.getPublicKeyLocked()
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), pubKey...), nil
}

// getPublicKeyLocked derives and caches the public key, returning the internal
// slice. The caller must hold kp.mu and must not expose the slice outside the
// lock without copying it.
func (kp *KeyPair) getPublicKeyLocked() ([]byte, error) {
	if kp.publicKey == nil {
		pubKey, err := crypto.GetPublicKey(kp.privateKey)
		if err != nil {
			return nil, err
		}
		kp.publicKey = pubKey
	}
	return kp.publicKey, nil
}

// GetAddress returns the Zenon address, deriving it if necessary.
//
// A copy of the cached address is returned so the caller cannot mutate the
// KeyPair's internal state.
func (kp *KeyPair) GetAddress() (*types.Address, error) {
	kp.mu.Lock()
	defer kp.mu.Unlock()
	if kp.address == nil {
		pubKey, err := kp.getPublicKeyLocked()
		if err != nil {
			return nil, err
		}

		// Derive address from public key
		addr := types.PubKeyToAddress(pubKey)
		kp.address = &addr
	}
	addrCopy := *kp.address
	return &addrCopy, nil
}

// Sign signs a message with the private key
func (kp *KeyPair) Sign(message []byte) ([]byte, error) {
	kp.mu.Lock()
	defer kp.mu.Unlock()
	return crypto.Sign(message, kp.privateKey)
}

// Verify verifies a signature against a message using this keypair's public key.
//
// The lock is held for the entire verification so a concurrent Destroy cannot
// zero the key material mid-read.
func (kp *KeyPair) Verify(signature []byte, message []byte) (bool, error) {
	kp.mu.Lock()
	defer kp.mu.Unlock()
	pubKey, err := kp.getPublicKeyLocked()
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
	kp.mu.Lock()
	defer kp.mu.Unlock()

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
