package utils

import (
	"github.com/zenon-network/go-zenon/common/types"
)

// HashHeight represents a combination of a block hash and its height.
//
// This is commonly used for momentum acknowledgment in transactions.
// The serialization format is: [32 bytes hash][8 bytes height (big-endian)],
// totaling 40 bytes.
//
// Note: The go-zenon package provides types.HashHeight which is used in
// nom.AccountBlock. This SDK type provides additional serialization support
// via GetBytes() for transaction hashing.
type HashHeight struct {
	Hash   types.Hash `json:"hash"`
	Height uint64     `json:"height"`
}

// EmptyHashHeight represents an empty HashHeight with zero hash and zero height.
//
// Use this as a default value or to check if a HashHeight is uninitialized.
var EmptyHashHeight = HashHeight{
	Hash:   types.ZeroHash,
	Height: 0,
}

// NewHashHeight creates a new HashHeight from a hash and height.
//
// Parameters:
//   - hash: The block hash (32 bytes)
//   - height: The block height
//
// Returns a HashHeight struct.
//
// Example:
//
//	// Create from momentum data
//	momentum, _ := client.LedgerApi.GetFrontierMomentum()
//	hh := utils.NewHashHeight(momentum.Hash, momentum.Height)
//
//	// Serialize for transaction
//	bytes := hh.GetBytes() // 40 bytes
func NewHashHeight(hash types.Hash, height uint64) HashHeight {
	return HashHeight{
		Hash:   hash,
		Height: height,
	}
}

// GetBytes serializes the HashHeight to bytes.
//
// Format: [32 bytes hash][8 bytes height (big-endian)]
//
// Returns a 40-byte slice containing the serialized HashHeight.
//
// This serialization format is used in transaction serialization for the
// momentumAcknowledged field.
//
// Example:
//
//	hh := utils.NewHashHeight(hash, 12345)
//	bytes := hh.GetBytes()
//	// bytes is 40 bytes: first 32 are hash, last 8 are height in big-endian
//	// For height 12345 (0x3039), last 8 bytes are: 0x0000000000003039
func (hh HashHeight) GetBytes() []byte {
	return Merge([][]byte{
		hh.Hash.Bytes(),
		LongToBytes(int64(hh.Height)),
	})
}

// IsEmpty returns true if this HashHeight represents an empty/zero value.
//
// A HashHeight is considered empty if both the hash is ZeroHash and
// the height is 0.
//
// Example:
//
//	hh := utils.EmptyHashHeight
//	fmt.Println(hh.IsEmpty()) // true
//
//	hh2 := utils.NewHashHeight(someHash, 100)
//	fmt.Println(hh2.IsEmpty()) // false
func (hh HashHeight) IsEmpty() bool {
	return hh.Hash == types.ZeroHash && hh.Height == 0
}
