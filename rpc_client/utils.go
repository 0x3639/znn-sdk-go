package rpc_client

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ValidateWsConnectionURL validates a WebSocket connection URL
// Returns an error if the URL is invalid
func ValidateWsConnectionURL(urlStr string) error {
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
	if scheme != "ws" && scheme != "wss" {
		return fmt.Errorf("invalid scheme '%s': must be 'ws' or 'wss'", parsedURL.Scheme)
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

// NormalizeWsURL normalizes a WebSocket URL by adding default port if missing
func NormalizeWsURL(urlStr string) (string, error) {
	if err := ValidateWsConnectionURL(urlStr); err != nil {
		return "", err
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		// This should not happen since ValidateWsConnectionURL already validated the URL
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add default port if missing
	if parsedURL.Port() == "" {
		switch parsedURL.Scheme {
		case "ws":
			parsedURL.Host = parsedURL.Host + ":80"
		case "wss":
			parsedURL.Host = parsedURL.Host + ":443"
		}
	}

	return parsedURL.String(), nil
}
