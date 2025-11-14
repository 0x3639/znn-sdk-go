package abi

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/zenon-network/go-zenon/common/types"
)

// ==================== Param Tests ====================

func TestNewParam(t *testing.T) {
	tests := []struct {
		name      string
		paramName string
		typeName  string
		wantErr   bool
	}{
		{
			name:      "uint256 param",
			paramName: "amount",
			typeName:  "uint256",
			wantErr:   false,
		},
		{
			name:      "address param",
			paramName: "recipient",
			typeName:  "address",
			wantErr:   false,
		},
		{
			name:      "invalid type",
			paramName: "test",
			typeName:  "invalid",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param, err := NewParam(tt.paramName, tt.typeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if param.Name != tt.paramName {
				t.Errorf("param.Name = %s, want %s", param.Name, tt.paramName)
			}
			if param.Indexed {
				t.Error("param.Indexed should be false by default")
			}
		})
	}
}

func TestDecodeList(t *testing.T) {
	// Create test parameters
	params := []Param{
		{Name: "a", Type: mustGetType("uint256")},
		{Name: "b", Type: mustGetType("uint256")},
	}

	// Encode test data: [1, 2]
	encoded := make([]byte, 64)
	copy(encoded[0:32], EncodeInt(1))
	copy(encoded[32:64], EncodeInt(2))

	// Decode
	result, err := DecodeList(params, encoded)
	if err != nil {
		t.Fatalf("DecodeList() error = %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("len(result) = %d, want 2", len(result))
	}

	// Check values
	val1, ok := result[0].(*big.Int)
	if !ok {
		t.Fatalf("result[0] type = %T, want *big.Int", result[0])
	}
	if val1.Int64() != 1 {
		t.Errorf("result[0] = %d, want 1", val1.Int64())
	}

	val2, ok := result[1].(*big.Int)
	if !ok {
		t.Fatalf("result[1] type = %T, want *big.Int", result[1])
	}
	if val2.Int64() != 2 {
		t.Errorf("result[1] = %d, want 2", val2.Int64())
	}
}

// ==================== Entry Tests ====================

func TestNewEntry(t *testing.T) {
	params := []Param{
		{Name: "x", Type: mustGetType("uint256")},
	}

	entry := NewEntry("test", params, Function)

	if entry.Name != "test" {
		t.Errorf("entry.Name = %s, want test", entry.Name)
	}
	if len(entry.Inputs) != 1 {
		t.Errorf("len(entry.Inputs) = %d, want 1", len(entry.Inputs))
	}
	if entry.Type != Function {
		t.Errorf("entry.Type = %v, want Function", entry.Type)
	}
}

func TestEntry_FormatSignature(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		params   []Param
		wantSig  string
	}{
		{
			name:     "no params",
			funcName: "test",
			params:   []Param{},
			wantSig:  "test()",
		},
		{
			name:     "single param",
			funcName: "transfer",
			params: []Param{
				{Name: "amount", Type: mustGetType("uint256")},
			},
			wantSig: "transfer(uint256)",
		},
		{
			name:     "multiple params",
			funcName: "transferFrom",
			params: []Param{
				{Name: "from", Type: mustGetType("address")},
				{Name: "to", Type: mustGetType("address")},
				{Name: "amount", Type: mustGetType("uint256")},
			},
			wantSig: "transferFrom(address,address,uint256)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := NewEntry(tt.funcName, tt.params, Function)
			sig := entry.FormatSignature()
			if sig != tt.wantSig {
				t.Errorf("FormatSignature() = %s, want %s", sig, tt.wantSig)
			}
		})
	}
}

func TestEntry_FingerprintSignature(t *testing.T) {
	// Test with transfer(address,uint256)
	params := []Param{
		{Name: "to", Type: mustGetType("address")},
		{Name: "value", Type: mustGetType("uint256")},
	}
	entry := NewEntry("transfer", params, Function)

	fingerprint := entry.FingerprintSignature()

	// Should be 32 bytes (SHA3-256)
	if len(fingerprint) != 32 {
		t.Errorf("len(fingerprint) = %d, want 32", len(fingerprint))
	}

	// The fingerprint should be deterministic
	fingerprint2 := entry.FingerprintSignature()
	if hex.EncodeToString(fingerprint) != hex.EncodeToString(fingerprint2) {
		t.Error("Fingerprint should be deterministic")
	}
}

