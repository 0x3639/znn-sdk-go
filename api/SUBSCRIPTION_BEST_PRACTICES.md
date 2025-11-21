# WebSocket Subscription Best Practices

## Overview

This guide provides best practices for managing WebSocket subscriptions in the Zenon Go SDK to prevent resource leaks and ensure proper cleanup.

**Last Updated**: 2025-11-16

---

## The Problem: Resource Leaks

WebSocket subscriptions consume resources:
- **Network connections**: Open TCP sockets
- **Goroutines**: Background threads reading from the connection
- **Memory buffers**: Queued messages waiting to be processed
- **Server resources**: Node tracks each subscription

**If not properly cleaned up**, these resources accumulate and can lead to:
- ❌ Connection exhaustion (file descriptor limits)
- ❌ Memory leaks (buffered messages pile up)
- ❌ Goroutine leaks (threads never terminate)
- ❌ Node resource exhaustion (server-side tracking)

---

## Solution: SubscriptionManager

The SDK provides `SubscriptionManager` to centralize subscription lifecycle management.

### Basic Usage

```go
import (
    "context"
    "github.com/0x3639/znn-sdk-go/api"
)

func subscribeToMomentums() error {
    ctx := context.Background()

    // Create manager
    manager := api.NewSubscriptionManager()

    // CRITICAL: Always defer cleanup
    defer manager.UnsubscribeAll()

    // Subscribe to momentums
    momentumSub, err := client.SubscriberApi.ToMomentums(ctx)
    if err != nil {
        return err
    }

    // Add to manager for tracking
    manager.Add(momentumSub)

    // Subscribe to account blocks
    accountSub, err := client.SubscriberApi.ToAllAccountBlocks(ctx)
    if err != nil {
        return err
    }
    manager.Add(accountSub)

    // Process events...
    for {
        select {
        case momentum := <-momentumSub.Channel:
            processMomentum(momentum)
        case block := <-accountSub.Channel:
            processBlock(block)
        }
    }

    // Cleanup happens automatically via defer
}
```

---

## Pattern: Single Subscription

For a single subscription, use the `defer sub.Unsubscribe()` pattern:

```go
func watchFrontierMomentum() error {
    ctx := context.Background()
    sub, err := client.SubscriberApi.ToMomentums(ctx)
    if err != nil {
        return err
    }
    defer sub.Unsubscribe()  // ALWAYS defer cleanup

    // Process events
    for momentum := range sub.Channel {
        fmt.Printf("New momentum: %d\n", momentum.Height)
    }

    return nil
}
```

### ✅ Correct Pattern

```go
sub, _ := client.SubscriberApi.ToMomentums(ctx)
defer sub.Unsubscribe()  // ✅ Cleanup guaranteed
```

### ❌ WRONG - No Cleanup

```go
sub, _ := client.SubscriberApi.ToMomentums(ctx)
// ❌ No defer - subscription leaks if function returns early
```

---

## Pattern: Multiple Subscriptions

For multiple subscriptions, use `SubscriptionManager`:

```go
func monitorNetwork() error {
    ctx := context.Background()
    manager := api.NewSubscriptionManager()
    defer manager.UnsubscribeAll()  // Single defer cleans up all

    // Subscribe to multiple event streams
    subs := []struct{
        name string
        sub  *server.ClientSubscription
        err  error
    }{
        {"momentums", client.SubscriberApi.ToMomentums(ctx)},
        {"blocks", client.SubscriberApi.ToAllAccountBlocks(ctx)},
        {"unconfirmed", client.SubscriberApi.ToAllUnconfirmedAccountBlocks(ctx)},
    }

    // Add all subscriptions to manager
    for _, s := range subs {
        if s.err != nil {
            return fmt.Errorf("failed to subscribe to %s: %w", s.name, s.err)
        }
        manager.Add(s.sub)
    }

    // Process events...
    // Cleanup happens automatically
}
```

### Benefits

✅ **Single cleanup point**: One `defer` for all subscriptions
✅ **Early return safety**: Cleanup happens even if function returns early
✅ **Panic safety**: Cleanup happens even if code panics
✅ **Readable**: Clear intent to clean up resources

---

## Pattern: Long-Running Services

For services that run indefinitely, clean up on shutdown:

