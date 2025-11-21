package pow

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"sync"

	"github.com/zenon-network/go-zenon/common/types"
	"golang.org/x/crypto/sha3"
)

const (
	// MaxProtocolDifficulty is the absolute maximum difficulty the Zenon protocol can require.
	// This is calculated as: MaxPoWPlasmaForAccountBlock × PoWDifficultyPerPlasma
	// = 94,500 plasma × 1,500 = 141,750,000
	//
	// Any difficulty value above this is either:
	// - A malfunctioning node
	// - A malicious attempt to DoS the client
	// - An incompatible protocol version
	MaxProtocolDifficulty uint64 = 141_750_000

	// MaxReasonableDifficulty is a safety cap with a 50% buffer above the protocol maximum.
	// Difficulties above this threshold will be rejected as obvious attacks or errors.
	MaxReasonableDifficulty uint64 = 200_000_000

	// DefaultMaxPoWWorkers is the default maximum number of concurrent PoW computations.
	// This value of 8 matches the Dart SDK's parallelism level and prevents CPU
	// exhaustion when multiple transactions are submitted concurrently.
	//
	// This can be overridden via:
	// - SetMaxPoWWorkers() function
	// - POW_MAX_WORKERS environment variable
	DefaultMaxPoWWorkers = 8
)

var (
	// ErrCancelled is returned when PoW generation is cancelled via context
	ErrCancelled = errors.New("pow generation cancelled")

	// ErrDifficultyTooHigh is returned when difficulty exceeds the reasonable maximum
	ErrDifficultyTooHigh = errors.New("difficulty exceeds reasonable maximum (possible DoS attack)")
)

// workerPool manages concurrent PoW generation operations.
// It uses a semaphore pattern to limit the number of simultaneous PoW computations,
// preventing CPU exhaustion when multiple transactions are submitted concurrently.
type workerPool struct {
	semaphore chan struct{}
}

var (
	pool     *workerPool
	poolOnce sync.Once
)

// initWorkerPool initializes the global worker pool.
// This is called lazily on first use of GeneratePowAsync or GeneratePowBigIntAsync.
func initWorkerPool() {
	poolOnce.Do(func() {
		maxWorkers := DefaultMaxPoWWorkers

		// Check for environment variable override
		if envVal := os.Getenv("POW_MAX_WORKERS"); envVal != "" {
			if val, err := strconv.Atoi(envVal); err == nil && val > 0 {
				maxWorkers = val
			}
		}

		pool = &workerPool{
			semaphore: make(chan struct{}, maxWorkers),
		}
	})
}

// acquire blocks until a worker slot is available or context is cancelled.
// Returns an error if the context is cancelled while waiting.
func (p *workerPool) acquire(ctx context.Context) error {
	select {
	case p.semaphore <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ErrCancelled
	}
}

// release frees a worker slot, allowing another PoW computation to start.
func (p *workerPool) release() {
	<-p.semaphore
}

// SetMaxPoWWorkers configures the maximum number of concurrent PoW computations.
// This must be called before any PoW generation operations to take effect.
//
// Parameters:
//   - maxWorkers: Maximum number of concurrent PoW computations (must be > 0)
//
// Setting this value affects all subsequent calls to GeneratePowAsync and
// GeneratePowBigIntAsync. It does not affect already-running PoW operations.
//
// Example:
//
//	// Limit to 4 concurrent PoW operations (useful on low-end hardware)
//	pow.SetMaxPoWWorkers(4)
//
//	// Allow 16 concurrent operations (for high-performance servers)
//	pow.SetMaxPoWWorkers(16)
//
// Note: This function is NOT thread-safe and should only be called during
// application initialization, before any PoW generation begins.
func SetMaxPoWWorkers(maxWorkers int) {
	if maxWorkers <= 0 {
		maxWorkers = DefaultMaxPoWWorkers
	}

	poolOnce.Do(func() {}) // Ensure poolOnce is marked as initialized

	pool = &workerPool{
		semaphore: make(chan struct{}, maxWorkers),
	}
}

// GetMaxPoWWorkers returns the current maximum number of concurrent PoW workers.
// If the worker pool hasn't been initialized yet, returns the default value.
//
// Returns:
//   - The current worker pool capacity
//
// Example:
//
//	fmt.Printf("Max concurrent PoW workers: %d\n", pow.GetMaxPoWWorkers())
func GetMaxPoWWorkers() int {
	if pool == nil {
		return DefaultMaxPoWWorkers
	}
	return cap(pool.semaphore)
}

// PowResult contains the result of an asynchronous PoW generation
type PowResult struct {
	// Nonce is the generated nonce as a hex string (without 0x prefix)
	Nonce string
	// Error is set if PoW generation failed or was cancelled
	Error error
}

