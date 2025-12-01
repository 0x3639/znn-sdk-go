package embedded

import (
	"encoding/json"
	"math/big"

	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
)

// TokenTuple represents a token configuration for liquidity rewards.
//
// This type defines how rewards are distributed for a specific token in the
// liquidity program. Each token can have different reward percentages and
// minimum staking requirements.
//
// Fields:
//   - TokenStandard: ZTS identifier for the token
//   - ZnnPercentage: Percentage of ZNN rewards allocated to this token
//   - QsrPercentage: Percentage of QSR rewards allocated to this token
//   - MinAmount: Minimum amount required to stake (in base units, 8 decimals)
type TokenTuple struct {
	TokenStandard types.ZenonTokenStandard `json:"tokenStandard"`
	ZnnPercentage uint32                   `json:"znnPercentage"`
	QsrPercentage uint32                   `json:"qsrPercentage"`
	MinAmount     *big.Int                 `json:"minAmount"`
}

// tokenTupleJSON is used for JSON unmarshaling with string amounts
type tokenTupleJSON struct {
	TokenStandard types.ZenonTokenStandard `json:"tokenStandard"`
	ZnnPercentage uint32                   `json:"znnPercentage"`
	QsrPercentage uint32                   `json:"qsrPercentage"`
	MinAmount     string                   `json:"minAmount"`
}

func (t *TokenTuple) UnmarshalJSON(data []byte) error {
	var aux tokenTupleJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	t.TokenStandard = aux.TokenStandard
	t.ZnnPercentage = aux.ZnnPercentage
	t.QsrPercentage = aux.QsrPercentage
	t.MinAmount = common.StringToBigInt(aux.MinAmount)
	return nil
}

// LiquidityInfo represents liquidity contract configuration.
//
// This type contains the global configuration for the liquidity rewards program,
// including reward amounts and supported token configurations.
//
// Fields:
//   - Administrator: Address that can configure the liquidity contract
//   - IsHalted: Whether the liquidity program is currently paused
//   - ZnnReward: Total ZNN rewards available per epoch (in base units, 8 decimals)
//   - QsrReward: Total QSR rewards available per epoch (in base units, 8 decimals)
//   - TokenTuples: List of supported tokens and their reward configurations
type LiquidityInfo struct {
	Administrator types.Address `json:"administrator"`
	IsHalted      bool          `json:"isHalted"`
	ZnnReward     *big.Int      `json:"znnReward"`
	QsrReward     *big.Int      `json:"qsrReward"`
	TokenTuples   []*TokenTuple `json:"tokenTuples"`
}

// liquidityInfoJSON is used for JSON unmarshaling with string amounts
type liquidityInfoJSON struct {
	Administrator types.Address `json:"administrator"`
	IsHalted      bool          `json:"isHalted"`
	ZnnReward     string        `json:"znnReward"`
	QsrReward     string        `json:"qsrReward"`
	TokenTuples   []*TokenTuple `json:"tokenTuples"`
}

func (l *LiquidityInfo) UnmarshalJSON(data []byte) error {
	var aux liquidityInfoJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	l.Administrator = aux.Administrator
	l.IsHalted = aux.IsHalted
	l.ZnnReward = common.StringToBigInt(aux.ZnnReward)
	l.QsrReward = common.StringToBigInt(aux.QsrReward)
	l.TokenTuples = aux.TokenTuples
	return nil
}

// LiquidityStakeEntry represents a single liquidity stake entry.
//
// Users can stake supported tokens in the liquidity program to earn ZNN and QSR
// rewards. This type contains information about an individual liquidity stake.
//
// Fields:
//   - Amount: Staked amount (in base units, 8 decimals)
//   - TokenStandard: ZTS identifier of the staked token
//   - WeightedAmount: Amount multiplied by duration factor for reward calculation
//   - StartTime: Unix timestamp when the stake was created
//   - RevokeTime: Unix timestamp when revocation was initiated (0 if not revoked)
//   - ExpirationTime: Unix timestamp when the stake can be canceled
//   - StakeAddress: Address that owns this stake
//   - Id: Unique identifier for this stake entry
type LiquidityStakeEntry struct {
	Amount         *big.Int                 `json:"amount"`
	TokenStandard  types.ZenonTokenStandard `json:"tokenStandard"`
	WeightedAmount *big.Int                 `json:"weightedAmount"`
	StartTime      int64                    `json:"startTime"`
	RevokeTime     int64                    `json:"revokeTime"`
	ExpirationTime int64                    `json:"expirationTime"`
	StakeAddress   types.Address            `json:"stakeAddress"`
	Id             types.Hash               `json:"id"`
}

// liquidityStakeEntryJSON is used for JSON unmarshaling with string amounts
type liquidityStakeEntryJSON struct {
	Amount         string                   `json:"amount"`
	TokenStandard  types.ZenonTokenStandard `json:"tokenStandard"`
	WeightedAmount string                   `json:"weightedAmount"`
	StartTime      int64                    `json:"startTime"`
	RevokeTime     int64                    `json:"revokeTime"`
	ExpirationTime int64                    `json:"expirationTime"`
	StakeAddress   types.Address            `json:"stakeAddress"`
	Id             types.Hash               `json:"id"`
}

func (l *LiquidityStakeEntry) UnmarshalJSON(data []byte) error {
	var aux liquidityStakeEntryJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	l.Amount = common.StringToBigInt(aux.Amount)
	l.TokenStandard = aux.TokenStandard
	l.WeightedAmount = common.StringToBigInt(aux.WeightedAmount)
	l.StartTime = aux.StartTime
	l.RevokeTime = aux.RevokeTime
	l.ExpirationTime = aux.ExpirationTime
	l.StakeAddress = aux.StakeAddress
	l.Id = aux.Id
	return nil
}

// LiquidityStakeList represents a paginated list of liquidity stake entries.
//
// This type is returned by methods that list liquidity stakes. It includes
// aggregate totals for all stakes belonging to the queried address.
//
// Fields:
//   - TotalAmount: Sum of all staked amounts (in base units, 8 decimals)
//   - TotalWeightedAmount: Sum of all weighted amounts for reward calculation
//   - Count: Total number of stake entries matching the query
//   - List: Slice of LiquidityStakeEntry entries for the current page
type LiquidityStakeList struct {
	TotalAmount         *big.Int               `json:"totalAmount"`
	TotalWeightedAmount *big.Int               `json:"totalWeightedAmount"`
	Count               int                    `json:"count"`
	List                []*LiquidityStakeEntry `json:"list"`
}

// liquidityStakeListJSON is used for JSON unmarshaling with string amounts
type liquidityStakeListJSON struct {
	TotalAmount         string                 `json:"totalAmount"`
	TotalWeightedAmount string                 `json:"totalWeightedAmount"`
	Count               int                    `json:"count"`
	List                []*LiquidityStakeEntry `json:"list"`
}

func (l *LiquidityStakeList) UnmarshalJSON(data []byte) error {
	var aux liquidityStakeListJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	l.TotalAmount = common.StringToBigInt(aux.TotalAmount)
	l.TotalWeightedAmount = common.StringToBigInt(aux.TotalWeightedAmount)
	l.Count = aux.Count
	l.List = aux.List
	return nil
}
