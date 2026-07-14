package rpc_client

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ValidateConnectionURL validates a Zenon JSON-RPC connection URL.
//
// Parameters:
//   - urlStr: Absolute HTTP, HTTPS, WebSocket, or secure WebSocket URL.
//
// ValidateConnectionURL returns an error for an empty URL, unsupported scheme,
// missing hostname, or port outside 1 through 65535. Validation occurs before
// any connection is attempted.
//
// Example:
//
//	err := rpc_client.ValidateConnectionURL("https://node.example.com")
//
// Supported schemes are http, https, ws, and wss. Subscriptions require ws or
// wss; HTTP transports support ordinary JSON-RPC reads and writes.
func ValidateConnectionURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" && scheme != "ws" && scheme != "wss" {
		return fmt.Errorf("invalid scheme '%s': must be http, https, ws, or wss", parsedURL.Scheme)
	}

	// Check host is not empty
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must contain a host")
	}

	// Check hostname is not empty (host could be just ":port")
	if parsedURL.Hostname() == "" {
		return fmt.Errorf("URL must contain a hostname")
	}

	// Validate port if specified
	if parsedURL.Port() != "" {
		port, err := strconv.Atoi(parsedURL.Port())
		if err != nil {
			return fmt.Errorf("invalid port '%s': %w", parsedURL.Port(), err)
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("invalid port %d: must be between 1 and 65535", port)
		}
	}

	return nil
}

// NormalizeConnectionURL validates a transport URL and adds its default port.
//
// HTTP and WebSocket URLs default to port 80; HTTPS and secure WebSocket URLs
// default to port 443. Existing paths, queries, and explicit ports are
// preserved.
func NormalizeConnectionURL(urlStr string) (string, error) {
	if err := ValidateConnectionURL(urlStr); err != nil {
		return "", err
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		// This should not happen since ValidateConnectionURL already validated the URL.
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add default port if missing
	if parsedURL.Port() == "" {
		switch parsedURL.Scheme {
		case "http", "ws":
			parsedURL.Host = parsedURL.Host + ":80"
		case "https", "wss":
			parsedURL.Host = parsedURL.Host + ":443"
		}
	}

	return parsedURL.String(), nil
}

// ValidateWsConnectionURL validates a Zenon RPC lifecycle URL.
//
// Deprecated: Use [ValidateConnectionURL]. The legacy name now accepts HTTP,
// HTTPS, WS, and WSS so existing callers gain the stable transport lifecycle.
func ValidateWsConnectionURL(urlStr string) error {
	return ValidateConnectionURL(urlStr)
}

// NormalizeWsURL normalizes a Zenon RPC lifecycle URL.
//
// Deprecated: Use [NormalizeConnectionURL].
func NormalizeWsURL(urlStr string) (string, error) {
	return NormalizeConnectionURL(urlStr)
}
