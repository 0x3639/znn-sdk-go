package wallet

import (
	"os"
	"path/filepath"
	"testing"
)

// Regression tests for path traversal (findings #29, #30): caller-supplied
// wallet names must not be able to escape the managed wallet directory in
// any KeyStoreManager entry point.

func TestSaveKeyStore_NameCannotEscapeWalletDir(t *testing.T) {
	root := t.TempDir()
	walletDir := filepath.Join(root, "wallets")

	manager, err := NewKeyStoreManager(walletDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	store, err := NewKeyStoreRandom()
	if err != nil {
		t.Fatalf("NewKeyStoreRandom() error = %v", err)
	}

	hostileNames := []string{
		"../escaped-wallet",
		"sub/dir-wallet",
		"..",
		".",
	}
	for _, name := range hostileNames {
		if err := manager.SaveKeyStore(store, "password123", name); err == nil {
			t.Errorf("SaveKeyStore(%q) accepted a hostile wallet name", name)
		}
	}

	outside := filepath.Join(root, "escaped-wallet")
	if _, statErr := os.Stat(outside); statErr == nil {
		t.Fatalf("SaveKeyStore wrote outside the wallet directory: %s exists", outside)
	}
}

func TestReadKeyStore_NameCannotEscapeWalletDir(t *testing.T) {
	root := t.TempDir()
	otherDir := filepath.Join(root, "other")
	walletDir := filepath.Join(root, "wallets")

	// Keystore that lives OUTSIDE the managed wallet directory.
	other, err := NewKeyStoreManager(otherDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}
	if _, err := other.CreateNew("password123", "victim-wallet"); err != nil {
		t.Fatalf("CreateNew() error = %v", err)
	}

	manager, err := NewKeyStoreManager(walletDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	if store, err := manager.ReadKeyStore("password123", "../other/victim-wallet"); err == nil {
		t.Fatalf("ReadKeyStore escaped the wallet directory and decrypted a foreign keystore (mnemonic length %d)", len(store.Mnemonic))
	}
}

func TestGetKeystoreInfo_NameCannotEscapeWalletDir(t *testing.T) {
	root := t.TempDir()
	otherDir := filepath.Join(root, "other")
	walletDir := filepath.Join(root, "wallets")

	other, err := NewKeyStoreManager(otherDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}
	if _, err := other.CreateNew("password123", "victim-wallet"); err != nil {
		t.Fatalf("CreateNew() error = %v", err)
	}

	manager, err := NewKeyStoreManager(walletDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	if info, err := manager.GetKeystoreInfo("../other/victim-wallet"); err == nil {
		t.Fatalf("GetKeystoreInfo escaped the wallet directory and disclosed foreign metadata: %v", info)
	}
}

func TestFindKeyStore_NameCannotEscapeWalletDir(t *testing.T) {
	root := t.TempDir()
	otherDir := filepath.Join(root, "other")
	walletDir := filepath.Join(root, "wallets")

	other, err := NewKeyStoreManager(otherDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}
	if _, err := other.CreateNew("password123", "victim-wallet"); err != nil {
		t.Fatalf("CreateNew() error = %v", err)
	}

	manager, err := NewKeyStoreManager(walletDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	if name, err := manager.FindKeyStore("../other/victim-wallet"); err == nil {
		t.Fatalf("FindKeyStore acted as a file-existence oracle outside the wallet directory: %q", name)
	}
}

// A plain wallet name must keep working end to end.
func TestKeyStoreManager_PlainNamesStillWork(t *testing.T) {
	manager, err := NewKeyStoreManager(t.TempDir())
	if err != nil {
		t.Fatalf("NewKeyStoreManager() error = %v", err)
	}

	created, err := manager.CreateNew("password123", "main-wallet")
	if err != nil {
		t.Fatalf("CreateNew() error = %v", err)
	}

	loaded, err := manager.ReadKeyStore("password123", "main-wallet")
	if err != nil {
		t.Fatalf("ReadKeyStore() error = %v", err)
	}
	if loaded.Mnemonic != created.Mnemonic {
		t.Fatal("round-tripped keystore does not match")
	}

	if _, err := manager.GetKeystoreInfo("main-wallet"); err != nil {
		t.Fatalf("GetKeystoreInfo() error = %v", err)
	}
	if name, err := manager.FindKeyStore("main-wallet"); err != nil || name != "main-wallet" {
		t.Fatalf("FindKeyStore() = %q, %v", name, err)
	}
}
