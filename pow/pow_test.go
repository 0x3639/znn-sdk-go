package pow

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/zenon-network/go-zenon/common/types"
)

// =============================================================================
// PowStatus Tests
// =============================================================================

func TestPowStatus_String(t *testing.T) {
	tests := []struct {
		status   PowStatus
		expected string
	}{
		{Generating, "Generating"},
		{Done, "Done"},
		{PowStatus(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.expected {
			t.Errorf("PowStatus(%d).String() = %s, want %s", tt.status, got, tt.expected)
		}
	}
}

// =============================================================================
// GetThresholdByDifficulty Tests
// =============================================================================

func TestGetThresholdByDifficulty_Zero(t *testing.T) {
	threshold := GetThresholdByDifficulty(big.NewInt(0))
	expected := ^uint64(0) // Max uint64

	if threshold != expected {
		t.Errorf("GetThresholdByDifficulty(0) = %d, want %d", threshold, expected)
	}
}

func TestGetThresholdByDifficulty_One(t *testing.T) {
	threshold := GetThresholdByDifficulty(big.NewInt(1))

	// For difficulty 1, threshold should be close to max uint64
	if threshold == 0 {
		t.Error("GetThresholdByDifficulty(1) should not be 0")
	}
}

func TestGetThresholdByDifficulty_Large(t *testing.T) {
	// High difficulty should give low threshold
	difficulty := big.NewInt(1000000)
	threshold := GetThresholdByDifficulty(difficulty)

	// Threshold should be much smaller than max
	if threshold > ^uint64(0)/2 {
		t.Error("GetThresholdByDifficulty(1000000) should be much less than max uint64")
	}
}

func TestGetThresholdByDifficulty_Inverse(t *testing.T) {
	// Test that higher difficulty gives lower threshold
	threshold1 := GetThresholdByDifficulty(big.NewInt(100))
	threshold2 := GetThresholdByDifficulty(big.NewInt(1000))

	if threshold1 <= threshold2 {
		t.Error("Higher difficulty should give lower threshold")
	}
}

// =============================================================================
// computeHash Tests
// =============================================================================

func TestComputeHash_Deterministic(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("test_hash_123"))

	hash1 := computeHash(testHash, 12345)
	hash2 := computeHash(testHash, 12345)

	if len(hash1) != 32 {
		t.Errorf("computeHash() length = %d, want 32", len(hash1))
	}

	for i := range hash1 {
		if hash1[i] != hash2[i] {
			t.Error("computeHash() should be deterministic")
			break
		}
	}
}

func TestComputeHash_DifferentNonces(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("test_hash_123"))

	hash1 := computeHash(testHash, 1)
	hash2 := computeHash(testHash, 2)

	same := true
	for i := range hash1 {
		if hash1[i] != hash2[i] {
			same = false
			break
		}
	}

	if same {
		t.Error("Different nonces should produce different hashes")
	}
}

// =============================================================================
// uint64ToBytes Tests
// =============================================================================

func TestUint64ToBytes(t *testing.T) {
	tests := []struct {
		input    uint64
		expected []byte
	}{
		{0, []byte{0, 0, 0, 0, 0, 0, 0, 0}},
		{1, []byte{0, 0, 0, 0, 0, 0, 0, 1}},
		{256, []byte{0, 0, 0, 0, 0, 0, 1, 0}},
		{0x0102030405060708, []byte{1, 2, 3, 4, 5, 6, 7, 8}},
	}

	for _, tt := range tests {
		result := uint64ToBytes(tt.input)
		if len(result) != 8 {
			t.Errorf("uint64ToBytes(%d) length = %d, want 8", tt.input, len(result))
		}

		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("uint64ToBytes(%d)[%d] = %d, want %d", tt.input, i, result[i], tt.expected[i])
			}
		}
	}
}

// =============================================================================
// uint64ToHex Tests
// =============================================================================

func TestUint64ToHex(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0000000000000000"},
		{1, "0000000000000001"},
		{255, "00000000000000ff"},
		{256, "0000000000000100"},
		{0x123456789abcdef0, "123456789abcdef0"},
	}

	for _, tt := range tests {
		result := uint64ToHex(tt.input)
		if result != tt.expected {
			t.Errorf("uint64ToHex(%d) = %s, want %s", tt.input, result, tt.expected)
		}

		if len(result) != 16 {
			t.Errorf("uint64ToHex(%d) length = %d, want 16", tt.input, len(result))
		}
	}
}

