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
