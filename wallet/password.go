package wallet

import (
	"fmt"
	"unicode"
)

const (
	// MinPasswordLength is the minimum recommended password length for wallet encryption
	MinPasswordLength = 8
)

// PasswordStrength represents the strength level of a password
type PasswordStrength int

const (
	// PasswordWeak indicates a password that doesn't meet minimum requirements
	PasswordWeak PasswordStrength = iota
	// PasswordModerate indicates a password that meets basic requirements
	PasswordModerate
	// PasswordStrong indicates a password with good character diversity
	PasswordStrong
	// PasswordVeryStrong indicates a password with excellent character diversity
	PasswordVeryStrong
)

// String returns the string representation of PasswordStrength
func (s PasswordStrength) String() string {
	switch s {
	case PasswordWeak:
		return "Weak"
	case PasswordModerate:
		return "Moderate"
	case PasswordStrong:
		return "Strong"
	case PasswordVeryStrong:
		return "Very Strong"
	default:
		return "Unknown"
	}
}

// ValidatePassword checks if a password meets minimum security requirements.
//
// Requirements:
//   - Minimum 8 characters (configurable via MinPasswordLength)
//   - At least one character from any category (to prevent all-same-char passwords)
//
// This function returns an error if the password doesn't meet requirements.
// For a more detailed analysis, use AnalyzePasswordStrength.
//
// Example:
//
//	err := ValidatePassword("mypassword123")
//	if err != nil {
//	    fmt.Println("Password too weak:", err)
//	}
func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", MinPasswordLength)
	}

	// Check for all-same-character passwords (e.g., "aaaaaaaa")
	if isAllSameChar(password) {
		return fmt.Errorf("password cannot be all the same character")
	}

	return nil
}

// AnalyzePasswordStrength provides a detailed analysis of password strength.
//
// Strength scoring:
//   - Weak: < 8 chars OR all same character
//   - Moderate: >= 8 chars with one character class (e.g., all lowercase)
//   - Strong: >= 8 chars with two character classes (e.g., letters + numbers)
//   - Very Strong: >= 12 chars with three or more character classes
//
// Character classes considered:
//   - Lowercase letters (a-z)
//   - Uppercase letters (A-Z)
//   - Digits (0-9)
//   - Special characters (punctuation, symbols, spaces)
//
// Returns the strength level. Use ValidatePassword for simple pass/fail checking.
//
// Example:
//
//	strength := AnalyzePasswordStrength("MyP@ssw0rd123")
//	fmt.Println("Strength:", strength.String())  // Output: "Very Strong"
func AnalyzePasswordStrength(password string) PasswordStrength {
	if len(password) < MinPasswordLength || isAllSameChar(password) {
		return PasswordWeak
	}

	charClasses := countCharacterClasses(password)

	// Determine strength based on length and character diversity
	if len(password) >= 12 && charClasses >= 3 {
		return PasswordVeryStrong
	}

	if charClasses >= 2 {
		return PasswordStrong
	}

	return PasswordModerate
}

// isAllSameChar checks if all characters in the string are identical
func isAllSameChar(s string) bool {
	if len(s) == 0 {
		return false
	}

	first := rune(s[0])
	for _, r := range s {
		if r != first {
			return false
		}
	}
	return true
}

// countCharacterClasses counts how many different character classes are present
func countCharacterClasses(s string) int {
	var hasLower, hasUpper, hasDigit, hasSpecial bool

	for _, r := range s {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r) || unicode.IsSpace(r):
			hasSpecial = true
		}
	}

	count := 0
	if hasLower {
		count++
	}
	if hasUpper {
		count++
	}
	if hasDigit {
		count++
	}
	if hasSpecial {
		count++
	}

	return count
}
