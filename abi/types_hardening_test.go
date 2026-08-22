package abi

import (
	"math"
	"strings"
	"testing"
)

// CodeRabbit finding: negative/overflowing offsets must return errors, never panic.
func TestDecoders_RejectNegativeAndOverflowingOffsets(t *testing.T) {
	encoded := make([]byte, 64)
	addr, _ := NewAddressType()
	hsh, _ := NewHashType()
	fb, _ := NewFixedBytesType("bytes8")
	tst, _ := NewTokenStandardType()
	decoders := map[string]func(int) (interface{}, error){
		"uint":          func(o int) (interface{}, error) { return DecodeUint(encoded, o) },
		"int":           func(o int) (interface{}, error) { return DecodeInt(encoded, o) },
		"address":       func(o int) (interface{}, error) { return addr.Decode(encoded, o) },
		"hash":          func(o int) (interface{}, error) { return hsh.Decode(encoded, o) },
		"fixedBytes":    func(o int) (interface{}, error) { return fb.Decode(encoded, o) },
		"tokenStandard": func(o int) (interface{}, error) { return tst.Decode(encoded, o) },
	}
	for name, decode := range decoders {
		for _, offset := range []int{-1, math.MinInt, math.MaxInt, math.MaxInt - 10, 33} {
			if _, err := decode(offset); err == nil {
				t.Errorf("%s: offset %d decoded without error", name, offset)
			}
		}
		if _, err := decode(32); err != nil {
			t.Errorf("%s: valid offset 32 failed: %v", name, err)
		}
	}
}

// CodeRabbit finding: array type parsing must bound nesting and length.
func TestGetType_BoundsNestingAndLength(t *testing.T) {
	if _, err := GetType("uint256" + strings.Repeat("[]", MaxArrayNesting+1)); err == nil {
		t.Error("GetType accepted nesting above MaxArrayNesting")
	}
	if _, err := GetType("uint256" + strings.Repeat("[1]", MaxTypeNameLength)); err == nil {
		t.Error("GetType accepted over-long declaration")
	}
	if _, err := GetType("uint256[][]"); err != nil {
		t.Errorf("GetType rejected a valid nested declaration: %v", err)
	}
	if _, err := GetType("uint256[99999999]"); err == nil {
		t.Error("GetType accepted static array size above MaxStaticArraySize")
	}
}

// CodeRabbit finding: tuple lengths must be validated against the encoded
// head before allocating the result slice.
func TestDecodeTuple_RejectsLengthBeyondData(t *testing.T) {
	dyn, err := NewDynamicArrayType("uint256[]")
	if err != nil {
		t.Fatal(err)
	}
	encoded := make([]byte, 64)
	for _, length := range []int{-1, 3, math.MaxInt} {
		if _, tupleErr := dyn.DecodeTuple(encoded, 0, length); tupleErr == nil {
			t.Errorf("DynamicArrayType.DecodeTuple accepted length %d", length)
		}
	}
	static, err := NewStaticArrayType("uint256[4]")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := static.Decode(encoded, 0); err == nil {
		t.Error("StaticArrayType.Decode accepted 4 elements with only 2 words of data")
	}
	if _, err := static.DecodeTuple(encoded, 0, math.MaxInt); err == nil {
		t.Error("StaticArrayType.DecodeTuple accepted MaxInt length")
	}
}
