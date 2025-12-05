package utils

import (
	"math/big"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
)

// =============================================================================
// Block Type Constants
// =============================================================================

const (
	// BlockTypeUnknown represents an unknown block type
	BlockTypeUnknown = 0

	// BlockTypeGenesisReceive represents a genesis receive block
	BlockTypeGenesisReceive = 1

	// BlockTypeUserSend represents a user send block
	BlockTypeUserSend = 2

	// BlockTypeUserReceive represents a user receive block
	BlockTypeUserReceive = 3

	// BlockTypeContractSend represents a contract send block
	BlockTypeContractSend = 4

	// BlockTypeContractReceive represents a contract receive block
	BlockTypeContractReceive = 5
)

// =============================================================================
// Block Type Utilities
// =============================================================================

// IsSendBlock checks if the block type is a send block
func IsSendBlock(blockType int) bool {
	return blockType == BlockTypeUserSend || blockType == BlockTypeContractSend
}

// IsReceiveBlock checks if the block type is a receive block
func IsReceiveBlock(blockType int) bool {
	return blockType == BlockTypeUserReceive ||
		blockType == BlockTypeGenesisReceive ||
		blockType == BlockTypeContractReceive
}

// =============================================================================
// Block Serialization
// =============================================================================

// GetTransactionBytes serializes an AccountBlock to bytes for hashing.
//
// This follows the exact serialization format required by the Zenon protocol:
//   - version: 8 bytes (big-endian int64)
//   - chainIdentifier: 8 bytes (big-endian int64)
//   - blockType: 8 bytes (big-endian int64)
//   - previousHash: 32 bytes
//   - height: 8 bytes (big-endian int64)
//   - momentumAcknowledged: 40 bytes (hash + height)
//   - address: 20 bytes
//   - toAddress: 20 bytes
//   - amount: 32 bytes (big-endian, unsigned)
//   - tokenStandard: 10 bytes
//   - fromBlockHash: 32 bytes
//   - descendentBlocks: 32 bytes (SHA3-256 of empty bytes)
//   - data: 32 bytes (SHA3-256 of data field)
//   - fusedPlasma: 8 bytes (big-endian int64)
//   - difficulty: 8 bytes (big-endian int64)
//   - nonce: 8 bytes
//
// Total: 306 bytes
//
// Parameters:
//   - block: The AccountBlock to serialize. Must have all fields populated.
//     If Amount is nil, it will be treated as zero.
//
// Returns the 306-byte serialized representation of the block.
//
// Example:
//
//	// Serialize a block for manual hashing
//	bytes := utils.GetTransactionBytes(block)
//	fmt.Printf("Serialized block: %d bytes\n", len(bytes)) // 306 bytes
//
// Note: This function is typically used internally by GetTransactionHash.
// Most users should call GetTransactionHash directly.
//
// Reference: znn_sdk_dart/lib/src/utils/block.dart:31-70
func GetTransactionBytes(block *nom.AccountBlock) []byte {
	versionBytes := Uint64ToBytes(block.Version)
	chainIdentifierBytes := Uint64ToBytes(block.ChainIdentifier)
	blockTypeBytes := Uint64ToBytes(block.BlockType)
	previousHashBytes := block.PreviousHash.Bytes()
	heightBytes := Uint64ToBytes(block.Height)

	// MomentumAcknowledged serialization: hash (32 bytes) + height (8 bytes)
	momentumAcknowledgedBytes := Merge([][]byte{
		block.MomentumAcknowledged.Hash.Bytes(),
		Uint64ToBytes(block.MomentumAcknowledged.Height),
	})

	addressBytes := block.Address.Bytes()
	toAddressBytes := block.ToAddress.Bytes()

	// Amount: convert to 32-byte big-endian representation
	amount := block.Amount
	if amount == nil {
		amount = big.NewInt(0)
	}
	amountBytes := BigIntToBytes(amount, 32)

	tokenStandardBytes := block.TokenStandard.Bytes()
	fromBlockHashBytes := block.FromBlockHash.Bytes()

	// DescendentBlocks: always hash of empty bytes (descendant blocks not included directly)
	descendentBlocksBytes := HashDigestEmpty().Bytes()

	// Data: hash of the data field
	dataBytes := HashDigest(block.Data).Bytes()

	fusedPlasmaBytes := Uint64ToBytes(block.FusedPlasma)
	difficultyBytes := Uint64ToBytes(block.Difficulty)

	// Nonce: already 8 bytes in the Nonce struct
	nonceBytes := block.Nonce.Data[:]

	return Merge([][]byte{
		versionBytes,
		chainIdentifierBytes,
		blockTypeBytes,
		previousHashBytes,
		heightBytes,
		momentumAcknowledgedBytes,
		addressBytes,
		toAddressBytes,
		amountBytes,
		tokenStandardBytes,
		fromBlockHashBytes,
		descendentBlocksBytes,
		dataBytes,
		fusedPlasmaBytes,
		difficultyBytes,
		nonceBytes,
	})
}

// GetTransactionHash computes the transaction hash for an AccountBlock.
//
// This is computed as SHA3-256(GetTransactionBytes(block)).
//
// IMPORTANT: This hash is used as the ID for stakes, plasma fusions,
// liquidity stakes, HTLCs, and accelerator projects/phases. When you
// create one of these entries, the protocol assigns Id = transaction hash.
//
// Use cases:
//   - Predict the ID of a stake/fusion/HTLC before sending
//   - Verify the ID matches after the transaction is confirmed
//   - Compute the ID needed to cancel an entry
//
// Parameters:
//   - block: The AccountBlock to hash (must be fully populated)
//
// Returns the transaction hash (which equals the entry ID).
//
// Example - Predicting a stake ID:
//
//	// Create stake template
//	template := client.StakeApi.Stake(duration)
//
//	// Autofill transaction parameters (height, previousHash, momentum, etc.)
//	// ... populate template fields ...
//
//	// Compute the ID before sending
//	stakeId := utils.GetTransactionHash(template)
//	fmt.Println("Stake will have ID:", stakeId.String())
//
//	// Now send the transaction
//	// The on-chain stake entry will have this ID
//
// Reference: znn_sdk_dart/lib/src/utils/block.dart:27-29
func GetTransactionHash(block *nom.AccountBlock) types.Hash {
	return HashDigest(GetTransactionBytes(block))
}

// GetPoWData computes the data hash used for PoW generation.
//
// This is computed as SHA3-256(address || previousHash), where || denotes
// concatenation.
//
// The PoW nonce must satisfy: SHA3-256(PoWData || nonce) < target
// where target is derived from the difficulty.
//
// Parameters:
//   - block: The AccountBlock template. Must have Address and PreviousHash set.
//     Other fields are not used in the computation.
//
// Returns the PoW data hash (32 bytes).
//
// Example:
//
//	// Get PoW data for generating a nonce
//	powData := utils.GetPoWData(block)
//	fmt.Printf("PoW data hash: %s\n", powData.String())
//
//	// Use with PoW generation
//	nonce, err := pow.GeneratePow(powData, difficulty)
//
// Reference: znn_sdk_dart/lib/src/utils/block.dart:72-75
func GetPoWData(block *nom.AccountBlock) types.Hash {
	return HashDigest(Merge([][]byte{
		block.Address.Bytes(),
		block.PreviousHash.Bytes(),
	}))
}
