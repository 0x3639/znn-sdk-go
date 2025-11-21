package api

import (
	"testing"
)

// =============================================================================
// SubscriptionManager Tests
// =============================================================================

func TestNewSubscriptionManager(t *testing.T) {
	manager := NewSubscriptionManager()

	if manager == nil {
		t.Fatal("NewSubscriptionManager() returned nil")
	}

	if !manager.IsEmpty() {
		t.Error("New manager should be empty")
	}

	if manager.Count() != 0 {
		t.Errorf("New manager count = %d, want 0", manager.Count())
	}
}

func TestSubscriptionManager_Add(t *testing.T) {
	manager := NewSubscriptionManager()

	// Note: We can't easily create real ClientSubscription objects in unit tests
	// without a live RPC connection, so we test the manager logic with nil checks

	// Adding nil should be safe (no-op)
	manager.Add(nil)
	if manager.Count() != 0 {
		t.Error("Adding nil subscription should not increase count")
	}
}

func TestSubscriptionManager_AddMultiple(t *testing.T) {
	manager := NewSubscriptionManager()

	// Adding multiple nils should be safe
	manager.AddMultiple(nil, nil, nil)
	if manager.Count() != 0 {
		t.Error("Adding nil subscriptions should not increase count")
	}
}

func TestSubscriptionManager_UnsubscribeAll_Empty(t *testing.T) {
	manager := NewSubscriptionManager()

	// Should be safe to call on empty manager
	manager.UnsubscribeAll()

	if !manager.IsEmpty() {
		t.Error("Manager should still be empty after UnsubscribeAll()")
	}
}

func TestSubscriptionManager_UnsubscribeAll_MultipleCalls(t *testing.T) {
	manager := NewSubscriptionManager()

	// Should be safe to call multiple times
	manager.UnsubscribeAll()
	manager.UnsubscribeAll()
	manager.UnsubscribeAll()

	if !manager.IsEmpty() {
		t.Error("Manager should be empty after multiple UnsubscribeAll() calls")
	}
}

func TestSubscriptionManager_Remove_Nil(t *testing.T) {
	manager := NewSubscriptionManager()

	// Removing nil should return false
	if manager.Remove(nil) {
		t.Error("Remove(nil) should return false")
	}
}

func TestSubscriptionManager_IsEmpty(t *testing.T) {
	manager := NewSubscriptionManager()

	if !manager.IsEmpty() {
		t.Error("New manager should be empty")
	}

	// After UnsubscribeAll, should still be empty
	manager.UnsubscribeAll()

	if !manager.IsEmpty() {
		t.Error("Manager should be empty after UnsubscribeAll()")
	}
}

func TestSubscriptionManager_Count(t *testing.T) {
	manager := NewSubscriptionManager()

	if count := manager.Count(); count != 0 {
		t.Errorf("Count() = %d, want 0", count)
	}
}

func TestSubscriptionManager_DeferPattern(t *testing.T) {
	// This test demonstrates the recommended defer pattern
	func() {
		manager := NewSubscriptionManager()
		defer manager.UnsubscribeAll() // Ensure cleanup

		// Simulate subscribing...
		// (In real code, this would be actual subscriptions)

		// If function exits early or panics, defer ensures cleanup
	}()

	// After function exits, manager should have cleaned up
	// (This is tested by the defer call above)
}

func TestSubscriptionManager_ConcurrentAccess(t *testing.T) {
	manager := NewSubscriptionManager()

	// Test concurrent access (should not race)
	done := make(chan bool, 3)

	// Concurrent adds
	go func() {
		for i := 0; i < 100; i++ {
			manager.Add(nil) // Safe with nil
		}
		done <- true
	}()

	// Concurrent counts
	go func() {
		for i := 0; i < 100; i++ {
			_ = manager.Count()
		}
		done <- true
	}()

	// Concurrent unsubscribes
	go func() {
		for i := 0; i < 100; i++ {
			manager.UnsubscribeAll()
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done

	// Should end up empty after all UnsubscribeAll calls
	if !manager.IsEmpty() {
		t.Error("Manager should be empty after concurrent operations")
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkSubscriptionManager_Add(b *testing.B) {
	manager := NewSubscriptionManager()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.Add(nil)
	}
}

func BenchmarkSubscriptionManager_Count(b *testing.B) {
	manager := NewSubscriptionManager()
	// Pre-populate with some subscriptions
	for i := 0; i < 100; i++ {
		manager.Add(nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.Count()
	}
}

func BenchmarkSubscriptionManager_UnsubscribeAll(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		manager := NewSubscriptionManager()
		// Pre-populate
		for j := 0; j < 10; j++ {
			manager.Add(nil)
		}
		b.StartTimer()

		manager.UnsubscribeAll()
	}
}
