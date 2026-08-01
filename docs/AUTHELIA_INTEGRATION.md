# Authelia Integration Plan for Statping-ng

## Overview

Integrate [Authelia](https://www.authelia.com/) as an authentication provider for Statping-ng. Authelia is a lightweight self-hosted SSO portal that uses forward-auth patterns with reverse proxies (Traefik, NGINX, Caddy, HAProxy).

## Integration Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Browser   │────▶│   Traefik   │────▶│  Authelia   │     │  Statping   │
│             │◀────│   (proxy)   │◀────│   (auth)    │     │    (app)    │
└─────────────┘     └──────┬──────┘     └─────────────┘     └──────▲──────┘
                           │                                       │
                           │  Remote-User: john                    │
                           │  Remote-Groups: admins,users          │
                           │  Remote-Email: john@example.com       │
                           └───────────────────────────────────────┘
```

**Flow:**
1. User requests Statping page
2. Traefik forwards auth request to Authelia (`/api/authz/forward-auth`)
3. Authelia validates session cookie or redirects to login
4. On success (HTTP 200), Authelia returns `Remote-*` headers
5. Traefik forwards request to Statping with `Remote-*` headers
6. Statping trusts headers and creates/finds user session

## Implementation Phases

### Phase 1: Remote User Header Authentication

**Files to create/modify:**
- `handlers/authelia.go` - Header extraction and validation
- `handlers/authentication.go` - Add header auth check
- `handlers/routes.go` - Register middleware
- `types/core/core.go` - Add Authelia config fields

**New Configuration Fields:**
```go
// In types/core/core.go
type Core struct {
    // ... existing fields ...
    
    // Authelia/Forward-Auth Settings
    AutheliaEnabled       null.NullBool   `json:"authelia_enabled"`
    AutheliaHeaderUser    null.NullString `json:"authelia_header_user"`    // Default: "Remote-User"
    AutheliaHeaderEmail   null.NullString `json:"authelia_header_email"`   // Default: "Remote-Email"
    AutheliaHeaderGroups  null.NullString `json:"authelia_header_groups"`  // Default: "Remote-Groups"
    AutheliaHeaderName    null.NullString `json:"authelia_header_name"`    // Default: "Remote-Name"
    AutheliaAdminGroups   null.NullString `json:"authelia_admin_groups"`   // Semicolon-separated, e.g., "admins;statping-admins"
    AutheliaTrustedProxies null.NullString `json:"authelia_trusted_proxies"` // CIDR ranges, e.g., "10.0.0.0/8;172.16.0.0/12"
}
```

**Environment Variables:**
```bash
AUTHELIA_ENABLED=true
AUTHELIA_HEADER_USER=Remote-User
AUTHELIA_HEADER_EMAIL=Remote-Email
AUTHELIA_HEADER_GROUPS=Remote-Groups
AUTHELIA_HEADER_NAME=Remote-Name
AUTHELIA_ADMIN_GROUPS=admins;statping-admins
AUTHELIA_TRUSTED_PROXIES=10.0.0.0/8;172.16.0.0/12
```

**New Handler: `handlers/authelia.go`**
```go
package handlers

import (
    "net"
    "net/http"
    "strings"
    
    "github.com/statping-ng/statping-ng/types/core"
    "github.com/statping-ng/statping-ng/types/users"
    "github.com/statping-ng/statping-ng/utils"
)

// autheliaAuth extracts user info from Authelia's Remote-* headers
// Returns nil if headers not present or request not from trusted proxy
func autheliaAuth(r *http.Request) *users.User {
    if !core.App.AutheliaEnabled.Bool {
        return nil
    }
    
    // Verify request comes from trusted proxy
    if !isFromTrustedProxy(r) {
        log.Warnln("Authelia headers present but request not from trusted proxy")
        return nil
    }
    
    // Extract headers (configurable names)
    headerUser := core.App.AutheliaHeaderUser.String
    if headerUser == "" {
        headerUser = "Remote-User"
    }
    
    username := r.Header.Get(headerUser)
    if username == "" {
        return nil // No Authelia auth
    }
    
    email := r.Header.Get(getHeaderName("email", "Remote-Email"))
    groups := r.Header.Get(getHeaderName("groups", "Remote-Groups"))
    name := r.Header.Get(getHeaderName("name", "Remote-Name"))
    
    // Find or create user
    user, err := users.FindByUsername(username)
    if err != nil {
        // Auto-provision new user from Authelia
        user = &users.User{
            Username: username,
            Email:    email,
            Admin:    null.NewNullBool(isAdminGroup(groups)),
        }
        if name != "" {
            user.Username = username // Keep username, could add display name field
        }
        if err := user.Create(); err != nil {
            log.Errorln("Failed to create Authelia user:", err)
            return nil
        }
        log.Infof("Created Authelia user: %s (admin=%v)", username, user.Admin.Bool)
    } else {
        // Update existing user's admin status based on groups
        newAdmin := isAdminGroup(groups)
        if user.Admin.Bool != newAdmin {
            user.Admin = null.NewNullBool(newAdmin)
            user.Update()
        }
    }
    
    return user
}

// isAdminGroup checks if any group matches configured admin groups
func isAdminGroup(groups string) bool {
    adminGroups := core.App.AutheliaAdminGroups.String
    if adminGroups == "" {
        return false
    }
    
    adminList := strings.Split(adminGroups, ";")
    userGroups := strings.Split(groups, ",")
    
    for _, ag := range adminList {
        ag = strings.TrimSpace(ag)
        for _, ug := range userGroups {
            if strings.TrimSpace(ug) == ag {
                return true
            }
        }
    }
    return false
}

// isFromTrustedProxy validates the request originates from a trusted proxy
func isFromTrustedProxy(r *http.Request) bool {
    trustedCIDRs := core.App.AutheliaTrustedProxies.String
    if trustedCIDRs == "" {
        // If no trusted proxies configured, reject all header auth
        return false
    }
    
    clientIP := getClientIP(r)
    if clientIP == nil {
        return false
    }
    
    for _, cidr := range strings.Split(trustedCIDRs, ";") {
        cidr = strings.TrimSpace(cidr)
        _, network, err := net.ParseCIDR(cidr)
        if err != nil {
            continue
        }
        if network.Contains(clientIP) {
            return true
        }
    }
    return false
}
```

**Modify `handlers/authentication.go`:**
```go
// Add to IsAuthenticated or create new check
func hasAutheliaAuth(r *http.Request) (*users.User, bool) {
    user := autheliaAuth(r)
    return user, user != nil
}
```

**Security Considerations:**
1. **Trusted Proxy Validation**: CRITICAL - must verify `X-Forwarded-For` comes from trusted proxy IP
2. **Header Spoofing Prevention**: Only accept `Remote-*` headers from configured proxy CIDRs
3. **No Local Bypass**: Authelia auth should not bypass local password requirements when disabled

### Phase 2: UI Configuration

**Files to modify:**
- `frontend/src/pages/Settings.vue` - Add Authelia configuration section
- `handlers/dashboard.go` - API endpoints for settings

**Settings UI Section:**
```
┌─────────────────────────────────────────────────────────┐
│ Forward Auth (Authelia/Authentik)                       │
├─────────────────────────────────────────────────────────┤
│ ☑ Enable Forward Auth                                   │
│                                                         │
│ Header Configuration:                                   │
│ ┌─────────────────┐  ┌─────────────────┐               │
│ │ User Header     │  │ Remote-User     │               │
│ └─────────────────┘  └─────────────────┘               │
│ ┌─────────────────┐  ┌─────────────────┐               │
│ │ Email Header    │  │ Remote-Email    │               │
│ └─────────────────┘  └─────────────────┘               │
│ ┌─────────────────┐  ┌─────────────────┐               │
│ │ Groups Header   │  │ Remote-Groups   │               │
│ └─────────────────┘  └─────────────────┘               │
│                                                         │
│ Authorization:                                          │
│ ┌─────────────────┐  ┌─────────────────────────────┐   │
│ │ Admin Groups    │  │ admins;statping-admins      │   │
│ └─────────────────┘  └─────────────────────────────┘   │
│ (semicolon-separated group names)                       │
│                                                         │
│ Security:                                               │
│ ┌─────────────────┐  ┌─────────────────────────────┐   │
│ │ Trusted Proxies │  │ 10.0.0.0/8;172.16.0.0/12    │   │
│ └─────────────────┘  └─────────────────────────────┘   │
│ (semicolon-separated CIDR ranges)                       │
│                                                         │
│                              [Test Connection] [Save]   │
└─────────────────────────────────────────────────────────┘
```

### Phase 3: Session Integration

**Behavior when Authelia is enabled:**

1. **Login Page**: Show "Authenticated via Authelia" message instead of login form
2. **Session Creation**: Create Statping session from Authelia headers on first request
3. **Session Refresh**: Re-validate headers on each request (headers always present from proxy)
4. **Logout**: Redirect to Authelia logout URL (configurable)

**New Config Field:**
```go
AutheliaLogoutURL null.NullString `json:"authelia_logout_url"` // e.g., "https://auth.example.com/logout"
```

### Phase 4: API Authentication Support

Support Authelia's OAuth 2.0 Bearer Token flow for API access:

**Headers to support:**
- `Authorization: Bearer <authelia_token>` - OAuth 2.0 access token from Authelia

**Implementation:**
```go
// In handlers/authentication.go
func hasAutheliaBearerToken(r *http.Request) bool {
    if !core.App.AutheliaEnabled.Bool {
        return false
    }
    
    // Authelia validates the token via forward-auth
    // If we receive the request with Remote-User header, token is valid
    return autheliaAuth(r) != nil
}
```

Note: The actual token validation happens at Authelia's `/api/authz/forward-auth` endpoint. Statping only needs to trust the resulting `Remote-*` headers.

## Database Migration

```go
// migrations/authelia.go
func MigrateAuthelia(db database.Database) error {
    return db.Table("core").AutoMigrate(
        &struct {
            AutheliaEnabled        bool   `gorm:"column:authelia_enabled;default:false"`
            AutheliaHeaderUser     string `gorm:"column:authelia_header_user;default:'Remote-User'"`
            AutheliaHeaderEmail    string `gorm:"column:authelia_header_email;default:'Remote-Email'"`
            AutheliaHeaderGroups   string `gorm:"column:authelia_header_groups;default:'Remote-Groups'"`
            AutheliaHeaderName     string `gorm:"column:authelia_header_name;default:'Remote-Name'"`
            AutheliaAdminGroups    string `gorm:"column:authelia_admin_groups"`
            AutheliaTrustedProxies string `gorm:"column:authelia_trusted_proxies"`
            AutheliaLogoutURL      string `gorm:"column:authelia_logout_url"`
        }{},
    ).Error()
}
```

## Proxy Configuration Examples

### Traefik (docker-compose)

```yaml
labels:
  - "traefik.http.routers.statping.middlewares=authelia@docker"
  - "traefik.http.middlewares.authelia.forwardauth.address=http://authelia:9091/api/authz/forward-auth"
  - "traefik.http.middlewares.authelia.forwardauth.trustForwardHeader=true"
  - "traefik.http.middlewares.authelia.forwardauth.authResponseHeaders=Remote-User,Remote-Groups,Remote-Email,Remote-Name"
```

### NGINX

```nginx
location / {
    auth_request /authelia;
    auth_request_set $user $upstream_http_remote_user;
    auth_request_set $groups $upstream_http_remote_groups;
    auth_request_set $email $upstream_http_remote_email;
    auth_request_set $name $upstream_http_remote_name;
    
    proxy_set_header Remote-User $user;
    proxy_set_header Remote-Groups $groups;
    proxy_set_header Remote-Email $email;
    proxy_set_header Remote-Name $name;
    
    proxy_pass http://statping:8080;
}

location /authelia {
    internal;
    proxy_pass http://authelia:9091/api/authz/auth-request;
    proxy_set_header X-Original-URL $scheme://$http_host$request_uri;
    proxy_set_header X-Original-Method $request_method;
}
```

### Caddy

```caddyfile
statping.example.com {
    forward_auth authelia:9091 {
        uri /api/authz/forward-auth
        copy_headers Remote-User Remote-Groups Remote-Email Remote-Name
    }
    reverse_proxy statping:8080
}
```

## Testing Plan

### Unit Tests (`handlers/authelia_test.go`)

```go
func TestAutheliaAuth(t *testing.T) {
    tests := []struct {
        name           string
        headers        map[string]string
        trustedProxies string
        clientIP       string
        wantUser       bool
        wantAdmin      bool
    }{
        {
            name: "valid headers from trusted proxy",
            headers: map[string]string{
                "Remote-User":   "john",
                "Remote-Email":  "john@example.com",
                "Remote-Groups": "users,admins",
            },
            trustedProxies: "10.0.0.0/8",
            clientIP:       "10.0.0.5",
            wantUser:       true,
            wantAdmin:      true,
        },
        {
            name: "headers from untrusted IP",
            headers: map[string]string{
                "Remote-User": "hacker",
            },
            trustedProxies: "10.0.0.0/8",
            clientIP:       "1.2.3.4",
            wantUser:       false,
        },
        {
            name: "no headers",
            headers: map[string]string{},
            trustedProxies: "10.0.0.0/8",
            clientIP:       "10.0.0.5",
            wantUser:       false,
        },
    }
    // ... test implementation
}
```

### Integration Tests

1. **Mock Proxy Test**: Simulate Traefik forwarding headers
2. **User Provisioning Test**: Verify users are created from headers
3. **Admin Group Test**: Verify group-to-admin mapping
4. **Spoofing Prevention Test**: Verify headers rejected from non-proxy IPs

## Security Checklist

- [ ] Trusted proxy CIDR validation (CRITICAL)
- [ ] Header names configurable (avoid hardcoded assumptions)
- [ ] No header auth when `AutheliaEnabled=false`
- [ ] Constant-time comparison for sensitive strings
- [ ] Log authentication events for audit
- [ ] Document security implications in README

## Compatibility

This integration pattern also works with:
- **Authentik** - Uses same `X-authentik-*` headers (configurable)
- **Keycloak Gatekeeper** - Similar forward-auth pattern
- **OAuth2-proxy** - `X-Auth-Request-*` headers
- **Pomerium** - `X-Pomerium-*` headers

The configurable header names allow supporting any forward-auth proxy.

## Estimated Effort

| Phase | Description | Effort |
|-------|-------------|--------|
| 1 | Header Authentication | 4-6 hours |
| 2 | UI Configuration | 2-3 hours |
| 3 | Session Integration | 2-3 hours |
| 4 | API Bearer Support | 1-2 hours |
| - | Tests & Documentation | 2-3 hours |
| **Total** | | **11-17 hours** |

## References

- [Authelia Proxy Authorization](https://www.authelia.com/reference/guides/proxy-authorization/)
- [Authelia OAuth 2.0 Bearer Token](https://www.authelia.com/integration/openid-connect/oauth-2.0-bearer-token-usage/)
- [Traefik ForwardAuth](https://www.authelia.com/integration/proxies/traefik/)
- [NGINX Auth Request](https://www.authelia.com/integration/proxies/nginx/)
- [Caddy Forward Auth](https://www.authelia.com/integration/proxies/caddy/)
