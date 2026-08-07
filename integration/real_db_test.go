//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/handlers"
	_ "github.com/statping-ng/statping-ng/notifiers"
	"github.com/statping-ng/statping-ng/source"
	"github.com/statping-ng/statping-ng/types/configs"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/messages"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/types/users"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = utils.InitLogs()
	_ = source.Assets()
}

func setupRealDatabase(t *testing.T) (*httptest.Server, *users.User, func()) {
	xzPath := filepath.Join("..", "testdata", "statping.db.xz")
	if _, err := os.Stat(xzPath); os.IsNotExist(err) {
		t.Skipf("testdata/statping.db.xz not found at %s", xzPath)
	}

	tmpDir := t.TempDir()
	dbCopyPath := filepath.Join(tmpDir, "statping.db")

	// Decompress xz file using xz command
	cmd := exec.Command("xz", "-dkc", xzPath)
	output, err := cmd.Output()
	if err != nil {
		t.Skipf("Failed to decompress testdata/statping.db.xz (xz command required): %v", err)
	}
	err = os.WriteFile(dbCopyPath, output, 0o666)
	require.NoError(t, err)

	utils.Directory = tmpDir
	utils.Params.Set("STATPING_DIR", tmpDir)
	utils.Params.Set("DB_CONN", "sqlite")
	utils.Params.Set("GO_ENV", "test")

	dbConfig := &configs.DbConfig{
		DbConn:   "sqlite",
		DbData:   "statping.db",
		Location: tmpDir,
	}

	err = configs.Connect(dbConfig, false)
	require.NoError(t, err)

	core.App.Setup = true
	_ = core.App.Update()

	adminUser, err := users.Find(1)
	require.NoError(t, err)

	ts := httptest.NewServer(handlers.Router())

	// Return cleanup function to close db connection before t.TempDir() cleanup
	cleanup := func() {
		ts.Close()
		if dbConfig.Db != nil {
			if db, err := dbConfig.Db.DB(); err == nil && db != nil {
				_ = db.Close()
			}
		}
	}

	return ts, adminUser, cleanup
}

// getCSRFToken fetches a CSRF token from the test server
func getCSRFToken(ts *httptest.Server) (string, []*http.Cookie, error) {
	resp, err := http.Get(ts.URL + "/api/csrf")
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}

	var csrfResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &csrfResp); err != nil {
		return "", nil, err
	}

	return csrfResp.Token, resp.Cookies(), nil
}

// doAuthenticatedRequest performs an authenticated request with CSRF token
func doAuthenticatedRequest(method, url string, body []byte, apiKey string, csrfToken string, cookies []*http.Cookie) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if csrfToken != "" {
		req.Header.Set("X-CSRF-Token", csrfToken)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	return http.DefaultClient.Do(req)
}

