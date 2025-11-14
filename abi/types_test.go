package abi

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/zenon-network/go-zenon/common/types"
)

// =============================================================================
// IntType Tests
// =============================================================================

func TestNewIntType(t *testing.T) {
	tests := []struct {
		name      string
		typeName  string
		wantSize  int
		wantError bool
	}{
		{"int defaults to int256", "int", 256, false},
		{"int8", "int8", 8, false},
		{"int16", "int16", 16, false},
		{"int32", "int32", 32, false},
		{"int64", "int64", 64, false},
		{"int128", "int128", 128, false},
		{"int256", "int256", 256, false},
		{"invalid size", "int7", 0, true},
		{"invalid size too large", "int512", 0, true},
		{"invalid name", "integer", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewIntType(tt.typeName)
			if (err != nil) != tt.wantError {
				t.Errorf("NewIntType() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && got.size != tt.wantSize {
				t.Errorf("NewIntType() size = %v, want %v", got.size, tt.wantSize)
			}
		})
	}
}

func TestIntType_GetCanonicalName(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		want     string
	}{
		{"int becomes int256", "int", "int256"},
		{"int32 stays int32", "int32", "int32"},
		{"int256 stays int256", "int256", "int256"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it, err := NewIntType(tt.typeName)
			if err != nil {
				t.Fatalf("NewIntType() error = %v", err)
			}
			if got := it.GetCanonicalName(); got != tt.want {
				t.Errorf("IntType.GetCanonicalName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntType_Encode(t *testing.T) {
	it, err := NewIntType("int256")
	if err != nil {
		t.Fatalf("NewIntType() error = %v", err)
	}

	tests := []struct {
		name      string
		value     interface{}
		want      []byte
		wantError bool
	}{
		{
			name:  "zero",
			value: 0,
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name:  "positive small",
			value: 1,
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
		},
		{
			name:  "positive 255",
			value: 255,
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255,
			},
		},
		{
			name:  "negative -1",
			value: -1,
			want: []byte{
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
			},
		},
		{
			name:  "negative -128",
			value: -128,
			want: []byte{
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 128,
			},
		},
		{
			name:  "big.Int positive",
			value: big.NewInt(12345),
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x30, 0x39,
			},
		},
		{
			name:  "string decimal",
			value: "42",
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42,
			},
		},
		{
			name:  "string hex with 0x prefix",
			value: "0xFF",
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255,
			},
		},
		{
			name:  "string hex without prefix",
			value: "ff",
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255,
			},
		},
		{
			name:      "invalid string",
			value:     "not a number",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := it.Encode(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("IntType.Encode() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && !bytes.Equal(got, tt.want) {
				t.Errorf("IntType.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntType_Decode(t *testing.T) {
	it, err := NewIntType("int256")
	if err != nil {
		t.Fatalf("NewIntType() error = %v", err)
	}

	tests := []struct {
		name    string
		encoded []byte
		offset  int
		want    *big.Int
	}{
		{
			name: "zero",
			encoded: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
			offset: 0,
			want:   big.NewInt(0),
		},
		{
			name: "positive 1",
			encoded: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
			offset: 0,
			want:   big.NewInt(1),
		},
		{
			name: "positive 255",
			encoded: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255,
			},
			offset: 0,
			want:   big.NewInt(255),
		},
		{
			name: "negative -1",
			encoded: []byte{
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
			},
			offset: 0,
			want:   big.NewInt(-1),
		},
		{
			name: "negative -128",
			encoded: []byte{
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 128,
			},
			offset: 0,
			want:   big.NewInt(-128),
		},
		{
			name: "with offset",
			encoded: []byte{
				// Offset bytes (ignored)
				99, 99, 99, 99,
				// Actual value (12345)
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x30, 0x39,
			},
			offset: 4,
			want:   big.NewInt(12345),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := it.Decode(tt.encoded, tt.offset)
			if err != nil {
				t.Errorf("IntType.Decode() error = %v", err)
				return
			}
			gotBigInt, ok := got.(*big.Int)
			if !ok {
				t.Errorf("IntType.Decode() returned non-*big.Int: %T", got)
				return
			}
			if gotBigInt.Cmp(tt.want) != 0 {
				t.Errorf("IntType.Decode() = %v, want %v", gotBigInt, tt.want)
			}
		})
	}
}

func TestIntType_RoundTrip(t *testing.T) {
	it, err := NewIntType("int256")
	if err != nil {
		t.Fatalf("NewIntType() error = %v", err)
	}

	tests := []struct {
		name  string
		value *big.Int
	}{
		{"zero", big.NewInt(0)},
		{"one", big.NewInt(1)},
		{"minus one", big.NewInt(-1)},
		{"max int64", big.NewInt(9223372036854775807)},
		{"min int64", big.NewInt(-9223372036854775808)},
		{"large positive", new(big.Int).Exp(big.NewInt(2), big.NewInt(200), nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := it.Encode(tt.value)
			if err != nil {
				t.Errorf("IntType.Encode() error = %v", err)
				return
			}

			// Decode
			decoded, err := it.Decode(encoded, 0)
			if err != nil {
				t.Errorf("IntType.Decode() error = %v", err)
				return
			}

			decodedBigInt, ok := decoded.(*big.Int)
			if !ok {
				t.Errorf("IntType.Decode() returned non-*big.Int: %T", decoded)
				return
			}

			// Compare
			if decodedBigInt.Cmp(tt.value) != 0 {
				t.Errorf("Round trip failed: original = %v, decoded = %v", tt.value, decodedBigInt)
			}
		})
	}
}

func TestIntType_GetFixedSize(t *testing.T) {
	it, err := NewIntType("int256")
	if err != nil {
		t.Fatalf("NewIntType() error = %v", err)
	}

	if got := it.GetFixedSize(); got != Int32Size {
		t.Errorf("IntType.GetFixedSize() = %v, want %v", got, Int32Size)
	}
}

func TestIntType_IsDynamicType(t *testing.T) {
	it, err := NewIntType("int256")
	if err != nil {
		t.Fatalf("NewIntType() error = %v", err)
	}

	if got := it.IsDynamicType(); got != false {
		t.Errorf("IntType.IsDynamicType() = %v, want false", got)
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestEncodeInt(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  []byte
	}{
		{
			name:  "zero",
			value: 0,
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name:  "positive",
			value: 100,
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100,
			},
		},
		{
			name:  "negative",
			value: -1,
			want: []byte{
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
				255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeInt(tt.value)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("EncodeInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeInt(t *testing.T) {
	tests := []struct {
		name    string
		encoded []byte
		offset  int
		want    *big.Int
		wantErr bool
	}{
		{
			name: "valid positive",
			encoded: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42,
			},
			offset: 0,
			want:   big.NewInt(42),
		},
		{
			name: "insufficient bytes",
			encoded: []byte{
				0, 0, 0, 0,
			},
			offset:  0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeInt(tt.encoded, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Cmp(tt.want) != 0 {
				t.Errorf("DecodeInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// UnsignedIntType Tests
// =============================================================================

func TestNewUnsignedIntType(t *testing.T) {
	tests := []struct {
		name      string
		typeName  string
		wantSize  int
		wantError bool
	}{
		{"uint defaults to uint256", "uint", 256, false},
		{"uint8", "uint8", 8, false},
		{"uint16", "uint16", 16, false},
		{"uint32", "uint32", 32, false},
		{"uint64", "uint64", 64, false},
		{"uint128", "uint128", 128, false},
		{"uint256", "uint256", 256, false},
		{"invalid size", "uint7", 0, true},
		{"invalid size too large", "uint512", 0, true},
		{"invalid name", "unsigned", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUnsignedIntType(tt.typeName)
			if (err != nil) != tt.wantError {
				t.Errorf("NewUnsignedIntType() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && got.size != tt.wantSize {
				t.Errorf("NewUnsignedIntType() size = %v, want %v", got.size, tt.wantSize)
			}
		})
	}
}

func TestUnsignedIntType_GetCanonicalName(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		want     string
	}{
		{"uint becomes uint256", "uint", "uint256"},
		{"uint32 stays uint32", "uint32", "uint32"},
		{"uint256 stays uint256", "uint256", "uint256"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uit, err := NewUnsignedIntType(tt.typeName)
			if err != nil {
				t.Fatalf("NewUnsignedIntType() error = %v", err)
			}
			if got := uit.GetCanonicalName(); got != tt.want {
				t.Errorf("UnsignedIntType.GetCanonicalName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnsignedIntType_Encode(t *testing.T) {
	uit, err := NewUnsignedIntType("uint256")
	if err != nil {
		t.Fatalf("NewUnsignedIntType() error = %v", err)
	}

	tests := []struct {
		name      string
		value     interface{}
		want      []byte
		wantError bool
	}{
		{
			name:  "zero",
			value: 0,
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name:  "positive small",
			value: uint64(1),
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
		},
		{
			name:  "positive 255",
			value: uint64(255),
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255,
			},
		},
		{
			name:  "max uint64",
			value: uint64(18446744073709551615),
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 255, 255, 255, 255, 255, 255,
			},
		},
		{
			name:  "big.Int positive",
			value: big.NewInt(12345),
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x30, 0x39,
			},
		},
		{
			name:  "string decimal",
			value: "42",
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42,
			},
		},
		{
			name:  "string hex with 0x prefix",
			value: "0xFF",
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255,
			},
		},
		{
			name:      "negative value rejected",
			value:     -1,
			wantError: true,
		},
		{
			name:      "negative big.Int rejected",
			value:     big.NewInt(-100),
			wantError: true,
		},
		{
			name:      "invalid string",
			value:     "not a number",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := uit.Encode(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("UnsignedIntType.Encode() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && !bytes.Equal(got, tt.want) {
				t.Errorf("UnsignedIntType.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnsignedIntType_Decode(t *testing.T) {
	uit, err := NewUnsignedIntType("uint256")
	if err != nil {
		t.Fatalf("NewUnsignedIntType() error = %v", err)
	}

	tests := []struct {
		name    string
		encoded []byte
		offset  int
		want    *big.Int
	}{
		{
			name: "zero",
			encoded: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
			offset: 0,
			want:   big.NewInt(0),
		},
		{
			name: "positive 1",
			encoded: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
			offset: 0,
			want:   big.NewInt(1),
		},
		{
			name: "positive 255",
			encoded: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255,
			},
			offset: 0,
			want:   big.NewInt(255),
		},
		{
			name: "max uint64",
			encoded: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 255, 255, 255, 255, 255, 255,
			},
			offset: 0,
			want:   new(big.Int).SetUint64(18446744073709551615),
		},
		{
			name: "large value (2^255)",
			encoded: []byte{
				128, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
			offset: 0,
			want:   new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil),
		},
		{
			name: "with offset",
			encoded: []byte{
				// Offset bytes (ignored)
				99, 99, 99, 99,
				// Actual value (12345)
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x30, 0x39,
			},
			offset: 4,
			want:   big.NewInt(12345),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := uit.Decode(tt.encoded, tt.offset)
			if err != nil {
				t.Errorf("UnsignedIntType.Decode() error = %v", err)
				return
			}
			gotBigInt, ok := got.(*big.Int)
			if !ok {
				t.Errorf("UnsignedIntType.Decode() returned non-*big.Int: %T", got)
				return
			}
			if gotBigInt.Cmp(tt.want) != 0 {
				t.Errorf("UnsignedIntType.Decode() = %v, want %v", gotBigInt, tt.want)
			}
		})
	}
}

func TestUnsignedIntType_RoundTrip(t *testing.T) {
	uit, err := NewUnsignedIntType("uint256")
	if err != nil {
		t.Fatalf("NewUnsignedIntType() error = %v", err)
	}

	tests := []struct {
		name  string
		value *big.Int
	}{
		{"zero", big.NewInt(0)},
		{"one", big.NewInt(1)},
		{"max uint64", new(big.Int).SetUint64(18446744073709551615)},
		{"large positive", new(big.Int).Exp(big.NewInt(2), big.NewInt(200), nil)},
		{"2^255", new(big.Int).Exp(big.NewInt(2), big.NewInt(255), nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := uit.Encode(tt.value)
			if err != nil {
				t.Errorf("UnsignedIntType.Encode() error = %v", err)
				return
			}

			// Decode
			decoded, err := uit.Decode(encoded, 0)
			if err != nil {
				t.Errorf("UnsignedIntType.Decode() error = %v", err)
				return
			}

			decodedBigInt, ok := decoded.(*big.Int)
			if !ok {
				t.Errorf("UnsignedIntType.Decode() returned non-*big.Int: %T", decoded)
				return
			}

			// Compare
			if decodedBigInt.Cmp(tt.value) != 0 {
				t.Errorf("Round trip failed: original = %v, decoded = %v", tt.value, decodedBigInt)
			}
		})
	}
}

func TestUnsignedIntType_GetFixedSize(t *testing.T) {
	uit, err := NewUnsignedIntType("uint256")
	if err != nil {
		t.Fatalf("NewUnsignedIntType() error = %v", err)
	}

	if got := uit.GetFixedSize(); got != Int32Size {
		t.Errorf("UnsignedIntType.GetFixedSize() = %v, want %v", got, Int32Size)
	}
}

func TestUnsignedIntType_IsDynamicType(t *testing.T) {
	uit, err := NewUnsignedIntType("uint256")
	if err != nil {
		t.Fatalf("NewUnsignedIntType() error = %v", err)
	}

	if got := uit.IsDynamicType(); got != false {
		t.Errorf("UnsignedIntType.IsDynamicType() = %v, want false", got)
	}
}

func TestEncodeUint(t *testing.T) {
	tests := []struct {
		name  string
		value uint64
		want  []byte
	}{
		{
			name:  "zero",
			value: 0,
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name:  "positive",
			value: 100,
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100,
			},
		},
		{
			name:  "max uint64",
			value: 18446744073709551615,
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 255, 255, 255, 255, 255, 255,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncodeUint(tt.value)
			if err != nil {
				t.Errorf("EncodeUint() error = %v", err)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("EncodeUint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncodeUintBig_NegativeRejection(t *testing.T) {
	_, err := EncodeUintBig(big.NewInt(-1))
	if err == nil {
		t.Error("EncodeUintBig() should reject negative values")
	}
}

func TestDecodeUint(t *testing.T) {
	tests := []struct {
		name    string
		encoded []byte
		offset  int
		want    *big.Int
		wantErr bool
	}{
		{
			name: "valid positive",
			encoded: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42,
			},
			offset: 0,
			want:   big.NewInt(42),
		},
		{
			name: "insufficient bytes",
			encoded: []byte{
				0, 0, 0, 0,
			},
			offset:  0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeUint(tt.encoded, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeUint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Cmp(tt.want) != 0 {
				t.Errorf("DecodeUint() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// BoolType Tests
// =============================================================================

func TestNewBoolType(t *testing.T) {
	bt, err := NewBoolType()
	if err != nil {
		t.Fatalf("NewBoolType() error = %v", err)
	}
	if bt.GetName() != "bool" {
		t.Errorf("NewBoolType() name = %v, want 'bool'", bt.GetName())
	}
}

func TestBoolType_Encode(t *testing.T) {
	bt, err := NewBoolType()
	if err != nil {
		t.Fatalf("NewBoolType() error = %v", err)
	}

	tests := []struct {
		name      string
		value     interface{}
		want      []byte
		wantError bool
	}{
		{
			name:  "bool true",
			value: true,
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
		},
		{
			name:  "bool false",
			value: false,
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name:  "string 'true'",
			value: "true",
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
		},
		{
			name:  "string 'True'",
			value: "True",
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
		},
		{
			name:  "string 'TRUE'",
			value: "TRUE",
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
		},
		{
			name:  "string '1'",
			value: "1",
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
		},
		{
			name:  "string 'false'",
			value: "false",
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name:  "string 'False'",
			value: "False",
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name:  "string '0'",
			value: "0",
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name:  "string empty",
			value: "",
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name:  "int 0",
			value: 0,
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name:  "int 1",
			value: 1,
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
		},
		{
			name:  "int 42 (truthy)",
			value: 42,
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
		},
		{
			name:  "uint64 0",
			value: uint64(0),
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name:  "uint64 1",
			value: uint64(1),
			want: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
		},
		{
			name:      "unsupported type",
			value:     []byte{1, 2, 3},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bt.Encode(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("BoolType.Encode() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && !bytes.Equal(got, tt.want) {
				t.Errorf("BoolType.Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoolType_Decode(t *testing.T) {
	bt, err := NewBoolType()
	if err != nil {
		t.Fatalf("NewBoolType() error = %v", err)
	}

	tests := []struct {
		name    string
		encoded []byte
		offset  int
		want    bool
	}{
		{
			name: "zero = false",
			encoded: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			},
			offset: 0,
			want:   false,
		},
		{
			name: "one = true",
			encoded: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
			offset: 0,
			want:   true,
		},
		{
			name: "non-zero = true (42)",
			encoded: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42,
			},
			offset: 0,
			want:   true,
		},
		{
			name: "non-zero = true (255)",
			encoded: []byte{
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255,
			},
			offset: 0,
			want:   true,
		},
		{
			name: "with offset",
			encoded: []byte{
				// Offset bytes (ignored)
				99, 99, 99, 99,
				// Actual value (1 = true)
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
			},
			offset: 4,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := bt.Decode(tt.encoded, tt.offset)
			if err != nil {
				t.Errorf("BoolType.Decode() error = %v", err)
				return
			}
			gotBool, ok := got.(bool)
			if !ok {
				t.Errorf("BoolType.Decode() returned non-bool: %T", got)
				return
			}
			if gotBool != tt.want {
				t.Errorf("BoolType.Decode() = %v, want %v", gotBool, tt.want)
			}
		})
	}
}

func TestBoolType_RoundTrip(t *testing.T) {
	bt, err := NewBoolType()
	if err != nil {
		t.Fatalf("NewBoolType() error = %v", err)
	}

	tests := []struct {
		name  string
		value bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := bt.Encode(tt.value)
			if err != nil {
				t.Errorf("BoolType.Encode() error = %v", err)
				return
			}

			// Decode
			decoded, err := bt.Decode(encoded, 0)
			if err != nil {
				t.Errorf("BoolType.Decode() error = %v", err)
				return
			}

			decodedBool, ok := decoded.(bool)
			if !ok {
				t.Errorf("BoolType.Decode() returned non-bool: %T", decoded)
				return
			}

			// Compare
			if decodedBool != tt.value {
				t.Errorf("Round trip failed: original = %v, decoded = %v", tt.value, decodedBool)
			}
		})
	}
}

func TestBoolType_GetFixedSize(t *testing.T) {
	bt, err := NewBoolType()
	if err != nil {
		t.Fatalf("NewBoolType() error = %v", err)
	}

	if got := bt.GetFixedSize(); got != Int32Size {
		t.Errorf("BoolType.GetFixedSize() = %v, want %v", got, Int32Size)
	}
}

func TestBoolType_IsDynamicType(t *testing.T) {
	bt, err := NewBoolType()
	if err != nil {
		t.Fatalf("NewBoolType() error = %v", err)
	}

	if got := bt.IsDynamicType(); got != false {
		t.Errorf("BoolType.IsDynamicType() = %v, want false", got)
	}
}

// =============================================================================
// AddressType Tests
// =============================================================================

func TestNewAddressType(t *testing.T) {
	at, err := NewAddressType()
	if err != nil {
		t.Fatalf("NewAddressType() error = %v", err)
	}

	if at.GetName() != "address" {
		t.Errorf("AddressType.GetName() = %v, want 'address'", at.GetName())
	}
}

func TestAddressType_GetCanonicalName(t *testing.T) {
	at, err := NewAddressType()
	if err != nil {
		t.Fatalf("NewAddressType() error = %v", err)
	}

	if at.GetCanonicalName() != "address" {
		t.Errorf("AddressType.GetCanonicalName() = %v, want 'address'", at.GetCanonicalName())
	}
}

func TestAddressType_Encode(t *testing.T) {
	at, err := NewAddressType()
	if err != nil {
		t.Fatalf("NewAddressType() error = %v", err)
	}

	tests := []struct {
		name        string
		value       interface{}
		wantErr     bool
		wantHexPart string // Expected hex of the last 20 bytes (address part)
	}{
		{
			name:        "string address",
			value:       "z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7",
			wantErr:     false,
			wantHexPart: "0025374a419f32736f61ecc5ac4059d2f1b5884d",
		},
		{
			name:        "zero address string",
			value:       "z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f",
			wantErr:     false,
			wantHexPart: "0000000000000000000000000000000000000000",
		},
		{
			name:    "invalid string",
			value:   "not-an-address",
			wantErr: true,
		},
		{
			name:    "unsupported type",
			value:   123,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := at.Encode(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddressType.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(encoded) != Int32Size {
					t.Errorf("AddressType.Encode() returned %d bytes, want %d", len(encoded), Int32Size)
					return
				}

				// Check first 12 bytes are zero (padding)
				for i := 0; i < 12; i++ {
					if encoded[i] != 0 {
						t.Errorf("AddressType.Encode() byte %d = %x, want 0x00 (padding)", i, encoded[i])
					}
				}

				// Check last 20 bytes match expected address
				if tt.wantHexPart != "" {
					actualHex := fmt.Sprintf("%x", encoded[12:])
					if actualHex != tt.wantHexPart {
						t.Errorf("AddressType.Encode() address bytes = %s, want %s", actualHex, tt.wantHexPart)
					}
				}
			}
		})
	}
}

func TestAddressType_Encode_AddressValue(t *testing.T) {
	at, err := NewAddressType()
	if err != nil {
		t.Fatalf("NewAddressType() error = %v", err)
	}

	// Test with types.Address value
	addr := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")
	encoded, err := at.Encode(addr)
	if err != nil {
		t.Errorf("AddressType.Encode() with Address value error = %v", err)
		return
	}

	if len(encoded) != Int32Size {
		t.Errorf("AddressType.Encode() returned %d bytes, want %d", len(encoded), Int32Size)
	}

	expectedHex := "0025374a419f32736f61ecc5ac4059d2f1b5884d"
	actualHex := fmt.Sprintf("%x", encoded[12:])
	if actualHex != expectedHex {
		t.Errorf("AddressType.Encode() address bytes = %s, want %s", actualHex, expectedHex)
	}
}

func TestAddressType_Encode_AddressPointer(t *testing.T) {
	at, err := NewAddressType()
	if err != nil {
		t.Fatalf("NewAddressType() error = %v", err)
	}

	// Test with *types.Address value
	addr := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")
	encoded, err := at.Encode(&addr)
	if err != nil {
		t.Errorf("AddressType.Encode() with *Address value error = %v", err)
		return
	}

	if len(encoded) != Int32Size {
		t.Errorf("AddressType.Encode() returned %d bytes, want %d", len(encoded), Int32Size)
	}

	// Test with nil pointer
	var nilAddr *types.Address
	_, err = at.Encode(nilAddr)
	if err == nil {
		t.Error("AddressType.Encode() with nil pointer should return error")
	}
}

func TestAddressType_Decode(t *testing.T) {
	at, err := NewAddressType()
	if err != nil {
		t.Fatalf("NewAddressType() error = %v", err)
	}

	tests := []struct {
		name       string
		encodedHex string
		offset     int
		wantAddr   string
		wantErr    bool
	}{
		{
			name:       "valid address at offset 0",
			encodedHex: "0000000000000000000000000025374a419f32736f61ecc5ac4059d2f1b5884d",
			offset:     0,
			wantAddr:   "z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7",
			wantErr:    false,
		},
		{
			name:       "zero address",
			encodedHex: "0000000000000000000000000000000000000000000000000000000000000000",
			offset:     0,
			wantAddr:   "z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f",
			wantErr:    false,
		},
		{
			name:       "valid address at offset 32",
			encodedHex: "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000025374a419f32736f61ecc5ac4059d2f1b5884d",
			offset:     32,
			wantAddr:   "z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7",
			wantErr:    false,
		},
		{
			name:       "insufficient bytes",
			encodedHex: "0000000000000000000000000025374a419f32",
			offset:     0,
			wantErr:    true,
		},
		{
			name:       "offset too large",
			encodedHex: "0000000000000000000000000025374a419f32736f61ecc5ac4059d2f1b5884d",
			offset:     100,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert hex string to bytes
			encoded := make([]byte, len(tt.encodedHex)/2)
			for i := 0; i < len(encoded); i++ {
				fmt.Sscanf(tt.encodedHex[i*2:i*2+2], "%x", &encoded[i])
			}

			decoded, err := at.Decode(encoded, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddressType.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				addr, ok := decoded.(types.Address)
				if !ok {
					t.Errorf("AddressType.Decode() returned non-Address type: %T", decoded)
					return
				}

				if addr.String() != tt.wantAddr {
					t.Errorf("AddressType.Decode() = %v, want %v", addr.String(), tt.wantAddr)
				}
			}
		})
	}
}

func TestAddressType_RoundTrip(t *testing.T) {
	at, err := NewAddressType()
	if err != nil {
		t.Fatalf("NewAddressType() error = %v", err)
	}

	tests := []struct {
		name string
		addr string
	}{
		{"normal address", "z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7"},
		{"zero address", "z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := at.Encode(tt.addr)
			if err != nil {
				t.Errorf("AddressType.Encode() error = %v", err)
				return
			}

			// Decode
			decoded, err := at.Decode(encoded, 0)
			if err != nil {
				t.Errorf("AddressType.Decode() error = %v", err)
				return
			}

			addr, ok := decoded.(types.Address)
			if !ok {
				t.Errorf("AddressType.Decode() returned non-Address: %T", decoded)
				return
			}

			// Compare
			if addr.String() != tt.addr {
				t.Errorf("Round trip failed: original = %v, decoded = %v", tt.addr, addr.String())
			}
		})
	}
}

func TestAddressType_GetFixedSize(t *testing.T) {
	at, err := NewAddressType()
	if err != nil {
		t.Fatalf("NewAddressType() error = %v", err)
	}

	if got := at.GetFixedSize(); got != Int32Size {
		t.Errorf("AddressType.GetFixedSize() = %v, want %v", got, Int32Size)
	}
}

func TestAddressType_IsDynamicType(t *testing.T) {
	at, err := NewAddressType()
	if err != nil {
		t.Fatalf("NewAddressType() error = %v", err)
	}

	if got := at.IsDynamicType(); got != false {
		t.Errorf("AddressType.IsDynamicType() = %v, want false", got)
	}
}

// =============================================================================
// HashType Tests
// =============================================================================

func TestNewHashType(t *testing.T) {
	ht, err := NewHashType()
	if err != nil {
		t.Fatalf("NewHashType() error = %v", err)
	}

	if ht.GetName() != "hash" {
		t.Errorf("HashType.GetName() = %v, want 'hash'", ht.GetName())
	}
}

func TestHashType_GetCanonicalName(t *testing.T) {
	ht, err := NewHashType()
	if err != nil {
		t.Fatalf("NewHashType() error = %v", err)
	}

	if ht.GetCanonicalName() != "hash" {
		t.Errorf("HashType.GetCanonicalName() = %v, want 'hash'", ht.GetCanonicalName())
	}
}

func TestHashType_Encode(t *testing.T) {
	ht, err := NewHashType()
	if err != nil {
		t.Fatalf("NewHashType() error = %v", err)
	}

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
		wantHex string // Expected hex of encoded bytes
	}{
		{
			name:    "hex string",
			value:   "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
			wantErr: false,
			wantHex: "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
		},
		{
			name:    "zero hash string",
			value:   "0000000000000000000000000000000000000000000000000000000000000000",
			wantErr: false,
			wantHex: "0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:    "invalid hex string",
			value:   "not-a-hash",
			wantErr: true,
		},
		{
			name:    "invalid hex length",
			value:   "c51c6c118265d36db508a1a3d0c16b11",
			wantErr: true,
		},
		{
			name:    "unsupported type",
			value:   123,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := ht.Encode(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashType.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(encoded) != Int32Size {
					t.Errorf("HashType.Encode() returned %d bytes, want %d", len(encoded), Int32Size)
					return
				}

				if tt.wantHex != "" {
					actualHex := fmt.Sprintf("%x", encoded)
					if actualHex != tt.wantHex {
						t.Errorf("HashType.Encode() = %s, want %s", actualHex, tt.wantHex)
					}
				}
			}
		})
	}
}

func TestHashType_Encode_ByteSlice(t *testing.T) {
	ht, err := NewHashType()
	if err != nil {
		t.Fatalf("NewHashType() error = %v", err)
	}

	// Valid 32-byte slice
	validBytes := make([]byte, 32)
	for i := range validBytes {
		validBytes[i] = byte(i)
	}

	encoded, err := ht.Encode(validBytes)
	if err != nil {
		t.Errorf("HashType.Encode() with valid byte slice error = %v", err)
		return
	}

	if len(encoded) != Int32Size {
		t.Errorf("HashType.Encode() returned %d bytes, want %d", len(encoded), Int32Size)
	}

	if !bytes.Equal(encoded, validBytes) {
		t.Errorf("HashType.Encode() modified bytes")
	}

	// Invalid byte slice (wrong length)
	invalidBytes := make([]byte, 20)
	_, err = ht.Encode(invalidBytes)
	if err == nil {
		t.Error("HashType.Encode() with invalid byte slice should return error")
	}
}

func TestHashType_Encode_HashValue(t *testing.T) {
	ht, err := NewHashType()
	if err != nil {
		t.Fatalf("NewHashType() error = %v", err)
	}

	// Test with types.Hash value
	hash := types.HexToHashPanic("c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5")
	encoded, err := ht.Encode(hash)
	if err != nil {
		t.Errorf("HashType.Encode() with Hash value error = %v", err)
		return
	}

	if len(encoded) != Int32Size {
		t.Errorf("HashType.Encode() returned %d bytes, want %d", len(encoded), Int32Size)
	}

	expectedHex := "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5"
	actualHex := fmt.Sprintf("%x", encoded)
	if actualHex != expectedHex {
		t.Errorf("HashType.Encode() = %s, want %s", actualHex, expectedHex)
	}
}

func TestHashType_Encode_HashPointer(t *testing.T) {
	ht, err := NewHashType()
	if err != nil {
		t.Fatalf("NewHashType() error = %v", err)
	}

	// Test with *types.Hash value
	hash := types.HexToHashPanic("c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5")
	encoded, err := ht.Encode(&hash)
	if err != nil {
		t.Errorf("HashType.Encode() with *Hash value error = %v", err)
		return
	}

	if len(encoded) != Int32Size {
		t.Errorf("HashType.Encode() returned %d bytes, want %d", len(encoded), Int32Size)
	}

	// Test with nil pointer
	var nilHash *types.Hash
	_, err = ht.Encode(nilHash)
	if err == nil {
		t.Error("HashType.Encode() with nil pointer should return error")
	}
}

func TestHashType_Decode(t *testing.T) {
	ht, err := NewHashType()
	if err != nil {
		t.Fatalf("NewHashType() error = %v", err)
	}

	tests := []struct {
		name       string
		encodedHex string
		offset     int
		wantHash   string
		wantErr    bool
	}{
		{
			name:       "valid hash at offset 0",
			encodedHex: "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
			offset:     0,
			wantHash:   "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
			wantErr:    false,
		},
		{
			name:       "zero hash",
			encodedHex: "0000000000000000000000000000000000000000000000000000000000000000",
			offset:     0,
			wantHash:   "0000000000000000000000000000000000000000000000000000000000000000",
			wantErr:    false,
		},
		{
			name:       "valid hash at offset 32",
			encodedHex: "0000000000000000000000000000000000000000000000000000000000000000c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
			offset:     32,
			wantHash:   "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
			wantErr:    false,
		},
		{
			name:       "insufficient bytes",
			encodedHex: "c51c6c118265d36db508a1a3d0c16b11",
			offset:     0,
			wantErr:    true,
		},
		{
			name:       "offset too large",
			encodedHex: "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
			offset:     100,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert hex string to bytes
			encoded := make([]byte, len(tt.encodedHex)/2)
			for i := 0; i < len(encoded); i++ {
				fmt.Sscanf(tt.encodedHex[i*2:i*2+2], "%x", &encoded[i])
			}

			decoded, err := ht.Decode(encoded, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashType.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				hash, ok := decoded.(types.Hash)
				if !ok {
					t.Errorf("HashType.Decode() returned non-Hash type: %T", decoded)
					return
				}

				if hash.String() != tt.wantHash {
					t.Errorf("HashType.Decode() = %v, want %v", hash.String(), tt.wantHash)
				}
			}
		})
	}
}

func TestHashType_RoundTrip(t *testing.T) {
	ht, err := NewHashType()
	if err != nil {
		t.Fatalf("NewHashType() error = %v", err)
	}

	tests := []struct {
		name string
		hash string
	}{
		{"normal hash", "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5"},
		{"zero hash", "0000000000000000000000000000000000000000000000000000000000000000"},
		{"all ones", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := ht.Encode(tt.hash)
			if err != nil {
				t.Errorf("HashType.Encode() error = %v", err)
				return
			}

			// Decode
			decoded, err := ht.Decode(encoded, 0)
			if err != nil {
				t.Errorf("HashType.Decode() error = %v", err)
				return
			}

			hash, ok := decoded.(types.Hash)
			if !ok {
				t.Errorf("HashType.Decode() returned non-Hash: %T", decoded)
				return
			}

			// Compare
			if hash.String() != tt.hash {
				t.Errorf("Round trip failed: original = %v, decoded = %v", tt.hash, hash.String())
			}
		})
	}
}

func TestHashType_GetFixedSize(t *testing.T) {
	ht, err := NewHashType()
	if err != nil {
		t.Fatalf("NewHashType() error = %v", err)
	}

	if got := ht.GetFixedSize(); got != Int32Size {
		t.Errorf("HashType.GetFixedSize() = %v, want %v", got, Int32Size)
	}
}

func TestHashType_IsDynamicType(t *testing.T) {
	ht, err := NewHashType()
	if err != nil {
		t.Fatalf("NewHashType() error = %v", err)
	}

	if got := ht.IsDynamicType(); got != false {
		t.Errorf("HashType.IsDynamicType() = %v, want false", got)
	}
}

// =============================================================================
// Bytes32Type Tests
// =============================================================================

func TestNewBytes32Type(t *testing.T) {
	bt, err := NewBytes32Type("bytes32")
	if err != nil {
		t.Fatalf("NewBytes32Type() error = %v", err)
	}

	if bt.GetName() != "bytes32" {
		t.Errorf("Bytes32Type.GetName() = %v, want 'bytes32'", bt.GetName())
	}
}

func TestBytes32Type_GetCanonicalName(t *testing.T) {
	bt, err := NewBytes32Type("bytes32")
	if err != nil {
		t.Fatalf("NewBytes32Type() error = %v", err)
	}

	if bt.GetCanonicalName() != "bytes32" {
		t.Errorf("Bytes32Type.GetCanonicalName() = %v, want 'bytes32'", bt.GetCanonicalName())
	}
}

func TestBytes32Type_Encode_HexString(t *testing.T) {
	bt, err := NewBytes32Type("bytes32")
	if err != nil {
		t.Fatalf("NewBytes32Type() error = %v", err)
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
		wantHex string
	}{
		{
			name:    "hex string without 0x",
			value:   "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
			wantErr: false,
			wantHex: "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
		},
		{
			name:    "hex string with 0x",
			value:   "0xc51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
			wantErr: false,
			wantHex: "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
		},
		{
			name:    "zero bytes",
			value:   "0000000000000000000000000000000000000000000000000000000000000000",
			wantErr: false,
			wantHex: "0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:    "invalid hex string (too short)",
			value:   "c51c6c118265d36db508a1a3d0c16b11",
			wantErr: true,
		},
		{
			name:    "invalid hex string (too long)",
			value:   "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5ff",
			wantErr: true,
		},
		{
			name:    "invalid hex characters",
			value:   "not-a-valid-hex-string-with-64-characters-here-padding-it",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := bt.Encode(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bytes32Type.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(encoded) != Int32Size {
					t.Errorf("Bytes32Type.Encode() returned %d bytes, want %d", len(encoded), Int32Size)
					return
				}

				if tt.wantHex != "" {
					actualHex := fmt.Sprintf("%x", encoded)
					if actualHex != tt.wantHex {
						t.Errorf("Bytes32Type.Encode() = %s, want %s", actualHex, tt.wantHex)
					}
				}
			}
		})
	}
}

func TestBytes32Type_Encode_ByteSlice(t *testing.T) {
	bt, err := NewBytes32Type("bytes32")
	if err != nil {
		t.Fatalf("NewBytes32Type() error = %v", err)
	}

	// Valid 32-byte slice
	validBytes := make([]byte, 32)
	for i := range validBytes {
		validBytes[i] = byte(i)
	}

	encoded, err := bt.Encode(validBytes)
	if err != nil {
		t.Errorf("Bytes32Type.Encode() with valid byte slice error = %v", err)
		return
	}

	if len(encoded) != Int32Size {
		t.Errorf("Bytes32Type.Encode() returned %d bytes, want %d", len(encoded), Int32Size)
	}

	if !bytes.Equal(encoded, validBytes) {
		t.Errorf("Bytes32Type.Encode() modified bytes")
	}

	// Shorter byte slice (should be right-padded with zeros)
	shortBytes := []byte{1, 2, 3, 4, 5}
	encoded, err = bt.Encode(shortBytes)
	if err != nil {
		t.Errorf("Bytes32Type.Encode() with short byte slice error = %v", err)
		return
	}

	if len(encoded) != Int32Size {
		t.Errorf("Bytes32Type.Encode() returned %d bytes, want %d", len(encoded), Int32Size)
	}

	// Check first 5 bytes match
	if !bytes.Equal(encoded[:5], shortBytes) {
		t.Errorf("Bytes32Type.Encode() did not copy bytes correctly")
	}

	// Check remaining bytes are zero
	for i := 5; i < 32; i++ {
		if encoded[i] != 0 {
			t.Errorf("Bytes32Type.Encode() byte %d = %x, want 0x00 (padding)", i, encoded[i])
		}
	}

	// Invalid byte slice (too long)
	longBytes := make([]byte, 33)
	_, err = bt.Encode(longBytes)
	if err == nil {
		t.Error("Bytes32Type.Encode() with too long byte slice should return error")
	}
}

func TestBytes32Type_Encode_Numeric(t *testing.T) {
	bt, err := NewBytes32Type("bytes32")
	if err != nil {
		t.Fatalf("NewBytes32Type() error = %v", err)
	}

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"int", int(42), false},
		{"int8", int8(42), false},
		{"int16", int16(42), false},
		{"int32", int32(42), false},
		{"int64", int64(42), false},
		{"big.Int", big.NewInt(42), false},
		{"negative big.Int", big.NewInt(-42), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := bt.Encode(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bytes32Type.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(encoded) != Int32Size {
					t.Errorf("Bytes32Type.Encode() returned %d bytes, want %d", len(encoded), Int32Size)
				}
			}
		})
	}
}

func TestBytes32Type_Encode_UnsupportedType(t *testing.T) {
	bt, err := NewBytes32Type("bytes32")
	if err != nil {
		t.Fatalf("NewBytes32Type() error = %v", err)
	}

	_, err = bt.Encode(3.14) // float64
	if err == nil {
		t.Error("Bytes32Type.Encode() with unsupported type should return error")
	}
}

func TestBytes32Type_Decode(t *testing.T) {
	bt, err := NewBytes32Type("bytes32")
	if err != nil {
		t.Fatalf("NewBytes32Type() error = %v", err)
	}

	tests := []struct {
		name       string
		encodedHex string
		offset     int
		wantBytes  string
		wantErr    bool
	}{
		{
			name:       "valid bytes at offset 0",
			encodedHex: "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
			offset:     0,
			wantBytes:  "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
			wantErr:    false,
		},
		{
			name:       "zero bytes",
			encodedHex: "0000000000000000000000000000000000000000000000000000000000000000",
			offset:     0,
			wantBytes:  "0000000000000000000000000000000000000000000000000000000000000000",
			wantErr:    false,
		},
		{
			name:       "valid bytes at offset 32",
			encodedHex: "0000000000000000000000000000000000000000000000000000000000000000c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
			offset:     32,
			wantBytes:  "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
			wantErr:    false,
		},
		{
			name:       "insufficient bytes",
			encodedHex: "c51c6c118265d36db508a1a3d0c16b11",
			offset:     0,
			wantErr:    true,
		},
		{
			name:       "offset too large",
			encodedHex: "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5",
			offset:     100,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert hex string to bytes
			encoded := make([]byte, len(tt.encodedHex)/2)
			for i := 0; i < len(encoded); i++ {
				fmt.Sscanf(tt.encodedHex[i*2:i*2+2], "%x", &encoded[i])
			}

			decoded, err := bt.Decode(encoded, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bytes32Type.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				result, ok := decoded.([]byte)
				if !ok {
					t.Errorf("Bytes32Type.Decode() returned non-[]byte type: %T", decoded)
					return
				}

				if len(result) != 32 {
					t.Errorf("Bytes32Type.Decode() returned %d bytes, want 32", len(result))
					return
				}

				resultHex := fmt.Sprintf("%x", result)
				if resultHex != tt.wantBytes {
					t.Errorf("Bytes32Type.Decode() = %s, want %s", resultHex, tt.wantBytes)
				}
			}
		})
	}
}

func TestBytes32Type_RoundTrip(t *testing.T) {
	bt, err := NewBytes32Type("bytes32")
	if err != nil {
		t.Fatalf("NewBytes32Type() error = %v", err)
	}

	tests := []struct {
		name  string
		value string
	}{
		{"normal bytes", "c51c6c118265d36db508a1a3d0c16b11b3a3c5d8f6f0f1c5a5f5c5f5c5f5c5f5"},
		{"zero bytes", "0000000000000000000000000000000000000000000000000000000000000000"},
		{"all ones", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := bt.Encode(tt.value)
			if err != nil {
				t.Errorf("Bytes32Type.Encode() error = %v", err)
				return
			}

			// Decode
			decoded, err := bt.Decode(encoded, 0)
			if err != nil {
				t.Errorf("Bytes32Type.Decode() error = %v", err)
				return
			}

			result, ok := decoded.([]byte)
			if !ok {
				t.Errorf("Bytes32Type.Decode() returned non-[]byte: %T", decoded)
				return
			}

			// Compare
			resultHex := fmt.Sprintf("%x", result)
			if resultHex != tt.value {
				t.Errorf("Round trip failed: original = %v, decoded = %v", tt.value, resultHex)
			}
		})
	}
}

func TestBytes32Type_GetFixedSize(t *testing.T) {
	bt, err := NewBytes32Type("bytes32")
	if err != nil {
		t.Fatalf("NewBytes32Type() error = %v", err)
	}

	if got := bt.GetFixedSize(); got != Int32Size {
		t.Errorf("Bytes32Type.GetFixedSize() = %v, want %v", got, Int32Size)
	}
}

func TestBytes32Type_IsDynamicType(t *testing.T) {
	bt, err := NewBytes32Type("bytes32")
	if err != nil {
		t.Fatalf("NewBytes32Type() error = %v", err)
	}

	if got := bt.IsDynamicType(); got != false {
		t.Errorf("Bytes32Type.IsDynamicType() = %v, want false", got)
	}
}

// =============================================================================
// TokenStandardType Tests
// =============================================================================

func TestNewTokenStandardType(t *testing.T) {
	tst, err := NewTokenStandardType()
	if err != nil {
		t.Fatalf("NewTokenStandardType() error = %v", err)
	}

	if tst.GetName() != "tokenStandard" {
		t.Errorf("TokenStandardType.GetName() = %v, want 'tokenStandard'", tst.GetName())
	}
}

func TestTokenStandardType_GetCanonicalName(t *testing.T) {
	tst, err := NewTokenStandardType()
	if err != nil {
		t.Fatalf("NewTokenStandardType() error = %v", err)
	}

	if tst.GetCanonicalName() != "tokenStandard" {
		t.Errorf("TokenStandardType.GetCanonicalName() = %v, want 'tokenStandard'", tst.GetCanonicalName())
	}
}

func TestTokenStandardType_Encode(t *testing.T) {
	tst, err := NewTokenStandardType()
	if err != nil {
		t.Fatalf("NewTokenStandardType() error = %v", err)
	}

	tests := []struct {
		name        string
		value       interface{}
		wantErr     bool
		wantHexPart string // Expected hex of the last 10 bytes (ZTS part)
	}{
		{
			name:        "ZNN token standard string",
			value:       "zts1znnxxxxxxxxxxxxx9z4ulx",
			wantErr:     false,
			wantHexPart: "14e66318c6318c6318c6",
		},
		{
			name:        "QSR token standard string",
			value:       "zts1qsrxxxxxxxxxxxxxmrhjll",
			wantErr:     false,
			wantHexPart: "04066318c6318c6318c6",
		},
		{
			name:    "invalid string",
			value:   "not-a-token-standard",
			wantErr: true,
		},
		{
			name:    "unsupported type",
			value:   123,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := tst.Encode(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenStandardType.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(encoded) != Int32Size {
					t.Errorf("TokenStandardType.Encode() returned %d bytes, want %d", len(encoded), Int32Size)
					return
				}

				// Check first 22 bytes are zero (padding)
				for i := 0; i < 22; i++ {
					if encoded[i] != 0 {
						t.Errorf("TokenStandardType.Encode() byte %d = %x, want 0x00 (padding)", i, encoded[i])
					}
				}

				// Check last 10 bytes match expected ZTS
				if tt.wantHexPart != "" {
					actualHex := fmt.Sprintf("%x", encoded[22:])
					if actualHex != tt.wantHexPart {
						t.Errorf("TokenStandardType.Encode() ZTS bytes = %s, want %s", actualHex, tt.wantHexPart)
					}
				}
			}
		})
	}
}

func TestTokenStandardType_Encode_ZTSValue(t *testing.T) {
	tst, err := NewTokenStandardType()
	if err != nil {
		t.Fatalf("NewTokenStandardType() error = %v", err)
	}

	// Test with types.ZenonTokenStandard value
	zts := types.ZnnTokenStandard
	encoded, err := tst.Encode(zts)
	if err != nil {
		t.Errorf("TokenStandardType.Encode() with ZTS value error = %v", err)
		return
	}

	if len(encoded) != Int32Size {
		t.Errorf("TokenStandardType.Encode() returned %d bytes, want %d", len(encoded), Int32Size)
	}

	expectedHex := "14e66318c6318c6318c6"
	actualHex := fmt.Sprintf("%x", encoded[22:])
	if actualHex != expectedHex {
		t.Errorf("TokenStandardType.Encode() ZTS bytes = %s, want %s", actualHex, expectedHex)
	}
}

func TestTokenStandardType_Encode_ZTSPointer(t *testing.T) {
	tst, err := NewTokenStandardType()
	if err != nil {
		t.Fatalf("NewTokenStandardType() error = %v", err)
	}

	// Test with *types.ZenonTokenStandard value
	zts := types.QsrTokenStandard
	encoded, err := tst.Encode(&zts)
	if err != nil {
		t.Errorf("TokenStandardType.Encode() with *ZTS value error = %v", err)
		return
	}

	if len(encoded) != Int32Size {
		t.Errorf("TokenStandardType.Encode() returned %d bytes, want %d", len(encoded), Int32Size)
	}

	// Test with nil pointer
	var nilZTS *types.ZenonTokenStandard
	_, err = tst.Encode(nilZTS)
	if err == nil {
		t.Error("TokenStandardType.Encode() with nil pointer should return error")
	}
}

func TestTokenStandardType_Decode(t *testing.T) {
	tst, err := NewTokenStandardType()
	if err != nil {
		t.Fatalf("NewTokenStandardType() error = %v", err)
	}

	tests := []struct {
		name       string
		encodedHex string
		offset     int
		wantZTS    string
		wantErr    bool
	}{
		{
			name:       "ZNN token standard at offset 0",
			encodedHex: "0000000000000000000000000000000000000000000014e66318c6318c6318c6",
			offset:     0,
			wantZTS:    "zts1znnxxxxxxxxxxxxx9z4ulx",
			wantErr:    false,
		},
		{
			name:       "QSR token standard at offset 0",
			encodedHex: "0000000000000000000000000000000000000000000004066318c6318c6318c6",
			offset:     0,
			wantZTS:    "zts1qsrxxxxxxxxxxxxxmrhjll",
			wantErr:    false,
		},
		{
			name:       "ZNN token standard at offset 32",
			encodedHex: "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000014e66318c6318c6318c6",
			offset:     32,
			wantZTS:    "zts1znnxxxxxxxxxxxxx9z4ulx",
			wantErr:    false,
		},
		{
			name:       "insufficient bytes",
			encodedHex: "000000000000000000000000000000000000000000000014e6",
			offset:     0,
			wantErr:    true,
		},
		{
			name:       "offset too large",
			encodedHex: "000000000000000000000000000000000000000000000014e66318c6318c6318c6",
			offset:     100,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert hex string to bytes
			encoded := make([]byte, len(tt.encodedHex)/2)
			for i := 0; i < len(encoded); i++ {
				fmt.Sscanf(tt.encodedHex[i*2:i*2+2], "%x", &encoded[i])
			}

			decoded, err := tst.Decode(encoded, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenStandardType.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				zts, ok := decoded.(types.ZenonTokenStandard)
				if !ok {
					t.Errorf("TokenStandardType.Decode() returned non-ZenonTokenStandard type: %T", decoded)
					return
				}

				if zts.String() != tt.wantZTS {
					t.Errorf("TokenStandardType.Decode() = %v, want %v", zts.String(), tt.wantZTS)
				}
			}
		})
	}
}

func TestTokenStandardType_RoundTrip(t *testing.T) {
	tst, err := NewTokenStandardType()
	if err != nil {
		t.Fatalf("NewTokenStandardType() error = %v", err)
	}

	tests := []struct {
		name string
		zts  string
	}{
		{"ZNN token standard", "zts1znnxxxxxxxxxxxxx9z4ulx"},
		{"QSR token standard", "zts1qsrxxxxxxxxxxxxxmrhjll"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := tst.Encode(tt.zts)
			if err != nil {
				t.Errorf("TokenStandardType.Encode() error = %v", err)
				return
			}

			// Decode
			decoded, err := tst.Decode(encoded, 0)
			if err != nil {
				t.Errorf("TokenStandardType.Decode() error = %v", err)
				return
			}

			zts, ok := decoded.(types.ZenonTokenStandard)
			if !ok {
				t.Errorf("TokenStandardType.Decode() returned non-ZenonTokenStandard: %T", decoded)
				return
			}

			// Compare
			if zts.String() != tt.zts {
				t.Errorf("Round trip failed: original = %v, decoded = %v", tt.zts, zts.String())
			}
		})
	}
}

func TestTokenStandardType_GetFixedSize(t *testing.T) {
	tst, err := NewTokenStandardType()
	if err != nil {
		t.Fatalf("NewTokenStandardType() error = %v", err)
	}

	if got := tst.GetFixedSize(); got != Int32Size {
		t.Errorf("TokenStandardType.GetFixedSize() = %v, want %v", got, Int32Size)
	}
}

func TestTokenStandardType_IsDynamicType(t *testing.T) {
	tst, err := NewTokenStandardType()
	if err != nil {
		t.Fatalf("NewTokenStandardType() error = %v", err)
	}

	if got := tst.IsDynamicType(); got != false {
		t.Errorf("TokenStandardType.IsDynamicType() = %v, want false", got)
	}
}

// =============================================================================
// BytesType Tests
// =============================================================================

func TestNewBytesType(t *testing.T) {
	bt, err := NewBytesType()
	if err != nil {
		t.Fatalf("NewBytesType() error = %v", err)
	}

	if bt.GetName() != "bytes" {
		t.Errorf("BytesType.GetName() = %v, want 'bytes'", bt.GetName())
	}
}

func TestBytesType_IsDynamicType(t *testing.T) {
	bt, err := NewBytesType()
	if err != nil {
		t.Fatalf("NewBytesType() error = %v", err)
	}

	if got := bt.IsDynamicType(); got != true {
		t.Errorf("BytesType.IsDynamicType() = %v, want true", got)
	}
}

func TestBytesType_Encode_EmptyBytes(t *testing.T) {
	bt, err := NewBytesType()
	if err != nil {
		t.Fatalf("NewBytesType() error = %v", err)
	}

	encoded, err := bt.Encode([]byte{})
	if err != nil {
		t.Errorf("BytesType.Encode() with empty bytes error = %v", err)
		return
	}

	// Empty bytes: length=0, no data
	if len(encoded) != 32 {
		t.Errorf("BytesType.Encode() empty bytes length = %d, want 32", len(encoded))
	}

	// Check length is 0
	lengthBig, _ := DecodeInt(encoded, 0)
	if lengthBig.Int64() != 0 {
		t.Errorf("BytesType.Encode() empty bytes length field = %d, want 0", lengthBig.Int64())
	}
}

func TestBytesType_Encode_SmallBytes(t *testing.T) {
	bt, err := NewBytesType()
	if err != nil {
		t.Fatalf("NewBytesType() error = %v", err)
	}

	data := []byte{1, 2, 3, 4, 5}
	encoded, err := bt.Encode(data)
	if err != nil {
		t.Errorf("BytesType.Encode() error = %v", err)
		return
	}

	// Length (32) + data padded to 32 = 64 bytes total
	if len(encoded) != 64 {
		t.Errorf("BytesType.Encode() length = %d, want 64", len(encoded))
	}

	// Check length field
	lengthBig, _ := DecodeInt(encoded, 0)
	if lengthBig.Int64() != 5 {
		t.Errorf("BytesType.Encode() length field = %d, want 5", lengthBig.Int64())
	}

	// Check data
	if !bytes.Equal(encoded[32:37], data) {
		t.Errorf("BytesType.Encode() data mismatch")
	}

	// Check padding (bytes 37-64 should be zero)
	for i := 37; i < 64; i++ {
		if encoded[i] != 0 {
			t.Errorf("BytesType.Encode() padding byte %d = %x, want 0x00", i, encoded[i])
		}
	}
}

func TestBytesType_Encode_ExactlyOneBlock(t *testing.T) {
	bt, err := NewBytesType()
	if err != nil {
		t.Fatalf("NewBytesType() error = %v", err)
	}

	// 32 bytes exactly
	data := make([]byte, 32)
	for i := range data {
		data[i] = byte(i)
	}

	encoded, err := bt.Encode(data)
	if err != nil {
		t.Errorf("BytesType.Encode() error = %v", err)
		return
	}

	// Length (32) + data (32) = 64 bytes
	if len(encoded) != 64 {
		t.Errorf("BytesType.Encode() length = %d, want 64", len(encoded))
	}
}

func TestBytesType_Encode_LargeBytes(t *testing.T) {
	bt, err := NewBytesType()
	if err != nil {
		t.Fatalf("NewBytesType() error = %v", err)
	}

	// 50 bytes (needs 2 blocks = 64 bytes for data)
	data := make([]byte, 50)
	for i := range data {
		data[i] = byte(i)
	}

	encoded, err := bt.Encode(data)
	if err != nil {
		t.Errorf("BytesType.Encode() error = %v", err)
		return
	}

	// Length (32) + data padded to 64 = 96 bytes total
	if len(encoded) != 96 {
		t.Errorf("BytesType.Encode() length = %d, want 96", len(encoded))
	}

	// Check padding
	for i := 32 + 50; i < 96; i++ {
		if encoded[i] != 0 {
			t.Errorf("BytesType.Encode() padding byte %d = %x, want 0x00", i, encoded[i])
		}
	}
}

func TestBytesType_Decode(t *testing.T) {
	bt, err := NewBytesType()
	if err != nil {
		t.Fatalf("NewBytesType() error = %v", err)
	}

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"small", []byte{1, 2, 3, 4, 5}},
		{"exact block", make([]byte, 32)},
		{"large", make([]byte, 50)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := bt.Encode(tt.data)
			if err != nil {
				t.Errorf("BytesType.Encode() error = %v", err)
				return
			}

			// Decode
			decoded, err := bt.Decode(encoded, 0)
			if err != nil {
				t.Errorf("BytesType.Decode() error = %v", err)
				return
			}

			result, ok := decoded.([]byte)
			if !ok {
				t.Errorf("BytesType.Decode() returned non-[]byte: %T", decoded)
				return
			}

			if !bytes.Equal(result, tt.data) {
				t.Errorf("Round trip failed: length mismatch got %d want %d", len(result), len(tt.data))
			}
		})
	}
}

