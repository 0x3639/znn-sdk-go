package embedded_test

import (
	"fmt"
	"log"
	"math/big"

	"github.com/0x3639/znn-sdk-go/rpc_client"
	"github.com/zenon-network/go-zenon/common/types"
)

// Example_listAcceleratorProjects demonstrates listing all Accelerator-Z projects.
func Example_listAcceleratorProjects() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Get all projects
	projects, err := client.AcceleratorApi.GetAll(0, 25)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total Accelerator Projects: %d\n\n", projects.Count)

	for i, project := range projects.List {
		fmt.Printf("%d. %s\n", i+1, project.Name)
		fmt.Printf("   Owner: %s\n", project.Owner)
		fmt.Printf("   Status: %d\n", project.Status)
		fmt.Printf("   ZNN Requested: %s\n", project.ZnnFundsNeeded)
		fmt.Printf("   QSR Requested: %s\n", project.QsrFundsNeeded)
	}
}

// Example_createAcceleratorProject demonstrates submitting a new project proposal.
func Example_createAcceleratorProject() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Project details
	name := "My Ecosystem Project"
	description := "Building a tool to enhance the Zenon ecosystem..."
	url := "https://myproject.com"

	// Funding request
	znnNeeded := big.NewInt(5000 * 100000000)  // 5,000 ZNN
	qsrNeeded := big.NewInt(50000 * 100000000) // 50,000 QSR

	_ = client.AcceleratorApi.CreateProject(
		name,
		description,
		url,
		znnNeeded,
		qsrNeeded,
	)

	fmt.Println("Accelerator Project Created")
	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Requesting: %s ZNN, %s QSR\n", znnNeeded, qsrNeeded)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Wait for Pillar voting period")
	fmt.Println("2. If approved, add project phases")
	fmt.Println("3. Complete milestones to receive funding")
}

// Example_voteOnAcceleratorProject demonstrates Pillar voting.
func Example_voteOnAcceleratorProject() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	projectId := types.HexToHashPanic("0x123...")
	pillarName := "MyPillar"

	// Vote yes on project
	_ = client.AcceleratorApi.VoteByName(projectId, pillarName, 1)

	fmt.Println("Vote Cast on Accelerator Project")
	fmt.Printf("Pillar: %s\n", pillarName)
	fmt.Printf("Project: %s\n", projectId)
	fmt.Println("Vote: YES")

	fmt.Println("\nVote options:")
	fmt.Println("  0 = Abstain")
	fmt.Println("  1 = Yes (approve)")
	fmt.Println("  2 = No (reject)")
}

// Example_getProjectVoteBreakdown demonstrates checking vote results.
func Example_getProjectVoteBreakdown() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	projectId := types.HexToHashPanic("0x123...")

	// Get vote breakdown
	breakdown, err := client.AcceleratorApi.GetVoteBreakdown(projectId)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Vote Breakdown for Project\n")
	fmt.Printf("ID: %s\n\n", projectId)
	fmt.Printf("Yes votes: %d\n", breakdown.Yes)
	fmt.Printf("No votes: %d\n", breakdown.No)
	fmt.Printf("Total votes: %d\n", breakdown.Total)
}

// Example_donateToAccelerator demonstrates donating to the Accelerator fund.
func Example_donateToAccelerator() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Donate 100 ZNN to Accelerator
	amount := big.NewInt(100 * 100000000)

	_ = client.AcceleratorApi.Donate(amount, types.ZnnTokenStandard)

	fmt.Println("Accelerator Donation")
	fmt.Printf("Amount: %s ZNN\n", amount)
	fmt.Println("\nDonations support ecosystem development")
	fmt.Println("Funds are used for approved projects")
}

// Example_listSporks demonstrates viewing protocol sporks.
func Example_listSporks() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Get all sporks
	sporks, err := client.SporkApi.GetAll(0, 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Protocol Sporks: %d\n\n", sporks.Count)

	for i, spork := range sporks.List {
		fmt.Printf("%d. %s\n", i+1, spork.Name)
		fmt.Printf("   ID: %s\n", spork.Id)
		fmt.Printf("   Activated: %t\n", spork.Activated)
	}

	fmt.Println("\nSporks enable protocol features without hard forks")
}
