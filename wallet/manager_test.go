package wallet

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// NewKeyStoreManager Tests
// =============================================================================

func TestNewKeyStoreManager_CreatesDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	walletPath := filepath.Join(tmpDir, "wallets")

	manager, err := NewKeyStoreManager(walletPath)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	if manager == nil {
		t.Error("Manager should not be nil")
	}

	// Check directory was created
	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		t.Error("Wallet directory should be created")
	}
}

func TestNewKeyStoreManager_ExistingDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory first
	walletPath := filepath.Join(tmpDir, "wallets")
	err = os.MkdirAll(walletPath, 0700)
	if err != nil {
		t.Fatalf("Failed to create wallet directory: %v", err)
	}

	manager, err := NewKeyStoreManager(walletPath)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	if manager.WalletPath != walletPath {
		t.Errorf("WalletPath = %s, want %s", manager.WalletPath, walletPath)
	}
}

// =============================================================================
// SaveKeyStore Tests
// =============================================================================

func TestSaveKeyStore_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	// Create test keystore
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	store, err := NewKeyStoreFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("NewKeyStoreFromMnemonic() error = %v", err)
	}

	// Save keystore
	err = manager.SaveKeyStore(store, "password123", "test-wallet")
	if err != nil {
		t.Fatalf("SaveKeyStore() error = %v", err)
	}

	// Check file exists
	filePath := filepath.Join(tmpDir, "test-wallet")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Keystore file should exist")
	}
}

func TestSaveKeyStore_NilKeyStore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	err = manager.SaveKeyStore(nil, "password", "test")
	if err == nil {
		t.Error("SaveKeyStore() should return error for nil keystore")
	}
}

func TestSaveKeyStore_EmptyPassword(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	store, _ := NewKeyStoreRandom()
	err = manager.SaveKeyStore(store, "", "test")
	if err == nil {
		t.Error("SaveKeyStore() should return error for empty password")
	}
}

func TestSaveKeyStore_EmptyName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	store, _ := NewKeyStoreRandom()
	err = manager.SaveKeyStore(store, "password", "")
	if err == nil {
		t.Error("SaveKeyStore() should return error for empty name")
	}
}

// =============================================================================
// ReadKeyStore Tests
// =============================================================================

func TestReadKeyStore_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	// Create and save keystore
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	store1, err := NewKeyStoreFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("NewKeyStoreFromMnemonic() error = %v", err)
	}

	err = manager.SaveKeyStore(store1, "password123", "test-wallet")
	if err != nil {
		t.Fatalf("SaveKeyStore() error = %v", err)
	}

	// Read keystore
	store2, err := manager.ReadKeyStore("password123", "test-wallet")
	if err != nil {
		t.Fatalf("ReadKeyStore() error = %v", err)
	}

	// Verify same keystore
	if store2.Mnemonic != mnemonic {
		t.Error("Read keystore should have same mnemonic")
	}
}

func TestReadKeyStore_WrongPassword(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	// Create and save keystore
	store, _ := NewKeyStoreRandom()
	err = manager.SaveKeyStore(store, "password123", "test-wallet")
	if err != nil {
		t.Fatalf("SaveKeyStore() error = %v", err)
	}

	// Try to read with wrong password
	_, err = manager.ReadKeyStore("wrongpassword", "test-wallet")
	if err == nil {
		t.Error("ReadKeyStore() should return error for wrong password")
	}
}

func TestReadKeyStore_FileNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	_, err = manager.ReadKeyStore("password", "nonexistent")
	if err == nil {
		t.Error("ReadKeyStore() should return error for nonexistent file")
	}
}

func TestReadKeyStore_EmptyPassword(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	_, err = manager.ReadKeyStore("", "test")
	if err == nil {
		t.Error("ReadKeyStore() should return error for empty password")
	}
}

// =============================================================================
// FindKeyStore Tests
// =============================================================================

func TestFindKeyStore_ExactMatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	// Create keystore
	store, _ := NewKeyStoreRandom()
	err = manager.SaveKeyStore(store, "password", "my-wallet")
	if err != nil {
		t.Fatalf("SaveKeyStore() error = %v", err)
	}

	// Find keystore
	found, err := manager.FindKeyStore("my-wallet")
	if err != nil {
		t.Fatalf("FindKeyStore() error = %v", err)
	}

	if found != "my-wallet" {
		t.Errorf("FindKeyStore() = %s, want my-wallet", found)
	}
}

