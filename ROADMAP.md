# Zenon Go SDK - Missing Encoding Methods Roadmap

This document outlines encoding methods that are present in the reference Dart SDK but missing from the Go SDK. These methods are essential for transaction serialization, hash generation, and ID computation.

## Primary Use Case: Stake/Fusion/HTLC ID Computation

**Problem:** When creating stakes, plasma fusions, liquidity stakes, HTLCs, or accelerator projects, the protocol assigns an ID to each entry. This ID is simply the **transaction hash** of the block that created the entry. Without `GetTransactionHash()`, SDK users cannot:

1. **Predict IDs before sending** - Useful for pre-computing cancel transaction data
2. **Verify IDs after sending** - Confirm the on-chain ID matches expected value
3. **Cancel entries** - The `Cancel()` methods require the entry ID

**Solution:** Implement `GetTransactionHash()` which computes the transaction hash (and thus the entry ID) from a fully-populated `AccountBlock`.

**Reference:** The Dart SDK provides `BlockUtils.getTransactionHash()` in `lib/src/utils/block.dart:27-29`:
```dart
static Hash getTransactionHash(AccountBlockTemplate transaction) {
  return Hash.digest(getTransactionBytes(transaction));
}
```

**Protocol Evidence:** The go-zenon node assigns IDs using `sendBlock.Hash`:
- `vm/embedded/implementation/stake.go:68`: `Id: sendBlock.Hash`
- `vm/embedded/implementation/plasma.go:61`: `Id: sendBlock.Hash`
- `vm/embedded/implementation/liquidity.go:419`: `Id: sendBlock.Hash`
- `vm/embedded/implementation/htlc.go:105`: `Id: sendBlock.Hash`
- `vm/embedded/implementation/accelerator.go:161,243,512`: `Id: sendBlock.Hash`

---

## Overview

| Method | Purpose | Priority | Status |
|--------|---------|----------|--------|
| `HashDigest()` | Create Hash from SHA3-256 of data | High (dependency) | Missing |
| `HashHeight.GetBytes()` | Serialize hash+height to 40 bytes | High (dependency) | Missing |
| `GetTransactionBytes()` | Serialize AccountBlock to bytes | High (dependency) | Missing |
| `GetTransactionHash()` | Compute transaction hash / entry ID | **Critical** | Missing |
| `GetPoWData()` | Get PoW input hash | Medium | Missing |
| `AddressFromPublicKey()` | Derive address from public key | Low (go-zenon has alternative) | Missing |

---

## Implementation Order

The implementation order is driven by dependencies, with the critical `GetTransactionHash()` as the end goal:

1. **`utils/hash.go`** - Foundation: `HashDigest()` wraps SHA3-256
2. **`utils/hash_height.go`** - `HashHeight` type with `GetBytes()` serialization
3. **`utils/block.go` extensions** - `GetTransactionBytes()`, `GetTransactionHash()`, `GetPoWData()`
4. **`utils/address.go`** - Optional: `AddressFromPublicKey()`

---

## 1. HashDigest

**Purpose:** Create a `types.Hash` by computing SHA3-256 of input data.

**Location:** `utils/hash.go` (new file)

**Reference:** `reference/znn_sdk_dart-master/lib/src/model/primitives/hash.dart:24`

```go
// HashDigest creates a types.Hash by computing SHA3-256 of the input data.
//
// This is the standard way to create hashes for transaction data, data fields,
// and other content that needs to be hashed in the Zenon protocol.
//
// Parameters:
//   - data: Arbitrary byte data to hash
//
// Returns a types.Hash (32-byte SHA3-256 hash).
//
// Example:
//
//     dataHash := utils.HashDigest(transactionData)
//     fmt.Println("Data hash:", dataHash.String())
func HashDigest(data []byte) types.Hash {
    digest := crypto.DigestDefault(data)
    hash, _ := types.BytesToHash(digest)
    return hash
}

// HashDigestEmpty returns the hash of empty data (SHA3-256 of empty bytes).
// This is commonly used for empty descendentBlocks in transaction serialization.
func HashDigestEmpty() types.Hash {
    return HashDigest([]byte{})
}
```

**Dependencies:** `crypto/crypto.go` (DigestDefault already exists)

---

## 2. HashHeight Type

**Purpose:** A struct combining Hash + Height with serialization support. Used for momentum acknowledgment in transactions.

