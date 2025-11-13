# Zenon SDK Port Roadmap: Dart → Go

**Status**: In Progress
**Approach**: Methodical function-by-function port with unit testing
**Goal**: Full feature parity with Dart SDK

## Progress Overview

- [x] Phase 1: Foundation (ABI + Embedded Definitions)
- [x] Phase 2: Utils Enhancement
- [x] Phase 3: Crypto & Argon2
- [x] Phase 4: Wallet System
- [ ] Phase 5: PoW Module
- [ ] Phase 6: WebSocket Client Enhancement
- [ ] Phase 7: HTLC API
- [ ] Phase 8: Testing & Documentation

---

## Phase 1: Foundation - ABI Module & Embedded Definitions

**Priority**: CRITICAL
**Status**: ✅ Complete
**Estimated**: 5-7 days

### 1.1 ABI Types (`abi/types.go`)

#### Core Type System
- [x] `AbiType` interface
  - [x] `GetName()` method
  - [x] `GetCanonicalName()` method
  - [x] `Encode(value interface{})` method
  - [x] `Decode(encoded []byte, offset int)` method
  - [x] `GetFixedSize()` method
  - [x] `IsDynamicType()` method
  - [x] Unit test: Interface compliance

#### Numeric Types
- [x] `NumericType` abstract base
  - [x] `EncodeInternal(value interface{})` - Convert to big.Int
  - [x] Unit test: Value conversion

- [x] `IntType` (signed integers)
  - [x] Constructor with size validation (int8 to int256)
  - [x] `GetCanonicalName()` - Returns "intN"
  - [x] `Encode(value)` - Encode signed integer
  - [x] `Decode(encoded, offset)` - Decode to big.Int
  - [x] `EncodeInt(i int)` - Static encoder
  - [x] `EncodeIntBig(bigInt *big.Int)` - Static big encoder
  - [x] `DecodeInt(encoded, offset)` - Static decoder
  - [x] Unit test: int8 encoding/decoding
  - [x] Unit test: int32 encoding/decoding
  - [x] Unit test: int256 encoding/decoding
  - [x] Unit test: Negative numbers
  - [x] Unit test: Zero value
  - [x] Unit test: Max/min values

- [x] `UnsignedIntType` (unsigned integers)
  - [x] Constructor with size validation (uint8 to uint256)
  - [x] `GetCanonicalName()` - Returns "uintN"
  - [x] `Encode(value)` - Encode unsigned integer
  - [x] `Decode(encoded, offset)` - Decode to big.Int
  - [x] `DecodeUint(encoded, offset)` - Static unsigned decoder
  - [x] `EncodeUint(i uint64)` - Static encoder
  - [x] `EncodeUintBig(bigInt *big.Int)` - Validate non-negative
  - [x] Unit test: uint8 encoding/decoding
  - [x] Unit test: uint256 encoding/decoding
  - [x] Unit test: Reject negative values
  - [x] Unit test: Max values (2^255, max uint64)

#### Boolean Type
- [x] `BoolType`
  - [x] Constructor
  - [x] `Encode(value)` - Encode as 0/1
  - [x] `Decode(encoded, offset)` - Decode to bool
  - [x] Unit test: true encoding
  - [x] Unit test: false encoding
  - [x] Unit test: Invalid values

#### Address Types
- [x] `AddressType`
  - [x] Constructor
  - [x] `Encode(value)` - Encode Address with padding
  - [x] `Decode(encoded, offset)` - Decode to types.Address
  - [x] Unit test: Valid address encoding
  - [x] Unit test: Address decoding
  - [x] Unit test: Invalid address handling

#### Hash Types
- [x] `HashType`
  - [x] Constructor
  - [x] `Encode(value)` - Encode Hash
  - [x] `Decode(encoded, offset)` - Decode to types.Hash
  - [x] Unit test: Hash encoding/decoding

- [x] `Bytes32Type`
  - [x] Constructor with name parameter
  - [x] `Encode(value)` - Encode fixed 32-byte value
  - [x] `Decode(encoded, offset)` - Decode 32 bytes
  - [x] Unit test: Fixed-size encoding

#### Token Standard Type
- [x] `TokenStandardType`
  - [x] Constructor
  - [x] `Encode(value)` - Encode TokenStandard
  - [x] `Decode(encoded, offset)` - Decode to types.ZenonTokenStandard
  - [x] Unit test: ZNN token standard
  - [x] Unit test: QSR token standard
  - [x] Unit test: Custom ZTS

#### Bytes Types
- [x] `BytesType` (dynamic bytes)
  - [x] Constructor
  - [x] `Encode(value)` - Encode with length prefix and padding
  - [x] `Decode(encoded, offset)` - Decode dynamic bytes
  - [x] `IsDynamicType()` - Returns true
  - [x] Unit test: Empty bytes
  - [x] Unit test: Small bytes (< 32)
  - [x] Unit test: Large bytes (> 32)
  - [x] Unit test: Padding verification

- [x] `StringType` (extends BytesType)
  - [x] Constructor
  - [x] `Encode(value)` - Encode UTF-8 string
  - [x] `Decode(encoded, offset)` - Decode to string
  - [x] Unit test: ASCII string
  - [x] Unit test: UTF-8 string
  - [x] Unit test: Empty string

#### Array Types
- [x] `ArrayType` interface
  - [x] `GetElementType()` method
  - [x] `EncodeTuple(l []interface{})` method
  - [x] `DecodeTuple(encoded, offset, len)` method
  - [x] `EncodeList(l []interface{})` abstract method

- [x] `StaticArrayType`
  - [x] Constructor with size parameter
  - [x] `GetCanonicalName()` - Returns "typeN"
  - [x] `EncodeList(l)` - Encode fixed-size array
  - [x] `Decode(encoded, offset)` - Decode fixed array
  - [x] `GetFixedSize()` - Returns element_size * count
  - [x] Unit test: uint256[3] encoding/decoding
  - [x] Unit test: address[5] encoding/decoding
  - [x] Unit test: Size mismatch error

