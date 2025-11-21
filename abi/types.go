package abi

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/zenon-network/go-zenon/common/types"
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
		// #nosec G115 -- numBytes is capped at 32, so numBytes*8 cannot overflow
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

// =============================================================================
// UnsignedIntType - Unsigned Integer Type
// =============================================================================

// UnsignedIntType represents unsigned integer types (uint8 to uint256)
type UnsignedIntType struct {
	NumericType
	size int // Size in bits
}

// NewUnsignedIntType creates a new unsigned integer type
func NewUnsignedIntType(name string) (*UnsignedIntType, error) {
	uit := &UnsignedIntType{}
	uit.name = name

	// Parse size from name (e.g., "uint256" -> 256)
	if name == "uint" {
		uit.size = 256
	} else if strings.HasPrefix(name, "uint") {
		sizeStr := strings.TrimPrefix(name, "uint")
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid uint type name: %s", name)
		}

		// Validate size (must be 8-256 in increments of 8)
		if size < 8 || size > 256 || size%8 != 0 {
			return nil, fmt.Errorf("invalid uint size: %d (must be 8-256 in increments of 8)", size)
		}
		uit.size = size
	} else {
		return nil, fmt.Errorf("invalid uint type name: %s", name)
	}

	return uit, nil
}

// GetCanonicalName returns the canonical name (uint defaults to uint256)
func (uit *UnsignedIntType) GetCanonicalName() string {
	if uit.name == "uint" {
		return "uint256"
	}
	return uit.name
}

// Encode encodes an unsigned integer value
func (uit *UnsignedIntType) Encode(value interface{}) ([]byte, error) {
	bigInt, err := uit.EncodeInternal(value)
	if err != nil {
		return nil, err
	}
	return EncodeUintBig(bigInt)
}

// Decode decodes an unsigned integer value
func (uit *UnsignedIntType) Decode(encoded []byte, offset int) (interface{}, error) {
	return DecodeUint(encoded, offset)
}

// EncodeUint encodes an unsigned int to 32 bytes
func EncodeUint(i uint64) ([]byte, error) {
	return EncodeUintBig(new(big.Int).SetUint64(i))
}

// EncodeUintBig encodes a big.Int to 32 bytes (unsigned)
// Returns error if the value is negative
func EncodeUintBig(bigInt *big.Int) ([]byte, error) {
	if bigInt.Sign() < 0 {
		return nil, fmt.Errorf("cannot encode negative value as unsigned integer: %s", bigInt.String())
	}
	return bigIntToBytes(bigInt, Int32Size), nil
}

// DecodeUint decodes an unsigned integer from encoded bytes at offset
func DecodeUint(encoded []byte, offset int) (*big.Int, error) {
	if len(encoded) < offset+Int32Size {
		return nil, fmt.Errorf("insufficient bytes for decoding uint")
	}

	bytes := encoded[offset : offset+Int32Size]
	return decodeBigInt(bytes), nil
}

// bigIntToBytes converts a big.Int to a fixed-size byte array (unsigned)
func bigIntToBytes(b *big.Int, numBytes int) []byte {
	// Create byte array filled with zeros
	bytes := make([]byte, numBytes)

	// Get big.Int bytes
	biBytes := b.Bytes()

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

// =============================================================================
// BoolType - Boolean Type
// =============================================================================

// BoolType represents boolean values (encoded as uint256: 0 or 1)
type BoolType struct {
	IntType
}

// NewBoolType creates a new boolean type
func NewBoolType() (*BoolType, error) {
	// Create an int256 type as the base (bool is encoded as int256)
	intType, err := NewIntType("int256")
	if err != nil {
		return nil, err
	}

	// Override the name to "bool"
	intType.name = "bool"

	return &BoolType{
		IntType: *intType,
	}, nil
}

// Encode encodes a boolean value as 0 (false) or 1 (true)
func (bt *BoolType) Encode(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case bool:
		if v {
			return bt.IntType.Encode(1)
		}
		return bt.IntType.Encode(0)

	case string:
		if v == "true" || v == "True" || v == "TRUE" || v == "1" {
			return bt.IntType.Encode(1)
		}
		return bt.IntType.Encode(0)

	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		// Allow numeric input: 0 = false, anything else = true
		bigInt, err := bt.EncodeInternal(value)
		if err != nil {
			return nil, err
		}
		if bigInt.Sign() == 0 {
			return bt.IntType.Encode(0)
		}
		return bt.IntType.Encode(1)

	default:
		return nil, fmt.Errorf("unsupported value type for boolean encoding: %T", value)
	}
}