**Location:** `utils/hash_height.go` (new file)

**Reference:** `reference/znn_sdk_dart-master/lib/src/model/primitives/hash_height.dart`

```go
// HashHeight represents a combination of a block hash and its height.
// This is commonly used for momentum acknowledgment in transactions.
type HashHeight struct {
    Hash   types.Hash `json:"hash"`
    Height uint64     `json:"height"`
}

// EmptyHashHeight represents an empty HashHeight with zero hash and zero height.
var EmptyHashHeight = HashHeight{
    Hash:   types.ZeroHash,
    Height: 0,
}

// NewHashHeight creates a new HashHeight from a hash and height.
//
// Parameters:
//   - hash: The block hash
//   - height: The block height
//
// Returns a HashHeight struct.
func NewHashHeight(hash types.Hash, height uint64) HashHeight {
    return HashHeight{
        Hash:   hash,
        Height: height,
    }
}

// GetBytes serializes the HashHeight to bytes.
// Format: [32 bytes hash][8 bytes height (big-endian)]
//
// Returns a 40-byte slice.
func (hh HashHeight) GetBytes() []byte {
    return Merge([][]byte{
        hh.Hash.Bytes(),
        LongToBytes(int64(hh.Height)),
    })
}
```

**Dependencies:** `utils/bytes.go` (LongToBytes, Merge already exist)

---

## 3. Block Serialization and ID Computation

**Purpose:** Serialize AccountBlock to bytes for hashing and compute transaction hash (which is the entry ID for stakes, fusions, HTLCs, etc.).

**Location:** `utils/block.go` (extend existing file)

**Reference:** `reference/znn_sdk_dart-master/lib/src/utils/block.dart:27-75`

### GetTransactionBytes

```go
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
//   - block: The AccountBlock to serialize
//
// Returns the serialized bytes.
func GetTransactionBytes(block *nom.AccountBlock) []byte {
    versionBytes := LongToBytes(int64(block.Version))
    chainIdentifierBytes := LongToBytes(int64(block.ChainIdentifier))
    blockTypeBytes := LongToBytes(int64(block.BlockType))
    previousHashBytes := block.PreviousHash.Bytes()
    heightBytes := LongToBytes(int64(block.Height))

    // MomentumAcknowledged serialization
    momentumAcknowledgedBytes := Merge([][]byte{
        block.MomentumAcknowledged.Hash.Bytes(),
        LongToBytes(int64(block.MomentumAcknowledged.Height)),
    })

    addressBytes := block.Address.Bytes()
    toAddressBytes := block.ToAddress.Bytes()
    amountBytes := BigIntToBytes(block.Amount, 32)
    tokenStandardBytes := block.TokenStandard.Bytes()
    fromBlockHashBytes := block.FromBlockHash.Bytes()
    descendentBlocksBytes := HashDigestEmpty().Bytes()
    dataBytes := HashDigest(block.Data).Bytes()
    fusedPlasmaBytes := LongToBytes(int64(block.FusedPlasma))
    difficultyBytes := LongToBytes(int64(block.Difficulty))

    // Nonce: decode hex string, left-pad to 8 bytes
    nonceBytes := make([]byte, 8)
    if block.Nonce != "" {
        decoded, _ := hex.DecodeString(block.Nonce)
        nonceBytes = LeftPadBytes(decoded, 8)
    }

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
```

### GetTransactionHash

```go
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
//     // Create stake template
//     template := client.StakeApi.Stake(duration)
//
//     // Autofill transaction parameters (height, previousHash, momentum, etc.)
//     // ... populate template fields ...
//
//     // Compute the ID before sending
//     stakeId := utils.GetTransactionHash(template)
//     fmt.Println("Stake will have ID:", stakeId.String())
//
//     // Now send the transaction
//     // The on-chain stake entry will have this ID
func GetTransactionHash(block *nom.AccountBlock) types.Hash {
    return HashDigest(GetTransactionBytes(block))
}
```

### GetPoWData

```go
// GetPoWData computes the data hash used for PoW generation.
//
// This is computed as SHA3-256(address || previousHash).
//
// Parameters:
//   - block: The AccountBlock template
//
// Returns the PoW data hash.
func GetPoWData(block *nom.AccountBlock) types.Hash {
    return HashDigest(Merge([][]byte{
        block.Address.Bytes(),
        block.PreviousHash.Bytes(),
    }))
}
```

