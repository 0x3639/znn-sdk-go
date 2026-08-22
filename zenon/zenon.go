// Package zenon provides the high-level transaction send flow for the Zenon SDK.
//
// While the rpc_client and embedded APIs return unsigned *nom.AccountBlock
// templates, turning a template into a published, on-chain transaction requires
// several coordinated steps:
//
//  1. Autofill the block's height, previous hash, and acknowledged momentum
//  2. Set the signing address and public key from the keypair
//  3. Query the required Proof-of-Work difficulty (or use available plasma)
//  4. Generate the PoW nonce when difficulty is required
//  5. Compute the transaction hash and sign it
//  6. Publish the raw transaction to the node
//
// The Zenon type wraps an *rpc_client.RpcClient and performs this whole flow via
// Send (autofill -> PoW -> sign -> publish) or PrepareBlock (everything except
// publish). This mirrors the official Dart and TypeScript SDKs' Zenon.send /
// prepareBlock helpers.
//
// Basic usage:
//
//	client, _ := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
//	defer client.Stop()
//
//	z := zenon.NewZenon(client)
//
//	keyStore, _ := manager.ReadKeyStore("password", "my-wallet")
//	keyPair, _ := keyStore.GetKeyPair(0)
//
//	template := client.LedgerApi.SendTemplate(
//	    recipient, types.ZnnTokenStandard, amount, nil,
//	)
//	published, err := z.Send(template, keyPair)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Published:", published.Hash)
package zenon

import (
	"context"
	"fmt"
	"time"

	"github.com/0x3639/znn-sdk-go/pow"
	"github.com/0x3639/znn-sdk-go/rpc_client"
	"github.com/0x3639/znn-sdk-go/wallet"
	"github.com/zenon-network/go-zenon/chain/nom"
)

// Zenon coordinates the full transaction send flow against a connected node.
//
// Construct one with NewZenon. A Zenon is a thin, stateless wrapper around an
// *rpc_client.RpcClient and is safe to reuse for many transactions. It holds no
// keys; a *wallet.KeyPair is supplied per call.
type Zenon struct {
	client *rpc_client.RpcClient

	// PowCallback, when non-nil, is invoked with pow.Generating immediately
	// before a Proof-of-Work computation begins and pow.Done immediately after
	// it completes. It is never called when a transaction is covered by fused
	// plasma (no PoW required). Use it to surface progress to users, since PoW
	// generation is synchronous and can take noticeable time at high difficulty.
	PowCallback func(pow.PowStatus)

	// PowTimeout bounds Proof-of-Work generation for Send and PrepareBlock,
	// which have no caller-supplied context. Zero selects DefaultPowTimeout; a
	// negative value disables the deadline (explicit opt-in to unbounded work).
	// SendWithContext/PrepareBlockWithContext use the caller's context instead.
	PowTimeout time.Duration

	// MaxDifficulty, when non-zero, is the largest PoW difficulty this Zenon is
	// willing to compute. The node controls the required difficulty, so this is
	// an application-level work budget: Send/PrepareBlock fail with an error
	// instead of starting a nonce search when the node asks for more. When zero,
	// difficulties up to pow.MaxProtocolDifficulty are accepted (higher values
	// are always rejected).
	MaxDifficulty uint64
}

// DefaultPowTimeout is the Proof-of-Work deadline applied by Send and
// PrepareBlock when Zenon.PowTimeout is zero. Protocol-maximum difficulty
// typically completes well within it on commodity hardware, while a hostile
// node cannot pin a worker indefinitely.
const DefaultPowTimeout = 5 * time.Minute

// NewZenon creates a Zenon send-flow helper bound to the given RPC client.
//
// Parameters:
//   - client: A connected *rpc_client.RpcClient. The client must remain open for
//     the lifetime of the returned Zenon.
//
// Returns a ready-to-use *Zenon.
//
// Example:
//
//	client, _ := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
//	z := zenon.NewZenon(client)
func NewZenon(client *rpc_client.RpcClient) *Zenon {
	return &Zenon{client: client}
}

// Client returns the underlying RPC client, for callers that need direct API
// access alongside the send flow.
func (z *Zenon) Client() *rpc_client.RpcClient {
	return z.client
}

// Send finalizes and publishes a transaction template in one call.
//
// It performs, in order: field validation and autofill (height, previous hash,
// acknowledged momentum, signing address and public key), PoW/plasma resolution
// (querying the node and generating a nonce only if difficulty is required),
// hashing and signing, and finally publishing the raw transaction.
//
// Parameters:
//   - transaction: An unsigned *nom.AccountBlock template, typically returned by
//     a LedgerApi or embedded contract method. It is mutated in place.
//   - keyPair: The *wallet.KeyPair that signs the transaction. Its address
//     becomes the block's sender.
//
// Returns the fully populated, published *nom.AccountBlock (the same pointer that
// was passed in) or an error if any step fails. A nil error means the node
// accepted the raw transaction for processing; on-chain execution may still fail
// independently.
//
// Note: PoW generation is synchronous and can be slow at high difficulty. Set
// PowCallback to observe progress. PoW is bounded by PowTimeout
// (DefaultPowTimeout when zero); use SendWithContext to supply your own
// deadline. For transactions covered by fused plasma, no PoW is generated.
//
// Example:
//
//	template := client.TokenApi.IssueToken(...)
//	published, err := z.Send(template, keyPair)
func (z *Zenon) Send(transaction *nom.AccountBlock, keyPair *wallet.KeyPair) (*nom.AccountBlock, error) {
	ctx, cancel := z.powContext()
	defer cancel()
	return z.SendWithContext(ctx, transaction, keyPair)
}

