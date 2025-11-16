package pow

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/zenon-network/go-zenon/common/types"
)

// =============================================================================
// Difficulty Validation Tests
// =============================================================================

func TestValidateAndCapDifficulty_WithinRange(t *testing.T) {
	testCases := []uint64{
		0,
		1,
		1000,
		1000000,
		31500000,  // Base transaction
		78750000,  // Embedded simple
		141750000, // Max protocol
		MaxProtocolDifficulty,
	}

	for _, difficulty := range testCases {
		capped, err := validateAndCapDifficulty(difficulty)
		if err != nil {
			t.Errorf("validateAndCapDifficulty(%d) returned error: %v", difficulty, err)
		}
		if capped != difficulty {
			t.Errorf("validateAndCapDifficulty(%d) = %d, want %d (no capping expected)",
				difficulty, capped, difficulty)
		}
	}
}

func TestValidateAndCapDifficulty_AboveProtocol(t *testing.T) {
	testCases := []uint64{
		MaxProtocolDifficulty + 1,
		MaxProtocolDifficulty + 1000,
		150000000,
		MaxReasonableDifficulty - 1,
		MaxReasonableDifficulty,
	}

	for _, difficulty := range testCases {
		capped, err := validateAndCapDifficulty(difficulty)
		if err != nil {
			t.Errorf("validateAndCapDifficulty(%d) returned error: %v (should cap, not error)",
				difficulty, err)
		}
		if capped != MaxProtocolDifficulty {
			t.Errorf("validateAndCapDifficulty(%d) = %d, want %d (should be capped)",
				difficulty, capped, MaxProtocolDifficulty)
		}
	}
}

func TestValidateAndCapDifficulty_TooHigh(t *testing.T) {
	testCases := []uint64{
		MaxReasonableDifficulty + 1,
		MaxReasonableDifficulty + 1000000,
		1000000000,
		^uint64(0), // max uint64
	}

	for _, difficulty := range testCases {
		_, err := validateAndCapDifficulty(difficulty)
		if err == nil {
			t.Errorf("validateAndCapDifficulty(%d) should return error", difficulty)
		}
		if err != nil && !strings.Contains(err.Error(), "difficulty exceeds reasonable maximum") {
			t.Errorf("validateAndCapDifficulty(%d) returned wrong error: %v", difficulty, err)
		}
	}
}

func TestValidateAndCapDifficultyBigInt_WithinRange(t *testing.T) {
	testCases := []*big.Int{
		big.NewInt(0),
		big.NewInt(1),
		big.NewInt(1000000),
		big.NewInt(141750000), // Max protocol
	}

	for _, difficulty := range testCases {
		capped, err := validateAndCapDifficultyBigInt(difficulty)
		if err != nil {
			t.Errorf("validateAndCapDifficultyBigInt(%s) returned error: %v", difficulty, err)
		}
		if capped.Cmp(difficulty) != 0 {
			t.Errorf("validateAndCapDifficultyBigInt(%s) = %s, want %s",
				difficulty, capped, difficulty)
		}
	}
}

func TestValidateAndCapDifficultyBigInt_TooLarge(t *testing.T) {
	// Create a big.Int that's larger than uint64 max
	tooLarge := new(big.Int).SetUint64(^uint64(0))
	tooLarge.Add(tooLarge, big.NewInt(1))

	_, err := validateAndCapDifficultyBigInt(tooLarge)
	if err == nil {
		t.Error("validateAndCapDifficultyBigInt(>uint64) should return error")
	}
}

// =============================================================================
// GeneratePoW with Difficulty Cap Tests
// =============================================================================

func TestGeneratePoW_WithCappedDifficulty(t *testing.T) {
	hash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")

	// Use a difficulty slightly above protocol max (should be capped)
	difficulty := MaxProtocolDifficulty + 10000

	// Should cap to MaxProtocolDifficulty and generate successfully
	nonce := GeneratePoW(hash, difficulty)
	if nonce == "" {
		t.Error("GeneratePoW should return a nonce")
	}
}

