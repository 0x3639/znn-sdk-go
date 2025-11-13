# Zenon SDK Port Roadmap: Dart → Go

**Status**: In Progress
**Approach**: Methodical function-by-function port with unit testing
**Goal**: Full feature parity with Dart SDK

## Progress Overview

- [ ] Phase 1: Foundation (ABI + Embedded Definitions)
- [ ] Phase 2: Utils Enhancement
- [ ] Phase 3: Crypto & Argon2
- [ ] Phase 4: Wallet System
- [ ] Phase 5: PoW Module
- [ ] Phase 6: WebSocket Client Enhancement
- [ ] Phase 7: HTLC API
- [ ] Phase 8: Testing & Documentation

---

## Phase 1: Foundation - ABI Module & Embedded Definitions

**Priority**: CRITICAL
**Status**: Not Started
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
- [ ] `Param` struct
  - [ ] `Indexed` field (bool)
  - [ ] `Name` field (string)
  - [ ] `Type` field (AbiType)
  - [ ] Constructor

- [ ] `DecodeList(params []Param, encoded []interface{})` function
  - [ ] Decode parameter list
  - [ ] Handle static vs dynamic offsets
  - [ ] Unit test: Multiple parameters
  - [ ] Unit test: Mixed static/dynamic

#### Entry Model
- [ ] `Entry` struct
  - [ ] `Name` field (string)
  - [ ] `Inputs` field ([]Param)
  - [ ] `Type` field (TypeEnum)
  - [ ] Constructor

- [ ] `FormatSignature()` method
  - [ ] Format as "functionName(type1,type2,...)"
  - [ ] Unit test: Simple signature
  - [ ] Unit test: Complex signature

- [ ] `FingerprintSignature()` method
  - [ ] SHA3-256 hash of signature
  - [ ] Return as bytes
  - [ ] Unit test: Known function fingerprint

- [ ] `EncodeSignature()` method
  - [ ] Full signature hash
  - [ ] Unit test: Signature encoding

- [ ] `EncodeArguments(args []interface{})` method
  - [ ] Encode function arguments
  - [ ] Handle head/tail for dynamic types
  - [ ] Unit test: Static arguments only
  - [ ] Unit test: Dynamic arguments
  - [ ] Unit test: Mixed arguments

#### ABI Function
- [ ] `AbiFunction` struct (extends Entry)
  - [ ] `EncodedSignLength` constant = 4
  - [ ] Constructor

- [ ] `Decode(encoded []byte)` method
  - [ ] Verify signature match
  - [ ] Decode arguments
  - [ ] Unit test: Valid function call
  - [ ] Unit test: Invalid signature

- [ ] `Encode(args []interface{})` method
  - [ ] Encode 4-byte signature + arguments
  - [ ] Unit test: Complete encoding

- [ ] `EncodeSignature()` method (override)
  - [ ] Return first 4 bytes of hash
  - [ ] Unit test: Signature extraction

- [ ] `ExtractSignature(data []byte)` static method
  - [ ] Extract first 4 bytes
  - [ ] Unit test: Signature extraction

#### ABI Container
- [ ] `Abi` struct
  - [ ] `Entries` field ([]Entry)
  - [ ] Constructor

- [ ] `NewAbi(entries []Entry)` constructor
  - [ ] Initialize from entry list
  - [ ] Unit test: Construction

- [ ] `FromJson(jsonStr string)` constructor
  - [ ] Parse ABI from JSON
  - [ ] Parse function entries
  - [ ] Parse event entries
  - [ ] Unit test: Valid JSON
  - [ ] Unit test: Invalid JSON

- [ ] `EncodeFunction(name string, args []interface{})` method
  - [ ] Find function by name
  - [ ] Encode call
  - [ ] Unit test: Known function
  - [ ] Unit test: Unknown function

- [ ] `DecodeFunction(encoded []byte)` method
  - [ ] Extract signature
  - [ ] Find matching function
  - [ ] Decode arguments
  - [ ] Unit test: Complete decoding

### 1.3 Embedded Contract Definitions (`embedded/definitions.go`)

