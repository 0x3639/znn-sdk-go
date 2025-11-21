package embedded_test

import (
	"fmt"
	"log"
	"math/big"

	"github.com/0x3639/znn-sdk-go/rpc_client"
	"github.com/zenon-network/go-zenon/common/types"
)

// Example_listActiveSentinels demonstrates listing all active Sentinels.
func Example_listActiveSentinels() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Get all active sentinels
	sentinels, err := client.SentinelApi.GetAllActive(0, 50)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Active Sentinels: %d\n\n", sentinels.Count)

	for i, sentinel := range sentinels.List {
		fmt.Printf("%d. Owner: %s\n", i+1, sentinel.Owner)
		fmt.Printf("   Registration: Momentum %d\n", sentinel.RegistrationTimestamp)
		fmt.Printf("   Active: %t\n", sentinel.Active)
	}
}

// Example_registerSentinel demonstrates registering a new Sentinel.
func Example_registerSentinel() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Create Sentinel registration transaction
	_ = client.SentinelApi.Register()

	fmt.Println("Sentinel Registration Transaction Created")
	fmt.Println("\nRequirements:")
	fmt.Println("- 5,000 ZNN (locked as collateral)")
	fmt.Println("- 50,000 QSR (locked as collateral)")
	fmt.Println("- Running Sentinel node infrastructure")

	fmt.Println("\nBenefits:")
	fmt.Println("- Earn ZNN and QSR rewards")
	fmt.Println("- Support network infrastructure")
	fmt.Println("- Lower barrier than Pillar")
	fmt.Println("- Full collateral return on revocation")
}

// Example_checkSentinelInfo demonstrates querying Sentinel information.
func Example_checkSentinelInfo() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	sentinelOwner := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get sentinel info
	sentinel, err := client.SentinelApi.GetByOwner(sentinelOwner)
	if err != nil {
		log.Fatal(err)
	}

	if sentinel.Owner.String() != "z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f" {
		fmt.Printf("Sentinel Owner: %s\n", sentinel.Owner)
		fmt.Printf("Revocable: %t\n", sentinel.CanBeRevoked)
		fmt.Printf("Registration: Momentum %d\n", sentinel.RegistrationTimestamp)
		fmt.Printf("Active: %t\n", sentinel.Active)
	} else {
		fmt.Println("No Sentinel registered for this address")
	}
}

// Example_checkSentinelRewards demonstrates checking Sentinel rewards.
func Example_checkSentinelRewards() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	sentinelOwner := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get uncollected rewards
	rewards, err := client.SentinelApi.GetUncollectedReward(sentinelOwner)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Sentinel Rewards")
	fmt.Printf("Uncollected ZNN: %s\n", rewards.Znn)
	fmt.Printf("Uncollected QSR: %s\n", rewards.Qsr)

	if rewards.Znn.Cmp(big.NewInt(0)) > 0 || rewards.Qsr.Cmp(big.NewInt(0)) > 0 {
		fmt.Println("\nRewards available for collection!")
	}
}

// Example_collectSentinelRewards demonstrates collecting Sentinel rewards.
func Example_collectSentinelRewards() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Create reward collection transaction
	_ = client.SentinelApi.CollectReward()

	fmt.Println("Sentinel reward collection transaction created")
	fmt.Println("Collects both ZNN and QSR rewards")
	fmt.Println("\nSentinel rewards come from:")
	fmt.Println("- Network infrastructure participation")
	fmt.Println("- Protocol-level reward distribution")
}

// Example_depositQsrForSentinel demonstrates depositing QSR for a Sentinel.
func Example_depositQsrForSentinel() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Deposit 10,000 QSR
	amount := big.NewInt(10000 * 100000000)

	_ = client.SentinelApi.DepositQsr(amount)

	fmt.Println("QSR Deposit Transaction Created")
	fmt.Printf("Depositing: %s QSR\n", amount)
	fmt.Println("\nDeposit benefits:")
	fmt.Println("- Increase Sentinel rewards weight")
	fmt.Println("- Strengthen network participation")
	fmt.Println("- Can withdraw after lock period")
}

// Example_withdrawQsrFromSentinel demonstrates withdrawing deposited QSR.
func Example_withdrawQsrFromSentinel() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	sentinelOwner := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Check deposited QSR first
	deposited, err := client.SentinelApi.GetDepositedQsr(sentinelOwner)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Deposited QSR: %s\n", deposited)

	if deposited.Cmp(big.NewInt(0)) > 0 {
		// Create withdrawal transaction
		_ = client.SentinelApi.WithdrawQsr()

		fmt.Println("\nWithdrawal transaction created")
		fmt.Println("All deposited QSR will be returned")
	} else {
		fmt.Println("\nNo QSR deposited to withdraw")
	}
}

// Example_revokeSentinel demonstrates revoking a Sentinel.
func Example_revokeSentinel() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Create revocation transaction
	_ = client.SentinelApi.Revoke()

	fmt.Println("Sentinel Revocation Transaction Created")
	fmt.Println("\nRevocation process:")
	fmt.Println("- Sentinel stops participating")
	fmt.Println("- Collateral enters cooldown period")
	fmt.Println("- After cooldown: 5,000 ZNN returned")
	fmt.Println("- After cooldown: 50,000 QSR returned")

	fmt.Println("\nNote: Must wait for cooldown before collateral is released")
}

// Example_viewSentinelRewardHistory demonstrates viewing reward collection history.
func Example_viewSentinelRewardHistory() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	sentinelOwner := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get reward history
	history, err := client.SentinelApi.GetFrontierRewardByPage(sentinelOwner, 0, 25)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Reward Collection History: %d entries\n\n", history.Count)

	if history.Count > 0 {
		totalZnn := big.NewInt(0)
		totalQsr := big.NewInt(0)

		for i, entry := range history.List {
			totalZnn.Add(totalZnn, entry.Znn)
			totalQsr.Add(totalQsr, entry.Qsr)

			fmt.Printf("%d. Epoch %d\n", i+1, entry.Epoch)
			fmt.Printf("   ZNN: %s, QSR: %s\n", entry.Znn, entry.Qsr)
		}

		fmt.Printf("\nTotal collected:\n")
		fmt.Printf("  ZNN: %s\n", totalZnn)
		fmt.Printf("  QSR: %s\n", totalQsr)
	} else {
		fmt.Println("No reward collections yet")
	}
}