func TestEntry_EncodeArguments_StaticOnly(t *testing.T) {
	// Function with only static types: test(uint256,uint256)
	params := []Param{
		{Name: "a", Type: mustGetType("uint256")},
		{Name: "b", Type: mustGetType("uint256")},
	}
	entry := NewEntry("test", params, Function)

	args := []interface{}{42, 100}
	encoded, err := entry.EncodeArguments(args)
	if err != nil {
		t.Fatalf("EncodeArguments() error = %v", err)
	}

	// Should be 64 bytes (2 * 32)
	if len(encoded) != 64 {
		t.Errorf("len(encoded) = %d, want 64", len(encoded))
	}

	// Verify values
	val1, _ := DecodeInt(encoded, 0)
	if val1.Int64() != 42 {
		t.Errorf("first arg = %d, want 42", val1.Int64())
	}

	val2, _ := DecodeInt(encoded, 32)
	if val2.Int64() != 100 {
		t.Errorf("second arg = %d, want 100", val2.Int64())
	}
}

func TestEntry_EncodeArguments_WithDynamic(t *testing.T) {
	// Function with dynamic type: test(uint256,string)
	params := []Param{
		{Name: "id", Type: mustGetType("uint256")},
		{Name: "name", Type: mustGetType("string")},
	}
	entry := NewEntry("test", params, Function)

	args := []interface{}{123, "hello"}
	encoded, err := entry.EncodeArguments(args)
	if err != nil {
		t.Fatalf("EncodeArguments() error = %v", err)
	}

	// Head: 32 bytes (uint256) + 32 bytes (offset pointer)
	// Tail: 32 bytes (string length) + 32 bytes (string data padded)
	// Total: 128 bytes
	expectedLen := 128
	if len(encoded) != expectedLen {
		t.Errorf("len(encoded) = %d, want %d", len(encoded), expectedLen)
	}

	// First arg should be 123
	val1, _ := DecodeInt(encoded, 0)
	if val1.Int64() != 123 {
		t.Errorf("first arg = %d, want 123", val1.Int64())
	}

	// Second arg should be an offset pointer (64)
	offset, _ := DecodeInt(encoded, 32)
	if offset.Int64() != 64 {
		t.Errorf("offset = %d, want 64", offset.Int64())
	}
}

// ==================== AbiFunction Tests ====================

func TestNewAbiFunction(t *testing.T) {
	params := []Param{
		{Name: "x", Type: mustGetType("uint256")},
	}

	fn := NewAbiFunction("test", params)

	if fn.Name != "test" {
		t.Errorf("fn.Name = %s, want test", fn.Name)
	}
	if fn.Type != Function {
		t.Errorf("fn.Type = %v, want Function", fn.Type)
	}
}

func TestAbiFunction_EncodeSignature(t *testing.T) {
	// Test with transfer(address,uint256)
	params := []Param{
		{Name: "to", Type: mustGetType("address")},
		{Name: "value", Type: mustGetType("uint256")},
	}
	fn := NewAbiFunction("transfer", params)

	sig := fn.EncodeSignature()

	// Should be 4 bytes
	if len(sig) != EncodedSignLength {
		t.Errorf("len(sig) = %d, want %d", len(sig), EncodedSignLength)
	}
}

