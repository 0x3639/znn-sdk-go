package utils

import (
	"encoding/hex"
	"testing"

	"github.com/zenon-network/go-zenon/common/types"
)

func TestNewHashHeight(t *testing.T) {
	hash := types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	height := uint64(12345)

	hh := NewHashHeight(hash, height)

	if hh.Hash != hash {
		t.Errorf("NewHashHeight hash = %s, want %s", hh.Hash.String(), hash.String())
	}
	if hh.Height != height {
		t.Errorf("NewHashHeight height = %d, want %d", hh.Height, height)
	}
}

func TestHashHeightGetBytes(t *testing.T) {
	tests := []struct {
		name           string
		hash           string
		height         uint64
		expectedLength int
	}{
		{
			name:           "zero hash zero height",
			hash:           "0000000000000000000000000000000000000000000000000000000000000000",
			height:         0,
			expectedLength: 40,
		},
		{
			name:           "non-zero hash with height",
			hash:           "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			height:         12345,
			expectedLength: 40,
		},
		{
			name:           "max height",
			hash:           "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			height:         ^uint64(0), // max uint64
			expectedLength: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := types.HexToHashPanic(tt.hash)
			hh := NewHashHeight(hash, tt.height)
			bytes := hh.GetBytes()

			if len(bytes) != tt.expectedLength {
				t.Errorf("GetBytes() length = %d, want %d", len(bytes), tt.expectedLength)
			}

			// Verify hash bytes (first 32 bytes)
			hashBytes := bytes[:32]
			if hex.EncodeToString(hashBytes) != tt.hash {
				t.Errorf("GetBytes() hash portion = %s, want %s", hex.EncodeToString(hashBytes), tt.hash)
			}

			// Verify height bytes (last 8 bytes, big-endian)
			heightBytes := bytes[32:40]
			expectedHeightBytes := LongToBytes(int64(tt.height))
			if hex.EncodeToString(heightBytes) != hex.EncodeToString(expectedHeightBytes) {
				t.Errorf("GetBytes() height portion = %s, want %s",
					hex.EncodeToString(heightBytes), hex.EncodeToString(expectedHeightBytes))
			}
		})
	}
}

func TestHashHeightGetBytesSpecificValue(t *testing.T) {
	// Test with known values to verify big-endian encoding
	hash := types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	height := uint64(12345) // 0x3039 in hex

	hh := NewHashHeight(hash, height)
	bytes := hh.GetBytes()

	// Height 12345 should be encoded as 0x0000000000003039 (big-endian)
	expectedHeightHex := "0000000000003039"
	actualHeightHex := hex.EncodeToString(bytes[32:40])

	if actualHeightHex != expectedHeightHex {
		t.Errorf("Height encoding = %s, want %s", actualHeightHex, expectedHeightHex)
	}
}

func TestEmptyHashHeight(t *testing.T) {
	if EmptyHashHeight.Hash != types.ZeroHash {
		t.Error("EmptyHashHeight.Hash should be ZeroHash")
	}
	if EmptyHashHeight.Height != 0 {
		t.Error("EmptyHashHeight.Height should be 0")
	}
}

func TestHashHeightIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		hh       HashHeight
		expected bool
	}{
		{
			name:     "empty hash height",
			hh:       EmptyHashHeight,
			expected: true,
		},
		{
			name:     "zero hash zero height",
			hh:       NewHashHeight(types.ZeroHash, 0),
			expected: true,
		},
		{
			name: "non-zero hash",
			hh: NewHashHeight(
				types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
				0,
			),
			expected: false,
		},
		{
			name:     "non-zero height",
			hh:       NewHashHeight(types.ZeroHash, 1),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.hh.IsEmpty() != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", tt.hh.IsEmpty(), tt.expected)
			}
		})
	}
}
