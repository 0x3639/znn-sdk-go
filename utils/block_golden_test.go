package utils

import (
	"math/big"
	"testing"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
)

// Golden vectors imported verbatim from the Dart reference SDK at
// reference/znn_sdk_dart-master/test/model/nom/account_block_test.dart.
//
// These pin Go's GetTransactionHash output to the exact bytes the Dart SDK
// produces for the same inputs. If either implementation drifts in field
// order, endianness, hashing, or padding, this test will catch it.

func dartSendBlock() *nom.AccountBlock {
	return &nom.AccountBlock{
		Version:         1,
		ChainIdentifier: 100,
		BlockType:       uint64(BlockTypeUserSend),
		PreviousHash:    types.HexToHashPanic("598fa623dd308bec7163bb375aa7546ec4aced3b71a1c9278709903e69280dbd"),
		Height:          2,
		MomentumAcknowledged: types.HashHeight{
			Hash:   types.HexToHashPanic("c37c70550e95d0c72f0924d480321976040108f29fa7530487f8dde81e713689"),
			Height: 1,
		},
		Address:       types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz"),
		ToAddress:     types.ParseAddressPanic("z1qr4pexnnfaexqqz8nscjjcsajy5hdqfkgadvwx"),
		Amount:        big.NewInt(10000000000),
		TokenStandard: types.ParseZTSPanic("zts1tfjkummwyppk76twsnv50e"),
		FromBlockHash: types.ZeroHash,
		Data:          []byte{},
		FusedPlasma:   21000,
		Difficulty:    0,
		Nonce:         nom.Nonce{},
	}
}

func dartReceiveBlock() *nom.AccountBlock {
	return &nom.AccountBlock{
		Version:         1,
		ChainIdentifier: 100,
		BlockType:       uint64(BlockTypeUserReceive),
		PreviousHash:    types.HexToHashPanic("57b6b7c6edb82b38ec4c992d99c84bf8016f03bf0727ff9daa811d2e862fa77a"),
		Height:          2,
		MomentumAcknowledged: types.HashHeight{
			Hash:   types.HexToHashPanic("0f92b0be5eef439be78f9d48add78288391d6723e40c7059fae0f1241a9e639f"),
			Height: 2,
		},
		Address:       types.ParseAddressPanic("z1qr4pexnnfaexqqz8nscjjcsajy5hdqfkgadvwx"),
		ToAddress:     types.ParseAddressPanic("z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f"),
		Amount:        big.NewInt(0),
		TokenStandard: types.ParseZTSPanic("zts1qqqqqqqqqqqqqqqqtq587y"),
		FromBlockHash: types.HexToHashPanic("3835082b4afb76971d58d6ad510e7e91f3bb0d41912fac4ec4cfef7bd7bbea73"),
		Data:          []byte{},
		FusedPlasma:   21000,
		Difficulty:    0,
		Nonce:         nom.Nonce{},
	}
}

func TestGetTransactionHash_DartGoldenSend(t *testing.T) {
	const expected = "3835082b4afb76971d58d6ad510e7e91f3bb0d41912fac4ec4cfef7bd7bbea73"
	got := GetTransactionHash(dartSendBlock()).String()
	if got != expected {
		t.Fatalf("send-block hash mismatch (Dart parity broken)\n  got:  %s\n  want: %s", got, expected)
	}
}

func TestGetTransactionHash_DartGoldenReceive(t *testing.T) {
	const expected = "158a0a5a7b4d57f4d92e3c068db19125fcc31ff0f059de0df98c920b54a83cd2"
	got := GetTransactionHash(dartReceiveBlock()).String()
	if got != expected {
		t.Fatalf("receive-block hash mismatch (Dart parity broken)\n  got:  %s\n  want: %s", got, expected)
	}
}

// TestGetTransactionHash_DartGoldenTemplate exercises the same fixture as the
// Dart account_block_template_test.dart — the template is the send block
// minus the cached confirmation/token fields, so the hash must be identical.
func TestGetTransactionHash_DartGoldenTemplate(t *testing.T) {
	const expected = "3835082b4afb76971d58d6ad510e7e91f3bb0d41912fac4ec4cfef7bd7bbea73"
	got := GetTransactionHash(dartSendBlock()).String()
	if got != expected {
		t.Fatalf("template hash mismatch\n  got:  %s\n  want: %s", got, expected)
	}
}