```go
type MonitorService struct {
    client  *rpc_client.RpcClient
    manager *api.SubscriptionManager
    stop    chan struct{}
}

func (s *MonitorService) Start() error {
    ctx := context.Background()
    s.manager = api.NewSubscriptionManager()
    s.stop = make(chan struct{})

    // Subscribe to events
    sub, err := s.client.SubscriberApi.ToMomentums(ctx)
    if err != nil {
        return err
    }
    s.manager.Add(sub)

    // Process events in background
    go s.processEvents(sub)

    return nil
}

func (s *MonitorService) Stop() {
    close(s.stop)  // Signal shutdown
    s.manager.UnsubscribeAll()  // Clean up subscriptions
}

func (s *MonitorService) processEvents(sub *server.ClientSubscription) {
    for {
        select {
        case event := <-sub.Channel:
            // Process event
        case <-s.stop:
            return  // Shutdown signal
        }
    }
}
```

### Usage

```go
service := &MonitorService{client: rpcClient}

if err := service.Start(); err != nil {
    log.Fatal(err)
}

// Handle shutdown signal
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt)
<-sigChan

service.Stop()  // Clean shutdown
```

---

## Pattern: Testing

Always clean up in tests to prevent leaks between test cases:

```go
func TestSubscriptions(t *testing.T) {
    ctx := context.Background()
    client, _ := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
    defer client.Stop()

    manager := api.NewSubscriptionManager()
    defer manager.UnsubscribeAll()  // Test cleanup

    sub, err := client.SubscriberApi.ToMomentums(ctx)
    if err != nil {
        t.Fatal(err)
    }
    manager.Add(sub)

    // Test subscription behavior...
}
```

### Table-Driven Tests

```go
func TestMultipleSubscriptions(t *testing.T) {
    ctx := context.Background()
    tests := []struct {
        name string
        fn   func(*rpc_client.RpcClient) (*server.ClientSubscription, error)
    }{
        {"momentums", func(c *rpc_client.RpcClient) (*server.ClientSubscription, error) {
            return c.SubscriberApi.ToMomentums(ctx)
        }},
        {"blocks", func(c *rpc_client.RpcClient) (*server.ClientSubscription, error) {
            return c.SubscriberApi.ToAllAccountBlocks(ctx)
        }},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client, _ := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
            defer client.Stop()

            sub, err := tt.fn(client)
            if err != nil {
                t.Fatal(err)
            }
            defer sub.Unsubscribe()  // Per-test cleanup

            // Test logic...
        })
    }
}
```

---

## Common Mistakes

### ❌ Forgetting Cleanup

```go
func bad() {
    ctx := context.Background()
    sub, _ := client.SubscriberApi.ToMomentums(ctx)
    // ❌ No cleanup - subscription leaks

    for momentum := range sub.Channel {
        if momentum.Height > 1000 {
            return  // ❌ Early return leaks subscription
        }
    }
}
```

### ✅ Correct Version

```go
func good() {
    ctx := context.Background()
    sub, _ := client.SubscriberApi.ToMomentums(ctx)
    defer sub.Unsubscribe()  // ✅ Always cleaned up

    for momentum := range sub.Channel {
        if momentum.Height > 1000 {
            return  // ✅ Cleanup happens via defer
        }
    }
}
```

### ❌ Deferred in Loop

```go
func bad() {
    ctx := context.Background()
    for i := 0; i < 10; i++ {
        sub, _ := client.SubscriberApi.ToMomentums(ctx)
        defer sub.Unsubscribe()  // ❌ All defers execute at function end
                                  // Creates 10 subscriptions, cleans up at end
    }
}
```

### ✅ Correct Version

```go
func good() {
    ctx := context.Background()
    manager := api.NewSubscriptionManager()
    defer manager.UnsubscribeAll()

    for i := 0; i < 10; i++ {
        sub, _ := client.SubscriberApi.ToMomentums(ctx)
        manager.Add(sub)

        // Process events...

        // Optionally remove when done with this subscription
        manager.Remove(sub)
    }
}
```

### ❌ Ignoring Errors on Cleanup

```go
func bad() {
    ctx := context.Background()
    sub, _ := client.SubscriberApi.ToMomentums(ctx)
    sub.Unsubscribe()  // ❌ No defer - won't run if panic occurs
}
```

### ✅ Correct Version

```go
func good() {
    ctx := context.Background()
    sub, _ := client.SubscriberApi.ToMomentums(ctx)
    defer sub.Unsubscribe()  // ✅ Runs even if panic occurs
}
```

---

## SubscriptionManager API

### Core Methods

