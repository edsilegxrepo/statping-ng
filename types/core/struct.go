package core

import (
	"time"

	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/utils"
)

var App *Core

func New(version, commit string) {
	App = new(Core)
	App.Version = version
	App.Commit = commit
	App.Started = utils.Now()
}

// Core struct contains all the required fields for Statping. All application settings
// will be saved into 1 row in the 'core' table. You can use the core.CoreApp
// global variable to interact with the attributes to the application, such as services.
type Core struct {
	Name           string          `gorm:"not null;column:name" json:"name,omitempty"`
	Description    string          `gorm:"not null;column:description" json:"description,omitempty"`
	ConfigFile     string          `gorm:"column:config" json:"-"`
	ApiSecret      string          `gorm:"column:api_secret" json:"api_secret" scope:"admin"`
	EncryptionKey  string          `gorm:"column:encryption_key" json:"-"` // Never exposed - used only for encrypting secrets
	Style          string          `gorm:"not null;column:style" json:"style,omitempty"`
	Footer         null.NullString `gorm:"column:footer" json:"footer"`
	Domain         string          `gorm:"not null;column:domain" json:"domain"`
	Version        string          `gorm:"column:version" json:"version"`
	Commit         string          `gorm:"-" json:"commit"`
	Language       string          `gorm:"column:language" json:"language"`
	Setup          bool            `gorm:"-" json:"setup"`
	MigrationId    int64           `gorm:"column:migration_id" json:"migration_id,omitempty"`
	AllowReports   null.NullBool   `gorm:"column:allow_reports;default:true" json:"allow_reports,omitempty"`
	SessionTimeout int             `gorm:"column:session_timeout;default:720" json:"session_timeout"`
	DigestEnabled  null.NullBool   `gorm:"column:digest_enabled;default:false" json:"digest_enabled"`
	DigestEmails   string          `gorm:"column:digest_emails" json:"digest_emails"`
	DigestHour     int             `gorm:"column:digest_hour;default:8" json:"digest_hour"`
	CreatedAt      time.Time       `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"column:updated_at" json:"updated_at"`
	Started        time.Time       `gorm:"-" json:"started_on"`
	Notifications  []AllNotifiers  `gorm:"-" json:"-"`
	Integrations   []Integrator    `gorm:"-" json:"-"`

	OAuth       `json:"-"`
	LDAP        `json:"-"`
	LogShipping `json:"-"`
	ForwardAuth `json:"-"`
}

type LDAP struct {
	LdapEnabled                null.NullBool `gorm:"column:ldap_enabled;default:false" json:"ldap_enabled"`
	LdapHost                   string        `gorm:"column:ldap_host" json:"ldap_host"`
	LdapPort                   int           `gorm:"column:ldap_port;default:636" json:"ldap_port"`
	LdapStartTLS               null.NullBool `gorm:"column:ldap_start_tls;default:false" json:"ldap_start_tls"`
	LdapSkipVerify             null.NullBool `gorm:"column:ldap_skip_verify;default:false" json:"ldap_skip_verify"`
	LdapBindDN                 string        `gorm:"column:ldap_bind_dn" json:"ldap_bind_dn" scope:"admin"`
	LdapBindPassword           string        `gorm:"column:ldap_bind_password" json:"ldap_bind_password" scope:"admin"`
	LdapBaseDN                 string        `gorm:"column:ldap_base_dn" json:"ldap_base_dn"`
	LdapUserFilter             string        `gorm:"column:ldap_user_filter" json:"ldap_user_filter"`
	LdapUsernameAttr           string        `gorm:"column:ldap_username_attr" json:"ldap_username_attr"`
	LdapEmailAttr              string        `gorm:"column:ldap_email_attr" json:"ldap_email_attr"`
	LdapAuthorizedGroupEnabled null.NullBool `gorm:"column:ldap_authorized_group_enabled;default:false" json:"ldap_authorized_group_enabled"`
	LdapAuthorizedGroup        string        `gorm:"column:ldap_authorized_group" json:"ldap_authorized_group"`
	LdapTemplate               string        `gorm:"column:ldap_template" json:"ldap_template"`
}

