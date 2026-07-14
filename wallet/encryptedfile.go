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

// EncryptedFile represents a versioned, encrypted Zenon wallet file.
//
// Metadata contains top-level key-file properties such as baseAddress and
// walletType. Crypto describes the Argon2id and AES-256-GCM parameters needed
// to decrypt the payload. Timestamp is expressed as Unix seconds, and Version
// identifies the key-file format.
//
// Use [Encrypt] to create an EncryptedFile, [FromJSON] to parse one, and
// [EncryptedFile.Decrypt] to authenticate and decrypt its payload. Wallet
// callers should normally use [FromEncryptedFile], which additionally decodes
// the wallet entropy and validates the base address.
type EncryptedFile struct {
	Metadata  map[string]interface{} `json:",inline"`
	Crypto    *CryptoParams          `json:"crypto"`
	Timestamp int64                  `json:"timestamp"`
	Version   int                    `json:"version"`
}

// CryptoParams contains the self-describing encryption parameters stored in a
// Zenon key file.
//
// Argon2Params controls key derivation. CipherData contains AES-GCM ciphertext
// followed by its authentication tag. Nonce and CipherData are hexadecimal
// strings with an optional 0x prefix.
type CryptoParams struct {
	Argon2Params *Argon2Params `json:"argon2Params"`
	CipherData   string        `json:"cipherData"` // Hex encoded
	CipherName   string        `json:"cipherName"` // "aes-256-gcm"
	Kdf          string        `json:"kdf"`        // "argon2.IDKey"
	Nonce        string        `json:"nonce"`      // Hex encoded
}

// Argon2Params contains the Argon2id key-derivation parameters persisted in a
// Zenon key file.
//
// Legacy files contain only Salt. Missing cost fields use the stable Zenon
// defaults during decryption and cause [EncryptedFile.NeedsUpgrade] to return
// true. Newly encrypted files always persist every field.
type Argon2Params struct {
	Salt        string `json:"salt"` // Hex encoded
	TimeCost    uint32 `json:"timeCost,omitempty"`
	MemoryCost  uint32 `json:"memoryCost,omitempty"`
	HashLength  uint32 `json:"hashLength,omitempty"`
	Parallelism uint8  `json:"parallelism,omitempty"`
}

// Encrypt creates a version-one encrypted key file from plaintext data.
//
// Parameters:
//   - data: Plaintext bytes to authenticate and encrypt.
//   - password: UTF-8 password used by Argon2id.
//   - metadata: Optional top-level key-file metadata. The map is retained by
//     the returned value, so callers that need isolation should pass a copy.
//
// Encrypt returns a self-describing EncryptedFile using the stable Argon2id
// defaults and AES-256-GCM, or an error if secure randomness or cipher setup
// fails.
//
// Example:
//
//	file, err := Encrypt(entropy, "correct horse battery staple", metadata)
//	if err != nil {
//		return err
//	}
//	encoded, err := file.ToJSON()
//
// Security Note: Encryption authenticates both the ciphertext and the fixed
// Zenon associated data. Prefer [KeyStore.ToEncryptedFile] for wallet entropy,
// because it also records the derived base address.
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
				Salt:        "0x" + hex.EncodeToString(salt),
				TimeCost:    params.Iterations,
				MemoryCost:  params.Memory,
				HashLength:  params.KeyLength,
				Parallelism: params.Parallelism,
			},
			CipherData: "0x" + hex.EncodeToString(ciphertext),
			CipherName: "aes-256-gcm",
			Kdf:        "argon2.IDKey",
			Nonce:      "0x" + hex.EncodeToString(nonce),
		},
	}

	return ef, nil
}

// Decrypt authenticates and decrypts the encrypted file with password.
//
// Legacy files that omit Argon2 cost fields are decrypted with the stable
// default configuration. Self-describing files use their persisted Argon2
// parameters. The method validates the key-file version, algorithms, salt,
// nonce, and AES-256 key length before decrypting.
//
// Parameters:
//   - password: UTF-8 password used to derive the AES key.
//
// Decrypt returns the plaintext bytes. It returns [ErrIncorrectPassword] when
// AES-GCM authentication fails and a descriptive error for malformed or
// unsupported key-file parameters.
//
// Example:
//
//	plaintext, err := file.Decrypt("correct horse battery staple")
//	if errors.Is(err, ErrIncorrectPassword) {
//		log.Print("password or key-file authentication is invalid")
//	}
//
// Security Note: Callers are responsible for clearing plaintext when it is no
// longer needed. Use [FromEncryptedFile] to validate wallet metadata after
// decryption.
func (ef *EncryptedFile) Decrypt(password string) ([]byte, error) {
	if err := ef.validateEncryptionEnvelope(); err != nil {
		return nil, err
	}
	salt, nonce, ciphertext, err := ef.decodeEncryptionPayload()
	if err != nil {
		return nil, err
	}

	// Derive key using the persisted parameters, falling back to all stable
	// defaults for legacy salt-only files.
	params, err := ef.argon2Parameters()
	if err != nil {
		return nil, err
	}
	key := crypto.DeriveKey([]byte(password), salt, params)

	return decryptAESGCM(key, nonce, ciphertext)
}

