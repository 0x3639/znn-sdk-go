package wallet

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// Review finding #rev5: Verify released the mutex before reading the returned
// key slice, so Destroy could zero it concurrently; getters also exposed the
// internal slices. Exercise Verify racing Destroy and confirm getters copy.
func TestKeyPairVerifyRacesDestroyAndGettersCopy(t *testing.T) {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i + 7)
	}

	// Getters must return copies: mutating the result must not affect the pair.
	kp, err := NewKeyPairFromSeed(seed)
	if err != nil {
		t.Fatalf("NewKeyPairFromSeed: %v", err)
	}
	pub, err := kp.GetPublicKey()
	if err != nil {
		t.Fatalf("GetPublicKey: %v", err)
	}
	pubCopy := append([]byte(nil), pub...)
	for i := range pub {
		pub[i] ^= 0xFF
	}
	pub2, _ := kp.GetPublicKey()
	if !bytes.Equal(pub2, pubCopy) {
		t.Error("GetPublicKey exposed internal slice (caller mutation leaked)")
	}

	priv := kp.GetPrivateKey()
	if len(priv) > 0 {
		privCopy := append([]byte(nil), priv...)
		priv[0] ^= 0xFF
		if !bytes.Equal(kp.GetPrivateKey(), privCopy) {
			t.Error("GetPrivateKey exposed internal slice (caller mutation leaked)")
		}
	}

	// NewKeyPair must copy caller-owned key bytes.
	raw := append([]byte(nil), kp.GetPrivateKey()...)
	kp2 := NewKeyPair(raw)
	before, _ := kp2.GetPublicKey()
	beforeCopy := append([]byte(nil), before...)
	for i := range raw {
		raw[i] ^= 0xFF
	}
	after, _ := kp2.GetPublicKey()
	if !bytes.Equal(after, beforeCopy) {
		t.Error("NewKeyPair retained caller-owned slice (mutation leaked)")
	}

	// Verify must not race Destroy: run many concurrent verifies and destroys.
	for iter := 0; iter < 50; iter++ {
		racer, err := NewKeyPairFromSeed(seed)
		if err != nil {
			t.Fatalf("NewKeyPairFromSeed: %v", err)
		}
		sig, err := racer.Sign([]byte("msg"))
		if err != nil {
			t.Fatalf("Sign: %v", err)
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); _, _ = racer.Verify(sig, []byte("msg")) }()
		go func() { defer wg.Done(); racer.Destroy() }()
		wg.Wait()
	}
}

// Regression test for finding #24: DerivePath accepted apostrophe-less
// components in [2^31, 2^32) and silently aliased them to hardened indices
// ("m/2147483648" derived the key of "m/0'"). Per the BIP32 path grammar an
// unhardened index must be < 2^31.
func TestDerivePathRejectsUnhardenedIndexAboveHardenedStart(t *testing.T) {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}

	hostilePaths := []string{
		"m/2147483648", // 2^31, aliases m/0'
		"m/4294967295", // 2^32-1, aliases m/2147483647'
	}
	for _, path := range hostilePaths {
		if _, err := DerivePath(path, seed); err == nil {
			t.Errorf("DerivePath(%q) accepted an out-of-range unhardened index", path)
		}
	}

	// Legitimate hardened paths must keep working.
	if _, err := DerivePath("m/44'/73404'/0'", seed); err != nil {
		t.Fatalf("DerivePath(valid hardened path) error = %v", err)
	}
}

// Regression test for finding #31: KeyPair lazily caches publicKey/address
// with no synchronization, racing Sign/GetPublicKey/GetAddress across
// goroutines (run with -race to detect).
func TestKeyPairConcurrentUseIsRaceFree(t *testing.T) {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	kp, err := NewKeyPairFromSeed(seed)
	if err != nil {
		t.Fatalf("NewKeyPairFromSeed: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := kp.GetPublicKey(); err != nil {
				t.Errorf("GetPublicKey: %v", err)
			}
			if _, err := kp.GetAddress(); err != nil {
				t.Errorf("GetAddress: %v", err)
			}
			if _, err := kp.Sign([]byte("message")); err != nil {
				t.Errorf("Sign: %v", err)
			}
			if ok, err := kp.Verify(mustSign(t, kp, []byte("message")), []byte("message")); err != nil || !ok {
				t.Errorf("Verify = %v, %v", ok, err)
			}
		}()
	}
	wg.Wait()
}

func mustSign(t *testing.T, kp *KeyPair, message []byte) []byte {
	t.Helper()
	sig, err := kp.Sign(message)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	return sig
}

// Regression test for finding #36: DeriveAddressesByRange sized its result
// allocation from the unbounded caller-supplied span, so a huge range
// panicked (makeslice: cap out of range) or ground through 2^31 derivations.
func TestDeriveAddressesByRangeRejectsExcessiveSpan(t *testing.T) {
	store, err := NewKeyStoreRandom()
	if err != nil {
		t.Fatalf("NewKeyStoreRandom: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DeriveAddressesByRange panicked on excessive span: %v", r)
		}
	}()

	if _, err = store.DeriveAddressesByRange(0, math.MaxInt); err == nil {
		t.Error("DeriveAddressesByRange accepted a span of math.MaxInt")
	}
	if _, err = store.DeriveAddressesByRange(0, DefaultMaxIndex+1); err == nil {
		t.Error("DeriveAddressesByRange accepted a span above DefaultMaxIndex")
	}

	addresses, err := store.DeriveAddressesByRange(0, 3)
	if err != nil {
		t.Fatalf("DeriveAddressesByRange(0, 3) error = %v", err)
	}
	if len(addresses) != 3 {
		t.Fatalf("DeriveAddressesByRange(0, 3) returned %d addresses", len(addresses))
	}
}

func TestFindAddress_RejectsHugeMaxAccounts(t *testing.T) {
	store, err := NewKeyStoreRandom()
	if err != nil {
		t.Fatal(err)
	}
	addr, _ := store.GetBaseAddress()
	if _, err := store.FindAddress(*addr, DefaultMaxIndex+1); err == nil {
		t.Error("FindAddress accepted maxAccounts above DefaultMaxIndex")
	}
}

// TestSaveKeyStore_DoesNotFollowSymlinkPlantedAfterCheck simulates the
// time-of-check/time-of-use race: a symlink at the target is replaced by the
// new keystore file rather than followed.
func TestSaveKeyStore_ReplacesSymlinkInsteadOfFollowing(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "victim")
	if err := os.WriteFile(outside, []byte("sensitive"), 0600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "wallet")
	if err := os.WriteFile(target, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	// Bypass resolveKeyStorePath to emulate the link being planted after the
	// Lstat check succeeded.
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, target); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(target, []byte("new keystore")); err != nil {
		t.Fatal(err)
	}
	victim, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(victim) != "sensitive" {
		t.Fatalf("file outside wallet directory was overwritten: %q", victim)
	}
	info, err := os.Lstat(target)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() {
		t.Fatalf("target is not a regular file after write: %v", info.Mode())
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("target perm = %o, want 0600", info.Mode().Perm())
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("temporary file left behind: %v", entries)
	}
}