// Decode decodes a boolean value (0 = false, non-zero = true)
func (bt *BoolType) Decode(encoded []byte, offset int) (interface{}, error) {
	result, err := bt.IntType.Decode(encoded, offset)
	if err != nil {
		return nil, err
	}

	bigInt, ok := result.(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected decode result type: %T", result)
	}

	return bigInt.Sign() != 0, nil
}

// =============================================================================
// AddressType - Address Type
// =============================================================================

// AddressType represents Zenon address values (20 bytes, left-padded to 32)
type AddressType struct {
	baseType
}

// NewAddressType creates a new address type
func NewAddressType() (*AddressType, error) {
	return &AddressType{
		baseType: baseType{name: "address"},
	}, nil
}

// Encode encodes an address value to 32 bytes (20-byte address left-padded with 12 zero bytes)
func (at *AddressType) Encode(value interface{}) ([]byte, error) {
	var addr types.Address
	var err error

	switch v := value.(type) {
	case string:
		addr, err = types.ParseAddress(v)
		if err != nil {
			return nil, fmt.Errorf("invalid address string: %s, error: %w", v, err)
		}

	case types.Address:
		addr = v

	case *types.Address:
		if v == nil {
			return nil, fmt.Errorf("nil address pointer")
		}
		addr = *v

	default:
		return nil, fmt.Errorf("unsupported value type for address encoding: %T", value)
	}

	// Address is 20 bytes, needs to be left-padded to 32 bytes
	// Result: [12 zero bytes][20 address bytes]
	result := make([]byte, Int32Size)
	addrBytes := addr.Bytes()
	copy(result[12:], addrBytes)

	return result, nil
}

// Decode decodes an address value from encoded bytes at offset
func (at *AddressType) Decode(encoded []byte, offset int) (interface{}, error) {
	if len(encoded) < offset+Int32Size {
		return nil, fmt.Errorf("insufficient bytes for decoding address")
	}

	// Address bytes are at offset+12 (skip 12 padding bytes) and are 20 bytes long
	addrBytes := encoded[offset+12 : offset+Int32Size]

	addr, err := types.BytesToAddress(addrBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode address: %w", err)
	}

	return addr, nil
}

// =============================================================================
// HashType - Hash Type
// =============================================================================

// HashType represents hash values (32 bytes, no padding needed)
type HashType struct {
	baseType
}

// NewHashType creates a new hash type
func NewHashType() (*HashType, error) {
	return &HashType{
		baseType: baseType{name: "hash"},
	}, nil
}

// Encode encodes a hash value to 32 bytes
func (ht *HashType) Encode(value interface{}) ([]byte, error) {
	var hash types.Hash
	var err error

	switch v := value.(type) {
	case string:
		// Try parsing as hex string
		hash, err = types.HexToHash(v)
		if err != nil {
			return nil, fmt.Errorf("invalid hash string: %s, error: %w", v, err)
		}

	case []byte:
		// Convert byte slice to hash
		hash, err = types.BytesToHash(v)
		if err != nil {
			return nil, fmt.Errorf("invalid hash bytes (expected 32 bytes): %w", err)
		}

	case types.Hash:
		hash = v

	case *types.Hash:
		if v == nil {
			return nil, fmt.Errorf("nil hash pointer")
		}
		hash = *v

	default:
		return nil, fmt.Errorf("unsupported value type for hash encoding: %T", value)
	}

	// Hash is already 32 bytes, no padding needed
	return hash.Bytes(), nil
}