// PowStatus represents the status of PoW generation
type PowStatus int

const (
	// Generating indicates PoW is being generated
	Generating PowStatus = iota
	// Done indicates PoW generation is complete
	Done
)

// String returns the string representation of PowStatus
func (s PowStatus) String() string {
	switch s {
	case Generating:
		return "Generating"
	case Done:
		return "Done"
	default:
		return "Unknown"
	}
}

// validateAndCapDifficulty validates the difficulty and caps it if necessary.
//
// Returns:
//   - The capped difficulty value (safe to use)
//   - An error if difficulty exceeds MaxReasonableDifficulty
//
// Behavior:
//   - difficulty <= MaxProtocolDifficulty: Returns as-is, no warning
//   - MaxProtocolDifficulty < difficulty <= MaxReasonableDifficulty: Caps to MaxProtocolDifficulty, logs warning
//   - difficulty > MaxReasonableDifficulty: Returns error (obvious attack)
func validateAndCapDifficulty(difficulty uint64) (uint64, error) {
	// Check if obviously too high (probable DoS attack)
	if difficulty > MaxReasonableDifficulty {
		return 0, fmt.Errorf("%w: difficulty=%d, max=%d",
			ErrDifficultyTooHigh, difficulty, MaxReasonableDifficulty)
	}

	// Check if above protocol maximum (cap it and warn)
	if difficulty > MaxProtocolDifficulty {
		log.Printf("WARNING: Difficulty %d exceeds protocol maximum %d. "+
			"Capping to protocol maximum. This may indicate a malfunctioning or malicious node.",
			difficulty, MaxProtocolDifficulty)
		return MaxProtocolDifficulty, nil
	}

	// Within normal range
	return difficulty, nil
}

// validateAndCapDifficultyBigInt is like validateAndCapDifficulty but for *big.Int
func validateAndCapDifficultyBigInt(difficulty *big.Int) (*big.Int, error) {
	// Check if difficulty fits in uint64
	if !difficulty.IsUint64() {
		return nil, fmt.Errorf("%w: difficulty too large for uint64",
			ErrDifficultyTooHigh)
	}

	// Validate as uint64
	diffUint64 := difficulty.Uint64()
	cappedUint64, err := validateAndCapDifficulty(diffUint64)
	if err != nil {
		return nil, err
	}

	return new(big.Int).SetUint64(cappedUint64), nil
}

// GeneratePoW generates a valid proof-of-work nonce for the given hash and difficulty
// Returns the nonce as a hex string (without 0x prefix)
//
// The algorithm:
// 1. Iterate through nonce values starting from 0
// 2. For each nonce, compute: hash = SHA3-256(dataHash + nonce)
// 3. Interpret hash as big-endian uint256
// 4. Check if: hash * difficulty < 2^256
// 5. Return nonce when condition is met
//
// Note: This function panics if difficulty exceeds MaxReasonableDifficulty.
// For error handling, use GeneratePowWithContext instead.
func GeneratePoW(dataHash types.Hash, difficulty uint64) string {
	if difficulty == 0 {
		return "0000000000000000"
	}

	// Validate and cap difficulty
	cappedDifficulty, err := validateAndCapDifficulty(difficulty)
	if err != nil {
		panic(err) // Panic for synchronous API consistency
	}

	difficultyBig := new(big.Int).SetUint64(cappedDifficulty)
	threshold := GetThresholdByDifficulty(difficultyBig)
	nonce := uint64(0)

	for {
		hash := computeHash(dataHash, nonce)
		hashValue := hashToUint64(hash)

		if hashValue <= threshold {
			return uint64ToHex(nonce)
		}

		nonce++
	}
}

// GeneratePowBigInt is like GeneratePoW but accepts difficulty as *big.Int
//
// Note: This function panics if difficulty exceeds MaxReasonableDifficulty.
// For error handling, use GeneratePowBigIntWithContext instead.
func GeneratePowBigInt(dataHash types.Hash, difficulty *big.Int) string {
	if difficulty.Cmp(big.NewInt(0)) == 0 {
		return "0000000000000000"
	}

	// Validate and cap difficulty
	cappedDifficulty, err := validateAndCapDifficultyBigInt(difficulty)
	if err != nil {
		panic(err) // Panic for synchronous API consistency
	}

	threshold := GetThresholdByDifficulty(cappedDifficulty)
	nonce := uint64(0)

	for {
		hash := computeHash(dataHash, nonce)
		hashValue := hashToUint64(hash)

		if hashValue <= threshold {
			return uint64ToHex(nonce)
		}

		nonce++
	}
}

// GeneratePowBytes is like GeneratePoW but returns nonce as bytes
func GeneratePowBytes(dataHash types.Hash, difficulty uint64) []byte {
	hexStr := GeneratePoW(dataHash, difficulty)
	return hexToBytes(hexStr)
}

