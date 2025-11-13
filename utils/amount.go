package utils

import (
	"fmt"
	"math/big"
	"strings"
)

// =============================================================================
// Amount Utilities
// =============================================================================

// ExtractDecimals parses a decimal string and converts it to big.Int
// Example: "1.5" with 8 decimals becomes 150000000
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
