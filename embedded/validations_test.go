package embedded

import (
	"strings"
	"testing"
)

// =============================================================================
// Token Name Validation Tests
// =============================================================================

func TestValidateTokenName_Valid(t *testing.T) {
	validNames := []string{
		"MyToken",
		"token123",
		"Test-Token",
		"my.token",
		"test_token",
		"Token123",
		"A",
		"ABC123",
	}

	for _, name := range validNames {
		err := ValidateTokenName(name)
		if err != nil {
			t.Errorf("ValidateTokenName(%s) returned error: %v", name, err)
		}
	}
}

func TestValidateTokenName_Empty(t *testing.T) {
	err := ValidateTokenName("")
	if err == nil {
		t.Error("ValidateTokenName(\"\") should return error for empty string")
	}
}

func TestValidateTokenName_TooLong(t *testing.T) {
	// Create a name longer than TokenNameMaxLength (40)
	longName := strings.Repeat("a", TokenNameMaxLength+1)
	err := ValidateTokenName(longName)
	if err == nil {
		t.Error("ValidateTokenName() should return error for name too long")
	}
}

func TestValidateTokenName_InvalidPattern(t *testing.T) {
	invalidNames := []string{
		"-invalid",
		"invalid-",
		".invalid",
		"invalid.",
		"invalid--name",
		"invalid..name",
		"invalid__name",
		"token with spaces",
		"token@symbol",
	}

	for _, name := range invalidNames {
		err := ValidateTokenName(name)
		if err == nil {
			t.Errorf("ValidateTokenName(%s) should return error for invalid pattern", name)
		}
	}
}

// =============================================================================
// Token Symbol Validation Tests
// =============================================================================

func TestValidateTokenSymbol_Valid(t *testing.T) {
	validSymbols := []string{
		"TEST",
		"TOKEN",
		"ABC",
		"TOKEN123",
		"XYZ",
	}

	for _, symbol := range validSymbols {
		err := ValidateTokenSymbol(symbol)
		if err != nil {
			t.Errorf("ValidateTokenSymbol(%s) returned error: %v", symbol, err)
		}
	}
}

func TestValidateTokenSymbol_Empty(t *testing.T) {
	err := ValidateTokenSymbol("")
	if err == nil {
		t.Error("ValidateTokenSymbol(\"\") should return error for empty string")
	}
}

func TestValidateTokenSymbol_Reserved(t *testing.T) {
	reservedSymbols := []string{"ZNN", "QSR"}

	for _, symbol := range reservedSymbols {
		err := ValidateTokenSymbol(symbol)
		if err == nil {
			t.Errorf("ValidateTokenSymbol(%s) should return error for reserved symbol", symbol)
		}
	}
}

func TestValidateTokenSymbol_TooLong(t *testing.T) {
	// Create a symbol longer than TokenSymbolMaxLength (10)
	longSymbol := strings.Repeat("A", TokenSymbolMaxLength+1)
	err := ValidateTokenSymbol(longSymbol)
	if err == nil {
		t.Error("ValidateTokenSymbol() should return error for symbol too long")
	}
}

func TestValidateTokenSymbol_InvalidPattern(t *testing.T) {
	invalidSymbols := []string{
		"test",        // lowercase
		"Test",        // mixed case
		"TOKEN-1",     // hyphen
		"TOKEN_1",     // underscore
		"token.sym",   // dot
		"TOKEN SPACE", // space
	}

	for _, symbol := range invalidSymbols {
		err := ValidateTokenSymbol(symbol)
		if err == nil {
			t.Errorf("ValidateTokenSymbol(%s) should return error for invalid pattern", symbol)
		}
	}
}

// =============================================================================
// Token Domain Validation Tests
// =============================================================================

func TestValidateTokenDomain_Valid(t *testing.T) {
	validDomains := []string{
		"example.com",
		"test.example.com",
		"my-domain.io",
		"test123.org",
		"subdomain.example.co.uk",
	}

	for _, domain := range validDomains {
		err := ValidateTokenDomain(domain)
		if err != nil {
			t.Errorf("ValidateTokenDomain(%s) returned error: %v", domain, err)
		}
	}
}

func TestValidateTokenDomain_Empty(t *testing.T) {
	err := ValidateTokenDomain("")
	if err == nil {
		t.Error("ValidateTokenDomain(\"\") should return error for empty string")
	}
}

func TestValidateTokenDomain_Invalid(t *testing.T) {
	invalidDomains := []string{
		"invalid",
		"-invalid.com",
		"invalid-.com",
		".com",
		"com.",
		"invalid..com",
		"invalid .com",
	}

	for _, domain := range invalidDomains {
		err := ValidateTokenDomain(domain)
		if err == nil {
			t.Errorf("ValidateTokenDomain(%s) should return error for invalid domain", domain)
		}
	}
}

