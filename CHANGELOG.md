# Changelog

Notable changes to the Zenon Go SDK are documented in this file.

## Unreleased

Follow-up to the v0.4.0 hardening, from a Codex (GPT-6) re-review of the
CodeRabbit findings. One additive API change: the `abi.ErrUnsupportedArrayShape`
sentinel. No signatures change.
Also carries the dependency and toolchain patch bumps for two new govulncheck
advisories (merged from master, PR #12); those bullets are listed last.

### Security

- ABI: array decoders bound the aggregate size of decoded dynamic elements to
  the encoded payload. Previously many element offset words could alias one
  `bytes`/`string`/nested-array tail, so roughly 32N+L input bytes decoded to
  N×L bytes, and nested aliasing of empty values allocated N² result slots
  (CWE-400). Every element, fixed or dynamic, is charged at least one word,
  and bytes/strings longer than a word their length word plus content, which
  never exceeds what a canonical encoding occupies, so the check only rejects
  overlapping elements. Fixed elements are charged because a static array of
  dynamic elements is encoded inline with a fixed size of zero (matching
  go-zenon's ABI), so the fixed-element loop re-decodes one offset for every
  element. Applies to `StaticArrayType.DecodeTuple` and
  `DynamicArrayType.DecodeTuple`; `StaticArrayType.Decode` now always goes
  through `DecodeTuple`. Per-tuple budgets bound the work at each nesting
  level (at most `MaxArrayNesting` levels), not one global budget.
- ABI: an array holding more than one static array of dynamic elements (for
  example `bytes[][1][]` or `string[][2]`, outer dimension first) now fails with the new
  `abi.ErrUnsupportedArrayShape` before any element is decoded. The wrapper is
  encoded inline with a fixed size of zero (as in go-zenon's ABI), so the
  array has no well-defined stride; previously every element decoded the
  same offset and returned duplicated values. A single wrapper, or a wrapper
  used directly (`string[2]`), is unaffected. No embedded contract uses this
  shape.
- `rpc_client`: an oversized non-2xx HTTP response body now fails with
  `ErrResponseTooLarge` and releases the connection immediately. go-zenon's
  HTTP codec buffers error bodies in full, discards the read error, and does
  not close the body on that path, so the limit failure previously surfaced
  as a generic HTTP error while the connection lingered until `HTTPTimeout`.
  Redirects that `net/http` will follow (301/302/303; 307/308 only when the
  chain's original request has a replayable body, which go-zenon's POSTs
  lack) are left to it, since it closes their body after draining at most a
  couple of KiB rather than reading it in full;
  every other non-2xx response, including an unfollowed 3xx anywhere in a
  redirect chain, is buffered under the limit. The body limiter
  also closes the underlying body when it reports the limit.
- SECURITY.md documents the two trust assumptions behind accepted residual
  risk: the node URL is trusted unless `ClientOptions.DialPolicy` is set, and
  the wallet directory ancestry must be private to the owning user.

- `github.com/gorilla/websocket` v1.5.0 -> v1.5.3, the latest release, which
  clears the affected version range of GO-2026-6278 (GHSA-w67g-5rqw-f597,
  "cryptographically weak PRNG for the WebSocket mask key"). Note that v1.5.3
  still derives the client mask key from `math/rand.Uint32` exactly as v1.5.0
  did; the bump satisfies the advisory, it does not change the mask generator.
  On Go 1.20+ the global `math/rand` source is auto-seeded (ChaCha8 since Go
  1.22), and the mask key is a framing detail rather than a secret, so the
  practical exposure for this SDK is low. Every `rpc_client` WebSocket path was
  reported reachable.
- `toolchain go1.26.10` (was go1.26.7) raises this repository's suggested
  toolchain to a Go release containing the GO-2026-6809 (CVE-2026-111943) fix.
  Under `GOTOOLCHAIN=auto` an installed Go older than 1.26.10 switches to it; a
  newer installed Go (1.27.x) keeps running itself, so the gate also depends on
  that release being patched (1.27.3+). The fix: `crypto/x509` now distrusts a
  set of compromised intermediate certificates in `Certificate.Verify`, which
  the TLS dial path reaches. The `go 1.26.0` floor is unchanged. A dependency's `toolchain`
  directive does not apply to the applications that import it, so consumers
  must build with a patched Go (1.26.10+ or 1.27.3+) themselves.
- CI: `actions/setup-go` steps set `check-latest: true` so the `1.26.x` jobs
  resolve to the newest 1.26 patch release instead of the runner's cached
  toolchain, which is what let the x509 advisory surface as a gate failure.

### Fixed

- `rpc_client`: `MaxHTTPResponseBytes = math.MaxInt64` overflowed the body
  limiter's read-size arithmetic and panicked with a negative slice bound.
- Tests: `TestSubscriptionReadLimitEnforced` could fail spuriously when
  subscription cleanup closed the events channel before the test's select
  observed the terminal error; `crypto.Digest` now has an explicit test at
  `MaxDigestSize` and `MaxDigestSize+1`.

## v0.5.0 - 2026-08-22

Toolchain floor bump. No API changes.

### Breaking Changes

- Minimum Go version is now 1.26 (`go 1.26.0`, `toolchain go1.26.7`). Go 1.24
  and 1.25 are end-of-life and carry reachable standard-library advisories in
  the TLS/x509/net/http paths the RPC client uses. Consumers on an older
  toolchain with the default `GOTOOLCHAIN=auto` will download 1.26
  transparently.

### Changed

- `crypto.Digest` and the PoW hasher use the standard library `crypto/sha3`
  for SHA3-256 instead of the `golang.org/x/crypto/sha3` forwarding wrapper.
  Output is identical.
- `github.com/gorilla/websocket` is declared as a direct dependency (it was
  already imported directly by `rpc_client`).
- CI: test and govulncheck matrices run on the go.mod minimum (1.26.x) and the
  current stable release, with vulnerability advisories fully enforced on both;
  golangci-lint upgraded to v2.13.1.

## v0.4.0 - 2026-08-22

Security hardening addressing 20 CodeRabbit findings (9 High, 9 Medium, 2 Low).
Every code fix ships with a regression test. **Five exported signatures change —
see Breaking Changes.**

### Breaking Changes

- `crypto.DeriveKey` and `crypto.DeriveKeyDefault` now return `([]byte, error)`.
  Argon2 parameters are validated with the new `crypto.ValidateArgon2Parameters`
  before any work is performed; zero iterations/parallelism (which panicked
  inside Argon2), oversized memory (> 256 MiB), iterations (> 16),
  parallelism, key-length, short-salt, or a memory×iterations product above
  `crypto.MaxArgon2WorkKiB` (1 GiB-pass) return
  `crypto.ErrInvalidArgon2Parameters`.
- `pow.GeneratePoW`, `pow.GeneratePowBigInt`, and `pow.GeneratePowBytes` now
  return an error instead of panicking when the difficulty exceeds
  `pow.MaxReasonableDifficulty`. They are equivalent to the `*WithContext`
  variants with a background context. All `*BigInt*` generators return the new
  `pow.ErrInvalidDifficulty` for a nil or negative difficulty instead of
  panicking.

### Security

- `crypto.Digest` returns nil for `digestSize > crypto.MaxDigestSize` (1024)
  instead of allocating a caller-sized buffer.
- `pow.GeneratePowAsync`/`GeneratePowBigIntAsync` apply bounded admission: at
  most `maxWorkers × pow.DefaultPendingPoWMultiplier` requests may be active or
  queued; further requests receive `pow.ErrPoolOverloaded` immediately instead
  of spawning an unbounded number of goroutines.
- New `pow.GeneratePowBytesWithContext` provides a cancellable form of
  `GeneratePowBytes`.
- `zenon.Zenon` gains `SendWithContext`/`PrepareBlockWithContext`, a
  `MaxDifficulty` work budget, and a `PowTimeout` applied by the context-free
  `Send`/`PrepareBlock` (`zenon.DefaultPowTimeout` = 5 min; negative disables).
  PoW runs through the cancellable, worker-limited pool path; a node
  requesting difficulty above the budget (or `pow.MaxProtocolDifficulty`)
  fails before any nonce search starts.
- `rpc_client.ClientOptions` gains `DialPolicy` (applied to every resolved
  destination for initial dials, reconnects, subscription sockets, and HTTP
  redirects; environment proxies are disabled while a policy is active so a
  proxy cannot reach the destination on the client's behalf;
  `rpc_client.RejectNonPublicDestinations` is provided for untrusted URLs), `MaxHTTPResponseBytes` (default 16 MiB, enforced on success
  and error bodies), `HTTPTimeout` (default 60s), and
  `MaxSubscriptionMessageBytes` (default 15 MiB, applied via `SetReadLimit`
  before the first read on every subscription socket).
- `wallet.KeyStoreManager.SaveKeyStore` writes through an exclusive temporary
  file and `rename(2)`, so a symlink planted between the path check and the
  write is replaced rather than followed (CWE-367).
- `wallet.KeyStore.FindAddress` rejects `maxAccounts > DefaultMaxIndex`.
- `utils.ExtractDecimals` returns an error and `utils.AddDecimals` returns the
  plain integer string for `decimals > utils.MaxDecimals` (32).
- ABI: every fixed-width decoder obtains its word through one overflow-safe
  bounds check (negative or overflowing offsets return errors); `GetType`
  enforces `MaxTypeNameLength`, `MaxArrayNesting`, and `MaxStaticArraySize`;
  array `Decode`/`DecodeTuple` verify the element count fits the encoded head
  before allocating.
- `conformance/modelwire.RoundTrip` rejects inputs above `MaxInputSize`
  (256 KiB) or `MaxInputNodes` (4096 objects/arrays), rejects expanded
  encodings above `MaxEncodedSize` (4 MiB), and converts panics from dependency decoders (e.g. `BalanceInfo` with a
  missing token, `AccountBlockList` with a null element) into errors.
- CI: scanner failures are no longer suppressed. `govulncheck` fails on any
  reachable advisory not listed in `.govulncheck-ignore`
  (`scripts/govulncheck-gate.sh`, which also fails on scanner crashes, package
  load errors, or empty output); `gosec` runs without `-no-fail`. Workflows
  use least-privilege job permissions (`security-events: write` only on the
  SARIF-upload job), `persist-credentials: false`, and pin every third-party
  action and tool to an immutable commit/version. The govulncheck gate runs on
  both the go.mod minimum toolchain (dependency advisories enforced, stdlib
  advisories reported) and the current stable release (fully enforced).

### Fixed

- `pow.GeneratePowAsync`/`GeneratePowBigIntAsync` bind each request to the
  worker pool it was admitted to, and `pow.SetMaxPoWWorkers` is now safe to
  call concurrently with generation; previously a resize mid-request could
  pair a slot reserved in one pool with a release on another (deadlock and
  slot leak).
- `zenon.Zenon.Send`/`PrepareBlock` leave the template untouched when PoW is
  cancelled, overloaded, or fails, so the same block can be retried; previously
  a non-zero difficulty with an empty nonce was left behind and rejected on
  retry.
- README and package documentation examples check the errors returned by the
  changed APIs, compute the PoW data hash with `utils.GetPoWData`, and convert
  the hex nonce correctly for `pow.CheckPoW`.

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