func TestFindKeyStore_CaseInsensitive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	// Create keystore
	store, _ := NewKeyStoreRandom()
	err = manager.SaveKeyStore(store, "password", "MyWallet")
	if err != nil {
		t.Fatalf("SaveKeyStore() error = %v", err)
	}

	// Find with different case
	found, err := manager.FindKeyStore("mywallet")
	if err != nil {
		t.Fatalf("FindKeyStore() error = %v", err)
	}

	if strings.ToLower(found) != "mywallet" {
		t.Errorf("FindKeyStore() should find case-insensitive match")
	}
}

func TestFindKeyStore_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	_, err = manager.FindKeyStore("nonexistent")
	if !errors.Is(err, ErrKeystoreNotFound) {
		t.Errorf("FindKeyStore() error = %v, want ErrKeystoreNotFound", err)
	}
}

func TestFindKeyStore_EmptyName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	_, err = manager.FindKeyStore("")
	if err == nil {
		t.Error("FindKeyStore() should return error for empty name")
	}
}

// =============================================================================
// ListAllKeyStores Tests
// =============================================================================

func TestListAllKeyStores_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	keystores, err := manager.ListAllKeyStores()
	if err != nil {
		t.Fatalf("ListAllKeyStores() error = %v", err)
	}

	if len(keystores) != 0 {
		t.Errorf("ListAllKeyStores() = %d keystores, want 0", len(keystores))
	}
}

func TestListAllKeyStores_Multiple(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	// Create multiple keystores
	names := []string{"wallet1", "wallet2", "wallet3"}
	for _, name := range names {
		store, _ := NewKeyStoreRandom()
		saveErr := manager.SaveKeyStore(store, "password", name)
		if saveErr != nil {
			t.Fatalf("SaveKeyStore() error = %v", saveErr)
		}
	}

	keystores, err := manager.ListAllKeyStores()
	if err != nil {
		t.Fatalf("ListAllKeyStores() error = %v", err)
	}

	if len(keystores) != 3 {
		t.Errorf("ListAllKeyStores() = %d keystores, want 3", len(keystores))
	}

	// Check all names present
	for _, name := range names {
		found := false
		for _, ks := range keystores {
			if ks == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ListAllKeyStores() missing %s", name)
		}
	}
}

func TestListAllKeyStores_IgnoresHiddenFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	// Create regular keystore
	store, _ := NewKeyStoreRandom()
	err = manager.SaveKeyStore(store, "password", "wallet")
	if err != nil {
		t.Fatalf("SaveKeyStore() error = %v", err)
	}

	// Create hidden file
	hiddenPath := filepath.Join(tmpDir, ".hidden")
	err = os.WriteFile(hiddenPath, []byte("test"), 0600)
	if err != nil {
		t.Fatalf("Failed to create hidden file: %v", err)
	}

	keystores, err := manager.ListAllKeyStores()
	if err != nil {
		t.Fatalf("ListAllKeyStores() error = %v", err)
	}

	// Should only have the regular wallet
	if len(keystores) != 1 {
		t.Errorf("ListAllKeyStores() = %d keystores, want 1", len(keystores))
	}

	if keystores[0] != "wallet" {
		t.Errorf("ListAllKeyStores()[0] = %s, want wallet", keystores[0])
	}
}

// =============================================================================
// CreateNew Tests
// =============================================================================

func TestCreateNew_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	store, err := manager.CreateNew("password123", "new-wallet")
	if err != nil {
		t.Fatalf("CreateNew() error = %v", err)
	}

	if store == nil {
		t.Fatal("CreateNew() should return keystore")
	}

	if store.Mnemonic == "" {
		t.Error("CreateNew() should generate mnemonic")
	}

	// Check file exists
	filePath := filepath.Join(tmpDir, "new-wallet")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("CreateNew() should save keystore file")
	}
}

func TestCreateNew_EmptyName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	_, err = manager.CreateNew("password", "")
	if err == nil {
		t.Error("CreateNew() should return error for empty name")
	}
}

// =============================================================================
// CreateFromMnemonic Tests
// =============================================================================

