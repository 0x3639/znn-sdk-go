package wallet

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	sdkcrypto "github.com/0x3639/znn-sdk-go/crypto"
)

func TestFromEncryptedFileLegacyRawEntropyVectors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		keyFileJSON string
		entropyHex  string
		baseAddress string
	}{
		{
			name: "256-bit",
			keyFileJSON: `{
                "baseAddress":"z1qq9n7fpaqd8lpcljandzmx4xtku9w4ftwyg0mq",
                "crypto":{"argon2Params":{"salt":"0xab4801d422d25662820f75b53878bf08"},"cipherData":"0x652514c94526bbca6d82f5c663d047803b18819ef7be0dd6bc45822343b70a46d7ffda6730ccd8a26f636bacfcb318d3","cipherName":"aes-256-gcm","kdf":"argon2.IDKey","nonce":"0xf52d55466f05414a5a9f528b"},
                "timestamp":1639039880,"version":1
            }`,
			entropyHex:  "00e089c2d43064b3462ce24fc09099fe9fd2cf3657b6335462972baa911d31fc",
			baseAddress: "z1qq9n7fpaqd8lpcljandzmx4xtku9w4ftwyg0mq",
		},
		{
			name: "128-bit",
			keyFileJSON: `{
                "baseAddress":"z1qrf825tea0hha086vjnn4dhpl5wsdcesktxh5x",
                "crypto":{"argon2Params":{"salt":"0x4cb0009a61148aa2874dbb8450c2cfca"},"cipherData":"0x142b5bcfdac54ad3a6a2cfb627f30f80a4080e02500cab75a9b79b3ccf2752ef","cipherName":"aes-256-gcm","kdf":"argon2.IDKey","nonce":"0xa31fb4d6027c482fd9d85c1d"},
                "timestamp":1639637010,"version":1
            }`,
			entropyHex:  "bbefd88e1ff3f673d24da98b51f04ee7",
			baseAddress: "z1qrf825tea0hha086vjnn4dhpl5wsdcesktxh5x",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, err := FromJSON([]byte(test.keyFileJSON))
			if err != nil {
				t.Fatalf("FromJSON() error = %v", err)
			}
			if !file.NeedsUpgrade() {
				t.Fatal("NeedsUpgrade() = false, want true for a legacy salt-only file")
			}

			store, err := FromEncryptedFile(file, "password")
			if err != nil {
				t.Fatalf("FromEncryptedFile() error = %v", err)
			}
			if got := hex.EncodeToString(store.Entropy); got != test.entropyHex {
				t.Fatalf("entropy = %s, want %s", got, test.entropyHex)
			}
			address, err := store.GetBaseAddress()
			if err != nil {
				t.Fatalf("GetBaseAddress() error = %v", err)
			}
			if got := address.String(); got != test.baseAddress {
				t.Fatalf("base address = %s, want %s", got, test.baseAddress)
			}
		})
	}
}

func TestToEncryptedFileEmitsRawEntropyAndCompleteArgon2Parameters(t *testing.T) {
	entropy := bytes.Repeat([]byte{0x42}, 32)
	store, err := NewKeyStoreFromEntropy(entropy)
	if err != nil {
		t.Fatalf("NewKeyStoreFromEntropy() error = %v", err)
	}

	file, err := store.ToEncryptedFile("password", nil)
	if err != nil {
		t.Fatalf("ToEncryptedFile() error = %v", err)
	}
	plaintext, err := file.Decrypt("password")
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if !bytes.Equal(plaintext, entropy) {
		t.Fatalf("plaintext = %x, want raw entropy %x", plaintext, entropy)
	}
	if json.Valid(plaintext) {
		t.Fatal("plaintext unexpectedly contains the legacy Go JSON payload")
	}

	defaults := sdkcrypto.DefaultArgon2Parameters()
	params := file.Crypto.Argon2Params
	if params.TimeCost != defaults.Iterations || params.MemoryCost != defaults.Memory ||
		params.HashLength != defaults.KeyLength || params.Parallelism != defaults.Parallelism {
		t.Fatalf("Argon2 parameters = %+v, want %+v", params, defaults)
	}
	if file.NeedsUpgrade() {
		t.Fatal("NeedsUpgrade() = true for a newly generated key file")
	}

	encoded, err := file.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}
	var document struct {
		Crypto struct {
			Argon2Params map[string]json.RawMessage `json:"argon2Params"`
		} `json:"crypto"`
	}
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	for _, field := range []string{"salt", "timeCost", "memoryCost", "hashLength", "parallelism"} {
		if _, ok := document.Crypto.Argon2Params[field]; !ok {
			t.Errorf("argon2Params is missing %q", field)
		}
	}
}

