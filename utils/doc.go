// Package utils provides utility functions for working with Zenon Network data types,
// including amount conversions, byte operations, and block utilities.
//
// This package contains helper functions that simplify common operations when working
// with the Zenon SDK, particularly for handling amounts, addresses, and data conversions.
//
// # Amount Conversions
//
// Zenon uses 8 decimal places for token amounts. Utilities help convert between
// user-facing amounts and base units:
//
//	// Convert 1.5 ZNN to base units
//	amount := utils.AmountToRaw("1.5", 8)  // Returns *big.Int representing 150000000
//
//	// Convert base units to user-facing amount
//	userAmount := utils.RawToAmount(bigIntAmount, 8)  // Returns "1.50000000"
//
// # Byte Operations
//
// Common byte slice operations:
//
//	// Reverse bytes
//	reversed := utils.ReverseBytes(data)
//
//	// Pad bytes to specific length
//	padded := utils.PadBytes(data, 32)
//
// # Block Utilities
//
// Helper functions for account blocks:
//
//	// Check if block hash is valid
//	if utils.IsValidHash(hashString) {
//	    // Process hash
//	}
//
//	// Check if address is valid
//	if utils.IsValidAddress(addressString) {
//	    // Process address
//	}
//
// # Common Patterns
//
// Working with token amounts:
//
//	// User wants to send 10.5 ZNN
//	amountStr := "10.5"
//	amount := utils.AmountToRaw(amountStr, 8)
//
//	// Create transaction with base units
//	template := client.LedgerApi.SendTemplate(
//	    toAddress,
//	    types.ZnnTokenStandard,
//	    amount,
//	    []byte{},
//	)
//
// # Data Encoding
//
// Utilities for encoding transaction data:
//
//	// Encode string data for transaction
//	data := utils.EncodeString("Hello Zenon")
//
//	// Decode data from transaction
//	message := utils.DecodeString(block.Data)
//
// For more information, see https://pkg.go.dev/github.com/0x3639/znn-sdk-go/utils
package utils
