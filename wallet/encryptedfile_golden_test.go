package wallet

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Golden vectors imported verbatim from the Dart reference SDK at
// reference/znn_sdk_dart-master/test/wallet/encryptedfile_metadata_test.dart.
//
// These pin the on-disk wallet file format. A keystore written by the Dart
// SDK must be loadable by the Go SDK and vice versa, which is only true if
// both sides agree on every JSON field name and shape.

const dartKeyStoreNoMetadata = `{
    "crypto": {
        "argon2Params": {
          "salt": "0x5b85100f186953332faddeaf2b6d68de"
        },
        "cipherData": "0xcfe2e1aa498229aaf78ab9634592a19c1f45ad5178ed4d530fdeef8fe528e4b78955a389e0d1e832fd0c21346b3a45cd",
        "cipherName": "aes-256-gcm",
        "kdf": "argon2.IDKey",
        "nonce": "0x86d6b27ba67a72238af12958"
      },
    "timestamp": 1707418422,
    "version": 1
  }`

const dartKeyStoreWithMetadata = `{
    "baseAddress": "z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7",
    "walletType": "keystore",
    "crypto": {
        "argon2Params": {
          "salt": "0x5b85100f186953332faddeaf2b6d68de"
        },
        "cipherData": "0xcfe2e1aa498229aaf78ab9634592a19c1f45ad5178ed4d530fdeef8fe528e4b78955a389e0d1e832fd0c21346b3a45cd",
        "cipherName": "aes-256-gcm",
        "kdf": "argon2.IDKey",
        "nonce": "0x86d6b27ba67a72238af12958"
      },
    "timestamp": 1707418422,
    "version": 1
  }`

// jsonEqual compares two JSON byte slices semantically (whitespace/order independent).
func jsonEqual(t *testing.T, a, b []byte) bool {
	t.Helper()
	var av, bv interface{}
	if err := json.Unmarshal(a, &av); err != nil {
		t.Fatalf("invalid JSON a: %v", err)
	}
	if err := json.Unmarshal(b, &bv); err != nil {
		t.Fatalf("invalid JSON b: %v", err)
	}
	return reflect.DeepEqual(av, bv)
}

func TestEncryptedFile_DartGoldenNoMetadata(t *testing.T) {
	ef, err := FromJSON([]byte(dartKeyStoreNoMetadata))
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if ef.Version != 1 {
		t.Errorf("version: got %d, want 1", ef.Version)
	}
	if ef.Timestamp != 1707418422 {
		t.Errorf("timestamp: got %d, want 1707418422", ef.Timestamp)
	}
	if ef.Crypto == nil {
		t.Fatal("crypto field missing")
	}
	if ef.Crypto.CipherName != "aes-256-gcm" {
		t.Errorf("cipherName: got %q, want %q", ef.Crypto.CipherName, "aes-256-gcm")
	}
	if ef.Crypto.Kdf != "argon2.IDKey" {
		t.Errorf("kdf: got %q, want %q", ef.Crypto.Kdf, "argon2.IDKey")
	}
	if ef.Crypto.Argon2Params.Salt != "0x5b85100f186953332faddeaf2b6d68de" {
		t.Errorf("salt mismatch: got %q", ef.Crypto.Argon2Params.Salt)
	}
	if ef.Crypto.Nonce != "0x86d6b27ba67a72238af12958" {
		t.Errorf("nonce mismatch: got %q", ef.Crypto.Nonce)
	}
	if ef.Crypto.CipherData != "0xcfe2e1aa498229aaf78ab9634592a19c1f45ad5178ed4d530fdeef8fe528e4b78955a389e0d1e832fd0c21346b3a45cd" {
		t.Errorf("cipherData mismatch")
	}

	// Dart asserts metadata is null when absent. Go's FromJSON always
	// allocates an empty map; treat empty as semantically equivalent to
	// "no metadata". This is the boundary we care about — that no spurious
	// keys snuck in.
	if len(ef.Metadata) != 0 {
		t.Errorf("expected no metadata, got %v", ef.Metadata)
	}

	// Round-trip
	out, err := ef.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if !jsonEqual(t, out, []byte(dartKeyStoreNoMetadata)) {
		t.Errorf("round-trip mismatch\n  got:  %s\n  want: %s", out, dartKeyStoreNoMetadata)
	}
}

func TestEncryptedFile_DartGoldenWithMetadata(t *testing.T) {
	ef, err := FromJSON([]byte(dartKeyStoreWithMetadata))
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if got, want := ef.Metadata["baseAddress"], "z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7"; got != want {
		t.Errorf("metadata baseAddress: got %v, want %s", got, want)
	}
	if got, want := ef.Metadata["walletType"], "keystore"; got != want {
		t.Errorf("metadata walletType: got %v, want %s", got, want)
	}
	if len(ef.Metadata) != 2 {
		t.Errorf("expected exactly 2 metadata keys (baseAddress, walletType), got %d: %v", len(ef.Metadata), ef.Metadata)
	}

	// Round-trip
	out, err := ef.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if !jsonEqual(t, out, []byte(dartKeyStoreWithMetadata)) {
		t.Errorf("round-trip mismatch\n  got:  %s\n  want: %s", out, dartKeyStoreWithMetadata)
	}
}
