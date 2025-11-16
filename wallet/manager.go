package wallet

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// KeyStoreManager manages keystore files in a directory
type KeyStoreManager struct {
	WalletPath string
}

// NewKeyStoreManager creates a new keystore manager for managing encrypted wallet files
// in the specified directory.
//
// The manager handles:
//   - Creating new wallets with random mnemonics
//   - Importing wallets from existing mnemonics
//   - Saving encrypted keystore files
//   - Loading encrypted keystore files
//   - Listing all wallets in the directory
//
// Parameters:
//   - walletPath: Directory path where keystore files will be stored
//
// The directory will be created with 0700 permissions if it doesn't exist, ensuring
// only the owner can read/write wallet files.
//
// Returns a KeyStoreManager instance or an error if directory creation fails.
//
// Example:
//
//	manager, err := wallet.NewKeyStoreManager("./my-wallets")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create a new wallet
//	keystore, _ := manager.CreateNew("password123", "main-wallet")
//	fmt.Println("Mnemonic:", keystore.Mnemonic)
func NewKeyStoreManager(walletPath string) (*KeyStoreManager, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(walletPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create wallet directory: %w", err)
	}

	return &KeyStoreManager{
		WalletPath: walletPath,
	}, nil
}

// SaveKeyStore encrypts a keystore and saves it to a file in the managed directory.
//
// The keystore is encrypted using Argon2 key derivation with the provided password.
// The file is saved with 0600 permissions (readable/writable only by owner).
//
// Parameters:
//   - store: KeyStore instance to save
//   - password: Passphrase for encryption (must be non-empty)
//   - name: Filename for the keystore
//
// Returns an error if encryption or file writing fails.
//
// Example:
//
//	// Create keystore in memory
//	keystore, _ := wallet.NewKeyStoreRandom()
//
//	// Save to file
//	manager, _ := wallet.NewKeyStoreManager("./wallets")
//	err := manager.SaveKeyStore(keystore, "secure-password", "backup-wallet")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (m *KeyStoreManager) SaveKeyStore(store *KeyStore, password, name string) error {
	if store == nil {
		return fmt.Errorf("keystore cannot be nil")
	}

	// Validate password strength
	if err := ValidatePassword(password); err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}

	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	// Get base address for metadata
	baseAddr, err := store.GetBaseAddress()
	if err != nil {
		return fmt.Errorf("failed to get base address: %w", err)
	}

	// Create metadata
	metadata := map[string]interface{}{
		BaseAddressKey: baseAddr.String(),
		WalletTypeKey:  KeyStoreWalletType,
		"name":         name,
	}

	// Encrypt keystore
	ef, err := store.ToEncryptedFile(password, metadata)
	if err != nil {
		return fmt.Errorf("failed to encrypt keystore: %w", err)
	}

	// Serialize to JSON
	jsonData, err := ef.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize keystore: %w", err)
	}

	// Construct file path
	filePath := filepath.Join(m.WalletPath, name)

	// Write to file
	if err := os.WriteFile(filePath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write keystore file: %w", err)
	}

	return nil
}

// ReadKeyStore loads and decrypts an existing keystore file from the managed directory.
//
// This method:
//  1. Reads the encrypted keystore file
//  2. Parses the JSON structure
//  3. Decrypts using the provided password
//  4. Returns the KeyStore ready for use
//
// Parameters:
//   - password: Passphrase used when the keystore was created/saved
//   - keyStoreFile: Filename of the keystore (not full path, just the name)
//
// Returns the decrypted KeyStore or an error if:
//   - File doesn't exist
//   - Password is incorrect
//   - File is corrupted
//
// Example:
//
//	manager, _ := wallet.NewKeyStoreManager("./wallets")
//	keystore, err := manager.ReadKeyStore("my-password", "main-wallet")
//	if err != nil {
//	    log.Fatal("Failed to load wallet:", err)
//	}
//
//	// Use the keystore
//	keypair, _ := keystore.GetKeyPair(0)
//	address, _ := keypair.GetAddress()
//	fmt.Println("Address:", address)
func (m *KeyStoreManager) ReadKeyStore(password string, keyStoreFile string) (*KeyStore, error) {
	// Note: When reading, we don't validate password strength since the keystore
	// may have been created before validation was added. We only check non-empty.
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	if keyStoreFile == "" {
		return nil, fmt.Errorf("keystore file cannot be empty")
	}

	// Construct file path
	filePath := filepath.Join(m.WalletPath, keyStoreFile)

	// Read file
	// #nosec G304 - filePath is constructed from controlled wallet directory
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore file: %w", err)
	}

	// Parse JSON
	ef, err := FromJSON(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse keystore file: %w", err)
	}

	// Decrypt
	store, err := FromEncryptedFile(ef, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt keystore: %w", err)
	}

	return store, nil
}

