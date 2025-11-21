package embedded_test

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/0x3639/znn-sdk-go/rpc_client"
	"github.com/zenon-network/go-zenon/common/types"
)

// Example_checkUncollectedRewards demonstrates checking pending staking rewards.
func Example_checkUncollectedRewards() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	address := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get uncollected rewards
	rewards, err := client.StakeApi.GetUncollectedReward(address)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Uncollected Rewards:\n")
	fmt.Printf("  ZNN: %s\n", rewards.Znn)
	fmt.Printf("  QSR: %s\n", rewards.Qsr)

	// Check if rewards are available to collect
	if rewards.Znn.Cmp(big.NewInt(0)) > 0 || rewards.Qsr.Cmp(big.NewInt(0)) > 0 {
		fmt.Println("\nRewards available for collection!")
	} else {
		fmt.Println("\nNo rewards to collect yet")
	}
}

// Example_listStakeEntries demonstrates viewing all active stakes.
func Example_listStakeEntries() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	address := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get stake entries
	stakes, err := client.StakeApi.GetEntriesByAddress(address, 0, 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Active stakes: %d\n", stakes.Count)

	if stakes.Count > 0 {
		totalStaked := big.NewInt(0)
		now := time.Now().Unix()

		fmt.Println("\nStake details:")
		for i, stake := range stakes.Entries {
			totalStaked.Add(totalStaked, stake.Amount)
			expired := stake.ExpirationTimestamp <= now

			fmt.Printf("%d. Amount: %s ZNN\n", i+1, stake.Amount)
			fmt.Printf("   Expires: %v\n", time.Unix(stake.ExpirationTimestamp, 0))
			fmt.Printf("   Status: %s\n", map[bool]string{true: "Expired (can cancel)", false: "Active"}[expired])
		}

		fmt.Printf("\nTotal staked: %s ZNN\n", totalStaked)
	} else {
		fmt.Println("No active stakes")
	}
}

// Example_stakeFor1Month demonstrates staking ZNN for 1 month.
func Example_stakeFor1Month() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Stake 100 ZNN for 1 month
	amount := big.NewInt(100 * 100000000)
	duration := int64(2592000) // 30 days in seconds

	template := client.StakeApi.Stake(duration, amount)

	fmt.Println("1-Month Stake Created")
	fmt.Printf("Amount: %s ZNN\n", amount)
	fmt.Printf("Duration: %d seconds (30 days)\n", duration)
	fmt.Printf("To: %s\n", template.ToAddress)

	fmt.Println("\nStake characteristics:")
	fmt.Println("- Shortest duration")
	fmt.Println("- Lowest reward multiplier")
	fmt.Println("- Fastest liquidity return")
	fmt.Println("- Good for testing or short-term holds")

	// Template must be autofilled, enhanced with PoW, signed, and published
}

// Example_stakeFor12Months demonstrates staking for maximum rewards.
func Example_stakeFor12Months() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Stake 1000 ZNN for 12 months (maximum duration)
	amount := big.NewInt(1000 * 100000000)
	duration := int64(31536000) // 365 days in seconds

	_ = client.StakeApi.Stake(duration, amount)

	fmt.Println("12-Month Stake Created")
	fmt.Printf("Amount: %s ZNN\n", amount)
	fmt.Printf("Duration: %d seconds (365 days)\n", duration)

	fmt.Println("\nMaximum rewards strategy:")
	fmt.Println("- Highest reward multiplier")
	fmt.Println("- Best for long-term holders")
	fmt.Println("- Maximizes ZNN and QSR earnings")
	fmt.Println("- Requires commitment to 1-year lockup")

	// Note: ZNN will be locked for the full 12 months
}

// Example_diversifiedStaking demonstrates creating multiple stakes.
func Example_diversifiedStaking() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	fmt.Println("Diversified Staking Strategy")
	fmt.Println("Creating multiple stakes with different durations:")

	// Short-term stake (1 month)
	_ = client.StakeApi.Stake(2592000, big.NewInt(100*100000000))
	fmt.Println("1. Short-term: 100 ZNN for 1 month")
	fmt.Println("   - Quick liquidity return")

	// Medium-term stake (6 months)
	_ = client.StakeApi.Stake(15552000, big.NewInt(300*100000000))
	fmt.Println("\n2. Medium-term: 300 ZNN for 6 months")
	fmt.Println("   - Balanced rewards/liquidity")

	// Long-term stake (12 months)
	_ = client.StakeApi.Stake(31536000, big.NewInt(600*100000000))
	fmt.Println("\n3. Long-term: 600 ZNN for 12 months")
	fmt.Println("   - Maximum rewards")

	fmt.Println("\nBenefits:")
	fmt.Println("- Staggered expiration dates")
	fmt.Println("- Risk diversification")
	fmt.Println("- Regular liquidity events")
	fmt.Println("- Optimized reward/flexibility balance")
}

// Example_collectRewards demonstrates claiming accumulated rewards.
func Example_collectRewards() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	address := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Check uncollected rewards first
	rewards, err := client.StakeApi.GetUncollectedReward(address)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Reward Collection")
	fmt.Printf("Uncollected ZNN: %s\n", rewards.Znn)
	fmt.Printf("Uncollected QSR: %s\n", rewards.Qsr)

	// Only collect if rewards are available
	if rewards.Znn.Cmp(big.NewInt(0)) > 0 || rewards.Qsr.Cmp(big.NewInt(0)) > 0 {
		_ = client.StakeApi.CollectReward()

		fmt.Println("\nReward collection transaction created")
		fmt.Println("After confirmation:")
		fmt.Println("- Rewards will be added to your balance")
		fmt.Println("- Uncollected amount resets to 0")
		fmt.Println("- New rewards continue accruing")
	} else {
		fmt.Println("\nNo rewards to collect yet")
		fmt.Println("Tip: Wait for rewards to accumulate before collecting")
	}
}