- [x] `DynamicArrayType`
  - [x] Constructor
  - [x] `GetCanonicalName()` - Returns "type[]"
  - [x] `EncodeList(l)` - Encode with length prefix
  - [x] `Decode(encoded, offset)` - Decode dynamic array
  - [x] `IsDynamicType()` - Returns true
  - [x] Unit test: Empty array
  - [x] Unit test: uint256[] encoding/decoding
  - [x] Unit test: Nested arrays

#### Function Type
- [x] `FunctionType`
  - [x] Constructor
  - [x] `Encode(value)` - Encode 24-byte function selector
  - [x] `Decode(encoded, offset)` - Decode (unimplemented)
  - [x] Unit test: Function selector encoding

#### Type Factory
- [x] `GetType(typeName string)` - Factory method
  - [x] Parse "int" types (int8-int256)
  - [x] Parse "uint" types (uint8-uint256)
  - [x] Parse "bool" type
  - [x] Parse "address" type
  - [x] Parse "string" type
  - [x] Parse "bytes" type
  - [x] Parse "bytesN" types (bytes1-bytes32)
  - [x] Parse "hash" type
  - [x] Parse "tokenStandard" type
  - [x] Parse static arrays "type[N]"
  - [x] Parse dynamic arrays "type[]"
  - [x] Parse "function" type
  - [x] Unit test: All type parsing
  - [x] Unit test: Invalid type names

### 1.2 ABI Encoding/Decoding (`abi/abi.go`)

#### Parameter Model
- [x] `Param` struct
  - [x] `Indexed` field (bool)
  - [x] `Name` field (string)
  - [x] `Type` field (AbiType)
  - [x] Constructor (`NewParam`)

- [x] `DecodeList(params []Param, encoded []interface{})` function
  - [x] Decode parameter list
  - [x] Handle static vs dynamic offsets
  - [x] Unit test: Multiple parameters
  - [x] Unit test: Mixed static/dynamic

#### Entry Model
- [x] `Entry` struct
  - [x] `Name` field (string)
  - [x] `Inputs` field ([]Param)
  - [x] `Type` field (TypeEnum)
  - [x] Constructor (`NewEntry`)

- [x] `FormatSignature()` method
  - [x] Format as "functionName(type1,type2,...)"
  - [x] Unit test: Simple signature
  - [x] Unit test: Complex signature

- [x] `FingerprintSignature()` method
  - [x] SHA3-256 hash of signature
  - [x] Return as bytes
  - [x] Unit test: Known function fingerprint

- [x] `EncodeSignature()` method
  - [x] Full signature hash
  - [x] Unit test: Signature encoding

- [x] `EncodeArguments(args []interface{})` method
  - [x] Encode function arguments
  - [x] Handle head/tail for dynamic types
  - [x] Unit test: Static arguments only
  - [x] Unit test: Dynamic arguments
  - [x] Unit test: Mixed arguments

#### ABI Function
- [x] `AbiFunction` struct (extends Entry)
  - [x] `EncodedSignLength` constant = 4
  - [x] Constructor (`NewAbiFunction`)

- [x] `Decode(encoded []byte)` method
  - [x] Skip signature and decode arguments
  - [x] Decode arguments
  - [x] Unit test: Valid function call

- [x] `Encode(args []interface{})` method
  - [x] Encode 4-byte signature + arguments
  - [x] Unit test: Complete encoding

- [x] `EncodeSignature()` method (override)
  - [x] Return first 4 bytes of hash
  - [x] Unit test: Signature extraction

- [x] `ExtractSignature(data []byte)` helper function
  - [x] Extract first 4 bytes
  - [x] Unit test: Signature extraction

#### ABI Container
- [x] `Abi` struct
  - [x] `Entries` field ([]Entry)
  - [x] Constructor (`NewAbi`)

- [x] `NewAbi(entries []Entry)` constructor
  - [x] Initialize from entry list
  - [x] Unit test: Construction

- [x] `FromJson(jsonStr string)` constructor
  - [x] Parse ABI from JSON
  - [x] Parse function entries (only functions supported)
  - [x] Unit test: Valid JSON
  - [x] Unit test: Invalid JSON
  - [x] Unit test: Missing name/type fields
  - [x] Unit test: Invalid param types

- [x] `EncodeFunction(name string, args []interface{})` method
  - [x] Find function by name
  - [x] Encode call
  - [x] Unit test: Known function
  - [x] Unit test: Unknown function

- [x] `DecodeFunction(encoded []byte)` method
  - [x] Extract signature
  - [x] Find matching function
  - [x] Decode arguments
  - [x] Unit test: Complete decoding
  - [x] Unit test: Unknown signature
  - [x] Unit test: Round-trip encoding/decoding

### 1.3 Embedded Contract Definitions (`embedded/definitions.go`)

#### Contract ABI Definitions (JSON strings)
- [x] `PlasmaDefinition` constant
  - [x] JSON ABI for Plasma contract
  - [x] Methods: Fuse, CancelFuse

- [x] `PillarDefinition` constant
  - [x] JSON ABI for Pillar contract
  - [x] Methods: Register, RegisterLegacy, UpdatePillar, Revoke, Delegate, Undelegate

- [x] `TokenDefinition` constant
  - [x] JSON ABI for Token contract
  - [x] Methods: IssueToken, Mint, Burn, UpdateToken

- [x] `SentinelDefinition` constant
  - [x] JSON ABI for Sentinel contract
  - [x] Methods: Register, Revoke

- [x] `SwapDefinition` constant
  - [x] JSON ABI for Swap contract
  - [x] Methods: RetrieveAssets

- [x] `StakeDefinition` constant
  - [x] JSON ABI for Stake contract
  - [x] Methods: Stake, Cancel

- [x] `AcceleratorDefinition` constant
  - [x] JSON ABI for Accelerator contract
  - [x] Methods: CreateProject, AddPhase, UpdatePhase, Donate, VoteByName, VoteByProdAddress

- [x] `SporkDefinition` constant
  - [x] JSON ABI for Spork contract
  - [x] Methods: CreateSpork, ActivateSpork

- [x] `HtlcDefinition` constant
  - [x] JSON ABI for HTLC contract
  - [x] Methods: Create, Reclaim, Unlock, DenyProxyUnlock, AllowProxyUnlock

