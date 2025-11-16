package wallet

import (
	"strings"
	"testing"
)

// =============================================================================
// ValidatePassword Tests
// =============================================================================

func TestValidatePassword_Valid(t *testing.T) {
	validPasswords := []string{
		"12345678",                           // Minimum length
		"abcdefgh",                           // All lowercase
		"ABCDEFGH",                           // All uppercase
		"password123",                        // Mixed alphanumeric
		"MyP@ssw0rd",                         // Mixed with special chars
		"a very long passphrase with spaces", // Long phrase
		"こんにちは世界",                            // Unicode characters
	}

	for _, password := range validPasswords {
		err := ValidatePassword(password)
		if err != nil {
			t.Errorf("ValidatePassword(%q) returned error: %v", password, err)
		}
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	shortPasswords := []string{
		"",        // Empty
		"a",       // 1 char
		"abc",     // 3 chars
		"1234567", // 7 chars (just under minimum)
	}

	for _, password := range shortPasswords {
		err := ValidatePassword(password)
		if err == nil {
			t.Errorf("ValidatePassword(%q) should reject password shorter than %d chars",
				password, MinPasswordLength)
		}
		if err != nil && !strings.Contains(err.Error(), "at least") {
			t.Errorf("ValidatePassword(%q) returned wrong error: %v", password, err)
		}
	}
}

func TestValidatePassword_AllSameChar(t *testing.T) {
	sameCharPasswords := []string{
		"aaaaaaaa",       // All 'a'
		"00000000",       // All '0'
		"        ",       // All spaces
		"!!!!!!!!",       // All '!'
		"zzzzzzzzzzzzzz", // Longer but all same
	}

	for _, password := range sameCharPasswords {
		err := ValidatePassword(password)
		if err == nil {
			t.Errorf("ValidatePassword(%q) should reject all-same-character passwords", password)
		}
		if err != nil && !strings.Contains(err.Error(), "same character") {
			t.Errorf("ValidatePassword(%q) returned wrong error: %v", password, err)
		}
	}
}

// =============================================================================
// AnalyzePasswordStrength Tests
// =============================================================================

func TestAnalyzePasswordStrength_Weak(t *testing.T) {
	weakPasswords := []string{
		"",         // Empty
		"1234567",  // Too short
		"aaaaaaaa", // All same char
		"00000000", // All same char (minimum length)
	}

	for _, password := range weakPasswords {
		strength := AnalyzePasswordStrength(password)
		if strength != PasswordWeak {
			t.Errorf("AnalyzePasswordStrength(%q) = %v, want Weak", password, strength)
		}
	}
}

func TestAnalyzePasswordStrength_Moderate(t *testing.T) {
	moderatePasswords := []string{
		"abcdefgh",  // All lowercase, minimum length
		"ABCDEFGH",  // All uppercase, minimum length
		"12345678",  // All digits, minimum length
		"password",  // Single class, >= 8 chars
		"lowercase", // All lowercase
	}

	for _, password := range moderatePasswords {
		strength := AnalyzePasswordStrength(password)
		if strength != PasswordModerate {
			t.Errorf("AnalyzePasswordStrength(%q) = %v, want Moderate", password, strength)
		}
	}
}

func TestAnalyzePasswordStrength_Strong(t *testing.T) {
	strongPasswords := []string{
		"Password1",   // Upper + lower + digit
		"mypass123",   // Lower + digit
		"HELLO123",    // Upper + digit
		"Pass@word",   // Upper + lower + special
		"hello-world", // Lower + special
	}

	for _, password := range strongPasswords {
		strength := AnalyzePasswordStrength(password)
		if strength != PasswordStrong {
			t.Errorf("AnalyzePasswordStrength(%q) = %v, want Strong", password, strength)
		}
	}
}

func TestAnalyzePasswordStrength_VeryStrong(t *testing.T) {
	veryStrongPasswords := []string{
		"MyP@ssw0rd123",           // 13 chars, all 4 classes
		"Tr0ub4dor&34",            // 12 chars, all 4 classes
		"Correct Horse Battery 1", // 24 chars, 4 classes (upper+lower+digit+space)
		"aB3!def@9hij",            // 12+ chars, 4 classes
	}

	for _, password := range veryStrongPasswords {
		strength := AnalyzePasswordStrength(password)
		if strength != PasswordVeryStrong {
			t.Errorf("AnalyzePasswordStrength(%q) = %v, want VeryStrong", password, strength)
		}
	}
}

// =============================================================================
// PasswordStrength.String() Tests
// =============================================================================

func TestPasswordStrength_String(t *testing.T) {
	tests := []struct {
		strength PasswordStrength
		expected string
	}{
		{PasswordWeak, "Weak"},
		{PasswordModerate, "Moderate"},
		{PasswordStrong, "Strong"},
		{PasswordVeryStrong, "Very Strong"},
		{PasswordStrength(999), "Unknown"},
	}

	for _, test := range tests {
		result := test.strength.String()
		if result != test.expected {
			t.Errorf("PasswordStrength(%d).String() = %q, want %q",
				test.strength, result, test.expected)
		}
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestIsAllSameChar(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"a", true},
		{"aaaa", true},
		{"0000", true},
		{"    ", true},
		{"ab", false},
		{"aaa b", false},
		{"123", false},
	}

	for _, test := range tests {
		result := isAllSameChar(test.input)
		if result != test.expected {
			t.Errorf("isAllSameChar(%q) = %v, want %v", test.input, result, test.expected)
		}
	}
}

func TestCountCharacterClasses(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"abc", 1},         // lowercase only
		{"ABC", 1},         // uppercase only
		{"123", 1},         // digits only
		{"!@#", 1},         // special only
		{"Aa", 2},          // upper + lower
		{"a1", 2},          // lower + digit
		{"A!", 2},          // upper + special
		{"Aa1", 3},         // upper + lower + digit
		{"Aa!", 3},         // upper + lower + special
		{"a1!", 3},         // lower + digit + special
		{"Aa1!", 4},        // all 4 classes
		{"P@ssw0rd", 4},    // all 4 classes
		{"hello world", 2}, // lower + special (space)
	}

	for _, test := range tests {
		result := countCharacterClasses(test.input)
		if result != test.expected {
			t.Errorf("countCharacterClasses(%q) = %d, want %d",
				test.input, result, test.expected)
		}
	}
}

