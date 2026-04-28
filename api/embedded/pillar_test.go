package embedded

import (
	"bytes"
	"testing"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

// TestPillarApi_Revoke_EncodesName guards the fix for the Pillar.Revoke
// signature mismatch with the Dart reference SDK.
//
// The on-chain ABI defines Revoke as: {"name":"name","type":"string"}
// (see go-zenon/vm/embedded/definition/pillars.go:47). Calling Revoke
// without the name argument produces malformed call data the contract
// will reject. Dart passes the name (pillar.dart:143-146) — Go must too.
func TestPillarApi_Revoke_EncodesName(t *testing.T) {
	api := NewPillarApi(nil)

	const pillarName = "MyPillar"
	block := api.Revoke(pillarName)

	if block == nil {
		t.Fatal("Revoke returned nil")
	}
	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}
	if block.ToAddress != types.PillarContract {
		t.Errorf("ToAddress = %s, want PillarContract", block.ToAddress.String())
	}
	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard.String())
	}
	if block.Amount == nil || block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}

	// The encoded data must equal what we'd get by directly packing the
	// Revoke method with the same name argument. Anything else means
	// either the wrong method was called or the argument was dropped.
	expected := definition.ABIPillars.PackMethodPanic(definition.RevokeMethodName, pillarName)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch\n  got:  %x\n  want: %x", block.Data, expected)
	}

	// Decoding the encoded data with the same ABI should round-trip the name
	// argument back, confirming the contract will see what we intended.
	var decoded struct {
		Name string `abi:"name"`
	}
	if err := definition.ABIPillars.UnpackMethod(&decoded, definition.RevokeMethodName, block.Data); err != nil {
		t.Fatalf("encoded Revoke call failed to decode against the ABI: %v", err)
	}
	if decoded.Name != pillarName {
		t.Errorf("decoded name = %q, want %q", decoded.Name, pillarName)
	}
}

func TestPillarApi_Revoke_DifferentNamesProduceDifferentEncodings(t *testing.T) {
	api := NewPillarApi(nil)
	a := api.Revoke("alpha")
	b := api.Revoke("beta")
	if bytes.Equal(a.Data, b.Data) {
		t.Error("Revoke encodings should differ when the name differs")
	}
}
