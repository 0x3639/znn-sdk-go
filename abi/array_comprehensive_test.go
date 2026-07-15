package abi

import (
	"reflect"
	"testing"
)

func TestTypeEnumStringValues(t *testing.T) {
	if Function.String() != "function" {
		t.Fatalf("Function.String() = %q", Function.String())
	}
	if TypeEnum(99).String() != "unknown" {
		t.Fatalf("unknown TypeEnum.String() = %q", TypeEnum(99).String())
	}
}

func TestStaticArrayDynamicElementTupleRoundTrip(t *testing.T) {
	array, err := NewStaticArrayType("string[2]")
	if err != nil {
		t.Fatal(err)
	}
	if array.GetElementType().GetCanonicalName() != "string" {
		t.Fatalf("element type = %s", array.GetElementType().GetCanonicalName())
	}
	encoded, err := array.Encode([]string{"alpha", "beta"})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	decoded, err := array.DecodeTuple(encoded, 0, 2)
	if err != nil {
		t.Fatalf("DecodeTuple: %v", err)
	}
	if !reflect.DeepEqual(decoded, []interface{}{"alpha", "beta"}) {
		t.Fatalf("decoded = %#v", decoded)
	}
	direct, err := array.Decode(encoded, 0)
	if err != nil || !reflect.DeepEqual(direct, decoded) {
		t.Fatalf("Decode = %#v, %v", direct, err)
	}
}

func TestDynamicArrayDynamicElementRoundTrip(t *testing.T) {
	array, err := NewDynamicArrayType("string[]")
	if err != nil {
		t.Fatal(err)
	}
	if array.GetElementType().GetCanonicalName() != "string" {
		t.Fatalf("element type = %s", array.GetElementType().GetCanonicalName())
	}
	encoded, err := array.Encode([]string{"alpha", "beta"})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	decoded, err := array.Decode(encoded, 0)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !reflect.DeepEqual(decoded, []interface{}{"alpha", "beta"}) {
		t.Fatalf("decoded = %#v", decoded)
	}
	tuple, err := array.DecodeTuple(encoded[32:], 0, 2)
	if err != nil {
		t.Fatalf("DecodeTuple: %v", err)
	}
	if !reflect.DeepEqual(tuple, decoded) {
		t.Fatalf("tuple = %#v, decoded = %#v", tuple, decoded)
	}
}

func TestArrayTypedSliceAndErrorPaths(t *testing.T) {
	static, err := NewStaticArrayType("uint256[2]")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := static.Encode([]int{1, 2}); err != nil {
		t.Fatalf("static typed slice: %v", err)
	}
	if _, err := static.Encode(42); err == nil {
		t.Fatal("static array accepted a scalar")
	}
	if _, err := static.EncodeList([]interface{}{1}); err == nil {
		t.Fatal("static array accepted the wrong element count")
	}

	dynamic, err := NewDynamicArrayType("uint256[]")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dynamic.Encode([]uint64{1, 2}); err != nil {
		t.Fatalf("dynamic typed slice: %v", err)
	}
	if _, err := dynamic.Encode(42); err == nil {
		t.Fatal("dynamic array accepted a scalar")
	}

	staticStrings, _ := NewStaticArrayType("string[2]")
	if _, err := staticStrings.EncodeTuple([]interface{}{"valid", 42}); err == nil {
		t.Fatal("static tuple accepted an invalid element")
	}
	dynamicStrings, _ := NewDynamicArrayType("string[]")
	if _, err := dynamicStrings.EncodeTuple([]interface{}{"valid", 42}); err == nil {
		t.Fatal("dynamic tuple accepted an invalid element")
	}
	if _, err := staticStrings.DecodeTuple([]byte{1}, 0, 1); err == nil {
		t.Fatal("static tuple decoded truncated data")
	}
	if _, err := dynamicStrings.DecodeTuple([]byte{1}, 0, 1); err == nil {
		t.Fatal("dynamic tuple decoded truncated data")
	}
	negativeLength := EncodeInt(-1)
	if _, err := dynamic.Decode(negativeLength, 0); err == nil {
		t.Fatal("dynamic array accepted a negative length")
	}
}
