package pow

import (
	"context"
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/zenon-network/go-zenon/common/types"
	"golang.org/x/crypto/sha3"
)

var (
	// ErrCancelled is returned when PoW generation is cancelled via context
	ErrCancelled = errors.New("pow generation cancelled")
)

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

// GeneratePoW generates a valid proof-of-work nonce for the given hash and difficulty
// Returns the nonce as a hex string (without 0x prefix)
//
// The algorithm:
// 1. Iterate through nonce values starting from 0
// 2. For each nonce, compute: hash = SHA3-256(dataHash + nonce)
// 3. Interpret hash as big-endian uint256
// 4. Check if: hash * difficulty < 2^256
// 5. Return nonce when condition is met
func GeneratePoW(dataHash types.Hash, difficulty uint64) string {
	if difficulty == 0 {
		return "0000000000000000"
	}

	difficultyBig := new(big.Int).SetUint64(difficulty)
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
func GeneratePowBigInt(dataHash types.Hash, difficulty *big.Int) string {
	if difficulty.Cmp(big.NewInt(0)) == 0 {
		return "0000000000000000"
	}

	threshold := GetThresholdByDifficulty(difficulty)
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
func GeneratePowWithContext(ctx context.Context, dataHash types.Hash, difficulty uint64) (string, error) {
	if difficulty == 0 {
		return "0000000000000000", nil
	}

	difficultyBig := new(big.Int).SetUint64(difficulty)
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
func GeneratePowBigIntWithContext(ctx context.Context, dataHash types.Hash, difficulty *big.Int) (string, error) {
	if difficulty.Cmp(big.NewInt(0)) == 0 {
		return "0000000000000000", nil
	}

	threshold := GetThresholdByDifficulty(difficulty)
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
