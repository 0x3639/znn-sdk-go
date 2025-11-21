package crypto

import (
	"crypto/ed25519"
	"testing"
)

// FuzzSign tests Ed25519 signing with random messages
func FuzzSign(f *testing.F) {
	// Add seed corpus
	f.Add([]byte("test message"))
	f.Add([]byte(""))
	f.Add([]byte("a"))
	f.Add(make([]byte, 100))
	f.Add(make([]byte, 1000))

	// Create a fixed test keypair
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		f.Fatalf("failed to generate test key: %v", err)
	}

	f.Fuzz(func(t *testing.T, message []byte) {
		// Sign message
		signature, err := Sign(message, priv)
		if err != nil {
			t.Fatalf("signing failed: %v", err)
		}

		// Verify signature length
		if len(signature) != ed25519.SignatureSize {
			t.Fatalf("invalid signature length: got %d, want %d", len(signature), ed25519.SignatureSize)
		}

		// Verify signature is valid
		valid, err := Verify(signature, message, pub)
		if err != nil {
			t.Fatalf("verify error: %v", err)
		}
		if !valid {
			t.Fatalf("signature verification failed")
		}

		// Verify wrong message fails
		if len(message) > 0 {
			wrongMessage := append([]byte{0xFF}, message...)
			valid, err = Verify(signature, wrongMessage, pub)
			if err != nil {
				t.Fatalf("verify error on wrong message: %v", err)
			}
			if valid {
				t.Fatalf("signature verified with wrong message")
			}
		}
	})
}

// FuzzDigest tests SHA3-256 hashing with random data
func FuzzDigest(f *testing.F) {
	// Add seed corpus
	f.Add([]byte("test data"), 32)
	f.Add([]byte(""), 32)
	f.Add(make([]byte, 100), 32)
	f.Add([]byte("data"), 64)
	f.Add([]byte("data"), 16)

	f.Fuzz(func(t *testing.T, data []byte, digestSize int) {
		// Skip invalid digest sizes
		if digestSize <= 0 || digestSize > 1024 {
			t.Skip("invalid digest size")
		}

		// Compute digest
		hash := Digest(data, digestSize)

		// Verify hash length
		expectedLen := digestSize
		if digestSize == 0 || digestSize == 32 {
			expectedLen = 32
		}

		if len(hash) != expectedLen {
			t.Fatalf("invalid hash length: got %d, want %d", len(hash), expectedLen)
		}

		// Verify determinism
		hash2 := Digest(data, digestSize)
		if string(hash) != string(hash2) {
			t.Fatalf("digest not deterministic")
		}

		// Verify different data produces different hash (unless empty)
		if len(data) > 0 {
			differentData := append([]byte{0xFF}, data...)
			hash3 := Digest(differentData, digestSize)
			if string(hash) == string(hash3) {
				t.Fatalf("different data produced same hash")
			}
		}
	})
}

// FuzzVerify tests signature verification with random inputs
func FuzzVerify(f *testing.F) {
	// Create a valid signature for testing
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		f.Fatalf("failed to generate test key: %v", err)
	}

	message := []byte("test message")
	validSig, err := Sign(message, priv)
	if err != nil {
		f.Fatalf("failed to create test signature: %v", err)
	}

	// Add seed corpus
	f.Add(validSig, message, []byte(pub))
	f.Add(make([]byte, ed25519.SignatureSize), message, []byte(pub))
	f.Add(validSig, []byte("wrong message"), []byte(pub))

	f.Fuzz(func(t *testing.T, sig []byte, msg []byte, pubKey []byte) {
		// Call Verify - should not panic with any input
		_, _ = Verify(sig, msg, pubKey)

		// No assertions - just ensure no panics occur
	})
}
