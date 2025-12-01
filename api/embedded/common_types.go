// Package embedded provides client-side types for Zenon Network embedded contract APIs.
//
// These types are designed for deserializing JSON-RPC responses from Zenon nodes.
// They include custom UnmarshalJSON implementations to properly convert string
// representations of big integers (as sent by the RPC protocol) into Go's *big.Int.
//
// This follows the same pattern as the Dart SDK, which defines dedicated client-side
// types with fromJson factories rather than reusing server-side types.
package embedded

import (
	"encoding/json"
	"math/big"

	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
)

// UncollectedReward represents pending rewards that haven't been collected yet.
//
// This type is returned by GetUncollectedReward methods across pillar, sentinel,
// stake, and liquidity APIs. Rewards accumulate over time and can be collected
// using the respective CollectReward transaction methods.
//
// Fields:
//   - Address: The address that owns the uncollected rewards
//   - ZnnAmount: Uncollected ZNN rewards (in base units, 8 decimals)
//   - QsrAmount: Uncollected QSR rewards (in base units, 8 decimals)
//
// Example:
//
//	rewards, err := client.StakeApi.GetUncollectedReward(address)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if rewards.ZnnAmount.Sign() > 0 {
//	    fmt.Printf("Uncollected ZNN: %s\n", rewards.ZnnAmount)
//	}
type UncollectedReward struct {
	Address   types.Address `json:"address"`
	ZnnAmount *big.Int      `json:"znnAmount"`
	QsrAmount *big.Int      `json:"qsrAmount"`
}

// uncollectedRewardJSON is used for JSON unmarshaling with string amounts
type uncollectedRewardJSON struct {
	Address   types.Address `json:"address"`
	ZnnAmount string        `json:"znnAmount"`
	QsrAmount string        `json:"qsrAmount"`
}

func (r *UncollectedReward) UnmarshalJSON(data []byte) error {
	var aux uncollectedRewardJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.Address = aux.Address
	r.ZnnAmount = common.StringToBigInt(aux.ZnnAmount)
	r.QsrAmount = common.StringToBigInt(aux.QsrAmount)
	return nil
}

// RewardHistoryEntry represents a single reward collection event.
//
// This type records historical reward distributions by epoch. It is used
// in paginated lists returned by GetFrontierRewardByPage methods.
//
// Fields:
//   - Epoch: The epoch number when rewards were distributed
//   - ZnnAmount: ZNN rewards for this epoch (in base units, 8 decimals)
//   - QsrAmount: QSR rewards for this epoch (in base units, 8 decimals)
type RewardHistoryEntry struct {
	Epoch     uint64   `json:"epoch"`
	ZnnAmount *big.Int `json:"znnAmount"`
	QsrAmount *big.Int `json:"qsrAmount"`
}

// rewardHistoryEntryJSON is used for JSON unmarshaling with string amounts
type rewardHistoryEntryJSON struct {
	Epoch     uint64 `json:"epoch"`
	ZnnAmount string `json:"znnAmount"`
	QsrAmount string `json:"qsrAmount"`
}

func (r *RewardHistoryEntry) UnmarshalJSON(data []byte) error {
	var aux rewardHistoryEntryJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.Epoch = aux.Epoch
	r.ZnnAmount = common.StringToBigInt(aux.ZnnAmount)
	r.QsrAmount = common.StringToBigInt(aux.QsrAmount)
	return nil
}

// RewardHistoryList represents a paginated list of reward history entries
type RewardHistoryList struct {
	Count int                   `json:"count"`
	List  []*RewardHistoryEntry `json:"list"`
}

// SecurityInfo represents security configuration for embedded contracts
type SecurityInfo struct {
	Guardians          []types.Address `json:"guardians"`
	GuardiansVotes     []types.Address `json:"guardiansVotes"`
	AdministratorDelay uint64          `json:"administratorDelay"`
	SoftDelay          uint64          `json:"softDelay"`
}

// TimeChallengeInfo represents a time challenge for administrative operations
type TimeChallengeInfo struct {
	MethodName           string     `json:"MethodName"`
	ParamsHash           types.Hash `json:"ParamsHash"`
	ChallengeStartHeight uint64     `json:"ChallengeStartHeight"`
}

// TimeChallengesList represents a list of time challenges
type TimeChallengesList struct {
	Count int                  `json:"count"`
	List  []*TimeChallengeInfo `json:"list"`
}