// GeneratePowWithContext generates PoW with context support for cancellation
// Returns the nonce as a hex string or ErrCancelled if context is cancelled
// Checks context cancellation every 10000 iterations for efficiency
//
// Returns ErrDifficultyTooHigh if difficulty exceeds MaxReasonableDifficulty.
func GeneratePowWithContext(ctx context.Context, dataHash types.Hash, difficulty uint64) (string, error) {
	if difficulty == 0 {
		return "0000000000000000", nil
	}

	// Validate and cap difficulty
	cappedDifficulty, err := validateAndCapDifficulty(difficulty)
	if err != nil {
		return "", err
	}

	difficultyBig := new(big.Int).SetUint64(cappedDifficulty)
	threshold := GetThresholdByDifficulty(difficultyBig)
	nonce := uint64(0)
	checkInterval := uint64(10000) // Check context every 10k iterations

	for {
		// Check context cancellation periodically
		if nonce%checkInterval == 0 {
			select {
			case <-ctx.Done():
				return "", ErrCancelled
			default:
			}
		}

		hash := computeHash(dataHash, nonce)
		hashValue := hashToUint64(hash)

		if hashValue <= threshold {
			return uint64ToHex(nonce), nil
		}

		nonce++
	}
}

// GeneratePowBigIntWithContext is like GeneratePowWithContext but accepts difficulty as *big.Int
//
// Returns ErrDifficultyTooHigh if difficulty exceeds MaxReasonableDifficulty.
func GeneratePowBigIntWithContext(ctx context.Context, dataHash types.Hash, difficulty *big.Int) (string, error) {
	if difficulty.Cmp(big.NewInt(0)) == 0 {
		return "0000000000000000", nil
	}

	// Validate and cap difficulty
	cappedDifficulty, err := validateAndCapDifficultyBigInt(difficulty)
	if err != nil {
		return "", err
	}

	threshold := GetThresholdByDifficulty(cappedDifficulty)
	nonce := uint64(0)
	checkInterval := uint64(10000) // Check context every 10k iterations

	for {
		// Check context cancellation periodically
		if nonce%checkInterval == 0 {
			select {
			case <-ctx.Done():
				return "", ErrCancelled
			default:
			}
		}

		hash := computeHash(dataHash, nonce)
		hashValue := hashToUint64(hash)

		if hashValue <= threshold {
			return uint64ToHex(nonce), nil
		}

		nonce++
	}
}

// GeneratePowAsync generates PoW asynchronously and returns a channel.
// This provides a Dart-like async pattern while maintaining Go's context cancellation.
// The returned channel will receive exactly one result and then be closed.
//
// Worker Pool: This function uses a worker pool to limit concurrent PoW operations.
// By default, at most 8 PoW computations run simultaneously. Additional requests
// queue until a worker becomes available. This prevents CPU exhaustion when many
// transactions are submitted concurrently.
//
// Configure the worker pool using SetMaxPoWWorkers() or the POW_MAX_WORKERS
// environment variable before calling this function.
//
// This function immediately returns a read-only channel and spawns a goroutine
// to generate the PoW. The caller can wait for the result by reading from the channel.
//
// Usage:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	resultChan := pow.GeneratePowAsync(ctx, hash, difficulty)
//	result := <-resultChan
//	if result.Error != nil {
//	    // Handle error (timeout, cancellation, etc.)
//	    return
//	}
//	// Use result.Nonce
//
// For concurrent operations (automatically queued if >8 simultaneous):
//
//	results := make([]<-chan PowResult, 20)
//	for i := 0; i < 20; i++ {
//	    results[i] = pow.GeneratePowAsync(ctx, hashes[i], difficulty)
//	}
//	for i := 0; i < 20; i++ {
//	    result := <-results[i]  // First 8 start immediately, rest queue
//	    // Process result
//	}
func GeneratePowAsync(ctx context.Context, dataHash types.Hash, difficulty uint64) <-chan PowResult {
	initWorkerPool()
	resultChan := make(chan PowResult, 1)

	go func() {
		defer close(resultChan)

		// Acquire worker slot (blocks if pool is full)
		if err := pool.acquire(ctx); err != nil {
			resultChan <- PowResult{
				Nonce: "",
				Error: err,
			}
			return
		}
		defer pool.release()

		nonce, err := GeneratePowWithContext(ctx, dataHash, difficulty)
		resultChan <- PowResult{
			Nonce: nonce,
			Error: err,
		}
	}()

	return resultChan
}