func TestAbiFunction_Encode(t *testing.T) {
	// Simple function: test(uint256)
	params := []Param{
		{Name: "x", Type: mustGetType("uint256")},
	}
	fn := NewAbiFunction("test", params)

	args := []interface{}{42}
	encoded, err := fn.Encode(args)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Should be 4 bytes (signature) + 32 bytes (uint256) = 36 bytes
	expectedLen := 4 + 32
	if len(encoded) != expectedLen {
		t.Errorf("len(encoded) = %d, want %d", len(encoded), expectedLen)
	}

	// First 4 bytes should be the signature
	signature := encoded[:4]
	expectedSig := fn.EncodeSignature()
	if hex.EncodeToString(signature) != hex.EncodeToString(expectedSig) {
		t.Error("First 4 bytes should be function signature")
	}

	// Remaining bytes should be the encoded argument
	val, _ := DecodeInt(encoded, 4)
	if val.Int64() != 42 {
		t.Errorf("encoded arg = %d, want 42", val.Int64())
	}
}

func TestAbiFunction_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		params   []Param
		args     []interface{}
	}{
		{
			name:     "single uint256",
			funcName: "setValue",
			params: []Param{
				{Name: "value", Type: mustGetType("uint256")},
			},
			args: []interface{}{100},
		},
		{
			name:     "multiple params",
			funcName: "setValues",
			params: []Param{
				{Name: "a", Type: mustGetType("uint256")},
				{Name: "b", Type: mustGetType("bool")},
			},
			args: []interface{}{42, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := NewAbiFunction(tt.funcName, tt.params)

			// Encode
			encoded, err := fn.Encode(tt.args)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}

			// Decode
			decoded, err := fn.Decode(encoded)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}

			if len(decoded) != len(tt.args) {
				t.Fatalf("len(decoded) = %d, want %d", len(decoded), len(tt.args))
			}

			// Compare values
			for i := range tt.args {
				switch expected := tt.args[i].(type) {
				case int:
					actual, ok := decoded[i].(*big.Int)
					if !ok {
						t.Errorf("decoded[%d] type = %T, want *big.Int", i, decoded[i])
						continue
					}
					if actual.Int64() != int64(expected) {
						t.Errorf("decoded[%d] = %d, want %d", i, actual.Int64(), expected)
					}
				case bool:
					actual, ok := decoded[i].(bool)
					if !ok {
						t.Errorf("decoded[%d] type = %T, want bool", i, decoded[i])
						continue
					}
					if actual != expected {
						t.Errorf("decoded[%d] = %v, want %v", i, actual, expected)
					}
				}
			}
		})
	}
}

// ==================== Abi Tests ====================

func TestNewAbi(t *testing.T) {
	entries := []Entry{
		{
			Name:   "test",
			Inputs: []Param{{Name: "x", Type: mustGetType("uint256")}},
			Type:   Function,
		},
	}

	abi := NewAbi(entries)

	if len(abi.Entries) != 1 {
		t.Errorf("len(abi.Entries) = %d, want 1", len(abi.Entries))
	}
	if abi.Entries[0].Name != "test" {
		t.Errorf("abi.Entries[0].Name = %s, want test", abi.Entries[0].Name)
	}
}

func TestFromJson_Valid(t *testing.T) {
	jsonStr := `[
		{
			"name": "transfer",
			"type": "function",
			"inputs": [
				{"name": "to", "type": "address"},
				{"name": "amount", "type": "uint256"}
			]
		},
		{
			"name": "approve",
			"type": "function",
			"inputs": [
				{"name": "spender", "type": "address"},
				{"name": "amount", "type": "uint256"}
			]
		}
	]`

	abi, err := FromJson(jsonStr)
	if err != nil {
		t.Fatalf("FromJson() error = %v", err)
	}

	if len(abi.Entries) != 2 {
		t.Fatalf("len(abi.Entries) = %d, want 2", len(abi.Entries))
	}

	// Check first entry
	if abi.Entries[0].Name != "transfer" {
		t.Errorf("abi.Entries[0].Name = %s, want transfer", abi.Entries[0].Name)
	}
	if len(abi.Entries[0].Inputs) != 2 {
		t.Errorf("len(abi.Entries[0].Inputs) = %d, want 2", len(abi.Entries[0].Inputs))
	}

	// Check second entry
	if abi.Entries[1].Name != "approve" {
		t.Errorf("abi.Entries[1].Name = %s, want approve", abi.Entries[1].Name)
	}
}

