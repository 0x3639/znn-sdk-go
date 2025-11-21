package embedded

import (
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/api/embedded"
	"github.com/zenon-network/go-zenon/rpc/server"
)

type SwapApi struct {
	client *server.Client
}

func NewSwapApi(client *server.Client) *SwapApi {
	return &SwapApi{
		client: client,
	}
}

// GetAssetsByKeyIdHash retrieves swap asset information by key ID hash.
//
// The Swap contract manages the legacy swap from old Zenon assets. This query
// returns information about swapped assets for a specific key.
//
// Parameters:
//   - keyIdHash: Hash of the key ID to query
//
// Returns swap asset entry or an error.
func (sa *SwapApi) GetAssetsByKeyIdHash(keyIdHash types.Hash) (*embedded.SwapAssetEntry, error) {
	ans := new(embedded.SwapAssetEntry)
	if err := sa.client.Call(ans, "embedded.swap.getAssetsByKeyIdHash", keyIdHash.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *SwapApi) GetAssets() (map[types.Hash]*embedded.SwapAssetEntrySimple, error) {
	var ans map[types.Hash]*embedded.SwapAssetEntrySimple
	if err := sa.client.Call(ans, "embedded.swap.getAssets"); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *SwapApi) GetLegacyPillars() ([]*embedded.SwapLegacyPillarEntry, error) {
	var ans []*embedded.SwapLegacyPillarEntry
	if err := sa.client.Call(&ans, "embedded.swap.getLegacyPillars"); err != nil {
		return nil, err
	}
	return ans, nil
}
