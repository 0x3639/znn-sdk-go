package wallet

import (
	"strings"
	"testing"
)

// FuzzPasswordValidation tests password validation with fuzzy inputs
func FuzzPasswordValidation(f *testing.F) {
	// Add seed corpus
	f.Add("password")                // Valid password
	f.Add("12345678")                // Valid 8 chars
	f.Add("short")                   // Too short
	f.Add("")                        // Empty
	f.Add("a")                       // Too short
	f.Add("aaaaaaa")                 // Too short (7 chars)
	f.Add("aaaaaaaa")                // Valid length but all same chars
	f.Add("abcdefgh")                // Valid
	f.Add(string(make([]byte, 100))) // Long password

	f.Fuzz(func(t *testing.T, password string) {
		// Call ValidatePassword - should not panic
		err := ValidatePassword(password)

		// Check expected errors for obvious cases
		if len(password) < MinPasswordLength {
			if err == nil {
				t.Fatalf("expected error for password length %d < %d", len(password), MinPasswordLength)
			}
			return
		}

		// For valid length passwords, just ensure no panic
		// (don't try to replicate the all-same-char logic as it's complex with UTF-8)
	})
}

// FuzzPasswordStrengthAnalysis tests password strength analyzer
func FuzzPasswordStrengthAnalysis(f *testing.F) {
	// Add seed corpus
	f.Add("password")     // Weak/Moderate
	f.Add("Password1")    // Strong
	f.Add("P@ssw0rd!")    // Strong
	f.Add("12345678")     // Moderate
	f.Add("abcdefgh")     // Moderate
	f.Add("ABCDEFGH")     // Moderate
	f.Add("aBcDeFgH")     // Strong
	f.Add("aB3dE!gH")     // Strong
	f.Add("aB3dE!gHiJkL") // Very Strong (12+ chars, 3+ classes)
	f.Add("")             // Weak
	f.Add("short")        // Weak

	f.Fuzz(func(t *testing.T, password string) {
		// Call AnalyzePasswordStrength - should not panic
		strength := AnalyzePasswordStrength(password)

		// Verify strength is one of the defined constants
		if strength != PasswordWeak && strength != PasswordModerate &&
			strength != PasswordStrong && strength != PasswordVeryStrong {
			t.Fatalf("invalid password strength value: %d", strength)
		}

		// Verify empty or short passwords are weak
		if len(password) < MinPasswordLength {
			if strength != PasswordWeak {
				t.Fatalf("short password should be weak, got %d", strength)
			}
			return
		}

		// For valid length passwords, just ensure no panic and valid strength value
	})
}

// FuzzMnemonicValidation tests BIP39 mnemonic validation
func FuzzMnemonicValidation(f *testing.F) {
	// Valid 24-word mnemonic
	validMnemonic := "route become dream access impulse price inform obtain engage ski believe awful absent pig thing vibrant possible exotic flee pepper marble rural fire fancy"

	// Add seed corpus
	f.Add(validMnemonic)                  // Valid
	f.Add("")                             // Empty
	f.Add("invalid words here")           // Invalid words
	f.Add("route become dream")           // Too few words
	f.Add(strings.Repeat("abandon ", 24)) // 24 same words
	f.Add(strings.Repeat("abandon ", 12)) // 12 same words

	f.Fuzz(func(t *testing.T, mnemonic string) {
		// Skip very long inputs to avoid timeouts
		if len(mnemonic) > 10000 {
			t.Skip("mnemonic too long")
		}

		// Call ValidateMnemonic - should not panic
		words := strings.Fields(mnemonic)
		valid := ValidateMnemonic(words)

		// Check expected results
		if len(words) == 0 {
			if valid {
				t.Fatalf("empty mnemonic should be invalid")
			}
			return
		}

		// Valid word counts are 12, 15, 18, 21, 24
		validWordCount := len(words) == 12 || len(words) == 15 ||
			len(words) == 18 || len(words) == 21 || len(words) == 24

		if !validWordCount && valid {
			t.Fatalf("invalid word count %d should be invalid", len(words))
		}
	})
}

// FuzzKeyPairCreation tests KeyPair creation with fuzzy private keys
func FuzzKeyPairCreation(f *testing.F) {
	// Add seed corpus
	validKey := make([]byte, 64)
	f.Add(validKey)
	f.Add(make([]byte, 32))  // Wrong size
	f.Add(make([]byte, 0))   // Empty
	f.Add(make([]byte, 100)) // Too large

	f.Fuzz(func(t *testing.T, privateKey []byte) {
		// Create keypair - should not panic
		kp := NewKeyPair(privateKey)
		if kp == nil {
			t.Fatal("NewKeyPair returned nil")
		}
		defer kp.Destroy()

		// Verify key is accessible (may be zero if invalid input)
		_ = kp.GetPrivateKey()
	})
}
