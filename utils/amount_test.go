package utils

import (
	"math/big"
	"strings"
	"testing"
)

// =============================================================================
// ExtractDecimals Tests
// =============================================================================

func TestExtractDecimals_NoDecimal(t *testing.T) {
	result, err := ExtractDecimals("100", 8)
	if err != nil {
		t.Fatalf("ExtractDecimals() error = %v", err)
	}

	expected := big.NewInt(10000000000) // 100 * 10^8
	if result.Cmp(expected) != 0 {
		t.Errorf("ExtractDecimals(\"100\", 8) = %s, want %s", result.String(), expected.String())
	}
}

func TestExtractDecimals_WithDecimal(t *testing.T) {
	result, err := ExtractDecimals("1.5", 8)
	if err != nil {
		t.Fatalf("ExtractDecimals() error = %v", err)
	}

	expected := big.NewInt(150000000) // 1.5 * 10^8
	if result.Cmp(expected) != 0 {
		t.Errorf("ExtractDecimals(\"1.5\", 8) = %s, want %s", result.String(), expected.String())
	}
}

func TestExtractDecimals_SmallAmount(t *testing.T) {
	result, err := ExtractDecimals("0.00000001", 8)
	if err != nil {
		t.Fatalf("ExtractDecimals() error = %v", err)
	}

	expected := big.NewInt(1) // 0.00000001 * 10^8 = 1
	if result.Cmp(expected) != 0 {
		t.Errorf("ExtractDecimals(\"0.00000001\", 8) = %s, want %s", result.String(), expected.String())
	}
}

func TestExtractDecimals_TruncatesExtraDecimals(t *testing.T) {
	result, err := ExtractDecimals("1.123456789", 8)
	if err != nil {
		t.Fatalf("ExtractDecimals() error = %v", err)
	}

	expected := big.NewInt(112345678) // Truncated to 8 decimals
	if result.Cmp(expected) != 0 {
		t.Errorf("ExtractDecimals(\"1.123456789\", 8) = %s, want %s", result.String(), expected.String())
	}
}

func TestExtractDecimals_PadsShortDecimals(t *testing.T) {
	result, err := ExtractDecimals("1.5", 8)
	if err != nil {
		t.Fatalf("ExtractDecimals() error = %v", err)
	}

	expected := big.NewInt(150000000) // Padded to 8 decimals
	if result.Cmp(expected) != 0 {
		t.Errorf("ExtractDecimals(\"1.5\", 8) = %s, want %s", result.String(), expected.String())
	}
}

func TestExtractDecimals_Empty(t *testing.T) {
	_, err := ExtractDecimals("", 8)
	if err == nil {
		t.Error("ExtractDecimals(\"\", 8) should return error")
	}
}

func TestExtractDecimals_Invalid(t *testing.T) {
	testCases := []string{
		"abc",
		"1.2.3",
		"1..2",
	}

	for _, tc := range testCases {
		_, err := ExtractDecimals(tc, 8)
		if err == nil {
			t.Errorf("ExtractDecimals(%s, 8) should return error", tc)
		}
	}
}

func TestExtractDecimals_NegativeAmount_NoDecimal(t *testing.T) {
	_, err := ExtractDecimals("-100", 8)
	if err == nil {
		t.Error("ExtractDecimals(\"-100\", 8) should return error for negative amount")
	}
	if err != nil && !contains(err.Error(), "cannot be negative") {
		t.Errorf("Expected 'cannot be negative' error, got: %v", err)
	}
}

func TestExtractDecimals_NegativeAmount_WithDecimal(t *testing.T) {
	_, err := ExtractDecimals("-1.5", 8)
	if err == nil {
		t.Error("ExtractDecimals(\"-1.5\", 8) should return error for negative amount")
	}
	if err != nil && !contains(err.Error(), "cannot be negative") {
		t.Errorf("Expected 'cannot be negative' error, got: %v", err)
	}
}

func TestExtractDecimals_NegativeZero(t *testing.T) {
	// Edge case: "-0" or "-0.0" should be treated as zero (not negative)
	result, err := ExtractDecimals("-0", 8)
	if err != nil {
		// Note: big.Int parses "-0" as 0, so Sign() == 0
		// This test documents expected behavior
		t.Logf("ExtractDecimals(\"-0\", 8) error = %v (acceptable)", err)
	}
	if result != nil && result.Sign() != 0 {
		t.Error("ExtractDecimals(\"-0\", 8) should result in zero")
	}
}

func TestExtractDecimals_NegativeSmallAmount(t *testing.T) {
	_, err := ExtractDecimals("-0.00000001", 8)
	if err == nil {
		t.Error("ExtractDecimals(\"-0.00000001\", 8) should return error for negative amount")
	}
	if err != nil && !contains(err.Error(), "cannot be negative") {
		t.Errorf("Expected 'cannot be negative' error, got: %v", err)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// =============================================================================
// AddDecimals Tests
// =============================================================================

func TestAddDecimals_WholeNumber(t *testing.T) {
	input := big.NewInt(100000000) // 1.0 with 8 decimals
	result := AddDecimals(input, 8)

	expected := "1"
	if result != expected {
		t.Errorf("AddDecimals(%s, 8) = %s, want %s", input.String(), result, expected)
	}
}

func TestAddDecimals_Decimal(t *testing.T) {
	input := big.NewInt(150000000) // 1.5 with 8 decimals
	result := AddDecimals(input, 8)

	expected := "1.5"
	if result != expected {
		t.Errorf("AddDecimals(%s, 8) = %s, want %s", input.String(), result, expected)
	}
}

func TestAddDecimals_SmallAmount(t *testing.T) {
	input := big.NewInt(1) // 0.00000001 with 8 decimals
	result := AddDecimals(input, 8)

	expected := "0.00000001"
	if result != expected {
		t.Errorf("AddDecimals(%s, 8) = %s, want %s", input.String(), result, expected)
	}
}

func TestAddDecimals_Zero(t *testing.T) {
	input := big.NewInt(0)
	result := AddDecimals(input, 8)

	expected := "0"
	if result != expected {
		t.Errorf("AddDecimals(0, 8) = %s, want %s", result, expected)
	}
}

func TestAddDecimals_NoDecimals(t *testing.T) {
	input := big.NewInt(100)
	result := AddDecimals(input, 0)

	expected := "100"
	if result != expected {
		t.Errorf("AddDecimals(100, 0) = %s, want %s", result, expected)
	}
}

func TestAddDecimals_StripsTrailingZeros(t *testing.T) {
	input := big.NewInt(120000000) // 1.2 with trailing zeros
	result := AddDecimals(input, 8)

	expected := "1.2"
	if result != expected {
		t.Errorf("AddDecimals(%s, 8) = %s, want %s", input.String(), result, expected)
	}
}

// =============================================================================
// Round Trip Tests
// =============================================================================

func TestAmountRoundTrip(t *testing.T) {
	testCases := []string{
		"1",
		"1.5",
		"0.00000001",
		"100",
		"99.99999999",
	}

	for _, tc := range testCases {
		// Extract decimals
		extracted, err := ExtractDecimals(tc, 8)
		if err != nil {
			t.Fatalf("ExtractDecimals(%s) error = %v", tc, err)
		}

		// Add decimals back
		result := AddDecimals(extracted, 8)

		// Should match original (after normalization)
		original, _ := ExtractDecimals(tc, 8)
		reconstructed, _ := ExtractDecimals(result, 8)

		if original.Cmp(reconstructed) != 0 {
			t.Errorf("Round trip failed for %s: got %s", tc, result)
		}
	}
}