// =============================================================================
// Pillar Name Validation Tests
// =============================================================================

func TestValidatePillarName_Valid(t *testing.T) {
	validNames := []string{
		"MyPillar",
		"pillar-1",
		"my.pillar",
		"test_pillar",
		"Pillar123",
		"A",
		"ABC",
	}

	for _, name := range validNames {
		err := ValidatePillarName(name)
		if err != nil {
			t.Errorf("ValidatePillarName(%s) returned error: %v", name, err)
		}
	}
}

func TestValidatePillarName_Empty(t *testing.T) {
	err := ValidatePillarName("")
	if err == nil {
		t.Error("ValidatePillarName(\"\") should return error for empty string")
	}
}

func TestValidatePillarName_TooLong(t *testing.T) {
	// Create a name longer than PillarNameMaxLength (40)
	longName := strings.Repeat("a", PillarNameMaxLength+1)
	err := ValidatePillarName(longName)
	if err == nil {
		t.Error("ValidatePillarName() should return error for name too long")
	}
}

func TestValidatePillarName_InvalidPattern(t *testing.T) {
	invalidNames := []string{
		"-invalid",
		"invalid-",
		".invalid",
		"invalid.",
		"invalid--name",
		"invalid..name",
		"invalid__name",
		"pillar with spaces",
		"pillar@symbol",
	}

	for _, name := range invalidNames {
		err := ValidatePillarName(name)
		if err == nil {
			t.Errorf("ValidatePillarName(%s) should return error for invalid pattern", name)
		}
	}
}

// =============================================================================
// Project Name Validation Tests
// =============================================================================

func TestValidateProjectName_Valid(t *testing.T) {
	validNames := []string{
		"My Project",
		"Test",
		"Project 123",
		"A",
		strings.Repeat("a", ProjectNameMaxLength), // Exactly max length
	}

	for _, name := range validNames {
		err := ValidateProjectName(name)
		if err != nil {
			t.Errorf("ValidateProjectName(%s) returned error: %v", name, err)
		}
	}
}

func TestValidateProjectName_Empty(t *testing.T) {
	err := ValidateProjectName("")
	if err == nil {
		t.Error("ValidateProjectName(\"\") should return error for empty string")
	}
}

func TestValidateProjectName_TooLong(t *testing.T) {
	// Create a name longer than ProjectNameMaxLength (30)
	longName := strings.Repeat("a", ProjectNameMaxLength+1)
	err := ValidateProjectName(longName)
	if err == nil {
		t.Error("ValidateProjectName() should return error for name too long")
	}
}

// =============================================================================
// Project Description Validation Tests
// =============================================================================

func TestValidateProjectDescription_Valid(t *testing.T) {
	validDescriptions := []string{
		"This is a valid project description",
		"Test",
		"A",
		strings.Repeat("a", ProjectDescriptionMaxLength), // Exactly max length
	}

	for _, desc := range validDescriptions {
		err := ValidateProjectDescription(desc)
		if err != nil {
			t.Errorf("ValidateProjectDescription(%s) returned error: %v", desc, err)
		}
	}
}

func TestValidateProjectDescription_Empty(t *testing.T) {
	err := ValidateProjectDescription("")
	if err == nil {
		t.Error("ValidateProjectDescription(\"\") should return error for empty string")
	}
}

func TestValidateProjectDescription_TooLong(t *testing.T) {
	// Create a description longer than ProjectDescriptionMaxLength (240)
	longDesc := strings.Repeat("a", ProjectDescriptionMaxLength+1)
	err := ValidateProjectDescription(longDesc)
	if err == nil {
		t.Error("ValidateProjectDescription() should return error for description too long")
	}
}

// =============================================================================
// Boundary Tests
// =============================================================================

func TestValidateTokenName_AtMaxLength(t *testing.T) {
	// Exactly at max length should be valid
	name := strings.Repeat("a", TokenNameMaxLength)
	err := ValidateTokenName(name)
	if err != nil {
		t.Errorf("ValidateTokenName() should accept name at max length (%d)", TokenNameMaxLength)
	}
}

func TestValidateTokenSymbol_AtMaxLength(t *testing.T) {
	// Exactly at max length should be valid
	symbol := strings.Repeat("A", TokenSymbolMaxLength)
	err := ValidateTokenSymbol(symbol)
	if err != nil {
		t.Errorf("ValidateTokenSymbol() should accept symbol at max length (%d)", TokenSymbolMaxLength)
	}
}

func TestValidatePillarName_AtMaxLength(t *testing.T) {
	// Exactly at max length should be valid
	name := strings.Repeat("a", PillarNameMaxLength)
	err := ValidatePillarName(name)
	if err != nil {
		t.Errorf("ValidatePillarName() should accept name at max length (%d)", PillarNameMaxLength)
	}
}