// Decode decodes a hash value from encoded bytes at offset
func (ht *HashType) Decode(encoded []byte, offset int) (interface{}, error) {
	if len(encoded) < offset+Int32Size {
		return nil, fmt.Errorf("insufficient bytes for decoding hash")
	}

	// Extract 32 bytes for the hash
	hashBytes := encoded[offset : offset+Int32Size]

	hash, err := types.BytesToHash(hashBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	return hash, nil
}

// =============================================================================
// Bytes32Type - Fixed 32-byte Type
// =============================================================================

// Bytes32Type represents fixed 32-byte values (similar to HashType but decodes to []byte)
type Bytes32Type struct {
	baseType
}

// NewBytes32Type creates a new bytes32 type
func NewBytes32Type(name string) (*Bytes32Type, error) {
	return &Bytes32Type{
		baseType: baseType{name: name},
	}, nil
}

// Encode encodes a bytes32 value to 32 bytes
func (bt *Bytes32Type) Encode(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case string:
		// Try parsing as hex string
		if len(v) > 2 && v[:2] == "0x" {
			v = v[2:] // Remove 0x prefix
		}

		// Decode hex string
		if len(v) != 64 {
			return nil, fmt.Errorf("invalid hex string length: expected 64 chars, got %d", len(v))
		}

		bytes := make([]byte, 32)
		for i := 0; i < 32; i++ {
			_, err := fmt.Sscanf(v[i*2:i*2+2], "%x", &bytes[i])
			if err != nil {
				return nil, fmt.Errorf("invalid hex string: %w", err)
			}
		}
		return bytes, nil

	case []byte:
		// Byte slice must be exactly 32 bytes or less (will be right-padded)
		if len(v) > 32 {
			return nil, fmt.Errorf("byte slice too long: expected max 32 bytes, got %d", len(v))
		}

		result := make([]byte, 32)
		copy(result, v)
		return result, nil

	case *big.Int:
		// Encode as big.Int (for numeric values)
		return EncodeIntBig(v), nil

	case int, int8, int16, int32, int64:
		// Convert to big.Int and encode
		var bigInt *big.Int
		switch val := v.(type) {
		case int:
			bigInt = big.NewInt(int64(val))
		case int8:
			bigInt = big.NewInt(int64(val))
		case int16:
			bigInt = big.NewInt(int64(val))
		case int32:
			bigInt = big.NewInt(int64(val))
		case int64:
			bigInt = big.NewInt(val)
		}
		return EncodeIntBig(bigInt), nil

	default:
		return nil, fmt.Errorf("unsupported value type for bytes32 encoding: %T", value)
	}
}

// Decode decodes a bytes32 value from encoded bytes at offset
func (bt *Bytes32Type) Decode(encoded []byte, offset int) (interface{}, error) {
	if len(encoded) < offset+Int32Size {
		return nil, fmt.Errorf("insufficient bytes for decoding bytes32")
	}

	// Extract 32 bytes
	result := make([]byte, Int32Size)
	copy(result, encoded[offset:offset+Int32Size])

	return result, nil
}

// =============================================================================
// TokenStandardType - Token Standard Type
// =============================================================================

// TokenStandardType represents Zenon token standard values (10 bytes, left-padded to 32)
type TokenStandardType struct {
	baseType
}

// NewTokenStandardType creates a new token standard type
func NewTokenStandardType() (*TokenStandardType, error) {
	return &TokenStandardType{
		baseType: baseType{name: "tokenStandard"},
	}, nil
}

// Encode encodes a token standard value to 32 bytes (10-byte ZTS left-padded with 22 zero bytes)
func (tst *TokenStandardType) Encode(value interface{}) ([]byte, error) {
	var zts types.ZenonTokenStandard
	var err error

	switch v := value.(type) {
	case string:
		zts, err = types.ParseZTS(v)
		if err != nil {
			return nil, fmt.Errorf("invalid token standard string: %s, error: %w", v, err)
		}

	case types.ZenonTokenStandard:
		zts = v

	case *types.ZenonTokenStandard:
		if v == nil {
			return nil, fmt.Errorf("nil token standard pointer")
		}
		zts = *v

	default:
		return nil, fmt.Errorf("unsupported value type for token standard encoding: %T", value)
	}

	// ZTS is 10 bytes, needs to be left-padded to 32 bytes
	// Result: [22 zero bytes][10 ZTS bytes]
	result := make([]byte, Int32Size)
	ztsBytes := zts.Bytes()
	copy(result[22:], ztsBytes)

	return result, nil
}

// Decode decodes a token standard value from encoded bytes at offset
func (tst *TokenStandardType) Decode(encoded []byte, offset int) (interface{}, error) {
	if len(encoded) < offset+Int32Size {
		return nil, fmt.Errorf("insufficient bytes for decoding token standard")
	}

	// ZTS bytes are at offset+22 (skip 22 padding bytes) and are 10 bytes long
	ztsBytes := encoded[offset+22 : offset+Int32Size]

	zts, err := types.BytesToZTS(ztsBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token standard: %w", err)
	}

	return zts, nil
}