#### Contract ABI Definitions (JSON strings)
- [ ] `PlasmaDefinition` constant
  - [ ] JSON ABI for Plasma contract
  - [ ] Methods: Fuse, Cancel

- [ ] `PillarDefinition` constant
  - [ ] JSON ABI for Pillar contract
  - [ ] Methods: Register, UpdatePillar, Revoke, Delegate, Undelegate

- [ ] `TokenDefinition` constant
  - [ ] JSON ABI for Token contract
  - [ ] Methods: IssueToken, Mint, Burn, UpdateToken, TransferOwnership, DisableMint

- [ ] `SentinelDefinition` constant
  - [ ] JSON ABI for Sentinel contract
  - [ ] Methods: Register, Revoke, CollectReward

- [ ] `SwapDefinition` constant
  - [ ] JSON ABI for Swap contract
  - [ ] Methods: RetrieveAssets

- [ ] `StakeDefinition` constant
  - [ ] JSON ABI for Stake contract
  - [ ] Methods: Stake, Cancel, CollectReward

- [ ] `AcceleratorDefinition` constant
  - [ ] JSON ABI for Accelerator contract
  - [ ] Methods: CreateProject, AddPhase, UpdatePhase, Donate, VoteByName, VoteByProdAddress

- [ ] `SporkDefinition` constant
  - [ ] JSON ABI for Spork contract
  - [ ] Methods: CreateSpork, ActivateSpork

- [ ] `HtlcDefinition` constant
  - [ ] JSON ABI for HTLC contract
  - [ ] Methods: Create, Reclaim, Unlock, DenyProxyUnlock, AllowProxyUnlock

- [ ] `BridgeDefinition` constant
  - [ ] JSON ABI for Bridge contract
  - [ ] Methods: WrapToken, UnwrapToken, SetNetwork, RemoveNetwork, SetTokenPair, RemoveTokenPair, Halt, Unhalt, SetAllowKeyGen, SetBridgeMetadata, RevokeUnwrapRequest, Emergency, ChangeTssECDSAPubKey, ChangeAdministrator, ProposeAdministrator, SetOrchestrator, SetRedeemDelay, NominateGuardians, SetAllowedToRedeem, SetRedeemed

- [ ] `LiquidityDefinition` constant
  - [ ] JSON ABI for Liquidity contract
  - [ ] Methods: Fund, BurnZnn, Update, SetIsHalted, SetTokenTuple, NominateGuardians, ProposeAdministrator, Emergency, UnlockLiquidityEntries, SetAdditionalReward

- [ ] `CommonDefinition` constant
  - [ ] JSON ABI for Common contract methods
  - [ ] Methods: DepositQsr, WithdrawQsr, CollectReward

#### Parsed ABI Objects
- [ ] `Plasma` variable - Parsed Abi
  - [ ] Initialize from PlasmaDefinition
  - [ ] Unit test: Parse success

- [ ] `Pillar` variable - Parsed Abi
  - [ ] Initialize from PillarDefinition
  - [ ] Unit test: Parse success

- [ ] `Token` variable - Parsed Abi
  - [ ] Initialize from TokenDefinition
  - [ ] Unit test: Parse success
  - [ ] Unit test: IssueToken encoding

- [ ] `Sentinel` variable - Parsed Abi
  - [ ] Initialize from SentinelDefinition
  - [ ] Unit test: Parse success

- [ ] `Swap` variable - Parsed Abi
  - [ ] Initialize from SwapDefinition
  - [ ] Unit test: Parse success

- [ ] `Stake` variable - Parsed Abi
  - [ ] Initialize from StakeDefinition
  - [ ] Unit test: Parse success

- [ ] `Accelerator` variable - Parsed Abi
  - [ ] Initialize from AcceleratorDefinition
  - [ ] Unit test: Parse success

- [ ] `Spork` variable - Parsed Abi
  - [ ] Initialize from SporkDefinition
  - [ ] Unit test: Parse success

- [ ] `Htlc` variable - Parsed Abi
  - [ ] Initialize from HtlcDefinition
  - [ ] Unit test: Parse success

