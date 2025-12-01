package embedded

import (
	"encoding/json"
	"math/big"

	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
)

// Pillar type constants identify the origin of a Pillar.
const (
	// UnknownPillarType indicates a pillar with unknown origin.
	UnknownPillarType = 0
	// LegacyPillarType indicates a pillar migrated from the legacy network.
	LegacyPillarType = 1
	// RegularPillarType indicates a pillar registered on the current network.
	RegularPillarType = 2
)

// PillarEpochStats represents momentum production statistics for an epoch.
//
// This type tracks how many momentums a Pillar was expected to produce versus
// how many it actually produced during a given epoch. This data is used for
// calculating Pillar performance metrics.
//
// Fields:
//   - ProducedMomentums: Number of momentums the Pillar actually produced
//   - ExpectedMomentums: Number of momentums the Pillar was expected to produce
type PillarEpochStats struct {
	ProducedMomentums int32 `json:"producedMomentums"`
	ExpectedMomentums int32 `json:"expectedMomentums"`
}

// PillarInfo represents detailed information about a Pillar.
//
// Pillars are the consensus nodes of the Zenon Network. They produce momentums
// and participate in network governance. This type contains all configuration
// and status information for a Pillar.
//
// Fields:
//   - Name: Unique name of the Pillar
//   - Rank: Current rank based on total weight (delegations)
//   - Type: Origin type (UnknownPillarType, LegacyPillarType, or RegularPillarType)
//   - OwnerAddress: Address that owns and controls the Pillar
//   - ProducerAddress: Address used for producing momentums
//   - WithdrawAddress: Address for withdrawing rewards
//   - GiveMomentumRewardPercentage: Percentage of momentum rewards shared with delegators
//   - GiveDelegateRewardPercentage: Percentage of delegate rewards shared
//   - IsRevocable: Whether the Pillar can be revoked
//   - RevokeCooldown: Remaining cooldown time before revocation completes
//   - RevokeTimestamp: Unix timestamp when revocation was initiated
//   - CurrentStats: Momentum production statistics for current epoch
//   - Weight: Total delegation weight (in base units, 8 decimals)
//
// Example:
//
//	pillar, err := client.PillarApi.GetByName("MyPillar")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Pillar rank: %d, Weight: %s\n", pillar.Rank, pillar.Weight)
type PillarInfo struct {
	Name                         string            `json:"name"`
	Rank                         int32             `json:"rank"`
	Type                         int32             `json:"type"`
	OwnerAddress                 types.Address     `json:"ownerAddress"`
	ProducerAddress              types.Address     `json:"producerAddress"`
	WithdrawAddress              types.Address     `json:"withdrawAddress"`
	GiveMomentumRewardPercentage int32             `json:"giveMomentumRewardPercentage"`
	GiveDelegateRewardPercentage int32             `json:"giveDelegateRewardPercentage"`
	IsRevocable                  bool              `json:"isRevocable"`
	RevokeCooldown               int64             `json:"revokeCooldown"`
	RevokeTimestamp              int64             `json:"revokeTimestamp"`
	CurrentStats                 *PillarEpochStats `json:"currentStats"`
	Weight                       *big.Int          `json:"weight"`
}

// pillarInfoJSON is used for JSON unmarshaling with string amounts
type pillarInfoJSON struct {
	Name                         string            `json:"name"`
	Rank                         int32             `json:"rank"`
	Type                         int32             `json:"type"`
	OwnerAddress                 types.Address     `json:"ownerAddress"`
	ProducerAddress              types.Address     `json:"producerAddress"`
	WithdrawAddress              types.Address     `json:"withdrawAddress"`
	GiveMomentumRewardPercentage int32             `json:"giveMomentumRewardPercentage"`
	GiveDelegateRewardPercentage int32             `json:"giveDelegateRewardPercentage"`
	IsRevocable                  bool              `json:"isRevocable"`
	RevokeCooldown               int64             `json:"revokeCooldown"`
	RevokeTimestamp              int64             `json:"revokeTimestamp"`
	CurrentStats                 *PillarEpochStats `json:"currentStats"`
	Weight                       string            `json:"weight"`
}

func (p *PillarInfo) UnmarshalJSON(data []byte) error {
	var aux pillarInfoJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	p.Name = aux.Name
	p.Rank = aux.Rank
	p.Type = aux.Type
	p.OwnerAddress = aux.OwnerAddress
	p.ProducerAddress = aux.ProducerAddress
	p.WithdrawAddress = aux.WithdrawAddress
	p.GiveMomentumRewardPercentage = aux.GiveMomentumRewardPercentage
	p.GiveDelegateRewardPercentage = aux.GiveDelegateRewardPercentage
	p.IsRevocable = aux.IsRevocable
	p.RevokeCooldown = aux.RevokeCooldown
	p.RevokeTimestamp = aux.RevokeTimestamp
	p.CurrentStats = aux.CurrentStats
	p.Weight = common.StringToBigInt(aux.Weight)
	return nil
}

