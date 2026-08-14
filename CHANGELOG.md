# Changelog

Notable changes to the Zenon Go SDK are documented in this file.

## Unreleased

## v0.3.1 - 2026-08-14

### Fixed

- `api.SubscriberApi` subscription methods (`ToMomentums`,
  `ToAllAccountBlocks`, `ToAccountBlocksByAddress`,
  `ToUnreceivedAccountBlocksByAddress`) no longer panic with a nil pointer
  dereference when the underlying WebSocket client is absent. `RpcClient.Stop()`
  detaches the client via `SetClient(nil)` so post-Stop calls fail cleanly, but
  the subscribe methods dereferenced the nil client — a subscribe racing
  `Stop()` (e.g. a wallet auto-receive racing a node disconnect) crashed the
  process. They now return the new exported sentinel `api.ErrNotConnected`,
  restoring the pre-v0.3.0 error-not-panic behavior. Regression test included.

## v0.3.0 - 2026-08-14

Security hardening release remediating 36 findings from an external code review
(2 High, 9 Medium, 25 Low) plus seven follow-up issues raised in review. Every
fix ships with a regression test. **This release changes five exported function
signatures — see Breaking Changes before upgrading.**

### Breaking Changes

- `utils.GetTransactionBytes` and `utils.GetTransactionHash` now return an
  `error` alongside their result. They reject block amounts that are negative
  or do not fit in 32 unsigned bytes, which previously produced colliding,
  non-injective transaction hashes in the signing preimage.
- `utils.BigIntToBytes` and `utils.BigIntToBytesSigned` now return
  `([]byte, error)`. They reject negative or over-wide values (rather than
  silently truncating or aliasing them) and out-of-range output widths.
- `pow.BenchmarkPoW` now returns `(nonce string, iterations uint64, err error)`
  and validates the requested difficulty before searching, instead of spinning
  a core uncancellably on an out-of-range value.

### Security

- ABI decoding validates attacker-controlled offset and length words in
  `DecodeList`, `BytesType.Decode`, and the array decoders with overflow-safe,
  injective bounds checks, so hostile calldata returns an error instead of
  panicking or triggering an unbounded allocation.
- Wallet keystore names are confined to the managed directory: names containing
  path separators or `..` are rejected, and symlinks / non-regular files are
  refused by save, read, metadata, and listing operations (path-traversal and
  symlink-traversal hardening).
- Argon2 key-derivation costs read from an imported keystore are capped
  (256 MiB memory, 10 iterations, 8 lanes) to prevent resource-exhaustion on
  import.
- Password validation counts characters rather than bytes and correctly
  detects all-same-character passwords in multi-byte UTF-8.
- BIP32 derivation rejects unhardened path indices at or above 2^31 instead of
  aliasing them to hardened indices.
- `KeyPair` is safe for concurrent use: lazy derivation, signing, verification,
  and `Destroy` are serialized, and the accessors return defensive copies.
- `KeyStore.DeriveAddressesByRange` bounds the requested span.
- Embedded API decoders reject null list, map, and required-nested entries from
  a hostile node response, and `SwapApi.GetAssets` now passes a pointer
  destination so swap-asset queries decode.
- The RPC client synchronizes its connection-lifecycle fields, makes `Stop()`
  final (an in-flight reconnect can no longer resurrect a stopped client),
  keeps the exported API objects stable across reconnects, and terminates a
  subscription whose handshake is cancelled mid-flight.
- Additional guards: `crypto.Digest` rejects negative sizes,
  `ValidateTokenDomain` enforces the 128-character protocol maximum,
  `utils.AddDecimals` formats negative amounts correctly and no longer panics
  on negative decimals, and `utils.Arraycopy` bounds-checks its arguments.

### Added

- `api.SubscriberApi.SetClient` swaps the underlying WebSocket client on
  reconnect while keeping the `SubscriberApi` instance stable.

### Validation

- Regression tests were added for every finding. The native test suite, race
  tests, `go vet`, `gofmt`, and golangci-lint all pass.

## v0.2.1 - 2026-07-14

This patch release corrects ABI decoding for arrays with dynamic element types
and expands regression coverage across SDK error and lifecycle paths.

### Fixed

- Static and dynamic ABI arrays now advance through each encoded head pointer
  when their elements are dynamically sized, rather than decoding every
  element from the first pointer.
- Dynamic ABI array decoding now rejects negative encoded lengths.

### Validation

- Added exhaustive error-injection and lifecycle tests for API, RPC client,
  subscription, transport, wallet, Zenon transaction, ABI, and wire-model
  behavior.
- Total Go statement coverage is 93.1%, with the native test suite, race tests,
  vet, and golangci-lint passing.

## v0.2.0 - 2026-07-14

This release brings the Go SDK into conformance with the stable Zenon SDK
specification and the pinned canonical `go-zenon` behavior.

### Added

- HTTP and HTTPS JSON-RPC support alongside the existing WebSocket transports.
- Normalized RPC errors containing the node error code, message, data, method,
  and positional parameters.
- Reconnecting normalized subscriptions that expose the subscription ID and
  update batches, clean up on disconnect, and resubscribe after reconnection.
- Wallet key-file upgrade detection through `EncryptedFile.NeedsUpgrade`.
- Executable typed model conformance plumbing covering all 72 stable wire
  models.

### Changed

- Newly written wallet key files store interoperable raw BIP39 entropy and all
  Argon2 parameters: time cost, memory cost, hash length, parallelism, and salt.
- Wallet decryption now derives account zero and verifies
  `metadata.baseAddress` before accepting a key file.
- ABI validation now enforces exact lengths for `bytes1` through `bytes32`,
  signed and unsigned bounds for every width from 8 through 256, and canonical
  boolean values on both encode and decode.
- Paginated RPC methods reject page sizes or counts above their endpoint limit
  before sending the request. Standard endpoints allow at most 1024 items;
  memory-pool and liquidity-stake endpoints allow at most 50.
- Decimal-to-base-unit conversion accepts signed values and truncates excess
  precision toward zero.

### Fixed

- Added the 11 missing concrete embedded ABI entries for Accelerator, Pillar,
  Sentinel, and Stake. All concrete catalogs now cover all 84 stable functions.
- `ledger.publishRawTransaction` now treats only a JSON `null` result as
  success.
- Empty HTLC hash locks preserve their canonical empty base64 string instead of
  re-encoding as JSON `null`.

### Compatibility

- Legacy raw-entropy key files and key files produced by earlier Go SDK
  versions remain readable through the public wallet API.
- Existing WebSocket request APIs remain supported. HTTP transports support
  request/response calls; subscriptions continue to require WebSocket or secure
  WebSocket transport.
- Applications that previously supplied out-of-range ABI values, non-boolean
  values for ABI booleans, oversized pages, or relied on a non-null publish
  response will now receive an error before invalid data is accepted.

### Validation

- Stable conformance corpus: 764/764 cases.
- Stable model fixtures: 72/72 models through typed Go decoding and encoding.
- Embedded ABI inventory: 84/84 functions.
- RPC inventory: 76/76 methods with canonical positional ordering.
- Native tests, race tests, vet, canonical-node tests, stable-spec validation,
  and the HTTP/WebSocket transport fixture all pass.
