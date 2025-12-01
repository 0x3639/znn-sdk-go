package embedded

import (
	"encoding/json"
	"math/big"

	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
)

// StakeEntry represents a single staking entry.
//
// Users can stake ZNN for various durations (1-12 months) to earn QSR rewards.
// Longer staking durations receive higher reward multipliers. This type contains
// information about an individual stake.
//
// Fields:
//   - Amount: Staked amount (in base units, 8 decimals)
//   - WeightedAmount: Amount multiplied by duration factor for reward calculation
//   - StartTimestamp: Unix timestamp when the stake was created
//   - ExpirationTimestamp: Unix timestamp when the stake can be canceled
//   - Address: Address that owns this stake
//   - Id: Unique identifier for this stake entry
//
// Duration Multipliers:
//   - 1 month:  1x weight
//   - 3 months: 3x weight
//   - 6 months: 6x weight
//   - 12 months: 12x weight
//
// Example:
//
//	stakes, err := client.StakeApi.GetEntriesByAddress(address, 0, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, stake := range stakes.List {
//	    expired := stake.ExpirationTimestamp <= time.Now().Unix()
//	    fmt.Printf("Stake: %s ZNN, Expired: %t\n", stake.Amount, expired)
//	}
type StakeEntry struct {
	Amount              *big.Int      `json:"amount"`
	WeightedAmount      *big.Int      `json:"weightedAmount"`
	StartTimestamp      int64         `json:"startTimestamp"`
	ExpirationTimestamp int64         `json:"expirationTimestamp"`
	Address             types.Address `json:"address"`
	Id                  types.Hash    `json:"id"`
}

// stakeEntryJSON is used for JSON unmarshaling with string amounts
type stakeEntryJSON struct {
	Amount              string        `json:"amount"`
	WeightedAmount      string        `json:"weightedAmount"`
	StartTimestamp      int64         `json:"startTimestamp"`
	ExpirationTimestamp int64         `json:"expirationTimestamp"`
	Address             types.Address `json:"address"`
	Id                  types.Hash    `json:"id"`
}

func (s *StakeEntry) UnmarshalJSON(data []byte) error {
	var aux stakeEntryJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	s.Amount = common.StringToBigInt(aux.Amount)
	s.WeightedAmount = common.StringToBigInt(aux.WeightedAmount)
	s.StartTimestamp = aux.StartTimestamp
	s.ExpirationTimestamp = aux.ExpirationTimestamp
	s.Address = aux.Address
	s.Id = aux.Id
	return nil
}

// StakeList represents a paginated list of stake entries.
//
// This type is returned by methods that list stakes, such as GetEntriesByAddress.
// It includes aggregate totals for all stakes belonging to the address.
//
// Fields:
//   - TotalAmount: Sum of all staked amounts (in base units, 8 decimals)
//   - TotalWeightedAmount: Sum of all weighted amounts for reward calculation
//   - Count: Total number of stake entries matching the query
//   - List: Slice of StakeEntry entries for the current page
//
// Example:
//
//	stakes, err := client.StakeApi.GetEntriesByAddress(address, 0, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Total staked: %s ZNN across %d entries\n",
//	    stakes.TotalAmount, stakes.Count)
type StakeList struct {
	TotalAmount         *big.Int      `json:"totalAmount"`
	TotalWeightedAmount *big.Int      `json:"totalWeightedAmount"`
	Count               int           `json:"count"`
	List                []*StakeEntry `json:"list"`
}

// stakeListJSON is used for JSON unmarshaling with string amounts
type stakeListJSON struct {
	TotalAmount         string        `json:"totalAmount"`
	TotalWeightedAmount string        `json:"totalWeightedAmount"`
	Count               int           `json:"count"`
	List                []*StakeEntry `json:"list"`
}

func (s *StakeList) UnmarshalJSON(data []byte) error {
	var aux stakeListJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	s.TotalAmount = common.StringToBigInt(aux.TotalAmount)
	s.TotalWeightedAmount = common.StringToBigInt(aux.TotalWeightedAmount)
	s.Count = aux.Count
	s.List = aux.List
	return nil
}
