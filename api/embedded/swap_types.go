package embedded

import (
	"encoding/json"
	"math/big"

	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
)

// SwapAssetEntry represents swap asset information with full details.
//
// This type is used for the legacy network swap functionality, which allows
// users who held tokens on the legacy Zenon network to claim their tokens
// on the current network.
//
// Fields:
//   - KeyIdHash: Hash identifier linking to the legacy network address
//   - Qsr: QSR amount available to claim (in base units, 8 decimals)
//   - Znn: ZNN amount available to claim (in base units, 8 decimals)
type SwapAssetEntry struct {
	KeyIdHash types.Hash `json:"keyIdHash"`
	Qsr       *big.Int   `json:"qsr"`
	Znn       *big.Int   `json:"znn"`
}

// swapAssetEntryJSON is used for JSON unmarshaling with string amounts
type swapAssetEntryJSON struct {
	KeyIdHash types.Hash `json:"keyIdHash"`
	Qsr       string     `json:"qsr"`
	Znn       string     `json:"znn"`
}

func (s *SwapAssetEntry) UnmarshalJSON(data []byte) error {
	var aux swapAssetEntryJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	s.KeyIdHash = aux.KeyIdHash
	s.Qsr = common.StringToBigInt(aux.Qsr)
	s.Znn = common.StringToBigInt(aux.Znn)
	return nil
}

// HasBalance returns true if there is any remaining balance.
//
// This method checks whether there are any unclaimed tokens (either ZNN or QSR)
// available in this swap entry.
//
// Returns:
//   - true if either Qsr or Znn has a positive balance
//   - false if both balances are zero or negative
func (s *SwapAssetEntry) HasBalance() bool {
	return s.Qsr.Sign() > 0 || s.Znn.Sign() > 0
}

// SwapAssetEntrySimple represents simplified swap asset information.
//
// A simplified version of SwapAssetEntry without the KeyIdHash, used when
// only the token amounts are needed.
//
// Fields:
//   - Qsr: QSR amount (in base units, 8 decimals)
//   - Znn: ZNN amount (in base units, 8 decimals)
type SwapAssetEntrySimple struct {
	Qsr *big.Int `json:"qsr"`
	Znn *big.Int `json:"znn"`
}

// swapAssetEntrySimpleJSON is used for JSON unmarshaling with string amounts
type swapAssetEntrySimpleJSON struct {
	Qsr string `json:"qsr"`
	Znn string `json:"znn"`
}

func (s *SwapAssetEntrySimple) UnmarshalJSON(data []byte) error {
	var aux swapAssetEntrySimpleJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	s.Qsr = common.StringToBigInt(aux.Qsr)
	s.Znn = common.StringToBigInt(aux.Znn)
	return nil
}

// SwapLegacyPillarEntry represents a legacy pillar swap entry.
//
// This type is used for legacy pillar holders who need to claim their
// pillar slots on the current network.
//
// Fields:
//   - NumPillars: Number of pillar slots to claim
//   - KeyIdHash: Hash identifier linking to the legacy network address
type SwapLegacyPillarEntry struct {
	NumPillars int        `json:"numPillars"`
	KeyIdHash  types.Hash `json:"keyIdHash"`
}