// =============================================================================
// BytesType - Dynamic Bytes Type
// =============================================================================

// BytesType represents dynamic byte arrays (length-prefixed, padded to 32-byte multiples)
type BytesType struct {
	baseType
}

// NewBytesType creates a new bytes type
func NewBytesType() (*BytesType, error) {
	return &BytesType{
		baseType: baseType{name: "bytes"},
	}, nil
}

// Encode encodes dynamic bytes with length prefix and padding
// Format: [32 bytes: length][data padded to 32-byte multiple]
func (bt *BytesType) Encode(value interface{}) ([]byte, error) {
	var data []byte

	switch v := value.(type) {
	case []byte:
		data = v

	case string:
		// Decode hex string
		if len(v) > 2 && v[:2] == "0x" {
			v = v[2:] // Remove 0x prefix
		}

		// Decode hex string to bytes
		if len(v)%2 != 0 {
			return nil, fmt.Errorf("invalid hex string: odd length")
		}

		data = make([]byte, len(v)/2)
		for i := 0; i < len(data); i++ {
			_, err := fmt.Sscanf(v[i*2:i*2+2], "%x", &data[i])
			if err != nil {
				return nil, fmt.Errorf("invalid hex string: %w", err)
			}
		}

	default:
		return nil, fmt.Errorf("unsupported value type for bytes encoding: %T", value)
	}

	// Calculate padded length (round up to nearest 32-byte multiple)
	paddedLen := 0
	if len(data) > 0 {
		paddedLen = ((len(data)-1)/Int32Size + 1) * Int32Size
	}

	// Create result: [length][padded data]
	result := make([]byte, Int32Size+paddedLen)

	// Encode length
	lengthBytes := EncodeInt(len(data))
	copy(result[0:Int32Size], lengthBytes)

	// Copy data
	copy(result[Int32Size:], data)

	// Remaining bytes are already zero (padding)
	return result, nil
}

// Decode decodes dynamic bytes from encoded data at offset
func (bt *BytesType) Decode(encoded []byte, offset int) (interface{}, error) {
	if len(encoded) < offset+Int32Size {
		return nil, fmt.Errorf("insufficient bytes for decoding bytes length")
	}

	// Decode length from first 32 bytes
	lengthBig, err := DecodeInt(encoded, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to decode bytes length: %w", err)
	}

	length := int(lengthBig.Int64())
	if length < 0 {
		return nil, fmt.Errorf("invalid bytes length: %d", length)
	}

	if length == 0 {
		return []byte{}, nil
	}

	// Check if we have enough bytes for the data
	dataOffset := offset + Int32Size
	if len(encoded) < dataOffset+length {
		return nil, fmt.Errorf("insufficient bytes for decoding bytes data")
	}

	// Extract data
	result := make([]byte, length)
	copy(result, encoded[dataOffset:dataOffset+length])

	return result, nil
}

// IsDynamicType returns true for BytesType
func (bt *BytesType) IsDynamicType() bool {
	return true
}

// =============================================================================
// StringType - Dynamic String Type
// =============================================================================

// StringType represents UTF-8 encoded strings (extends BytesType)
type StringType struct {
	BytesType
}

// NewStringType creates a new string type
func NewStringType() (*StringType, error) {
	return &StringType{
		BytesType: BytesType{
			baseType: baseType{name: "string"},
		},
	}, nil
}

// GetFixedSize returns 0 as string is a dynamic type
func (st *StringType) GetFixedSize() int {
	return 0
}

// Encode encodes a string as UTF-8 bytes using BytesType encoding
func (st *StringType) Encode(value interface{}) ([]byte, error) {
	var str string

	switch v := value.(type) {
	case string:
		str = v
	default:
		return nil, fmt.Errorf("unsupported value type for string encoding: %T", value)
	}

	// Convert string to UTF-8 bytes and use BytesType encoding
	utf8Bytes := []byte(str)
	return st.BytesType.Encode(utf8Bytes)
}