- [x] `BridgeDefinition` constant
  - [x] JSON ABI for Bridge contract
  - [x] Methods: 21 functions including WrapToken, UnwrapToken, SetNetwork, SetTokenPair, etc.

- [x] `LiquidityDefinition` constant
  - [x] JSON ABI for Liquidity contract
  - [x] Methods: 15 functions including Fund, BurnZnn, Update, SetIsHalted, etc.

- [x] `CommonDefinition` constant
  - [x] JSON ABI for Common contract methods
  - [x] Methods: DepositQsr, WithdrawQsr, CollectReward, Update, Donate, VoteByName, VoteByProdAddress

#### Parsed ABI Objects
- [x] `Plasma` variable - Parsed Abi
  - [x] Initialize from PlasmaDefinition (init function)
  - [x] Unit test: Parse success
  - [x] Unit test: Fuse encoding

- [x] `Pillar` variable - Parsed Abi
  - [x] Initialize from PillarDefinition (init function)
  - [x] Unit test: Parse success
  - [x] Unit test: Delegate encoding

- [x] `Token` variable - Parsed Abi
  - [x] Initialize from TokenDefinition (init function)
  - [x] Unit test: Parse success
  - [x] Unit test: IssueToken encoding

- [x] `Sentinel` variable - Parsed Abi
  - [x] Initialize from SentinelDefinition (init function)
  - [x] Unit test: Parse success
  - [x] Unit test: Register encoding

- [x] `Swap` variable - Parsed Abi
  - [x] Initialize from SwapDefinition (init function)
  - [x] Unit test: Parse success

- [x] `Stake` variable - Parsed Abi
  - [x] Initialize from StakeDefinition (init function)
  - [x] Unit test: Parse success
  - [x] Unit test: Stake encoding

- [x] `Accelerator` variable - Parsed Abi
  - [x] Initialize from AcceleratorDefinition (init function)
  - [x] Unit test: Parse success

- [x] `Spork` variable - Parsed Abi
  - [x] Initialize from SporkDefinition (init function)
  - [x] Unit test: Parse success

- [x] `Htlc` variable - Parsed Abi
  - [x] Initialize from HtlcDefinition (init function)
  - [x] Unit test: Parse success

- [x] `Bridge` variable - Parsed Abi
  - [x] Initialize from BridgeDefinition (init function)
  - [x] Unit test: Parse success

- [x] `Liquidity` variable - Parsed Abi
  - [x] Initialize from LiquidityDefinition (init function)
  - [x] Unit test: Parse success

- [x] `Common` variable - Parsed Abi
  - [x] Initialize from CommonDefinition (init function)
  - [x] Unit test: Parse success

### 1.4 Embedded Constants (`embedded/constants.go`)

- [x] Base constants (CoinDecimals, OneZnn, OneQsr)
- [x] `GenesisTimestamp` constant = 1637755200
- [x] Unit tests for all constants

#### Plasma Constants
- [x] `FuseMinQsrAmount` constant (10 QSR)
- [x] `MinPlasmaAmount` constant (21000)
- [x] Unit tests

#### Pillar Constants
- [x] `MinDelegationAmount` constant (1 ZNN)
- [x] `PillarRegisterZnnAmount` constant (15000 ZNN)
- [x] `PillarRegisterQsrAmount` constant (150000 QSR)
- [x] `PillarNameMaxLength` constant (40)
- [x] `PillarNameRegExp` constant
- [x] Unit tests with valid/invalid name patterns

#### Sentinel Constants
- [x] `SentinelRegisterZnnAmount` constant (5000 ZNN)
- [x] `SentinelRegisterQsrAmount` constant (50000 QSR)
- [x] Unit tests

#### Staking Constants
- [x] `StakeMinZnnAmount` constant (1 ZNN)
- [x] `StakeTimeUnitSec` constant (30 days)
- [x] `StakeTimeMinSec` constant (1 month)
- [x] `StakeTimeMaxSec` constant (12 months)
- [x] `StakeUnitDurationName` constant = "month"
- [x] Unit tests

#### Token Constants
- [x] `TokenZtsIssueFeeInZnn` constant (1 ZNN)
- [x] `MinTokenTotalMaxSupply` constant (1)
- [x] `BigP255` constant (2^255)
- [x] `BigP255m1` constant (2^255 - 1)
- [x] `TokenNameMaxLength` constant (40)
- [x] `TokenSymbolMaxLength` constant (10)
- [x] `TokenSymbolExceptions` slice (ZNN, QSR)
- [x] `TokenNameRegExp` constant
- [x] `TokenSymbolRegExp` constant
- [x] `TokenDomainRegExp` constant
- [x] Unit tests with valid/invalid patterns

#### Accelerator Constants
- [x] `ProjectCreationFeeInZnn` constant (1 ZNN)
- [x] `ZnnProjectMaximumFunds` constant (5000 ZNN)
- [x] `QsrProjectMaximumFunds` constant (50000 QSR)
- [x] `ZnnProjectMinimumFunds` constant (10 ZNN)
- [x] `QsrProjectMinimumFunds` constant (100 QSR)
- [x] `ProjectDescriptionMaxLength` constant (240)
- [x] `ProjectNameMaxLength` constant (30)
- [x] `ProjectVotingStatus` constant (0)
- [x] `ProjectActiveStatus` constant (1)
- [x] `ProjectPaidStatus` constant (2)
- [x] `ProjectClosedStatus` constant (3)
- [x] `ProjectUrlRegExp` constant
- [x] Unit tests

#### Swap Constants
- [x] `SwapAssetDecayTimestampStart` constant (1645531200)
- [x] `SwapAssetDecayEpochsOffset` constant (90)
- [x] `SwapAssetDecayTickEpochs` constant (30)
- [x] `SwapAssetDecayTickValuePercentage` constant (10)
- [x] Unit tests

#### Spork Constants
- [x] `SporkNameMinLength` constant (5)
- [x] `SporkNameMaxLength` constant (40)
- [x] `SporkDescriptionMaxLength` constant (400)
- [x] Unit tests

