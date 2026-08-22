package utils

import (
	"fmt"
	"math/big"
	"strings"
)

// =============================================================================
// Amount Utilities
// =============================================================================

// MaxDecimals is the largest decimal precision accepted by ExtractDecimals and
// AddDecimals.
//
// Zenon token standards use at most 18 decimals; 32 leaves generous headroom
// while preventing an untrusted decimals value from requesting an enormous
// string or big.Int allocation (CWE-400).
const MaxDecimals = 32

// ExtractDecimals parses a decimal string amount and converts it to a big.Int
// in base units according to the specified number of decimal places.
//
// This function is used to convert human-readable token amounts (e.g., "1.5 ZNN")
// into the integer representation required by the Zenon protocol.
//
// Parameters:
//   - amount: Decimal string representation (e.g., "1.5", "100", "0.00000001")
//   - decimals: Number of decimal places (e.g., 8 for ZNN/QSR)
//
// Returns the amount in base units as a big.Int, or an error if:
//   - Amount is empty (unless decimals is 0)
//   - Amount has invalid decimal format (multiple decimal points)
//   - Amount contains non-numeric characters
//   - Decimals is negative
//
// Behavior:
//   - Integer amounts: "100" with 8 decimals becomes 10000000000
//   - Decimal amounts: "1.5" with 8 decimals becomes 150000000
//   - Excess decimals are truncated: "1.123456789" with 8 decimals becomes 112345678
//   - Negative excess decimals truncate toward zero: "-1.239" with 2 decimals becomes -123
//   - Missing decimals are padded: "1.5" with 8 decimals becomes "1.50000000"
//   - Negative amounts preserve their sign
//
// Example - ZNN transfer (8 decimals):
//
//	amount, err := ExtractDecimals("1.5", 8)
//	// Returns: 150000000 (1.5 * 10^8)
//
// Example - Custom token (2 decimals):
//
//	amount, err := ExtractDecimals("99.99", 2)
//	// Returns: 9999 (99.99 * 10^2)
//
// Example - Signed calculation with truncation toward zero:
//
//	amount, err := ExtractDecimals("-123.456", 2)
//	// Returns: -12345
//
// Note: Signed conversion is useful for deltas, quotes, and calculations. A
// protocol operation that requires a non-negative transfer amount must enforce
// that separate domain rule before constructing an account block.
func ExtractDecimals(amount string, decimals int) (*big.Int, error) {
	if decimals < 0 {
		return nil, fmt.Errorf("decimals cannot be negative: %d", decimals)
	}
	if decimals > MaxDecimals {
		return nil, fmt.Errorf("decimals %d exceeds maximum %d", decimals, MaxDecimals)
	}
	if amount == "" {
		if decimals == 0 {
			return big.NewInt(0), nil
		}
		return nil, fmt.Errorf("amount cannot be empty")
	}

	// No decimal point - just append zeros
	if !strings.Contains(amount, ".") {
		// Parse the integer part
		intPart, ok := new(big.Int).SetString(amount, 10)
		if !ok {
			return nil, fmt.Errorf("invalid amount format: %s", amount)
		}

		// Multiply by 10^decimals
		multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
		result := new(big.Int).Mul(intPart, multiplier)
		return result, nil
	}

	// Has decimal point - split and process
	parts := strings.Split(amount, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid decimal format: %s", amount)
	}

	intPart := parts[0]
	decPart := parts[1]

	// Truncate or pad the decimal part
	if len(decPart) > decimals {
		decPart = decPart[:decimals]
	} else if len(decPart) < decimals {
		decPart = decPart + strings.Repeat("0", decimals-len(decPart))
	}

	// Combine integer and decimal parts
	combined := intPart + decPart
	result, ok := new(big.Int).SetString(combined, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount format: %s", amount)
	}

	return result, nil
}

// AddDecimals converts big.Int to decimal string representation
// Example: 150000000 with 8 decimals becomes "1.5"
//
// Negative amounts are formatted with a leading '-' (e.g. -5 with 8 decimals
// becomes "-0.00000005"). A decimals value <= 0 or above MaxDecimals returns
// the plain integer string instead of slicing out of range or allocating an
// attacker-sized padding string.
func AddDecimals(number *big.Int, decimals int) string {
	if number.Sign() == 0 {
		return "0"
	}

	// Convert magnitude to string; the sign is re-attached at the end so it
	// can never end up embedded in the fraction (e.g. "0.000000-5").
	sign := ""
	if number.Sign() < 0 {
		sign = "-"
	}
	str := new(big.Int).Abs(number).String()

	// If decimals is 0 (or invalid / above MaxDecimals), return as-is
	if decimals <= 0 || decimals > MaxDecimals {
		return sign + str
	}

	// Pad with zeros if needed
	if len(str) <= decimals {
		str = strings.Repeat("0", decimals-len(str)+1) + str
	}

	// Insert decimal point
	insertPos := len(str) - decimals
	result := str[:insertPos] + "." + str[insertPos:]

	// Strip trailing zeros after decimal point
	result = strings.TrimRight(result, "0")
	result = strings.TrimRight(result, ".")

	return sign + result
}
