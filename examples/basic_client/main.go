package main

import (
	"fmt"
	"time"

	"github.com/MoonBaZZe/znn-sdk-go/rpc_client"
)

func main() {
	fmt.Println("Zenon Go SDK - Basic Client Example")
	fmt.Println("====================================")

	// Connect to local node with default options
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer client.Stop()

	// Register connection callbacks
	client.AddOnConnectionEstablishedCallback(func() {
		fmt.Println("✓ Connected to Zenon node")
	})

	client.AddOnConnectionLostCallback(func(err error) {
		fmt.Printf("✗ Connection lost: %v\n", err)
	})

	fmt.Printf("Connection status: %s\n\n", client.Status())

	// Query frontier momentum
	fmt.Println("Querying frontier momentum...")
	momentum, err := client.LedgerApi.GetFrontierMomentum()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Current height: %d\n", momentum.Height)
	fmt.Printf("Momentum hash: %s\n", momentum.Hash.String())
	fmt.Printf("Timestamp: %d\n\n", momentum.Timestamp)

	// Query network info
	fmt.Println("Querying network info...")
	networkInfo, err := client.StatsApi.OsInfo()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("OS: %s\n", networkInfo.Os)
	fmt.Printf("Platform: %s\n", networkInfo.Platform)
	fmt.Printf("Platform version: %s\n\n", networkInfo.PlatformVersion)

	// Keep running for a bit to demonstrate callbacks
	fmt.Println("Monitoring connection for 5 seconds...")
	time.Sleep(5 * time.Second)

	fmt.Println("\n✓ Example completed successfully")
}
