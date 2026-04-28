package embedded

import (
	sdkembedded "github.com/0x3639/znn-sdk-go/embedded"
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/server"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

// SecondsPerDay is the number of seconds in a day, used for epoch math
// in the swap-decay schedule. Mirrors Dart's Duration.secondsPerDay.
const SecondsPerDay = 24 * 60 * 60

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
func (sa *SwapApi) GetAssetsByKeyIdHash(keyIdHash types.Hash) (*SwapAssetEntry, error) {
	ans := new(SwapAssetEntry)
	if err := sa.client.Call(ans, "embedded.swap.getAssetsByKeyIdHash", keyIdHash.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *SwapApi) GetAssets() (map[types.Hash]*SwapAssetEntrySimple, error) {
	var ans map[types.Hash]*SwapAssetEntrySimple
	if err := sa.client.Call(ans, "embedded.swap.getAssets"); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *SwapApi) GetLegacyPillars() ([]*SwapLegacyPillarEntry, error) {
	var ans []*SwapLegacyPillarEntry
	if err := sa.client.Call(&ans, "embedded.swap.getLegacyPillars"); err != nil {
		return nil, err
	}
	return ans, nil
}

// RetrieveAssets creates a transaction template that retrieves swapped assets
// using a public key and signature proving ownership of the legacy key.
//
// Reference: znn_sdk_dart/lib/src/api/embedded/swap.dart:39-42
func (sa *SwapApi) RetrieveAssets(publicKey, signature string) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.SwapContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABISwap.PackMethodPanic(
			definition.RetrieveAssetsMethodName,
			publicKey,
			signature,
		),
	}
}

// GetSwapDecayPercentage returns the percentage of the swap that has decayed
// at the given Unix timestamp. The result is the inverse of percentageToGive
// in the Dart reference (i.e. 0 means no decay, 100 means fully decayed).
//
// Decay schedule (mirrors reference/.../lib/src/embedded/constants.dart):
//   - Before SwapAssetDecayTimestampStart: 0% decay
//   - After: every SwapAssetDecayTickEpochs days past the offset, the decay
//     grows by SwapAssetDecayTickValuePercentage percentage points until 100%
//
// Pure helper — no RPC call.
//
// Reference: znn_sdk_dart/lib/src/api/embedded/swap.dart:44-61
func (sa *SwapApi) GetSwapDecayPercentage(currentTimestamp int64) int {
	var percentageToGive int
	currentEpoch := (currentTimestamp - sdkembedded.GenesisTimestamp) / SecondsPerDay
	if currentTimestamp < sdkembedded.SwapAssetDecayTimestampStart {
		percentageToGive = 100
	} else {
		numTicks := (currentEpoch - sdkembedded.SwapAssetDecayEpochsOffset + 1) / sdkembedded.SwapAssetDecayTickEpochs
		decayFactor := sdkembedded.SwapAssetDecayTickValuePercentage * int(numTicks)
		if decayFactor > 100 {
			percentageToGive = 0
		} else {
			percentageToGive = 100 - decayFactor
		}
	}
	return 100 - percentageToGive
}