#### HTLC Constants
- [x] `HtlcPreimageMinLength` constant (1)
- [x] `HtlcPreimageMaxLength` constant (255)
- [x] `HtlcPreimageDefaultLength` constant (32)
- [x] `HtlcHashTypeSha3` constant = 0
- [x] `HtlcHashTypeSha256` constant = 1
- [x] Unit tests

#### Bridge Constants
- [x] `BridgeMinGuardians` constant = 5
- [x] `BridgeMaximumFee` constant = 10000
- [x] Unit tests

### 1.5 Embedded Validations (`embedded/validations.go`)

- [x] `ValidateTokenName(value string)` function
  - [x] Check length (max 40 characters)
  - [x] Check regex pattern (alphanumeric with separators)
  - [x] Return error or nil
  - [x] Unit test: Valid names
  - [x] Unit test: Invalid names (empty, too long, invalid pattern)

- [x] `ValidateTokenSymbol(value string)` function
  - [x] Check length (max 10 characters)
  - [x] Check regex pattern (uppercase alphanumeric)
  - [x] Check exceptions list (ZNN, QSR reserved)
  - [x] Return error or nil
  - [x] Unit test: Valid symbols
  - [x] Unit test: Reserved symbols
  - [x] Unit test: Invalid patterns

- [x] `ValidateTokenDomain(value string)` function
  - [x] Check empty
  - [x] Check regex pattern (valid domain format)
  - [x] Return error or nil
  - [x] Unit test: Valid domains
  - [x] Unit test: Invalid domains

- [x] `ValidatePillarName(value string)` function
  - [x] Check length (max 40 characters)
  - [x] Check regex pattern (alphanumeric with separators)
  - [x] Return error or nil
  - [x] Unit test: Valid pillar names
  - [x] Unit test: Invalid pillar names

- [x] `ValidateProjectName(value string)` function
  - [x] Check length (max 30 characters)
  - [x] Return error or nil
  - [x] Unit test: Valid names
  - [x] Unit test: Too long

- [x] `ValidateProjectDescription(value string)` function
  - [x] Check length (max 240 characters)
  - [x] Return error or nil
  - [x] Unit test: Valid descriptions
  - [x] Unit test: Too long

---

## Phase 2: Utils Enhancement

**Priority**: HIGH
**Status**: ✅ Complete
**Estimated**: 2-3 days

### 2.1 NOM Constants (`utils/constants.go`)

- [x] `CoinDecimals` constant = 8 (imported from embedded)
- [x] `OneZnn` constant = 100000000 (imported from embedded)
- [x] `OneQsr` constant = 100000000 (imported from embedded)
- [x] `IntervalBetweenMomentums` constant = 10 seconds
- [x] Unit test: Constant values

### 2.2 Bytes Utilities (`utils/bytes.go`)

- [x] `Arraycopy(src, startPos, dest, destPos, len)` function
  - [x] Copy array slice
  - [x] Unit test: Normal copy
  - [x] Unit test: Boundary conditions (with offsets)

- [x] `DecodeBigInt(bytes []byte)` function
  - [x] Convert bytes to big.Int (big-endian)
  - [x] Unit test: Positive numbers
  - [x] Unit test: Large numbers

- [x] `EncodeBigInt(number *big.Int)` function
  - [x] Convert big.Int to bytes (big-endian)
  - [x] Unit test: Encode/decode round-trip (including 2^128)

- [x] `BigIntToBytes(b *big.Int, numBytes int)` function
  - [x] Convert to fixed-size byte array (32 bytes tested)
  - [x] Unit test: Various sizes

- [x] `BigIntToBytesSigned(b *big.Int, numBytes int)` function
  - [x] Convert signed big.Int
  - [x] Unit test: Negative numbers (0xFF padding)
  - [x] Unit test: Positive numbers (0x00 padding)

- [x] `BytesToBigInt(bb []byte)` function
  - [x] Convert bytes to big.Int
  - [x] Unit test: Round-trip
  - [x] Unit test: Empty bytes

- [x] `Merge(arrays [][]byte)` function
  - [x] Concatenate byte arrays
  - [x] Unit test: Multiple arrays
  - [x] Unit test: Empty arrays
  - [x] Unit test: With nil arrays

- [x] `IntToBytes(integer int32)` function
  - [x] Convert int32 to 4 bytes (big-endian)
  - [x] Unit test: Various integers

- [x] `LongToBytes(longValue int64)` function
  - [x] Convert int64 to 8 bytes (big-endian)
  - [x] Unit test: Various longs

- [x] `Base64ToBytes(base64Str string)` function
  - [x] Decode base64
  - [x] Unit test: Valid base64
  - [x] Unit test: Empty string

- [x] `BytesToBase64(bytes []byte)` function
  - [x] Encode to base64
  - [x] Unit test: Round-trip

- [x] `BytesToHex(bytes []byte)` function
  - [x] Convert to hex string
  - [x] Unit test: Various inputs (empty, single byte, multiple bytes)

- [x] `LeftPadBytes(bytes []byte, size int)` function
  - [x] Pad bytes on left with zeros
  - [x] Unit test: Padding to 6 bytes
  - [x] Unit test: Already correct size
  - [x] Unit test: Larger than target

### 2.3 Amount Utilities (`utils/amount.go`)

- [x] `ExtractDecimals(amount string, decimals int)` function
  - [x] Parse decimal string to big.Int
  - [x] Remove decimal point and pad/truncate
  - [x] Unit test: "1.5" with 8 decimals → 150000000
  - [x] Unit test: "100" with 8 decimals → 10000000000
  - [x] Unit test: "0.00000001" with 8 decimals → 1
  - [x] Unit test: Truncation of extra decimals
  - [x] Unit test: Padding of short decimals
  - [x] Unit test: Invalid formats

- [x] `AddDecimals(number *big.Int, decimals int)` function
  - [x] Convert big.Int to decimal string
  - [x] Insert decimal point
  - [x] Strip trailing zeros
  - [x] Unit test: 100000000 with 8 decimals → "1"
  - [x] Unit test: 1 with 8 decimals → "0.00000001"
  - [x] Unit test: Zero handling
  - [x] Unit test: No decimals (decimals=0)
  - [x] Unit test: Round-trip with ExtractDecimals

