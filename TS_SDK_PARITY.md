# TypeScript SDK Parity Audit

Comparison of the Go SDK (`znn-sdk-go`) against the TypeScript SDK
[`digitalSloth/znn-typescript-sdk`](https://github.com/digitalSloth/znn-typescript-sdk)
(v1.0.3), checked into `reference/znn-typescript-sdk-main/`.

Where the two SDKs disagree, the canonical reference is the official Dart SDK
(`reference/znn_sdk_dart-master/`) and the `go-zenon` node ABI/RPC surface — so
several TS "features" that call non-existent node methods are TS bugs, not Go gaps.

## Method coverage summary

Every embedded contract and core API method present in the TS SDK has a Go
equivalent. Verified contract-by-contract:

| Area | RPC methods | Contract calls | Status |
|------|-------------|----------------|--------|
| pillar | 11 | 9 | ✅ full parity |
| sentinel | 5 | 5 | ✅ full parity |
| stake | 3 | 3 | ✅ full parity |
| plasma | 3 (+2 Go helpers) | 2 | ⚠️ see #2 |
| token | 3 | 4 | ✅ full parity |
| accelerator | 5 | 6 | ✅ full parity |
| bridge | 15 | 20 | ⚠️ see #1 |
| liquidity | 6 | 11 | ✅ full parity |
| swap | 3 | 1 | ✅ full parity |
| spork | 1 | 2 | ✅ full parity |
| htlc | 2 | 7 | ✅ full parity |
| ledger / stats / subscribe | — | — | ✅ (stats.extraData is a TS-only dead RPC) |

ABI method names and contract addresses match in every case. Several Go function
**names** differ cosmetically from TS (e.g. `AddNetwork` vs `setNetwork`,
`VoteByProducerAddress` vs `voteByProdAddress`, `CancelLiquidity` vs
`cancelLiquidityStake`) but all pack the identical ABI method name, so wire
behavior is unchanged.

## Confirmed Go-side issues — all RESOLVED on this branch

### 0. ⛔️ CRITICAL (newly discovered): `pow/pow.go` was incompatible with go-zenon — FIXED
While wiring the send flow we found the pure-Go PoW implementation did not match the
node's algorithm. Proven empirically: a nonce the SDK generated was **rejected** by
go-zenon's own `pow.CheckPoWNonce`. The differences:

| | SDK (before) | go-zenon (canonical) |
|---|---|---|
| preimage order | `dataHash ‖ nonce` | `nonce ‖ dataHash` |
| nonce endianness | big-endian | little-endian |
| hash compared | full 32 bytes, big-endian | first 8 bytes, little-endian |
| accept when | `hash·diff < 2^64` (reversed) | `hash ≥ 2^64 − 2^64/diff` |

This was latent because there was no send flow exercising it and fused-plasma
transactions skip PoW — but any PoW transaction would have been rejected on-chain.
**Fixed**: `pow/pow.go` now matches the canonical algorithm byte-for-byte (public API
unchanged). The broken internal unit tests were replaced, and the golden test
`pow.TestPoWAcceptedByNode` verifies generated nonces against the node's own
`CheckPoWNonce`.

### 1. Bridge `GetAllWrapTokenRequestsByToAddressNetworkClassAndChainId` dropped `chainId` — FIXED
`api/embedded/bridge.go:87` omitted the `chainId` argument (sent 4 params; the node
and TS/Dart expect 5: `toAddress, networkClass, chainId, pageIndex, pageSize`), so the
node received `pageIndex` where it expected `chainId`. The parameter was added.

### 2. Plasma `GetRequiredFusionAmount` called a non-existent node RPC — FIXED
It called `embedded.plasma.getRequiredFusionAmount`, which does not exist in
`go-zenon@v0.0.8-alphanet` (the node exposes only `Get`, `GetEntriesByAddress`,
`GetRequiredPoWForAccountBlock`). The TS SDK had already removed it. The dead method
was removed; the pure-local helper `GetPlasmaByQsr` was kept.

### 3. No high-level transaction send flow — ADDED
The TS/Dart SDKs provide `Zenon.send()` / `prepareBlock()`. The Go SDK had the
building blocks but nothing tying them together, and `CLAUDE.md` referenced a
non-existent `zenon/` package. Added the `zenon` package (`zenon/zenon.go`,
`zenon/utils.go`): `zenon.NewZenon(client).Send(template, keyPair)` runs
autofill → PoW → sign → publish, with `PrepareBlock` (no publish) and `RequiresPoW`.
Ports `znn_sdk_dart/lib/src/utils/block.dart`; `CLAUDE.md` updated to match.

## Discrepancies that are NOT Go bugs (verified)

- **BIP44 derivation path** — *false alarm.* Both derive `m/44'/73404'/account'`
  (`wallet/derivation.go:23` vs `wallet/derivation.ts:5`). Misleading `/0'/0'`
  comments in the Go source caused the confusion; the code is correct and
  cross-SDK wallet-compatible. (Keystore: identical Argon2id + AES-256-GCM, AAD
  "zenon".)
- **Bridge `UnwrapToken` "parameter reordering"** — *false alarm.* The Go function
  signature orders parameters differently, but the ABI pack order
  (`networkClass, chainId, txHash, logIndex, toAddress, tokenAddress, amount,
  signature`) is identical to TS. Wire-correct.
- **Accelerator project creation fee** — TS uses `10`; Go uses `1 * OneZnn`. Go
  matches the canonical Dart reference (`projectCreationFeeInZnn = 1 * oneZnn`).
  **The TS value is the bug.**
- **`stats.extraData`** — present in TS, absent in Go. `go-zenon`'s StatsApi has no
  such method (only OsInfo/ProcessInfo/NetworkInfo/SyncInfo), so the TS method is a
  dead RPC. Go is correct to omit it.
- **`liquidity.setAdditionalReward`** — TS types args as `number`, Go as `*big.Int`.
  Go is the more correct representation for uint256 token amounts.
- **Swap return types** — TS wraps results in list classes; Go returns bare
  maps/slices. Same data, idiomatic difference.

## Go-only enhancements (no TS equivalent)

`LedgerApi.PublishRawTransactionWithRetry` (exponential backoff), `SubscriptionManager`,
connection health checks + lifecycle callbacks, PoW worker pool / context cancellation /
difficulty capping, secure key zeroing, and additional wallet helpers
(`FindAddress`, range derivation).

## Methodology

Coverage was mapped by five parallel readers (one per domain), then every
"critical"/"needs-verification" finding was confirmed by hand against the Go source,
the TS source, the Dart reference, and the `go-zenon` dependency. The five-agent pass
produced several false positives (items #BIP44 and #UnwrapToken above) that direct
verification overturned.