// Decode decodes a string from encoded bytes at offset
func (st *StringType) Decode(encoded []byte, offset int) (interface{}, error) {
	// Use BytesType to decode the bytes
	decoded, err := st.BytesType.Decode(encoded, offset)
	if err != nil {
		return nil, err
	}

	bytes, ok := decoded.([]byte)
	if !ok {
		return nil, fmt.Errorf("unexpected decode result type: %T", decoded)
	}

	// Convert bytes to UTF-8 string
	return string(bytes), nil
}

// =============================================================================
// GetType - Type Factory Function
// =============================================================================

// GetType creates an ABI type from a type name string
func GetType(typeName string) (AbiType, error) {
	// Check for array types
	if strings.Contains(typeName, "[") {
		idx1 := strings.Index(typeName, "[")
		idx2 := strings.Index(typeName[idx1:], "]")
		if idx1 == -1 || idx2 == -1 {
			return nil, fmt.Errorf("invalid array type: %s", typeName)
		}
		idx2 += idx1

		// Check if it's a dynamic array (empty brackets)
		if idx1+1 == idx2 {
			return NewDynamicArrayType(typeName)
		}
		return NewStaticArrayType(typeName)
	}

	// Handle basic types
	switch {
	case strings.HasPrefix(typeName, "int"):
		return NewIntType(typeName)
	case strings.HasPrefix(typeName, "uint"):
		return NewUnsignedIntType(typeName)
	case typeName == "bool":
		return NewBoolType()
	case typeName == "address":
		return NewAddressType()
	case typeName == "hash":
		return NewHashType()
	case typeName == "bytes32":
		return NewBytes32Type(typeName)
	case typeName == "tokenStandard":
		return NewTokenStandardType()
	case typeName == "bytes":
		return NewBytesType()
	case typeName == "string":
		return NewStringType()
	case typeName == "function":
		return NewFunctionType()
	default:
		return nil, fmt.Errorf("unknown type: %s", typeName)
	}
}

// =============================================================================
// ArrayType - Array Types Interface
// =============================================================================

// ArrayType is an interface for array types (static and dynamic)
type ArrayType interface {
	AbiType
	GetElementType() AbiType
	EncodeTuple(values []interface{}) ([]byte, error)
	DecodeTuple(encoded []byte, origOffset int, length int) ([]interface{}, error)
	EncodeList(values []interface{}) ([]byte, error)
}

// =============================================================================
// StaticArrayType - Fixed-Size Array Type
// =============================================================================

// StaticArrayType represents fixed-size arrays like uint256[3]
type StaticArrayType struct {
	baseType
	elementType AbiType
	size        int
}

// NewStaticArrayType creates a new static array type
// typeName should be in format "elementType[size]" e.g. "uint256[3]"
func NewStaticArrayType(typeName string) (*StaticArrayType, error) {
	// Parse type name to extract element type and size
	idx1 := strings.Index(typeName, "[")
	if idx1 == -1 {
		return nil, fmt.Errorf("invalid static array type name: %s", typeName)
	}

	idx2 := strings.Index(typeName[idx1:], "]")
	if idx2 == -1 {
		return nil, fmt.Errorf("invalid static array type name: %s", typeName)
	}
	idx2 += idx1

	// Extract size
	sizeStr := typeName[idx1+1 : idx2]
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		return nil, fmt.Errorf("invalid array size: %s", sizeStr)
	}

	// Extract element type name
	elementTypeName := typeName[0:idx1]

	// Handle sub-dimensions (e.g., "uint256[3][2]")
	subDim := ""
	if idx2+1 < len(typeName) {
		subDim = typeName[idx2+1:]
	}
	fullElementTypeName := elementTypeName + subDim

	// Get element type
	elementType, err := GetType(fullElementTypeName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse element type: %w", err)
	}

	return &StaticArrayType{
		baseType:    baseType{name: typeName},
		elementType: elementType,
		size:        size,
	}, nil
}

// GetElementType returns the element type
func (sat *StaticArrayType) GetElementType() AbiType {
	return sat.elementType
}

// GetCanonicalName returns the canonical type name
func (sat *StaticArrayType) GetCanonicalName() string {
	return fmt.Sprintf("%s[%d]", sat.elementType.GetCanonicalName(), sat.size)
}

// GetFixedSize returns the total size (element size * count)
func (sat *StaticArrayType) GetFixedSize() int {
	return sat.elementType.GetFixedSize() * sat.size
}