func (ef *EncryptedFile) validateEncryptionEnvelope() error {
	if ef == nil {
		return fmt.Errorf("invalid encrypted file: nil")
	}
	if ef.Version != 1 {
		return fmt.Errorf("unsupported encrypted file version: %d", ef.Version)
	}
	if ef.Crypto == nil || ef.Crypto.Argon2Params == nil {
		return fmt.Errorf("invalid encrypted file: missing crypto parameters")
	}
	if ef.Crypto.CipherName != "aes-256-gcm" {
		return fmt.Errorf("unsupported cipher: %q", ef.Crypto.CipherName)
	}
	if ef.Crypto.Kdf != "argon2.IDKey" {
		return fmt.Errorf("unsupported key derivation function: %q", ef.Crypto.Kdf)
	}
	return nil
}

func (ef *EncryptedFile) decodeEncryptionPayload() ([]byte, []byte, []byte, error) {
	salt, err := hexToBytes(ef.Crypto.Argon2Params.Salt)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid Argon2 salt: %w", err)
	}
	if len(salt) != 16 {
		return nil, nil, nil, fmt.Errorf("invalid Argon2 salt length: got %d, want 16", len(salt))
	}
	nonce, err := hexToBytes(ef.Crypto.Nonce)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid AES-GCM nonce: %w", err)
	}
	if len(nonce) != 12 {
		return nil, nil, nil, fmt.Errorf("invalid AES-GCM nonce length: got %d, want 12", len(nonce))
	}
	ciphertext, err := hexToBytes(ef.Crypto.CipherData)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid cipher data: %w", err)
	}
	return salt, nonce, ciphertext, nil
}

func decryptAESGCM(key, nonce, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	aad := []byte("zenon")
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, aad)
	if err != nil {
		return nil, ErrIncorrectPassword
	}
	return plaintext, nil
}

// NeedsUpgrade reports whether the key file should be re-encrypted with the
// current Argon2id configuration.
//
// Parameters:
//   - target: Optional target Argon2 configuration. When omitted, the stable
//     Zenon defaults from [crypto.DefaultArgon2Parameters] are used. At most
//     one target may be supplied.
//
// NeedsUpgrade returns true for legacy salt-only files, incomplete parameter
// sets, unsupported algorithms or versions, and parameters that differ from
// the target configuration. It does not decrypt the file or validate the
// password.
//
// Example:
//
//	if file.NeedsUpgrade() {
//		log.Print("re-encrypt this wallet after the next successful unlock")
//	}
//
// Security Note: A true result is advisory; decrypt and validate the existing
// file before replacing it. See [FromEncryptedFile] for integrity validation.
func (ef *EncryptedFile) NeedsUpgrade(target ...crypto.Argon2Parameters) bool {
	if len(target) > 1 || !ef.hasCurrentEncryptionEnvelope() {
		return true
	}

	desired := crypto.DefaultArgon2Parameters()
	if len(target) == 1 {
		desired = target[0]
	}
	actual := ef.Crypto.Argon2Params
	return !actual.isComplete() || !actual.matches(desired)
}

func (ef *EncryptedFile) hasCurrentEncryptionEnvelope() bool {
	return ef != nil && ef.Version == 1 && ef.Crypto != nil &&
		ef.Crypto.Argon2Params != nil && ef.Crypto.CipherName == "aes-256-gcm" &&
		ef.Crypto.Kdf == "argon2.IDKey"
}

func (params *Argon2Params) isComplete() bool {
	return params.TimeCost != 0 && params.MemoryCost != 0 &&
		params.HashLength != 0 && params.Parallelism != 0
}

func (params *Argon2Params) matches(target crypto.Argon2Parameters) bool {
	return params.TimeCost == target.Iterations && params.MemoryCost == target.Memory &&
		params.HashLength == target.KeyLength && params.Parallelism == target.Parallelism
}

func (ef *EncryptedFile) argon2Parameters() (crypto.Argon2Parameters, error) {
	defaults := crypto.DefaultArgon2Parameters()
	stored := ef.Crypto.Argon2Params
	if stored.TimeCost == 0 && stored.MemoryCost == 0 &&
		stored.HashLength == 0 && stored.Parallelism == 0 {
		return defaults, nil
	}

	params := defaults
	if stored.TimeCost != 0 {
		params.Iterations = stored.TimeCost
	}
	if stored.MemoryCost != 0 {
		params.Memory = stored.MemoryCost
	}
	if stored.HashLength != 0 {
		params.KeyLength = stored.HashLength
	}
	if stored.Parallelism != 0 {
		params.Parallelism = stored.Parallelism
	}
	if params.KeyLength != 32 {
		return crypto.Argon2Parameters{}, fmt.Errorf("invalid AES-256 key length: got %d, want 32", params.KeyLength)
	}
	return params, nil
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
