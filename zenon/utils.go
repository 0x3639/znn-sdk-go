package zenon

import (
	"fmt"
	"math/big"

	"github.com/0x3639/znn-sdk-go/api/embedded"
	"github.com/0x3639/znn-sdk-go/pow"
	"github.com/0x3639/znn-sdk-go/utils"
	"github.com/0x3639/znn-sdk-go/wallet"
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	gozenonpow "github.com/zenon-network/go-zenon/pow"
)

// defaultBlockVersion is the only account-block version go-zenon accepts. The
// node rejects blocks with version 0 (ErrABVersionMissing) or any version other
// than 1 (ErrABVersionInvalid), so templates must carry it before hashing.
const defaultBlockVersion = 1

// normalizeBlockDefaults fills in the protocol defaults that a raw template may
// leave unset, so that the hashed-and-signed block matches what go-zenon expects.
//
// It mirrors the TypeScript SDK's AccountBlockTemplate defaults
// (reference/znn-typescript-sdk-main/src/model/nom/accountBlock.ts):
//   - Version defaults to 1 when unset (caller-supplied versions are preserved).
//   - A nil Amount becomes a non-nil zero, so serialization never emits "<nil>".
//   - Receive blocks have their routing fields (ToAddress, TokenStandard) zeroed,
//     matching the node's receive-block verification, which requires a zero
//     amount, zero token standard, and zero to-address.
//
// ChainIdentifier is not set here because it depends on node state; it is filled
// in by autofillTransactionParameters from the frontier momentum.
func normalizeBlockDefaults(transaction *nom.AccountBlock) {
	if transaction.Version == 0 {
		transaction.Version = defaultBlockVersion
	}
	if transaction.Amount == nil {
		transaction.Amount = big.NewInt(0)
	}
	if !transaction.IsSendBlock() {
		transaction.ToAddress = types.ZeroAddress
		transaction.TokenStandard = types.ZeroTokenStandard
		transaction.Amount = big.NewInt(0)
	}
}

// checkAndSetFields populates the signing identity and chain-position fields of a
// transaction and validates receive blocks.
//
// It sets Address and PublicKey from the keypair, autofills height/previousHash/
// momentumAcknowledged, and for receive blocks verifies that the referenced send
// block exists, targets this address, and that no data is attached.
//
// Reference: znn_sdk_dart/lib/src/utils/block.dart:_checkAndSetFields
func (z *Zenon) checkAndSetFields(transaction *nom.AccountBlock, keyPair *wallet.KeyPair) error {
	address, err := keyPair.GetAddress()
	if err != nil {
		return fmt.Errorf("failed to derive address: %w", err)
	}
	publicKey, err := keyPair.GetPublicKey()
	if err != nil {
		return fmt.Errorf("failed to derive public key: %w", err)
	}

	transaction.Address = *address
	transaction.PublicKey = publicKey

	normalizeBlockDefaults(transaction)

	if err := z.autofillTransactionParameters(transaction); err != nil {
		return err
	}

	if !transaction.IsSendBlock() {
		if transaction.FromBlockHash == types.ZeroHash {
			return fmt.Errorf("receive block requires a non-empty fromBlockHash")
		}

		sendBlock, err := z.client.LedgerApi.GetAccountBlockByHash(transaction.FromBlockHash)
		if err != nil {
			return fmt.Errorf("failed to fetch source send block %s: %w", transaction.FromBlockHash, err)
		}
		if sendBlock == nil {
			return fmt.Errorf("source send block %s not found", transaction.FromBlockHash)
		}
		if sendBlock.ToAddress.String() != transaction.Address.String() {
			return fmt.Errorf("source send block recipient %s does not match account %s",
				sendBlock.ToAddress, transaction.Address)
		}
		if len(transaction.Data) != 0 {
			return fmt.Errorf("receive block must not carry data")
		}
	}

	if transaction.Difficulty > 0 && transaction.Nonce.Data == ([8]byte{}) {
		return fmt.Errorf("transaction has difficulty %d but no nonce", transaction.Difficulty)
	}

	return nil
}

