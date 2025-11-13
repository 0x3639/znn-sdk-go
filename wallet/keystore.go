package wallet

import (
	"encoding/hex"
	"fmt"

	"github.com/zenon-network/go-zenon/common/types"
)

// KeyStore represents a hierarchical deterministic wallet
type KeyStore struct {
	Mnemonic string
	Entropy  []byte
	Seed     []byte
}

// NewKeyStoreFromMnemonic creates a KeyStore from a BIP39 mnemonic
func NewKeyStoreFromMnemonic(mnemonic string) (*KeyStore, error) {
	if !ValidateMnemonicString(mnemonic) {
		return nil, ErrInvalidMnemonic
	}

	entropy, err := MnemonicToEntropy(mnemonic)
	if err != nil {
		return nil, err
	}

	seed := MnemonicToSeed(mnemonic, "")

	return &KeyStore{
		Mnemonic: mnemonic,
		Entropy:  entropy,
		Seed:     seed,
	}, nil
}

// NewKeyStoreFromSeed creates a KeyStore from a seed
func NewKeyStoreFromSeed(seedHex string) (*KeyStore, error) {
	seed, err := hex.DecodeString(seedHex)
	if err != nil {
		return nil, fmt.Errorf("invalid seed hex: %w", err)
	}

	return &KeyStore{
		Seed: seed,
	}, nil
}

// NewKeyStoreFromEntropy creates a KeyStore from entropy bytes
func NewKeyStoreFromEntropy(entropy []byte) (*KeyStore, error) {
	if len(entropy) != 16 && len(entropy) != 32 {
		return nil, ErrInvalidEntropy
	}

	mnemonic, err := EntropyToMnemonic(entropy)
	if err != nil {
		return nil, err
	}

	return NewKeyStoreFromMnemonic(mnemonic)
}

// NewKeyStoreRandom creates a new KeyStore with random entropy (256 bits)
func NewKeyStoreRandom() (*KeyStore, error) {
	mnemonic, err := GenerateMnemonic(256)
	if err != nil {
		return nil, err
	}

	return NewKeyStoreFromMnemonic(mnemonic)
}

// GetKeyPair derives a KeyPair at the given account index
// Uses BIP44 path: m/44'/73404'/account'
func (ks *KeyStore) GetKeyPair(account int) (*KeyPair, error) {
	if ks.Seed == nil {
		return nil, fmt.Errorf("keystore seed not initialized")
	}

	// Derive using BIP44 path
	path := GetDerivationAccount(account)
	keyData, err := DerivePath(path, ks.Seed)
	if err != nil {
		return nil, err
	}

	// Create keypair from derived key
	kp, err := NewKeyPairFromSeed(keyData.Key)
	if err != nil {
		return nil, err
	}

	return kp, nil
}

// DeriveAddressesByRange derives addresses for a range of account indices
// Returns slice of addresses from left (inclusive) to right (exclusive)
func (ks *KeyStore) DeriveAddressesByRange(left, right int) ([]*types.Address, error) {
	if left < 0 || right < left {
		return nil, fmt.Errorf("invalid range: [%d, %d)", left, right)
	}

	addresses := make([]*types.Address, 0, right-left)

	for i := left; i < right; i++ {
		kp, err := ks.GetKeyPair(i)
		if err != nil {
			return nil, fmt.Errorf("failed to derive account %d: %w", i, err)
		}

		addr, err := kp.GetAddress()
		if err != nil {
			return nil, fmt.Errorf("failed to get address for account %d: %w", i, err)
		}

		addresses = append(addresses, addr)
	}

	return addresses, nil
}

// FindResponse represents the result of finding an address in the keystore
type FindResponse struct {
	Index   int
	KeyPair *KeyPair
}

// FindAddress searches for an address in the keystore up to maxAccounts
// Returns FindResponse with the account index and keypair if found
func (ks *KeyStore) FindAddress(address types.Address, maxAccounts int) (*FindResponse, error) {
	if maxAccounts <= 0 {
		maxAccounts = DefaultMaxIndex
	}

	for i := 0; i < maxAccounts; i++ {
		kp, err := ks.GetKeyPair(i)
		if err != nil {
			return nil, err
		}

		addr, err := kp.GetAddress()
		if err != nil {
			return nil, err
		}

		if addr.String() == address.String() {
			return &FindResponse{
				Index:   i,
				KeyPair: kp,
			}, nil
		}
	}

	return nil, ErrAddressNotFound
}