// Example_cancelExpiredStake demonstrates canceling a stake after expiration.
func Example_cancelExpiredStake() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	address := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get stake entries
	stakes, err := client.StakeApi.GetEntriesByAddress(address, 0, 10)
	if err != nil {
		log.Fatal(err)
	}

	if stakes.Count == 0 {
		fmt.Println("No stakes to cancel")
		return
	}

	// Find expired stakes
	now := time.Now().Unix()
	expiredCount := 0

	fmt.Println("Checking for expired stakes...")
	for _, stake := range stakes.Entries {
		if stake.ExpirationTimestamp <= now {
			expiredCount++
			_ = client.StakeApi.Cancel(stake.Id)

			fmt.Printf("Canceling expired stake:\n")
			fmt.Printf("  Amount: %s ZNN\n", stake.Amount)
			fmt.Printf("  Expired: %v\n", time.Unix(stake.ExpirationTimestamp, 0))
			fmt.Printf("  ID: %s...\n", stake.Id.String()[:16])
		}
	}

	if expiredCount == 0 {
		fmt.Println("No expired stakes found")
		fmt.Println("All stakes are still locked")
	} else {
		fmt.Printf("\nCancellation transactions created: %d\n", expiredCount)
		fmt.Println("After confirmation, ZNN will return to balance")
	}
}

// Example_viewRewardHistory demonstrates checking reward collection history.
func Example_viewRewardHistory() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	address := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get reward history
	history, err := client.StakeApi.GetFrontierRewardByPage(address, 0, 25)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Reward Collection History\n")
	fmt.Printf("Total collections: %d\n\n", history.Count)

	if history.Count > 0 {
		totalZnn := big.NewInt(0)
		totalQsr := big.NewInt(0)

		for i, entry := range history.List {
			totalZnn.Add(totalZnn, entry.Znn)
			totalQsr.Add(totalQsr, entry.Qsr)

			fmt.Printf("%d. Collected at epoch %d\n", i+1, entry.Epoch)
			fmt.Printf("   ZNN: %s, QSR: %s\n", entry.Znn, entry.Qsr)
		}

		fmt.Printf("\nTotal rewards collected:\n")
		fmt.Printf("  ZNN: %s\n", totalZnn)
		fmt.Printf("  QSR: %s\n", totalQsr)
	} else {
		fmt.Println("No reward collections yet")
	}
}

// Example_compoundRewards demonstrates collecting and restaking rewards.
func Example_compoundRewards() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	address := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Check uncollected rewards
	rewards, err := client.StakeApi.GetUncollectedReward(address)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Compound Staking Strategy")
	fmt.Printf("Uncollected ZNN: %s\n", rewards.Znn)
	fmt.Printf("Uncollected QSR: %s\n", rewards.Qsr)

	minStake := big.NewInt(100000000) // 1 ZNN minimum

	if rewards.Znn.Cmp(minStake) >= 0 {
		fmt.Println("\nStep 1: Collect rewards")
		_ = client.StakeApi.CollectReward()

		fmt.Println("Step 2: Restake collected ZNN")
		_ = client.StakeApi.Stake(31536000, rewards.Znn)

		fmt.Println("\nCompounding benefits:")
		fmt.Println("- Rewards earn rewards")
		fmt.Println("- Exponential growth over time")
		fmt.Println("- Maximizes long-term returns")
		fmt.Println("- Best for multi-year holders")
	} else {
		fmt.Printf("\nNot enough ZNN to compound yet")
		fmt.Printf("Minimum stake: %s ZNN\n", minStake)
		fmt.Println("Continue accumulating rewards")
	}
}

// Example_stakingDurationComparison demonstrates comparing different durations.
func Example_stakingDurationComparison() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	amount := big.NewInt(1000 * 100000000) // 1000 ZNN

	fmt.Println("Staking Duration Comparison")
	fmt.Printf("Amount: %s ZNN\n\n", amount)

	durations := []struct {
		name    string
		seconds int64
		days    int
	}{
		{"1 Month", 2592000, 30},
		{"3 Months", 7776000, 90},
		{"6 Months", 15552000, 180},
		{"12 Months", 31536000, 365},
	}

	for i, d := range durations {
		_ = client.StakeApi.Stake(d.seconds, amount)

		fmt.Printf("%d. %s (%d days)\n", i+1, d.name, d.days)
		fmt.Printf("   Duration: %d seconds\n", d.seconds)
		fmt.Printf("   Liquidity: Returns after %d days\n", d.days)

		// Reward multiplier increases with duration
		multiplier := float64(d.days) / 30.0
		fmt.Printf("   Reward multiplier: ~%.1fx\n\n", multiplier)
	}

	fmt.Println("Selection guide:")
	fmt.Println("- Short term: Need liquidity, willing to accept lower rewards")
	fmt.Println("- Medium term: Balance of rewards and flexibility")
	fmt.Println("- Long term: Maximum rewards, committed to holding")
}