// FindKeyStore searches for a keystore file by name
// Returns the filename if found, empty string if not found
func (m *KeyStoreManager) FindKeyStore(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("name cannot be empty")
	}

	// Try exact match first
	filePath := filepath.Join(m.WalletPath, name)
	if _, err := os.Stat(filePath); err == nil {
		return name, nil
	}

	// Try case-insensitive search
	files, err := m.ListAllKeyStores()
	if err != nil {
		return "", err
	}

	lowerName := strings.ToLower(name)
	for _, file := range files {
		if strings.ToLower(file) == lowerName {
			return file, nil
		}
	}

	return "", ErrKeystoreNotFound
}

// ListAllKeyStores returns all keystore files in the directory
func (m *KeyStoreManager) ListAllKeyStores() ([]string, error) {
	// Read directory
	entries, err := os.ReadDir(m.WalletPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet directory: %w", err)
	}

	// Filter for regular files (no directories)
	var keystores []string
	for _, entry := range entries {
		if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			keystores = append(keystores, entry.Name())
		}
	}

	return keystores, nil
}

// CreateNew generates a new wallet with a random BIP39 mnemonic and saves it as an
// encrypted keystore file.
//
// This is the primary method for creating new Zenon wallets. It:
//  1. Generates a cryptographically secure 24-word BIP39 mnemonic
//  2. Derives the master seed from the mnemonic
//  3. Encrypts the keystore with the provided passphrase using Argon2
//  4. Saves the encrypted keystore to a file
//
// Parameters:
//   - passphrase: Password to encrypt the keystore (must be non-empty)
//   - name: Filename for the keystore (e.g., "main-wallet")
//
// Returns the created KeyStore containing the mnemonic and seed, or an error.
//
// IMPORTANT: The mnemonic must be securely backed up. It's the only way to recover
// the wallet if the keystore file is lost.
//
// Example:
//
//	manager, _ := wallet.NewKeyStoreManager("./wallets")
//	keystore, err := manager.CreateNew("secure-password", "my-wallet")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// IMPORTANT: Back up this mnemonic securely!
//	fmt.Println("Mnemonic:", keystore.Mnemonic)
//	fmt.Println("Base address:", keystore.GetBaseAddress())
func (m *KeyStoreManager) CreateNew(passphrase, name string) (*KeyStore, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	// Generate random keystore
	store, err := NewKeyStoreRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate keystore: %w", err)
	}

	// Save to file
	if err := m.SaveKeyStore(store, passphrase, name); err != nil {
		return nil, err
	}

	return store, nil
}

// CreateFromMnemonic imports a wallet from an existing BIP39 mnemonic phrase and
// saves it as an encrypted keystore file.
//
// Use this method to:
//   - Restore a wallet from a backup mnemonic
//   - Import a wallet from another device
//   - Migrate from another Zenon wallet application
//
// The mnemonic must be a valid BIP39 phrase (12 or 24 words). The same mnemonic
// will always generate the same addresses.
//
// Parameters:
//   - mnemonic: Valid BIP39 mnemonic phrase (space-separated words)
//   - passphrase: Password to encrypt the keystore (can be different from original)
//   - name: Filename for the keystore
//
// Returns the imported KeyStore or an error if the mnemonic is invalid.
//
// Example:
//
//	manager, _ := wallet.NewKeyStoreManager("./wallets")
//	mnemonic := "route become dream access impulse price inform obtain engage ski believe awful"
//	keystore, err := manager.CreateFromMnemonic(mnemonic, "new-password", "imported-wallet")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Verify it matches expected address
//	address, _ := keystore.GetBaseAddress()
//	fmt.Println("Restored address:", address)
func (m *KeyStoreManager) CreateFromMnemonic(mnemonic, passphrase, name string) (*KeyStore, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	// Create from mnemonic
	store, err := NewKeyStoreFromMnemonic(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to create keystore from mnemonic: %w", err)
	}

	// Save to file
	if err := m.SaveKeyStore(store, passphrase, name); err != nil {
		return nil, err
	}

	return store, nil
}

// GetKeystoreInfo reads metadata from a keystore file without decrypting
func (m *KeyStoreManager) GetKeystoreInfo(keyStoreFile string) (map[string]interface{}, error) {
	if keyStoreFile == "" {
		return nil, fmt.Errorf("keystore file cannot be empty")
	}

	// Construct file path
	filePath := filepath.Join(m.WalletPath, keyStoreFile)

	// Read file
	// #nosec G304 - filePath is constructed from controlled wallet directory
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore file: %w", err)
	}

	// Parse JSON
	ef, err := FromJSON(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse keystore file: %w", err)
	}

	return ef.Metadata, nil
}