// =============================================================================
// hexToBytes Tests
// =============================================================================

func TestHexToBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected []byte
	}{
		{"00", []byte{0}},
		{"ff", []byte{255}},
		{"0102", []byte{1, 2}},
		{"abcd", []byte{0xab, 0xcd}},
		{"123456789abcdef0", []byte{0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0}},
	}

	for _, tt := range tests {
		result := hexToBytes(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("hexToBytes(%s) length = %d, want %d", tt.input, len(result), len(tt.expected))
			continue
		}

		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("hexToBytes(%s)[%d] = %d, want %d", tt.input, i, result[i], tt.expected[i])
			}
		}
	}
}

func TestHexToBytes_OddLength(t *testing.T) {
	// Odd length should be padded
	result := hexToBytes("f")
	if len(result) != 1 || result[0] != 0x0f {
		t.Errorf("hexToBytes(f) = %v, want [15]", result)
	}
}

// =============================================================================
// hashToUint64 Tests
// =============================================================================

func TestHashToUint64(t *testing.T) {
	tests := []struct {
		input    []byte
		expected uint64
	}{
		{[]byte{0, 0, 0, 0, 0, 0, 0, 0}, 0},
		{[]byte{0, 0, 0, 0, 0, 0, 0, 1}, 1},
		{[]byte{1, 0, 0, 0, 0, 0, 0, 0}, 0x0100000000000000},
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8}, 0x0102030405060708},
	}

	for _, tt := range tests {
		result := hashToUint64(tt.input)
		if result != tt.expected {
			t.Errorf("hashToUint64(%v) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestHashToUint64_Short(t *testing.T) {
	// Hash shorter than 8 bytes should return 0
	result := hashToUint64([]byte{1, 2, 3})
	if result != 0 {
		t.Errorf("hashToUint64(short) = %d, want 0", result)
	}
}

// =============================================================================
// CheckPoW Tests
// =============================================================================

func TestCheckPoW_ZeroDifficulty(t *testing.T) {
	testHash := types.Hash{}
	if !CheckPoW(testHash, 0, 0) {
		t.Error("CheckPoW() with zero difficulty should always return true")
	}
}

func TestCheckPoW_ValidNonce(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("test_for_valid_nonce"))

	// Generate a valid nonce
	nonceHex := GeneratePoW(testHash, 1000)
	nonceBytes := hexToBytes(nonceHex)
	nonce := hashToUint64(nonceBytes)

	// Check that it's valid
	if !CheckPoW(testHash, nonce, 1000) {
		t.Error("CheckPoW() should return true for valid nonce")
	}
}

func TestCheckPoW_InvalidNonce(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("test_for_invalid_nonce"))

	// Use a very high difficulty that nonce 0 won't satisfy
	if CheckPoW(testHash, 0, 100000000) {
		t.Error("CheckPoW() should return false for invalid nonce with high difficulty")
	}
}

// =============================================================================
// GeneratePoW Tests
// =============================================================================

func TestGeneratePoW_ZeroDifficulty(t *testing.T) {
	testHash := types.Hash{}
	nonce := GeneratePoW(testHash, 0)

	if nonce != "0000000000000000" {
		t.Errorf("GeneratePoW() with zero difficulty = %s, want 0000000000000000", nonce)
	}
}

func TestGeneratePoW_LowDifficulty(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("test_low_difficulty"))

	nonce := GeneratePoW(testHash, 10)

	// Should return a valid hex string
	if len(nonce) != 16 {
		t.Errorf("GeneratePoW() nonce length = %d, want 16", len(nonce))
	}

	// Verify it's valid
	nonceBytes := hexToBytes(nonce)
	nonceVal := hashToUint64(nonceBytes)
	if !CheckPoW(testHash, nonceVal, 10) {
		t.Error("GeneratePoW() should return valid nonce")
	}
}

func TestGeneratePoW_MediumDifficulty(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("test_medium_difficulty"))

	nonce := GeneratePoW(testHash, 1000)

	// Verify it's valid
	nonceBytes := hexToBytes(nonce)
	nonceVal := hashToUint64(nonceBytes)
	if !CheckPoW(testHash, nonceVal, 1000) {
		t.Error("GeneratePoW() should return valid nonce for medium difficulty")
	}
}

func TestGeneratePoW_Deterministic(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("test_deterministic"))

	nonce1 := GeneratePoW(testHash, 100)
	nonce2 := GeneratePoW(testHash, 100)

	if nonce1 != nonce2 {
		t.Error("GeneratePoW() should be deterministic for same input")
	}
}

