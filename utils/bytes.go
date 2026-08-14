package utils

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
)

// =============================================================================
// Array Operations
// =============================================================================

// Arraycopy copies length bytes from src[startPos:] to dest[destPos:].
//
// Arguments are bounds-checked: any negative index/length, a length that
// overflows a start position, or a range that extends past either slice is a
// safe no-op rather than a panic (CWE-129). Callers that must detect a bad
// request should validate the arguments themselves.
func Arraycopy(src []byte, startPos int, dest []byte, destPos int, length int) {
	if length <= 0 || startPos < 0 || destPos < 0 {
		return
	}
	// Guard the additions against int overflow before comparing to lengths.
	if startPos > len(src)-length || destPos > len(dest)-length {
		return
	}
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

// maxEncodedByteWidth bounds the requested output width of the big.Int byte
// encoders. It is far above any real use (transaction amounts use 32 bytes),
// yet small enough that numBytes*8 cannot overflow int on any platform.
const maxEncodedByteWidth = 1 << 20

// BigIntToBytes converts a non-negative big.Int to a fixed-size big-endian
// byte array of exactly numBytes bytes.
//
// The encoding is injective over its domain: the value must be non-negative
// and representable in numBytes bytes. Negative or over-wide values are
// rejected with an error rather than silently truncated or aliased (which
// would let distinct amounts collide in a signed transaction preimage, see
// CWE-682). numBytes must be positive and no larger than maxEncodedByteWidth.
func BigIntToBytes(b *big.Int, numBytes int) ([]byte, error) {
	if numBytes <= 0 || numBytes > maxEncodedByteWidth {
		return nil, fmt.Errorf("numBytes out of range: %d", numBytes)
	}
	if b == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}
	if b.Sign() < 0 {
		return nil, fmt.Errorf("value must be non-negative: %s", b)
	}
	// FillBytes panics if the value does not fit; reject that case up front.
	if b.BitLen() > numBytes*8 {
		return nil, fmt.Errorf("value %s does not fit in %d bytes", b, numBytes)
	}
	return b.FillBytes(make([]byte, numBytes)), nil
}

// BigIntToBytesSigned converts a signed big.Int to a fixed-size two's
// complement big-endian byte array of exactly numBytes bytes.
//
// The value must be representable in numBytes bytes of two's complement,
// i.e. within [-2^(8n-1), 2^(8n-1)-1]; out-of-range values are rejected
// rather than truncated. numBytes must be positive.
func BigIntToBytesSigned(b *big.Int, numBytes int) ([]byte, error) {
	if numBytes <= 0 || numBytes > maxEncodedByteWidth {
		return nil, fmt.Errorf("numBytes out of range: %d", numBytes)
	}
	if b == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}

	bits := uint(numBytes * 8)
	minimum := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), bits-1))
	maximum := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), bits-1), big.NewInt(1))
	if b.Cmp(minimum) < 0 || b.Cmp(maximum) > 0 {
		return nil, fmt.Errorf("value %s is out of range for a signed %d-byte integer", b, numBytes)
	}

	if b.Sign() >= 0 {
		return b.FillBytes(make([]byte, numBytes)), nil
	}

	// Two's complement: 2^(numBytes*8) + b.
	twosComp := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), bits), b)
	return twosComp.FillBytes(make([]byte, numBytes)), nil
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
