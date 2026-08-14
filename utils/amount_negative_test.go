package utils

import (
	"math/big"
	"testing"
)

// Regression test for finding #20: AddDecimals sliced the formatted string
// without accounting for a leading '-', embedding the sign inside the
// fraction ("0.000000-5") or emitting non-canonical output ("-.15").
func TestAddDecimalsFormatsNegativeAmounts(t *testing.T) {
	cases := []struct {
		amount   int64
		decimals int
		want     string
	}{
		{-5, 8, "-0.00000005"},
		{-12345, 8, "-0.00012345"},
		{-15000000, 8, "-0.15"},
		{-1, 8, "-0.00000001"},
		{-150000000, 8, "-1.5"},
		{5, 8, "0.00000005"},
		{150000000, 8, "1.5"},
	}
	for _, tc := range cases {
		got := AddDecimals(big.NewInt(tc.amount), tc.decimals)
		if got != tc.want {
			t.Errorf("AddDecimals(%d, %d) = %q, want %q", tc.amount, tc.decimals, got, tc.want)
		}
	}
}

// Negative amounts must round-trip through the documented inverse.
func TestAddDecimalsExtractDecimalsRoundTripNegative(t *testing.T) {
	for _, amount := range []int64{-1, -5, -15, -12345, -15000000} {
		formatted := AddDecimals(big.NewInt(amount), 8)
		parsed, err := ExtractDecimals(formatted, 8)
		if err != nil {
			t.Errorf("ExtractDecimals(AddDecimals(%d)) error = %v (formatted %q)", amount, err, formatted)
			continue
		}
		if parsed.Int64() != amount {
			t.Errorf("round-trip of %d = %d (formatted %q)", amount, parsed.Int64(), formatted)
		}
	}
}

// Regression test for finding #21: AddDecimals panicked with a slice-bounds
// error for negative decimals; it must behave like decimals == 0.
func TestAddDecimalsNegativeDecimalsDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("AddDecimals panicked on negative decimals: %v", r)
		}
	}()

	if got := AddDecimals(big.NewInt(1), -1); got != "1" {
		t.Errorf("AddDecimals(1, -1) = %q, want %q", got, "1")
	}
	if got := AddDecimals(big.NewInt(12345), -7); got != "12345" {
		t.Errorf("AddDecimals(12345, -7) = %q, want %q", got, "12345")
	}
}