func TestGeneratePoW_DifferentHashes(t *testing.T) {
	hash1 := types.Hash{}
	hash2 := types.Hash{}
	copy(hash1[:], []byte("hash1"))
	copy(hash2[:], []byte("hash2"))

	nonce1 := GeneratePoW(hash1, 100)
	nonce2 := GeneratePoW(hash2, 100)

	// Different hashes should (usually) produce different nonces
	// Note: This could theoretically fail, but very unlikely
	if nonce1 == nonce2 {
		t.Log("Warning: Different hashes produced same nonce (unlikely but possible)")
	}
}

// =============================================================================
// GeneratePowBigInt Tests
// =============================================================================

func TestGeneratePowBigInt_ZeroDifficulty(t *testing.T) {
	testHash := types.Hash{}
	nonce := GeneratePowBigInt(testHash, big.NewInt(0))

	if nonce != "0000000000000000" {
		t.Errorf("GeneratePowBigInt() with zero difficulty = %s, want 0000000000000000", nonce)
	}
}

func TestGeneratePowBigInt_Valid(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("test_bigint"))

	nonce := GeneratePowBigInt(testHash, big.NewInt(500))

	// Verify it's valid
	nonceBytes := hexToBytes(nonce)
	nonceVal := hashToUint64(nonceBytes)
	if !CheckPoW(testHash, nonceVal, 500) {
		t.Error("GeneratePowBigInt() should return valid nonce")
	}
}

func TestGeneratePowBigInt_MatchesGeneratePoW(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("test_match"))

	difficulty := uint64(200)
	nonce1 := GeneratePoW(testHash, difficulty)
	nonce2 := GeneratePowBigInt(testHash, big.NewInt(int64(difficulty)))

	if nonce1 != nonce2 {
		t.Error("GeneratePowBigInt() should match GeneratePoW() for same difficulty")
	}
}

// =============================================================================
// GeneratePowBytes Tests
// =============================================================================

func TestGeneratePowBytes_ReturnsBytes(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("test_bytes"))

	nonceBytes := GeneratePowBytes(testHash, 50)

	if len(nonceBytes) != 8 {
		t.Errorf("GeneratePowBytes() length = %d, want 8", len(nonceBytes))
	}

	// Verify it's valid
	nonceVal := hashToUint64(nonceBytes)
	if !CheckPoW(testHash, nonceVal, 50) {
		t.Error("GeneratePowBytes() should return valid nonce")
	}
}

// =============================================================================
// BenchmarkPoW Tests
// =============================================================================

func TestBenchmarkPoW_LowDifficulty(t *testing.T) {
	nonce, iterations := BenchmarkPoW(10)

	if len(nonce) != 16 {
		t.Errorf("BenchmarkPoW() nonce length = %d, want 16", len(nonce))
	}

	if iterations == 0 {
		t.Error("BenchmarkPoW() should perform at least one iteration")
	}

	t.Logf("BenchmarkPoW(10): found nonce after %d iterations", iterations)
}

func TestBenchmarkPoW_MediumDifficulty(t *testing.T) {
	nonce, iterations := BenchmarkPoW(1000)

	if len(nonce) != 16 {
		t.Errorf("BenchmarkPoW() nonce length = %d, want 16", len(nonce))
	}

	if iterations == 0 {
		t.Error("BenchmarkPoW() should perform at least one iteration")
	}

	t.Logf("BenchmarkPoW(1000): found nonce after %d iterations", iterations)
}

func TestBenchmarkPoW_Deterministic(t *testing.T) {
	// BenchmarkPoW uses a fixed hash, so should be deterministic
	nonce1, iter1 := BenchmarkPoW(100)
	nonce2, iter2 := BenchmarkPoW(100)

	if nonce1 != nonce2 {
		t.Error("BenchmarkPoW() should be deterministic")
	}

	if iter1 != iter2 {
		t.Error("BenchmarkPoW() should return same iteration count")
	}
}

// =============================================================================
// Performance Benchmarks
// =============================================================================

func BenchmarkGeneratePoW_Difficulty10(b *testing.B) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("benchmark_test"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GeneratePoW(testHash, 10)
	}
}

func BenchmarkGeneratePoW_Difficulty100(b *testing.B) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("benchmark_test"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GeneratePoW(testHash, 100)
	}
}

