package api

import (
	"errors"
	"testing"
)

// =============================================================================
// isTransientError Tests
// =============================================================================

func TestIsTransientError_TransientErrors(t *testing.T) {
	transientErrors := []error{
		errors.New("connection refused"),
		errors.New("Connection Reset by peer"),
		errors.New("i/o timeout occurred"),
		errors.New("TIMEOUT: request exceeded deadline"),
		errors.New("temporary failure in name resolution"),
		errors.New("network unreachable"),
		errors.New("deadline exceeded"),
		errors.New("broken pipe"),
	}

	for _, err := range transientErrors {
		if !isTransientError(err) {
			t.Errorf("Expected transient error, got permanent for: %v", err)
		}
	}
}

func TestIsTransientError_PermanentErrors(t *testing.T) {
	permanentErrors := []error{
		errors.New("invalid signature provided"),
		errors.New("Insufficient balance for transaction"),
		errors.New("Invalid hash format"),
		errors.New("invalid height specified"),
		errors.New("invalid data in transaction"),
		errors.New("Invalid Amount: negative value"),
		errors.New("invalid address format"),
		errors.New("account chain error"),
	}

	for _, err := range permanentErrors {
		if isTransientError(err) {
			t.Errorf("Expected permanent error, got transient for: %v", err)
		}
	}
}

func TestIsTransientError_NilError(t *testing.T) {
	if isTransientError(nil) {
		t.Error("nil error should not be considered transient")
	}
}

func TestIsTransientError_UnknownError(t *testing.T) {
	// Unknown errors should be treated as transient (safer to retry)
	unknownErr := errors.New("some unknown error message")
	if !isTransientError(unknownErr) {
		t.Error("Unknown errors should be treated as transient")
	}
}

func TestIsTransientError_CaseSensitivity(t *testing.T) {
	// Should be case-insensitive
	testCases := []struct {
		err      error
		expected bool
	}{
		{errors.New("CONNECTION REFUSED"), true},
		{errors.New("Connection Refused"), true},
		{errors.New("connection refused"), true},
		{errors.New("INVALID SIGNATURE"), false},
		{errors.New("Invalid Signature"), false},
		{errors.New("invalid signature"), false},
	}

	for _, tc := range testCases {
		result := isTransientError(tc.err)
		if result != tc.expected {
			t.Errorf("For error %q, expected transient=%v, got %v",
				tc.err, tc.expected, result)
		}
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkIsTransientError_Transient(b *testing.B) {
	err := errors.New("connection refused")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isTransientError(err)
	}
}

func BenchmarkIsTransientError_Permanent(b *testing.B) {
	err := errors.New("invalid signature")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isTransientError(err)
	}
}

func BenchmarkIsTransientError_Unknown(b *testing.B) {
	err := errors.New("unknown error message")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isTransientError(err)
	}
}
