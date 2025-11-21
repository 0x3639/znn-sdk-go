package api_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/0x3639/znn-sdk-go/rpc_client"
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
)

// Example_subscribeMomentums demonstrates subscribing to new momentums.
func Example_subscribeMomentums() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Subscribe to momentum events
	sub, momentumChan, err := client.SubscriberApi.ToMomentums(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	fmt.Println("Monitoring new momentums...")
	fmt.Println("Press Ctrl+C to stop")

	// Monitor momentums
	timeout := time.After(30 * time.Second)

	for {
		select {
		case momentums := <-momentumChan:
			for _, m := range momentums {
				fmt.Printf("\nNew Momentum:\n")
				fmt.Printf("  Height: %d\n", m.Height)
				fmt.Printf("  Hash: %s...\n", m.Hash.String()[:16])
			}
		case <-timeout:
			fmt.Println("\nMonitoring complete")
			return
		}
	}
}

// Example_monitorAccountActivity demonstrates monitoring a specific address.
func Example_monitorAccountActivity() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	ctx := context.Background()
	myAddress := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Subscribe to account blocks
	sub, blockChan, err := client.SubscriberApi.ToAccountBlocksByAddress(ctx, myAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	fmt.Printf("Monitoring activity for: %s\n", myAddress)
	fmt.Println("Waiting for transactions...")

	timeout := time.After(30 * time.Second)

	for {
		select {
		case blocks := <-blockChan:
			for _, block := range blocks {
				switch block.BlockType {
				case nom.BlockTypeUserReceive:
					fmt.Printf("\n✓ Received transaction\n")
					fmt.Printf("  Height: %d\n", block.Height)
					fmt.Printf("  From: %s\n", block.Address)
				case nom.BlockTypeUserSend:
					fmt.Printf("\n→ Sent transaction\n")
					fmt.Printf("  Height: %d\n", block.Height)
					fmt.Printf("  To: %s\n", block.ToAddress)
				}
				fmt.Printf("  Hash: %s...\n", block.Hash.String()[:16])
			}
		case <-timeout:
			fmt.Println("\nMonitoring stopped")
			return
		}
	}
}

// Example_autoReceivePayments demonstrates automated payment receiving.
func Example_autoReceivePayments() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	ctx := context.Background()

	myAddress := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Subscribe to unreceived blocks
	sub, blockChan, err := client.SubscriberApi.ToUnreceivedAccountBlocksByAddress(ctx, myAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	fmt.Println("Auto-receive service started")
	fmt.Printf("Monitoring: %s\n", myAddress)

	timeout := time.After(30 * time.Second)

	for {
		select {
		case blocks := <-blockChan:
			for _, block := range blocks {
				fmt.Printf("\nIncoming payment detected:\n")
				fmt.Printf("  From: %s\n", block.Address)
				fmt.Printf("  Height: %d\n", block.Height)
				fmt.Printf("  Block: %s...\n", block.Hash.String()[:16])

				// Create receive block template
				_ = client.LedgerApi.ReceiveTemplate(block.Hash)

				fmt.Println("  → Created receive transaction")
				fmt.Println("  → Sign and publish to complete reception")
			}
		case <-timeout:
			fmt.Println("\nAuto-receive service stopped")
			return
		}
	}
}

// Example_networkActivityMonitor demonstrates monitoring all network transactions.
func Example_networkActivityMonitor() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	ctx := context.Background()

	// Subscribe to all account blocks (high volume!)
	sub, blockChan, err := client.SubscriberApi.ToAllAccountBlocks(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	fmt.Println("Network Activity Monitor")
	fmt.Println("Tracking all transactions...")

	txCount := 0
	timeout := time.After(10 * time.Second)

	for {
		select {
		case blocks := <-blockChan:
			for _, block := range blocks {
				txCount++
				fmt.Printf("[%d] %s -> %s (Height: %d)\n",
					txCount,
					block.Address.String()[:16]+"...",
					block.ToAddress.String()[:16]+"...",
					block.Height)
			}
		case <-timeout:
			fmt.Printf("\nTotal transactions observed: %d\n", txCount)
			return
		}
	}
}