- [ ] `Bridge` variable - Parsed Abi
  - [ ] Initialize from BridgeDefinition
  - [ ] Unit test: Parse success

- [ ] `Liquidity` variable - Parsed Abi
  - [ ] Initialize from LiquidityDefinition
  - [ ] Unit test: Parse success

- [ ] `Common` variable - Parsed Abi
  - [ ] Initialize from CommonDefinition
  - [ ] Unit test: Parse success

### 1.4 Embedded Constants (`embedded/constants.go`)

- [ ] `GenesisTimestamp` constant = 1637755200

#### Plasma Constants
- [ ] `FuseMinQsrAmount` constant
- [ ] `MinPlasmaAmount` constant

#### Pillar Constants
- [ ] `MinDelegationAmount` constant
- [ ] `PillarRegisterZnnAmount` constant
- [ ] `PillarRegisterQsrAmount` constant
- [ ] `PillarNameMaxLength` constant
- [ ] `PillarNameRegExp` constant

#### Sentinel Constants
- [ ] `SentinelRegisterZnnAmount` constant
- [ ] `SentinelRegisterQsrAmount` constant

#### Staking Constants
- [ ] `StakeMinZnnAmount` constant
- [ ] `StakeTimeUnitSec` constant
- [ ] `StakeTimeMinSec` constant
- [ ] `StakeTimeMaxSec` constant
- [ ] `StakeUnitDurationName` constant = "month"

#### Token Constants
- [ ] `TokenZtsIssueFeeInZnn` constant
- [ ] `MinTokenTotalMaxSupply` constant
- [ ] `BigP255` constant
- [ ] `BigP255m1` constant
- [ ] `TokenNameMaxLength` constant
- [ ] `TokenSymbolMaxLength` constant
- [ ] `TokenSymbolExceptions` slice
- [ ] `TokenNameRegExp` constant
- [ ] `TokenSymbolRegExp` constant
- [ ] `TokenDomainRegExp` constant

#### Accelerator Constants
- [ ] `ProjectCreationFeeInZnn` constant
- [ ] `ZnnProjectMaximumFunds` constant
- [ ] `QsrProjectMaximumFunds` constant
- [ ] `ZnnProjectMinimumFunds` constant
- [ ] `QsrProjectMinimumFunds` constant
- [ ] `ProjectDescriptionMaxLength` constant
- [ ] `ProjectNameMaxLength` constant
- [ ] `ProjectVotingStatus` constant
- [ ] `ProjectActiveStatus` constant
- [ ] `ProjectPaidStatus` constant
- [ ] `ProjectClosedStatus` constant
- [ ] `ProjectUrlRegExp` constant

#### Swap Constants
- [ ] `SwapAssetDecayTimestampStart` constant
- [ ] `SwapAssetDecayEpochsOffset` constant
- [ ] `SwapAssetDecayTickEpochs` constant
- [ ] `SwapAssetDecayTickValuePercentage` constant

#### Spork Constants
- [ ] `SporkNameMinLength` constant
- [ ] `SporkNameMaxLength` constant
- [ ] `SporkDescriptionMaxLength` constant

#### HTLC Constants
- [ ] `HtlcPreimageMinLength` constant
- [ ] `HtlcPreimageMaxLength` constant
- [ ] `HtlcPreimageDefaultLength` constant
- [ ] `HtlcHashTypeSha3` constant = 0
- [ ] `HtlcHashTypeSha256` constant = 1

#### Bridge Constants
- [ ] `BridgeMinGuardians` constant = 5
- [ ] `BridgeMaximumFee` constant = 10000

### 1.5 Embedded Validations (`embedded/validations.go`)

- [ ] `ValidateTokenName(value string)` function
  - [ ] Check length
  - [ ] Check regex pattern
  - [ ] Return error or nil
  - [ ] Unit test: Valid names
  - [ ] Unit test: Invalid names

- [ ] `ValidateTokenSymbol(value string)` function
  - [ ] Check length
  - [ ] Check regex pattern
  - [ ] Check exceptions list
  - [ ] Return error or nil
  - [ ] Unit test: Valid symbols
  - [ ] Unit test: Reserved symbols
  - [ ] Unit test: Invalid patterns

