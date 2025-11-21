package utils

import (
	"fmt"
	"math/big"
	"strings"
)

// =============================================================================
// Amount Utilities
// =============================================================================

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
//   - Amount is negative
//
// Behavior:
//   - Integer amounts: "100" with 8 decimals becomes 10000000000
//   - Decimal amounts: "1.5" with 8 decimals becomes 150000000
//   - Excess decimals are truncated: "1.123456789" with 8 decimals becomes 112345678
//   - Missing decimals are padded: "1.5" with 8 decimals becomes "1.50000000"
//   - Negative amounts are rejected (returns error)
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
// Security:
//   - Validates against negative amounts to prevent balance underflow
//   - Dual validation in both integer and decimal paths
func ExtractDecimals(amount string, decimals int) (*big.Int, error) {
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

		// Validate that amount is not negative
		if intPart.Sign() < 0 {
			return nil, fmt.Errorf("amount cannot be negative: %s", amount)
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

	// Validate that amount is not negative
	if result.Sign() < 0 {
		return nil, fmt.Errorf("amount cannot be negative: %s", amount)
	}

	return result, nil
}

// AddDecimals converts big.Int to decimal string representation
// Example: 150000000 with 8 decimals becomes "1.5"
func AddDecimals(number *big.Int, decimals int) string {
	if number.Sign() == 0 {
		return "0"
	}

	// Convert to string
	str := number.String()

	// If decimals is 0, return as-is
	if decimals == 0 {
		return str
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

	return result
}
