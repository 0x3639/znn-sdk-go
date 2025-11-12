package abi

import (
	"bytes"
	"fmt"
	"math/big"
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
		name         string
		encodedHex   string
		offset       int
		wantAddr     string
		wantErr      bool
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
		name  string
		addr  string
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
