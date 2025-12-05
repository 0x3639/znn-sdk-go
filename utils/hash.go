package utils

import (
	"github.com/0x3639/znn-sdk-go/crypto"
	"github.com/zenon-network/go-zenon/common/types"
)

// HashDigest creates a types.Hash by computing SHA3-256 of the input data.
//
// This is the standard way to create hashes for transaction data, data fields,
// and other content that needs to be hashed in the Zenon protocol. It wraps
// the crypto.DigestDefault function and returns a properly typed Hash.
//
// Parameters:
//   - data: Arbitrary byte data to hash. Can be empty, in which case
//     the hash of empty bytes is returned.
//
// Returns a types.Hash containing the 32-byte SHA3-256 hash.
//
// Example:
//
//	// Hash transaction data
//	dataHash := utils.HashDigest(transactionData)
//	fmt.Println("Data hash:", dataHash.String())
//
//	// Hash a simple string
//	hash := utils.HashDigest([]byte("hello world"))
//	fmt.Printf("Hash: %s\n", hash.String())
//
// See also: HashDigestEmpty for hashing empty data.
func HashDigest(data []byte) types.Hash {
	digest := crypto.DigestDefault(data)
	// DigestDefault always returns exactly 32 bytes, and BytesToHash only
	// fails if input length != 32, so error is impossible here
	hash, err := types.BytesToHash(digest)
	if err != nil {
		// This should never happen with SHA3-256 output
		panic("HashDigest: unexpected error from BytesToHash: " + err.Error())
	}
	return hash
}

// HashDigestEmpty returns the hash of empty data (SHA3-256 of empty bytes).
//
// This is commonly used for empty descendentBlocks in transaction serialization.
// The result is a constant value: the SHA3-256 hash of an empty byte array.
//
// Returns a types.Hash representing SHA3-256([]).
//
// Example:
//
//	// Get the hash used for empty descendentBlocks
//	emptyHash := utils.HashDigestEmpty()
//	fmt.Printf("Empty hash: %s\n", emptyHash.String())
//	// Output: a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a
//
// Note: This is equivalent to HashDigest([]byte{}).
func HashDigestEmpty() types.Hash {
	return HashDigest([]byte{})
}
