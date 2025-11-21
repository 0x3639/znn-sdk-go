package wallet

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// =============================================================================
// Encrypt Tests
// =============================================================================

func TestEncrypt_BasicUsage(t *testing.T) {
	data := []byte("secret data")
	password := "password123"
	metadata := map[string]interface{}{
		"name": "test-wallet",
	}

	ef, err := Encrypt(data, password, metadata)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if ef == nil {
		t.Fatal("Encrypted file should not be nil")
	}

	if ef.Crypto == nil {
		t.Error("Crypto params should not be nil")
	}

	if ef.Version != 1 {
		t.Errorf("Version = %d, want 1", ef.Version)
	}
}

func TestEncrypt_CryptoParams(t *testing.T) {
	data := []byte("secret data")
	password := "password123"

	ef, err := Encrypt(data, password, nil)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Check crypto params
	if ef.Crypto.CipherName != "aes-256-gcm" {
		t.Errorf("CipherName = %s, want aes-256-gcm", ef.Crypto.CipherName)
	}

	if ef.Crypto.Kdf != "argon2.IDKey" {
		t.Errorf("Kdf = %s, want argon2.IDKey", ef.Crypto.Kdf)
	}

	if ef.Crypto.Argon2Params == nil {
		t.Error("Argon2Params should not be nil")
	}

	// Check hex format
	if !strings.HasPrefix(ef.Crypto.CipherData, "0x") {
		t.Error("CipherData should have 0x prefix")
	}

	if !strings.HasPrefix(ef.Crypto.Nonce, "0x") {
		t.Error("Nonce should have 0x prefix")
	}

	if !strings.HasPrefix(ef.Crypto.Argon2Params.Salt, "0x") {
		t.Error("Salt should have 0x prefix")
	}
}

func TestEncrypt_Metadata(t *testing.T) {
	data := []byte("secret data")
	password := "password123"
	metadata := map[string]interface{}{
		"name":   "my-wallet",
		"type":   "keystore",
		"custom": 42,
	}

	ef, err := Encrypt(data, password, metadata)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if ef.Metadata["name"] != "my-wallet" {
		t.Error("Metadata not preserved")
	}

	if ef.Metadata["type"] != "keystore" {
		t.Error("Metadata not preserved")
	}

	if ef.Metadata["custom"] != 42 {
		t.Error("Metadata not preserved")
	}
}

func TestEncrypt_EmptyData(t *testing.T) {
	data := []byte{}
	password := "password123"

	ef, err := Encrypt(data, password, nil)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if ef == nil {
		t.Error("Should be able to encrypt empty data")
	}
}

func TestEncrypt_Randomness(t *testing.T) {
	data := []byte("secret data")
	password := "password123"

	ef1, err := Encrypt(data, password, nil)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	ef2, err := Encrypt(data, password, nil)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Same data and password should produce different ciphertexts (due to random salt/nonce)
	if ef1.Crypto.CipherData == ef2.Crypto.CipherData {
		t.Error("Ciphertexts should be different due to random salt/nonce")
	}

	if ef1.Crypto.Nonce == ef2.Crypto.Nonce {
		t.Error("Nonces should be different")
	}

	if ef1.Crypto.Argon2Params.Salt == ef2.Crypto.Argon2Params.Salt {
		t.Error("Salts should be different")
	}
}

// =============================================================================
// Decrypt Tests
// =============================================================================

func TestDecrypt_CorrectPassword(t *testing.T) {
	original := []byte("secret data")
	password := "password123"

	ef, err := Encrypt(original, password, nil)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := ef.Decrypt(password)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, original) {
		t.Error("Decrypted data does not match original")
	}
}

func TestDecrypt_IncorrectPassword(t *testing.T) {
	data := []byte("secret data")
	password := "password123"

	ef, err := Encrypt(data, password, nil)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	_, err = ef.Decrypt("wrongpassword")
	if !errors.Is(err, ErrIncorrectPassword) {
		t.Errorf("Decrypt() error = %v, want ErrIncorrectPassword", err)
	}
}

