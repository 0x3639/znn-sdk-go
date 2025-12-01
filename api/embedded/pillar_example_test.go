package embedded_test

import (
	"fmt"
	"log"

	"github.com/0x3639/znn-sdk-go/rpc_client"
	"github.com/zenon-network/go-zenon/common/types"
)

// Example_listAllPillars demonstrates listing all network Pillars.
func Example_listAllPillars() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Get all pillars
	pillars, err := client.PillarApi.GetAll(0, 50)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total Pillars: %d\n\n", pillars.Count)

	for i, pillar := range pillars.List {
		fmt.Printf("%d. %s\n", i+1, pillar.Name)
		fmt.Printf("   Weight: %s\n", pillar.Weight)
		fmt.Printf("   Rank: %d\n", pillar.Rank)
	}
}

// Example_checkPillarNameAvailability demonstrates checking if a name is available.
func Example_checkPillarNameAvailability() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	name := "MyNewPillar"

	// Check if name is available
	available, err := client.PillarApi.CheckNameAvailability(name)
	if err != nil {
		log.Fatal(err)
	}

	if *available {
		fmt.Printf("Name '%s' is available\n", name)
		fmt.Println("You can register a Pillar with this name")
	} else {
		fmt.Printf("Name '%s' is already taken\n", name)
		fmt.Println("Choose a different name")
	}
}

// Example_delegateToP illar demonstrates delegating to a Pillar.
func Example_delegateToPillar() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	pillarName := "MyFavoritePillar"

	// Create delegation transaction
	_ = client.PillarApi.Delegate(pillarName)

	fmt.Println("Delegation transaction created")
	fmt.Printf("Delegating to: %s\n", pillarName)
	fmt.Println("\nBenefits:")
	fmt.Println("- Earn rewards from pillar")
	fmt.Println("- Support network decentralization")
	fmt.Println("- ZNN stays in your wallet")
	fmt.Println("- Can change delegation anytime")
}

// Example_checkDelegation demonstrates checking your current delegation.
func Example_checkDelegation() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	address := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get current delegation
	delegation, err := client.PillarApi.GetDelegatedPillar(address)
	if err != nil {
		log.Fatal(err)
	}

	if delegation.Name != "" {
		fmt.Printf("Currently delegated to: %s\n", delegation.Name)
		fmt.Println("Earning delegation rewards")
	} else {
		fmt.Println("Not delegated to any Pillar")
		fmt.Println("Tip: Delegate to earn rewards")
	}
}

// Example_undelegateFromPillar demonstrates removing delegation.
func Example_undelegateFromPillar() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Create undelegation transaction
	_ = client.PillarApi.Undelegate()

	fmt.Println("Undelegation transaction created")
	fmt.Println("\nAfter undelegation:")
	fmt.Println("- Stop earning pillar rewards")
	fmt.Println("- Voting weight returns to unallocated")
	fmt.Println("- Can delegate to different pillar")
	fmt.Println("- No cooldown period required")
}

// Example_checkPillarRewards demonstrates checking Pillar operator rewards.
func Example_checkPillarRewards() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	pillarAddress := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get uncollected rewards
	rewards, err := client.PillarApi.GetUncollectedReward(pillarAddress)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Pillar Operator Rewards")
	fmt.Printf("Uncollected ZNN: %s\n", rewards.ZnnAmount)
	fmt.Printf("Uncollected QSR: %s\n", rewards.QsrAmount)
}

// Example_collectPillarRewards demonstrates collecting Pillar rewards.
func Example_collectPillarRewards() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Create reward collection transaction
	_ = client.PillarApi.CollectReward()

	fmt.Println("Pillar reward collection transaction created")
	fmt.Println("Collects both ZNN and QSR rewards")
	fmt.Println("\nReward sources:")
	fmt.Println("- Block production rewards")
	fmt.Println("- Delegation rewards")
	fmt.Println("- Momentum production participation")
}

// Example_getPillarByName demonstrates querying specific Pillar details.
func Example_getPillarByName() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	pillarName := "MyPillar"

	// Get pillar info
	pillar, err := client.PillarApi.GetByName(pillarName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Pillar: %s\n", pillar.Name)
	fmt.Printf("Owner: %s\n", pillar.OwnerAddress)
	fmt.Printf("Producer: %s\n", pillar.ProducerAddress)
	fmt.Printf("Reward: %s\n", pillar.WithdrawAddress)
	fmt.Printf("Weight: %s\n", pillar.Weight)
	fmt.Printf("Give Momentum Reward: %d%%\n", pillar.GiveMomentumRewardPercentage)
	fmt.Printf("Give Delegation Reward: %d%%\n", pillar.GiveDelegateRewardPercentage)
}

// Example_getOwnedPillars demonstrates listing Pillars owned by an address.
func Example_getOwnedPillars() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	owner := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get pillars owned by address
	pillars, err := client.PillarApi.GetByOwner(owner)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Pillars owned: %d\n", len(pillars))

	if len(pillars) > 0 {
		for _, pillar := range pillars {
			fmt.Printf("- %s (Weight: %s)\n", pillar.Name, pillar.Weight)
		}
	} else {
		fmt.Println("No pillars owned by this address")
	}
}