// PillarInfoList represents a paginated list of Pillars.
//
// This type is returned by methods that list multiple Pillars, such as GetAll.
//
// Fields:
//   - Count: Total number of Pillars matching the query
//   - List: Slice of PillarInfo entries for the current page
type PillarInfoList struct {
	Count int           `json:"count"`
	List  []*PillarInfo `json:"list"`
}

// PillarEpochHistory represents historical data for a Pillar in a specific epoch.
//
// This type records a Pillar's configuration and performance during a past epoch,
// useful for analyzing historical trends and reward distributions.
//
// Fields:
//   - Name: Name of the Pillar
//   - Epoch: The epoch number this history entry represents
//   - GiveBlockRewardPercentage: Momentum reward percentage during this epoch
//   - GiveDelegateRewardPercentage: Delegate reward percentage during this epoch
//   - ProducedBlockNum: Number of momentums produced during this epoch
//   - ExpectedBlockNum: Number of momentums expected during this epoch
//   - Weight: Total delegation weight during this epoch (in base units, 8 decimals)
type PillarEpochHistory struct {
	Name                         string   `json:"name"`
	Epoch                        uint64   `json:"epoch"`
	GiveBlockRewardPercentage    int32    `json:"giveBlockRewardPercentage"`
	GiveDelegateRewardPercentage int32    `json:"giveDelegateRewardPercentage"`
	ProducedBlockNum             int32    `json:"producedBlockNum"`
	ExpectedBlockNum             int32    `json:"expectedBlockNum"`
	Weight                       *big.Int `json:"weight"`
}

// pillarEpochHistoryJSON is used for JSON unmarshaling with string amounts
type pillarEpochHistoryJSON struct {
	Name                         string `json:"name"`
	Epoch                        uint64 `json:"epoch"`
	GiveBlockRewardPercentage    int32  `json:"giveBlockRewardPercentage"`
	GiveDelegateRewardPercentage int32  `json:"giveDelegateRewardPercentage"`
	ProducedBlockNum             int32  `json:"producedBlockNum"`
	ExpectedBlockNum             int32  `json:"expectedBlockNum"`
	Weight                       string `json:"weight"`
}

func (p *PillarEpochHistory) UnmarshalJSON(data []byte) error {
	var aux pillarEpochHistoryJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	p.Name = aux.Name
	p.Epoch = aux.Epoch
	p.GiveBlockRewardPercentage = aux.GiveBlockRewardPercentage
	p.GiveDelegateRewardPercentage = aux.GiveDelegateRewardPercentage
	p.ProducedBlockNum = aux.ProducedBlockNum
	p.ExpectedBlockNum = aux.ExpectedBlockNum
	p.Weight = common.StringToBigInt(aux.Weight)
	return nil
}

// PillarEpochHistoryList represents a paginated list of Pillar epoch histories.
//
// This type is returned by methods that query historical Pillar data, such as
// GetPillarEpochHistory.
//
// Fields:
//   - Count: Total number of history entries matching the query
//   - List: Slice of PillarEpochHistory entries for the current page
type PillarEpochHistoryList struct {
	Count int                   `json:"count"`
	List  []*PillarEpochHistory `json:"list"`
}

// DelegationInfo represents delegation information for an address.
//
// When a user delegates their ZNN to a Pillar, they earn a share of the Pillar's
// rewards proportional to their delegation weight. This type contains information
// about a delegation relationship.
//
// Fields:
//   - Name: Name of the Pillar being delegated to (empty if no delegation)
//   - Status: Delegation status (1 = active, 0 = inactive/no delegation)
//   - Weight: Delegated amount (in base units, 8 decimals)
//
// Example:
//
//	delegation, err := client.PillarApi.GetDelegatedPillar(address)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if delegation.IsPillarActive() {
//	    fmt.Printf("Delegated to: %s, Weight: %s\n", delegation.Name, delegation.Weight)
//	}
type DelegationInfo struct {
	Name       string   `json:"name"`
	Status     int32    `json:"status"`
	Weight     *big.Int `json:"weight"`
	WeightJson string   `json:"-"` // Internal field to track original
}

// delegationInfoJSON is used for JSON unmarshaling with string amounts
type delegationInfoJSON struct {
	Name   string `json:"name"`
	Status int32  `json:"status"`
	Weight string `json:"weight"`
}

func (d *DelegationInfo) UnmarshalJSON(data []byte) error {
	var aux delegationInfoJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	d.Name = aux.Name
	d.Status = aux.Status
	d.Weight = common.StringToBigInt(aux.Weight)
	return nil
}

// IsPillarActive returns true if the delegation is to an active Pillar.
//
// A delegation is considered active when Status equals 1, meaning the user
// has an active delegation to a Pillar that is currently participating in
// the network.
//
// Returns:
//   - true if delegated to an active Pillar
//   - false if no delegation or Pillar is inactive
func (d *DelegationInfo) IsPillarActive() bool {
	return d.Status == 1
}
