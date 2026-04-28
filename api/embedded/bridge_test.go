package embedded

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

func TestBridgeApi_ProposeAdministrator(t *testing.T) {
	api := NewBridgeApi(nil)
	addr := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")

	block := api.ProposeAdministrator(addr)
	if block == nil {
		t.Fatal("ProposeAdministrator returned nil")
	}
	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}
	if block.ToAddress != types.BridgeContract {
		t.Errorf("ToAddress = %s, want BridgeContract", block.ToAddress.String())
	}
	expected := definition.ABIBridge.PackMethodPanic(definition.ProposeAdministratorMethodName, addr)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch\n  got:  %x\n  want: %x", block.Data, expected)
	}
}

func TestBridgeApi_RevokeUnwrapRequest(t *testing.T) {
	api := NewBridgeApi(nil)
	txHash := types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	const logIndex uint32 = 7

	block := api.RevokeUnwrapRequest(txHash, logIndex)
	if block == nil {
		t.Fatal("RevokeUnwrapRequest returned nil")
	}
	if block.ToAddress != types.BridgeContract {
		t.Errorf("ToAddress = %s, want BridgeContract", block.ToAddress.String())
	}
	expected := definition.ABIBridge.PackMethodPanic(definition.RevokeUnwrapRequestMethodName, txHash, logIndex)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch\n  got:  %x\n  want: %x", block.Data, expected)
	}
}

// TestZtsFeesInfo_Unmarshal verifies the response shape matches what
// embedded.bridge.getFeeTokenPair returns: a JSON object with a string
// tokenStandard and a string accumulatedFee (BigInt-as-string convention).
func TestZtsFeesInfo_Unmarshal(t *testing.T) {
	raw := []byte(`{"tokenStandard":"zts1znnxxxxxxxxxxxxx9z4ulx","accumulatedFee":"123456789012345"}`)
	var z ZtsFeesInfo
	if err := json.Unmarshal(raw, &z); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if z.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("tokenStandard = %s, want ZNN", z.TokenStandard.String())
	}
	if z.AccumulatedFee == nil || z.AccumulatedFee.String() != "123456789012345" {
		t.Errorf("accumulatedFee = %v, want 123456789012345", z.AccumulatedFee)
	}
}