- [ ] `ValidateTokenDomain(value string)` function
  - [ ] Check length
  - [ ] Check regex pattern
  - [ ] Return error or nil
  - [ ] Unit test: Valid domains
  - [ ] Unit test: Invalid domains

- [ ] `ValidatePillarName(value string)` function
  - [ ] Check length
  - [ ] Check regex pattern
  - [ ] Return error or nil
  - [ ] Unit test: Valid pillar names
  - [ ] Unit test: Invalid pillar names

- [ ] `ValidateProjectName(value string)` function
  - [ ] Check length
  - [ ] Return error or nil
  - [ ] Unit test: Valid names
  - [ ] Unit test: Too long

- [ ] `ValidateProjectDescription(value string)` function
  - [ ] Check length
  - [ ] Return error or nil
  - [ ] Unit test: Valid descriptions
  - [ ] Unit test: Too long

---

## Phase 2: Utils Enhancement

**Priority**: HIGH
**Status**: Not Started
**Estimated**: 2-3 days

### 2.1 NOM Constants (`utils/constants.go`)

- [ ] `CoinDecimals` constant = 8
- [ ] `OneZnn` constant = 100000000
- [ ] `OneQsr` constant = 100000000
- [ ] `IntervalBetweenMomentums` constant = 10 seconds
- [ ] Unit test: Constant values

### 2.2 Bytes Utilities (`utils/bytes.go`)

- [ ] `Arraycopy(src, startPos, dest, destPos, len)` function
  - [ ] Copy array slice
  - [ ] Unit test: Normal copy
  - [ ] Unit test: Boundary conditions

- [ ] `DecodeBigInt(bytes []byte)` function
  - [ ] Convert bytes to big.Int
  - [ ] Unit test: Positive numbers
  - [ ] Unit test: Large numbers

- [ ] `EncodeBigInt(number *big.Int)` function
  - [ ] Convert big.Int to bytes
  - [ ] Unit test: Encode/decode round-trip

- [ ] `BigIntToBytes(b *big.Int, numBytes int)` function
  - [ ] Convert to fixed-size byte array
  - [ ] Unit test: Various sizes

- [ ] `BigIntToBytesSigned(b *big.Int, numBytes int)` function
  - [ ] Convert signed big.Int
  - [ ] Unit test: Negative numbers

- [ ] `BytesToBigInt(bb []byte)` function
  - [ ] Convert bytes to big.Int
  - [ ] Unit test: Round-trip

- [ ] `Merge(arrays [][]byte)` function
  - [ ] Concatenate byte arrays
  - [ ] Unit test: Multiple arrays
  - [ ] Unit test: Empty arrays

- [ ] `IntToBytes(integer int)` function
  - [ ] Convert int32 to 4 bytes
  - [ ] Unit test: Various integers

- [ ] `LongToBytes(longValue int64)` function
  - [ ] Convert int64 to 8 bytes
  - [ ] Unit test: Various longs

- [ ] `Base64ToBytes(base64Str string)` function
  - [ ] Decode base64
  - [ ] Unit test: Valid base64
  - [ ] Unit test: Invalid base64

- [ ] `BytesToBase64(bytes []byte)` function
  - [ ] Encode to base64
  - [ ] Unit test: Round-trip

- [ ] `BytesToHex(bytes []byte)` function
  - [ ] Convert to hex string
  - [ ] Unit test: Various inputs

- [ ] `LeftPadBytes(bytes []byte, size int)` function
  - [ ] Pad bytes on left with zeros
  - [ ] Unit test: Padding to 32 bytes
  - [ ] Unit test: Already correct size

### 2.3 Amount Utilities (`utils/amount.go`)

- [ ] `ExtractDecimals(amount string, decimals int)` function
  - [ ] Parse decimal string to big.Int
  - [ ] Remove decimal point
  - [ ] Unit test: "1.5" with 8 decimals
  - [ ] Unit test: "100" with 8 decimals
  - [ ] Unit test: "0.00000001" with 8 decimals