---

## 4. AddressFromPublicKey (Optional)

**Purpose:** Derive a Zenon address from an Ed25519 public key.

**Note:** The go-zenon package provides `types.PubKeyToAddress()` which is already used in `wallet/keypair.go`. This utility provides an explicit implementation for cases where the go-zenon function is not suitable.

**Location:** `utils/address.go` (new file)

**Reference:** `reference/znn_sdk_dart-master/lib/src/model/primitives/address.dart:106-111`

```go
// AddressFromPublicKey derives a Zenon address from an Ed25519 public key.
//
// This function computes the SHA3-256 hash of the public key and uses
// the first 19 bytes (plus a user type byte) to create a 20-byte address core.
//
// Parameters:
//   - publicKey: 32-byte Ed25519 public key
//
// Returns the derived Zenon address or an error if the public key is invalid.
//
// Note: This is a utility function. The go-zenon package provides
// types.PubKeyToAddress() which should be preferred when available.
func AddressFromPublicKey(publicKey []byte) (types.Address, error) {
    if len(publicKey) != 32 {
        return types.Address{}, errors.New("public key must be 32 bytes")
    }

    // SHA3-256 hash of public key
    digest := crypto.DigestDefault(publicKey)

    // Take first 19 bytes, prepend user type byte (0x00)
    core := make([]byte, 20)
    core[0] = 0x00 // user type byte
    copy(core[1:], digest[:19])

    return types.BytesToAddress(core)
}
```

---

## Serialization Field Sizes Summary

| Field | Size (bytes) | Notes |
|-------|--------------|-------|
| version | 8 | int64 big-endian |
| chainIdentifier | 8 | int64 big-endian |
| blockType | 8 | int64 big-endian |
| previousHash | 32 | Hash bytes |
| height | 8 | int64 big-endian |
| momentumAcknowledged | 40 | Hash(32) + Height(8) |
| address | 20 | Address core bytes |
| toAddress | 20 | Address core bytes |
| amount | 32 | big.Int unsigned |
| tokenStandard | 10 | ZTS bytes |
| fromBlockHash | 32 | Hash bytes |
| descendentBlocks | 32 | SHA3-256([]) |
| data | 32 | SHA3-256(data) |
| fusedPlasma | 8 | int64 big-endian |
| difficulty | 8 | int64 big-endian |
| nonce | 8 | hex decoded, left-padded |
| **Total** | **306** | |

---

## Test Vectors

### HashDigest

```go
// SHA3-256 of empty bytes
emptyHash := HashDigest([]byte{})
// Expected: c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470

// SHA3-256 of "test"
testHash := HashDigest([]byte("test"))
// Verify against known SHA3-256 output
```

### HashHeight.GetBytes

```go
hh := NewHashHeight(knownHash, 12345)
bytes := hh.GetBytes()
// Expected: 40 bytes total
// First 32 bytes: hash bytes
// Last 8 bytes: 0x0000000000003039 (12345 in big-endian)
```

### GetTransactionHash (ID Computation)

```go
// Create a stake transaction template with all fields populated
// Compute the hash
stakeId := GetTransactionHash(stakeBlock)

// Verify:
// - Total byte length of GetTransactionBytes should be 306
// - Hash should match the on-chain stake entry ID after sending
```

---

## Files to Create/Modify

| File | Action | Priority |
|------|--------|----------|
| `utils/hash.go` | CREATE | High |
| `utils/hash_test.go` | CREATE | High |
| `utils/hash_height.go` | CREATE | High |
| `utils/hash_height_test.go` | CREATE | High |
| `utils/block.go` | MODIFY (add serialization functions) | **Critical** |
| `utils/block_test.go` | CREATE | **Critical** |
| `utils/address.go` | CREATE (optional) | Low |
| `utils/address_test.go` | CREATE (optional) | Low |

---

## Use Cases

These utilities enable:

1. **Stake/Fusion/HTLC ID computation**: Compute `GetTransactionHash()` to predict or verify entry IDs
2. **Transaction verification**: Compute transaction hash locally and compare with on-chain hash
3. **Block reconstruction**: Serialize blocks from indexed data
4. **PoW validation**: Use `GetPoWData()` to verify PoW solutions
5. **Address derivation**: Derive addresses from public keys for account mapping
