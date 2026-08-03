package handlers

import (
	"net/http"

	"github.com/statping-ng/statping-ng/utils"
)

// Security audit event types
const (
	AuditLoginSuccess       = "login_success"
	AuditLoginFailed        = "login_failed"
	AuditLoginRateLimited   = "login_rate_limited"
	AuditLogout             = "logout"
	AuditOAuthLogin         = "oauth_login"
	AuditOAuthUserCreated   = "oauth_user_created"
	AuditOAuthFailed        = "oauth_failed"
	AuditUserCreated        = "user_created"
	AuditUserUpdated        = "user_updated"
	AuditUserDeleted        = "user_deleted"
	AuditAdminPromoted      = "admin_promoted"
	AuditAdminDemoted       = "admin_demoted"
	AuditUserEnabled        = "user_enabled"
	AuditUserDisabled       = "user_disabled"
	AuditPasswordChanged    = "password_changed"
	AuditOAuthConfigChanged = "oauth_config_changed"
	AuditLDAPConfigChanged  = "ldap_config_changed"
	AuditForwardAuthLogin   = "forward_auth_login"
)

var auditLog = utils.Log.WithField("type", "audit")

// AuditLog logs a security-relevant event with structured fields
func AuditLog(event string, r *http.Request, fields map[string]interface{}) {
	entry := auditLog.WithField("event", event)

	if r != nil {
		entry = entry.WithField("ip", getClientIP(r))
		entry = entry.WithField("user_agent", r.UserAgent())
	}

	for k, v := range fields {
		entry = entry.WithField(k, v)
	}

	entry.Info("security event")
}

// AuditLogSimple logs a security event with just username
func AuditLogSimple(event string, r *http.Request, username string) {
	AuditLog(event, r, map[string]interface{}{"username": username})
}