### 2.4 Block Utilities (`utils/block.go`)

- [x] Block type constants (Unknown, GenesisReceive, UserSend, UserReceive, ContractSend, ContractReceive)
- [x] `IsSendBlock(blockType int)` function
  - [x] Check if block type is send (UserSend or ContractSend)
  - [x] Unit test: Send block types
  - [x] Unit test: Receive block types
  - [x] Unit test: Unknown type

- [x] `IsReceiveBlock(blockType int)` function
  - [x] Check if block type is receive (UserReceive, GenesisReceive, or ContractReceive)
  - [x] Unit test: Receive block types
  - [x] Unit test: Send block types
  - [x] Unit test: Unknown type

- [ ] `GetTransactionHash(transaction *nom.AccountBlock)` function
  - [ ] Calculate SHA3 hash of transaction
  - [ ] Unit test: Known transaction hash

- [ ] `GetTransactionBytes(transaction *nom.AccountBlock)` function
  - [ ] Serialize transaction to bytes
  - [ ] Unit test: Serialization

- [ ] `Send(transaction *nom.AccountBlock, keyPair *wallet.KeyPair, ...)` function
  - [ ] Complete transaction preparation
  - [ ] Auto-fill parameters
  - [ ] Calculate PoW
  - [ ] Sign transaction
  - [ ] Publish to node
  - [ ] Unit test: Mock send

- [ ] `RequiresPoW(transaction *nom.AccountBlock, ...)` function
  - [ ] Check if PoW is required
  - [ ] Query plasma availability
  - [ ] Unit test: With plasma
  - [ ] Unit test: Without plasma

#### Private Block Utilities
- [ ] `getPoWData(transaction)` function
  - [ ] Extract PoW challenge data
  - [ ] Unit test: PoW data extraction

- [ ] `autofillTransactionParameters(transaction)` function
  - [ ] Query frontier block
  - [ ] Set height
  - [ ] Set previous hash
  - [ ] Set momentum acknowledged
  - [ ] Unit test: First block (height 1)
  - [ ] Unit test: Subsequent block

- [ ] `checkAndSetFields(transaction, keyPair)` function
  - [ ] Set address
  - [ ] Set public key
  - [ ] Validate receive blocks
  - [ ] Validate nonce
  - [ ] Unit test: Send block
  - [ ] Unit test: Receive block
  - [ ] Unit test: Invalid receive

- [ ] `setDifficulty(transaction, ...)` function
  - [ ] Query required PoW
  - [ ] Generate nonce if needed
  - [ ] Set fused plasma
  - [ ] Unit test: No PoW required
  - [ ] Unit test: PoW required

- [ ] `setHashAndSignature(transaction, keyPair)` function
  - [ ] Compute hash
  - [ ] Sign with keypair
  - [ ] Unit test: Signature verification

---

## Phase 3: Crypto & Argon2

**Priority**: HIGH
**Status**: ✅ Complete
**Estimated**: 2-3 days

### 3.1 Argon2 Wrapper (`crypto/argon2.go`)

- [x] `Argon2Parameters` struct
  - [x] Memory, Iterations, Parallelism, SaltLength, KeyLength fields
  - [x] Unit test: Default parameters
- [x] `DefaultArgon2Parameters()` function
  - [x] Returns default parameters (64MB, 1 iteration, 4 threads, 32-byte key)
  - [x] Unit test: Parameter values
- [x] `DeriveKey(password, salt []byte, params Argon2Parameters)` function
  - [x] Use golang.org/x/crypto/argon2.IDKey
  - [x] Return derived key
  - [x] Unit test: Same inputs same output
  - [x] Unit test: Different passwords different outputs
  - [x] Unit test: Different salts different outputs
  - [x] Unit test: Custom parameters
  - [x] Unit test: Empty password
  - [x] Unit test: Known vector
- [x] `DeriveKeyDefault(password, salt []byte)` function
  - [x] Use default parameters
  - [x] Unit test: Matches manual call with defaults
  - [x] Unit test: Output length

### 3.2 Crypto Utilities (`crypto/crypto.go`)

- [x] `GetPublicKey(privateKey []byte)` function
  - [x] Derive Ed25519 public key
  - [x] Validate private key size
  - [x] Unit test: Valid private key
  - [x] Unit test: Invalid key sizes
  - [x] Unit test: Output size

- [x] `Sign(message, privateKey []byte)` function
  - [x] Ed25519 signature
  - [x] Validate private key size
  - [x] Unit test: Valid signature
  - [x] Unit test: Invalid key sizes
  - [x] Unit test: Empty message
  - [x] Unit test: Different messages different signatures
  - [x] Unit test: Same message same signature

- [x] `Verify(signature, message, publicKey []byte)` function
  - [x] Verify Ed25519 signature
  - [x] Validate signature and public key sizes
  - [x] Unit test: Valid signature
  - [x] Unit test: Invalid signature (tampered)
  - [x] Unit test: Wrong message
  - [x] Unit test: Wrong public key
  - [x] Unit test: Invalid sizes

- [x] `Digest(data []byte, digestSize int)` function
  - [x] SHA3-256 digest (default 32 bytes)
  - [x] SHAKE256 for custom sizes
  - [x] Unit test: Default size (32)
  - [x] Unit test: Zero size (uses default)
  - [x] Unit test: Custom size (64)
  - [x] Unit test: Empty data
  - [x] Unit test: Deterministic
  - [x] Unit test: Different data different hash
  - [x] Unit test: Known SHA3-256 vector

- [x] `DigestDefault(data []byte)` function
  - [x] SHA3-256 with 32-byte output
  - [x] Unit test: Matches Digest(data, 32)

- [x] `SHA256Bytes(data []byte)` function
  - [x] SHA-256 hash
  - [x] Unit test: Empty data (known vector)
  - [x] Unit test: Deterministic
  - [x] Unit test: Different data different hash

- [x] Round-trip tests
  - [x] Unit test: Sign/Verify round-trip with various message sizes

