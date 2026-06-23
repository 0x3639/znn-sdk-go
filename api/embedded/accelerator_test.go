package embedded

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

// TestAcceleratorApi_UpdatePhase_PacksUpdatePhaseMethod guards against packing
// the wrong ABI method name. The Accelerator ABI defines two distinct methods:
// "Update" (no inputs) and "UpdatePhase" (id, name, description, url,
// znnFundsNeeded, qsrFundsNeeded). UpdatePhase must pack UpdatePhaseMethodName;
// packing UpdateMethodName ("Update") with six arguments makes
// PackMethodPanic panic (arg count mismatch), so this test also guards the
// builder against panicking at all.
func TestAcceleratorApi_UpdatePhase_PacksUpdatePhaseMethod(t *testing.T) {
	api := NewAcceleratorApi(nil)

	id := types.HexToHashPanic("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	znn := big.NewInt(100)
	qsr := big.NewInt(200)

	// Must not panic.
	block := api.UpdatePhase(id, "Phase", "desc", "https://zenon.org", znn, qsr)
	if block == nil {
		t.Fatal("UpdatePhase returned nil")
	}
	if block.ToAddress != types.AcceleratorContract {
		t.Errorf("ToAddress = %s, want AcceleratorContract", block.ToAddress.String())
	}
	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard.String())
	}
	if block.Amount == nil || block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}

	want := definition.ABIAccelerator.PackMethodPanic(
		definition.UpdatePhaseMethodName, id, "Phase", "desc", "https://zenon.org", znn, qsr,
	)
	if !bytes.Equal(block.Data, want) {
		t.Errorf("UpdatePhase packed the wrong method; Data does not match UpdatePhaseMethodName encoding")
	}
}
