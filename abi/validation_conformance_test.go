package abi

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"
)

func TestFixedBytesWidthsAndExactLengths(t *testing.T) {
	t.Parallel()
	for size := 1; size <= 32; size++ {
		size := size
		t.Run(fmt.Sprintf("bytes%d", size), func(t *testing.T) {
			typeObject, err := GetType(fmt.Sprintf("bytes%d", size))
			if err != nil {
				t.Fatalf("GetType() error = %v", err)
			}
			value := make([]byte, size)
			for index := range value {
				value[index] = byte(index + 1)
			}
			encoded, err := typeObject.Encode(value)
			if err != nil {
				t.Fatalf("Encode(exact) error = %v", err)
			}
			if len(encoded) != Int32Size || !bytes.Equal(encoded[:size], value) ||
				!bytes.Equal(encoded[size:], make([]byte, Int32Size-size)) {
				t.Fatalf("Encode(exact) = %x", encoded)
			}
			if _, encodeErr := typeObject.Encode(value[:size-1]); encodeErr == nil {
				t.Fatal("Encode(short) accepted a non-exact value")
			}
			if _, encodeErr := typeObject.Encode(append(value, 0xff)); encodeErr == nil {
				t.Fatal("Encode(long) accepted a non-exact value")
			}

			decoded, err := typeObject.Decode(encoded, 0)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got, ok := decoded.([]byte); !ok || !bytes.Equal(got, value) {
				t.Fatalf("Decode() = %x, want %x", got, value)
			}
			if size < Int32Size {
				nonCanonical := append([]byte(nil), encoded...)
				nonCanonical[size] = 1
				if _, err := typeObject.Decode(nonCanonical, 0); err == nil {
					t.Fatal("Decode() accepted non-zero right padding")
				}
			}
		})
	}
}

func TestIntegerBoundsForEveryWidth(t *testing.T) {
	t.Parallel()
	for width := 8; width <= 256; width += 8 {
		width := width
		t.Run(fmt.Sprintf("uint%d", width), func(t *testing.T) {
			typeObject, err := GetType(fmt.Sprintf("uint%d", width))
			if err != nil {
				t.Fatalf("GetType() error = %v", err)
			}
			maximum := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(width)), big.NewInt(1))
			for _, value := range []*big.Int{big.NewInt(0), maximum} {
				encoded, err := typeObject.Encode(value)
				if err != nil {
					t.Fatalf("Encode(%s) error = %v", value, err)
				}
				decoded, err := typeObject.Decode(encoded, 0)
				if err != nil {
					t.Fatalf("Decode(%s) error = %v", value, err)
				}
				if decoded.(*big.Int).Cmp(value) != 0 {
					t.Fatalf("round trip = %s, want %s", decoded, value)
				}
			}
			if _, err := typeObject.Encode(big.NewInt(-1)); err == nil {
				t.Fatal("Encode(-1) accepted a negative unsigned value")
			}
			overflow := new(big.Int).Add(maximum, big.NewInt(1))
			if _, err := typeObject.Encode(overflow); err == nil {
				t.Fatalf("Encode(%s) accepted unsigned overflow", overflow)
			}
			if width < 256 {
				word := bigIntToBytes(overflow, Int32Size)
				if _, err := typeObject.Decode(word, 0); err == nil {
					t.Fatal("Decode() accepted a word outside the declared unsigned width")
				}
			}
		})

		t.Run(fmt.Sprintf("int%d", width), func(t *testing.T) {
			typeObject, err := GetType(fmt.Sprintf("int%d", width))
			if err != nil {
				t.Fatalf("GetType() error = %v", err)
			}
			minimum := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), uint(width-1)))
			maximum := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(width-1)), big.NewInt(1))
			for _, value := range []*big.Int{minimum, maximum} {
				encoded, err := typeObject.Encode(value)
				if err != nil {
					t.Fatalf("Encode(%s) error = %v", value, err)
				}
				decoded, err := typeObject.Decode(encoded, 0)
				if err != nil {
					t.Fatalf("Decode(%s) error = %v", value, err)
				}
				if decoded.(*big.Int).Cmp(value) != 0 {
					t.Fatalf("round trip = %s, want %s", decoded, value)
				}
			}
			underflow := new(big.Int).Sub(minimum, big.NewInt(1))
			overflow := new(big.Int).Add(maximum, big.NewInt(1))
			if _, err := typeObject.Encode(underflow); err == nil {
				t.Fatalf("Encode(%s) accepted signed underflow", underflow)
			}
			if _, err := typeObject.Encode(overflow); err == nil {
				t.Fatalf("Encode(%s) accepted signed overflow", overflow)
			}
			if width < 256 {
				if _, err := typeObject.Decode(EncodeIntBig(overflow), 0); err == nil {
					t.Fatal("Decode() accepted a word outside the declared signed width")
				}
			}
		})
	}
}

func TestFunctionEncodingRequiresExactArity(t *testing.T) {
	t.Parallel()
	param, err := NewParam("value", "uint8")
	if err != nil {
		t.Fatalf("NewParam() error = %v", err)
	}
	function := NewAbiFunction("Set", []Param{*param})
	if _, err := function.Encode(nil); err == nil {
		t.Fatal("Encode() accepted a missing argument")
	}
	if _, err := function.Encode([]interface{}{1, 2}); err == nil {
		t.Fatal("Encode() accepted an extra argument")
	}
	if _, err := function.Encode([]interface{}{1}); err != nil {
		t.Fatalf("Encode() rejected exact arity: %v", err)
	}
}