func TestCreateFromMnemonic_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	store, err := manager.CreateFromMnemonic(mnemonic, "password123", "imported-wallet")
	if err != nil {
		t.Fatalf("CreateFromMnemonic() error = %v", err)
	}

	if store.Mnemonic != mnemonic {
		t.Error("CreateFromMnemonic() should preserve mnemonic")
	}

	// Check file exists
	filePath := filepath.Join(tmpDir, "imported-wallet")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("CreateFromMnemonic() should save keystore file")
	}
}

func TestCreateFromMnemonic_InvalidMnemonic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	_, err = manager.CreateFromMnemonic("invalid mnemonic", "password", "wallet")
	if err == nil {
		t.Error("CreateFromMnemonic() should return error for invalid mnemonic")
	}
}

func TestCreateFromMnemonic_EmptyName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	_, err = manager.CreateFromMnemonic(mnemonic, "password", "")
	if err == nil {
		t.Error("CreateFromMnemonic() should return error for empty name")
	}
}

// =============================================================================
// GetKeystoreInfo Tests
// =============================================================================

func TestGetKeystoreInfo_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	// Create keystore
	store, _ := NewKeyStoreRandom()
	err = manager.SaveKeyStore(store, "password", "test-wallet")
	if err != nil {
		t.Fatalf("SaveKeyStore() error = %v", err)
	}

	// Get info
	info, err := manager.GetKeystoreInfo("test-wallet")
	if err != nil {
		t.Fatalf("GetKeystoreInfo() error = %v", err)
	}

	// Check metadata fields
	if info[WalletTypeKey] != KeyStoreWalletType {
		t.Error("GetKeystoreInfo() should contain wallet type")
	}

	if _, ok := info[BaseAddressKey]; !ok {
		t.Error("GetKeystoreInfo() should contain base address")
	}
}

func TestGetKeystoreInfo_FileNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	_, err = manager.GetKeystoreInfo("nonexistent")
	if err == nil {
		t.Error("GetKeystoreInfo() should return error for nonexistent file")
	}
}

func TestGetKeystoreInfo_EmptyName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	_, err = manager.GetKeystoreInfo("")
	if err == nil {
		t.Error("GetKeystoreInfo() should return error for empty name")
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestKeyStoreManager_FullWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keystore-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager, err := NewKeyStoreManager(tmpDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	// 1. Create new keystore
	store1, err := manager.CreateNew("password123", "wallet1")
	if err != nil {
		t.Fatalf("CreateNew() error = %v", err)
	}

	// 2. Import from mnemonic
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	store2, err := manager.CreateFromMnemonic(mnemonic, "password456", "wallet2")
	if err != nil {
		t.Fatalf("CreateFromMnemonic() error = %v", err)
	}

	// 3. List all keystores
	keystores, err := manager.ListAllKeyStores()
	if err != nil {
		t.Fatalf("ListAllKeyStores() error = %v", err)
	}

	if len(keystores) != 2 {
		t.Errorf("ListAllKeyStores() = %d keystores, want 2", len(keystores))
	}

	// 4. Find keystore
	found, err := manager.FindKeyStore("wallet1")
	if err != nil {
		t.Fatalf("FindKeyStore() error = %v", err)
	}

	if found != "wallet1" {
		t.Error("FindKeyStore() should find wallet1")
	}

	// 5. Get info without decrypting
	info, err := manager.GetKeystoreInfo("wallet2")
	if err != nil {
		t.Fatalf("GetKeystoreInfo() error = %v", err)
	}

	if info[WalletTypeKey] != KeyStoreWalletType {
		t.Error("GetKeystoreInfo() should return correct wallet type")
	}

	// 6. Read and decrypt keystores
	readStore1, err := manager.ReadKeyStore("password123", "wallet1")
	if err != nil {
		t.Fatalf("ReadKeyStore() error = %v", err)
	}

	if readStore1.Mnemonic != store1.Mnemonic {
		t.Error("Read keystore should match original")
	}

	readStore2, err := manager.ReadKeyStore("password456", "wallet2")
	if err != nil {
		t.Fatalf("ReadKeyStore() error = %v", err)
	}

	if readStore2.Mnemonic != store2.Mnemonic {
		t.Error("Read keystore should match original")
	}
}
