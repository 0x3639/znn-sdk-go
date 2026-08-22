package crypto

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// Argon2 parameter bounds enforced by DeriveKey.
//
// These are defensive limits for parameters that may originate from untrusted
// sources (for example Argon2 cost fields read from a key file). Values above
// the maximums would let a small input force gigabytes of allocation or
// minutes of CPU work; zero iterations or zero parallelism cause a panic
// inside the Argon2 implementation.
const (
	// MinArgon2Iterations is the minimum accepted time cost.
	MinArgon2Iterations uint32 = 1
	// MaxArgon2Iterations is the maximum accepted time cost.
	MaxArgon2Iterations uint32 = 16
	// MinArgon2Memory is the minimum accepted memory cost in KiB. The Argon2
	// specification requires at least 8 KiB per lane; DeriveKey additionally
	// checks Memory >= 8*Parallelism.
	MinArgon2Memory uint32 = 8
	// MaxArgon2Memory is the maximum accepted memory cost in KiB (256 MiB).
	// One DeriveKey call can never allocate more than this.
	MaxArgon2Memory uint32 = 256 * 1024
	// MaxArgon2WorkKiB is the maximum accepted total work budget, measured as
	// Memory × Iterations in KiB-passes (1 GiB-pass). It bounds CPU time
	// independently of the per-parameter maxima so that the largest memory
	// and iteration values cannot be combined. The SDK default (64 MiB × 1)
	// uses 1/16 of this budget.
	MaxArgon2WorkKiB uint64 = 1024 * 1024
	// MinArgon2Parallelism is the minimum accepted degree of parallelism.
	MinArgon2Parallelism uint8 = 1
	// MaxArgon2Parallelism is the maximum accepted degree of parallelism.
	MaxArgon2Parallelism uint8 = 64
	// MinArgon2KeyLength is the minimum accepted derived key length in bytes.
	MinArgon2KeyLength uint32 = 16
	// MaxArgon2KeyLength is the maximum accepted derived key length in bytes.
	MaxArgon2KeyLength uint32 = 1024
	// MinArgon2SaltLength is the minimum accepted salt length in bytes.
	MinArgon2SaltLength = 8
)

// ErrInvalidArgon2Parameters is returned (wrapped) by DeriveKey and
// ValidateArgon2Parameters when a parameter is outside its accepted range.
var ErrInvalidArgon2Parameters = errors.New("invalid argon2 parameters")

// Argon2Parameters represents the parameters for Argon2 key derivation
type Argon2Parameters struct {
	Memory      uint32 // Memory in KB
	Iterations  uint32 // Number of iterations
	Parallelism uint8  // Degree of parallelism
	SaltLength  uint32 // Length of salt in bytes
	KeyLength   uint32 // Length of derived key in bytes
}

// DefaultArgon2Parameters returns the default parameters used by the Zenon SDK
// These match the parameters used in the Dart SDK:
// - Memory: 64 * 1024 KB (64 MB)
// - Iterations: 1
// - Parallelism: 4
// - KeyLength: 32 bytes
func DefaultArgon2Parameters() Argon2Parameters {
	return Argon2Parameters{
		Memory:      64 * 1024,
		Iterations:  1,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}
}

// ValidateArgon2Parameters checks that every Argon2 cost and output parameter
// is within the documented safe range.
//
// Parameters:
//   - params: The Argon2 parameters to validate.
//
// Returns an error wrapping ErrInvalidArgon2Parameters if Iterations, Memory,
// Parallelism, or KeyLength is outside [Min*, Max*]. SaltLength is not
// checked here because the salt itself is passed separately to DeriveKey.
//
// Example:
//
//	if err := crypto.ValidateArgon2Parameters(params); err != nil {
//	    return err
//	}
func ValidateArgon2Parameters(params Argon2Parameters) error {
	if params.Iterations < MinArgon2Iterations || params.Iterations > MaxArgon2Iterations {
		return fmt.Errorf("%w: iterations %d outside [%d, %d]",
			ErrInvalidArgon2Parameters, params.Iterations, MinArgon2Iterations, MaxArgon2Iterations)
	}
	if params.Memory < MinArgon2Memory || params.Memory > MaxArgon2Memory {
		return fmt.Errorf("%w: memory %d KiB outside [%d, %d]",
			ErrInvalidArgon2Parameters, params.Memory, MinArgon2Memory, MaxArgon2Memory)
	}
	if params.Parallelism < MinArgon2Parallelism || params.Parallelism > MaxArgon2Parallelism {
		return fmt.Errorf("%w: parallelism %d outside [%d, %d]",
			ErrInvalidArgon2Parameters, params.Parallelism, MinArgon2Parallelism, MaxArgon2Parallelism)
	}
	if params.KeyLength < MinArgon2KeyLength || params.KeyLength > MaxArgon2KeyLength {
		return fmt.Errorf("%w: key length %d outside [%d, %d]",
			ErrInvalidArgon2Parameters, params.KeyLength, MinArgon2KeyLength, MaxArgon2KeyLength)
	}
	if work := uint64(params.Memory) * uint64(params.Iterations); work > MaxArgon2WorkKiB {
		return fmt.Errorf("%w: memory×iterations %d KiB-passes exceeds work budget %d",
			ErrInvalidArgon2Parameters, work, MaxArgon2WorkKiB)
	}
	// Argon2 requires memory >= 8*parallelism KiB.
	if params.Memory < 8*uint32(params.Parallelism) {
		return fmt.Errorf("%w: memory %d KiB is below 8*parallelism (%d)",
			ErrInvalidArgon2Parameters, params.Memory, 8*uint32(params.Parallelism))
	}
	return nil
}

// DeriveKey derives a key from a password using Argon2id.
//
// This is used for wallet encryption/decryption. Parameters are validated with
// ValidateArgon2Parameters before any work is performed so that untrusted
// values (for example from a key file) cannot trigger a panic inside Argon2 or
// an unbounded memory/CPU allocation.
//
// Parameters:
//   - password: The password bytes.
//   - salt: The salt; must be at least MinArgon2SaltLength bytes.
//   - params: Argon2 cost and output parameters.
//
// Returns the derived key, or an error wrapping ErrInvalidArgon2Parameters
// when a parameter or the salt is out of range.
//
// Example:
//
//	key, err := crypto.DeriveKey([]byte(password), salt, crypto.DefaultArgon2Parameters())
//	if err != nil {
//	    return err
//	}
func DeriveKey(password []byte, salt []byte, params Argon2Parameters) ([]byte, error) {
	if err := ValidateArgon2Parameters(params); err != nil {
		return nil, err
	}
	if len(salt) < MinArgon2SaltLength {
		return nil, fmt.Errorf("%w: salt length %d below minimum %d",
			ErrInvalidArgon2Parameters, len(salt), MinArgon2SaltLength)
	}
	return argon2.IDKey(
		password,
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	), nil
}

// DeriveKeyDefault derives a key using the default parameters.
//
// It never fails on parameter validation (the defaults are always valid) but
// still returns an error for a salt shorter than MinArgon2SaltLength.
func DeriveKeyDefault(password []byte, salt []byte) ([]byte, error) {
	return DeriveKey(password, salt, DefaultArgon2Parameters())
}