func TestFromJson_NoInputs(t *testing.T) {
	jsonStr := `[
		{
			"name": "claim",
			"type": "function"
		}
	]`

	abi, err := FromJson(jsonStr)
	if err != nil {
		t.Fatalf("FromJson() error = %v", err)
	}

	if len(abi.Entries) != 1 {
		t.Fatalf("len(abi.Entries) = %d, want 1", len(abi.Entries))
	}

	if abi.Entries[0].Name != "claim" {
		t.Errorf("abi.Entries[0].Name = %s, want claim", abi.Entries[0].Name)
	}

	if len(abi.Entries[0].Inputs) != 0 {
		t.Errorf("len(abi.Entries[0].Inputs) = %d, want 0", len(abi.Entries[0].Inputs))
	}
}

func TestFromJson_InvalidJSON(t *testing.T) {
	jsonStr := `not valid json`

	_, err := FromJson(jsonStr)
	if err == nil {
		t.Error("FromJson() expected error for invalid JSON, got nil")
	}
}

func TestFromJson_MissingName(t *testing.T) {
	jsonStr := `[
		{
			"type": "function",
			"inputs": []
		}
	]`

	_, err := FromJson(jsonStr)
	if err == nil {
		t.Error("FromJson() expected error for missing name, got nil")
	}
}

func TestFromJson_MissingType(t *testing.T) {
	jsonStr := `[
		{
			"name": "test",
			"inputs": []
		}
	]`

	_, err := FromJson(jsonStr)
	if err == nil {
		t.Error("FromJson() expected error for missing type, got nil")
	}
}

func TestFromJson_InvalidType(t *testing.T) {
	jsonStr := `[
		{
			"name": "test",
			"type": "event",
			"inputs": []
		}
	]`

	_, err := FromJson(jsonStr)
	if err == nil {
		t.Error("FromJson() expected error for non-function type, got nil")
	}
}

func TestFromJson_InvalidParamType(t *testing.T) {
	jsonStr := `[
		{
			"name": "test",
			"type": "function",
			"inputs": [
				{"name": "x", "type": "invalid_type"}
			]
		}
	]`

	_, err := FromJson(jsonStr)
	if err == nil {
		t.Error("FromJson() expected error for invalid param type, got nil")
	}
}

func TestAbi_EncodeFunction(t *testing.T) {
	jsonStr := `[
		{
			"name": "setValue",
			"type": "function",
			"inputs": [
				{"name": "value", "type": "uint256"}
			]
		}
	]`

	abi, err := FromJson(jsonStr)
	if err != nil {
		t.Fatalf("FromJson() error = %v", err)
	}

	args := []interface{}{42}
	encoded, err := abi.EncodeFunction("setValue", args)
	if err != nil {
		t.Fatalf("EncodeFunction() error = %v", err)
	}

	// Should be 4 bytes (signature) + 32 bytes (uint256) = 36 bytes
	expectedLen := 36
	if len(encoded) != expectedLen {
		t.Errorf("len(encoded) = %d, want %d", len(encoded), expectedLen)
	}

	// Verify the value
	val, _ := DecodeInt(encoded, 4)
	if val.Int64() != 42 {
		t.Errorf("encoded value = %d, want 42", val.Int64())
	}
}

func TestAbi_EncodeFunction_UnknownFunction(t *testing.T) {
	jsonStr := `[
		{
			"name": "setValue",
			"type": "function",
			"inputs": []
		}
	]`

	abi, err := FromJson(jsonStr)
	if err != nil {
		t.Fatalf("FromJson() error = %v", err)
	}

	_, err = abi.EncodeFunction("unknownFunction", []interface{}{})
	if err == nil {
		t.Error("EncodeFunction() expected error for unknown function, got nil")
	}
}

