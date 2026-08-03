package users

import (
	"time"

	"github.com/statping-ng/statping-ng/types/null"
)

// AuthProvider constants for user authentication
const (
	AuthProviderLocal       = "local"
	AuthProviderLDAP        = "ldap"
	AuthProviderOAuthGoogle = "oauth_google"
	AuthProviderOAuthGitHub = "oauth_github"
	AuthProviderOAuthSlack  = "oauth_slack"
	AuthProviderOAuthCustom = "oauth_custom"
	AuthProviderForwardAuth = "forward_auth"
)

// AuthProviderInfo contains display information for an auth provider
type AuthProviderInfo struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// GetAuthProviders returns all available authentication providers
func GetAuthProviders() []AuthProviderInfo {
	return []AuthProviderInfo{
		{Value: AuthProviderLocal, Label: "Local"},
		{Value: AuthProviderLDAP, Label: "LDAP"},
		{Value: AuthProviderOAuthGoogle, Label: "OAuth - Google"},
		{Value: AuthProviderOAuthGitHub, Label: "OAuth - GitHub"},
		{Value: AuthProviderOAuthSlack, Label: "OAuth - Slack"},
		{Value: AuthProviderOAuthCustom, Label: "OAuth - Custom"},
		{Value: AuthProviderForwardAuth, Label: "Forward Auth"},
	}
}

// User is the main struct for Users
type User struct {
	Id                 int64         `gorm:"primary_key;column:id" json:"id"`
	Username           string        `gorm:"type:varchar(100);unique;column:username;" json:"username,omitempty"`
	Password           string        `gorm:"column:password" json:"password,omitempty" private:"true" scope:"admin"`
	Email              string        `gorm:"type:varchar(100);column:email" json:"email,omitempty" scope:"user,admin"`
	ApiKey             string        `gorm:"uniqueIndex;type:varchar(100);column:api_key" json:"api_key,omitempty" private:"true" scope:"admin"`
	Scopes             string        `gorm:"column:scopes" json:"scopes,omitempty"`
	AuthProvider       string        `gorm:"type:varchar(50);column:auth_provider;default:'local'" json:"auth_provider,omitempty"`
	Admin              null.NullBool `gorm:"column:administrator" json:"admin,omitempty"`
	Enabled            null.NullBool `gorm:"column:enabled;default:true" json:"enabled,omitempty"`
	ForwardAuthManaged null.NullBool `gorm:"column:forward_auth_managed;default:false" json:"forward_auth_managed,omitempty"`
	CreatedAt          time.Time     `gorm:"column:created_at" json:"created_at"`
	UpdatedAt          time.Time     `gorm:"column:updated_at" json:"updated_at"`
	Token              string        `gorm:"-" json:"token"`
}