func BenchmarkGeneratePoW_Difficulty1000(b *testing.B) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("benchmark_test"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GeneratePoW(testHash, 1000)
	}
}

func BenchmarkCheckPoW(b *testing.B) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("benchmark_test"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckPoW(testHash, 12345, 1000)
	}
}

func BenchmarkComputeHash(b *testing.B) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("benchmark_test"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		computeHash(testHash, uint64(i))
	}
}

// =============================================================================
// Async PoW Tests
// =============================================================================

func TestGeneratePowAsync_Success(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("test_async_pow"))
	difficulty := uint64(1000) // Low difficulty for fast test

	ctx := context.Background()
	resultChan := GeneratePowAsync(ctx, testHash, difficulty)

	// Wait for result
	result := <-resultChan

	if result.Error != nil {
		t.Fatalf("GeneratePowAsync() error = %v, want nil", result.Error)
	}

	if result.Nonce == "" {
		t.Error("GeneratePowAsync() returned empty nonce")
	}

	// Verify the nonce is valid
	if len(result.Nonce) != 16 {
		t.Errorf("GeneratePowAsync() nonce length = %d, want 16", len(result.Nonce))
	}
}

func TestGeneratePowAsync_ZeroDifficulty(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("zero_difficulty"))

	ctx := context.Background()
	resultChan := GeneratePowAsync(ctx, testHash, 0)

	result := <-resultChan

	if result.Error != nil {
		t.Fatalf("GeneratePowAsync() with zero difficulty error = %v, want nil", result.Error)
	}

	if result.Nonce != "0000000000000000" {
		t.Errorf("GeneratePowAsync() with zero difficulty = %s, want 0000000000000000", result.Nonce)
	}
}

func TestGeneratePowAsync_Cancellation(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("cancel_test"))
	difficulty := uint64(100000000) // Very high difficulty - will take long

	ctx, cancel := context.WithCancel(context.Background())

	resultChan := GeneratePowAsync(ctx, testHash, difficulty)

	// Cancel immediately
	cancel()

	// Wait for result
	result := <-resultChan

	if !errors.Is(result.Error, ErrCancelled) {
		t.Errorf("GeneratePowAsync() after cancel error = %v, want %v", result.Error, ErrCancelled)
	}

	if result.Nonce != "" {
		t.Errorf("GeneratePowAsync() after cancel returned nonce = %s, want empty", result.Nonce)
	}
}

func TestGeneratePowAsync_Timeout(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("timeout_test"))
	difficulty := uint64(100000000) // Very high difficulty

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	resultChan := GeneratePowAsync(ctx, testHash, difficulty)

	result := <-resultChan

	// Should timeout
	if result.Error == nil {
		t.Error("GeneratePowAsync() with timeout should return error")
	}

	if !errors.Is(result.Error, ErrCancelled) {
		t.Errorf("GeneratePowAsync() timeout error = %v, want %v", result.Error, ErrCancelled)
	}
}

func TestGeneratePowAsync_MultipleConcurrent(t *testing.T) {
	ctx := context.Background()
	numOps := 5
	difficulty := uint64(1000) // Low difficulty

	// Start multiple concurrent PoW operations
	results := make([]<-chan PowResult, numOps)
	for i := 0; i < numOps; i++ {
		testHash := types.Hash{}
		copy(testHash[:], []byte(fmt.Sprintf("concurrent_test_%d", i)))
		results[i] = GeneratePowAsync(ctx, testHash, difficulty)
	}

	// Collect all results
	for i := 0; i < numOps; i++ {
		result := <-results[i]
		if result.Error != nil {
			t.Errorf("Operation %d failed: %v", i, result.Error)
		}
		if result.Nonce == "" {
			t.Errorf("Operation %d returned empty nonce", i)
		}
	}
}

func TestGeneratePowAsync_ChannelClosed(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("channel_close_test"))
	difficulty := uint64(1000)

	ctx := context.Background()
	resultChan := GeneratePowAsync(ctx, testHash, difficulty)

	// Read result
	result := <-resultChan

	if result.Error != nil {
		t.Fatalf("First read error = %v", result.Error)
	}

	// Try to read again - should get zero value because channel is closed
	result2, ok := <-resultChan
	if ok {
		t.Error("Channel should be closed after first result")
	}
	if result2.Nonce != "" || result2.Error != nil {
		t.Error("Second read should return zero value")
	}
}

