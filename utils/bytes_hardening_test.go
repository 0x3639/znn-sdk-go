package utils

import (
	"math/big"
	"testing"
)

// Regression test for finding #27: Arraycopy sliced src/dest with unchecked
// indices and lengths, panicking on any out-of-range or negative argument.
func TestArraycopyDoesNotPanicOnBadArguments(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Arraycopy panicked: %v", r)
		}
	}()

	src := []byte{1, 2, 3, 4}
	dest := make([]byte, 4)

	// All of these are out of range and must be safe no-ops.
	Arraycopy(src, 0, dest, 0, 100)   // length beyond src/dest
	Arraycopy(src, -1, dest, 0, 2)    // negative startPos
	Arraycopy(src, 0, dest, -1, 2)    // negative destPos
	Arraycopy(src, 2, dest, 0, 1<<62) // startPos+length overflow
	Arraycopy(src, 0, dest, 0, -1)    // negative length

	// A valid copy must still work.
	Arraycopy(src, 1, dest, 0, 2)
	if dest[0] != 2 || dest[1] != 3 {
		t.Errorf("valid Arraycopy produced %v", dest)
	}
}

// Regression test for finding #28: BigIntToBytes and BigIntToBytesSigned
// passed numBytes straight into make(), panicking on a negative width.
func TestBigIntToBytesRejectsNegativeNumBytes(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("BigInt byte encoder panicked on negative numBytes: %v", r)
		}
	}()

	if _, err := BigIntToBytes(big.NewInt(1), -1); err == nil {
		t.Error("BigIntToBytes(1, -1) did not return an error")
	}
	if _, err := BigIntToBytesSigned(big.NewInt(1), -1); err == nil {
		t.Error("BigIntToBytesSigned(1, -1) did not return an error")
	}
}

// Regression test for review finding #rev7: a huge numBytes must not overflow
// numBytes*8 (wrapping to a small/zero width) and reach an impossible make().
func TestBigIntToBytesRejectsHugeNumBytes(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("BigInt byte encoder panicked on huge numBytes: %v", r)
		}
	}()

	huge := 1 << 61
	if _, err := BigIntToBytes(big.NewInt(0), huge); err == nil {
		t.Error("BigIntToBytes(0, 1<<61) did not return an error")
	}
	if _, err := BigIntToBytesSigned(big.NewInt(0), huge); err == nil {
		t.Error("BigIntToBytesSigned(0, 1<<61) did not return an error")
	}
}

// Regression test for finding #26: BigIntToBytes aliased negative and
// over-wide amounts (BigIntToBytes(-5, 32) == BigIntToBytes(251, 32)), so
// distinct amounts produced identical unsigned encodings.
func TestBigIntToBytesRejectsOutOfDomainAmounts(t *testing.T) {
	if _, err := BigIntToBytes(big.NewInt(-5), 32); err == nil {
		t.Error("BigIntToBytes accepted a negative amount")
	}

	// 2^256 does not fit in 32 unsigned bytes.
	tooWide := new(big.Int).Lsh(big.NewInt(1), 256)
	if _, err := BigIntToBytes(tooWide, 32); err == nil {
		t.Error("BigIntToBytes accepted an over-wide amount")
	}

	// A valid in-range amount encodes correctly and injectively.
	five, err := BigIntToBytes(big.NewInt(5), 32)
	if err != nil {
		t.Fatalf("BigIntToBytes(5, 32) error = %v", err)
	}
	if len(five) != 32 || five[31] != 5 {
		t.Errorf("BigIntToBytes(5, 32) = %x", five)
	}
	twoFiftyOne, err := BigIntToBytes(big.NewInt(251), 32)
	if err != nil {
		t.Fatalf("BigIntToBytes(251, 32) error = %v", err)
	}
	if twoFiftyOne[31] != 251 {
		t.Errorf("BigIntToBytes(251, 32) = %x", twoFiftyOne)
	}
}
