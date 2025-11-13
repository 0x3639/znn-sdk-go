package utils

import (
	"bytes"
	"math/big"
	"testing"
)

// =============================================================================
// Arraycopy Tests
// =============================================================================

func TestArraycopy(t *testing.T) {
	src := []byte{1, 2, 3, 4, 5}
	dest := make([]byte, 10)

	Arraycopy(src, 0, dest, 0, 5)

	for i := 0; i < 5; i++ {
		if dest[i] != src[i] {
			t.Errorf("dest[%d] = %d, want %d", i, dest[i], src[i])
		}
	}
}

func TestArraycopy_WithOffset(t *testing.T) {
	src := []byte{1, 2, 3, 4, 5}
	dest := make([]byte, 10)

	Arraycopy(src, 2, dest, 3, 3)

	expected := []byte{0, 0, 0, 3, 4, 5, 0, 0, 0, 0}
	if !bytes.Equal(dest, expected) {
		t.Errorf("dest = %v, want %v", dest, expected)
	}
}

// =============================================================================
// BigInt Encoding/Decoding Tests
// =============================================================================

func TestDecodeBigInt(t *testing.T) {
	testCases := []struct {
		bytes    []byte
		expected int64
	}{
		{[]byte{0x00}, 0},
		{[]byte{0x01}, 1},
		{[]byte{0xFF}, 255},
		{[]byte{0x01, 0x00}, 256},
		{[]byte{0x01, 0x00, 0x00}, 65536},
	}

	for _, tc := range testCases {
		result := DecodeBigInt(tc.bytes)
		if result.Int64() != tc.expected {
			t.Errorf("DecodeBigInt(%v) = %d, want %d", tc.bytes, result.Int64(), tc.expected)
		}
	}
}

func TestEncodeBigInt(t *testing.T) {
	testCases := []struct {
		value    int64
		expected []byte
	}{
		{0, []byte{0x00}},
		{1, []byte{0x01}},
		{255, []byte{0xFF}},
		{256, []byte{0x01, 0x00}},
		{65536, []byte{0x01, 0x00, 0x00}},
	}

	for _, tc := range testCases {
		result := EncodeBigInt(big.NewInt(tc.value))
		if !bytes.Equal(result, tc.expected) {
			t.Errorf("EncodeBigInt(%d) = %v, want %v", tc.value, result, tc.expected)
		}
	}
}

func TestBigIntRoundTrip(t *testing.T) {
	values := []*big.Int{
		big.NewInt(0),
		big.NewInt(1),
		big.NewInt(255),
		big.NewInt(256),
		big.NewInt(65536),
		big.NewInt(1000000),
		new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil),
	}

	for _, val := range values {
		encoded := EncodeBigInt(val)
		decoded := DecodeBigInt(encoded)
		if decoded.Cmp(val) != 0 {
			t.Errorf("Round trip failed for %s", val.String())
		}
	}
}

func TestBigIntToBytes(t *testing.T) {
	// Test with 32 bytes
	val := big.NewInt(42)
	result := BigIntToBytes(val, 32)

	if len(result) != 32 {
		t.Errorf("len(result) = %d, want 32", len(result))
	}

	// Last byte should be 42
	if result[31] != 42 {
		t.Errorf("result[31] = %d, want 42", result[31])
	}

	// All other bytes should be 0
	for i := 0; i < 31; i++ {
		if result[i] != 0 {
			t.Errorf("result[%d] = %d, want 0", i, result[i])
		}
	}
}

func TestBigIntToBytesSigned_Positive(t *testing.T) {
	val := big.NewInt(42)
	result := BigIntToBytesSigned(val, 4)

	if len(result) != 4 {
		t.Errorf("len(result) = %d, want 4", len(result))
	}

	// Should be padded with 0x00
	if result[0] != 0x00 {
		t.Errorf("result[0] = %x, want 0x00", result[0])
	}
}

func TestBigIntToBytesSigned_Negative(t *testing.T) {
	val := big.NewInt(-1)
	result := BigIntToBytesSigned(val, 4)

	if len(result) != 4 {
		t.Errorf("len(result) = %d, want 4", len(result))
	}

	// Should be padded with 0xFF for negative numbers
	for i := 0; i < 4; i++ {
		if result[i] != 0xFF {
			t.Errorf("result[%d] = %x, want 0xFF", i, result[i])
		}
	}
}

func TestBytesToBigInt(t *testing.T) {
	// Empty bytes
	result := BytesToBigInt([]byte{})
	if result.Int64() != 0 {
		t.Errorf("BytesToBigInt([]) = %d, want 0", result.Int64())
	}

	// Non-empty bytes
	result = BytesToBigInt([]byte{0x01, 0x00})
	if result.Int64() != 256 {
		t.Errorf("BytesToBigInt([0x01, 0x00]) = %d, want 256", result.Int64())
	}
}

