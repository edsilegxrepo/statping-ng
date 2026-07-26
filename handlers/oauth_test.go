package handlers

import (
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/stretchr/testify/assert"
)

func TestGenerateOAuthState(t *testing.T) {
	state1, err := generateOAuthState()
	assert.Nil(t, err, "Should generate OAuth state without error")
	assert.NotEmpty(t, state1, "Generated state should not be empty")

	state2, err := generateOAuthState()
	assert.Nil(t, err, "Should generate second OAuth state without error")
	assert.NotEqual(t, state1, state2, "Generated states should be unique")
}

func TestValidateOAuthState(t *testing.T) {
	// Generate a valid state
	state, err := generateOAuthState()
	assert.Nil(t, err, "Should generate OAuth state without error")

	// First validation should succeed
	assert.True(t, validateOAuthState(state), "First validation should succeed")

	// Second validation should fail (one-time use)
	assert.False(t, validateOAuthState(state), "Second validation should fail (state consumed)")
}

func TestValidateOAuthStateEmpty(t *testing.T) {
	assert.False(t, validateOAuthState(""), "Empty state should not validate")
}

func TestValidateOAuthStateInvalid(t *testing.T) {
	assert.False(t, validateOAuthState("invalid-state-token"), "Invalid state should not validate")
}

func TestValidateOAuthStateExpired(t *testing.T) {
	// Manually add an expired state
	expiredState := "expired-test-state-" + time.Now().String()
	oauthStateStoreLock.Lock()
	oauthStateStore[expiredState] = time.Now().Add(-1 * time.Hour)
	oauthStateStoreLock.Unlock()

	assert.False(t, validateOAuthState(expiredState), "Expired state should not validate")
}

func TestCleanupExpiredOAuthStates(t *testing.T) {
	// Add some expired states with unique names
	prefix := time.Now().UnixNano()
	expired1 := "expired1-" + string(rune(prefix))
	expired2 := "expired2-" + string(rune(prefix))
	valid := "valid-" + string(rune(prefix))

	oauthStateStoreLock.Lock()
	oauthStateStore[expired1] = time.Now().Add(-1 * time.Hour)
	oauthStateStore[expired2] = time.Now().Add(-2 * time.Hour)
	oauthStateStore[valid] = time.Now().Add(1 * time.Hour)
	oauthStateStoreLock.Unlock()

	cleanupExpiredOAuthStates()

	oauthStateStoreLock.RLock()
	_, exists1 := oauthStateStore[expired1]
	_, exists2 := oauthStateStore[expired2]
	_, existsValid := oauthStateStore[valid]
	oauthStateStoreLock.RUnlock()

	assert.False(t, exists1, "expired1 should have been cleaned up")
	assert.False(t, exists2, "expired2 should have been cleaned up")
	assert.True(t, existsValid, "valid state should still exist")

	// Cleanup test state
	oauthStateStoreLock.Lock()
	delete(oauthStateStore, valid)
	oauthStateStoreLock.Unlock()
}

func TestOAuthRoutes(t *testing.T) {
	tests := []HTTPTest{
		{
			Name: "OAuth Save",
			URL:  "/api/oauth",
			Body: `{
						"gh_client_id": "githubid",
						"gh_client_secret": "githubsecret",
						"google_client_id": "googleid",
						"google_client_secret": "googlesecret",
						"oauth_domains": "gmail.com,yahoo.com,socialeck.com",
						"oauth_providers": "local,slack,google,github",
						"slack_client_id": "example.iddd",
						"slack_client_secret": "exampleeesecret",
						"slack_team": "dev"
					}`,
			Method:         "POST",
			ExpectedStatus: 200,
			BeforeTest:     SetTestENV,
		},
		{
			Name:             "OAuth Values",
			URL:              "/api/oauth",
			Method:           "GET",
			ExpectedStatus:   200,
			ExpectedContains: []string{`"slack_client_id":"example.iddd"`},
			AfterTest:        UnsetTestENV,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			res, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
			t.Log(res)
		})
	}
}

func TestOAuthLoginRoutes(t *testing.T) {
	ensureHandlerSetup(t)

	tests := []HTTPTest{
		{
			Name:           "OAuth GitHub Callback - Invalid State",
			URL:            "/oauth/github?state=invalid&code=testcode",
			Method:         "GET",
			ExpectedStatus: 200,
			ExpectedContains: []string{"invalid or expired OAuth state"},
			NoAuth:         true,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
		})
	}
}

