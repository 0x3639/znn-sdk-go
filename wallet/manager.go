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

// validateKeyStoreName ensures a caller-supplied wallet name resolves to a
// single file directly inside the managed wallet directory.
//
// Wallet names reach this package from CLIs and wallet backends, so they must
// be treated as untrusted input: a name containing path separators or ".."
// components would let filepath.Join escape the wallet directory and read or
// overwrite arbitrary files (CWE-22).
//
// Returns an error if the name is empty, contains a path separator, or is a
// relative path component ("." or "..").
func validateKeyStoreName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("invalid keystore name: %q", name)
	}
	if strings.ContainsAny(name, `/\`) || name != filepath.Base(name) {
		return fmt.Errorf("keystore name must not contain path separators: %q", name)
	}
	return nil
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

// resolveKeyStorePath validates a caller-supplied keystore name and returns
// its absolute path inside the managed wallet directory.
//
// It rejects names that escape the directory lexically (see
// validateKeyStoreName) and, when the target already exists, rejects anything
// that is not a regular file. This stops a planted symlink from redirecting a
// read or write to a file outside the wallet directory (CWE-59/CWE-22), which
// os.ReadFile/os.WriteFile would otherwise follow.
//
// A small time-of-check/time-of-use window remains between this Lstat and the
// subsequent open; it is far narrower than the unchecked path and closes the
// practical "planted symlink" vector.
func (m *KeyStoreManager) resolveKeyStorePath(name string) (string, error) {
	if err := validateKeyStoreName(name); err != nil {
		return "", err
	}
	filePath := filepath.Join(m.WalletPath, name)
	info, err := os.Lstat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// A new keystore is allowed; the parent directory is the managed one.
			return filePath, nil
		}
		return "", fmt.Errorf("failed to inspect keystore file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("keystore file %q is not a regular file", name)
	}
	return filePath, nil
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

	filePath, err := m.resolveKeyStorePath(name)
	if err != nil {
		return err
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

	// Write atomically (filePath was validated by resolveKeyStorePath above).
	if err := writeFileAtomic(filePath, jsonData); err != nil {
		return fmt.Errorf("failed to write keystore file: %w", err)
	}

	return nil
}

// writeFileAtomic writes data to path without ever following a symlink at
// path.
//
// The Argon2 work between resolveKeyStorePath's Lstat and the final write is
// long enough for a local attacker with write access to the wallet directory
// to replace the validated entry with a symlink (CWE-367). os.WriteFile would
// follow that link and truncate a file outside the wallet directory. Instead
// the data is written to a freshly created (O_CREATE|O_EXCL, 0600) temporary
// file in the same directory and then renamed over path. rename(2) replaces
// the directory entry itself and never dereferences a symlink at the target,
// so the worst an attacker can achieve is to have their planted link replaced
// by the new keystore file.
func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := func() {
		// #nosec G104 -- best-effort cleanup on an error path
		tmp.Close() //nolint:errcheck
		// #nosec G104 -- best-effort cleanup on an error path
		os.Remove(tmpName) //nolint:errcheck
	}
	// os.CreateTemp uses 0600 already; make it explicit in case of a
	// permissive umask on non-POSIX platforms.
	if err := tmp.Chmod(0600); err != nil {
		cleanup()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		// #nosec G104 -- best-effort cleanup on an error path
		os.Remove(tmpName) //nolint:errcheck
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		// #nosec G104 -- best-effort cleanup on an error path
		os.Remove(tmpName) //nolint:errcheck
		return err
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

	filePath, err := m.resolveKeyStorePath(keyStoreFile)
	if err != nil {
		return nil, err
	}

	// Read file
	// #nosec G304 - keyStoreFile is confined to the wallet directory and rejected if a symlink by resolveKeyStorePath
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
	if err := validateKeyStoreName(name); err != nil {
		return "", err
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

	// Filter for regular files only: skip directories, dotfiles, and symlinks
	// (a symlink could point outside the managed directory).
	var keystores []string
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}
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
	filePath, err := m.resolveKeyStorePath(keyStoreFile)
	if err != nil {
		return nil, err
	}

	// Read file
	// #nosec G304 - keyStoreFile is confined to the wallet directory and rejected if a symlink by resolveKeyStorePath
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
