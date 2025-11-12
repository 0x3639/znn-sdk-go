package abi

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// Constants
const (
	Int32Size = 32 // Size of encoded values in bytes
)

// AbiType is the interface that all ABI types must implement
type AbiType interface {
	GetName() string
	GetCanonicalName() string
	Encode(value interface{}) ([]byte, error)
	Decode(encoded []byte, offset int) (interface{}, error)
	GetFixedSize() int
	IsDynamicType() bool
}

// =============================================================================
// Base Types
// =============================================================================

// baseType provides common functionality for all ABI types
type baseType struct {
	name string
}

func (bt *baseType) GetName() string {
	return bt.name
}

func (bt *baseType) GetCanonicalName() string {
	return bt.name
}

func (bt *baseType) GetFixedSize() int {
	return Int32Size
}

func (bt *baseType) IsDynamicType() bool {
	return false
}

// =============================================================================
// Numeric Type Base
// =============================================================================

// NumericType is the base for all numeric types (int, uint)
type NumericType struct {
	baseType
}

// EncodeInternal converts various value types to big.Int
func (nt *NumericType) EncodeInternal(value interface{}) (*big.Int, error) {
	switch v := value.(type) {
	case string:
		s := strings.ToLower(strings.TrimSpace(v))
		radix := 10

		// Handle hex strings
		if strings.HasPrefix(s, "0x") {
			s = s[2:]
			radix = 16
		} else if strings.ContainsAny(s, "abcdef") {
			radix = 16
		}

		bigInt := new(big.Int)
		if _, ok := bigInt.SetString(s, radix); !ok {
			return nil, fmt.Errorf("invalid numeric string: %s", v)
		}
		return bigInt, nil

	case *big.Int:
		return v, nil

	case int:
		return big.NewInt(int64(v)), nil

	case int8:
		return big.NewInt(int64(v)), nil

	case int16:
		return big.NewInt(int64(v)), nil

	case int32:
		return big.NewInt(int64(v)), nil

	case int64:
		return big.NewInt(v), nil

	case uint:
		return new(big.Int).SetUint64(uint64(v)), nil

	case uint8:
		return new(big.Int).SetUint64(uint64(v)), nil

	case uint16:
		return new(big.Int).SetUint64(uint64(v)), nil

	case uint32:
		return new(big.Int).SetUint64(uint64(v)), nil

	case uint64:
		return new(big.Int).SetUint64(v), nil

	case []byte:
		return new(big.Int).SetBytes(v), nil

	default:
		return nil, fmt.Errorf("unsupported value type for numeric encoding: %T", value)
	}
}

// =============================================================================
// IntType - Signed Integer Type
// =============================================================================

// IntType represents signed integer types (int8 to int256)
type IntType struct {
	NumericType
	size int // Size in bits
}

// NewIntType creates a new signed integer type
func NewIntType(name string) (*IntType, error) {
	it := &IntType{}
	it.name = name

	// Parse size from name (e.g., "int256" -> 256)
	if name == "int" {
		it.size = 256
	} else if strings.HasPrefix(name, "int") {
		sizeStr := strings.TrimPrefix(name, "int")
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid int type name: %s", name)
		}

		// Validate size (must be 8-256 in increments of 8)
		if size < 8 || size > 256 || size%8 != 0 {
			return nil, fmt.Errorf("invalid int size: %d (must be 8-256 in increments of 8)", size)
		}
		it.size = size
	} else {
		return nil, fmt.Errorf("invalid int type name: %s", name)
	}

	return it, nil
}

// GetCanonicalName returns the canonical name (int defaults to int256)
func (it *IntType) GetCanonicalName() string {
	if it.name == "int" {
		return "int256"
	}
	return it.name
}

// Encode encodes a signed integer value
func (it *IntType) Encode(value interface{}) ([]byte, error) {
	bigInt, err := it.EncodeInternal(value)
	if err != nil {
		return nil, err
	}
	return EncodeIntBig(bigInt), nil
}

// Decode decodes a signed integer value
func (it *IntType) Decode(encoded []byte, offset int) (interface{}, error) {
	return DecodeInt(encoded, offset)
}

// EncodeInt encodes an int to 32 bytes
func EncodeInt(i int) []byte {
	return EncodeIntBig(big.NewInt(int64(i)))
}

// EncodeIntBig encodes a big.Int to 32 bytes (signed, two's complement)
func EncodeIntBig(bigInt *big.Int) []byte {
	return bigIntToBytesSigned(bigInt, Int32Size)
}

// DecodeInt decodes a signed integer from encoded bytes at offset
func DecodeInt(encoded []byte, offset int) (*big.Int, error) {
	if len(encoded) < offset+Int32Size {
		return nil, fmt.Errorf("insufficient bytes for decoding int")
	}

	bytes := encoded[offset : offset+Int32Size]

	// Convert bytes to big.Int (handling sign)
	bigInt := new(big.Int).SetBytes(bytes)

	// Check if negative (MSB set)
	if bytes[0]&0x80 != 0 {
		// Two's complement for negative numbers
		// Create a mask of all 1s for 256 bits
		max := new(big.Int).Lsh(big.NewInt(1), 256) // 2^256
		bigInt.Sub(bigInt, max)
	}

	return bigInt, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// bigIntToBytesSigned converts a big.Int to a fixed-size byte array (signed, two's complement)
func bigIntToBytesSigned(b *big.Int, numBytes int) []byte {
	// Determine fill byte based on sign
	fillByte := byte(0x00)
	if b.Sign() < 0 {
		fillByte = 0xFF
	}

	// Create byte array filled with appropriate byte
	bytes := make([]byte, numBytes)
	for i := range bytes {
		bytes[i] = fillByte
	}

	// Get big.Int bytes
	biBytes := b.Bytes()

	// For negative numbers, we need two's complement
	if b.Sign() < 0 {
		// Convert to two's complement
		// First, get absolute value
		absVal := new(big.Int).Abs(b)
		// Create a value with numBytes*8 bits set
		maxVal := new(big.Int).Lsh(big.NewInt(1), uint(numBytes*8))
		// Subtract absolute value from 2^(numBytes*8)
		twosComp := new(big.Int).Sub(maxVal, absVal)
		biBytes = twosComp.Bytes()
	}

	// Calculate where to start copying
	start := 0
	length := len(biBytes)

	// Handle case where encoded value has extra leading byte
	if length == numBytes+1 {
		start = 1
		length = numBytes
	} else if length > numBytes {
		length = numBytes
	}

	// Copy bytes to the end of the array (big-endian)
	copy(bytes[numBytes-length:], biBytes[start:start+length])

	return bytes
}

// decodeBigInt decodes bytes to a big.Int (unsigned interpretation)
func decodeBigInt(bytes []byte) *big.Int {
	if len(bytes) == 0 {
		return big.NewInt(0)
	}
	result := big.NewInt(0)
	for _, b := range bytes {
		result.Mul(result, big.NewInt(256))
		result.Add(result, big.NewInt(int64(b)))
	}
	return result
}