// Encode encodes a static array
func (sat *StaticArrayType) Encode(value interface{}) ([]byte, error) {
	// Convert value to slice
	var values []interface{}
	switch v := value.(type) {
	case []interface{}:
		values = v
	case []string, []int, []int64, []uint64, []bool:
		// Use reflection to convert typed slices
		rv := reflect.ValueOf(v)
		values = make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			values[i] = rv.Index(i).Interface()
		}
	default:
		return nil, fmt.Errorf("unsupported value type for array encoding: %T", value)
	}

	return sat.EncodeList(values)
}

// EncodeList encodes a list of values as a static array
func (sat *StaticArrayType) EncodeList(values []interface{}) ([]byte, error) {
	if len(values) != sat.size {
		return nil, fmt.Errorf("array size mismatch: got %d elements, expected %d", len(values), sat.size)
	}
	return sat.EncodeTuple(values)
}

// EncodeTuple encodes array elements as a tuple
func (sat *StaticArrayType) EncodeTuple(values []interface{}) ([]byte, error) {
	var elems [][]byte

	// If element type is dynamic, use offset encoding
	if sat.elementType.IsDynamicType() {
		elems = make([][]byte, len(values)*2)
		offset := len(values) * Int32Size

		for i := 0; i < len(values); i++ {
			// Encode offset
			elems[i] = EncodeInt(offset)

			// Encode element
			encoded, err := sat.elementType.Encode(values[i])
			if err != nil {
				return nil, fmt.Errorf("failed to encode element %d: %w", i, err)
			}
			elems[len(values)+i] = encoded

			// Update offset (round up to 32-byte multiple)
			offset += (len(encoded) / Int32Size) * Int32Size
			if len(encoded)%Int32Size != 0 {
				offset += Int32Size
			}
		}
	} else {
		// For fixed-size elements, encode directly
		elems = make([][]byte, len(values))
		for i := 0; i < len(values); i++ {
			encoded, err := sat.elementType.Encode(values[i])
			if err != nil {
				return nil, fmt.Errorf("failed to encode element %d: %w", i, err)
			}
			elems[i] = encoded
		}
	}

	// Merge all encoded elements
	return bytes.Join(elems, nil), nil
}

// Decode decodes a static array from encoded data
func (sat *StaticArrayType) Decode(encoded []byte, offset int) (interface{}, error) {
	result := make([]interface{}, sat.size)

	for i := 0; i < sat.size; i++ {
		decoded, err := sat.elementType.Decode(encoded, offset+i*sat.elementType.GetFixedSize())
		if err != nil {
			return nil, fmt.Errorf("failed to decode element %d: %w", i, err)
		}
		result[i] = decoded
	}

	return result, nil
}

// DecodeTuple decodes array elements from a tuple encoding
func (sat *StaticArrayType) DecodeTuple(encoded []byte, origOffset int, length int) ([]interface{}, error) {
	offset := origOffset
	result := make([]interface{}, length)

	for i := 0; i < length; i++ {
		if sat.elementType.IsDynamicType() {
			// For dynamic types, read offset and decode from there
			offsetBig, err := DecodeInt(encoded, offset)
			if err != nil {
				return nil, fmt.Errorf("failed to decode offset for element %d: %w", i, err)
			}
			elemOffset := origOffset + int(offsetBig.Int64())

			decoded, err := sat.elementType.Decode(encoded, elemOffset)
			if err != nil {
				return nil, fmt.Errorf("failed to decode element %d: %w", i, err)
			}
			result[i] = decoded
		} else {
			// For fixed types, decode directly
			decoded, err := sat.elementType.Decode(encoded, offset)
			if err != nil {
				return nil, fmt.Errorf("failed to decode element %d: %w", i, err)
			}
			result[i] = decoded
		}
		offset += sat.elementType.GetFixedSize()
	}

	return result, nil
}

// =============================================================================
// DynamicArrayType - Variable-Size Array Type
// =============================================================================

// DynamicArrayType represents dynamic arrays like uint256[]
type DynamicArrayType struct {
	baseType
	elementType AbiType
}

