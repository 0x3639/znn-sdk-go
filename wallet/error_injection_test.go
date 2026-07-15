package wallet

import (
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/0x3639/znn-sdk-go/crypto"
	"github.com/zenon-network/go-zenon/common/types"
)

func validMalformedTestEnvelope() *EncryptedFile {
	return &EncryptedFile{
		Version: 1,
		Crypto: &CryptoParams{
			Argon2Params: &Argon2Params{Salt: "0x" + strings.Repeat("00", 16)},
			CipherData:   "0x" + strings.Repeat("00", 16),
			CipherName:   "aes-256-gcm",
			Kdf:          "argon2.IDKey",
			Nonce:        "0x" + strings.Repeat("00", 12),
		},
	}
}

func TestEncryptedFileRejectsMalformedEnvelopeFields(t *testing.T) {
	var nilFile *EncryptedFile
	if _, err := nilFile.Decrypt("password"); err == nil {
		t.Fatal("nil encrypted file was accepted")
	}

	tests := []struct {
		name   string
		mutate func(*EncryptedFile)
		want   string
	}{
		{"version", func(file *EncryptedFile) { file.Version = 2 }, "unsupported encrypted file version"},
		{"crypto", func(file *EncryptedFile) { file.Crypto = nil }, "missing crypto parameters"},
		{"argon2", func(file *EncryptedFile) { file.Crypto.Argon2Params = nil }, "missing crypto parameters"},
		{"cipher", func(file *EncryptedFile) { file.Crypto.CipherName = "aes-cbc" }, "unsupported cipher"},
		{"kdf", func(file *EncryptedFile) { file.Crypto.Kdf = "scrypt" }, "unsupported key derivation"},
		{"salt hex", func(file *EncryptedFile) { file.Crypto.Argon2Params.Salt = "0xzz" }, "invalid Argon2 salt"},
		{"salt length", func(file *EncryptedFile) { file.Crypto.Argon2Params.Salt = "0x00" }, "invalid Argon2 salt length"},
		{"nonce hex", func(file *EncryptedFile) { file.Crypto.Nonce = "0xzz" }, "invalid AES-GCM nonce"},
		{"nonce length", func(file *EncryptedFile) { file.Crypto.Nonce = "0x00" }, "invalid AES-GCM nonce length"},
		{"cipher data", func(file *EncryptedFile) { file.Crypto.CipherData = "0xzz" }, "invalid cipher data"},
		{"key length", func(file *EncryptedFile) {
			file.Crypto.Argon2Params.TimeCost = 1
			file.Crypto.Argon2Params.MemoryCost = 8
			file.Crypto.Argon2Params.HashLength = 16
			file.Crypto.Argon2Params.Parallelism = 1
		}, "invalid AES-256 key length"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := validMalformedTestEnvelope()
			test.mutate(file)
			if _, err := file.Decrypt("password"); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Decrypt error = %v, want substring %q", err, test.want)
			}
		})
	}

	if _, err := decryptAESGCM([]byte{1}, make([]byte, 12), make([]byte, 16)); err == nil {
		t.Fatal("AES-GCM accepted an invalid key length")
	}
}

func TestEncryptedFileNeedsUpgradeVariants(t *testing.T) {
	if !(*EncryptedFile)(nil).NeedsUpgrade() {
		t.Fatal("nil file does not require upgrade")
	}
	file := validMalformedTestEnvelope()
	defaults := crypto.DefaultArgon2Parameters()
	file.Crypto.Argon2Params.TimeCost = defaults.Iterations
	file.Crypto.Argon2Params.MemoryCost = defaults.Memory
	file.Crypto.Argon2Params.HashLength = defaults.KeyLength
	file.Crypto.Argon2Params.Parallelism = defaults.Parallelism
	if file.NeedsUpgrade(defaults) {
		t.Fatal("matching explicit parameters require upgrade")
	}
	if !file.NeedsUpgrade(defaults, defaults) {
		t.Fatal("multiple target configurations were accepted")
	}
	custom := defaults
	custom.Iterations++
	if !file.NeedsUpgrade(custom) {
		t.Fatal("different target parameters do not require upgrade")
	}
	file.Crypto.Argon2Params.Parallelism = 0
	if !file.NeedsUpgrade(defaults) {
		t.Fatal("incomplete parameters do not require upgrade")
	}
}

func TestLegacyKeyStorePayloadVariantsAndErrors(t *testing.T) {
	entropy := strings.Repeat("11", 16)
	fromEntropy, err := keyStoreFromLegacyPlaintext([]byte(`{"entropy":"` + entropy + `"}`))
	if err != nil || len(fromEntropy.Entropy) != 16 {
		t.Fatalf("entropy payload = %#v, %v", fromEntropy, err)
	}
	seed := strings.Repeat("22", 32)
	fromSeed, err := keyStoreFromLegacyPlaintext([]byte(`{"seed":"` + seed + `"}`))
	if err != nil || hex.EncodeToString(fromSeed.Seed) != seed {
		t.Fatalf("seed payload = %#v, %v", fromSeed, err)
	}

	for _, payload := range []string{
		`{`,
		`{}`,
		`{"mnemonic":"invalid mnemonic"}`,
		`{"entropy":"zz"}`,
		`{"entropy":"00"}`,
		`{"seed":"zz"}`,
	} {
		if _, err := keyStoreFromLegacyPlaintext([]byte(payload)); err == nil {
			t.Errorf("legacy payload %q was accepted", payload)
		}
	}
	if _, err := deserializeKeyStoreData([]byte(`{`)); err == nil {
		t.Fatal("malformed keystore JSON was accepted")
	}
}

func TestKeyStoreErrorPropagationFromMissingSeedAndEntropy(t *testing.T) {
	store := &KeyStore{}
	if _, err := store.GetKeyPair(0); err == nil {
		t.Fatal("GetKeyPair accepted a missing seed")
	}
	if _, err := store.DeriveAddressesByRange(0, 1); err == nil {
		t.Fatal("DeriveAddressesByRange accepted a missing seed")
	}
	if _, err := store.FindAddress(types.ZeroAddress, 1); err == nil {
		t.Fatal("FindAddress accepted a missing seed")
	}
	if _, err := store.GetBaseAddress(); err == nil {
		t.Fatal("GetBaseAddress accepted a missing seed")
	}
	if _, err := store.ToEncryptedFile("password", nil); !errors.Is(err, ErrInvalidKeyStore) {
		t.Fatalf("ToEncryptedFile error = %v", err)
	}
	if _, err := (*KeyStore)(nil).ToEncryptedFile("password", nil); !errors.Is(err, ErrInvalidKeyStore) {
		t.Fatalf("nil ToEncryptedFile error = %v", err)
	}
}

func TestFromEncryptedFileRejectsMissingMetadataAndInvalidPlaintext(t *testing.T) {
	file, err := Encrypt(make([]byte, 16), "password", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, fromErr := FromEncryptedFile(file, "password"); !errors.Is(fromErr, ErrInvalidKeyStore) {
		t.Fatalf("missing metadata error = %v", fromErr)
	}

	invalid, err := Encrypt([]byte("not a keystore"), "password", map[string]interface{}{
		BaseAddressKey: types.ZeroAddress.String(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := FromEncryptedFile(invalid, "password"); !errors.Is(err, ErrInvalidKeyStore) {
		t.Fatalf("invalid plaintext error = %v", err)
	}
}