- [ ] `AddDecimals(number *big.Int, decimals int)` function
  - [ ] Convert big.Int to decimal string
  - [ ] Insert decimal point
  - [ ] Unit test: 100000000 with 8 decimals → "1.0"
  - [ ] Unit test: 1 with 8 decimals → "0.00000001"

### 2.4 Block Utilities (`utils/block.go`)

- [ ] `IsSendBlock(blockType int)` function
  - [ ] Check if block type is send
  - [ ] Unit test: Send block type
  - [ ] Unit test: Receive block type

- [ ] `IsReceiveBlock(blockType int)` function
  - [ ] Check if block type is receive
  - [ ] Unit test: Receive block type
  - [ ] Unit test: Send block type

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
**Status**: Not Started
**Estimated**: 2-3 days

### 3.1 Argon2 Wrapper (`crypto/argon2.go`)

- [ ] `Argon2IDKey(password, salt []byte, time, memory uint32, threads uint8, keyLen uint32)` function
  - [ ] Use golang.org/x/crypto/argon2
  - [ ] Return derived key
  - [ ] Unit test: Known test vector
  - [ ] Unit test: Wallet encryption parameters

### 3.2 Crypto Utilities (`crypto/crypto.go`)

- [ ] `GetPublicKey(privateKey []byte)` function
  - [ ] Derive Ed25519 public key
  - [ ] Unit test: Known keypair

- [ ] `Sign(message, privateKey, publicKey []byte)` function
  - [ ] Ed25519 signature
  - [ ] Unit test: Sign and verify

- [ ] `Verify(signature, message, publicKey []byte)` function
  - [ ] Verify Ed25519 signature
  - [ ] Unit test: Valid signature
  - [ ] Unit test: Invalid signature

- [ ] `DeriveKey(path, seed string)` function
  - [ ] BIP32 key derivation
  - [ ] Unit test: BIP44 path

- [ ] `Digest(data []byte, digestSize int)` function
  - [ ] SHA3-256 digest
  - [ ] Unit test: Known hash

- [ ] `SHA256Bytes(data []byte)` function
  - [ ] SHA256 hash
  - [ ] Unit test: Known hash

### 3.3 Ed25519 Implementation (`crypto/ed25519.go`)

**Note**: This is a complex module. Most of Go's crypto/ed25519 can be used directly.
Only implement custom BIP32 derivation if needed.

- [ ] `KeyData` struct
  - [ ] `Key` field
  - [ ] `ChainCode` field

#### BIP32 Key Derivation
- [ ] `GetMasterKeyFromSeed(seed string)` function
  - [ ] HMAC-SHA512 derivation
  - [ ] Unit test: Master key derivation

- [ ] `DerivePath(path, seed string)` function
  - [ ] Parse BIP32 path
  - [ ] Derive child keys
  - [ ] Unit test: m/44'/73404'/0'/0/0

- [ ] `getCKDPriv(data *KeyData, index uint32)` function (private)
  - [ ] Child key derivation
  - [ ] Handle hardened keys
  - [ ] Unit test: Hardened derivation
  - [ ] Unit test: Normal derivation

---

## Phase 4: Wallet System

**Priority**: HIGH
**Status**: Not Started
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

- [ ] `BaseAddressKey` constant = "baseAddress"
- [ ] `WalletTypeKey` constant = "walletType"
- [ ] `KeyStoreWalletType` constant = "keystore"

### 4.4 Derivation (`wallet/derivation.go`)

- [ ] `CoinType` constant = "73404"
- [ ] `DerivationPath` constant = "m/44'/73404'"

- [ ] `GetDerivationAccount(account int)` function
  - [ ] Return BIP44 path for account
  - [ ] Unit test: Account 0 path
  - [ ] Unit test: Account 5 path

### 4.5 Mnemonic (`wallet/mnemonic.go`)

- [ ] `GenerateMnemonic(strength int)` function
  - [ ] Use go-bip39 library
  - [ ] Generate random mnemonic
  - [ ] Unit test: 128-bit (12 words)
  - [ ] Unit test: 256-bit (24 words)

