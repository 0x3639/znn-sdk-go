package api

import (
	"context"

	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/api/subscribe"
	"github.com/zenon-network/go-zenon/rpc/server"
)

type SubscriberApi struct {
	client *server.Client
}

func NewSubscriberApi(client *server.Client) *SubscriberApi {
	return &SubscriberApi{
		client: client,
	}
}

// ToMomentums subscribes to real-time momentum (block) production events.
//
// Momentums are Zenon's equivalent of blocks - each momentum contains a batch of
// confirmed account blocks. This subscription allows monitoring the blockchain as
// new momentums are produced by Pillars.
//
// Use cases:
//   - Monitor blockchain height and progress
//   - Detect new blocks in real-time
//   - Build block explorers and analytics
//   - Trigger actions on new momentums
//
// Parameters:
//   - ctx: Context for cancellation and timeout control. The subscription will
//     be automatically cancelled when this context is cancelled.
//
// Returns:
//   - ClientSubscription: Subscription handle for management
//   - Channel: Receives arrays of new Momentum events
//   - Error: If subscription fails
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//
//	sub, momentumChan, err := client.SubscriberApi.ToMomentums(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer sub.Unsubscribe()
//
//	for momentums := range momentumChan {
//	    for _, m := range momentums {
//	        fmt.Printf("New momentum: Height %d, Hash %s\n", m.Height, m.Hash)
//	    }
//	}
//
// Note: The subscription will stop when ctx is cancelled or Unsubscribe() is called.
func (sa *SubscriberApi) ToMomentums(ctx context.Context) (*server.ClientSubscription, chan []subscribe.Momentum, error) {
	ch := make(chan []subscribe.Momentum)
	subscription, err := sa.client.Subscribe(ctx, "ledger", ch, "momentums")
	if err != nil {
		return nil, nil, err
	}
	return subscription, ch, err
}

// ToAllAccountBlocks subscribes to all account block events across the entire network.
//
// This provides a real-time stream of every transaction on Zenon Network as they
// are confirmed. Use with caution - high volume on active networks.
//
// Use cases:
//   - Network-wide transaction monitoring
//   - Building comprehensive block explorers
//   - Analytics and metrics collection
//   - Detecting protocol-level events
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - ClientSubscription: Subscription handle for management
//   - Channel: Receives arrays of AccountBlock events
//   - Error: If subscription fails
//
// Example:
//
//	ctx := context.Background()
//	sub, blockChan, err := client.SubscriberApi.ToAllAccountBlocks(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer sub.Unsubscribe()
//
//	for blocks := range blockChan {
//	    for _, block := range blocks {
//	        fmt.Printf("Block: %s -> %s, Amount: %s\n",
//	            block.Address, block.ToAddress, block.Amount)
//	    }
//	}
//
// Warning: This subscription can generate high data volume on busy networks.
// Consider using ToAccountBlocksByAddress for specific addresses instead.
func (sa *SubscriberApi) ToAllAccountBlocks(ctx context.Context) (*server.ClientSubscription, chan []subscribe.AccountBlock, error) {
	ch := make(chan []subscribe.AccountBlock)
	subscription, err := sa.client.Subscribe(ctx, "ledger", ch, "allAccountBlocks")
	if err != nil {
		return nil, nil, err
	}
	return subscription, ch, err
}

// ToAccountBlocksByAddress subscribes to account block events for a specific address.
//
// Monitors all transactions (both send and receive blocks) for a single address.
// This is the most common subscription pattern for wallet and application monitoring.
//
// Use cases:
//   - Wallet transaction notifications
//   - Payment processing confirmations
//   - Account activity monitoring
//   - Real-time balance tracking
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - address: Address to monitor for transactions
//
// Returns:
//   - ClientSubscription: Subscription handle for management
//   - Channel: Receives arrays of AccountBlock events for this address
//   - Error: If subscription fails
//
// Example:
//
//	ctx := context.Background()
//	myAddress := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")
//	sub, blockChan, err := client.SubscriberApi.ToAccountBlocksByAddress(ctx, myAddress)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer sub.Unsubscribe()
//
//	for blocks := range blockChan {
//	    for _, block := range blocks {
//	        if block.BlockType == nom.BlockTypeUserReceive {
//	            fmt.Printf("Received: %s from %s\n", block.Amount, block.Address)
//	        } else {
//	            fmt.Printf("Sent: %s to %s\n", block.Amount, block.ToAddress)
//	        }
//	    }
//	}
//
// This is ideal for monitoring a single wallet or application address.
func (sa *SubscriberApi) ToAccountBlocksByAddress(ctx context.Context, address types.Address) (*server.ClientSubscription, chan []subscribe.AccountBlock, error) {
	ch := make(chan []subscribe.AccountBlock)
	subscription, err := sa.client.Subscribe(ctx, "ledger", ch, "accountBlocksByAddress", address.String())
	if err != nil {
		return nil, nil, err
	}
	return subscription, ch, err
}

// ToUnreceivedAccountBlocksByAddress subscribes to incoming unreceived blocks for an address.
//
// Monitors specifically for send blocks directed to this address that haven't been
// received yet. Perfect for payment processing where you need to know when funds
// arrive and automatically create receive blocks.
//
// Use cases:
//   - Payment gateway implementations
//   - Auto-receive transaction automation
//   - Incoming payment notifications
//   - Fund monitoring for services
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - address: Address to monitor for incoming transactions
//
// Returns:
//   - ClientSubscription: Subscription handle for management
//   - Channel: Receives arrays of unreceived AccountBlock events
//   - Error: If subscription fails
//
// Example - Auto-receive payments:
//
//	ctx := context.Background()
//	myAddress := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")
//	sub, blockChan, err := client.SubscriberApi.ToUnreceivedAccountBlocksByAddress(ctx, myAddress)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer sub.Unsubscribe()
//
//	for blocks := range blockChan {
//	    for _, block := range blocks {
//	        fmt.Printf("Incoming payment: %s %s from %s\n",
//	            block.Amount, block.TokenStandard, block.Address)
//
//	        // Create receive block automatically
//	        receiveTemplate := client.LedgerApi.ReceiveTemplate(block.Hash)
//	        // Autofill, sign, and publish receive block...
//	    }
//	}
//
// This is essential for automated payment processing and wallet auto-receive features.
func (sa *SubscriberApi) ToUnreceivedAccountBlocksByAddress(ctx context.Context, address types.Address) (*server.ClientSubscription, chan []subscribe.AccountBlock, error) {
	ch := make(chan []subscribe.AccountBlock)
	subscription, err := sa.client.Subscribe(ctx, "ledger", ch, "unreceivedAccountBlocksByAddress", address.String())
	if err != nil {
		return nil, nil, err
	}
	return subscription, ch, err
}
