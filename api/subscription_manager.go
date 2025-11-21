package api

import (
	"sync"

	"github.com/zenon-network/go-zenon/rpc/server"
)

// SubscriptionManager manages multiple WebSocket subscriptions and provides
// batch operations to prevent memory leaks.
//
// Problem it solves:
// - Manual subscription management is error-prone
// - Forgotten unsubscribe() calls lead to goroutine and channel leaks
// - No easy way to track and clean up multiple subscriptions
//
// Usage:
//
//	manager := api.NewSubscriptionManager()
//	defer manager.UnsubscribeAll() // Ensure cleanup
//
//	// Add subscriptions
//	ctx := context.Background()
//	sub1, ch1, _ := client.SubscriberApi.ToMomentums(ctx)
//	manager.Add(sub1)
//
//	sub2, ch2, _ := client.SubscriberApi.ToAllAccountBlocks(ctx)
//	manager.Add(sub2)
//
//	// Process events...
//
//	// Cleanup (or use defer)
//	manager.UnsubscribeAll()
type SubscriptionManager struct {
	subscriptions []*server.ClientSubscription
	mu            sync.RWMutex
}

// NewSubscriptionManager creates a new subscription manager.
//
// Example:
//
//	manager := api.NewSubscriptionManager()
//	defer manager.UnsubscribeAll()
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		subscriptions: make([]*server.ClientSubscription, 0),
	}
}

// Add registers a subscription with the manager.
//
// The subscription will be unsubscribed when UnsubscribeAll() is called.
//
// Parameters:
//   - sub: ClientSubscription to manage
//
// Example:
//
//	ctx := context.Background()
//	sub, ch, err := client.SubscriberApi.ToMomentums(ctx)
//	if err != nil {
//	    return err
//	}
//	manager.Add(sub)
//
//	// Use subscription...
//	for momentum := range ch {
//	    fmt.Printf("New momentum: %d\n", momentum.Height)
//	}
func (sm *SubscriptionManager) Add(sub *server.ClientSubscription) {
	if sub == nil {
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.subscriptions = append(sm.subscriptions, sub)
}

// AddMultiple registers multiple subscriptions at once.
//
// Convenience method for adding multiple subscriptions.
//
// Parameters:
//   - subs: Variadic list of subscriptions to manage
//
// Example:
//
//	ctx := context.Background()
//	sub1, _, _ := client.SubscriberApi.ToMomentums(ctx)
//	sub2, _, _ := client.SubscriberApi.ToAllAccountBlocks(ctx)
//	sub3, _, _ := client.SubscriberApi.ToAccountBlocksByAddress(ctx, addr)
//
//	manager.AddMultiple(sub1, sub2, sub3)
func (sm *SubscriptionManager) AddMultiple(subs ...*server.ClientSubscription) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, sub := range subs {
		if sub != nil {
			sm.subscriptions = append(sm.subscriptions, sub)
		}
	}
}

// Remove unsubscribes and removes a specific subscription from management.
//
// Parameters:
//   - sub: Subscription to remove
//
// Returns true if the subscription was found and removed, false otherwise.
//
// Example:
//
//	// Stop managing a specific subscription
//	if manager.Remove(sub1) {
//	    fmt.Println("Subscription removed")
//	}
func (sm *SubscriptionManager) Remove(sub *server.ClientSubscription) bool {
	if sub == nil {
		return false
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, s := range sm.subscriptions {
		if s == sub {
			// Unsubscribe
			s.Unsubscribe()

			// Remove from list
			sm.subscriptions = append(sm.subscriptions[:i], sm.subscriptions[i+1:]...)
			return true
		}
	}

	return false
}

// UnsubscribeAll unsubscribes all managed subscriptions and clears the list.
//
// This method is safe to call multiple times and safe for concurrent use.
// It's recommended to use defer for automatic cleanup:
//
//	manager := api.NewSubscriptionManager()
//	defer manager.UnsubscribeAll()
//
// Example:
//
//	// Manual cleanup
//	manager.UnsubscribeAll()
//
//	// Or automatic cleanup
//	func processEvents() {
//	    manager := api.NewSubscriptionManager()
//	    defer manager.UnsubscribeAll()
//
//	    // ... subscribe and process ...
//	}
func (sm *SubscriptionManager) UnsubscribeAll() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Unsubscribe all
	for _, sub := range sm.subscriptions {
		if sub != nil {
			sub.Unsubscribe()
		}
	}

	// Clear the list
	sm.subscriptions = make([]*server.ClientSubscription, 0)
}

// Count returns the number of currently managed subscriptions.
//
// Example:
//
//	fmt.Printf("Managing %d subscriptions\n", manager.Count())
func (sm *SubscriptionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.subscriptions)
}

// IsEmpty returns true if no subscriptions are being managed.
//
// Example:
//
//	if manager.IsEmpty() {
//	    fmt.Println("No active subscriptions")
//	}
func (sm *SubscriptionManager) IsEmpty() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.subscriptions) == 0
}