type LogShipping struct {
	LogShipEnabled    null.NullBool `gorm:"column:log_ship_enabled;default:false" json:"log_ship_enabled"`
	LogShipType       string        `gorm:"column:log_ship_type" json:"log_ship_type"`                 // loki, elasticsearch, splunk, cribl, webhook
	LogShipEndpoint   string        `gorm:"column:log_ship_endpoint" json:"log_ship_endpoint"`         // URL
	LogShipToken      string        `gorm:"column:log_ship_token" json:"log_ship_token" scope:"admin"` // Encrypted token
	LogShipIndex      string        `gorm:"column:log_ship_index" json:"log_ship_index"`               // Splunk index
	LogShipSourcetype string        `gorm:"column:log_ship_sourcetype" json:"log_ship_sourcetype"`     // Splunk/Cribl sourcetype
	LogShipLabels     string        `gorm:"column:log_ship_labels" json:"log_ship_labels"`             // key=value,key2=value2
}

// ForwardAuth enables authentication via reverse proxy headers (Authelia, Authentik, etc.)
// The proxy authenticates users (LDAP, OIDC, SAML, local) and passes identity via headers.
type ForwardAuth struct {
	ForwardAuthEnabled        null.NullBool `gorm:"column:forward_auth_enabled;default:false" json:"forward_auth_enabled"`
	ForwardAuthHeaderUser     string        `gorm:"column:forward_auth_header_user;default:Remote-User" json:"forward_auth_header_user"`
	ForwardAuthHeaderEmail    string        `gorm:"column:forward_auth_header_email;default:Remote-Email" json:"forward_auth_header_email"`
	ForwardAuthHeaderGroups   string        `gorm:"column:forward_auth_header_groups;default:Remote-Groups" json:"forward_auth_header_groups"`
	ForwardAuthHeaderName     string        `gorm:"column:forward_auth_header_name;default:Remote-Name" json:"forward_auth_header_name"`
	ForwardAuthAdminGroups    string        `gorm:"column:forward_auth_admin_groups" json:"forward_auth_admin_groups"`       // Semicolon-separated
	ForwardAuthTrustedProxies string        `gorm:"column:forward_auth_trusted_proxies" json:"forward_auth_trusted_proxies"` // CIDR ranges, semicolon-separated
	ForwardAuthLogoutURL      string        `gorm:"column:forward_auth_logout_url" json:"forward_auth_logout_url"`
}

type OAuth struct {
	Providers           string        `gorm:"column:oauth_providers;" json:"oauth_providers"`
	GithubClientID      string        `gorm:"column:gh_client_id" json:"gh_client_id"`
	GithubClientSecret  string        `gorm:"column:gh_client_secret" json:"gh_client_secret" scope:"admin"`
	GithubUsers         string        `gorm:"column:gh_users" json:"gh_users" scope:"admin"`
	GithubOrgs          string        `gorm:"column:gh_orgs" json:"gh_orgs" scope:"admin"`
	GoogleClientID      string        `gorm:"column:google_client_id" json:"google_client_id"`
	GoogleClientSecret  string        `gorm:"column:google_client_secret" json:"google_client_secret" scope:"admin"`
	GoogleUsers         string        `gorm:"column:google_users" json:"google_users" scope:"admin"`
	SlackClientID       string        `gorm:"column:slack_client_id" json:"slack_client_id"`
	SlackClientSecret   string        `gorm:"column:slack_client_secret" json:"slack_client_secret" scope:"admin"`
	SlackTeam           string        `gorm:"column:slack_team" json:"slack_team" scope:"admin"`
	SlackUsers          string        `gorm:"column:slack_users" json:"slack_users" scope:"admin"`
	CustomName          string        `gorm:"column:custom_name" json:"custom_name"`
	CustomClientID      string        `gorm:"column:custom_client_id" json:"custom_client_id"`
	CustomClientSecret  string        `gorm:"column:custom_client_secret" json:"custom_client_secret" scope:"admin"`
	CustomEndpointAuth  string        `gorm:"column:custom_endpoint_auth" json:"custom_endpoint_auth"`
	CustomEndpointToken string        `gorm:"column:custom_endpoint_token" json:"custom_endpoint_token" scope:"admin"`
	CustomScopes        string        `gorm:"column:custom_scopes" json:"custom_scopes"`
	CustomIsOpenID      null.NullBool `gorm:"column:custom_open_id" json:"custom_open_id"`
}

// AllNotifiers contains all the Notifiers loaded
type AllNotifiers interface{}

type Integrator interface{}

func (Core) TableName() string {
	return "core"
}
