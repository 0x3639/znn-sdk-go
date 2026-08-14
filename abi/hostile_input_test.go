package abi

import (
	"math/big"
	"strings"
	"testing"
)

// Regression tests for hostile ABI-encoded input. Offset and length words are
// attacker-controlled whenever this package decodes chain data or calldata;
// every malformed value must surface as an error, never a panic or an
// unbounded allocation.

// Finding #1: DecodeList used the dynamic-type offset word without
// validation, so negative or huge offsets reached the slice expressions.
func TestDecodeListRejectsMaliciousDynamicOffsets(t *testing.T) {
	params := []Param{{Name: "data", Type: mustGetType("bytes")}}

	cases := []struct {
		name   string
		offset *big.Int
	}{
		{"negative offset", big.NewInt(-1)},
		{"offset overflows offset+32", new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 63), big.NewInt(8))},
		{"offset wider than int64 aliases small offset", new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(32))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := append(EncodeIntBig(tc.offset), make([]byte, 32)...)

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("DecodeList panicked on %s: %v", tc.name, r)
				}
			}()

			if _, err := DecodeList(params, payload); err == nil {
				t.Errorf("DecodeList accepted %s without error", tc.name)
			}
		})
	}
}

// Finding #1 (entry point): hostile calldata through Abi.DecodeFunction.
func TestAbiDecodeFunctionRejectsHostileOffsetPayload(t *testing.T) {
	const definition = `[{"name":"f","type":"function","inputs":[{"name":"data","type":"bytes"}]}]`
	target, err := FromJson(definition)
	if err != nil {
		t.Fatalf("FromJson() error = %v", err)
	}

	fn := NewAbiFunction("f", []Param{{Name: "data", Type: mustGetType("bytes")}})
	payload := fn.EncodeSignature()
	payload = append(payload, EncodeIntBig(big.NewInt(-1))...)
	payload = append(payload, make([]byte, 32)...)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Abi.DecodeFunction panicked on hostile calldata: %v", r)
		}
	}()

	if _, err := target.DecodeFunction(payload); err == nil {
		t.Error("Abi.DecodeFunction accepted hostile calldata without error")
	}
}

// Finding #3: DynamicArrayType.Decode allocated the result slice from the
// attacker-controlled length word before checking the payload can hold that
// many elements. The payload here is 32 bytes long.
func TestDynamicArrayDecodeRejectsUnboundedLength(t *testing.T) {
	at := mustGetType("uint256[]")

	cases := []struct {
		name   string
		length *big.Int
	}{
		{"length 2^62", new(big.Int).Lsh(big.NewInt(1), 62)},
		{"length exceeds payload", big.NewInt(1 << 28)},
		{"length wider than int64", new(big.Int).Lsh(big.NewInt(1), 64)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc, err := EncodeUintBig(tc.length)
			if err != nil {
				t.Fatal(err)
			}

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("DynamicArrayType.Decode panicked on oversized length word: %v", r)
				}
			}()
			if _, err := at.Decode(enc, 0); err == nil {
				t.Error("expected error for array length exceeding the available data")
			}
		})
	}
}

// Finding #4: BytesType.Decode's bounds check computed dataOffset+length,
// which overflows int for a length word of 2^63-32, bypassing validation.
func TestBytesDecodeRejectsOverflowingLength(t *testing.T) {
	bt := mustGetType("bytes")

	cases := []struct {
		name   string
		length *big.Int
	}{
		{"length overflows dataOffset+length", new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 63), big.NewInt(Int32Size))},
		{"length wider than int64 aliases small length", new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(5))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc, err := EncodeUintBig(tc.length)
			if err != nil {
				t.Fatal(err)
			}

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("BytesType.Decode panicked on hostile length word: %v", r)
				}
			}()
			if _, err := bt.Decode(enc, 0); err == nil {
				t.Error("expected error for length word that exceeds the available data")
			}
		})
	}
}

// Finding #5: array decoders added the attacker-controlled relative element
// offset to origOffset without validation, so negative offsets reached the
// slice expressions as negative indices.
func TestDynamicArrayDecodeRejectsNegativeElementOffset(t *testing.T) {
	at := mustGetType("string[]")
	negOffset := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 40)) // -2^40
	lengthWord, err := EncodeUintBig(big.NewInt(1))
	if err != nil {
		t.Fatal(err)
	}
	enc := append(lengthWord, EncodeIntBig(negOffset)...) // one element

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("DynamicArrayType.Decode panicked on negative element offset: %v", r)
		}
	}()
	if _, err := at.Decode(enc, 0); err == nil {
		t.Fatal("expected error for out-of-range element offset")
	}
}

func TestStaticArrayDecodeTupleRejectsNegativeElementOffset(t *testing.T) {
	sat, ok := mustGetType("string[1]").(*StaticArrayType)
	if !ok {
		t.Fatal("string[1] did not resolve to *StaticArrayType")
	}
	negOffset := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 40)) // -2^40
	enc := EncodeIntBig(negOffset)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("StaticArrayType.DecodeTuple panicked on negative element offset: %v", r)
		}
	}()
	if _, err := sat.DecodeTuple(enc, 0, 1); err == nil {
		t.Fatal("expected error for out-of-range element offset")
	}
}

// Valid encodings must keep decoding after the bounds checks are added.
func TestHostileInputFixesPreserveValidDecoding(t *testing.T) {
	// bytes round-trip
	bt := mustGetType("bytes")
	encBytes, err := bt.Encode([]byte("hello world"))
	if err != nil {
		t.Fatalf("bytes Encode: %v", err)
	}
	decBytes, err := bt.Decode(encBytes, 0)
	if err != nil {
		t.Fatalf("bytes Decode: %v", err)
	}
	if string(decBytes.([]byte)) != "hello world" {
		t.Fatalf("bytes round-trip = %q", decBytes)
	}

	// dynamic string array round-trip
	at := mustGetType("string[]")
	encArr, err := at.Encode([]string{"alpha", "beta"})
	if err != nil {
		t.Fatalf("string[] Encode: %v", err)
	}
	decArr, err := at.Decode(encArr, 0)
	if err != nil {
		t.Fatalf("string[] Decode: %v", err)
	}
	got := decArr.([]interface{})
	if len(got) != 2 || got[0].(string) != "alpha" || got[1].(string) != "beta" {
		t.Fatalf("string[] round-trip = %v", got)
	}

	// DecodeList with a dynamic param
	params := []Param{{Name: "s", Type: mustGetType("string")}}
	strType := mustGetType("string")
	tail, err := strType.Encode("zenon")
	if err != nil {
		t.Fatalf("string Encode: %v", err)
	}
	payload := append(EncodeInt(Int32Size), tail...)
	decoded, err := DecodeList(params, payload)
	if err != nil {
		t.Fatalf("DecodeList: %v", err)
	}
	if s, ok := decoded[0].(string); !ok || !strings.Contains(s, "zenon") {
		t.Fatalf("DecodeList round-trip = %v", decoded)
	}
}
