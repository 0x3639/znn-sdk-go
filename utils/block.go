package utils

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