func TestDecrypt_EmptyPassword(t *testing.T) {
	data := []byte("secret data")
	password := ""

	ef, err := Encrypt(data, password, nil)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := ef.Decrypt(password)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, data) {
		t.Error("Should work with empty password")
	}
}

// =============================================================================
// Round Trip Tests
// =============================================================================

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	testCases := []struct {
		name     string
		data     []byte
		password string
		metadata map[string]interface{}
	}{
		{
			name:     "Simple",
			data:     []byte("hello world"),
			password: "password",
			metadata: nil,
		},
		{
			name:     "Empty data",
			data:     []byte{},
			password: "password",
			metadata: nil,
		},
		{
			name:     "Large data",
			data:     bytes.Repeat([]byte("a"), 10000),
			password: "password",
			metadata: nil,
		},
		{
			name:     "With metadata",
			data:     []byte("secret"),
			password: "password",
			metadata: map[string]interface{}{"name": "wallet"},
		},
		{
			name:     "Binary data",
			data:     []byte{0x00, 0xFF, 0x01, 0xFE, 0x02, 0xFD},
			password: "password",
			metadata: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ef, err := Encrypt(tc.data, tc.password, tc.metadata)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			decrypted, err := ef.Decrypt(tc.password)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if !bytes.Equal(decrypted, tc.data) {
				t.Error("Round trip failed: data mismatch")
			}

			// Check metadata preserved
			if tc.metadata != nil {
				for k, v := range tc.metadata {
					if ef.Metadata[k] != v {
						t.Errorf("Metadata[%s] not preserved", k)
					}
				}
			}
		})
	}
}

// =============================================================================
// ToJSON / FromJSON Tests
// =============================================================================

func TestToJSON_ValidStructure(t *testing.T) {
	data := []byte("secret data")
	password := "password123"
	metadata := map[string]interface{}{
		"name": "my-wallet",
	}

	ef, err := Encrypt(data, password, metadata)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	jsonData, err := ef.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Parse JSON to verify structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Check required fields
	if _, ok := parsed["crypto"]; !ok {
		t.Error("JSON should contain 'crypto' field")
	}

	if _, ok := parsed["timestamp"]; !ok {
		t.Error("JSON should contain 'timestamp' field")
	}

	if _, ok := parsed["version"]; !ok {
		t.Error("JSON should contain 'version' field")
	}

	if _, ok := parsed["name"]; !ok {
		t.Error("JSON should contain metadata 'name' field")
	}
}

func TestFromJSON_ValidJSON(t *testing.T) {
	jsonStr := `{
		"name": "my-wallet",
		"crypto": {
			"argon2Params": {
				"salt": "0x0123456789abcdef0123456789abcdef"
			},
			"cipherData": "0xabcdef",
			"cipherName": "aes-256-gcm",
			"kdf": "argon2.IDKey",
			"nonce": "0x0123456789abcdef01234567"
		},
		"timestamp": 1234567890,
		"version": 1
	}`

	ef, err := FromJSON([]byte(jsonStr))
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if ef.Version != 1 {
		t.Errorf("Version = %d, want 1", ef.Version)
	}

	if ef.Timestamp != 1234567890 {
		t.Errorf("Timestamp = %d, want 1234567890", ef.Timestamp)
	}

	if ef.Metadata["name"] != "my-wallet" {
		t.Error("Metadata not parsed correctly")
	}

	if ef.Crypto == nil {
		t.Error("Crypto params not parsed")
	}
}

func TestFromJSON_InvalidJSON(t *testing.T) {
	testCases := []string{
		"invalid json",
		"{}",
		`{"crypto": "invalid"}`,
	}

	for _, jsonStr := range testCases {
		_, err := FromJSON([]byte(jsonStr))
		// Should either error or return incomplete data
		// We don't enforce strict validation, just check it doesn't panic
		_ = err
	}
}

