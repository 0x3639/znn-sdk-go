package wallet

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// Regression test for finding #34: isAllSameChar compared the first *byte*
// of the string against decoded runes, so any password whose first character
// is multi-byte UTF-8 was never reported as all-same-character.
func TestValidatePassword_RejectsAllSameCharMultibyte(t *testing.T) {
	passwords := []string{
		// 8 identical runes (U+00E9), 16 bytes
		"éééééééé",
		// 8 identical runes (U+20AC), 24 bytes
		"€€€€€€€€",
		// 8 identical runes (U+1F600), 32 bytes
		strings.Repeat("😀", 8),
		// 8 repetitions of one invalid UTF-8 byte value
		strings.Repeat("\xff", 8),
	}
	for _, password := range passwords {
		if !isAllSameChar(password) {
			t.Errorf("isAllSameChar(%q) = false, want true", password)
		}
		if err := ValidatePassword(password); err == nil {
			t.Errorf("ValidatePassword(%q) accepted an all-same-character password", password)
		}
		if strength := AnalyzePasswordStrength(password); strength != PasswordWeak {
			t.Errorf("AnalyzePasswordStrength(%q) = %v, want Weak", password, strength)
		}
	}
}

// Regression test for finding #35: len(password) counts bytes, so passwords
// with far fewer than MinPasswordLength characters passed validation whenever
// their UTF-8 encoding was wide enough.
func TestValidatePassword_MinLengthCountsCharactersNotBytes(t *testing.T) {
	passwords := []string{
		// 2 characters (8 bytes)
		"😀😀",
		// 3 characters (9 bytes)
		"€€€",
		// 5 characters (15 bytes), not all-same-character
		"日本語語語",
	}
	for _, password := range passwords {
		if n := utf8.RuneCountInString(password); n >= MinPasswordLength {
			t.Fatalf("bad vector %q: %d runes >= %d", password, n, MinPasswordLength)
		}
		if err := ValidatePassword(password); err == nil {
			t.Errorf("ValidatePassword(%q) accepted a %d-character password", password, utf8.RuneCountInString(password))
		}
		if strength := AnalyzePasswordStrength(password); strength != PasswordWeak {
			t.Errorf("AnalyzePasswordStrength(%q) = %v, want Weak", password, strength)
		}
	}
}

// Valid multi-byte passwords of sufficient character count must still pass.
func TestValidatePassword_AcceptsSufficientUnicodePasswords(t *testing.T) {
	passwords := []string{
		"こんにちは世界です",  // 9 runes
		"Привет123",  // 9 runes
		"café☕pass1", // 10 runes
	}
	for _, password := range passwords {
		if err := ValidatePassword(password); err != nil {
			t.Errorf("ValidatePassword(%q) = %v, want nil", password, err)
		}
	}
}
