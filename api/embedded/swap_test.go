package embedded

import (
	"bytes"
	"testing"

	sdkembedded "github.com/0x3639/znn-sdk-go/embedded"
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

func TestSwapApi_RetrieveAssets(t *testing.T) {
	api := NewSwapApi(nil)
	const pubKey = "0x04abcdef"
	const sig = "0xdeadbeef"

	block := api.RetrieveAssets(pubKey, sig)
	if block == nil {
		t.Fatal("RetrieveAssets returned nil")
	}
	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}
	if block.ToAddress != types.SwapContract {
		t.Errorf("ToAddress = %s, want SwapContract", block.ToAddress.String())
	}
	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard.String())
	}
	if block.Amount == nil || block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}
	expected := definition.ABISwap.PackMethodPanic(definition.RetrieveAssetsMethodName, pubKey, sig)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch\n  got:  %x\n  want: %x", block.Data, expected)
	}
}

// TestSwapApi_GetSwapDecayPercentage walks the decay schedule defined in
// reference/znn_sdk_dart-master/lib/src/api/embedded/swap.dart:44-61.
//
// The expected values were computed by hand from the schedule (genesis epoch +
// SwapAssetDecayEpochsOffset, then +SwapAssetDecayTickValuePercentage per
// SwapAssetDecayTickEpochs days). They MUST match what the Dart helper returns
// for the same inputs.
func TestSwapApi_GetSwapDecayPercentage(t *testing.T) {
	api := NewSwapApi(nil)

	tests := []struct {
		name      string
		timestamp int64
		want      int
	}{
		{
			name:      "before decay start: no decay",
			timestamp: sdkembedded.SwapAssetDecayTimestampStart - 1,
			want:      0,
		},
		{
			name:      "at decay start: still 0",
			timestamp: sdkembedded.SwapAssetDecayTimestampStart,
			want:      0,
		},
		{
			// 30 days after start crosses the first tick boundary -> 10% decay
			name:      "after first tick (30 days)",
			timestamp: sdkembedded.SwapAssetDecayTimestampStart + 30*int64(SecondsPerDay),
			want:      10,
		},
		{
			// 300 days after start -> 10 ticks -> 100% decay
			name:      "after 10 ticks (300 days)",
			timestamp: sdkembedded.SwapAssetDecayTimestampStart + 300*int64(SecondsPerDay),
			want:      100,
		},
		{
			// Past full decay -> capped at 100
			name:      "after 11 ticks (330 days, capped)",
			timestamp: sdkembedded.SwapAssetDecayTimestampStart + 330*int64(SecondsPerDay),
			want:      100,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := api.GetSwapDecayPercentage(tc.timestamp)
			if got != tc.want {
				t.Errorf("GetSwapDecayPercentage(%d) = %d, want %d", tc.timestamp, got, tc.want)
			}
		})
	}
}
