package utils

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// IsInternalIP checks if an IP address is internal/private/loopback
func IsInternalIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	// Convert IPv4-mapped IPv6 to IPv4 for consistent checking
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4
	}
	// Explicit check for cloud metadata endpoints (169.254.169.254)
	cloudMetadata := net.ParseIP("169.254.169.254")
	if ip.Equal(cloudMetadata) {
		return true
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}

// IsInternalHost checks if a hostname resolves to an internal IP
func IsInternalHost(host string) (bool, error) {
	// Remove port if present
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}

	// Check for obvious internal hostnames
	lowerHost := strings.ToLower(hostname)
	if lowerHost == "localhost" || strings.HasSuffix(lowerHost, ".local") ||
		strings.HasSuffix(lowerHost, ".internal") || strings.HasSuffix(lowerHost, ".localhost") {
		return true, nil
	}

	// Resolve hostname to IPs
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return false, fmt.Errorf("failed to resolve host %q: %w", hostname, err)
	}

	for _, ip := range ips {
		if IsInternalIP(ip) {
			return true, nil
		}
	}

	return false, nil
}

// ValidateExternalURL checks if a URL is safe to request (not internal/SSRF)
// Returns an error if the URL targets internal resources
func ValidateExternalURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("empty URL")
	}

	// Reject URLs with backslashes (parser inconsistency risk)
	if strings.ContainsAny(rawURL, "\\") {
		return fmt.Errorf("URL contains invalid characters")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow http/https
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme %q: only http/https allowed", parsed.Scheme)
	}

	host := parsed.Host
	if host == "" {
		return fmt.Errorf("URL missing host")
	}

	// Check if host is an IP literal
	hostname := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}

	// Try parsing as IP directly
	if ip := net.ParseIP(hostname); ip != nil {
		if IsInternalIP(ip) {
			return fmt.Errorf("URL targets internal IP %s", ip)
		}
		return nil
	}

	// Resolve hostname and check all IPs
	isInternal, err := IsInternalHost(host)
	if err != nil {
		return err
	}
	if isInternal {
		return fmt.Errorf("URL targets internal host %q", hostname)
	}

	return nil
}

// ValidateAndResolveURL validates a URL and returns the resolved IPs to prevent DNS rebinding
// The caller should use these IPs with SafeHTTPClient to make the actual request
// Respects ALLOW_INTERNAL_URLS for testing environments
func ValidateAndResolveURL(rawURL string) ([]net.IP, error) {
	if err := ValidateExternalURLAllowInternal(rawURL); err != nil {
		return nil, err
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	hostname := parsed.Hostname()

	// If it's already an IP, return it
	if ip := net.ParseIP(hostname); ip != nil {
		return []net.IP{ip}, nil
	}

	// Resolve and return IPs
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host %q: %w", hostname, err)
	}

	return ips, nil
}

// SafeHTTPClient returns an HTTP client that pins DNS resolution to prevent rebinding attacks
// The resolvedIPs should come from ValidateAndResolveURL called at validation time
func SafeHTTPClient(resolvedIPs []net.IP, timeout time.Duration) *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// Check if internal URLs are allowed (for testing)
	allowInternal := Params.GetBool("ALLOW_INTERNAL_URLS")

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Extract port from addr
			_, port, err := net.SplitHostPort(addr)
			if err != nil {
				port = "443" // default
			}

			// Use the pre-resolved IP instead of re-resolving
			if len(resolvedIPs) > 0 {
				// Re-validate the IP (defense in depth) - skip if internal URLs allowed
				if !allowInternal {
					for _, ip := range resolvedIPs {
						if IsInternalIP(ip) {
							return nil, fmt.Errorf("resolved IP %s is internal", ip)
						}
					}
				}
				// Use the first valid IP
				pinnedAddr := net.JoinHostPort(resolvedIPs[0].String(), port)
				return dialer.DialContext(ctx, network, pinnedAddr)
			}

			// Fallback to normal dialing (shouldn't happen if used correctly)
			return dialer.DialContext(ctx, network, addr)
		},
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// ValidateExternalURLAllowInternal is like ValidateExternalURL but allows internal URLs
// when ALLOW_INTERNAL_URLS env var is set (for development/testing)
func ValidateExternalURLAllowInternal(rawURL string) error {
	if Params.GetBool("ALLOW_INTERNAL_URLS") {
		// Still validate basic URL structure
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return fmt.Errorf("invalid URL scheme %q: only http/https allowed", parsed.Scheme)
		}
		if parsed.Host == "" {
			return fmt.Errorf("URL missing host")
		}
		return nil
	}
	return ValidateExternalURL(rawURL)
}