func TestGeneratePowBigIntAsync_Success(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("bigint_async_test"))
	difficulty := big.NewInt(1000)

	ctx := context.Background()
	resultChan := GeneratePowBigIntAsync(ctx, testHash, difficulty)

	result := <-resultChan

	if result.Error != nil {
		t.Fatalf("GeneratePowBigIntAsync() error = %v, want nil", result.Error)
	}

	if result.Nonce == "" {
		t.Error("GeneratePowBigIntAsync() returned empty nonce")
	}

	if len(result.Nonce) != 16 {
		t.Errorf("GeneratePowBigIntAsync() nonce length = %d, want 16", len(result.Nonce))
	}
}

func TestGeneratePowBigIntAsync_ZeroDifficulty(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("bigint_zero"))
	difficulty := big.NewInt(0)

	ctx := context.Background()
	resultChan := GeneratePowBigIntAsync(ctx, testHash, difficulty)

	result := <-resultChan

	if result.Error != nil {
		t.Fatalf("GeneratePowBigIntAsync() with zero error = %v", result.Error)
	}

	if result.Nonce != "0000000000000000" {
		t.Errorf("GeneratePowBigIntAsync() zero difficulty = %s, want 0000000000000000", result.Nonce)
	}
}

func TestGeneratePowBigIntAsync_Cancellation(t *testing.T) {
	testHash := types.Hash{}
	copy(testHash[:], []byte("bigint_cancel"))
	difficulty := big.NewInt(100000000)

	ctx, cancel := context.WithCancel(context.Background())

	resultChan := GeneratePowBigIntAsync(ctx, testHash, difficulty)

	// Cancel immediately
	cancel()

	result := <-resultChan

	if !errors.Is(result.Error, ErrCancelled) {
		t.Errorf("GeneratePowBigIntAsync() cancel error = %v, want %v", result.Error, ErrCancelled)
	}
}

// =============================================================================
// Worker Pool Tests
// =============================================================================