// Example_paymentGateway demonstrates a simple payment gateway implementation.
func Example_paymentGateway() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	ctx := context.Background()

	gatewayAddress := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Subscribe to incoming payments
	sub, blockChan, err := client.SubscriberApi.ToUnreceivedAccountBlocksByAddress(ctx, gatewayAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	fmt.Println("Payment Gateway Active")
	fmt.Printf("Merchant Address: %s\n", gatewayAddress)
	fmt.Println("\nWaiting for customer payments...")

	timeout := time.After(30 * time.Second)

	for {
		select {
		case blocks := <-blockChan:
			for _, block := range blocks {
				// Process payment
				fmt.Printf("\n💰 Payment Received!\n")
				fmt.Printf("   Customer: %s\n", block.Address)
				fmt.Printf("   Height: %d\n", block.Height)
				fmt.Printf("   Transaction: %s\n", block.Hash)

				// In production:
				// 1. Verify payment amount matches order
				// 2. Create receive block
				// 3. Update order status
				// 4. Send confirmation to customer

				fmt.Println("   → Processing order...")
				fmt.Println("   → Creating receive block...")
				fmt.Println("   → Updating database...")
			}
		case <-timeout:
			fmt.Println("\nPayment gateway demo complete")
			return
		}
	}
}

// Example_blockExplorer demonstrates building a simple block explorer.
func Example_blockExplorer() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	ctx := context.Background()

	// Subscribe to momentums
	sub, momentumChan, err := client.SubscriberApi.ToMomentums(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	fmt.Println("Block Explorer - Live Feed")
	fmt.Println("=" + string(make([]byte, 50)))

	timeout := time.After(20 * time.Second)

	for {
		select {
		case momentums := <-momentumChan:
			for _, m := range momentums {
				fmt.Printf("\n📦 Block #%d\n", m.Height)
				fmt.Printf("   Hash: %s\n", m.Hash)

				// In a real explorer:
				// - Fetch detailed momentum info
				// - Display contained transactions
				// - Show producer statistics
				// - Calculate network metrics
			}
		case <-timeout:
			fmt.Println("\n" + string(make([]byte, 50)))
			fmt.Println("Explorer feed stopped")
			return
		}
	}
}

// Example_multipleSubscriptions demonstrates managing multiple subscriptions.
func Example_multipleSubscriptions() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	ctx := context.Background()

	myAddress := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Subscribe to momentums
	momentumSub, momentumChan, err := client.SubscriberApi.ToMomentums(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer momentumSub.Unsubscribe()

	// Subscribe to account activity
	accountSub, accountChan, err := client.SubscriberApi.ToAccountBlocksByAddress(ctx, myAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer accountSub.Unsubscribe()

	fmt.Println("Multi-Subscription Monitor")
	fmt.Println("Tracking both momentums and account activity")

	timeout := time.After(30 * time.Second)

	for {
		select {
		case momentums := <-momentumChan:
			for _, m := range momentums {
				fmt.Printf("\n[MOMENTUM] Height: %d\n", m.Height)
			}

		case blocks := <-accountChan:
			for _, block := range blocks {
				fmt.Printf("\n[ACCOUNT] Transaction at height: %d\n", block.Height)
			}

		case <-timeout:
			fmt.Println("\nMonitoring complete")
			return
		}
	}
}

// Example_subscriptionLifecycle demonstrates proper subscription management.
func Example_subscriptionLifecycle() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	ctx := context.Background()

	fmt.Println("Subscription Lifecycle Demo")

	// Create subscription
	fmt.Println("\n1. Creating subscription...")
	sub, momentumChan, err := client.SubscriberApi.ToMomentums(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("2. Subscription active")

	// Use subscription briefly
	timeout := time.After(5 * time.Second)
	received := false

	for !received {
		select {
		case momentums := <-momentumChan:
			fmt.Printf("3. Received %d momentum(s)\n", len(momentums))
			received = true
		case <-timeout:
			fmt.Println("3. Timeout waiting for momentums")
			received = true
		}
	}

	// Clean up
	fmt.Println("4. Unsubscribing...")
	sub.Unsubscribe()
	fmt.Println("5. Subscription closed")

	fmt.Println("\nBest practices:")
	fmt.Println("- Always call Unsubscribe() when done")
	fmt.Println("- Use defer for automatic cleanup")
	fmt.Println("- Handle channel closure gracefully")
	fmt.Println("- Set timeouts to prevent blocking")
}
