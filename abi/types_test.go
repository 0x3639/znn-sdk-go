package abi

import (
	"bytes"
	"math/big"
	"testing"
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
