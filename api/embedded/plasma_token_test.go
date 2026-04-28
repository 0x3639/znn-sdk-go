package embedded

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/constants"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

// =============================================================================
// PlasmaApi Fuse/Cancel Tests
// =============================================================================

func TestPlasmaApi_Fuse(t *testing.T) {
	api := NewPlasmaApi(nil)
	addr := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	amount := big.NewInt(10 * 100000000)

	block := api.Fuse(addr, amount)
	if block == nil {
		t.Fatal("Fuse returned nil")
	}
	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}
	if block.ToAddress != types.PlasmaContract {
		t.Errorf("ToAddress = %s, want PlasmaContract", block.ToAddress)
	}
	if block.TokenStandard != types.QsrTokenStandard {
		t.Errorf("TokenStandard = %s, want QSR", block.TokenStandard)
	}
	if block.Amount.Cmp(amount) != 0 {
		t.Errorf("Amount = %v, want %v", block.Amount, amount)
	}
	expected := definition.ABIPlasma.PackMethodPanic(definition.FuseMethodName, addr)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for Fuse")
	}
}

func TestPlasmaApi_Cancel(t *testing.T) {
	api := NewPlasmaApi(nil)
	id := types.HexToHashPanic("bbbb000000000000000000000000000000000000000000000000000000000001")

	block := api.Cancel(id)
	if block == nil {
		t.Fatal("Cancel returned nil")
	}
	if block.ToAddress != types.PlasmaContract {
		t.Errorf("ToAddress = %s, want PlasmaContract", block.ToAddress)
	}
	if block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}
	expected := definition.ABIPlasma.PackMethodPanic(definition.CancelFuseMethodName, id)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for Cancel")
	}
}

// =============================================================================
// TokenApi Tests
// =============================================================================

func TestNewTokenApi(t *testing.T) {
	api := NewTokenApi(nil)
	if api == nil {
		t.Fatal("NewTokenApi returned nil")
	}
}

func TestTokenApi_IssueToken(t *testing.T) {
	api := NewTokenApi(nil)
	totalSupply := big.NewInt(1000000 * 100000000)
	maxSupply := big.NewInt(1000000 * 100000000)

	block := api.IssueToken("TestToken", "TTK", "test.com", totalSupply, maxSupply, 8, true, true, false)
	if block == nil {
		t.Fatal("IssueToken returned nil")
	}
	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}
	if block.ToAddress != types.TokenContract {
		t.Errorf("ToAddress = %s, want TokenContract", block.ToAddress)
	}
	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard)
	}
	if block.Amount.Cmp(constants.TokenIssueAmount) != 0 {
		t.Errorf("Amount = %v, want %v", block.Amount, constants.TokenIssueAmount)
	}
}

func TestTokenApi_Mint(t *testing.T) {
	api := NewTokenApi(nil)
	receiver := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	amount := big.NewInt(1000 * 100000000)

	block := api.Mint(types.ZnnTokenStandard, amount, receiver)
	if block == nil {
		t.Fatal("Mint returned nil")
	}
	if block.ToAddress != types.TokenContract {
		t.Errorf("ToAddress = %s, want TokenContract", block.ToAddress)
	}
	if block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0 (ZNN fee is not sent, zero amount)", block.Amount)
	}
	expected := definition.ABIToken.PackMethodPanic(definition.MintMethodName, types.ZnnTokenStandard, amount, receiver)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for Mint")
	}
}

func TestTokenApi_Burn(t *testing.T) {
	api := NewTokenApi(nil)
	amount := big.NewInt(500 * 100000000)

	block := api.Burn(types.ZnnTokenStandard, amount)
	if block == nil {
		t.Fatal("Burn returned nil")
	}
	if block.ToAddress != types.TokenContract {
		t.Errorf("ToAddress = %s, want TokenContract", block.ToAddress)
	}
	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard)
	}
	if block.Amount.Cmp(amount) != 0 {
		t.Errorf("Amount = %v, want %v", block.Amount, amount)
	}
	expected := definition.ABIToken.PackMethodPanic(definition.BurnMethodName)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for Burn")
	}
}

func TestTokenApi_UpdateToken(t *testing.T) {
	api := NewTokenApi(nil)
	owner := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")

	block := api.UpdateToken(types.ZnnTokenStandard, owner, false, true)
	if block == nil {
		t.Fatal("UpdateToken returned nil")
	}
	if block.ToAddress != types.TokenContract {
		t.Errorf("ToAddress = %s, want TokenContract", block.ToAddress)
	}
	if block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}
	expected := definition.ABIToken.PackMethodPanic(definition.UpdateTokenMethodName, types.ZnnTokenStandard, owner, false, true)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for UpdateToken")
	}
}
