package pow

import (
	"testing"
	"time"
)

// Regression test for finding #18: BenchmarkPoW skipped the difficulty
// validation every other generator applies, so an out-of-range difficulty
// entered an uncancellable nonce-search loop that spins a core until process
// exit. It must reject such difficulty before entering the loop.
func TestBenchmarkPoWRejectsExcessiveDifficulty(t *testing.T) {
	done := make(chan error, 1)
	go func() {
		_, _, err := BenchmarkPoW(^uint64(0))
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("BenchmarkPoW accepted an out-of-range difficulty without error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("BenchmarkPoW never returned for an out-of-range difficulty (uncancellable spin)")
	}
}
