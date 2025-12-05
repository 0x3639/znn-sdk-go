package utils

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"math/big"
)

// =============================================================================
// Array Operations
// =============================================================================

// Arraycopy copies a slice from src to dest
func Arraycopy(src []byte, startPos int, dest []byte, destPos int, length int) {
	copy(dest[destPos:destPos+length], src[startPos:startPos+length])
}

// =============================================================================
// BigInt Encoding/Decoding
// =============================================================================

// DecodeBigInt converts bytes to big.Int (big-endian)
func DecodeBigInt(bytes []byte) *big.Int {
	result := big.NewInt(0)
	for i := 0; i < len(bytes); i++ {
		result.Mul(result, big.NewInt(256))
		result.Add(result, big.NewInt(int64(bytes[i])))
	}
	return result
}

// EncodeBigInt converts big.Int to bytes (big-endian, variable length)
func EncodeBigInt(number *big.Int) []byte {
	if number.Sign() == 0 {
		return []byte{0}
	}

	// Calculate size
	size := (number.BitLen() + 7) / 8
	result := make([]byte, size)

	// Convert
	temp := new(big.Int).Set(number)
	byteMask := big.NewInt(0xff)
	for i := 0; i < size; i++ {
		b := new(big.Int).And(temp, byteMask)
		result[size-i-1] = byte(b.Int64())
		temp.Rsh(temp, 8)
	}

	return result
}

// BigIntToBytes converts big.Int to fixed-size byte array
func BigIntToBytes(b *big.Int, numBytes int) []byte {
	bytes := make([]byte, numBytes)
	biBytes := EncodeBigInt(b)

	start := 0
	if len(biBytes) == numBytes+1 {
		start = 1
	}

	length := len(biBytes)
	if length > numBytes {
		length = numBytes
	}

	Arraycopy(biBytes, start, bytes, numBytes-length, length)
	return bytes
}

// BigIntToBytesSigned converts signed big.Int to fixed-size byte array
func BigIntToBytesSigned(b *big.Int, numBytes int) []byte {
	fillByte := byte(0x00)
	if b.Sign() < 0 {
		fillByte = 0xFF
	}

	bytes := make([]byte, numBytes)
	for i := 0; i < numBytes; i++ {
		bytes[i] = fillByte
	}

	biBytes := EncodeBigInt(b)
	start := 0
	if len(biBytes) == numBytes+1 {
		start = 1
	}

	length := len(biBytes)
	if length > numBytes {
		length = numBytes
	}

	Arraycopy(biBytes, start, bytes, numBytes-length, length)
	return bytes
}

// BytesToBigInt converts bytes to big.Int
func BytesToBigInt(bb []byte) *big.Int {
	if len(bb) == 0 {
		return big.NewInt(0)
	}
	return DecodeBigInt(bb)
}

// =============================================================================
// Array Merging
// =============================================================================

// Merge concatenates multiple byte arrays
func Merge(arrays [][]byte) []byte {
	// Calculate total size
	count := 0
	for _, array := range arrays {
		if array != nil {
			count += len(array)
		}
	}

	if count == 0 {
		return []byte{}
	}

	// Merge arrays
	merged := make([]byte, count)
	start := 0
	for _, array := range arrays {
		if len(array) > 0 {
			copy(merged[start:], array)
			start += len(array)
		}
	}

	return merged
}

// =============================================================================
// Integer Conversions
// =============================================================================

// IntToBytes converts int32 to 4 bytes (big-endian)
func IntToBytes(integer int32) []byte {
	bytes := make([]byte, 4)
	// #nosec G115 -- Safe conversion: int32 range fits in uint32
	binary.BigEndian.PutUint32(bytes, uint32(integer))
	return bytes
}

// LongToBytes converts int64 to 8 bytes (big-endian)
func LongToBytes(longValue int64) []byte {
	bytes := make([]byte, 8)
	// #nosec G115 -- Safe conversion: int64 range fits in uint64
	binary.BigEndian.PutUint64(bytes, uint64(longValue))
	return bytes
}

// Uint64ToBytes converts uint64 to 8 bytes (big-endian)
//
// This is the safe alternative to LongToBytes for unsigned values,
// avoiding the need for uint64->int64 conversion which can overflow.
func Uint64ToBytes(value uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, value)
	return bytes
}

// =============================================================================
// Base64 Encoding/Decoding
// =============================================================================

// Base64ToBytes decodes base64 string to bytes
func Base64ToBytes(base64Str string) ([]byte, error) {
	if base64Str == "" {
		return nil, nil
	}
	return base64.StdEncoding.DecodeString(base64Str)
}

// BytesToBase64 encodes bytes to base64 string
func BytesToBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

// =============================================================================
// Hex Encoding
// =============================================================================

// BytesToHex converts bytes to hex string
func BytesToHex(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

// =============================================================================
// Padding
// =============================================================================

// LeftPadBytes pads bytes on the left with zeros to reach the specified size
func LeftPadBytes(bytes []byte, size int) []byte {
	if len(bytes) >= size {
		return bytes
	}

	result := make([]byte, size)
	copy(result[size-len(bytes):], bytes)
	return result
}