// ==================== StringType Tests ====================

func TestNewStringType(t *testing.T) {
	st, err := NewStringType()
	if err != nil {
		t.Fatalf("NewStringType() error = %v", err)
	}
	if st == nil {
		t.Fatal("NewStringType() returned nil")
	}
}

func TestStringType_GetCanonicalName(t *testing.T) {
	st, _ := NewStringType()
	if got := st.GetCanonicalName(); got != "string" {
		t.Errorf("StringType.GetCanonicalName() = %v, want string", got)
	}
}

func TestStringType_IsDynamicType(t *testing.T) {
	st, _ := NewStringType()
	if !st.IsDynamicType() {
		t.Error("StringType.IsDynamicType() = false, want true")
	}
}

func TestStringType_GetFixedSize(t *testing.T) {
	st, _ := NewStringType()
	if size := st.GetFixedSize(); size != 0 {
		t.Errorf("StringType.GetFixedSize() = %d, want 0 (dynamic type)", size)
	}
}

func TestStringType_Encode_ASCII(t *testing.T) {
	st, _ := NewStringType()

	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantData string // hex of the data portion (without length prefix)
	}{
		{
			name:     "simple ASCII",
			input:    "hello",
			wantLen:  5,
			wantData: "68656c6c6f" + strings.Repeat("0", 54), // "hello" + padding to 32 bytes
		},
		{
			name:     "longer ASCII",
			input:    "The quick brown fox jumps",
			wantLen:  25,
			wantData: "54686520717569636b2062726f776e20666f78206a756d7073" + strings.Repeat("0", 14), // 25 bytes + 7 padding
		},
		{
			name:     "exactly 32 bytes",
			input:    "12345678901234567890123456789012",
			wantLen:  32,
			wantData: "3132333435363738393031323334353637383930313233343536373839303132",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := st.Encode(tt.input)
			if err != nil {
				t.Errorf("StringType.Encode() error = %v", err)
				return
			}

			// Check length prefix (first 32 bytes)
			lengthBytes := encoded[:32]
			expectedLength := make([]byte, 32)
			binary.BigEndian.PutUint64(expectedLength[24:], uint64(tt.wantLen))

			if !bytes.Equal(lengthBytes, expectedLength) {
				t.Errorf("Length prefix mismatch:\ngot  %x\nwant %x", lengthBytes, expectedLength)
			}

			// Check data portion
			dataHex := hex.EncodeToString(encoded[32:])
			if dataHex != tt.wantData {
				t.Errorf("Data mismatch:\ngot  %s\nwant %s", dataHex, tt.wantData)
			}
		})
	}
}

