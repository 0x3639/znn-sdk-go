package crypto

import (
	"bytes"
	"testing"
)

// Regression test for finding #14: Digest passed digestSize straight to
// make(), so a negative size panicked with "makeslice: len out of range".
func TestDigestRejectsInvalidSizes(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Digest panicked on invalid digestSize: %v", r)
		}
	}()

	for _, size := range []int{-1, -1 << 40} {
		got := Digest([]byte("test data"), size)
		if got != nil {
			t.Errorf("Digest(data, %d) = %x, want nil", size, got)
		}
	}

	// Valid sizes must keep working.
	if got := Digest([]byte("test data"), 0); len(got) != 32 {
		t.Errorf("Digest(data, 0) length = %d, want 32", len(got))
	}
	if got := Digest([]byte("test data"), 64); len(got) != 64 {
		t.Errorf("Digest(data, 64) length = %d, want 64", len(got))
	}
	if !bytes.Equal(Digest([]byte("test data"), 32), DigestDefault([]byte("test data"))) {
		t.Error("Digest(data, 32) != DigestDefault(data)")
	}
}

// Codex finding: the MaxDigestSize boundary itself was untested.
func TestDigestMaxSizeBoundary(t *testing.T) {
	if got := Digest([]byte("test data"), MaxDigestSize); len(got) != MaxDigestSize {
		t.Errorf("Digest(data, MaxDigestSize) length = %d, want %d", len(got), MaxDigestSize)
	}
	if got := Digest([]byte("test data"), MaxDigestSize+1); got != nil {
		t.Errorf("Digest(data, MaxDigestSize+1) = %d bytes, want nil", len(got))
	}
}
