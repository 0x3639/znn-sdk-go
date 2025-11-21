// Package crypto provides cryptographic primitives for the Zenon Network, including
// Ed25519 signature operations, SHA3 hashing, and Argon2 key derivation.
//
// This package implements the cryptographic operations required for:
//   - Transaction signing and verification (Ed25519)
//   - Address derivation and validation
//   - Keystore encryption (Argon2)
//   - Hash generation (SHA3-256)
//
// # Ed25519 Signatures
//
// Sign and verify messages with Ed25519:
//
//	// Sign a message
//	signature, err := crypto.Sign(message, privateKey)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Verify signature
//	valid := crypto.Verify(signature, message, publicKey)
//	if !valid {
//	    log.Fatal("Invalid signature")
//	}
//
// # Public Key Operations
//
// Derive public key from private key:
//
//	publicKey, err := crypto.GetPublicKey(privateKey)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # SHA3 Hashing
//
// Generate SHA3-256 hashes:
//
//	hash := crypto.Hash(data)
//	fmt.Printf("Hash: %x\n", hash)
//
// # Argon2 Key Derivation
//
// Used for keystore encryption (typically handled by wallet package):
//
//	derivedKey := crypto.DeriveKey(password, salt, iterations, memory, threads, keyLen)
//
// # Address Derivation
//
// Zenon addresses are derived from public keys using:
//  1. SHA3-256 hash of public key
//  2. Core bytes selection
//  3. Bech32 encoding with 'z' prefix
//
// This is typically handled by the wallet package's address derivation methods.
//
// # Security Considerations
//
// - Ed25519 provides 128-bit security level
// - All private keys should be stored encrypted
// - Use secure random number generation for key material
// - Never reuse nonces or expose private keys
//
// For more information, see https://pkg.go.dev/github.com/0x3639/znn-sdk-go/crypto
package crypto
