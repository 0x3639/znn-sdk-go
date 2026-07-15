# Changelog

Notable changes to the Zenon Go SDK are documented in this file.

## Unreleased

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