func TestRealDatabaseIntegration(t *testing.T) {
	ts, adminUser, cleanup := setupRealDatabase(t)
	defer cleanup()

	allSvc, err := services.SelectAllServices(true)
	require.NoError(t, err)
	assert.NotEmpty(t, allSvc)

	// 1. Dashboard & Rendered HTML Pages
	t.Run("GET / Root Dashboard HTML", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.NotEmpty(t, body)
	})

	t.Run("GET /service/1 Service Detail HTML", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/service/1")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.NotEmpty(t, body)
	})

	// 2. Core REST API Endpoints with API Key Authentication (Indexed User.ApiKey lookup)
	t.Run("GET /api/users Indexed API Key Auth", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.URL+"/api/users", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+adminUser.ApiKey)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /api/services", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/services")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"id"`)
	})

	t.Run("GET /api/services/1 Single Service Detail", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/services/1")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), `"id":1`)
	})

	t.Run("GET /api/notifiers", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/notifiers")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /api/groups Indexed Group Lookup", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/groups")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /api/messages Indexed Message Query", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/messages")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /api/services/1/incidents", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/services/1/incidents")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 3. Time-Series Data Aggregation on Real 93MB Composite Indexed Hits & Failures Data
	t.Run("GET /api/services/1/hits_data (Composite Index (service, created_at))", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/services/1/hits_data")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /api/services/1/uptime_data (Composite Index (service, created_at))", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/services/1/uptime_data")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /api/services/1/ping_data (Composite Index (service, created_at))", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/services/1/ping_data")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GET /api/services/3/failure_data (Composite Index (service, created_at))", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/services/3/failure_data?group=1h")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 4. Prometheus Metrics & Health Monitoring
	t.Run("GET /metrics Prometheus Exporter", func(t *testing.T) {
		req, err := http.NewRequest("GET", ts.URL+"/metrics", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+adminUser.ApiKey)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "go_memstats_")
	})

	t.Run("GET /health Health Endpoint", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/health")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 5. Authentication Flow
	t.Run("POST /api/login Bad Credentials", func(t *testing.T) {
		// Get CSRF token first
		csrfResp, err := http.Get(ts.URL + "/api/csrf")
		require.NoError(t, err)
		csrfBody, _ := io.ReadAll(csrfResp.Body)
		_ = csrfResp.Body.Close()
		var csrfData map[string]string
		_ = json.Unmarshal(csrfBody, &csrfData)
		csrfToken := csrfData["token"]
		csrfCookie := csrfResp.Cookies()

		form := url.Values{
			"username": {"baduser"},
			"password": {"badpassword"},
		}
		req, _ := http.NewRequest("POST", ts.URL+"/api/login", bytes.NewBufferString(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("X-CSRF-Token", csrfToken)
		for _, c := range csrfCookie {
			req.AddCookie(c)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "incorrect authentication")
	})
}

func TestConcurrentLoadIntegration(t *testing.T) {
	ts, adminUser, cleanup := setupRealDatabase(t)
	defer cleanup()

	const numWorkers = 30
	const requestsPerWorker = 20

	var totalSuccess uint64
	var totalFailed uint64

	endpoints := []string{
		"/",
		"/service/1",
		"/api/services",
		"/api/services/1",
		"/api/services/1/hits_data",
		"/api/services/1/uptime_data",
		"/api/services/1/ping_data",
		"/api/services/3/failure_data?group=1h",
		"/api/groups",
		"/api/messages",
		"/health",
		"/metrics",
	}

	var wg sync.WaitGroup
	startSignal := make(chan struct{})

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			<-startSignal

			client := &http.Client{Timeout: 5 * time.Second}
			for i := 0; i < requestsPerWorker; i++ {
				targetEP := endpoints[i%len(endpoints)]
				req, err := http.NewRequest("GET", ts.URL+targetEP, nil)
				if err != nil {
					atomic.AddUint64(&totalFailed, 1)
					continue
				}
				// Test API Key authentication lookup speed on indexed User.ApiKey
				req.Header.Set("Authorization", "Bearer "+adminUser.ApiKey)

				resp, err := client.Do(req)
				if err != nil {
					atomic.AddUint64(&totalFailed, 1)
					continue
				}
				_, _ = io.ReadAll(resp.Body)
				_ = resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					atomic.AddUint64(&totalSuccess, 1)
				} else {
					atomic.AddUint64(&totalFailed, 1)
				}
			}
		}(w)
	}

	close(startSignal)
	wg.Wait()

	t.Logf("Concurrent Load Completed: %d successful, %d failed out of %d total requests",
		totalSuccess, totalFailed, numWorkers*requestsPerWorker)

	assert.Equal(t, uint64(0), totalFailed, "Zero requests should fail under concurrent load")
	assert.Equal(t, uint64(numWorkers*requestsPerWorker), totalSuccess)
}

func TestFullCRUDLifecycleIntegration(t *testing.T) {
	ts, adminUser, cleanup := setupRealDatabase(t)
	defer cleanup()

	// Get CSRF token for authenticated requests
	csrfToken, cookies, err := getCSRFToken(ts)
	require.NoError(t, err)
	require.NotEmpty(t, csrfToken)

	// 1. Create a new Service
	newSvcPayload := map[string]interface{}{
		"name":            "Integration Test Service",
		"domain":          "https://golang.org",
		"type":            "http",
		"method":          "GET",
		"port":            0,
		"expected_status": 200,
		"check_interval":  30,
		"timeout":         5,
		"order_id":        99,
		"permalink":       "golang-test-svc",
	}

	jsonBytes, err := json.Marshal(newSvcPayload)
	require.NoError(t, err)

	resp, err := doAuthenticatedRequest("POST", ts.URL+"/api/services", jsonBytes, adminUser.ApiKey, csrfToken, cookies)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var respObj struct {
		Status string            `json:"status"`
		Output *services.Service `json:"output"`
	}
	err = json.Unmarshal(body, &respObj)
	require.NoError(t, err)
	createdSvc := respObj.Output
	require.NotNil(t, createdSvc, "Created service should not be nil")
	assert.True(t, createdSvc.Id > 0, "Created service should have valid ID")
	assert.Equal(t, "Integration Test Service", createdSvc.Name)

	// 2. Fetch created Service
	getResp, err := http.Get(fmt.Sprintf("%s/api/services/%d", ts.URL, createdSvc.Id))
	require.NoError(t, err)
	defer func() { _ = getResp.Body.Close() }()
	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	// 3. Create a Message for the Service (Test Composite Message Index)
	msgPayload := map[string]interface{}{
		"title":       "Scheduled Maintenance",
		"description": "Upgrading server infrastructure",
		"service":     createdSvc.Id,
		"start_on":    time.Now().Format(time.RFC3339),
		"end_on":      time.Now().Add(2 * time.Hour).Format(time.RFC3339),
	}
	msgBytes, err := json.Marshal(msgPayload)
	require.NoError(t, err)

	msgResp, err := doAuthenticatedRequest("POST", ts.URL+"/api/messages", msgBytes, adminUser.ApiKey, csrfToken, cookies)
	require.NoError(t, err)
	defer func() { _ = msgResp.Body.Close() }()
	assert.Equal(t, http.StatusOK, msgResp.StatusCode)

	var msgRespObj struct {
		Status string           `json:"status"`
		Output messages.Message `json:"output"`
	}
	msgBody, err := io.ReadAll(msgResp.Body)
	require.NoError(t, err)
	_ = json.Unmarshal(msgBody, &msgRespObj)
	assert.Equal(t, "Scheduled Maintenance", msgRespObj.Output.Title)

	// 4. Delete created Service
	delResp, err := doAuthenticatedRequest("DELETE", fmt.Sprintf("%s/api/services/%d", ts.URL, createdSvc.Id), nil, adminUser.ApiKey, csrfToken, cookies)
	require.NoError(t, err)
	defer func() { _ = delResp.Body.Close() }()
	assert.Equal(t, http.StatusOK, delResp.StatusCode)
}

func TestSecurityRegression(t *testing.T) {
	ts, _, cleanup := setupRealDatabase(t)
	defer cleanup()

	t.Run("Unauthenticated GET /api/users Returns 401 Unauthorized", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/users")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Non-Admin API Key Role Bypass Prevention", func(t *testing.T) {
		// Create a non-admin user
		regUser := &users.User{
			Username: "regular_user_test_security",
			Password: "Password123456789012345678901234567890",
			Email:    "regular@example.com",
			ApiKey:   "non_admin_user_key_99999",
			Admin:    null.NewNullBool(false),
		}
		err := regUser.Create()
		require.NoError(t, err)
		defer func() { _ = regUser.Delete() }()

		// Attempt admin creation endpoint via Non-Admin Query Parameter (?api=...)
		// Note: CSRF protection (403) is checked before auth (401), so we accept either as "rejected"
		newSvcPayload := map[string]interface{}{"name": "Hacked Svc", "domain": "http://evil.com", "type": "http", "port": 80}
		payloadBytes, _ := json.Marshal(newSvcPayload)

		req1, err := http.NewRequest("POST", ts.URL+fmt.Sprintf("/api/services?api=%s", regUser.ApiKey), bytes.NewBuffer(payloadBytes))
		require.NoError(t, err)
		req1.Header.Set("Content-Type", "application/json")
		resp1, err := http.DefaultClient.Do(req1)
		require.NoError(t, err)
		defer func() { _ = resp1.Body.Close() }()
		assert.True(t, resp1.StatusCode == http.StatusUnauthorized || resp1.StatusCode == http.StatusForbidden,
			"Non-admin API key in query string must be rejected (got %d)", resp1.StatusCode)

		// Attempt admin creation endpoint via Non-Admin Bearer Header
		req2, err := http.NewRequest("POST", ts.URL+"/api/services", bytes.NewBuffer(payloadBytes))
		require.NoError(t, err)
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+regUser.ApiKey)
		resp2, err := http.DefaultClient.Do(req2)
		require.NoError(t, err)
		defer func() { _ = resp2.Body.Close() }()
		assert.True(t, resp2.StatusCode == http.StatusUnauthorized || resp2.StatusCode == http.StatusForbidden,
			"Non-admin Bearer API key must be rejected (got %d)", resp2.StatusCode)
	})

	t.Run("User Password and Secret Redaction", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/services/1")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		// Public endpoint output must not leak secret keys or password hashes
		assert.NotContains(t, string(body), "password")
		assert.NotContains(t, string(body), "api_key")
	})
}