func TestStringType_Encode_UTF8(t *testing.T) {
	st, _ := NewStringType()

	tests := []struct {
		name    string
		input   string
		wantLen int // byte length of UTF-8 encoding
	}{
		{
			name:    "Japanese characters",
			input:   "こんにちは",
			wantLen: 15, // 5 chars * 3 bytes each
		},
		{
			name:    "emoji",
			input:   "Hello 👋🌍",
			wantLen: 14, // "Hello " (6) + 👋 (4) + 🌍 (4)
		},
		{
			name:    "mixed ASCII and UTF-8",
			input:   "Café",
			wantLen: 5, // C(1) + a(1) + f(1) + é(2)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := st.Encode(tt.input)
			if err != nil {
				t.Errorf("StringType.Encode() error = %v", err)
				return
			}

			// Check length prefix
			lengthBytes := encoded[:32]
			expectedLength := make([]byte, 32)
			binary.BigEndian.PutUint64(expectedLength[24:], uint64(tt.wantLen))

			if !bytes.Equal(lengthBytes, expectedLength) {
				t.Errorf("Length prefix = %d, want %d", binary.BigEndian.Uint64(lengthBytes[24:]), tt.wantLen)
			}

			// Verify UTF-8 bytes in data
			expectedUTF8 := []byte(tt.input)
			dataBytes := encoded[32 : 32+tt.wantLen]
			if !bytes.Equal(dataBytes, expectedUTF8) {
				t.Errorf("UTF-8 data mismatch:\ngot  %x\nwant %x", dataBytes, expectedUTF8)
			}
		})
	}
}