// NewDynamicArrayType creates a new dynamic array type
// typeName should be in format "elementType[]" e.g. "uint256[]"
func NewDynamicArrayType(typeName string) (*DynamicArrayType, error) {
	// Parse type name to extract element type
	idx1 := strings.Index(typeName, "[")
	if idx1 == -1 {
		return nil, fmt.Errorf("invalid dynamic array type name: %s", typeName)
	}

	idx2 := strings.Index(typeName[idx1:], "]")
	if idx2 == -1 {
		return nil, fmt.Errorf("invalid dynamic array type name: %s", typeName)
	}
	idx2 += idx1

	// Ensure it's actually a dynamic array (nothing between [])
	if idx1+1 != idx2 {
		return nil, fmt.Errorf("not a dynamic array type: %s", typeName)
	}

	// Extract element type name
	elementTypeName := typeName[0:idx1]

	// Handle sub-dimensions (e.g., "uint256[][2]")
	subDim := ""
	if idx2+1 < len(typeName) {
		subDim = typeName[idx2+1:]
	}
	fullElementTypeName := elementTypeName + subDim

	// Get element type
	elementType, err := GetType(fullElementTypeName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse element type: %w", err)
	}

	return &DynamicArrayType{
		baseType:    baseType{name: typeName},
		elementType: elementType,
	}, nil
}

// GetElementType returns the element type
func (dat *DynamicArrayType) GetElementType() AbiType {
	return dat.elementType
}

// GetCanonicalName returns the canonical type name
func (dat *DynamicArrayType) GetCanonicalName() string {
	return dat.elementType.GetCanonicalName() + "[]"
}

// IsDynamicType returns true as dynamic arrays are dynamic types
func (dat *DynamicArrayType) IsDynamicType() bool {
	return true
}

// GetFixedSize returns 0 as dynamic arrays don't have a fixed size
func (dat *DynamicArrayType) GetFixedSize() int {
	return 0
}

// Encode encodes a dynamic array
func (dat *DynamicArrayType) Encode(value interface{}) ([]byte, error) {
	// Convert value to slice
	var values []interface{}
	switch v := value.(type) {
	case []interface{}:
		values = v
	case []string, []int, []int64, []uint64, []bool:
		// Use reflection to convert typed slices
		rv := reflect.ValueOf(v)
		values = make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			values[i] = rv.Index(i).Interface()
		}
	default:
		return nil, fmt.Errorf("unsupported value type for array encoding: %T", value)
	}

	return dat.EncodeList(values)
}

// EncodeList encodes a list of values as a dynamic array with length prefix
func (dat *DynamicArrayType) EncodeList(values []interface{}) ([]byte, error) {
	// Encode length
	lengthBytes := EncodeInt(len(values))

	// Encode tuple
	tupleBytes, err := dat.EncodeTuple(values)
	if err != nil {
		return nil, err
	}

	// Merge length and tuple
	return append(lengthBytes, tupleBytes...), nil
}

// EncodeTuple encodes array elements as a tuple
func (dat *DynamicArrayType) EncodeTuple(values []interface{}) ([]byte, error) {
	var elems [][]byte

	// If element type is dynamic, use offset encoding
	if dat.elementType.IsDynamicType() {
		elems = make([][]byte, len(values)*2)
		offset := len(values) * Int32Size

		for i := 0; i < len(values); i++ {
			// Encode offset
			elems[i] = EncodeInt(offset)

			// Encode element
			encoded, err := dat.elementType.Encode(values[i])
			if err != nil {
				return nil, fmt.Errorf("failed to encode element %d: %w", i, err)
			}
			elems[len(values)+i] = encoded

			// Update offset (round up to 32-byte multiple)
			offset += (len(encoded) / Int32Size) * Int32Size
			if len(encoded)%Int32Size != 0 {
				offset += Int32Size
			}
		}
	} else {
		// For fixed-size elements, encode directly
		elems = make([][]byte, len(values))
		for i := 0; i < len(values); i++ {
			encoded, err := dat.elementType.Encode(values[i])
			if err != nil {
				return nil, fmt.Errorf("failed to encode element %d: %w", i, err)
			}
			elems[i] = encoded
		}
	}

	// Merge all encoded elements
	return bytes.Join(elems, nil), nil
}

