package embedded

import (
	"bytes"
	"testing"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

func TestSporkApi_CreateSpork(t *testing.T) {
	api := NewSporkApi(nil)
	const name = "halt-bridge"
	const description = "Halts the bridge in case of emergency"

	block := api.CreateSpork(name, description)
	if block == nil {
		t.Fatal("CreateSpork returned nil")
	}
	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}
	if block.ToAddress != types.SporkContract {
		t.Errorf("ToAddress = %s, want SporkContract", block.ToAddress.String())
	}
	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard.String())
	}
	if block.Amount == nil || block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}

	expected := definition.ABISpork.PackMethodPanic(definition.SporkCreateMethodName, name, description)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch\n  got:  %x\n  want: %x", block.Data, expected)
	}

	var decoded struct {
		Name        string `abi:"name"`
		Description string `abi:"description"`
	}
	if err := definition.ABISpork.UnpackMethod(&decoded, definition.SporkCreateMethodName, block.Data); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if decoded.Name != name || decoded.Description != description {
		t.Errorf("decoded = %+v, want name=%q description=%q", decoded, name, description)
	}
}

func TestSporkApi_ActivateSpork(t *testing.T) {
	api := NewSporkApi(nil)
	id := types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	block := api.ActivateSpork(id)
	if block == nil {
		t.Fatal("ActivateSpork returned nil")
	}
	if block.ToAddress != types.SporkContract {
		t.Errorf("ToAddress = %s, want SporkContract", block.ToAddress.String())
	}

	expected := definition.ABISpork.PackMethodPanic(definition.SporkActivateMethodName, id)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch\n  got:  %x\n  want: %x", block.Data, expected)
	}

	var decoded struct {
		Id types.Hash `abi:"id"`
	}
	if err := definition.ABISpork.UnpackMethod(&decoded, definition.SporkActivateMethodName, block.Data); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if decoded.Id != id {
		t.Errorf("decoded id = %s, want %s", decoded.Id.String(), id.String())
	}
}