// autofillTransactionParameters fills in the chain-position fields of a
// transaction from current node state.
//
// Height and PreviousHash come from the sender's frontier account block (height 1
// and the zero hash for a brand-new account). MomentumAcknowledged and
// ChainIdentifier come from the node's frontier momentum. ChainIdentifier is only
// set when the caller left it unset (zero), so an explicit chain ID is preserved;
// go-zenon rejects blocks whose chain identifier is zero or does not match the
// node, so it must be populated before hashing.
//
// Reference: znn_sdk_dart/lib/src/utils/block.dart:_autofillTransactionParameters
func (z *Zenon) autofillTransactionParameters(transaction *nom.AccountBlock) error {
	frontier, err := z.client.LedgerApi.GetFrontierAccountBlock(transaction.Address)
	if err != nil {
		return fmt.Errorf("failed to get frontier account block: %w", err)
	}

	height := uint64(1)
	previousHash := types.ZeroHash
	if frontier != nil && frontier.Height != 0 {
		height = frontier.Height + 1
		previousHash = frontier.Hash
	}
	transaction.Height = height
	transaction.PreviousHash = previousHash

	momentum, err := z.client.LedgerApi.GetFrontierMomentum()
	if err != nil {
		return fmt.Errorf("failed to get frontier momentum: %w", err)
	}
	if momentum == nil || momentum.Momentum == nil {
		return fmt.Errorf("frontier momentum unavailable")
	}
	transaction.MomentumAcknowledged = types.HashHeight{
		Hash:   momentum.Hash,
		Height: momentum.Height,
	}

	if transaction.ChainIdentifier == 0 {
		transaction.ChainIdentifier = momentum.ChainIdentifier
	}

	return nil
}

// requiredPoW asks the node how much Proof-of-Work, if any, the transaction needs.
func (z *Zenon) requiredPoW(transaction *nom.AccountBlock) (*embedded.GetRequiredResult, error) {
	param := embedded.GetRequiredParam{
		Address:   transaction.Address,
		BlockType: transaction.BlockType,
		ToAddress: transaction.ToAddress,
		Data:      transaction.Data,
	}
	return z.client.PlasmaApi.GetRequiredPoWForAccountBlock(param)
}

// setDifficulty resolves the transaction's plasma/PoW requirement and, when PoW
// is required, generates a node-compatible nonce.
//
// When the node reports a required difficulty, the available plasma and difficulty
// are recorded and a nonce is generated over the canonical PoW data hash
// (SHA3-256(address || previousHash)). Otherwise the transaction proceeds on fused
// plasma alone with a zero difficulty and nonce.
//
// Reference: znn_sdk_dart/lib/src/utils/block.dart:_setDifficulty
func (z *Zenon) setDifficulty(transaction *nom.AccountBlock) error {
	resp, err := z.requiredPoW(transaction)
	if err != nil {
		return fmt.Errorf("failed to query required PoW: %w", err)
	}

	if resp.RequiredDifficulty != 0 {
		// Guard against a malformed or hostile node response: pow.GeneratePowBytes
		// panics when the difficulty exceeds the safety cap. Surface it as an error
		// so Send/PrepareBlock fail cleanly instead of crashing the process.
		if resp.RequiredDifficulty > pow.MaxReasonableDifficulty {
			return fmt.Errorf("node requested PoW difficulty %d above the maximum supported %d",
				resp.RequiredDifficulty, pow.MaxReasonableDifficulty)
		}

		transaction.FusedPlasma = resp.AvailablePlasma
		transaction.Difficulty = resp.RequiredDifficulty

		if z.PowCallback != nil {
			z.PowCallback(pow.Generating)
		}

		// Use go-zenon's canonical data hash so the generated nonce is guaranteed
		// to satisfy the node's pow.CheckPoWNonce.
		dataHash := gozenonpow.GetAccountBlockHash(transaction)
		nonceBytes := pow.GeneratePowBytes(dataHash, transaction.Difficulty)
		copy(transaction.Nonce.Data[:], nonceBytes)

		if z.PowCallback != nil {
			z.PowCallback(pow.Done)
		}
	} else {
		transaction.FusedPlasma = resp.BasePlasma
		transaction.Difficulty = 0
		transaction.Nonce = nom.Nonce{}
	}

	return nil
}

// setHashAndSignature computes the transaction hash and signs it with the keypair.
//
// The signature is an ed25519 signature over the 32-byte transaction hash, matching
// go-zenon's verification and the Dart/TypeScript SDKs.
//
// Reference: znn_sdk_dart/lib/src/utils/block.dart:_setHashAndSignature
func (z *Zenon) setHashAndSignature(transaction *nom.AccountBlock, keyPair *wallet.KeyPair) error {
	transaction.Hash = utils.GetTransactionHash(transaction)

	signature, err := keyPair.Sign(transaction.Hash.Bytes())
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}
	transaction.Signature = signature

	return nil
}