```go
// Create new manager
manager := api.NewSubscriptionManager()

// Add subscription for tracking
manager.Add(subscription)

// Add multiple subscriptions
manager.AddMultiple(sub1, sub2, sub3)

// Remove specific subscription
removed := manager.Remove(subscription)  // Returns true if found

// Clean up all subscriptions
manager.UnsubscribeAll()

// Check if empty
empty := manager.IsEmpty()

// Get subscription count
count := manager.Count()
```

### Thread Safety

`SubscriptionManager` is thread-safe and can be used concurrently:

```go
ctx := context.Background()
manager := api.NewSubscriptionManager()

// Safe to call from multiple goroutines
go func() {
    sub, _ := client.SubscriberApi.ToMomentums(ctx)
    manager.Add(sub)  // Thread-safe
}()

go func() {
    sub, _ := client.SubscriberApi.ToAllAccountBlocks(ctx)
    manager.Add(sub)  // Thread-safe
}()

// Safe to clean up from any goroutine
defer manager.UnsubscribeAll()
```

---

## Monitoring Subscriptions

### Check for Leaks

```go
// Before operation
initialCount := manager.Count()

// Perform operation
doSomething()

// After operation
if manager.Count() > initialCount {
    log.Printf("WARNING: Subscription leak detected! Count increased from %d to %d",
        initialCount, manager.Count())
}
```

### Periodic Health Check

```go
func monitorSubscriptions(manager *api.SubscriptionManager) {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        count := manager.Count()
        log.Printf("Active subscriptions: %d", count)

        if count > 100 {
            log.Printf("WARNING: High subscription count: %d", count)
        }
    }
}
```

---

## Debugging Subscription Issues

### Common Symptoms

1. **File descriptor exhaustion**
   ```
   too many open files
   ```
   → Check if subscriptions are being cleaned up

2. **Memory growth**
   ```
   go tool pprof -http=:8080 mem.prof
   ```
   → Look for leaked goroutines from subscriptions

3. **Connection refused**
   ```
   dial tcp: connect: connection refused
   ```
   → May indicate node resource exhaustion from leaked subscriptions

### Debug Pattern

```go
func debugSubscriptions() {
    manager := api.NewSubscriptionManager()
    defer func() {
        finalCount := manager.Count()
        log.Printf("Cleaning up %d subscriptions", finalCount)
        manager.UnsubscribeAll()
    }()

    // Operations...
    log.Printf("Active subscriptions: %d", manager.Count())
}
```

---

## Production Checklist

Before deploying code that uses subscriptions:

- [ ] Every subscription has a cleanup path (`defer sub.Unsubscribe()`)
- [ ] Multiple subscriptions use `SubscriptionManager`
- [ ] Long-running services have shutdown handlers
- [ ] Tests clean up subscriptions in defer
- [ ] Error paths don't skip cleanup
- [ ] Subscription count is monitored in production
- [ ] Logs show subscription creation/cleanup
- [ ] Connection limits are understood (default: unlimited)

---

## Performance Considerations

### Subscription Overhead

Each subscription consumes:
- **Memory**: ~1-10 KB per subscription (channel buffers)
- **Goroutines**: 1-2 per subscription (reader + processor)
- **Network**: Minimal (events pushed by server)

**Guideline**: < 100 subscriptions per client is reasonable

### Batch Operations

Instead of subscribing per-account, subscribe once and filter:

```go
// ❌ BAD: 1000 subscriptions
for _, account := range accounts {
    sub, _ := client.SubscriberApi.ToAccountBlocksByAddress(ctx, account)
    manager.Add(sub)
}

// ✅ GOOD: 1 subscription, filter client-side
sub, _ := client.SubscriberApi.ToAllAccountBlocks(ctx)
manager.Add(sub)

for block := range sub.Channel {
    if accountMap[block.Address] {
        processBlock(block)
    }
}
```

---

## Summary

**Golden Rule**: Always clean up subscriptions using `defer`.

**Simple Pattern** (1 subscription):
```go
sub, _ := client.SubscriberApi.ToMomentums(ctx)
defer sub.Unsubscribe()
```

**Complex Pattern** (multiple subscriptions):
```go
manager := api.NewSubscriptionManager()
defer manager.UnsubscribeAll()
manager.Add(sub1)
manager.Add(sub2)
```

**Result**: Zero resource leaks, predictable behavior, production-ready code.

---

## References

- API Documentation: `api/subscription_manager.go`
- Tests: `api/subscription_manager_test.go`
- WebSocket Client: `rpc_client/client.go`
- Subscriber API: `api/subscriber.go`

---

**Author**: Zenon Go SDK Contributors
**Last Updated**: 2025-11-16
**Version**: 1.0
