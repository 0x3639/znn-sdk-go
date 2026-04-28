package embedded

import (
	"bytes"
	"testing"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

func TestLiquidityApi_ProposeAdministrator(t *testing.T) {
	api := NewLiquidityApi(nil)
	addr := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")

	block := api.ProposeAdministrator(addr)
	if block == nil {
		t.Fatal("ProposeAdministrator returned nil")
	}
	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}
	if block.ToAddress != types.LiquidityContract {
		t.Errorf("ToAddress = %s, want LiquidityContract", block.ToAddress.String())
	}
	expected := definition.ABILiquidity.PackMethodPanic(definition.ProposeAdministratorMethodName, addr)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch\n  got:  %x\n  want: %x", block.Data, expected)
	}
}

func TestLiquidityApi_ChangeAdministrator(t *testing.T) {
	api := NewLiquidityApi(nil)
	addr := types.ParseAddressPanic("z1qr4pexnnfaexqqz8nscjjcsajy5hdqfkgadvwx")

	block := api.ChangeAdministrator(addr)
	if block == nil {
		t.Fatal("ChangeAdministrator returned nil")
	}
	if block.ToAddress != types.LiquidityContract {
		t.Errorf("ToAddress = %s, want LiquidityContract", block.ToAddress.String())
	}
	expected := definition.ABILiquidity.PackMethodPanic(definition.ChangeAdministratorMethodName, addr)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch\n  got:  %x\n  want: %x", block.Data, expected)
	}
}

func TestLiquidityApi_Emergency(t *testing.T) {
	api := NewLiquidityApi(nil)
	block := api.Emergency()
	if block == nil {
		t.Fatal("Emergency returned nil")
	}
	if block.ToAddress != types.LiquidityContract {
		t.Errorf("ToAddress = %s, want LiquidityContract", block.ToAddress.String())
	}
	if block.Amount == nil || block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}
	expected := definition.ABILiquidity.PackMethodPanic(definition.EmergencyMethodName)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch\n  got:  %x\n  want: %x", block.Data, expected)
	}
}