func TestStringType_Encode_EmptyString(t *testing.T) {
	st, _ := NewStringType()

	encoded, err := st.Encode("")
	if err != nil {
		t.Fatalf("StringType.Encode(\"\") error = %v", err)
	}

	// Should be 32 bytes (length only, no data padding for empty)
	if len(encoded) != 32 {
		t.Errorf("Encoded empty string length = %d, want 32", len(encoded))
	}

	// Length should be 0
	lengthBytes := encoded[:32]
	expectedLength := make([]byte, 32)
	if !bytes.Equal(lengthBytes, expectedLength) {
		t.Errorf("Length prefix should be all zeros for empty string")
	}
}

func TestStringType_Encode_InvalidType(t *testing.T) {
	st, _ := NewStringType()

	tests := []struct {
		name  string
		input interface{}
	}{
		{"int", 42},
		{"[]byte", []byte{1, 2, 3}},
		{"bool", true},
		{"nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.Encode(tt.input)
			if err == nil {
				t.Errorf("StringType.Encode(%T) should return error", tt.input)
			}
		})
	}
}

func TestStringType_Decode(t *testing.T) {
	st, _ := NewStringType()

	tests := []struct {
		name   string
		str    string
		offset int
	}{
		{
			name:   "ASCII at offset 0",
			str:    "hello world",
			offset: 0,
		},
		{
			name:   "UTF-8 at offset 0",
			str:    "Hello 世界",
			offset: 0,
		},
		{
			name:   "ASCII at offset 64",
			str:    "test",
			offset: 64,
		},
		{
			name:   "empty string",
			str:    "",
			offset: 0,
		},
		{
			name:   "long string",
			str:    "This is a longer string that will span multiple 32-byte blocks for testing purposes",
			offset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First encode
			encoded, err := st.Encode(tt.str)
			if err != nil {
				t.Fatalf("Setup: encode error = %v", err)
			}

			// Add offset padding if needed
			var testData []byte
			if tt.offset > 0 {
				testData = make([]byte, tt.offset)
				testData = append(testData, encoded...)
			} else {
				testData = encoded
			}

			// Decode
			decoded, err := st.Decode(testData, tt.offset)
			if err != nil {
				t.Errorf("StringType.Decode() error = %v", err)
				return
			}

			result, ok := decoded.(string)
			if !ok {
				t.Errorf("StringType.Decode() returned non-string: %T", decoded)
				return
			}

			if result != tt.str {
				t.Errorf("StringType.Decode() = %q, want %q", result, tt.str)
			}
		})
	}
}

