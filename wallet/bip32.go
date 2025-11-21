package wallet

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// KeyData represents a BIP32 extended key (private key + chain code)
type KeyData struct {
	Key       []byte // 32 bytes for Ed25519
	ChainCode []byte // 32 bytes
}

const (
	// HardenedKeyStart is the index at which hardened keys start (2^31)
	HardenedKeyStart = 0x80000000

	// Ed25519Curve is the curve name for SLIP-0010
	Ed25519Curve = "ed25519 seed"
)

// GetMasterKeyFromSeed derives the master key from a seed using SLIP-0010
// This follows the SLIP-0010 specification for Ed25519
func GetMasterKeyFromSeed(seed []byte) (*KeyData, error) {
	if len(seed) < 16 || len(seed) > 64 {
		return nil, errors.New("seed must be between 16 and 64 bytes")
	}

	// HMAC-SHA512 with key = "ed25519 seed"
	h := hmac.New(sha512.New, []byte(Ed25519Curve))
	h.Write(seed)
	digest := h.Sum(nil)

	// Split into key (first 32 bytes) and chain code (last 32 bytes)
	key := digest[:32]
	chainCode := digest[32:]

	return &KeyData{
		Key:       key,
		ChainCode: chainCode,
	}, nil
}

// DerivePath derives a key from a BIP44 path string like "m/44'/73404'/0'"
func DerivePath(path string, seed []byte) (*KeyData, error) {
	if path == "" {
		return nil, errors.New("derivation path cannot be empty")
	}

	// Get master key
	master, err := GetMasterKeyFromSeed(seed)
	if err != nil {
		return nil, err
	}

	// Parse path
	components := strings.Split(path, "/")
	if len(components) == 0 {
		return nil, errors.New("invalid derivation path")
	}

	// Skip "m" prefix if present
	if components[0] == "m" {
		components = components[1:]
	}

	// If no components left after skipping "m", return master key
	if len(components) == 0 {
		return master, nil
	}

	// Derive each level
	current := master
	for _, component := range components {
		if component == "" {
			continue
		}

		// Check for hardened key (ends with ')
		hardened := strings.HasSuffix(component, "'")
		indexStr := strings.TrimSuffix(component, "'")

		index, err := strconv.ParseUint(indexStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid path component: %s", component)
		}

		// Convert to hardened index if needed
		if hardened {
			index += HardenedKeyStart
		}

		// Derive child key
		current, err = getCKDPriv(current, uint32(index))
		if err != nil {
			return nil, err
		}
	}

	return current, nil
}

// DeriveKey is a convenience function that derives a key from a path and seed
// and returns the Ed25519 private key
func DeriveKey(path string, seedHex string) ([]byte, error) {
	// Import encoding/hex at the top if not already imported
	// Convert hex seed to bytes using standard library
	seed, err := hexDecode(seedHex)
	if err != nil {
		return nil, fmt.Errorf("invalid seed hex: %w", err)
	}

	keyData, err := DerivePath(path, seed)
	if err != nil {
		return nil, err
	}

	// For Ed25519, we need to derive the actual private key from the key material
	// The 32-byte key from SLIP-0010 is used as the seed for Ed25519
	privateKey := ed25519.NewKeyFromSeed(keyData.Key)

	return privateKey, nil
}

// hexDecode decodes a hex string to bytes
func hexDecode(s string) ([]byte, error) {
	if len(s)%2 != 0 {
		return nil, errors.New("hex string must have even length")
	}

	result := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		b := byte(0)
		for j := 0; j < 2; j++ {
			c := s[i+j]
			b <<= 4
			if c >= '0' && c <= '9' {
				b |= c - '0'
			} else if c >= 'a' && c <= 'f' {
				b |= c - 'a' + 10
			} else if c >= 'A' && c <= 'F' {
				b |= c - 'A' + 10
			} else {
				return nil, fmt.Errorf("invalid hex character: %c", c)
			}
		}
		result[i/2] = b
	}

	return result, nil
}

// getCKDPriv derives a child key from a parent key using SLIP-0010
// This implements Child Key Derivation (CKD) for private keys
func getCKDPriv(parent *KeyData, index uint32) (*KeyData, error) {
	if parent == nil {
		return nil, errors.New("parent key is nil")
	}

	// For Ed25519, only hardened derivation is supported
	if index < HardenedKeyStart {
		return nil, errors.New("Ed25519 only supports hardened derivation")
	}

	// Create HMAC-SHA512
	h := hmac.New(sha512.New, parent.ChainCode)

	// Write 0x00 + parent key + index (big endian)
	h.Write([]byte{0x00})
	h.Write(parent.Key)

	indexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(indexBytes, index)
	h.Write(indexBytes)

	// Get digest
	digest := h.Sum(nil)

	// Split into key (first 32 bytes) and chain code (last 32 bytes)
	key := digest[:32]
	chainCode := digest[32:]

	return &KeyData{
		Key:       key,
		ChainCode: chainCode,
	}, nil
}

// GetPublicKey derives the Ed25519 public key from the key data
func (kd *KeyData) GetPublicKey() ([]byte, error) {
	if len(kd.Key) != 32 {
		return nil, errors.New("invalid key length")
	}

	// Derive Ed25519 private key from key material
	privateKey := ed25519.NewKeyFromSeed(kd.Key)

	// Get public key
	publicKey, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to derive public key: type assertion failed")
	}

	return []byte(publicKey), nil
}
