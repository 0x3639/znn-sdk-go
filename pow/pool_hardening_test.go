package pow

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/zenon-network/go-zenon/common/types"
)

// CodeRabbit finding: GeneratePowAsync must not spawn unbounded goroutines.
// With a 1-worker pool the admission capacity is DefaultPendingPoWMultiplier;
// extra requests must be rejected immediately with ErrPoolOverloaded.
func TestGeneratePowAsync_BoundedAdmission(t *testing.T) {
	// SetMaxPoWWorkers latches poolOnce, so restore a default-sized pool
	// afterwards rather than the possibly-nil previous value.
	t.Cleanup(func() { SetMaxPoWWorkers(DefaultMaxPoWWorkers) })
	SetMaxPoWWorkers(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")
	capacity := 1 * DefaultPendingPoWMultiplier
	// Very high difficulty so the active worker keeps running until cancel.
	chans := make([]<-chan PowResult, 0, capacity+2)
	for i := 0; i < capacity; i++ {
		chans = append(chans, GeneratePowAsync(ctx, hash, MaxProtocolDifficulty))
	}
	overloaded := GeneratePowAsync(ctx, hash, 1)
	res := <-overloaded
	if !errors.Is(res.Error, ErrPoolOverloaded) {
		t.Fatalf("expected ErrPoolOverloaded, got %v", res.Error)
	}
	overloadedBig := GeneratePowBigIntAsync(ctx, hash, big.NewInt(1))
	if bigRes := <-overloadedBig; !errors.Is(bigRes.Error, ErrPoolOverloaded) {
		t.Fatalf("expected ErrPoolOverloaded from big-int variant, got %v", bigRes.Error)
	}

	cancel()
	for _, c := range chans {
		<-c
	}
	// After the queue drains, admission works again.
	res = <-GeneratePowAsync(context.Background(), hash, 1)
	if res.Error != nil {
		t.Fatalf("pool did not recover after drain: %v", res.Error)
	}
}

func TestGeneratePowBytesWithContext_ReturnsErrors(t *testing.T) {
	hash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")
	if _, err := GeneratePowBytesWithContext(context.Background(), hash, MaxReasonableDifficulty+1); !errors.Is(err, ErrDifficultyTooHigh) {
		t.Fatalf("expected ErrDifficultyTooHigh, got %v", err)
	}
	nonce, err := GeneratePowBytesWithContext(context.Background(), hash, 1)
	if err != nil || len(nonce) != 8 {
		t.Fatalf("nonce=%x err=%v", nonce, err)
	}
}

// Codex finding: a nil *big.Int difficulty must return an error, never panic,
// through every big-int entry point.
func TestGeneratePowBigInt_NilAndNegativeDifficulty(t *testing.T) {
	hash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")
	if _, err := GeneratePowBigInt(hash, nil); !errors.Is(err, ErrInvalidDifficulty) {
		t.Errorf("GeneratePowBigInt(nil) err = %v", err)
	}
	if _, err := GeneratePowBigIntWithContext(context.Background(), hash, nil); !errors.Is(err, ErrInvalidDifficulty) {
		t.Errorf("GeneratePowBigIntWithContext(nil) err = %v", err)
	}
	if res := <-GeneratePowBigIntAsync(context.Background(), hash, nil); !errors.Is(res.Error, ErrInvalidDifficulty) {
		t.Errorf("GeneratePowBigIntAsync(nil) err = %v", res.Error)
	}
	if _, err := GeneratePowBigInt(hash, big.NewInt(-1)); !errors.Is(err, ErrInvalidDifficulty) {
		t.Errorf("GeneratePowBigInt(-1) err = %v", err)
	}
}