func TestStringType_RoundTrip(t *testing.T) {
	st, _ := NewStringType()

	tests := []struct {
		name string
		str  string
	}{
		{"simple ASCII", "hello"},
		{"with spaces", "hello world"},
		{"with numbers", "test123"},
		{"with symbols", "test!@#$%"},
		{"UTF-8 Japanese", "こんにちは世界"},
		{"UTF-8 emoji", "🎉🎊🎈"},
		{"mixed", "Hello 世界! 🌍"},
		{"empty", ""},
		{"single char", "x"},
		{"exactly 32 bytes", "12345678901234567890123456789012"},
		{"33 bytes", "123456789012345678901234567890123"},
		{"long string", strings.Repeat("test", 50)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			encoded, err := st.Encode(tt.str)
			if err != nil {
				t.Errorf("StringType.Encode() error = %v", err)
				return
			}

			// Decode
			decoded, err := st.Decode(encoded, 0)
			if err != nil {
				t.Errorf("StringType.Decode() error = %v", err)
				return
			}

			result, ok := decoded.(string)
			if !ok {
				t.Errorf("StringType.Decode() returned non-string: %T", decoded)
				return
			}

			if result != tt.str {
				t.Errorf("Round trip failed:\ngot  %q\nwant %q", result, tt.str)
			}
		})
	}
}