func TestAbi_DecodeFunction(t *testing.T) {
	jsonStr := `[
		{
			"name": "setValue",
			"type": "function",
			"inputs": [
				{"name": "value", "type": "uint256"}
			]
		}
	]`

	abi, err := FromJson(jsonStr)
	if err != nil {
		t.Fatalf("FromJson() error = %v", err)
	}

	// Encode first
	args := []interface{}{100}
	encoded, err := abi.EncodeFunction("setValue", args)
	if err != nil {
		t.Fatalf("EncodeFunction() error = %v", err)
	}

	// Decode
	decoded, err := abi.DecodeFunction(encoded)
	if err != nil {
		t.Fatalf("DecodeFunction() error = %v", err)
	}

	if len(decoded) != 1 {
		t.Fatalf("len(decoded) = %d, want 1", len(decoded))
	}

	val, ok := decoded[0].(*big.Int)
	if !ok {
		t.Fatalf("decoded[0] type = %T, want *big.Int", decoded[0])
	}

	if val.Int64() != 100 {
		t.Errorf("decoded[0] = %d, want 100", val.Int64())
	}
}

func TestAbi_DecodeFunction_UnknownSignature(t *testing.T) {
	jsonStr := `[
		{
			"name": "setValue",
			"type": "function",
			"inputs": [
				{"name": "value", "type": "uint256"}
			]
		}
	]`

	abi, err := FromJson(jsonStr)
	if err != nil {
		t.Fatalf("FromJson() error = %v", err)
	}

	// Create encoded data with unknown signature
	fakeEncoded := make([]byte, 36)
	fakeEncoded[0] = 0xFF
	fakeEncoded[1] = 0xFF
	fakeEncoded[2] = 0xFF
	fakeEncoded[3] = 0xFF

	_, err = abi.DecodeFunction(fakeEncoded)
	if err == nil {
		t.Error("DecodeFunction() expected error for unknown signature, got nil")
	}
}

func TestAbi_DecodeFunction_TooShort(t *testing.T) {
	jsonStr := `[
		{
			"name": "test",
			"type": "function",
			"inputs": []
		}
	]`

	abi, err := FromJson(jsonStr)
	if err != nil {
		t.Fatalf("FromJson() error = %v", err)
	}

	// Too short data (less than 4 bytes)
	shortData := []byte{0x01, 0x02}

	_, err = abi.DecodeFunction(shortData)
	if err == nil {
		t.Error("DecodeFunction() expected error for short data, got nil")
	}
}

func TestAbi_RoundTrip(t *testing.T) {
	jsonStr := `[
		{
			"name": "setValues",
			"type": "function",
			"inputs": [
				{"name": "a", "type": "uint256"},
				{"name": "b", "type": "bool"},
				{"name": "c", "type": "address"}
			]
		}
	]`

	abi, err := FromJson(jsonStr)
	if err != nil {
		t.Fatalf("FromJson() error = %v", err)
	}

	// Encode
	args := []interface{}{
		42,
		true,
		"z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7",
	}
	encoded, err := abi.EncodeFunction("setValues", args)
	if err != nil {
		t.Fatalf("EncodeFunction() error = %v", err)
	}

	// Decode
	decoded, err := abi.DecodeFunction(encoded)
	if err != nil {
		t.Fatalf("DecodeFunction() error = %v", err)
	}

	if len(decoded) != 3 {
		t.Fatalf("len(decoded) = %d, want 3", len(decoded))
	}

	// Check values
	val1, ok := decoded[0].(*big.Int)
	if !ok || val1.Int64() != 42 {
		t.Errorf("decoded[0] = %v, want 42", decoded[0])
	}

	val2, ok := decoded[1].(bool)
	if !ok || val2 != true {
		t.Errorf("decoded[1] = %v, want true", decoded[1])
	}

	val3, ok := decoded[2].(types.Address)
	if !ok {
		t.Errorf("decoded[2] type = %T, want types.Address", decoded[2])
	}
	if val3.String() != "z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7" {
		t.Errorf("decoded[2] = %s, want z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7", val3.String())
	}
}

// ==================== Helper Functions ====================

func mustGetType(typeName string) AbiType {
	t, err := GetType(typeName)
	if err != nil {
		panic(err)
	}
	return t
}
