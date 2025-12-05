package utils

import (
	"encoding/hex"
	"testing"
)

func TestHashDigest(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty bytes",
			input:    []byte{},
			expected: "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a",
		},
		{
			name:     "simple string",
			input:    []byte("test"),
			expected: "36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab80",
		},
		{
			name:     "hello world",
			input:    []byte("hello world"),
			expected: "644bcc7e564373040999aac89e7622f3ca71fba1d972fd94a31c3bfbf24e3938",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashDigest(tt.input)
			resultHex := hex.EncodeToString(result.Bytes())

			if resultHex != tt.expected {
				t.Errorf("HashDigest(%v) = %s, want %s", tt.input, resultHex, tt.expected)
			}
		})
	}
}

func TestHashDigestEmpty(t *testing.T) {
	// SHA3-256 of empty bytes should be consistent
	expected := "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"

	result := HashDigestEmpty()
	resultHex := hex.EncodeToString(result.Bytes())

	if resultHex != expected {
		t.Errorf("HashDigestEmpty() = %s, want %s", resultHex, expected)
	}

	// Should match HashDigest([]byte{})
	direct := HashDigest([]byte{})
	if result != direct {
		t.Error("HashDigestEmpty() should equal HashDigest([]byte{})")
	}
}

func TestHashDigestConsistency(t *testing.T) {
	// Multiple calls with same input should produce same output
	input := []byte("consistent test data")

	result1 := HashDigest(input)
	result2 := HashDigest(input)

	if result1 != result2 {
		t.Error("HashDigest should produce consistent results for same input")
	}
}

func TestHashDigestLength(t *testing.T) {
	// Result should always be 32 bytes
	inputs := [][]byte{
		{},
		{0x00},
		[]byte("short"),
		make([]byte, 1000), // large input
	}

	for _, input := range inputs {
		result := HashDigest(input)
		if len(result.Bytes()) != 32 {
			t.Errorf("HashDigest result length = %d, want 32", len(result.Bytes()))
		}
	}
}