// =============================================================================
// Edge Cases and Security Tests
// =============================================================================

func TestValidatePassword_Unicode(t *testing.T) {
	// Unicode passwords should work fine
	unicodePasswords := []string{
		"こんにちは世界",    // Japanese
		"Привет123",  // Russian + numbers
		"🔐🔑🗝️12345",  // Emojis + numbers
		"café☕️pass", // Mixed
	}

	for _, password := range unicodePasswords {
		if len(password) >= MinPasswordLength {
			err := ValidatePassword(password)
			if err != nil {
				t.Errorf("ValidatePassword(%q) should accept unicode: %v", password, err)
			}
		}
	}
}

func TestAnalyzePasswordStrength_CommonPatterns(t *testing.T) {
	// Common weak patterns (all should be at most Moderate)
	commonPatterns := []string{
		"password",
		"12345678",
		"qwertyui",
		"abcdefgh",
	}

	for _, password := range commonPatterns {
		strength := AnalyzePasswordStrength(password)
		if strength > PasswordModerate {
			t.Errorf("AnalyzePasswordStrength(%q) = %v, common pattern should not be Strong",
				password, strength)
		}
	}
}

func TestAnalyzePasswordStrength_Passphrases(t *testing.T) {
	// Long passphrases should be strong even with simple words
	passphrases := []string{
		"correct horse battery staple",   // 28 chars, spaces make it 4 classes
		"the quick brown fox jumps",      // 26 chars
		"i love zenon network very much", // 31 chars
	}

	for _, passphrase := range passphrases {
		strength := AnalyzePasswordStrength(passphrase)
		if strength < PasswordStrong {
			t.Errorf("AnalyzePasswordStrength(%q) = %v, long passphrase should be at least Strong",
				passphrase, strength)
		}
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkValidatePassword_Valid(b *testing.B) {
	password := "MySecureP@ssw0rd123"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ValidatePassword(password)
	}
}

func BenchmarkValidatePassword_Invalid(b *testing.B) {
	password := "short"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ValidatePassword(password)
	}
}

func BenchmarkAnalyzePasswordStrength(b *testing.B) {
	password := "MySecureP@ssw0rd123"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		AnalyzePasswordStrength(password)
	}
}

func BenchmarkCountCharacterClasses(b *testing.B) {
	password := "MySecureP@ssw0rd123"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		countCharacterClasses(password)
	}
}
