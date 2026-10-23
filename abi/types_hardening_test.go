package abi

import (
	"bytes"
	"errors"
	"math"
	"math/big"
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

// word returns a 32-byte big-endian encoding of n.
func word(n int) []byte {
	w := make([]byte, Int32Size)
	new(big.Int).SetInt64(int64(n)).FillBytes(w)
	return w
}

// aliasedBytesArray builds an array-of-bytes encoding (without the dynamic
// length word) whose n element offsets all point at one shared tail of
// tailLen bytes. Decoding it naively yields n×tailLen bytes from a payload of
// roughly 32n+tailLen bytes.
func aliasedBytesArray(n, tailLen int) []byte {
	var out []byte
	for i := 0; i < n; i++ {
		out = append(out, word(n*Int32Size)...)
	}
	out = append(out, word(tailLen)...)
	tail := bytes.Repeat([]byte{0xAB}, tailLen)
	pad := (Int32Size - tailLen%Int32Size) % Int32Size
	out = append(out, tail...)
	out = append(out, make([]byte, pad)...)
	return out
}

// Codex finding (#8 residual): many element offsets aliasing one dynamic tail
// must not amplify the decoded size beyond the encoded payload.
func TestArrayDecode_RejectsAliasedDynamicTails(t *testing.T) {
	const n, tailLen = 64, 4096
	body := aliasedBytesArray(n, tailLen)

	dyn, err := NewDynamicArrayType("bytes[]")
	if err != nil {
		t.Fatal(err)
	}
	encoded := append(word(n), body...)
	if _, decodeErr := dyn.Decode(encoded, 0); decodeErr == nil {
		t.Errorf("bytes[] decoded %d aliased elements (%d bytes) from %d encoded bytes", n, n*tailLen, len(encoded))
	}

	static, err := NewStaticArrayType("bytes[64]")
	if err != nil {
		t.Fatal(err)
	}
	if _, decodeErr := static.Decode(body, 0); decodeErr == nil {
		t.Errorf("bytes[64] decoded %d aliased elements (%d bytes) from %d encoded bytes", n, n*tailLen, len(body))
	}

	// Nested arrays amplify the same way: uint256[][] whose outer offsets all
	// alias one inner array of 64 words.
	nested, err := NewDynamicArrayType("uint256[][]")
	if err != nil {
		t.Fatal(err)
	}
	var inner []byte
	inner = append(inner, word(64)...)
	for i := 0; i < 64; i++ {
		inner = append(inner, word(i)...)
	}
	var outer []byte
	outer = append(outer, word(256)...)
	for i := 0; i < 256; i++ {
		outer = append(outer, word(256*Int32Size)...)
	}
	outer = append(outer, inner...)
	if _, err := nested.Decode(outer, 0); err == nil {
		t.Errorf("uint256[][] decoded 256 aliased inner arrays from %d encoded bytes", len(outer))
	}
}

// A canonical bytes[] encoding (distinct, non-overlapping tails) must still
// decode after the aliasing guard is in place.
func TestArrayDecode_CanonicalBytesArrayStillDecodes(t *testing.T) {
	dyn, err := NewDynamicArrayType("bytes[]")
	if err != nil {
		t.Fatal(err)
	}
	first := []byte("hello")
	second := bytes.Repeat([]byte{0x11}, 40)
	encoded := make([]byte, 0, 7*Int32Size)
	encoded = append(encoded, word(2)...)
	encoded = append(encoded, word(2*Int32Size)...)             // first tail right after heads
	encoded = append(encoded, word(2*Int32Size+2*Int32Size)...) // second tail after first (32 len + 32 padded data)
	encoded = append(encoded, word(len(first))...)
	encoded = append(encoded, first...)
	encoded = append(encoded, make([]byte, Int32Size-len(first))...)
	encoded = append(encoded, word(len(second))...)
	encoded = append(encoded, second...)
	encoded = append(encoded, make([]byte, 2*Int32Size-len(second))...)

	decoded, err := dyn.Decode(encoded, 0)
	if err != nil {
		t.Fatalf("canonical bytes[] rejected: %v", err)
	}
	list, ok := decoded.([]interface{})
	if !ok || len(list) != 2 {
		t.Fatalf("decoded = %#v, want 2 elements", decoded)
	}
	if got, _ := list[0].([]byte); !bytes.Equal(got, first) {
		t.Errorf("element 0 = %x, want %x", got, first)
	}
	if got, _ := list[1].([]byte); !bytes.Equal(got, second) {
		t.Errorf("element 1 = %x, want %x", got, second)
	}
}

// aliasedEmptyBytesMatrix builds a bytes[][] body (without the outer length
// word): n outer offsets alias one inner bytes[] whose n offsets alias one
// empty bytes value. Each inner decode allocates n result slots while every
// leaf is zero bytes.
func aliasedEmptyBytesMatrix(n int) []byte {
	out := make([]byte, 0, (2*n+3)*Int32Size)
	for i := 0; i < n; i++ {
		out = append(out, word(n*Int32Size)...) // inner array right after outer heads
	}
	out = append(out, word(n)...) // inner length
	for i := 0; i < n; i++ {
		out = append(out, word(n*Int32Size)...) // empty bytes right after inner heads
	}
	out = append(out, word(0)...) // empty bytes
	return out
}

// Codex finding: empty values and container storage must be charged, or
// nested aliasing of zero-length leaves allocates n² result slots.
func TestArrayDecode_RejectsAliasedEmptyNestedValues(t *testing.T) {
	const n = 4096
	body := aliasedEmptyBytesMatrix(n)

	dyn, err := NewDynamicArrayType("bytes[][]")
	if err != nil {
		t.Fatal(err)
	}
	encoded := append(word(n), body...)
	if _, decodeErr := dyn.Decode(encoded, 0); decodeErr == nil {
		t.Errorf("bytes[][] decoded %d×%d aliased empty elements from %d encoded bytes", n, n, len(encoded))
	}

	static, err := NewStaticArrayType("bytes[4096][]")
	if err != nil {
		t.Fatal(err)
	}
	if _, decodeErr := static.Decode(body, 0); decodeErr == nil {
		t.Errorf("bytes[4096][] decoded %d×%d aliased empty elements from %d encoded bytes", n, n, len(body))
	}
}

// Canonical nested encodings, including empty leaves and a non-zero origin,
// must still decode under the budget.
func TestArrayDecode_CanonicalNestedArraysStillDecode(t *testing.T) {
	// bytes[][] = [["", "ab"], []] encoded canonically, preceded by one junk
	// word so the tuple origin is non-zero.
	inner0 := bytes.Join([][]byte{
		word(2),
		word(2 * Int32Size), // "" tail
		word(3 * Int32Size), // "ab" tail
		word(0),             // ""
		word(2),             // "ab" length
		[]byte("ab"), make([]byte, Int32Size-2),
	}, nil)
	inner1 := word(0) // []
	outer := bytes.Join([][]byte{
		word(2),
		word(2 * Int32Size),
		word(2*Int32Size + len(inner0)),
		inner0,
		inner1,
	}, nil)
	encoded := append(word(0xdead), outer...)

	nested, err := NewDynamicArrayType("bytes[][]")
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := nested.Decode(encoded, Int32Size)
	if err != nil {
		t.Fatalf("canonical bytes[][] rejected: %v", err)
	}
	list := decoded.([]interface{})
	if len(list) != 2 || len(list[0].([]interface{})) != 2 || len(list[1].([]interface{})) != 0 {
		t.Fatalf("decoded = %#v", decoded)
	}
	if got := list[0].([]interface{})[1].([]byte); string(got) != "ab" {
		t.Errorf("element [0][1] = %q, want \"ab\"", got)
	}

	// uint256[][2] (dynamic array of uint256[2], outer dimension first in this
	// SDK) = [[1,2],[3,4]]: static inner arrays inline, no length words.
	fixedNested := bytes.Join([][]byte{word(2), word(1), word(2), word(3), word(4)}, nil)
	fixedType, err := NewDynamicArrayType("uint256[][2]")
	if err != nil {
		t.Fatal(err)
	}
	if _, decodeErr := fixedType.Decode(fixedNested, 0); decodeErr != nil {
		t.Fatalf("canonical uint256[][2] rejected: %v", decodeErr)
	}

	// string[2] (static array of dynamic elements) = ["", "xyz"].
	strs := bytes.Join([][]byte{
		word(2 * Int32Size),
		word(3 * Int32Size),
		word(0),
		word(3), []byte("xyz"), make([]byte, Int32Size-3),
	}, nil)
	strType, err := NewStaticArrayType("string[2]")
	if err != nil {
		t.Fatal(err)
	}
	if _, decodeErr := strType.Decode(strs, 0); decodeErr != nil {
		t.Fatalf("canonical string[2] rejected: %v", decodeErr)
	}
}

// Codex finding (round 3): a static wrapper around a dynamic array reports
// itself as a fixed type of size zero, so the fixed-element branch decoded
// the same offset repeatedly without charging the budget.
func TestArrayDecode_RejectsAliasingThroughStaticWrapper(t *testing.T) {
	const n = 4096
	// bytes[1][] element at offset 32: its one offset word (32, relative to
	// the wrapper origin) points at a bytes[] of n offsets aliasing one empty
	// bytes value.
	var innerOffsets [][]byte
	for i := 0; i < n; i++ {
		innerOffsets = append(innerOffsets, word(n*Int32Size))
	}
	body := bytes.Join(append([][]byte{word(Int32Size), word(n)}, append(innerOffsets, word(0))...), nil)

	dyn, err := NewDynamicArrayType("bytes[][1][]")
	if err != nil {
		t.Fatal(err)
	}
	encoded := append(word(n), body...)
	if _, decodeErr := dyn.Decode(encoded, 0); decodeErr == nil {
		t.Errorf("bytes[][1][] decoded %d aliased static wrappers from %d encoded bytes", n, len(encoded))
	}

	static, err := NewStaticArrayType("bytes[4096][1][]")
	if err != nil {
		t.Fatal(err)
	}
	if _, decodeErr := static.Decode(body, 0); decodeErr == nil {
		t.Errorf("bytes[4096][1][] decoded %d aliased static wrappers from %d encoded bytes", n, len(body))
	}
}

// Codex finding (round 3): fixed-width bytes decode to []byte but occupy one
// word, so charging them as dynamic bytes over-counted canonical nested
// fixed-byte arrays.
func TestArrayDecode_CanonicalNestedFixedBytesStillDecode(t *testing.T) {
	fixedWords := func(width, count int) [][]byte {
		var words [][]byte
		for i := 0; i < count; i++ {
			w := make([]byte, Int32Size)
			for j := 0; j < width; j++ {
				w[j] = byte(i + 1)
			}
			words = append(words, w)
		}
		return words
	}
	cases := map[string]struct {
		typeName string
		static   bool
		encoded  []byte
		want     int // elements in the single inner array
		width    int // fixed byte width of each element
	}{
		"bytes8[][] 9 values": {
			typeName: "bytes8[][]",
			encoded:  bytes.Join(append([][]byte{word(1), word(Int32Size), word(9)}, fixedWords(8, 9)...), nil),
			want:     9,
			width:    8,
		},
		"bytes8[1][] 9 values": {
			typeName: "bytes8[1][]",
			static:   true,
			encoded:  bytes.Join(append([][]byte{word(Int32Size), word(9)}, fixedWords(8, 9)...), nil),
			want:     9,
			width:    8,
		},
		"bytes32[][] 3 values": {
			typeName: "bytes32[][]",
			encoded:  bytes.Join(append([][]byte{word(1), word(Int32Size), word(3)}, fixedWords(32, 3)...), nil),
			want:     3,
			width:    32,
		},
	}
	for name, tc := range cases {
		var typ AbiType
		var err error
		if tc.static {
			typ, err = NewStaticArrayType(tc.typeName)
		} else {
			typ, err = NewDynamicArrayType(tc.typeName)
		}
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		decoded, err := typ.Decode(tc.encoded, 0)
		if err != nil {
			t.Errorf("%s: canonical encoding rejected: %v", name, err)
			continue
		}
		outer := decoded.([]interface{})
		if len(outer) != 1 || len(outer[0].([]interface{})) != tc.want {
			t.Errorf("%s: decoded = %#v, want 1 inner array of %d", name, decoded, tc.want)
			continue
		}
		// Values must be the fixed-width prefix of each word, left-aligned.
		for i, element := range outer[0].([]interface{}) {
			got := element.([]byte)
			if len(got) != tc.width || !bytes.Equal(got, bytes.Repeat([]byte{byte(i + 1)}, tc.width)) {
				t.Errorf("%s: element %d = %x, want %d bytes of %#x", name, i, got, tc.width, i+1)
			}
		}
	}
}

// Codex finding (round 4): an array of static wrappers around dynamic
// elements has no well-defined stride (the wrapper is encoded inline with a
// fixed size of zero, matching go-zenon's ABI), so every element decoded the
// same offset and returned duplicated values. That shape is now rejected
// with ErrUnsupportedArrayShape instead of silently returning wrong data or a
// misleading aliasing error.
func TestArrayDecode_RejectsZeroStrideStaticWrappers(t *testing.T) {
	// SDK encoding of uint256[][1][] = [[[1,2,3,4,5]], [[]]] (Codex's example):
	// words [2, 32, 5, 1, 2, 3, 4, 5, 32, 0]. Both wrappers are inline and
	// unequal in length, so there is no stride that reaches the second.
	body := bytes.Join([][]byte{word(32), word(5), word(1), word(2), word(3), word(4), word(5), word(32), word(0)}, nil)

	dyn, err := NewDynamicArrayType("uint256[][1][]")
	if err != nil {
		t.Fatal(err)
	}
	if _, decodeErr := dyn.Decode(append(word(2), body...), 0); !errors.Is(decodeErr, ErrUnsupportedArrayShape) {
		t.Errorf("uint256[][1][] with 2 wrappers: err = %v, want ErrUnsupportedArrayShape", decodeErr)
	}
	static, err := NewStaticArrayType("uint256[2][1][]")
	if err != nil {
		t.Fatal(err)
	}
	if _, decodeErr := static.Decode(body, 0); !errors.Is(decodeErr, ErrUnsupportedArrayShape) {
		t.Errorf("uint256[2][1][] with 2 wrappers: err = %v, want ErrUnsupportedArrayShape", decodeErr)
	}

	// A single wrapper needs no stride and still decodes.
	single := bytes.Join([][]byte{word(1), word(32), word(3), word(7), word(8), word(9)}, nil)
	decoded, err := dyn.Decode(single, 0)
	if err != nil {
		t.Fatalf("uint256[][1][] with 1 wrapper rejected: %v", err)
	}
	inner := decoded.([]interface{})[0].([]interface{})[0].([]interface{})
	if len(inner) != 3 || inner[2].(*big.Int).Int64() != 9 {
		t.Errorf("decoded = %#v, want [[[7 8 9]]]", decoded)
	}
}