// powContext returns the context Send and PrepareBlock use for PoW, applying
// PowTimeout (or DefaultPowTimeout) unless the timeout is negative.
func (z *Zenon) powContext() (context.Context, context.CancelFunc) {
	timeout := z.PowTimeout
	if timeout == 0 {
		timeout = DefaultPowTimeout
	}
	if timeout < 0 {
		return context.WithCancel(context.Background())
	}
	return context.WithTimeout(context.Background(), timeout)
}

// SendWithContext is like Send but honors ctx during Proof-of-Work generation.
//
// The node dictates the required difficulty, and a nonce search at the protocol
// maximum can run for minutes. Pass a context with a deadline or cancellation
// to bound that work; pow.ErrCancelled is returned (wrapped) when ctx is done
// before a nonce is found. PoW runs through the pow package's worker pool, so
// concurrent sends share the configured CPU limit.
//
// Parameters:
//   - ctx: Bounds PoW generation. It is not currently applied to RPC calls.
//   - transaction, keyPair: As for Send.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
//	defer cancel()
//	published, err := z.SendWithContext(ctx, template, keyPair)
func (z *Zenon) SendWithContext(ctx context.Context, transaction *nom.AccountBlock, keyPair *wallet.KeyPair) (*nom.AccountBlock, error) {
	if _, err := z.PrepareBlockWithContext(ctx, transaction, keyPair); err != nil {
		return nil, err
	}

	if err := z.client.LedgerApi.PublishRawTransaction(transaction); err != nil {
		return nil, fmt.Errorf("failed to publish transaction: %w", err)
	}

	return transaction, nil
}

// PrepareBlock runs the full send flow except the final publish step.
//
// This is useful when you need to inspect, persist, or hand off a signed
// transaction without immediately submitting it, or when you want to control the
// publish call yourself (for example, to publish through a different connection).
//
// Parameters:
//   - transaction: An unsigned *nom.AccountBlock template. It is mutated in place.
//   - keyPair: The *wallet.KeyPair that signs the transaction.
//
// Returns the populated and signed *nom.AccountBlock (the same pointer passed in)
// or an error. After a successful call the block carries a valid hash, signature,
// public key, and (if required) PoW nonce, and is ready for
// LedgerApi.PublishRawTransaction.
//
// Example:
//
//	signed, err := z.PrepareBlock(template, keyPair)
//	if err != nil {
//	    return err
//	}
//	// ... later ...
//	err = client.LedgerApi.PublishRawTransaction(signed)
func (z *Zenon) PrepareBlock(transaction *nom.AccountBlock, keyPair *wallet.KeyPair) (*nom.AccountBlock, error) {
	ctx, cancel := z.powContext()
	defer cancel()
	return z.PrepareBlockWithContext(ctx, transaction, keyPair)
}

// PrepareBlockWithContext is like PrepareBlock but honors ctx during
// Proof-of-Work generation. See SendWithContext for details.
func (z *Zenon) PrepareBlockWithContext(ctx context.Context, transaction *nom.AccountBlock, keyPair *wallet.KeyPair) (*nom.AccountBlock, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil context")
	}
	if err := z.checkAndSetFields(transaction, keyPair); err != nil {
		return nil, err
	}
	if err := z.setDifficulty(ctx, transaction); err != nil {
		return nil, err
	}
	if err := z.setHashAndSignature(transaction, keyPair); err != nil {
		return nil, err
	}
	return transaction, nil
}

// RequiresPoW reports whether a transaction would require Proof-of-Work, i.e.
// whether the sending address lacks sufficient fused plasma to cover it.
//
// This queries the node without modifying or sending anything (beyond setting the
// transaction's Address from the keypair so the query can be made). Use it to
// decide whether to warn a user about an upcoming PoW computation.
//
// Parameters:
//   - transaction: The *nom.AccountBlock template to evaluate. Its Address is set
//     from keyPair as a side effect.
//   - keyPair: The *wallet.KeyPair whose address will send the transaction.
//
// Returns true if PoW would be required, false if available plasma is sufficient,
// or an error if the node query fails.
//
// Example:
//
//	needed, err := z.RequiresPoW(template, keyPair)
//	if err == nil && needed {
//	    fmt.Println("This transaction will require Proof-of-Work.")
//	}
func (z *Zenon) RequiresPoW(transaction *nom.AccountBlock, keyPair *wallet.KeyPair) (bool, error) {
	address, err := keyPair.GetAddress()
	if err != nil {
		return false, fmt.Errorf("failed to derive address: %w", err)
	}
	transaction.Address = *address

	resp, err := z.requiredPoW(transaction)
	if err != nil {
		return false, fmt.Errorf("failed to query required PoW: %w", err)
	}
	return resp.RequiredDifficulty != 0, nil
}
