package rpc_client

import "testing"

// =============================================================================
// WebsocketStatus Tests
// =============================================================================

func TestWebsocketStatus_String(t *testing.T) {
	tests := []struct {
		status   WebsocketStatus
		expected string
	}{
		{Uninitialized, "Uninitialized"},
		{Connecting, "Connecting"},
		{Running, "Running"},
		{Stopped, "Stopped"},
		{WebsocketStatus(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.status.String()
		if got != tt.expected {
			t.Errorf("WebsocketStatus(%d).String() = %s, want %s", tt.status, got, tt.expected)
		}
	}
}

func TestWebsocketStatus_Values(t *testing.T) {
	// Test that constants have expected values
	if Uninitialized != 0 {
		t.Errorf("Uninitialized = %d, want 0", Uninitialized)
	}
	if Connecting != 1 {
		t.Errorf("Connecting = %d, want 1", Connecting)
	}
	if Running != 2 {
		t.Errorf("Running = %d, want 2", Running)
	}
	if Stopped != 3 {
		t.Errorf("Stopped = %d, want 3", Stopped)
	}
}
