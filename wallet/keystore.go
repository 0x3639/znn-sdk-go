package wallet

import (
	"encoding/hex"
	"encoding/json"
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

// GetKeyPair derives a keypair at the specified BIP44 account index.
//
// The derivation follows BIP44 path: m/44'/73404'/account'/0'/0' where:
//   - 44 is the BIP44 standard
//   - 73404 is Zenon's registered coin type
//   - account is the index you specify
//
// Each account index generates a unique address from the same mnemonic/seed.
// This allows deriving multiple addresses from a single backup mnemonic.
//
// Parameters:
//   - account: Account index (0 for first address, 1 for second, etc.)
//
// Returns a KeyPair that can:
//   - Get the Zenon address
//   - Sign transactions
//   - Access public/private keys
//
// Example:
//
//	// Get first address (index 0 - this is the default/base address)
//	keypair0, err := keystore.GetKeyPair(0)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	addr0, _ := keypair0.GetAddress()
//	fmt.Println("First address:", addr0)
//
//	// Get second address (index 1)
//	keypair1, _ := keystore.GetKeyPair(1)
//	addr1, _ := keypair1.GetAddress()
//	fmt.Println("Second address:", addr1)
//
// Note: GetKeyPair(0) returns the base address - the primary address for this wallet.
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

// DeriveAddressesByRange derives multiple addresses efficiently in a single operation.
//
// This is useful for:
//   - Displaying multiple addresses to the user
//   - Searching for addresses with specific properties
//   - Generating address pools for services
//
// The range is [left, right) - includes left, excludes right.
//
// Parameters:
//   - left: Starting account index (inclusive)
//   - right: Ending account index (exclusive)
//
// Returns a slice of addresses in order, or an error if derivation fails.
//
// Example:
//
//	// Derive first 5 addresses (indices 0-4)
//	addresses, err := keystore.DeriveAddressesByRange(0, 5)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for i, addr := range addresses {
//	    fmt.Printf("Address %d: %s\n", i, addr)
//	}
//
// Example output:
//
//	Address 0: z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7
//	Address 1: z1qqga8s8rkypgsg5qg2g7rp68nqh3r4lkm54tta
//	...
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

// FindAddress searches for a specific address within the keystore by trying account
// indices sequentially until found or maxAccounts is reached.
//
// This is useful when you know an address belongs to this wallet but don't know which
// account index it uses. Common scenarios:
//   - Finding the account index for an address shown in a block explorer
//   - Locating which derivation path was used for a transaction
//   - Verifying an address belongs to this wallet
//
// Parameters:
//   - address: The Zenon address to search for
//   - maxAccounts: Maximum number of indices to check (0 uses DefaultMaxIndex)
//
// Returns FindResponse containing the account index and keypair, or ErrAddressNotFound.
//
// Example:
//
//	// Search for address in first 100 accounts
//	targetAddr := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")
//	result, err := keystore.FindAddress(targetAddr, 100)
//	if err == wallet.ErrAddressNotFound {
//	    fmt.Println("Address not found in this wallet")
//	} else if err != nil {
//	    log.Fatal(err)
//	} else {
//	    fmt.Printf("Found at index %d\n", result.Index)
//	    // Use result.KeyPair to sign transactions
//	}
//
// Performance note: This is a linear search. If maxAccounts is large, it may take time.
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

// Helper functions for JSON serialization using standard encoding/json
func serializeKeyStoreData(data map[string]interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func deserializeKeyStoreData(data []byte) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse keystore data: %w", err)
	}
	return result, nil
}