// ==================== StaticArrayType Tests ====================

func TestNewStaticArrayType(t *testing.T) {
	tests := []struct {
		name        string
		typeName    string
		wantSize    int
		wantElemTyp string
		wantErr     bool
	}{
		{
			name:        "uint256[3]",
			typeName:    "uint256[3]",
			wantSize:    3,
			wantElemTyp: "uint256",
			wantErr:     false,
		},
		{
			name:        "address[5]",
			typeName:    "address[5]",
			wantSize:    5,
			wantElemTyp: "address",
			wantErr:     false,
		},
		{
			name:     "invalid - no brackets",
			typeName: "uint256",
			wantErr:  true,
		},
		{
			name:     "invalid - zero size",
			typeName: "uint256[0]",
			wantErr:  true,
		},
		{
			name:     "invalid - negative size",
			typeName: "uint256[-1]",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sat, err := NewStaticArrayType(tt.typeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStaticArrayType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if sat.size != tt.wantSize {
				t.Errorf("size = %d, want %d", sat.size, tt.wantSize)
			}
			if sat.elementType.GetCanonicalName() != tt.wantElemTyp {
				t.Errorf("element type = %s, want %s", sat.elementType.GetCanonicalName(), tt.wantElemTyp)
			}
		})
	}
}

func TestStaticArrayType_GetCanonicalName(t *testing.T) {
	sat, _ := NewStaticArrayType("uint256[3]")
	if got := sat.GetCanonicalName(); got != "uint256[3]" {
		t.Errorf("GetCanonicalName() = %s, want uint256[3]", got)
	}
}