// GetBaseAddress returns the address at account index 0
func (ks *KeyStore) GetBaseAddress() (*types.Address, error) {
	kp, err := ks.GetKeyPair(0)
	if err != nil {
		return nil, err
	}

	return kp.GetAddress()
}

// ToEncryptedFile encrypts the keystore to an EncryptedFile
func (ks *KeyStore) ToEncryptedFile(password string, metadata map[string]interface{}) (*EncryptedFile, error) {
	// Prepare keystore data for encryption
	data := make(map[string]interface{})

	if ks.Mnemonic != "" {
		data["mnemonic"] = ks.Mnemonic
	}

	if ks.Entropy != nil {
		data["entropy"] = hex.EncodeToString(ks.Entropy)
	}

	if ks.Seed != nil {
		data["seed"] = hex.EncodeToString(ks.Seed)
	}

	// Serialize to JSON
	jsonData, err := serializeKeyStoreData(data)
	if err != nil {
		return nil, err
	}

	// Add base address to metadata if not present
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	if _, hasBaseAddr := metadata[BaseAddressKey]; !hasBaseAddr {
		baseAddr, err := ks.GetBaseAddress()
		if err != nil {
			return nil, fmt.Errorf("failed to get base address: %w", err)
		}
		metadata[BaseAddressKey] = baseAddr.String()
	}

	if _, hasWalletType := metadata[WalletTypeKey]; !hasWalletType {
		metadata[WalletTypeKey] = KeyStoreWalletType
	}

	// Encrypt
	return Encrypt(jsonData, password, metadata)
}

// FromEncryptedFile decrypts an EncryptedFile to a KeyStore
func FromEncryptedFile(ef *EncryptedFile, password string) (*KeyStore, error) {
	// Decrypt
	plaintext, err := ef.Decrypt(password)
	if err != nil {
		return nil, err
	}

	// Parse keystore data
	data, err := deserializeKeyStoreData(plaintext)
	if err != nil {
		return nil, err
	}

	// Try to construct KeyStore from available data
	if mnemonic, ok := data["mnemonic"].(string); ok && mnemonic != "" {
		return NewKeyStoreFromMnemonic(mnemonic)
	}

	if entropyHex, ok := data["entropy"].(string); ok && entropyHex != "" {
		entropy, err := hex.DecodeString(entropyHex)
		if err != nil {
			return nil, fmt.Errorf("invalid entropy: %w", err)
		}
		return NewKeyStoreFromEntropy(entropy)
	}

	if seedHex, ok := data["seed"].(string); ok && seedHex != "" {
		return NewKeyStoreFromSeed(seedHex)
	}

	return nil, fmt.Errorf("encrypted file does not contain valid keystore data")
}

// Helper functions for JSON serialization
func serializeKeyStoreData(data map[string]interface{}) ([]byte, error) {
	// Simple JSON marshaling
	result := "{"
	first := true
	for k, v := range data {
		if !first {
			result += ","
		}
		first = false
		result += fmt.Sprintf(`"%s":"%v"`, k, v)
	}
	result += "}"
	return []byte(result), nil
}

func deserializeKeyStoreData(data []byte) (map[string]interface{}, error) {
	// Simple JSON parsing (for our simple key-value structure)
	result := make(map[string]interface{})

	// Parse manually (simplified for our use case)
	str := string(data)
	if len(str) < 2 || str[0] != '{' || str[len(str)-1] != '}' {
		return nil, fmt.Errorf("invalid JSON format")
	}

	// Remove braces
	str = str[1 : len(str)-1]

	// Parse key-value pairs
	// This is a simplified parser - in production, use encoding/json
	// For now, assume simple "key":"value" format

	// Split by comma (simplified - doesn't handle nested objects)
	parts := splitByComma(str)
	for _, part := range parts {
		// Split by colon
		colonIdx := -1
		inQuote := false
		for i, c := range part {
			if c == '"' {
				inQuote = !inQuote
			} else if c == ':' && !inQuote {
				colonIdx = i
				break
			}
		}

		if colonIdx == -1 {
			continue
		}

		key := trimQuotes(part[:colonIdx])
		value := trimQuotes(part[colonIdx+1:])
		result[key] = value
	}

	return result, nil
}

func splitByComma(s string) []string {
	var result []string
	var current string
	inQuote := false

	for _, c := range s {
		if c == '"' {
			inQuote = !inQuote
			current += string(c)
		} else if c == ',' && !inQuote {
			if len(current) > 0 {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}

	if len(current) > 0 {
		result = append(result, current)
	}

	return result
}

func trimQuotes(s string) string {
	s = trimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