// GeneratePowBigIntAsync is like GeneratePowAsync but accepts *big.Int difficulty.
// This is useful when difficulty exceeds uint64 range or comes from contract data.
//
// Worker Pool: Like GeneratePowAsync, this function uses the worker pool to limit
// concurrent PoW operations. At most 8 PoW computations run simultaneously by default.
//
// Usage:
//
//	difficulty := big.NewInt(100000)
//	resultChan := pow.GeneratePowBigIntAsync(ctx, hash, difficulty)
//	result := <-resultChan
//	if result.Error != nil {
//	    // Handle error
//	    return
//	}
//	// Use result.Nonce
func GeneratePowBigIntAsync(ctx context.Context, dataHash types.Hash, difficulty *big.Int) <-chan PowResult {
	initWorkerPool()
	resultChan := make(chan PowResult, 1)

	go func() {
		defer close(resultChan)

		// Acquire worker slot (blocks if pool is full)
		if err := pool.acquire(ctx); err != nil {
			resultChan <- PowResult{
				Nonce: "",
				Error: err,
			}
			return
		}
		defer pool.release()

		nonce, err := GeneratePowBigIntWithContext(ctx, dataHash, difficulty)
		resultChan <- PowResult{
			Nonce: nonce,
			Error: err,
		}
	}()

	return resultChan
}

// GetThresholdByDifficulty calculates the threshold value for a given difficulty
// threshold = 2^64 / difficulty
func GetThresholdByDifficulty(difficulty *big.Int) uint64 {
	if difficulty.Cmp(big.NewInt(0)) == 0 {
		return ^uint64(0) // Return max uint64 for zero difficulty
	}

	// Calculate 2^64 / difficulty
	maxUint64 := new(big.Int).SetUint64(^uint64(0))
	maxUint64.Add(maxUint64, big.NewInt(1)) // 2^64

	threshold := new(big.Int).Div(maxUint64, difficulty)

	// If threshold exceeds uint64, return max
	if !threshold.IsUint64() {
		return ^uint64(0)
	}

	return threshold.Uint64()
}

// CheckPoW verifies that a nonce is valid for the given hash and difficulty
func CheckPoW(dataHash types.Hash, nonce uint64, difficulty uint64) bool {
	if difficulty == 0 {
		return true
	}

	difficultyBig := new(big.Int).SetUint64(difficulty)
	threshold := GetThresholdByDifficulty(difficultyBig)

	hash := computeHash(dataHash, nonce)
	hashValue := hashToUint64(hash)

	return hashValue <= threshold
}

// BenchmarkPoW performs a quick PoW generation benchmark
// Returns the time taken and the nonce found
func BenchmarkPoW(difficulty uint64) (nonce string, iterations uint64) {
	// Use a fixed test hash for consistent benchmarking
	testHash := types.Hash{}
	copy(testHash[:], []byte("benchmark_test_hash_for_pow_"))

	difficultyBig := new(big.Int).SetUint64(difficulty)
	threshold := GetThresholdByDifficulty(difficultyBig)
	nonceVal := uint64(0)

	for {
		hash := computeHash(testHash, nonceVal)
		hashValue := hashToUint64(hash)

		if hashValue <= threshold {
			return uint64ToHex(nonceVal), nonceVal
		}

		nonceVal++
	}
}

// computeHash computes SHA3-256(dataHash || nonce)
func computeHash(dataHash types.Hash, nonce uint64) []byte {
	hasher := sha3.New256()
	hasher.Write(dataHash.Bytes())
	hasher.Write(uint64ToBytes(nonce))
	return hasher.Sum(nil)
}

// hashToUint64 converts the first 8 bytes of a hash to uint64 (big-endian)
func hashToUint64(hash []byte) uint64 {
	if len(hash) < 8 {
		return 0
	}
	return binary.BigEndian.Uint64(hash[:8])
}

// uint64ToBytes converts a uint64 to 8-byte array (big-endian)
func uint64ToBytes(n uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, n)
	return buf
}

// uint64ToHex converts a uint64 to a 16-character hex string
func uint64ToHex(n uint64) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, n)

	const hexChars = "0123456789abcdef"
	result := make([]byte, 16)
	for i := 0; i < 8; i++ {
		result[i*2] = hexChars[buf[i]>>4]
		result[i*2+1] = hexChars[buf[i]&0x0f]
	}
	return string(result)
}

// hexToBytes converts a hex string to bytes
func hexToBytes(hex string) []byte {
	if len(hex)%2 != 0 {
		hex = "0" + hex
	}

	result := make([]byte, len(hex)/2)
	for i := 0; i < len(hex); i += 2 {
		high := hexCharToValue(hex[i])
		low := hexCharToValue(hex[i+1])
		result[i/2] = (high << 4) | low
	}
	return result
}

// hexCharToValue converts a hex character to its numeric value
func hexCharToValue(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	default:
		return 0
	}
}