func TestStaticArrayType_GetFixedSize(t *testing.T) {
	sat, _ := NewStaticArrayType("uint256[3]")
	// uint256 is 32 bytes, so 3 elements = 96 bytes
	if got := sat.GetFixedSize(); got != 96 {
		t.Errorf("GetFixedSize() = %d, want 96", got)
	}
}

func TestStaticArrayType_IsDynamicType(t *testing.T) {
	sat, _ := NewStaticArrayType("uint256[3]")
	if sat.IsDynamicType() {
		t.Error("StaticArrayType.IsDynamicType() = true, want false")
	}
}

func TestStaticArrayType_Encode(t *testing.T) {
	sat, _ := NewStaticArrayType("uint256[3]")

	tests := []struct {
		name    string
		values  []interface{}
		wantErr bool
	}{
		{
			name:    "valid array",
			values:  []interface{}{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "size mismatch - too few",
			values:  []interface{}{1, 2},
			wantErr: true,
		},
		{
			name:    "size mismatch - too many",
			values:  []interface{}{1, 2, 3, 4},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sat.Encode(tt.values)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStaticArrayType_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		values   []interface{}
	}{
		{
			name:     "uint256[3]",
			typeName: "uint256[3]",
			values:   []interface{}{1, 2, 3},
		},
		{
			name:     "uint8[5]",
			typeName: "uint8[5]",
			values:   []interface{}{10, 20, 30, 40, 50},
		},
		{
			name:     "bool[2]",
			typeName: "bool[2]",
			values:   []interface{}{true, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sat, err := NewStaticArrayType(tt.typeName)
			if err != nil {
				t.Fatalf("NewStaticArrayType() error = %v", err)
			}

			// Encode
			encoded, err := sat.Encode(tt.values)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			// Decode
			decoded, err := sat.Decode(encoded, 0)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			result, ok := decoded.([]interface{})
			if !ok {
				t.Fatalf("Decode() returned non-array: %T", decoded)
			}

			if len(result) != len(tt.values) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.values))
			}

			// Compare values
			for i := range tt.values {
				switch expectedVal := tt.values[i].(type) {
				case int:
					expected := big.NewInt(int64(expectedVal))
					actual, ok := result[i].(*big.Int)
					if !ok {
						t.Errorf("element %d: got type %T, want *big.Int", i, result[i])
						continue
					}
					if expected.Cmp(actual) != 0 {
						t.Errorf("element %d: got %s, want %s", i, actual, expected)
					}
				case bool:
					actual, ok := result[i].(bool)
					if !ok {
						t.Errorf("element %d: got type %T, want bool", i, result[i])
						continue
					}
					if expectedVal != actual {
						t.Errorf("element %d: got %v, want %v", i, actual, expectedVal)
					}
				default:
					t.Errorf("element %d: unsupported test type %T", i, expectedVal)
				}
			}
		})
	}
}