func TestJSONRoundTrip(t *testing.T) {
	data := []byte("secret data")
	password := "password123"
	metadata := map[string]interface{}{
		"name": "my-wallet",
		"type": "keystore",
	}

	// Encrypt
	ef1, err := Encrypt(data, password, metadata)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Serialize
	jsonData, err := ef1.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Deserialize
	ef2, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	// Decrypt with second instance
	decrypted, err := ef2.Decrypt(password)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, data) {
		t.Error("JSON round trip failed: data mismatch")
	}

	// Check metadata preserved
	if ef2.Metadata["name"] != "my-wallet" {
		t.Error("Metadata not preserved through JSON round trip")
	}
}

// =============================================================================
// hexToBytes Helper Tests
// =============================================================================

func TestHexToBytes_WithPrefix(t *testing.T) {
	result, err := hexToBytes("0x0123456789abcdef")
	if err != nil {
		t.Fatalf("hexToBytes() error = %v", err)
	}

	expected := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}
	if !bytes.Equal(result, expected) {
		t.Errorf("hexToBytes() = %x, want %x", result, expected)
	}
}

func TestHexToBytes_WithoutPrefix(t *testing.T) {
	result, err := hexToBytes("0123456789abcdef")
	if err != nil {
		t.Fatalf("hexToBytes() error = %v", err)
	}

	expected := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}
	if !bytes.Equal(result, expected) {
		t.Errorf("hexToBytes() = %x, want %x", result, expected)
	}
}

func TestHexToBytes_Invalid(t *testing.T) {
	testCases := []string{
		"0xzz",
		"notahex",
		"0x1", // Odd length
	}

	for _, hex := range testCases {
		_, err := hexToBytes(hex)
		if err == nil {
			t.Errorf("hexToBytes(%q) should return error", hex)
		}
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestEncryptedFile_CompleteWorkflow(t *testing.T) {
	// Original data
	data := []byte("my secret wallet data")
	password := "strong-password-123"
	metadata := map[string]interface{}{
		BaseAddressKey: "z1xxx...",
		WalletTypeKey:  KeyStoreWalletType,
		"custom":       "metadata",
	}

	// 1. Encrypt
	ef, err := Encrypt(data, password, metadata)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// 2. Serialize to JSON
	jsonData, err := ef.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// 3. Deserialize from JSON
	ef2, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	// 4. Decrypt
	decrypted, err := ef2.Decrypt(password)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	// 5. Verify
	if !bytes.Equal(decrypted, data) {
		t.Error("Complete workflow failed: data mismatch")
	}

	if ef2.Metadata[BaseAddressKey] != "z1xxx..." {
		t.Error("Metadata not preserved")
	}

	// 6. Try wrong password
	_, err = ef2.Decrypt("wrong-password")
	if !errors.Is(err, ErrIncorrectPassword) {
		t.Error("Should fail with incorrect password")
	}
}

// =============================================================================
// Performance Benchmarks
// =============================================================================

func BenchmarkEncrypt(b *testing.B) {
	data := bytes.Repeat([]byte("a"), 1000)
	password := "password123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encrypt(data, password, nil)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	data := bytes.Repeat([]byte("a"), 1000)
	password := "password123"
	ef, _ := Encrypt(data, password, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ef.Decrypt(password)
	}
}

func BenchmarkToJSON(b *testing.B) {
	data := []byte("secret data")
	password := "password123"
	ef, _ := Encrypt(data, password, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ef.ToJSON()
	}
}

func BenchmarkFromJSON(b *testing.B) {
	data := []byte("secret data")
	password := "password123"
	ef, _ := Encrypt(data, password, nil)
	jsonData, _ := ef.ToJSON()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FromJSON(jsonData)
	}
}