// Decode decodes a dynamic array from encoded data
func (dat *DynamicArrayType) Decode(encoded []byte, origOffset int) (interface{}, error) {
	// Decode length
	lengthBig, err := DecodeInt(encoded, origOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to decode array length: %w", err)
	}
	length := int(lengthBig.Int64())

	// Move past length
	origOffset += 32
	offset := origOffset
	result := make([]interface{}, length)

	for i := 0; i < length; i++ {
		if dat.elementType.IsDynamicType() {
			// For dynamic types, read offset and decode from there
			offsetBig, err := DecodeInt(encoded, offset)
			if err != nil {
				return nil, fmt.Errorf("failed to decode offset for element %d: %w", i, err)
			}
			elemOffset := origOffset + int(offsetBig.Int64())

			decoded, err := dat.elementType.Decode(encoded, elemOffset)
			if err != nil {
				return nil, fmt.Errorf("failed to decode element %d: %w", i, err)
			}
			result[i] = decoded
		} else {
			// For fixed types, decode directly
			decoded, err := dat.elementType.Decode(encoded, offset)
			if err != nil {
				return nil, fmt.Errorf("failed to decode element %d: %w", i, err)
			}
			result[i] = decoded
		}
		offset += dat.elementType.GetFixedSize()
	}

	return result, nil
}

// DecodeTuple decodes array elements from a tuple encoding
func (dat *DynamicArrayType) DecodeTuple(encoded []byte, origOffset int, length int) ([]interface{}, error) {
	offset := origOffset
	result := make([]interface{}, length)

	for i := 0; i < length; i++ {
		if dat.elementType.IsDynamicType() {
			// For dynamic types, read offset and decode from there
			offsetBig, err := DecodeInt(encoded, offset)
			if err != nil {
				return nil, fmt.Errorf("failed to decode offset for element %d: %w", i, err)
			}
			elemOffset := origOffset + int(offsetBig.Int64())

			decoded, err := dat.elementType.Decode(encoded, elemOffset)
			if err != nil {
				return nil, fmt.Errorf("failed to decode element %d: %w", i, err)
			}
			result[i] = decoded
		} else {
			// For fixed types, decode directly
			decoded, err := dat.elementType.Decode(encoded, offset)
			if err != nil {
				return nil, fmt.Errorf("failed to decode element %d: %w", i, err)
			}
			result[i] = decoded
		}
		offset += dat.elementType.GetFixedSize()
	}

	return result, nil
}

// =============================================================================
// FunctionType - Function Selector Type
// =============================================================================

// FunctionType represents a 24-byte function selector (extends Bytes32Type)
type FunctionType struct {
	Bytes32Type
}

// NewFunctionType creates a new function type
func NewFunctionType() (*FunctionType, error) {
	bytes32Type, err := NewBytes32Type("function")
	if err != nil {
		return nil, err
	}

	return &FunctionType{
		Bytes32Type: *bytes32Type,
	}, nil
}

// Encode encodes a 24-byte function selector
// The input must be exactly 24 bytes, which will be padded to 32 bytes
func (ft *FunctionType) Encode(value interface{}) ([]byte, error) {
	var selector []byte

	switch v := value.(type) {
	case []byte:
		selector = v
	case string:
		// Decode hex string
		if len(v) > 2 && v[:2] == "0x" {
			v = v[2:]
		}
		if len(v)%2 != 0 {
			return nil, fmt.Errorf("invalid hex string: odd length")
		}
		selector = make([]byte, len(v)/2)
		for i := 0; i < len(selector); i++ {
			_, err := fmt.Sscanf(v[i*2:i*2+2], "%x", &selector[i])
			if err != nil {
				return nil, fmt.Errorf("invalid hex string: %w", err)
			}
		}
	default:
		return nil, fmt.Errorf("unsupported value type for function encoding: %T", value)
	}

	// Function selector must be exactly 24 bytes
	if len(selector) != 24 {
		return nil, fmt.Errorf("function selector must be 24 bytes, got %d", len(selector))
	}

	// Pad with 8 zero bytes to make 32 bytes total
	padded := make([]byte, 32)
	copy(padded[0:24], selector)
	// bytes 24-31 remain zero

	// Use Bytes32Type to encode the padded value
	return ft.Bytes32Type.Encode(padded)
}

// Decode is unimplemented for FunctionType
func (ft *FunctionType) Decode(encoded []byte, offset int) (interface{}, error) {
	return nil, fmt.Errorf("FunctionType.Decode is not implemented")
}