// ==================== DynamicArrayType Tests ====================

func TestNewDynamicArrayType(t *testing.T) {
	tests := []struct {
		name        string
		typeName    string
		wantElemTyp string
		wantErr     bool
	}{
		{
			name:        "uint256[]",
			typeName:    "uint256[]",
			wantElemTyp: "uint256",
			wantErr:     false,
		},
		{
			name:        "address[]",
			typeName:    "address[]",
			wantElemTyp: "address",
			wantErr:     false,
		},
		{
			name:     "invalid - no brackets",
			typeName: "uint256",
			wantErr:  true,
		},
		{
			name:     "invalid - static array",
			typeName: "uint256[5]",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dat, err := NewDynamicArrayType(tt.typeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDynamicArrayType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if dat.elementType.GetCanonicalName() != tt.wantElemTyp {
				t.Errorf("element type = %s, want %s", dat.elementType.GetCanonicalName(), tt.wantElemTyp)
			}
		})
	}
}

func TestDynamicArrayType_GetCanonicalName(t *testing.T) {
	dat, _ := NewDynamicArrayType("uint256[]")
	if got := dat.GetCanonicalName(); got != "uint256[]" {
		t.Errorf("GetCanonicalName() = %s, want uint256[]", got)
	}
}

func TestDynamicArrayType_IsDynamicType(t *testing.T) {
	dat, _ := NewDynamicArrayType("uint256[]")
	if !dat.IsDynamicType() {
		t.Error("DynamicArrayType.IsDynamicType() = false, want true")
	}
}

func TestDynamicArrayType_GetFixedSize(t *testing.T) {
	dat, _ := NewDynamicArrayType("uint256[]")
	if got := dat.GetFixedSize(); got != 0 {
		t.Errorf("GetFixedSize() = %d, want 0 (dynamic type)", got)
	}
}

func TestDynamicArrayType_Encode_EmptyArray(t *testing.T) {
	dat, _ := NewDynamicArrayType("uint256[]")

	encoded, err := dat.Encode([]interface{}{})
	if err != nil {
		t.Fatalf("Encode([]) error = %v", err)
	}

	// Should be 32 bytes for length = 0
	if len(encoded) != 32 {
		t.Errorf("Encoded empty array length = %d, want 32", len(encoded))
	}

	// Verify length is 0
	lengthBig, _ := DecodeInt(encoded, 0)
	if lengthBig.Int64() != 0 {
		t.Errorf("Length = %d, want 0", lengthBig.Int64())
	}
}

func TestDynamicArrayType_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		values   []interface{}
	}{
		{
			name:     "empty array",
			typeName: "uint256[]",
			values:   []interface{}{},
		},
		{
			name:     "single element",
			typeName: "uint256[]",
			values:   []interface{}{42},
		},
		{
			name:     "multiple elements",
			typeName: "uint256[]",
			values:   []interface{}{1, 2, 3, 4, 5},
		},
		{
			name:     "bool array",
			typeName: "bool[]",
			values:   []interface{}{true, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dat, err := NewDynamicArrayType(tt.typeName)
			if err != nil {
				t.Fatalf("NewDynamicArrayType() error = %v", err)
			}

			// Encode
			encoded, err := dat.Encode(tt.values)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			// Decode
			decoded, err := dat.Decode(encoded, 0)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			result, ok := decoded.([]interface{})
			if !ok {
				t.Fatalf("Decode() returned non-array: %T", decoded)
			}

			if len(result) != len(tt.values) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.values))
			}

			// Compare values
			for i := range tt.values {
				switch expectedVal := tt.values[i].(type) {
				case int:
					expected := big.NewInt(int64(expectedVal))
					actual, ok := result[i].(*big.Int)
					if !ok {
						t.Errorf("element %d: got type %T, want *big.Int", i, result[i])
						continue
					}
					if expected.Cmp(actual) != 0 {
						t.Errorf("element %d: got %s, want %s", i, actual, expected)
					}
				case bool:
					actual, ok := result[i].(bool)
					if !ok {
						t.Errorf("element %d: got type %T, want bool", i, result[i])
						continue
					}
					if expectedVal != actual {
						t.Errorf("element %d: got %v, want %v", i, actual, expectedVal)
					}
				default:
					t.Errorf("element %d: unsupported test type %T", i, expectedVal)
				}
			}
		})
	}
}

func TestGetType_Arrays(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		wantType string
	}{
		{
			name:     "static array",
			typeName: "uint256[3]",
			wantType: "*abi.StaticArrayType",
		},
		{
			name:     "dynamic array",
			typeName: "uint256[]",
			wantType: "*abi.DynamicArrayType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abiType, err := GetType(tt.typeName)
			if err != nil {
				t.Fatalf("GetType() error = %v", err)
			}

			gotType := fmt.Sprintf("%T", abiType)
			if gotType != tt.wantType {
				t.Errorf("GetType() type = %s, want %s", gotType, tt.wantType)
			}
		})
	}
}

// ==================== FunctionType Tests ====================

func TestNewFunctionType(t *testing.T) {
	ft, err := NewFunctionType()
	if err != nil {
		t.Fatalf("NewFunctionType() error = %v", err)
	}
	if ft == nil {
		t.Fatal("NewFunctionType() returned nil")
	}
}

func TestFunctionType_GetCanonicalName(t *testing.T) {
	ft, _ := NewFunctionType()
	if got := ft.GetCanonicalName(); got != "function" {
		t.Errorf("FunctionType.GetCanonicalName() = %v, want function", got)
	}
}

func TestFunctionType_GetFixedSize(t *testing.T) {
	ft, _ := NewFunctionType()
	if got := ft.GetFixedSize(); got != 32 {
		t.Errorf("FunctionType.GetFixedSize() = %d, want 32", got)
	}
}

func TestFunctionType_IsDynamicType(t *testing.T) {
	ft, _ := NewFunctionType()
	if ft.IsDynamicType() {
		t.Error("FunctionType.IsDynamicType() = true, want false")
	}
}

func TestFunctionType_Encode(t *testing.T) {
	ft, _ := NewFunctionType()

	tests := []struct {
		name    string
		value   interface{}
		wantHex string
		wantErr bool
	}{
		{
			name:    "24-byte selector from bytes",
			value:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24},
			wantHex: "0102030405060708090a0b0c0d0e0f1011121314151617180000000000000000",
			wantErr: false,
		},
		{
			name:    "24-byte selector from hex string",
			value:   "0x0102030405060708090a0b0c0d0e0f101112131415161718",
			wantHex: "0102030405060708090a0b0c0d0e0f1011121314151617180000000000000000",
			wantErr: false,
		},
		{
			name:    "24-byte selector from hex string without 0x",
			value:   "0102030405060708090a0b0c0d0e0f101112131415161718",
			wantHex: "0102030405060708090a0b0c0d0e0f1011121314151617180000000000000000",
			wantErr: false,
		},
		{
			name:    "invalid - too short",
			value:   []byte{1, 2, 3},
			wantErr: true,
		},
		{
			name:    "invalid - too long",
			value:   make([]byte, 32),
			wantErr: true,
		},
		{
			name:    "invalid - wrong type",
			value:   12345,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := ft.Encode(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("FunctionType.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			gotHex := hex.EncodeToString(encoded)
			if gotHex != tt.wantHex {
				t.Errorf("FunctionType.Encode() = %s, want %s", gotHex, tt.wantHex)
			}
		})
	}
}

func TestFunctionType_Decode(t *testing.T) {
	ft, _ := NewFunctionType()

	// Decode should always return an error (unimplemented)
	encoded := make([]byte, 32)
	_, err := ft.Decode(encoded, 0)
	if err == nil {
		t.Error("FunctionType.Decode() should return error (unimplemented)")
	}
}

func TestGetType_Function(t *testing.T) {
	abiType, err := GetType("function")
	if err != nil {
		t.Fatalf("GetType(\"function\") error = %v", err)
	}

	gotType := fmt.Sprintf("%T", abiType)
	if gotType != "*abi.FunctionType" {
		t.Errorf("GetType(\"function\") type = %s, want *abi.FunctionType", gotType)
	}
}
