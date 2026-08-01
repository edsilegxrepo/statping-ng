package handlers

import (
	"net"
	"net/http"
	"strings"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/types/users"
	"github.com/statping-ng/statping-ng/utils"
)

// ForwardAuthInfo contains user information extracted from forward auth headers
type ForwardAuthInfo struct {
	Username string
	Email    string
	Name     string
	Groups   []string
	IsAdmin  bool
}

// forwardAuthExtract extracts user info from forward auth headers (Authelia, Authentik, etc.)
// Returns nil if forward auth is disabled, headers not present, or request not from trusted proxy.
func forwardAuthExtract(r *http.Request) *ForwardAuthInfo {
	if !core.App.ForwardAuthEnabled.Bool {
		return nil
	}

	// Verify request comes from trusted proxy (CRITICAL for security)
	if !isFromForwardAuthTrustedProxy(r) {
		return nil
	}

	// Get header names (with defaults)
	headerUser := core.App.ForwardAuthHeaderUser
	if headerUser == "" {
		headerUser = "Remote-User"
	}
	headerEmail := core.App.ForwardAuthHeaderEmail
	if headerEmail == "" {
		headerEmail = "Remote-Email"
	}
	headerGroups := core.App.ForwardAuthHeaderGroups
	if headerGroups == "" {
		headerGroups = "Remote-Groups"
	}
	headerName := core.App.ForwardAuthHeaderName
	if headerName == "" {
		headerName = "Remote-Name"
	}

	// Extract username (required)
	username := r.Header.Get(headerUser)
	if username == "" {
		return nil
	}

	// Extract optional fields
	email := r.Header.Get(headerEmail)
	name := r.Header.Get(headerName)
	groupsStr := r.Header.Get(headerGroups)

	// Parse groups
	var groups []string
	if groupsStr != "" {
		for _, g := range strings.Split(groupsStr, ",") {
			g = strings.TrimSpace(g)
			if g != "" {
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

// forwardAuthUser finds or creates a user from forward auth headers
func forwardAuthUser(r *http.Request) *users.User {
	info := forwardAuthExtract(r)
	if info == nil {
		return nil
	}

	// Find existing user
	user, err := users.FindByUsername(info.Username)
	if err != nil {
		// Auto-provision new user from forward auth
		// Generate a random password - user authenticates via proxy, not locally
		email := info.Email
		if email == "" {
			email = info.Username + "@forward-auth.local"
		}
		user = &users.User{
			Username: info.Username,
			Email:    email,
			Password: utils.HashPassword(utils.RandomString(32)),
			Admin:    null.NewNullBool(info.IsAdmin),
		}
		if err := user.Create(); err != nil {
			log.Errorln("Failed to create forward auth user:", err)
			return nil
		}
		log.Infof("Created forward auth user: %s (admin=%v)", info.Username, info.IsAdmin)
	} else {
		// Update existing user's admin status based on groups
		if user.Admin.Bool != info.IsAdmin {
			user.Admin = null.NewNullBool(info.IsAdmin)
			if err := user.Update(); err != nil {
				log.Warnf("Failed to update forward auth user admin status: %v", err)
			}
		}
		// Update email if changed
		if info.Email != "" && user.Email != info.Email {
			user.Email = info.Email
			if err := user.Update(); err != nil {
				log.Warnf("Failed to update forward auth user email: %v", err)
			}
		}
	}

	return user
}

// isForwardAuthAdmin checks if any group matches configured admin groups
func isForwardAuthAdmin(userGroups []string) bool {
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

// isFromForwardAuthTrustedProxy validates the request originates from a trusted proxy
func isFromForwardAuthTrustedProxy(r *http.Request) bool {
	trustedCIDRs := core.App.ForwardAuthTrustedProxies
	if trustedCIDRs == "" {
		// If no trusted proxies configured, reject all header auth
		log.Warnln("Forward auth headers present but no trusted proxies configured")
		return false
	}

	// Get the direct connection IP (not X-Forwarded-For)
	clientIP := getDirectClientIP(r)
	if clientIP == nil {
		log.Warnln("Forward auth: could not determine client IP")
		return false
	}

	// Parse and check each CIDR
	for _, cidr := range strings.Split(trustedCIDRs, ";") {
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
		if network.Contains(clientIP) {
			return true
		}
	}

	log.Warnf("Forward auth: request from untrusted IP %s", clientIP)
	return false
}

// getDirectClientIP returns the IP of the direct connection (ignoring X-Forwarded-For)
func getDirectClientIP(r *http.Request) net.IP {
	// RemoteAddr is "IP:port" or "[IPv6]:port"
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// Try parsing as IP directly
		return net.ParseIP(r.RemoteAddr)
	}
	return net.ParseIP(host)
}

// ForwardAuthSettingsResponse is the API response for forward auth settings
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

// ForwardAuthSaveRequest is the request body for saving forward auth settings
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

	// Validate trusted proxies CIDRs
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

	// Update core settings
	core.App.ForwardAuthEnabled = null.NewNullBool(req.Enabled)
	core.App.ForwardAuthHeaderUser = req.HeaderUser
	core.App.ForwardAuthHeaderEmail = req.HeaderEmail
	core.App.ForwardAuthHeaderGroups = req.HeaderGroups
	core.App.ForwardAuthHeaderName = req.HeaderName
	core.App.ForwardAuthAdminGroups = req.AdminGroups
	core.App.ForwardAuthTrustedProxies = req.TrustedProxies
	core.App.ForwardAuthLogoutURL = req.LogoutURL

	if err := core.App.Update(); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	log.Infof("Forward auth settings updated (enabled=%v)", req.Enabled)
	returnJson(ForwardAuthSettingsResponse{
		Enabled:        req.Enabled,
		HeaderUser:     req.HeaderUser,
		HeaderEmail:    req.HeaderEmail,
		HeaderGroups:   req.HeaderGroups,
		HeaderName:     req.HeaderName,
		AdminGroups:    req.AdminGroups,
		TrustedProxies: req.TrustedProxies,
		LogoutURL:      req.LogoutURL,
	}, w, r)
}

var (
	errInvalidForwardAuthConfig = errors.New("Trusted proxies required when forward auth is enabled")
	errInvalidCIDR              = errors.New("Invalid CIDR format in trusted proxies")
)
