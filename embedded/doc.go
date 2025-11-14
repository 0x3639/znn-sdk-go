// Package embedded provides embedded smart contract definitions, constants, and
// validation utilities for Zenon Network's protocol-level contracts.
//
// This package contains the core definitions for all embedded contracts including
// contract addresses, method signatures, and validation rules. It's primarily used
// internally by the api/embedded package but provides useful constants for developers.
//
// # Embedded Contract Addresses
//
// All embedded contracts have fixed addresses in the Zenon Network:
//
//	embedded.PlasmaContract   // z1qxemdeddedxplasmaxxxxxxxxxxxxxxxxsctrp
//	embedded.PillarContract   // z1qxemdeddedxpyllarxxxxxxxxxxxxxxxsy3fmg
//	embedded.TokenContract    // z1qxemdeddedxt0kenxxxxxxxxxxxxxxxxh9amk0
//	embedded.SentinelContract // z1qxemdeddedxsentynelxxxxxxxxxxxxxwy0r2r
//	embedded.StakeContract    // z1qxemdeddedxstakexxxxxxxxxxxxxxxxjv8v62
//	// ... and more
//
// # Contract Methods
//
// Method names and signatures for each contract:
//
//	// Pillar contract methods
//	embedded.PillarRegisterMethod   = "Register"
//	embedded.PillarDelegateMethod   = "Delegate"
//	embedded.PillarUndelegateMethod = "Undelegate"
//
//	// Token contract methods
//	embedded.TokenIssueMethod = "IssueToken"
//	embedded.TokenMintMethod  = "Mint"
//	embedded.TokenBurnMethod  = "Burn"
//
// # Contract Constants
//
// Important constants for contract operations:
//
//	// Minimum requirements
//	embedded.PillarRegisterZnnAmount = 15000 * constants.Decimals
//	embedded.PillarRegisterQsrAmount = 150000 * constants.Decimals
//
//	// Staking durations
//	embedded.StakeMinDuration = 1   // months
//	embedded.StakeMaxDuration = 12  // months
//
//	// Token limits
//	embedded.TokenMaxSupply = big.NewInt(2^255)
//	embedded.TokenMaxDecimals = 18
//
// # Validation Utilities
//
// Validate contract parameters before submission:
//
//	// Validate pillar name
//	if !embedded.IsValidPillarName(name) {
//	    log.Fatal("Invalid pillar name")
//	}
//
//	// Validate token symbol
//	if !embedded.IsValidTokenSymbol(symbol) {
//	    log.Fatal("Invalid token symbol")
//	}
//
//	// Validate stake duration
//	if !embedded.IsValidStakeDuration(months) {
//	    log.Fatal("Stake duration must be 1-12 months")
//	}
//
// # Contract Definitions
//
// Each embedded contract has a definition struct containing:
//   - Contract address
//   - Available methods
//   - Parameter requirements
//   - Validation rules
//
// Example:
//
//	pillarDef := embedded.GetContractDefinition(embedded.PillarContract)
//	fmt.Println("Methods:", pillarDef.Methods)
//
// # Usage in Contract Calls
//
// Typically used indirectly through api/embedded:
//
//	// High-level API uses embedded package internally
//	template := client.PillarApi.Register(...)
//
//	// Under the hood:
//	// - Validates parameters using embedded.IsValidPillarName, etc.
//	// - Uses embedded.PillarContract address
//	// - Encodes using embedded.PillarRegisterMethod
//
// # Direct Usage
//
// For advanced use cases or debugging:
//
//	// Check if address is an embedded contract
//	if embedded.IsEmbeddedContract(address) {
//	    contractName := embedded.GetContractName(address)
//	    fmt.Println("Contract:", contractName)
//	}
//
// For more information, see https://pkg.go.dev/github.com/0x3639/znn-sdk-go/embedded
package embedded