- [ ] `ValidateMnemonic(words []string)` function
  - [ ] Check word count
  - [ ] Verify checksum
  - [ ] Unit test: Valid mnemonic
  - [ ] Unit test: Invalid checksum

- [ ] `IsValidWord(word string)` function
  - [ ] Check if word in BIP39 wordlist
  - [ ] Unit test: Valid word
  - [ ] Unit test: Invalid word

### 4.6 KeyPair (`wallet/keypair.go`)

- [ ] `KeyPair` struct
  - [ ] `PrivateKey` field
  - [ ] `PublicKey` field
  - [ ] `address` field (cached)

- [ ] `NewKeyPair(privateKey, publicKey []byte, address *types.Address)` constructor
  - [ ] Initialize keypair
  - [ ] Unit test: Constructor

- [ ] `GetPrivateKey()` method
  - [ ] Return private key bytes
  - [ ] Unit test: Key retrieval

- [ ] `GetPublicKey()` method
  - [ ] Derive if not set
  - [ ] Return public key
  - [ ] Unit test: Public key derivation

- [ ] `GetAddress()` method
  - [ ] Derive if not cached
  - [ ] Return address
  - [ ] Unit test: Address derivation

- [ ] `Sign(message []byte)` method
  - [ ] Ed25519 signature
  - [ ] Unit test: Signature generation

- [ ] `SignTx(tx *nom.AccountBlock)` method
  - [ ] Sign transaction hash
  - [ ] Unit test: Transaction signing

- [ ] `Verify(signature, message []byte)` method
  - [ ] Verify Ed25519 signature
  - [ ] Unit test: Valid signature
  - [ ] Unit test: Invalid signature

- [ ] `GeneratePublicKey(privateKey []byte)` static method
  - [ ] Derive public key
  - [ ] Unit test: Public key generation

### 4.7 Encrypted File (`wallet/encryptedfile.go`)

- [ ] `EncryptedFile` struct
  - [ ] `Metadata` field (map[string]interface{})
  - [ ] `Crypto` field (*cryptoParams)
  - [ ] `Timestamp` field (int64)
  - [ ] `Version` field (int)

- [ ] `cryptoParams` struct (private)
  - [ ] `Argon2Params` field (*argon2Params)
  - [ ] `CipherData` field ([]byte)
  - [ ] `CipherName` field (string)
  - [ ] `Kdf` field (string)
  - [ ] `Nonce` field ([]byte)

- [ ] `argon2Params` struct (private)
  - [ ] `Salt` field ([]byte)

- [ ] `Encrypt(data []byte, password string, metadata map[string]interface{})` static method
  - [ ] Generate salt
  - [ ] Derive key with Argon2
  - [ ] Encrypt with AES-256-GCM
  - [ ] Return EncryptedFile
  - [ ] Unit test: Encrypt/decrypt round-trip

- [ ] `FromJson(jsonData []byte)` constructor
  - [ ] Parse JSON
  - [ ] Unit test: Valid JSON
  - [ ] Unit test: Invalid JSON

- [ ] `Decrypt(password string)` method
  - [ ] Derive key with Argon2
  - [ ] Decrypt with AES-256-GCM
  - [ ] Verify authentication tag
  - [ ] Unit test: Correct password
  - [ ] Unit test: Incorrect password

- [ ] `ToJson()` method
  - [ ] Serialize to JSON
  - [ ] Unit test: JSON output

- [ ] `ToString()` method
  - [ ] Return JSON string
  - [ ] Unit test: String representation

### 4.8 KeyStore (`wallet/keystore.go`)

- [ ] `KeyStoreDefinition` struct
  - [ ] `File` field (file path)
  - [ ] Implements WalletDefinition

- [ ] `GetWalletId()` method
  - [ ] Return file path
  - [ ] Unit test: Wallet ID

- [ ] `GetWalletName()` method
  - [ ] Return file basename
  - [ ] Unit test: Wallet name

- [ ] `KeyStore` struct
  - [ ] `Mnemonic` field (string)
  - [ ] `Entropy` field (string)
  - [ ] `Seed` field (string)
  - [ ] Implements Wallet