func TestGeneratePoW_PanicsOnTooHighDifficulty(t *testing.T) {
	hash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")

	defer func() {
		if r := recover(); r == nil {
			t.Error("GeneratePoW should panic on difficulty > MaxReasonableDifficulty")
		}
	}()

	// This should panic
	_ = GeneratePoW(hash, MaxReasonableDifficulty+1)
}

func TestGeneratePowWithContext_RejectsHighDifficulty(t *testing.T) {
	ctx := context.Background()
	hash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")

	// Should return error, not panic
	_, err := GeneratePowWithContext(ctx, hash, MaxReasonableDifficulty+1)
	if err == nil {
		t.Error("GeneratePowWithContext should return error for high difficulty")
	}
	if !errors.Is(err, ErrDifficultyTooHigh) && !strings.Contains(err.Error(), "exceeds reasonable maximum") {
		t.Errorf("Expected ErrDifficultyTooHigh, got: %v", err)
	}
}

func TestGeneratePowWithContext_CapsAboveProtocol(t *testing.T) {
	ctx := context.Background()
	hash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")

	// Should cap to MaxProtocolDifficulty and succeed
	difficulty := MaxProtocolDifficulty + 10000
	nonce, err := GeneratePowWithContext(ctx, hash, difficulty)
	if err != nil {
		t.Errorf("GeneratePowWithContext should cap difficulty and succeed: %v", err)
	}
	if nonce == "" {
		t.Error("GeneratePowWithContext should return a nonce")
	}
}

func TestGeneratePowBigInt_WithValidation(t *testing.T) {
	hash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")

	// Within range - should work
	difficulty := big.NewInt(1000000)
	nonce := GeneratePowBigInt(hash, difficulty)
	if nonce == "" {
		t.Error("GeneratePowBigInt should return a nonce")
	}
}

func TestGeneratePowBigInt_PanicsOnTooHigh(t *testing.T) {
	hash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")

	defer func() {
		if r := recover(); r == nil {
			t.Error("GeneratePowBigInt should panic on too high difficulty")
		}
	}()

	// This should panic
	difficulty := big.NewInt(0).SetUint64(MaxReasonableDifficulty + 1)
	_ = GeneratePowBigInt(hash, difficulty)
}

func TestGeneratePowBigIntWithContext_RejectsHighDifficulty(t *testing.T) {
	ctx := context.Background()
	hash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")

	difficulty := big.NewInt(0).SetUint64(MaxReasonableDifficulty + 1)
	_, err := GeneratePowBigIntWithContext(ctx, hash, difficulty)
	if err == nil {
		t.Error("GeneratePowBigIntWithContext should return error for high difficulty")
	}
}

// =============================================================================
// Constant Validation Tests
// =============================================================================

func TestConstants_Values(t *testing.T) {
	// Verify the constants have the expected values
	if MaxProtocolDifficulty != 141_750_000 {
		t.Errorf("MaxProtocolDifficulty = %d, want 141750000", MaxProtocolDifficulty)
	}

	if MaxReasonableDifficulty != 200_000_000 {
		t.Errorf("MaxReasonableDifficulty = %d, want 200000000", MaxReasonableDifficulty)
	}

	// MaxReasonableDifficulty should be > MaxProtocolDifficulty
	if MaxReasonableDifficulty <= MaxProtocolDifficulty {
		t.Error("MaxReasonableDifficulty should be > MaxProtocolDifficulty")
	}
}

func TestConstants_Relationship(t *testing.T) {
	// Verify protocol max is based on correct formula
	// 94,500 plasma × 1,500 = 141,750,000
	maxPlasma := uint64(94_500)
	difficultyPerPlasma := uint64(1_500)
	expectedMax := maxPlasma * difficultyPerPlasma

	if MaxProtocolDifficulty != expectedMax {
		t.Errorf("MaxProtocolDifficulty = %d, expected %d (from formula)",
			MaxProtocolDifficulty, expectedMax)
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkValidateAndCapDifficulty_WithinRange(b *testing.B) {
	difficulty := uint64(100000000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		validateAndCapDifficulty(difficulty)
	}
}

func BenchmarkValidateAndCapDifficulty_Capped(b *testing.B) {
	difficulty := MaxProtocolDifficulty + 10000
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		validateAndCapDifficulty(difficulty)
	}
}
