// Package abi provides Application Binary Interface (ABI) encoding and decoding
// functionality for Zenon Network embedded smart contracts.
//
// The ABI package handles serialization and deserialization of contract calls, enabling
// interaction with protocol-level embedded contracts. This is primarily used internally
// by the embedded contract APIs but can be used directly for custom contract interactions.
//
// # Basic Concepts
//
// Embedded contracts in Zenon Network use ABI encoding to:
//   - Encode method calls into transaction data
//   - Decode method parameters from transaction data
//   - Serialize complex data structures
//
// # Contract Method Encoding
//
// Encode a contract method call:
//
//	// Define method parameters
//	params := []interface{}{
//	    "pillarName",
//	    big.NewInt(15000),
//	    types.ZnnTokenStandard,
//	}
//
//	// Encode method call
//	data, err := abi.EncodeMethod("Register", params...)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Contract Response Decoding
//
// Decode contract response data:
//
//	// Define expected return types
//	var result struct {
//	    Name   string
//	    Owner  types.Address
//	    Amount *big.Int
//	}
//
//	// Decode response
//	err := abi.DecodeResponse(responseData, &result)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Common Data Types
//
// The ABI package handles encoding/decoding of:
//   - Basic types: int8 through int256, uint8 through uint256, bool, string
//   - Fixed bytes: bytes1 through bytes32
//   - Big integers: *big.Int for large numbers
//   - Addresses: types.Address
//   - Token standards: types.ZenonTokenStandard
//   - Hashes: types.Hash
//   - Complex structures and arrays
//
// # Canonical Validation
//
// Encoding requires exactly one value for each declared parameter. Signed and
// unsigned integers must fit their declared bit width, fixed-byte values must
// have exactly their declared length, and boolean inputs must have Go type bool.
// Decoding rejects integer words outside the declared width, non-zero fixed-byte
// padding, and boolean words other than canonical zero or one. These checks prevent
// silent truncation and ambiguous wire encodings before a contract call is sent.
//
// # Internal Usage
//
// Most developers don't need to use the ABI package directly, as the embedded
// contract APIs handle encoding automatically:
//
//	// High-level API (recommended)
//	template := client.PillarApi.Register(name, producerAddress, rewardAddress, ...)
//
//	// Under the hood, this uses:
//	// data := abi.EncodeMethod("Register", name, producerAddress, ...)
//	// template.Data = data
//
// # Advanced Usage
//
// For custom contract interactions or debugging, you can use the ABI package directly:
//
//	// Create custom contract call
//	contractAddress := types.ParseAddressPanic("z1qxemdeddedxxxxxxxxxxxxxxxxxxxxxxxxxxx")
//	data, _ := abi.EncodeMethod("CustomMethod", param1, param2)
//
//	template := &nom.AccountBlock{
//	    Address:       myAddress,
//	    ToAddress:     contractAddress,
//	    Data:          data,
//	    TokenStandard: types.ZnnTokenStandard,
//	    Amount:        big.NewInt(0),
//	}
//
// For more information, see https://pkg.go.dev/github.com/0x3639/znn-sdk-go/abi
package abi
