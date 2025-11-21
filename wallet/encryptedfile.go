package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0x3639/znn-sdk-go/crypto"
)

// EncryptedFile represents an encrypted wallet file
type EncryptedFile struct {
	Metadata  map[string]interface{} `json:",inline"`
	Crypto    *CryptoParams          `json:"crypto"`
	Timestamp int64                  `json:"timestamp"`
	Version   int                    `json:"version"`
}

// CryptoParams contains encryption parameters
type CryptoParams struct {
	Argon2Params *Argon2Params `json:"argon2Params"`
	CipherData   string        `json:"cipherData"` // Hex encoded
	CipherName   string        `json:"cipherName"` // "aes-256-gcm"
	Kdf          string        `json:"kdf"`        // "argon2.IDKey"
	Nonce        string        `json:"nonce"`      // Hex encoded
}

// Argon2Params contains Argon2 key derivation parameters
type Argon2Params struct {
	Salt string `json:"salt"` // Hex encoded
}

// Encrypt creates an encrypted file from data and password
func Encrypt(data []byte, password string, metadata map[string]interface{}) (*EncryptedFile, error) {
	timestamp := time.Now().Unix()

	// Generate random salt (16 bytes)
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// Generate random nonce for AES-GCM (12 bytes)
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Derive key using Argon2
	params := crypto.DefaultArgon2Parameters()
	key := crypto.DeriveKey([]byte(password), salt, params)

	// Create AES-256-GCM cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Encrypt with additional authenticated data "zenon"
	aad := []byte("zenon")
	// #nosec G407 -- nonce is randomly generated on line 49, not hardcoded
	ciphertext := aesgcm.Seal(nil, nonce, data, aad)

	// Create encrypted file
	ef := &EncryptedFile{
		Metadata:  metadata,
		Timestamp: timestamp,
		Version:   1,
		Crypto: &CryptoParams{
			Argon2Params: &Argon2Params{
				Salt: "0x" + hex.EncodeToString(salt),
			},
			CipherData: "0x" + hex.EncodeToString(ciphertext),
			CipherName: "aes-256-gcm",
			Kdf:        "argon2.IDKey",
			Nonce:      "0x" + hex.EncodeToString(nonce),
		},
	}

	return ef, nil
}

// Decrypt decrypts the encrypted file with the given password
func (ef *EncryptedFile) Decrypt(password string) ([]byte, error) {
	// Decode hex values
	salt, err := hexToBytes(ef.Crypto.Argon2Params.Salt)
	if err != nil {
		return nil, err
	}

	nonce, err := hexToBytes(ef.Crypto.Nonce)
	if err != nil {
		return nil, err
	}

	ciphertext, err := hexToBytes(ef.Crypto.CipherData)
	if err != nil {
		return nil, err
	}

	// Derive key using Argon2
	params := crypto.DefaultArgon2Parameters()
	key := crypto.DeriveKey([]byte(password), salt, params)

	// Create AES-256-GCM cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt with additional authenticated data "zenon"
	aad := []byte("zenon")
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		// Authentication failed - likely incorrect password
		return nil, ErrIncorrectPassword
	}

	return plaintext, nil
}

// ToJSON serializes the encrypted file to JSON
func (ef *EncryptedFile) ToJSON() ([]byte, error) {
	// Flatten metadata into the main object
	result := make(map[string]interface{})

	// Copy metadata
	if ef.Metadata != nil {
		for k, v := range ef.Metadata {
			result[k] = v
		}
	}

	// Add standard fields
	result["crypto"] = ef.Crypto
	result["timestamp"] = ef.Timestamp
	result["version"] = ef.Version

	return json.MarshalIndent(result, "", "  ")
}

// FromJSON deserializes an encrypted file from JSON
func FromJSON(data []byte) (*EncryptedFile, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	ef := &EncryptedFile{
		Metadata: make(map[string]interface{}),
	}

	// Extract standard fields
	if crypto, ok := raw["crypto"].(map[string]interface{}); ok {
		cryptoJSON, err := json.Marshal(crypto)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal crypto params: %w", err)
		}
		var cp CryptoParams
		if err := json.Unmarshal(cryptoJSON, &cp); err != nil {
			return nil, err
		}
		ef.Crypto = &cp
		delete(raw, "crypto")
	}

	if timestamp, ok := raw["timestamp"].(float64); ok {
		ef.Timestamp = int64(timestamp)
		delete(raw, "timestamp")
	}

	if version, ok := raw["version"].(float64); ok {
		ef.Version = int(version)
		delete(raw, "version")
	}

	// Remaining fields are metadata
	for k, v := range raw {
		ef.Metadata[k] = v
	}

	return ef, nil
}

// hexToBytes converts a hex string (with or without 0x prefix) to bytes
func hexToBytes(s string) ([]byte, error) {
	// Remove 0x prefix if present
	if len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]
	}

	return hex.DecodeString(s)
}