// =============================================================================
// Merge Tests
// =============================================================================

func TestMerge(t *testing.T) {
	arrays := [][]byte{
		{1, 2, 3},
		{4, 5},
		{6, 7, 8, 9},
	}

	result := Merge(arrays)
	expected := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}

	if !bytes.Equal(result, expected) {
		t.Errorf("Merge() = %v, want %v", result, expected)
	}
}

func TestMerge_EmptyArrays(t *testing.T) {
	arrays := [][]byte{}
	result := Merge(arrays)

	if len(result) != 0 {
		t.Errorf("len(Merge([])) = %d, want 0", len(result))
	}
}

func TestMerge_WithNils(t *testing.T) {
	arrays := [][]byte{
		{1, 2},
		nil,
		{3, 4},
	}

	result := Merge(arrays)
	expected := []byte{1, 2, 3, 4}

	if !bytes.Equal(result, expected) {
		t.Errorf("Merge() = %v, want %v", result, expected)
	}
}

// =============================================================================
// Integer Conversion Tests
// =============================================================================

func TestIntToBytes(t *testing.T) {
	testCases := []struct {
		value    int32
		expected []byte
	}{
		{0, []byte{0x00, 0x00, 0x00, 0x00}},
		{1, []byte{0x00, 0x00, 0x00, 0x01}},
		{256, []byte{0x00, 0x00, 0x01, 0x00}},
		{65536, []byte{0x00, 0x01, 0x00, 0x00}},
	}

	for _, tc := range testCases {
		result := IntToBytes(tc.value)
		if !bytes.Equal(result, tc.expected) {
			t.Errorf("IntToBytes(%d) = %v, want %v", tc.value, result, tc.expected)
		}
	}
}

func TestLongToBytes(t *testing.T) {
	testCases := []struct {
		value    int64
		expected []byte
	}{
		{0, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
		{1, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{256, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00}},
	}

	for _, tc := range testCases {
		result := LongToBytes(tc.value)
		if !bytes.Equal(result, tc.expected) {
			t.Errorf("LongToBytes(%d) = %v, want %v", tc.value, result, tc.expected)
		}
	}
}

// =============================================================================
// Base64 Tests
// =============================================================================

func TestBase64RoundTrip(t *testing.T) {
	original := []byte("Hello, World!")
	encoded := BytesToBase64(original)
	decoded, err := Base64ToBytes(encoded)

	if err != nil {
		t.Fatalf("Base64ToBytes() error = %v", err)
	}

	if !bytes.Equal(decoded, original) {
		t.Errorf("Round trip failed: got %v, want %v", decoded, original)
	}
}

func TestBase64ToBytes_Empty(t *testing.T) {
	result, err := Base64ToBytes("")
	if err != nil {
		t.Errorf("Base64ToBytes(\"\") error = %v, want nil", err)
	}
	if result != nil {
		t.Errorf("Base64ToBytes(\"\") = %v, want nil", result)
	}
}

// =============================================================================
// Hex Tests
// =============================================================================

func TestBytesToHex(t *testing.T) {
	testCases := []struct {
		bytes    []byte
		expected string
	}{
		{[]byte{}, ""},
		{[]byte{0x00}, "00"},
		{[]byte{0xFF}, "ff"},
		{[]byte{0x01, 0x02, 0x03}, "010203"},
	}

	for _, tc := range testCases {
		result := BytesToHex(tc.bytes)
		if result != tc.expected {
			t.Errorf("BytesToHex(%v) = %s, want %s", tc.bytes, result, tc.expected)
		}
	}
}

// =============================================================================
// Padding Tests
// =============================================================================

func TestLeftPadBytes(t *testing.T) {
	input := []byte{1, 2, 3}
	result := LeftPadBytes(input, 6)

	expected := []byte{0, 0, 0, 1, 2, 3}
	if !bytes.Equal(result, expected) {
		t.Errorf("LeftPadBytes() = %v, want %v", result, expected)
	}
}

func TestLeftPadBytes_AlreadyCorrectSize(t *testing.T) {
	input := []byte{1, 2, 3}
	result := LeftPadBytes(input, 3)

	if !bytes.Equal(result, input) {
		t.Errorf("LeftPadBytes() = %v, want %v", result, input)
	}
}

func TestLeftPadBytes_LargerThanTarget(t *testing.T) {
	input := []byte{1, 2, 3, 4, 5}
	result := LeftPadBytes(input, 3)

	// Should return original when already larger
	if !bytes.Equal(result, input) {
		t.Errorf("LeftPadBytes() = %v, want %v", result, input)
	}
}