### 3.3 Ed25519 Implementation

**Note**: BIP32 key derivation will be implemented in Phase 4 (Wallet System) as it's tightly coupled with wallet functionality.

---

## Phase 4: Wallet System

**Priority**: HIGH
**Status**: ✅ Complete
**Estimated**: 5-7 days

### 4.1 Wallet Interfaces (`wallet/interfaces.go`)

- [ ] `WalletDefinition` interface
  - [ ] `GetWalletId()` method
  - [ ] `GetWalletName()` method

- [ ] `WalletOptions` interface
  - [ ] (Empty marker interface)

- [ ] `WalletManager` interface
  - [ ] `GetWalletDefinitions()` method
  - [ ] `GetWallet(definition, options)` method
  - [ ] `SupportsWallet(definition)` method

- [ ] `Wallet` interface
  - [ ] `GetAccount(index)` method

- [ ] `WalletAccount` interface
  - [ ] `GetPublicKey()` method
  - [ ] `GetAddress()` method
  - [ ] `Sign(message)` method
  - [ ] `SignTx(tx)` method

### 4.2 Wallet Exceptions (`wallet/exceptions.go`)

- [ ] `WalletException` struct
  - [ ] `Message` field
  - [ ] `Error()` method
  - [ ] Unit test: Error string

- [ ] `IncorrectPasswordException` struct
  - [ ] Constructor with default message
  - [ ] Unit test: Exception creation

### 4.3 Wallet Constants (`wallet/constants.go`)

- [x] `BaseAddressKey` constant = "baseAddress"
- [x] `WalletTypeKey` constant = "walletType"
- [x] `KeyStoreWalletType` constant = "keystore"
- [x] `DefaultMaxIndex` constant = 10000

### 4.4 Derivation (`wallet/derivation.go`)

- [x] `CoinType` constant = "73404"
- [x] `DerivationPath` constant = "m/44'/73404'"
- [x] `GetDerivationAccount(account int)` function
  - [x] Return BIP44 path for account
  - [x] Unit test: Account 0 path
  - [x] Unit test: Account 5 path

### 4.5 BIP32 Derivation (`wallet/bip32.go`)

- [x] SLIP-0010 Ed25519 implementation
- [x] `GetMasterKeyFromSeed(seed []byte)` function
  - [x] HMAC-SHA512 with "ed25519 seed"
  - [x] 30 unit tests with known vectors
- [x] `DerivePath(path string, seed []byte)` function
  - [x] BIP44 path parsing
  - [x] Hardened derivation only
- [x] `getCKDPriv(parent *KeyData, index uint32)` function
  - [x] Child key derivation

### 4.6 Mnemonic (`wallet/mnemonic.go`)

- [x] `GenerateMnemonic(strength int)` function
  - [x] Use go-bip39 library
  - [x] Generate random mnemonic
  - [x] Unit test: 128-bit (12 words)
  - [x] Unit test: 256-bit (24 words)

- [x] `ValidateMnemonic(words []string)` function
  - [x] Check word count
  - [x] Verify checksum
  - [x] Unit test: Valid mnemonic
  - [x] Unit test: Invalid checksum

- [x] `IsValidWord(word string)` function
  - [x] Check if word in BIP39 wordlist
  - [x] Unit test: Valid word
  - [x] Unit test: Invalid word

- [x] `MnemonicToSeed(mnemonic, passphrase string)` function
- [x] `MnemonicToEntropy(mnemonic string)` function
- [x] `EntropyToMnemonic(entropy []byte)` function
- [x] 24 comprehensive unit tests

### 4.7 KeyPair (`wallet/keypair.go`)

- [x] `KeyPair` struct
  - [x] `privateKey` field ([]byte)
  - [x] `publicKey` field ([]byte, cached)
  - [x] `address` field (*types.Address, cached)

- [x] `NewKeyPair(privateKey []byte)` constructor
  - [x] Initialize keypair
  - [x] Unit test: Constructor

- [x] `NewKeyPairFromSeed(seed []byte)` constructor
  - [x] Create from 32-byte seed
  - [x] Unit test: Seed initialization

- [x] `GetPrivateKey()` method
  - [x] Return private key bytes
  - [x] Unit test: Key retrieval

- [x] `GetPublicKey()` method
  - [x] Lazy derivation with caching
  - [x] Return public key
  - [x] Unit test: Public key derivation

- [x] `GetAddress()` method
  - [x] Lazy derivation with caching
  - [x] Return Zenon address
  - [x] Unit test: Address derivation

- [x] `Sign(message []byte)` method
  - [x] Ed25519 signature
  - [x] Unit test: Signature generation

- [x] `Verify(signature, message []byte)` method
  - [x] Verify Ed25519 signature
  - [x] Unit test: Valid signature
  - [x] Unit test: Invalid signature

- [x] `GeneratePublicKey(privateKey []byte)` static method
  - [x] Derive public key
  - [x] Unit test: Public key generation

- [x] 23 comprehensive unit tests

### 4.8 Encrypted File (`wallet/encryptedfile.go`)

- [x] `EncryptedFile` struct
  - [x] `Metadata` field (map[string]interface{})
  - [x] `Crypto` field (*CryptoParams)
  - [x] `Timestamp` field (int64)
  - [x] `Version` field (int)

- [x] `CryptoParams` struct
  - [x] `Argon2Params` field (*Argon2Params)
  - [x] `CipherData` field (hex string)
  - [x] `CipherName` field ("aes-256-gcm")
  - [x] `Kdf` field ("argon2.IDKey")
  - [x] `Nonce` field (hex string)

- [x] `Argon2Params` struct
  - [x] `Salt` field (hex string)

- [x] `Encrypt(data []byte, password string, metadata map[string]interface{})` function
  - [x] Generate 16-byte salt, 12-byte nonce
  - [x] Derive key with Argon2 (64MB, 1 iter, 4 threads)
  - [x] Encrypt with AES-256-GCM, AAD="zenon"
  - [x] Return EncryptedFile
  - [x] Unit test: Encrypt/decrypt round-trip

