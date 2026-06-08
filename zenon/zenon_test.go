package zenon

import (
	"math/big"
	"testing"

	"github.com/0x3639/znn-sdk-go/pow"
	"github.com/0x3639/znn-sdk-go/utils"
	"github.com/0x3639/znn-sdk-go/wallet"
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	gozenonpow "github.com/zenon-network/go-zenon/pow"
)

// testMnemonic is a well-known valid BIP39 mnemonic used only for deterministic
// offline tests.
const testMnemonic = "test test test test test test test test test test test junk"

func testKeyPair(t *testing.T) *wallet.KeyPair {
	t.Helper()
	ks, err := wallet.NewKeyStoreFromMnemonic(testMnemonic)
	if err != nil {
		t.Fatalf("NewKeyStoreFromMnemonic: %v", err)
	}
	kp, err := ks.GetKeyPair(0)
	if err != nil {
		t.Fatalf("GetKeyPair: %v", err)
	}
	return kp
}

func sampleSendBlock(t *testing.T, kp *wallet.KeyPair) *nom.AccountBlock {
	t.Helper()
	addr, err := kp.GetAddress()
	if err != nil {
		t.Fatalf("GetAddress: %v", err)
	}
	pub, err := kp.GetPublicKey()
	if err != nil {
		t.Fatalf("GetPublicKey: %v", err)
	}
	return &nom.AccountBlock{
		Version:              1,
		ChainIdentifier:      1,
		BlockType:            nom.BlockTypeUserSend,
		Address:              *addr,
		ToAddress:            types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz"),
		Amount:               big.NewInt(100000000),
		TokenStandard:        types.ZnnTokenStandard,
		Height:               5,
		PreviousHash:         types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
		MomentumAcknowledged: types.HashHeight{Hash: types.ZeroHash, Height: 100},
		PublicKey:            pub,
	}
}

// TestSetHashAndSignature verifies that the signing step produces the canonical
// transaction hash and a signature the keypair can verify.
func TestSetHashAndSignature(t *testing.T) {
	kp := testKeyPair(t)
	block := sampleSendBlock(t, kp)

	z := &Zenon{} // no client needed for signing
	if err := z.setHashAndSignature(block, kp); err != nil {
		t.Fatalf("setHashAndSignature: %v", err)
	}

	wantHash := utils.GetTransactionHash(block)
	if block.Hash != wantHash {
		t.Errorf("block.Hash = %s, want %s", block.Hash, wantHash)
	}

	ok, err := kp.Verify(block.Signature, block.Hash.Bytes())
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Error("signature does not verify against the transaction hash")
	}
}

// TestNormalizeBlockDefaultsSend verifies that a send template gets a default
// protocol version and a non-nil amount while its routing fields are preserved.
func TestNormalizeBlockDefaultsSend(t *testing.T) {
	to := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	block := &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     to,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        big.NewInt(42),
		// Version intentionally left at 0 to exercise the default.
	}

	normalizeBlockDefaults(block)

	if block.Version != 1 {
		t.Errorf("Version = %d, want 1", block.Version)
	}
	if block.Amount == nil || block.Amount.Cmp(big.NewInt(42)) != 0 {
		t.Errorf("Amount = %v, want 42", block.Amount)
	}
	if block.ToAddress != to {
		t.Errorf("ToAddress = %s, want %s (send routing must be preserved)", block.ToAddress, to)
	}
	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN (send routing must be preserved)", block.TokenStandard)
	}
}

// TestNormalizeBlockDefaultsReceive verifies that a receive template gets a
// non-nil zero amount and that any stray routing fields are zeroed, matching the
// node's receive-block verification (ErrABAmountMustBeZero/ZtsMustBeZero/
// ToAddressMustBeZero) and the TypeScript SDK defaults.
func TestNormalizeBlockDefaultsReceive(t *testing.T) {
	block := &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserReceive,
		FromBlockHash: types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
		// Stray routing fields that a dirty template might carry.
		ToAddress:     types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz"),
		TokenStandard: types.ZnnTokenStandard,
		// Amount left nil to exercise normalization.
	}

	normalizeBlockDefaults(block)

	if block.Version != 1 {
		t.Errorf("Version = %d, want 1", block.Version)
	}
	if block.Amount == nil || block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want non-nil zero", block.Amount)
	}
	if block.ToAddress != types.ZeroAddress {
		t.Errorf("ToAddress = %s, want zero for receive block", block.ToAddress)
	}
	if block.TokenStandard != types.ZeroTokenStandard {
		t.Errorf("TokenStandard = %s, want zero for receive block", block.TokenStandard)
	}
}

// TestNormalizeBlockDefaultsPreservesVersion confirms a caller-supplied version
// is not overwritten.
func TestNormalizeBlockDefaultsPreservesVersion(t *testing.T) {
	block := &nom.AccountBlock{BlockType: nom.BlockTypeUserSend, Version: 2}
	normalizeBlockDefaults(block)
	if block.Version != 2 {
		t.Errorf("Version = %d, want 2 (caller value must be preserved)", block.Version)
	}
}

// TestSendFlowNonceAcceptedByNode confirms the nonce that the send flow would
// generate for a block satisfies go-zenon's pow.CheckPoWNonce. This guards the
// integration between setDifficulty's data hash and the pow package.
func TestSendFlowNonceAcceptedByNode(t *testing.T) {
	kp := testKeyPair(t)
	block := sampleSendBlock(t, kp)
	block.Difficulty = 1000

	dataHash := gozenonpow.GetAccountBlockHash(block)
	nonce := pow.GeneratePowBytes(dataHash, block.Difficulty)
	copy(block.Nonce.Data[:], nonce)

	if !gozenonpow.CheckPoWNonce(block) {
		t.Errorf("send-flow nonce %x rejected by go-zenon CheckPoWNonce", nonce)
	}
}
