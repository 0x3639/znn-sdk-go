package wallet

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// KeyStoreManager manages keystore files in a directory
type KeyStoreManager struct {
	WalletPath string
}

// NewKeyStoreManager creates a new keystore manager for the given directory
func NewKeyStoreManager(walletPath string) (*KeyStoreManager, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(walletPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create wallet directory: %w", err)
	}

	return &KeyStoreManager{
		WalletPath: walletPath,
	}, nil
}

// SaveKeyStore encrypts and saves a keystore to a file
func (m *KeyStoreManager) SaveKeyStore(store *KeyStore, password, name string) error {
	if store == nil {
		return fmt.Errorf("keystore cannot be nil")
	}

	if password == "" {
		return fmt.Errorf("password cannot be empty")
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
	if err := ioutil.WriteFile(filePath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write keystore file: %w", err)
	}

	return nil
}

// ReadKeyStore reads and decrypts a keystore from a file
func (m *KeyStoreManager) ReadKeyStore(password string, keyStoreFile string) (*KeyStore, error) {
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	if keyStoreFile == "" {
		return nil, fmt.Errorf("keystore file cannot be empty")
	}

	// Construct file path
	filePath := filepath.Join(m.WalletPath, keyStoreFile)

	// Read file
	jsonData, err := ioutil.ReadFile(filePath)
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
	files, err := ioutil.ReadDir(m.WalletPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet directory: %w", err)
	}

	// Filter for regular files (no directories)
	var keystores []string
	for _, file := range files {
		if file.Mode().IsRegular() && !strings.HasPrefix(file.Name(), ".") {
			keystores = append(keystores, file.Name())
		}
	}

	return keystores, nil
}

// CreateNew generates a new random keystore and saves it
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

// CreateFromMnemonic creates a keystore from an existing mnemonic and saves it
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
	jsonData, err := ioutil.ReadFile(filePath)
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