- [ ] `FromMnemonic(mnemonic string)` constructor
  - [ ] Validate mnemonic
  - [ ] Derive seed
  - [ ] Unit test: Valid mnemonic
  - [ ] Unit test: Invalid mnemonic

- [ ] `FromSeed(seed string)` constructor
  - [ ] Set seed directly
  - [ ] Unit test: Seed initialization

- [ ] `FromEntropy(entropy string)` constructor
  - [ ] Convert to mnemonic
  - [ ] Unit test: Entropy conversion

- [ ] `NewRandom()` static method
  - [ ] Generate random mnemonic
  - [ ] Create keystore
  - [ ] Unit test: Random generation

- [ ] `SetMnemonic(mnemonic string)` method
  - [ ] Validate and set mnemonic
  - [ ] Derive seed
  - [ ] Unit test: Set valid mnemonic

- [ ] `SetSeed(seed string)` method
  - [ ] Set seed directly
  - [ ] Unit test: Set seed

- [ ] `SetEntropy(entropy string)` method
  - [ ] Convert to mnemonic
  - [ ] Unit test: Set entropy

- [ ] `GetAccount(index int)` method
  - [ ] Derive keypair at index
  - [ ] Unit test: Account 0
  - [ ] Unit test: Account 5

- [ ] `GetKeyPair(index int)` method
  - [ ] Derive BIP44 keypair
  - [ ] Unit test: Keypair derivation

- [ ] `DeriveAddressesByRange(left, right int)` method
  - [ ] Derive address range
  - [ ] Unit test: Range 0-10

- [ ] `FindAddress(address types.Address, numOfAddresses int)` method
  - [ ] Search for address in keystore
  - [ ] Return index and keypair
  - [ ] Unit test: Address found
  - [ ] Unit test: Address not found

- [ ] `FindResponse` struct
  - [ ] `Path` field (string)
  - [ ] `Index` field (int)
  - [ ] `KeyPair` field (*KeyPair)

### 4.9 KeyStore Manager (`wallet/manager.go`)

- [ ] `KeyStoreOptions` struct
  - [ ] `DecryptionPassword` field (string)
  - [ ] Implements WalletOptions

- [ ] `KeyStoreManager` struct
  - [ ] `WalletPath` field (directory path)
  - [ ] Implements WalletManager

- [ ] `NewKeyStoreManager(walletPath string)` constructor
  - [ ] Initialize manager
  - [ ] Create directory if needed
  - [ ] Unit test: Manager creation

- [ ] `SaveKeyStore(store *KeyStore, password, name string)` method
  - [ ] Serialize keystore to JSON
  - [ ] Encrypt with password
  - [ ] Write to file
  - [ ] Unit test: Save and load

- [ ] `ReadKeyStore(password string, keyStoreFile string)` method
  - [ ] Read encrypted file
  - [ ] Decrypt with password
  - [ ] Deserialize keystore
  - [ ] Unit test: Read keystore
  - [ ] Unit test: Wrong password

- [ ] `FindKeyStore(name string)` method
  - [ ] Search for keystore by name
  - [ ] Unit test: Find existing
  - [ ] Unit test: Not found

- [ ] `ListAllKeyStores()` method
  - [ ] List all keystores in directory
  - [ ] Unit test: Empty directory
  - [ ] Unit test: Multiple keystores

- [ ] `CreateNew(passphrase, name string)` method
  - [ ] Generate random keystore
  - [ ] Save to file
  - [ ] Unit test: Create new wallet

- [ ] `CreateFromMnemonic(mnemonic, passphrase, name string)` method
  - [ ] Create from existing mnemonic
  - [ ] Save to file
  - [ ] Unit test: Import mnemonic

- [ ] `GetWalletDefinitions()` method
  - [ ] Return all wallet definitions
  - [ ] Unit test: List definitions

- [ ] `GetWallet(definition WalletDefinition, options WalletOptions)` method
  - [ ] Load wallet from definition
  - [ ] Decrypt with options
  - [ ] Unit test: Get wallet

- [ ] `SupportsWallet(definition WalletDefinition)` method
  - [ ] Check if keystore file
  - [ ] Unit test: Support check

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
