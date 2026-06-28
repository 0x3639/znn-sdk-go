package embedded

import (
	"bytes"
	"encoding/base64"
	"math/big"
	"testing"

	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

// assertPayloadDecodes checks that a ProposalPayload targets the expected
// destination and that its base64 Data decodes to the same ABI bytes the
// destination contract's builder produces.
func assertPayloadDecodes(t *testing.T, p ProposalPayload, wantDest types.Address, wantData []byte) {
	t.Helper()
	if p.Destination != wantDest {
		t.Errorf("Destination = %s, want %s", p.Destination.String(), wantDest.String())
	}
	decoded, err := base64.StdEncoding.DecodeString(p.Data)
	if err != nil {
		t.Fatalf("Data is not valid base64: %v", err)
	}
	if !bytes.Equal(decoded, wantData) {
		t.Errorf("decoded data mismatch\n  got:  %x\n  want: %x", decoded, wantData)
	}
}

func TestPayloadSpork(t *testing.T) {
	g := NewGovernanceApi(nil)
	id := types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	create := g.PayloadSporkCreate("gov", "governance ratchet")
	wantCreate := definition.ABISpork.PackMethodPanic(definition.SporkCreateMethodName, "gov", "governance ratchet")
	assertPayloadDecodes(t, create, types.SporkContract, wantCreate)

	activate := g.PayloadSporkActivate(id)
	wantActivate := definition.ABISpork.PackMethodPanic(definition.SporkActivateMethodName, id)
	assertPayloadDecodes(t, activate, types.SporkContract, wantActivate)
}

func TestPayloadBridgeSetTokenPair(t *testing.T) {
	g := NewGovernanceApi(nil)
	min := big.NewInt(1000)
	p := g.PayloadBridgeSetTokenPair(2, 1, types.ZnnTokenStandard, "0xabc", true, true, false, min, 10, 20, "meta")
	want := definition.ABIBridge.PackMethodPanic(definition.SetTokenPairMethod,
		uint32(2), uint32(1), types.ZnnTokenStandard, "0xabc", true, true, false, min, uint32(10), uint32(20), "meta")
	assertPayloadDecodes(t, p, types.BridgeContract, want)
}

func TestPayloadBridgeRemoveNetwork(t *testing.T) {
	g := NewGovernanceApi(nil)
	p := g.PayloadBridgeRemoveNetwork(2, 1)
	want := definition.ABIBridge.PackMethodPanic(definition.RemoveNetworkMethodName, uint32(2), uint32(1))
	assertPayloadDecodes(t, p, types.BridgeContract, want)
}

func TestPayloadLiquidityFund(t *testing.T) {
	g := NewGovernanceApi(nil)
	znn := big.NewInt(500)
	qsr := big.NewInt(600)
	p := g.PayloadLiquidityFund(znn, qsr)
	want := definition.ABILiquidity.PackMethodPanic(definition.FundMethodName, znn, qsr)
	assertPayloadDecodes(t, p, types.LiquidityContract, want)
}

func TestPayloadLiquidityBurnZnn(t *testing.T) {
	g := NewGovernanceApi(nil)
	burn := big.NewInt(700)
	p := g.PayloadLiquidityBurnZnn(burn)
	want := definition.ABILiquidity.PackMethodPanic(definition.BurnZnnMethodName, burn)
	assertPayloadDecodes(t, p, types.LiquidityContract, want)
}

func TestPayloadLiquidityChangeAdministrator(t *testing.T) {
	g := NewGovernanceApi(nil)
	addr := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	p := g.PayloadLiquidityChangeAdministrator(addr)
	want := definition.ABILiquidity.PackMethodPanic(definition.ChangeAdministratorMethodName, addr)
	assertPayloadDecodes(t, p, types.LiquidityContract, want)
}

// TestEncodeProposalPayload verifies the generic primitive round-trips an
// arbitrary template's destination and ABI data.
func TestEncodeProposalPayload(t *testing.T) {
	block := NewSporkApi(nil).CreateSpork("n", "d")
	p := EncodeProposalPayload(block)
	if p.Destination != block.ToAddress {
		t.Errorf("Destination = %s, want %s", p.Destination, block.ToAddress)
	}
	decoded, err := base64.StdEncoding.DecodeString(p.Data)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	if !bytes.Equal(decoded, block.Data) {
		t.Errorf("decoded data != block.Data")
	}
}