- [x] `FromJSON(jsonData []byte)` constructor
  - [x] Parse JSON
  - [x] Unit test: Valid JSON
  - [x] Unit test: Invalid JSON

- [x] `Decrypt(password string)` method
  - [x] Derive key with Argon2
  - [x] Decrypt with AES-256-GCM
  - [x] Verify authentication tag
  - [x] Return ErrIncorrectPassword on failure
  - [x] Unit test: Correct password
  - [x] Unit test: Incorrect password

- [x] `ToJSON()` method
  - [x] Serialize to JSON
  - [x] Unit test: JSON output

- [x] 22 comprehensive unit tests

### 4.9 KeyStore (`wallet/keystore.go`)

- [x] `KeyStore` struct
  - [x] `Mnemonic` field (string)
  - [x] `Entropy` field ([]byte)
  - [x] `Seed` field ([]byte)

- [x] `NewKeyStoreFromMnemonic(mnemonic string)` constructor
  - [x] Validate mnemonic
  - [x] Derive entropy and seed
  - [x] Unit test: Valid mnemonic
  - [x] Unit test: Invalid mnemonic

- [x] `NewKeyStoreFromSeed(seedHex string)` constructor
  - [x] Set seed directly
  - [x] Unit test: Seed initialization

- [x] `NewKeyStoreFromEntropy(entropy []byte)` constructor
  - [x] Convert to mnemonic
  - [x] Unit test: Entropy conversion

- [x] `NewKeyStoreRandom()` function
  - [x] Generate random 256-bit mnemonic
  - [x] Create keystore
  - [x] Unit test: Random generation

