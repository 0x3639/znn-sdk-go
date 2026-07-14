package embedded

import (
	"strings"
	"testing"

	"github.com/zenon-network/go-zenon/common/types"
)

func TestEmbeddedPaginationRejectsOversizedArgumentsBeforeCalling(t *testing.T) {
	t.Parallel()
	address := types.Address{}
	tests := []struct {
		name string
		call func() error
	}{
		{"accelerator.getAll", func() error { _, err := NewAcceleratorApi(nil).GetAll(0, 1025); return err }},
		{"bridge.getAllWrapTokenRequests", func() error { _, err := NewBridgeApi(nil).GetAllWrapTokenRequests(0, 1025); return err }},
		{"bridge.getAllWrapTokenRequestsByToAddress", func() error { _, err := NewBridgeApi(nil).GetAllWrapTokenRequestsByToAddress("", 0, 1025); return err }},
		{"bridge.getAllWrapTokenRequestsByNetwork", func() error {
			_, err := NewBridgeApi(nil).GetAllWrapTokenRequestsByToAddressNetworkClassAndChainId("", 0, 0, 0, 1025)
			return err
		}},
		{"bridge.getAllNetworks", func() error { _, err := NewBridgeApi(nil).GetAllNetworks(0, 1025); return err }},
		{"bridge.getAllUnsignedWrapTokenRequests", func() error { _, err := NewBridgeApi(nil).GetAllUnsignedWrapTokenRequests(0, 1025); return err }},
		{"bridge.getAllUnwrapTokenRequests", func() error { _, err := NewBridgeApi(nil).GetAllUnwrapTokenRequests(0, 1025); return err }},
		{"bridge.getAllUnwrapTokenRequestsByToAddress", func() error {
			_, err := NewBridgeApi(nil).GetAllUnwrapTokenRequestsByToAddress("", 0, 1025)
			return err
		}},
		{"liquidity.getFrontierRewardByPage", func() error { _, err := NewLiquidityApi(nil).GetFrontierRewardByPage(address, 0, 1025); return err }},
		{"liquidity.getLiquidityStakeEntriesByAddress", func() error {
			_, err := NewLiquidityApi(nil).GetLiquidityStakeEntriesByAddress(address, 0, 51)
			return err
		}},
		{"pillar.getFrontierRewardByPage", func() error { _, err := NewPillarApi(nil).GetFrontierRewardByPage(address, 0, 1025); return err }},
		{"pillar.getAll", func() error { _, err := NewPillarApi(nil).GetAll(0, 1025); return err }},
		{"pillar.getPillarEpochHistory", func() error { _, err := NewPillarApi(nil).GetPillarEpochHistory("", 0, 1025); return err }},
		{"pillar.getPillarsHistoryByEpoch", func() error { _, err := NewPillarApi(nil).GetPillarsHistoryByEpoch(0, 0, 1025); return err }},
		{"plasma.getEntriesByAddress", func() error { _, err := NewPlasmaApi(nil).GetEntriesByAddress(address, 0, 1025); return err }},
		{"sentinel.getAllActive", func() error { _, err := NewSentinelApi(nil).GetAllActive(0, 1025); return err }},
		{"sentinel.getFrontierRewardByPage", func() error { _, err := NewSentinelApi(nil).GetFrontierRewardByPage(address, 0, 1025); return err }},
		{"spork.getAll", func() error { _, err := NewSporkApi(nil).GetAll(0, 1025); return err }},
		{"stake.getFrontierRewardByPage", func() error { _, err := NewStakeApi(nil).GetFrontierRewardByPage(address, 0, 1025); return err }},
		{"stake.getEntriesByAddress", func() error { _, err := NewStakeApi(nil).GetEntriesByAddress(address, 0, 1025); return err }},
		{"token.getAll", func() error { _, err := NewTokenApi(nil).GetAll(0, 1025); return err }},
		{"token.getByOwner", func() error { _, err := NewTokenApi(nil).GetByOwner(address, 0, 1025); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); err == nil || !strings.Contains(err.Error(), "exceeds maximum") {
				t.Fatalf("oversized request error = %v", err)
			}
		})
	}
}