func TestValidateGithub(t *testing.T) {
	// Save original OAuth config
	originalOAuth := core.App.OAuth
	defer func() { core.App.OAuth = originalOAuth }()

	t.Run("no restrictions allows all", func(t *testing.T) {
		core.App.OAuth.GithubUsers = ""
		core.App.OAuth.GithubOrgs = ""
		user := githubUser{Login: "anyuser", Name: "Any User"}
		assert.True(t, validateGithub(user, nil))
	})

	t.Run("allowed user passes", func(t *testing.T) {
		core.App.OAuth.GithubUsers = "alloweduser,anotheruser"
		core.App.OAuth.GithubOrgs = ""
		user := githubUser{Login: "alloweduser", Name: "Allowed User"}
		assert.True(t, validateGithub(user, nil))
	})

	t.Run("disallowed user fails", func(t *testing.T) {
		core.App.OAuth.GithubUsers = "alloweduser"
		core.App.OAuth.GithubOrgs = ""
		user := githubUser{Login: "notallowed", Name: "Not Allowed"}
		assert.False(t, validateGithub(user, nil))
	})

	t.Run("allowed org passes", func(t *testing.T) {
		core.App.OAuth.GithubUsers = ""
		core.App.OAuth.GithubOrgs = "myorg,otherorg"
		user := githubUser{Login: "someuser", Name: "Some User"}
		orgs := []githubOrgs{{Login: "myorg"}}
		assert.True(t, validateGithub(user, orgs))
	})

	t.Run("disallowed org fails", func(t *testing.T) {
		core.App.OAuth.GithubUsers = ""
		core.App.OAuth.GithubOrgs = "allowedorg"
		user := githubUser{Login: "someuser", Name: "Some User"}
		orgs := []githubOrgs{{Login: "differentorg"}}
		assert.False(t, validateGithub(user, orgs))
	})

	t.Run("case insensitive user match", func(t *testing.T) {
		core.App.OAuth.GithubUsers = "AllowedUser"
		core.App.OAuth.GithubOrgs = ""
		user := githubUser{Login: "alloweduser", Name: "Allowed User"}
		assert.True(t, validateGithub(user, nil))
	})
}

func TestValidateGoogle(t *testing.T) {
	originalOAuth := core.App.OAuth
	defer func() { core.App.OAuth = originalOAuth }()

	t.Run("no restrictions allows all", func(t *testing.T) {
		core.App.OAuth.GoogleUsers = ""
		info := googleUserInfo{Email: "anyone@example.com", Name: "Anyone"}
		assert.True(t, validateGoogle(info))
	})

	t.Run("allowed email passes", func(t *testing.T) {
		core.App.OAuth.GoogleUsers = "allowed@example.com,other@example.com"
		info := googleUserInfo{Email: "allowed@example.com", Name: "Allowed"}
		assert.True(t, validateGoogle(info))
	})

	t.Run("allowed domain passes via Hd", func(t *testing.T) {
		core.App.OAuth.GoogleUsers = "example.com"
		info := googleUserInfo{Email: "user@example.com", Hd: "example.com"}
		assert.True(t, validateGoogle(info))
	})

	t.Run("disallowed email fails", func(t *testing.T) {
		core.App.OAuth.GoogleUsers = "allowed@example.com"
		info := googleUserInfo{Email: "notallowed@example.com", Name: "Not Allowed"}
		assert.False(t, validateGoogle(info))
	})

	t.Run("case insensitive email match", func(t *testing.T) {
		core.App.OAuth.GoogleUsers = "Allowed@Example.com"
		info := googleUserInfo{Email: "allowed@example.com", Name: "Allowed"}
		assert.True(t, validateGoogle(info))
	})
}

func TestValidateSlack(t *testing.T) {
	originalOAuth := core.App.OAuth
	defer func() { core.App.OAuth = originalOAuth }()

	t.Run("no restrictions allows all", func(t *testing.T) {
		core.App.OAuth.SlackUsers = ""
		id := slackIdentity{Ok: true, User: struct {
			Name  string `json:"name"`
			ID    string `json:"id"`
			Email string `json:"email"`
		}{Name: "anyone", Email: "anyone@example.com"}}
		assert.True(t, validateSlack(id))
	})

	t.Run("allowed email passes", func(t *testing.T) {
		core.App.OAuth.SlackUsers = "allowed@example.com"
		id := slackIdentity{Ok: true, User: struct {
			Name  string `json:"name"`
			ID    string `json:"id"`
			Email string `json:"email"`
		}{Name: "allowed", Email: "allowed@example.com"}}
		assert.True(t, validateSlack(id))
	})

	t.Run("allowed name passes", func(t *testing.T) {
		core.App.OAuth.SlackUsers = "alloweduser"
		id := slackIdentity{Ok: true, User: struct {
			Name  string `json:"name"`
			ID    string `json:"id"`
			Email string `json:"email"`
		}{Name: "alloweduser", Email: "other@example.com"}}
		assert.True(t, validateSlack(id))
	})

	t.Run("disallowed user fails", func(t *testing.T) {
		core.App.OAuth.SlackUsers = "alloweduser"
		id := slackIdentity{Ok: true, User: struct {
			Name  string `json:"name"`
			ID    string `json:"id"`
			Email string `json:"email"`
		}{Name: "notallowed", Email: "notallowed@example.com"}}
		assert.False(t, validateSlack(id))
	})
}