func TestFromEncryptedFileRejectsTamperedBaseAddress(t *testing.T) {
	store, err := NewKeyStoreFromEntropy(bytes.Repeat([]byte{0x24}, 32))
	if err != nil {
		t.Fatalf("NewKeyStoreFromEntropy() error = %v", err)
	}
	file, err := store.ToEncryptedFile("password", nil)
	if err != nil {
		t.Fatalf("ToEncryptedFile() error = %v", err)
	}
	file.Metadata[BaseAddressKey] = "z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f"

	_, err = FromEncryptedFile(file, "password")
	if !errors.Is(err, ErrInvalidKeyStore) {
		t.Fatalf("FromEncryptedFile() error = %v, want ErrInvalidKeyStore", err)
	}
}

func TestFromEncryptedFileReadsLegacyGoJSONPayload(t *testing.T) {
	store, err := NewKeyStoreFromEntropy(bytes.Repeat([]byte{0x18}, 32))
	if err != nil {
		t.Fatalf("NewKeyStoreFromEntropy() error = %v", err)
	}
	legacyPlaintext, err := serializeKeyStoreData(map[string]interface{}{
		"mnemonic": store.Mnemonic,
		"entropy":  hex.EncodeToString(store.Entropy),
		"seed":     hex.EncodeToString(store.Seed),
	})
	if err != nil {
		t.Fatalf("serializeKeyStoreData() error = %v", err)
	}
	address, err := store.GetBaseAddress()
	if err != nil {
		t.Fatalf("GetBaseAddress() error = %v", err)
	}
	file, err := Encrypt(legacyPlaintext, "password", map[string]interface{}{
		BaseAddressKey: address.String(),
		WalletTypeKey:  KeyStoreWalletType,
	})
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	restored, err := FromEncryptedFile(file, "password")
	if err != nil {
		t.Fatalf("FromEncryptedFile() error = %v", err)
	}
	if !bytes.Equal(restored.Entropy, store.Entropy) {
		t.Fatalf("restored entropy = %x, want %x", restored.Entropy, store.Entropy)
	}
}

func TestToEncryptedFileOverridesCallerBaseAddressWithoutMutation(t *testing.T) {
	store, err := NewKeyStoreFromEntropy(bytes.Repeat([]byte{0x81}, 32))
	if err != nil {
		t.Fatalf("NewKeyStoreFromEntropy() error = %v", err)
	}
	metadata := map[string]interface{}{BaseAddressKey: "tampered", "name": "wallet"}
	file, err := store.ToEncryptedFile("password", metadata)
	if err != nil {
		t.Fatalf("ToEncryptedFile() error = %v", err)
	}
	if metadata[BaseAddressKey] != "tampered" {
		t.Fatalf("caller metadata was mutated: %v", metadata)
	}
	address, err := store.GetBaseAddress()
	if err != nil {
		t.Fatalf("GetBaseAddress() error = %v", err)
	}
	if got := file.Metadata[BaseAddressKey]; got != address.String() {
		t.Fatalf("file baseAddress = %v, want %s", got, address)
	}
}