- [x] `GetKeyPair(account int)` method
  - [x] Derive BIP44 keypair (m/44'/73404'/account')
  - [x] Unit test: Keypair derivation

- [x] `DeriveAddressesByRange(left, right int)` method
  - [x] Derive address range
  - [x] Unit test: Range 0-10

- [x] `FindAddress(address types.Address, maxAccounts int)` method
  - [x] Search for address in keystore
  - [x] Return FindResponse with index and keypair
  - [x] Unit test: Address found
  - [x] Unit test: Address not found

- [x] `FindResponse` struct
  - [x] `Index` field (int)
  - [x] `KeyPair` field (*KeyPair)

- [x] `GetBaseAddress()` method
  - [x] Return address at account 0

- [x] `ToEncryptedFile(password string, metadata map[string]interface{})` method
  - [x] Encrypt keystore to EncryptedFile

- [x] `FromEncryptedFile(ef *EncryptedFile, password string)` function
  - [x] Decrypt EncryptedFile to KeyStore

- [x] Custom JSON serialization helpers
- [x] 27 comprehensive unit tests

### 4.10 KeyStore Manager (`wallet/manager.go`)

- [x] `KeyStoreManager` struct
  - [x] `WalletPath` field (directory path)

- [x] `NewKeyStoreManager(walletPath string)` constructor
  - [x] Initialize manager
  - [x] Create directory if needed
  - [x] Unit test: Manager creation

- [x] `SaveKeyStore(store *KeyStore, password, name string)` method
  - [x] Serialize keystore to JSON
  - [x] Encrypt with password
  - [x] Write to file with 0600 permissions
  - [x] Unit test: Save and load

- [x] `ReadKeyStore(password string, keyStoreFile string)` method
  - [x] Read encrypted file
  - [x] Decrypt with password
  - [x] Deserialize keystore
  - [x] Unit test: Read keystore
  - [x] Unit test: Wrong password

- [x] `FindKeyStore(name string)` method
  - [x] Search for keystore by name
  - [x] Case-insensitive fallback
  - [x] Unit test: Find existing
  - [x] Unit test: Not found

- [x] `ListAllKeyStores()` method
  - [x] List all keystores in directory
  - [x] Filter hidden files
  - [x] Unit test: Empty directory
  - [x] Unit test: Multiple keystores

- [x] `CreateNew(passphrase, name string)` method
  - [x] Generate random keystore
  - [x] Save to file
  - [x] Unit test: Create new wallet

- [x] `CreateFromMnemonic(mnemonic, passphrase, name string)` method
  - [x] Create from existing mnemonic
  - [x] Save to file
  - [x] Unit test: Import mnemonic

- [x] `GetKeystoreInfo(keyStoreFile string)` method
  - [x] Read metadata without decryption
  - [x] Unit test: Get info

- [x] 26 comprehensive unit tests

---

## Phase 5: PoW Module

**Priority**: MEDIUM
**Status**: Not Started
**Estimated**: 3-4 days

### 5.1 PoW Generation (`pow/pow.go`)

- [ ] `PowStatus` enum
  - [ ] `Generating` value
  - [ ] `Done` value

- [ ] `GeneratePoW(hash types.Hash, difficulty uint64)` function
  - [ ] Option 1: Pure Go implementation
  - [ ] Option 2: CGO binding to C library
  - [ ] Option 3: Use go-zenon's implementation
  - [ ] Return nonce as hex string
  - [ ] Unit test: Known PoW solution
  - [ ] Unit test: Difficulty 1000
  - [ ] Benchmark test: Performance

- [ ] `BenchmarkPoW(difficulty uint64)` function
  - [ ] Benchmark PoW generation
  - [ ] Return test nonce
  - [ ] Unit test: Benchmark run

---

## Phase 6: WebSocket Client Enhancement

**Priority**: MEDIUM
**Status**: Not Started
**Estimated**: 2-3 days

### 6.1 WebSocket Status (`rpc_client/status.go`)

- [ ] `WebsocketStatus` enum
  - [ ] `Uninitialized` value
  - [ ] `Connecting` value
  - [ ] `Running` value
  - [ ] `Stopped` value

### 6.2 WebSocket Client Enhancement (`rpc_client/client.go`)

- [ ] `ConnectionEstablishedCallback` type
  - [ ] Function signature for callbacks

- [ ] Add fields to `RpcClient`:
  - [ ] `onConnectionCallbacks` slice
  - [ ] `websocketIntendedState` status
  - [ ] `restartedEventChan` channel

- [ ] `IsClosed()` method
  - [ ] Check if connection is closed
  - [ ] Unit test: Closed state

- [ ] `AddOnConnectionEstablishedCallback(callback)` method
  - [ ] Register connection callback
  - [ ] Unit test: Callback registration

- [ ] `Initialize(url string, retry bool)` method
  - [ ] Connect to WebSocket
  - [ ] Set up reconnection logic
  - [ ] Unit test: Successful connection
  - [ ] Unit test: Connection failure

- [ ] `Status()` method
  - [ ] Return current connection status
  - [ ] Unit test: Status check

- [ ] `Restart()` method
  - [ ] Reconnect if disconnected
  - [ ] Trigger callbacks
  - [ ] Unit test: Restart

- [ ] `Stop()` method (enhance existing)
  - [ ] Close connection gracefully
  - [ ] Unit test: Clean shutdown

### 6.3 Connection Monitoring (`rpc_client/monitor.go`)

- [ ] `startMonitoring()` function
  - [ ] Ping/pong health check
  - [ ] Auto-reconnect on disconnect
  - [ ] Unit test: Monitoring

- [ ] `handleDisconnect()` function
  - [ ] Detect disconnection
  - [ ] Trigger reconnection
  - [ ] Unit test: Disconnect handling

### 6.4 WebSocket Utils (`rpc_client/utils.go`)

- [ ] `ValidateWsConnectionURL(url string)` function
  - [ ] Parse URL
  - [ ] Check scheme (ws/wss)
  - [ ] Validate port
  - [ ] Unit test: Valid URL
  - [ ] Unit test: Invalid URL
  - [ ] Unit test: Missing port

---

## Phase 7: HTLC API

**Priority**: MEDIUM
**Status**: Not Started
**Estimated**: 1-2 days

### 7.1 HTLC API (`api/embedded/htlc.go`)

- [ ] `HtlcApi` struct
  - [ ] `client` field (*server.Client)

- [ ] `NewHtlcApi(client)` constructor
  - [ ] Initialize API
  - [ ] Unit test: Constructor

#### RPC Methods
- [ ] `GetById(id types.Hash)` method
  - [ ] Call "embedded.htlc.getById"
  - [ ] Return HtlcInfo
  - [ ] Unit test: Mock RPC call

- [ ] `GetProxyUnlockStatus(address types.Address)` method
  - [ ] Call "embedded.htlc.getProxyUnlockStatus"
  - [ ] Return bool
  - [ ] Unit test: Mock RPC call

#### Contract Methods (return *nom.AccountBlock)
- [ ] `Create(token types.ZenonTokenStandard, amount *big.Int, hashLocked types.Address, expirationTime int64, hashType int, keyMaxSize int, hashLock []byte)` method
  - [ ] Build AccountBlock template
  - [ ] Encode ABI call
  - [ ] Unit test: Template creation

- [ ] `Reclaim(id types.Hash)` method
  - [ ] Build reclaim template
  - [ ] Encode ABI call
  - [ ] Unit test: Reclaim template

- [ ] `Unlock(id types.Hash, preimage []byte)` method
  - [ ] Build unlock template
  - [ ] Encode ABI call
  - [ ] Unit test: Unlock template

- [ ] `DenyProxyUnlock()` method
  - [ ] Build deny template
  - [ ] Encode ABI call
  - [ ] Unit test: Deny template

- [ ] `AllowProxyUnlock()` method
  - [ ] Build allow template
  - [ ] Encode ABI call
  - [ ] Unit test: Allow template

---

## Phase 8: Testing & Documentation

**Priority**: HIGH
**Status**: Not Started
**Estimated**: 3-5 days

### 8.1 Integration Tests

- [ ] End-to-end transaction test
  - [ ] Create wallet
  - [ ] Connect to node
  - [ ] Send transaction
  - [ ] Verify on-chain

- [ ] Contract interaction tests
  - [ ] Token issuance
  - [ ] Pillar registration
  - [ ] Plasma fusion
  - [ ] HTLC creation/unlock

- [ ] Wallet compatibility tests
  - [ ] Create wallet with SDK
  - [ ] Import to Dart SDK
  - [ ] Verify addresses match

### 8.2 Examples

- [ ] Update `examples/` directory
  - [ ] Wallet creation example
  - [ ] Token operations example
  - [ ] HTLC example
  - [ ] Complete transaction flow

### 8.3 Documentation

- [ ] Update README.md
  - [ ] Installation instructions
  - [ ] Quick start guide
  - [ ] API reference links

- [ ] Update CLAUDE.md
  - [ ] New module documentation
  - [ ] Architecture updates
  - [ ] Testing instructions

- [ ] Create MIGRATION.md
  - [ ] Dart → Go migration guide
  - [ ] API mapping table
  - [ ] Common patterns

### 8.4 Benchmarks

- [ ] ABI encoding/decoding benchmarks
- [ ] PoW generation benchmarks
- [ ] Crypto operation benchmarks
- [ ] Transaction signing benchmarks

---

## Notes & Decisions

### Implementation Strategy
- **Function-level tracking**: Each function gets a checkbox
- **Unit test requirement**: No function is "complete" without passing unit tests
- **Incremental commits**: Commit after each major function or small group of related functions
- **Test-driven approach**: Write tests alongside implementation

### Testing Philosophy
- Unit tests for all public functions
- Integration tests for API interactions
- Compatibility tests with Dart SDK
- Benchmark tests for performance-critical code

### Dependencies
- **Avoid go-zenon where possible**: Create SDK-specific implementations
- **Use standard library**: Prefer Go standard library over external deps
- **Minimal external deps**:
  - `golang.org/x/crypto` for Argon2, Ed25519, SHA3
  - `github.com/tyler-smith/go-bip39` for mnemonic
  - Existing go-zenon types package (can't avoid - core types)

### Code Quality Standards
- Go fmt for all code
- Go vet passes
- golangci-lint passes
- 80%+ test coverage target

---

## Revision History

| Date | Change | Notes |
|------|--------|-------|
| 2025-11-12 | Initial roadmap created | Function-level inventory complete |
| | | |

---

**Next Step**: Implement `IntType` in `abi/types.go` with full unit test coverage.
