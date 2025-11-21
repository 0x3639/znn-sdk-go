package rpc_client

import (
	"strings"
	"testing"
)

// =============================================================================
// ValidateWsConnectionURL Tests
// =============================================================================

func TestValidateWsConnectionURL_Valid(t *testing.T) {
	validURLs := []string{
		"ws://localhost:35998",
		"wss://localhost:35998",
		"ws://127.0.0.1:35998",
		"wss://example.com:443",
		"ws://example.com",
		"wss://example.com",
		"ws://192.168.1.1:8080",
		"wss://sub.domain.com:9999",
	}

	for _, urlStr := range validURLs {
		err := ValidateWsConnectionURL(urlStr)
		if err != nil {
			t.Errorf("ValidateWsConnectionURL(%q) returned error: %v", urlStr, err)
		}
	}
}

func TestValidateWsConnectionURL_EmptyURL(t *testing.T) {
	err := ValidateWsConnectionURL("")
	if err == nil {
		t.Error("ValidateWsConnectionURL(\"\") should return error")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("Error should mention empty URL, got: %v", err)
	}
}

func TestValidateWsConnectionURL_InvalidScheme(t *testing.T) {
	invalidURLs := []string{
		"http://localhost:35998",
		"https://localhost:35998",
		"ftp://localhost:35998",
		"tcp://localhost:35998",
		"localhost:35998",
	}

	for _, urlStr := range invalidURLs {
		err := ValidateWsConnectionURL(urlStr)
		if err == nil {
			t.Errorf("ValidateWsConnectionURL(%q) should return error for invalid scheme", urlStr)
		}
		if !strings.Contains(err.Error(), "scheme") {
			t.Errorf("Error should mention scheme, got: %v", err)
		}
	}
}

func TestValidateWsConnectionURL_MissingHost(t *testing.T) {
	invalidURLs := []string{
		"ws://",
		"wss://",
		"ws://:35998",
	}

	for _, urlStr := range invalidURLs {
		err := ValidateWsConnectionURL(urlStr)
		if err == nil {
			t.Errorf("ValidateWsConnectionURL(%q) should return error for missing host", urlStr)
		}
	}
}

func TestValidateWsConnectionURL_InvalidPort(t *testing.T) {
	invalidURLs := []string{
		"ws://localhost:0",
		"ws://localhost:65536",
		"ws://localhost:99999",
		"ws://localhost:-1",
		"ws://localhost:abc",
	}

	for _, urlStr := range invalidURLs {
		err := ValidateWsConnectionURL(urlStr)
		if err == nil {
			t.Errorf("ValidateWsConnectionURL(%q) should return error for invalid port", urlStr)
		}
		if !strings.Contains(err.Error(), "port") {
			t.Errorf("Error should mention port, got: %v", err)
		}
	}
}

func TestValidateWsConnectionURL_MalformedURL(t *testing.T) {
	invalidURLs := []string{
		"://localhost",
		"ws//localhost",
		"ws:localhost",
		"not a url at all",
	}

	for _, urlStr := range invalidURLs {
		err := ValidateWsConnectionURL(urlStr)
		if err == nil {
			t.Errorf("ValidateWsConnectionURL(%q) should return error for malformed URL", urlStr)
		}
	}
}

func TestValidateWsConnectionURL_CaseInsensitive(t *testing.T) {
	validURLs := []string{
		"WS://localhost:35998",
		"Ws://localhost:35998",
		"WSS://localhost:35998",
		"Wss://localhost:35998",
	}

	for _, urlStr := range validURLs {
		err := ValidateWsConnectionURL(urlStr)
		if err != nil {
			t.Errorf("ValidateWsConnectionURL(%q) should accept case-insensitive scheme: %v", urlStr, err)
		}
	}
}

func TestValidateWsConnectionURL_WithPath(t *testing.T) {
	validURLs := []string{
		"ws://localhost:35998/path",
		"ws://localhost:35998/path/to/endpoint",
		"wss://example.com/websocket",
	}

	for _, urlStr := range validURLs {
		err := ValidateWsConnectionURL(urlStr)
		if err != nil {
			t.Errorf("ValidateWsConnectionURL(%q) should accept URL with path: %v", urlStr, err)
		}
	}
}

// =============================================================================
// NormalizeWsURL Tests
// =============================================================================

func TestNormalizeWsURL_WithPort(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ws://localhost:35998", "ws://localhost:35998"},
		{"wss://example.com:443", "wss://example.com:443"},
		{"ws://127.0.0.1:8080", "ws://127.0.0.1:8080"},
	}

	for _, tt := range tests {
		result, err := NormalizeWsURL(tt.input)
		if err != nil {
			t.Errorf("NormalizeWsURL(%q) returned error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("NormalizeWsURL(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeWsURL_WithoutPort(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ws://localhost", "ws://localhost:80"},
		{"wss://localhost", "wss://localhost:443"},
		{"ws://example.com", "ws://example.com:80"},
		{"wss://example.com", "wss://example.com:443"},
	}

	for _, tt := range tests {
		result, err := NormalizeWsURL(tt.input)
		if err != nil {
			t.Errorf("NormalizeWsURL(%q) returned error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("NormalizeWsURL(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeWsURL_InvalidURL(t *testing.T) {
	invalidURLs := []string{
		"",
		"http://localhost",
		"not a url",
		"ws://localhost:99999",
	}

	for _, urlStr := range invalidURLs {
		_, err := NormalizeWsURL(urlStr)
		if err == nil {
			t.Errorf("NormalizeWsURL(%q) should return error", urlStr)
		}
	}
}

func TestNormalizeWsURL_WithPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ws://localhost/path", "ws://localhost:80/path"},
		{"wss://example.com/websocket", "wss://example.com:443/websocket"},
		{"ws://localhost:35998/path", "ws://localhost:35998/path"},
	}

	for _, tt := range tests {
		result, err := NormalizeWsURL(tt.input)
		if err != nil {
			t.Errorf("NormalizeWsURL(%q) returned error: %v", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("NormalizeWsURL(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
