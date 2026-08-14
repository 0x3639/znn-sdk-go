package embedded

import (
	"strings"
	"testing"
)

// Regression test for finding #17: ValidateTokenDomain enforced no maximum
// length, so arbitrarily long regex-valid domains passed although the
// protocol (go-zenon constants.TokenDomainLengthMax) caps domains at 128.
func TestValidateTokenDomain_TooLong(t *testing.T) {
	longDomain := strings.Repeat("ab.", 4000) + "com" // 12,003 chars, regex-valid
	if err := ValidateTokenDomain(longDomain); err == nil {
		t.Error("ValidateTokenDomain accepted a 12,003-character domain")
	}

	// Exactly at the protocol limit must still pass.
	atLimit := strings.Repeat("ab.", 41) + "renec" // 41*3+5 = 128 chars
	if len(atLimit) != TokenDomainMaxLength {
		t.Fatalf("bad vector: %d chars, want %d", len(atLimit), TokenDomainMaxLength)
	}
	if err := ValidateTokenDomain(atLimit); err != nil {
		t.Errorf("ValidateTokenDomain rejected a %d-character domain: %v", TokenDomainMaxLength, err)
	}

	if err := ValidateTokenDomain("zenon.network"); err != nil {
		t.Errorf("ValidateTokenDomain rejected a normal domain: %v", err)
	}
}
