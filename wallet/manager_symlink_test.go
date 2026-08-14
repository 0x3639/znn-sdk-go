package wallet

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Review finding #rev6: a safe-looking filename inside the wallet directory may
// be a symlink to a file outside it, and os.WriteFile/os.ReadFile follow it.
// Save/Read/GetInfo must reject symlinks, and ListAllKeyStores must exclude
// them.
func TestKeyStoreManager_RejectsSymlinkedEntries(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on windows")
	}
	root := t.TempDir()
	walletDir := filepath.Join(root, "wallets")
	manager, err := NewKeyStoreManager(walletDir)
	if err != nil {
		t.Fatalf("NewKeyStoreManager: %v", err)
	}

	// A victim keystore outside the managed directory.
	other, err := NewKeyStoreManager(filepath.Join(root, "other"))
	if err != nil {
		t.Fatalf("NewKeyStoreManager: %v", err)
	}
	if _, err := other.CreateNew("password123", "victim"); err != nil {
		t.Fatalf("CreateNew: %v", err)
	}
	victimPath := filepath.Join(root, "other", "victim")

	// Plant a symlink inside the wallet dir pointing at the victim.
	link := filepath.Join(walletDir, "innocent")
	if err := os.Symlink(victimPath, link); err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	if _, err := manager.ReadKeyStore("password123", "innocent"); err == nil {
		t.Error("ReadKeyStore followed a symlink out of the wallet directory")
	}
	if _, err := manager.GetKeystoreInfo("innocent"); err == nil {
		t.Error("GetKeystoreInfo followed a symlink out of the wallet directory")
	}

	// A symlink target for a write: overwriting must not clobber the victim.
	store, err := NewKeyStoreRandom()
	if err != nil {
		t.Fatalf("NewKeyStoreRandom: %v", err)
	}
	if err := manager.SaveKeyStore(store, "password123", "innocent"); err == nil {
		t.Error("SaveKeyStore followed a symlink out of the wallet directory")
	}

	// ListAllKeyStores must not include the symlink.
	names, err := manager.ListAllKeyStores()
	if err != nil {
		t.Fatalf("ListAllKeyStores: %v", err)
	}
	for _, name := range names {
		if name == "innocent" {
			t.Error("ListAllKeyStores included a symlinked entry")
		}
	}
}