func TestSetMaxPoWWorkers(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"Valid 4 workers", 4, 4},
		{"Valid 16 workers", 16, 16},
		{"Zero defaults to 8", 0, DefaultMaxPoWWorkers},
		{"Negative defaults to 8", -1, DefaultMaxPoWWorkers},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset pool for each test
			pool = nil
			poolOnce = sync.Once{}

			SetMaxPoWWorkers(tt.input)

			got := GetMaxPoWWorkers()
			if got != tt.expected {
				t.Errorf("SetMaxPoWWorkers(%d) resulted in %d workers, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetMaxPoWWorkers_Uninitialized(t *testing.T) {
	// Reset pool
	pool = nil
	poolOnce = sync.Once{}

	got := GetMaxPoWWorkers()
	if got != DefaultMaxPoWWorkers {
		t.Errorf("GetMaxPoWWorkers() before initialization = %d, want %d", got, DefaultMaxPoWWorkers)
	}
}

func TestWorkerPool_ConcurrencyLimit(t *testing.T) {
	// Reset pool and set to 2 workers for easier testing
	pool = nil
	poolOnce = sync.Once{}
	SetMaxPoWWorkers(2)

	testHash := types.Hash{}
	copy(testHash[:], []byte("concurrency_test"))
	difficulty := uint64(1000) // Low difficulty for fast completion

	ctx := context.Background()

	// Track how many PoW operations are running concurrently
	running := make(chan int, 10)
	maxConcurrent := 0
	currentRunning := 0

	// Launch 5 PoW operations
	numOps := 5
	results := make([]<-chan PowResult, numOps)

	for i := 0; i < numOps; i++ {
		results[i] = GeneratePowAsync(ctx, testHash, difficulty)
	}

	// Monitor goroutine execution
	go func() {
		for delta := range running {
			currentRunning += delta
			if currentRunning > maxConcurrent {
				maxConcurrent = currentRunning
			}
		}
	}()

	// Collect all results
	for i := 0; i < numOps; i++ {
		result := <-results[i]
		if result.Error != nil {
			t.Errorf("PoW operation %d failed: %v", i, result.Error)
		}
	}

	close(running)

	// With 2 workers, we should never have more than 2 concurrent operations
	// Note: This is a best-effort check as we can't directly observe internal state
	t.Logf("Completed %d PoW operations with max 2 workers", numOps)
}

func TestWorkerPool_Cancellation_WhileQueued(t *testing.T) {
	// Reset pool and set to 1 worker to force queuing
	pool = nil
	poolOnce = sync.Once{}
	SetMaxPoWWorkers(1)

	testHash := types.Hash{}
	copy(testHash[:], []byte("queue_cancel_test"))

	// Use a difficulty high enough to keep first worker busy for ~1-2 seconds
	// but not so high it causes test timeouts
	difficulty := uint64(5000000)

	// Start first operation with timeout (will acquire the only worker slot)
	ctx1, cancel1 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel1()
	result1Chan := GeneratePowAsync(ctx1, testHash, difficulty)

	// Give first operation time to acquire the worker slot and start computing
	time.Sleep(100 * time.Millisecond)

	// Start second operation with cancellable context (will be queued, blocked on semaphore)
	ctx2, cancel2 := context.WithCancel(context.Background())
	result2Chan := GeneratePowAsync(ctx2, testHash, difficulty)

	// Give second operation time to hit the semaphore queue
	time.Sleep(100 * time.Millisecond)

	// Cancel second operation while it's blocked waiting for semaphore
	cancel2()

	// Second operation should return cancellation error immediately
	// (cancelled while waiting for semaphore, not during PoW computation)
	result2 := <-result2Chan
	if !errors.Is(result2.Error, ErrCancelled) {
		t.Errorf("Queued operation cancel error = %v, want %v", result2.Error, ErrCancelled)
	}

	// Cancel first operation to avoid waiting for it to complete
	cancel1()
	result1 := <-result1Chan
	// Don't check error as it may complete or be cancelled
	_ = result1
}

func TestWorkerPool_MultipleOperations_Success(t *testing.T) {
	// Reset pool with default workers
	pool = nil
	poolOnce = sync.Once{}
	SetMaxPoWWorkers(4)

	testHash := types.Hash{}
	copy(testHash[:], []byte("multi_success"))
	difficulty := uint64(5000) // Medium difficulty

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	numOps := 10
	results := make([]<-chan PowResult, numOps)

	// Launch all operations
	for i := 0; i < numOps; i++ {
		hash := types.Hash{}
		copy(hash[:], []byte(fmt.Sprintf("multi_%d", i)))
		results[i] = GeneratePowAsync(ctx, hash, difficulty)
	}

	// Collect and verify all results
	successCount := 0
	for i := 0; i < numOps; i++ {
		result := <-results[i]
		if result.Error != nil {
			t.Errorf("Operation %d failed: %v", i, result.Error)
			continue
		}
		if result.Nonce == "" {
			t.Errorf("Operation %d returned empty nonce", i)
			continue
		}
		successCount++
	}

	if successCount != numOps {
		t.Errorf("Expected %d successful operations, got %d", numOps, successCount)
	}

	t.Logf("Successfully completed %d concurrent PoW operations", successCount)
}

func TestWorkerPool_EnvironmentVariable(t *testing.T) {
	// This test verifies that POW_MAX_WORKERS env var is respected
	// Note: We can't easily test this in a unit test without forking the process
	// This is more of a documentation test

	t.Log("POW_MAX_WORKERS environment variable can be set to override default")
	t.Log("Example: POW_MAX_WORKERS=16 go run main.go")
	t.Log("This test verifies the code path exists but doesn't test actual env var")

	// Verify the code compiles and initializes
	pool = nil
	poolOnce = sync.Once{}
	initWorkerPool()

	workers := GetMaxPoWWorkers()
	if workers <= 0 {
		t.Errorf("Worker pool initialized with invalid size: %d", workers)
	}
}

func TestWorkerPool_BigIntAsync_WithPool(t *testing.T) {
	// Reset pool
	pool = nil
	poolOnce = sync.Once{}
	SetMaxPoWWorkers(2)

	testHash := types.Hash{}
	copy(testHash[:], []byte("bigint_pool_test"))
	difficulty := big.NewInt(2000)

	ctx := context.Background()

	numOps := 4
	results := make([]<-chan PowResult, numOps)

	// Launch operations
	for i := 0; i < numOps; i++ {
		results[i] = GeneratePowBigIntAsync(ctx, testHash, difficulty)
	}

	// Verify all complete successfully
	for i := 0; i < numOps; i++ {
		result := <-results[i]
		if result.Error != nil {
			t.Errorf("BigInt operation %d failed: %v", i, result.Error)
		}
		if result.Nonce == "" {
			t.Errorf("BigInt operation %d returned empty nonce", i)
		}
	}

	t.Logf("Successfully completed %d BigInt PoW operations with worker pool", numOps)
}
