package utils

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateServerURL validates that the provided string is a valid HTTP/HTTPS URL
func ValidateServerURL(serverURL string) error {
	if serverURL == "" {
		return fmt.Errorf("server URL cannot be empty")
	}

	// Parse the URL
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme, got: %s", parsedURL.Scheme)
	}

	// Check host
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	return nil
}

// NormalizeServerURL ensures the server URL has a proper format
func NormalizeServerURL(serverURL string) string {
	// Remove trailing slash
	return strings.TrimSuffix(serverURL, "/")
}
