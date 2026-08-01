package handlers

import (
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/types/users"
	"github.com/statping-ng/statping-ng/utils"
)

// Default header names for forward auth (used when config is empty)
const (
	DefaultForwardAuthHeaderUser   = "Remote-User"
	DefaultForwardAuthHeaderEmail  = "Remote-Email"
	DefaultForwardAuthHeaderGroups = "Remote-Groups"
	DefaultForwardAuthHeaderName   = "Remote-Name"

	// Limits
	MaxUsernameLength    = 255
	MaxEmailLength       = 320
	MaxTrustedProxiesLen = 4096
	MaxAdminGroupsLen    = 1024
	MaxLogoutURLLen      = 2048
	MaxHeaderNameLen     = 128

	// Rate limiting for user creation
	userCreationWindow    = time.Minute
	userCreationMaxPerMin = 10
)

var (
	// Mutex for user provisioning to prevent race conditions
	userProvisionMu sync.Mutex

	// Cached parsed CIDR networks (updated when config changes)
	trustedNetworksMu     sync.RWMutex
	trustedNetworks       []*net.IPNet
	trustedNetworksConfig string // The config string that was parsed

	// Rate limiter for user creation per IP
	userCreationLimiter   = make(map[string]*rateLimitEntry)
	userCreationLimiterMu sync.Mutex

	// Username validation: alphanumeric, underscore, hyphen, dot
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	// Email validation: basic format check
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

type rateLimitEntry struct {
	count       int
	windowStart time.Time
}

// ForwardAuthInfo contains user information extracted from forward auth headers.
// This struct holds the identity claims passed by the authenticating proxy.
type ForwardAuthInfo struct {
	Username string   // The authenticated username
	Email    string   // User's email address (optional)
	Name     string   // User's display name (optional)
	Groups   []string // Group memberships from the proxy
	IsAdmin  bool     // Whether user is in an admin group
}

// forwardAuthExtract extracts user info from forward auth headers (Authelia, Authentik, etc.)
// Returns nil if forward auth is disabled, headers not present, or request not from trusted proxy.
func forwardAuthExtract(r *http.Request) *ForwardAuthInfo {
	if core.App == nil || !core.App.ForwardAuthEnabled.Bool {
		return nil
	}

	// Verify request comes from trusted proxy (CRITICAL for security)
	if !isFromForwardAuthTrustedProxy(r) {
		return nil
	}

	// Get header names (with defaults)
	headerUser := core.App.ForwardAuthHeaderUser
	if headerUser == "" {
		headerUser = DefaultForwardAuthHeaderUser
	}
	headerEmail := core.App.ForwardAuthHeaderEmail
	if headerEmail == "" {
		headerEmail = DefaultForwardAuthHeaderEmail
	}
	headerGroups := core.App.ForwardAuthHeaderGroups
	if headerGroups == "" {
		headerGroups = DefaultForwardAuthHeaderGroups
	}
	headerName := core.App.ForwardAuthHeaderName
	if headerName == "" {
		headerName = DefaultForwardAuthHeaderName
	}

	// Extract username (required)
	username := r.Header.Get(headerUser)
	if username == "" {
		return nil
	}

	// Validate username format
	if !isValidUsername(username) {
		log.Warnf("Forward auth: invalid username format")
		return nil
	}

	// Extract and validate optional fields
	email := r.Header.Get(headerEmail)
	if email != "" && !isValidEmail(email) {
		log.Warnf("Forward auth: invalid email format, ignoring")
		email = ""
	}

	name := sanitizeDisplayName(r.Header.Get(headerName))
	groupsStr := r.Header.Get(headerGroups)

	// Parse groups
	var groups []string
	if groupsStr != "" {
		for _, g := range strings.Split(groupsStr, ",") {
			g = strings.TrimSpace(g)
			if g != "" && len(g) <= 256 {
				groups = append(groups, g)
			}
		}
	}

	// Check if user is admin based on groups
	isAdmin := isForwardAuthAdmin(groups)

	return &ForwardAuthInfo{
		Username: username,
		Email:    email,
		Name:     name,
		Groups:   groups,
		IsAdmin:  isAdmin,
	}
}

// isValidUsername checks if the username contains only allowed characters.
// Allows alphanumeric, underscore, hyphen, and dot. Max 255 chars.
func isValidUsername(s string) bool {
	if len(s) == 0 || len(s) > MaxUsernameLength {
		return false
	}
	return usernameRegex.MatchString(s)
}

// isValidEmail performs basic email format validation
func isValidEmail(s string) bool {
	if len(s) == 0 || len(s) > MaxEmailLength {
		return false
	}
	return emailRegex.MatchString(s)
}

// sanitizeDisplayName removes control characters from display name
func sanitizeDisplayName(s string) string {
	if len(s) > MaxUsernameLength {
		s = s[:MaxUsernameLength]
	}
	// Remove control characters
	var result strings.Builder
	for _, r := range s {
		if r >= 32 && r != 127 {
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

// forwardAuthUser finds or creates a user from forward auth headers.
// Uses mutex to prevent race conditions during user provisioning.
func forwardAuthUser(r *http.Request) *users.User {
	info := forwardAuthExtract(r)
	if info == nil {
		return nil
	}

	// Use mutex to prevent race condition when creating users
	userProvisionMu.Lock()
	defer userProvisionMu.Unlock()

	// Find existing user
	user, err := users.FindByUsername(info.Username)
	if err != nil {
		// Check rate limit before creating new user
		clientIP := getDirectClientIP(r)
		if clientIP != nil && !checkUserCreationRateLimit(clientIP.String()) {
			log.Warnf("Forward auth: user creation rate limited")
			return nil
		}

		// Auto-provision new user from forward auth
		// Generate a random password - user authenticates via proxy, not locally
		email := info.Email
		if email == "" {
			email = info.Username + "@forward-auth.local"
		}
		user = &users.User{
			Username:           info.Username,
			Email:              email,
			Password:           utils.HashPassword(utils.RandomString(32)),
			Admin:              null.NewNullBool(info.IsAdmin),
			ForwardAuthManaged: null.NewNullBool(true), // Mark as forward auth managed
		}
		if err := user.Create(); err != nil {
			log.Errorln("Failed to create forward auth user:", err)
			return nil
		}
		log.Infof("Created forward auth user: %s (admin=%v)", info.Username, info.IsAdmin)
	} else {
		// Only update admin status for forward-auth-managed users
		// This prevents demoting manually-created admin accounts
		if user.ForwardAuthManaged.Bool {
			needsUpdate := false

			if user.Admin.Bool != info.IsAdmin {
				user.Admin = null.NewNullBool(info.IsAdmin)
				needsUpdate = true
			}
			// Update email if changed and provided
			if info.Email != "" && user.Email != info.Email {
				user.Email = info.Email
				needsUpdate = true
			}

			if needsUpdate {
				if err := user.Update(); err != nil {
					log.Warnf("Failed to update forward auth user: %v", err)
				}
			}
		}
	}

	return user
}

// checkUserCreationRateLimit returns true if the IP is allowed to create a user
func checkUserCreationRateLimit(ip string) bool {
	userCreationLimiterMu.Lock()
	defer userCreationLimiterMu.Unlock()

	now := time.Now()
	entry, exists := userCreationLimiter[ip]

	if !exists || now.Sub(entry.windowStart) > userCreationWindow {
		userCreationLimiter[ip] = &rateLimitEntry{
			count:       1,
			windowStart: now,
		}
		return true
	}

	if entry.count >= userCreationMaxPerMin {
		return false
	}

	entry.count++
	return true
}

// isForwardAuthAdmin checks if any group matches configured admin groups
func isForwardAuthAdmin(userGroups []string) bool {
	if core.App == nil {
		return false
	}
	adminGroupsStr := core.App.ForwardAuthAdminGroups
	if adminGroupsStr == "" {
		return false
	}

	// Parse admin groups (semicolon-separated)
	var adminGroups []string
	for _, g := range strings.Split(adminGroupsStr, ";") {
		g = strings.TrimSpace(g)
		if g != "" {
			adminGroups = append(adminGroups, g)
		}
	}

	// Check for match
	for _, ag := range adminGroups {
		for _, ug := range userGroups {
			if ug == ag {
				return true
			}
		}
	}
	return false
}

// updateTrustedNetworksCache parses and caches the trusted proxy CIDRs.
// Call this when the config changes.
func updateTrustedNetworksCache(cidrs string) {
	trustedNetworksMu.Lock()
	defer trustedNetworksMu.Unlock()

	// Skip if unchanged
	if cidrs == trustedNetworksConfig {
		return
	}

	trustedNetworksConfig = cidrs
	trustedNetworks = nil

	if cidrs == "" {
		return
	}

	for _, cidr := range strings.Split(cidrs, ";") {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}

		// Handle single IP (add /32 or /128)
		if !strings.Contains(cidr, "/") {
			if strings.Contains(cidr, ":") {
				cidr += "/128"
			} else {
				cidr += "/32"
			}
		}

		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Warnf("Forward auth: invalid CIDR %q: %v", cidr, err)
			continue
		}
		trustedNetworks = append(trustedNetworks, network)
	}
}

// isFromForwardAuthTrustedProxy validates the request originates from a trusted proxy.
// Uses cached CIDR networks for performance.
func isFromForwardAuthTrustedProxy(r *http.Request) bool {
	if core.App == nil {
		return false
	}

	// Update cache if config changed
	currentConfig := core.App.ForwardAuthTrustedProxies
	trustedNetworksMu.RLock()
	needsUpdate := currentConfig != trustedNetworksConfig
	trustedNetworksMu.RUnlock()

	if needsUpdate {
		updateTrustedNetworksCache(currentConfig)
	}

	trustedNetworksMu.RLock()
	networks := trustedNetworks
	trustedNetworksMu.RUnlock()

	if len(networks) == 0 {
		log.Warnln("Forward auth: no trusted proxies configured")
		return false
	}

	// Get the direct connection IP (not X-Forwarded-For)
	clientIP := getDirectClientIP(r)
	if clientIP == nil {
		log.Warnln("Forward auth: could not determine client IP")
		return false
	}

	// Normalize IPv4-mapped IPv6 addresses
	if ipv4 := clientIP.To4(); ipv4 != nil {
		clientIP = ipv4
	}

	// Check against cached networks
	for _, network := range networks {
		if network.Contains(clientIP) {
			return true
		}
	}

	log.Warnf("Forward auth: request from untrusted IP %s", clientIP)
	return false
}

// getDirectClientIP returns the IP of the direct connection (ignoring X-Forwarded-For).
// This ensures we validate the immediate upstream proxy, not a spoofed header.
func getDirectClientIP(r *http.Request) net.IP {
	// RemoteAddr is "IP:port" or "[IPv6]:port"
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// Try parsing as IP directly
		return net.ParseIP(r.RemoteAddr)
	}
	return net.ParseIP(host)
}

// ForwardAuthSettingsResponse is the API response for forward auth settings.
// Returned by GET /api/forwardauth endpoint.
type ForwardAuthSettingsResponse struct {
	Enabled        bool   `json:"forward_auth_enabled"`
	HeaderUser     string `json:"forward_auth_header_user"`
	HeaderEmail    string `json:"forward_auth_header_email"`
	HeaderGroups   string `json:"forward_auth_header_groups"`
	HeaderName     string `json:"forward_auth_header_name"`
	AdminGroups    string `json:"forward_auth_admin_groups"`
	TrustedProxies string `json:"forward_auth_trusted_proxies"`
	LogoutURL      string `json:"forward_auth_logout_url"`
}

// forwardAuthSettingsHandler returns forward auth settings (GET /api/forwardauth)
func forwardAuthSettingsHandler(w http.ResponseWriter, r *http.Request) {
	resp := ForwardAuthSettingsResponse{
		Enabled:        core.App.ForwardAuthEnabled.Bool,
		HeaderUser:     core.App.ForwardAuthHeaderUser,
		HeaderEmail:    core.App.ForwardAuthHeaderEmail,
		HeaderGroups:   core.App.ForwardAuthHeaderGroups,
		HeaderName:     core.App.ForwardAuthHeaderName,
		AdminGroups:    core.App.ForwardAuthAdminGroups,
		TrustedProxies: core.App.ForwardAuthTrustedProxies,
		LogoutURL:      core.App.ForwardAuthLogoutURL,
	}
	returnJson(resp, w, r)
}

// ForwardAuthSaveRequest is the request body for saving forward auth settings.
// Used by POST /api/forwardauth endpoint.
type ForwardAuthSaveRequest struct {
	Enabled        bool   `json:"forward_auth_enabled"`
	HeaderUser     string `json:"forward_auth_header_user"`
	HeaderEmail    string `json:"forward_auth_header_email"`
	HeaderGroups   string `json:"forward_auth_header_groups"`
	HeaderName     string `json:"forward_auth_header_name"`
	AdminGroups    string `json:"forward_auth_admin_groups"`
	TrustedProxies string `json:"forward_auth_trusted_proxies"`
	LogoutURL      string `json:"forward_auth_logout_url"`
}

// forwardAuthSaveHandler saves forward auth settings (POST /api/forwardauth)
func forwardAuthSaveHandler(w http.ResponseWriter, r *http.Request) {
	var req ForwardAuthSaveRequest
	if err := DecodeJSON(r, &req); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	// Validate input lengths
	if len(req.TrustedProxies) > MaxTrustedProxiesLen {
		sendErrorJson(errTrustedProxiesTooLong, w, r)
		return
	}
	if len(req.AdminGroups) > MaxAdminGroupsLen {
		sendErrorJson(errAdminGroupsTooLong, w, r)
		return
	}
	if len(req.LogoutURL) > MaxLogoutURLLen {
		sendErrorJson(errLogoutURLTooLong, w, r)
		return
	}
	if len(req.HeaderUser) > MaxHeaderNameLen || len(req.HeaderEmail) > MaxHeaderNameLen ||
		len(req.HeaderGroups) > MaxHeaderNameLen || len(req.HeaderName) > MaxHeaderNameLen {
		sendErrorJson(errHeaderNameTooLong, w, r)
		return
	}

	// Validate trusted proxies CIDRs (required when enabled)
	if req.Enabled && req.TrustedProxies == "" {
		sendErrorJson(errInvalidForwardAuthConfig, w, r)
		return
	}

	// Validate CIDR syntax
	if req.TrustedProxies != "" {
		for _, cidr := range strings.Split(req.TrustedProxies, ";") {
			cidr = strings.TrimSpace(cidr)
			if cidr == "" {
				continue
			}
			// Add mask if missing
			if !strings.Contains(cidr, "/") {
				if strings.Contains(cidr, ":") {
					cidr += "/128"
				} else {
					cidr += "/32"
				}
			}
			if _, _, err := net.ParseCIDR(cidr); err != nil {
				sendErrorJson(errInvalidCIDR, w, r)
				return
			}
		}
	}

	// Validate logout URL format if provided
	if req.LogoutURL != "" {
		if !strings.HasPrefix(req.LogoutURL, "http://") && !strings.HasPrefix(req.LogoutURL, "https://") {
			sendErrorJson(errInvalidLogoutURL, w, r)
			return
		}
	}

	// Update core settings
	core.App.ForwardAuthEnabled = null.NewNullBool(req.Enabled)
	core.App.ForwardAuthHeaderUser = req.HeaderUser
	core.App.ForwardAuthHeaderEmail = req.HeaderEmail
	core.App.ForwardAuthHeaderGroups = req.HeaderGroups
	core.App.ForwardAuthHeaderName = req.HeaderName
	core.App.ForwardAuthAdminGroups = req.AdminGroups
	core.App.ForwardAuthTrustedProxies = req.TrustedProxies
	core.App.ForwardAuthLogoutURL = req.LogoutURL

	// Update cached networks
	updateTrustedNetworksCache(req.TrustedProxies)

	if err := core.App.Update(); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	log.Infof("Forward auth settings updated (enabled=%v)", req.Enabled)
	returnJson(ForwardAuthSettingsResponse(req), w, r)
}

// ForwardAuthLogoutURL returns the configured logout URL for forward auth.
// Returns empty string if forward auth is disabled or no logout URL is set.
func ForwardAuthLogoutURL() string {
	if core.App == nil || !core.App.ForwardAuthEnabled.Bool {
		return ""
	}
	return core.App.ForwardAuthLogoutURL
}

// Error definitions for forward auth
var (
	errInvalidForwardAuthConfig = errors.New("Trusted proxies required when forward auth is enabled")
	errInvalidCIDR              = errors.New("Invalid CIDR format in trusted proxies")
	errInvalidLogoutURL         = errors.New("Logout URL must start with http:// or https://")
	errTrustedProxiesTooLong    = errors.New("Trusted proxies configuration too long (max 4096 chars)")
	errAdminGroupsTooLong       = errors.New("Admin groups configuration too long (max 1024 chars)")
	errLogoutURLTooLong         = errors.New("Logout URL too long (max 2048 chars)")
	errHeaderNameTooLong        = errors.New("Header name too long (max 128 chars)")
)
