package embedded

import (
	"math/big"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/server"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

type StakeApi struct {
	client *server.Client
}

func NewStakeApi(client *server.Client) *StakeApi {
	return &StakeApi{
		client: client,
	}
}

// GetUncollectedReward retrieves pending staking rewards that haven't been collected yet.
//
// Staking ZNN generates rewards over time. These rewards accumulate and must be
// explicitly collected using CollectReward(). This method shows the current uncollected
// amount available for withdrawal.
//
// Returns a RewardDeposit containing:
//   - Qsr: Pending QSR rewards
//   - Znn: Pending ZNN rewards
//
// Parameters:
//   - address: Address to check for uncollected rewards
//
// Returns reward deposit information or an error.
//
// Example:
//
//	rewards, err := client.StakeApi.GetUncollectedReward(address)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Uncollected ZNN: %s\n", rewards.Znn)
//	fmt.Printf("Uncollected QSR: %s\n", rewards.Qsr)
//
//	// Collect rewards if any available
//	if rewards.Qsr.Cmp(big.NewInt(0)) > 0 || rewards.Znn.Cmp(big.NewInt(0)) > 0 {
//	    template := client.StakeApi.CollectReward()
//	    // Sign and publish transaction
//	}
func (sa *StakeApi) GetUncollectedReward(address types.Address) (*UncollectedReward, error) {
	ans := new(UncollectedReward)
	if err := sa.client.Call(ans, "embedded.stake.getUncollectedReward", address.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

// GetFrontierRewardByPage retrieves a paginated history of collected staking rewards.
//
// This provides a historical record of all reward collections, useful for:
//   - Tracking reward earnings over time
//   - Auditing reward history
//   - Analyzing staking performance
//   - Generating reward reports
//
// Each entry includes:
//   - ZNN and QSR amounts collected
//   - Timestamp of collection
//   - Momentum height when collected
//
// Parameters:
//   - address: Address to query reward history for
//   - pageIndex: Page number (0-indexed)
//   - pageSize: Number of entries per page
//
// Returns paginated reward history or an error.
//
// Example:
//
//	history, err := client.StakeApi.GetFrontierRewardByPage(address, 0, 25)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Total reward collections: %d\n", history.Count)
//	for _, entry := range history.List {
//	    fmt.Printf("Collected: %s ZNN, %s QSR at momentum %d\n",
//	        entry.Znn, entry.Qsr, entry.Height)
//	}
func (sa *StakeApi) GetFrontierRewardByPage(address types.Address, pageIndex, pageSize uint32) (*RewardHistoryList, error) {
	ans := new(RewardHistoryList)
	if err := sa.client.Call(ans, "embedded.stake.getFrontierRewardByPage", address.String(), pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

// GetEntriesByAddress retrieves all active stake entries for an address.
//
// Each stake entry represents a separate staking commitment with:
//   - Staked amount (ZNN)
//   - Start timestamp
//   - Expiration timestamp
//   - Duration in months (1, 3, 6, or 12)
//   - Unique stake ID for cancellation
//
// Use this to:
//   - Monitor active stakes
//   - Check when stakes expire
//   - Find stakes eligible for cancellation
//   - Calculate total staked amount
//
// Parameters:
//   - address: Address to query stake entries for
//   - pageIndex: Page number (0-indexed)
//   - pageSize: Number of entries per page
//
// Returns paginated stake list or an error.
//
// Example:
//
//	stakes, err := client.StakeApi.GetEntriesByAddress(address, 0, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Active stakes: %d\n", stakes.Count)
//	totalStaked := big.NewInt(0)
//	for _, stake := range stakes.List {
//	    totalStaked.Add(totalStaked, stake.Amount)
//	    fmt.Printf("Stake: %s ZNN, expires at %v\n",
//	        stake.Amount, stake.ExpirationTimestamp)
//	}
//	fmt.Printf("Total staked: %s ZNN\n", totalStaked)
func (sa *StakeApi) GetEntriesByAddress(address types.Address, pageIndex, pageSize uint32) (*StakeList, error) {
	ans := new(StakeList)
	if err := sa.client.Call(ans, "embedded.stake.getEntriesByAddress", address.String(), pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

// Contract calls

// Stake creates a transaction template to stake ZNN and earn rewards.
//
// Staking is Zenon's native yield mechanism. By locking ZNN for a specified duration,
// you earn both ZNN and QSR rewards proportional to the amount and duration.
//
// Staking parameters:
//   - Minimum amount: 1 ZNN (10^8 base units)
//   - Duration options (in seconds):
//   - 1 month: 2592000 (30 days)
//   - 3 months: 7776000 (90 days)
//   - 6 months: 15552000 (180 days)
//   - 12 months: 31536000 (365 days)
//   - Longer durations = higher rewards
//   - Can stake multiple times with different durations
//
// Reward mechanics:
//   - Rewards begin accruing immediately after staking
//   - Must call CollectReward() to claim accumulated rewards
//   - Stake entries are locked until duration expires
//   - Early cancellation not possible
//
// Parameters:
//   - durationInSec: Stake duration in seconds (must match valid options above)
//   - amount: Amount of ZNN to stake (in base units: 1 ZNN = 10^8)
//
// Returns an unsigned AccountBlock template ready for processing.
//
// Example - Stake for 1 month:
//
//	amount := big.NewInt(100 * 100000000)  // Stake 100 ZNN
//	duration := int64(2592000)             // 1 month in seconds
//
//	template := client.StakeApi.Stake(duration, amount)
//	// Sign and publish transaction
//
// Example - Stake for maximum rewards (12 months):
//
//	amount := big.NewInt(1000 * 100000000) // Stake 1000 ZNN
//	duration := int64(31536000)            // 12 months = highest rewards
//
//	template := client.StakeApi.Stake(duration, amount)
//	// Process through transaction pipeline
//
// Example - Multiple stake entries:
//
//	// Diversify by creating multiple stakes with different durations
//	stake1 := client.StakeApi.Stake(2592000, big.NewInt(100*100000000))  // 1 month
//	stake2 := client.StakeApi.Stake(15552000, big.NewInt(500*100000000)) // 6 months
//	// Each creates a separate entry with different expiration times
//
// Note: Staked ZNN is locked and cannot be withdrawn until the duration expires.
// Plan your liquidity needs accordingly.
func (sa *StakeApi) Stake(durationInSec int64, amount *big.Int) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.StakeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        amount,
		Data:          definition.ABIStake.PackMethodPanic(definition.StakeMethodName, durationInSec),
	}
}

// Cancel creates a transaction template to cancel an expired stake and reclaim ZNN.
//
// After a stake's duration expires, you must explicitly cancel it to reclaim your ZNN.
// The staked amount returns to your balance, but the stake entry remains until canceled.
//
// Cancellation requirements:
//   - Stake duration must have fully elapsed
//   - Cannot cancel before expiration timestamp
//   - Stake ID obtained from GetEntriesByAddress()
//
// Process:
//  1. Check stake entries with GetEntriesByAddress()
//  2. Find entries where ExpirationTimestamp has passed
//  3. Cancel using the stake's ID
//  4. Staked ZNN returns to your balance
//
// Parameters:
//   - id: Hash ID of the stake entry to cancel (from GetEntriesByAddress)
//
// Returns an unsigned AccountBlock template ready for processing.
//
// Example - Cancel expired stake:
//
//	// Get stake entries
//	stakes, _ := client.StakeApi.GetEntriesByAddress(address, 0, 10)
//
//	// Find expired stakes
//	now := time.Now().Unix()
//	for _, stake := range stakes.List {
//	    if stake.ExpirationTimestamp <= now {
//	        // This stake has expired, can cancel
//	        template := client.StakeApi.Cancel(stake.Id)
//	        // Sign and publish transaction
//	        fmt.Printf("Canceling stake: %s ZNN\n", stake.Amount)
//	    }
//	}
//
// Example - Cancel specific stake by ID:
//
//	stakeId := types.HexToHashPanic("0x123...")
//	template := client.StakeApi.Cancel(stakeId)
//	// Process and publish
//
// Note: Always check expiration timestamp before attempting to cancel. Canceling
// before expiration will fail.
func (sa *StakeApi) Cancel(id types.Hash) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.StakeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          definition.ABIStake.PackMethodPanic(definition.CancelStakeMethodName, id),
	}
}

// CollectReward creates a transaction template to claim accumulated staking rewards.
//
// Staking rewards (both ZNN and QSR) accumulate automatically but must be explicitly
// collected. This transaction withdraws all uncollected rewards to your balance.
//
// Reward collection:
//   - Collects both ZNN and QSR rewards simultaneously
//   - No minimum amount required to collect
//   - Can collect as frequently as desired
//   - Rewards continue accruing after collection
//
// Typical collection strategies:
//   - Regular collection (e.g., weekly/monthly)
//   - Collect when rewards reach target threshold
//   - Compound by restaking collected ZNN
//   - Collect before stake cancellation
//
// Returns an unsigned AccountBlock template ready for processing.
//
// Example - Basic reward collection:
//
//	// Check uncollected rewards
//	rewards, _ := client.StakeApi.GetUncollectedReward(myAddress)
//
//	if rewards.Znn.Cmp(big.NewInt(0)) > 0 || rewards.Qsr.Cmp(big.NewInt(0)) > 0 {
//	    fmt.Printf("Collecting: %s ZNN, %s QSR\n", rewards.Znn, rewards.Qsr)
//	    template := client.StakeApi.CollectReward()
//	    // Sign and publish transaction
//	}
//
// Example - Collect before canceling stakes:
//
//	// Always collect rewards before canceling stakes
//	collectTemplate := client.StakeApi.CollectReward()
//	// Publish collect transaction
//
//	// Then cancel expired stakes
//	// This ensures no rewards are left unclaimed
//
// Example - Compound rewards (collect and restake):
//
//	rewards, _ := client.StakeApi.GetUncollectedReward(myAddress)
//	if rewards.Znn.Cmp(big.NewInt(100000000)) > 0 { // At least 1 ZNN
//	    // Collect rewards
//	    collectTemplate := client.StakeApi.CollectReward()
//	    // After collection confirms, restake the ZNN
//	    stakeTemplate := client.StakeApi.Stake(31536000, rewards.Znn)
//	}
//
// Note: Collection requires a small amount of PoW/plasma. Ensure you have sufficient
// resources before attempting to collect.
func (sa *StakeApi) CollectReward() *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.StakeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          definition.ABIStake.PackMethodPanic(definition.CollectRewardMethodName),
	}
}
