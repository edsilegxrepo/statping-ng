package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	_ "github.com/statping-ng/statping-ng/notifiers"
	"github.com/statping-ng/statping-ng/source"
	"github.com/statping-ng/statping-ng/types/checkins"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/groups"
	"github.com/statping-ng/statping-ng/types/messages"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/types/users"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	dir             string
	setupOnce       sync.Once
	setupErr        error
	testSetupMutex  sync.Mutex
)

func init() {
	_ = utils.InitLogs()
	_ = source.Assets()
	core.New("test", "testcommithere")
}

func TestFailedHTTPServer(t *testing.T) {
	var err error
	go func() {
		_ = RunHTTPServer()
	}()
	go func() {
		time.Sleep(3 * time.Second)
		StopHTTPServer(nil)
	}()
	assert.Nil(t, err)
}

func TestSetupRoutes(t *testing.T) {
	form := url.Values{}
	form.Add("db_host", utils.Params.GetString("DB_HOST"))
	form.Add("db_user", utils.Params.GetString("DB_USER"))
	form.Add("db_password", utils.Params.GetString("DB_PASS"))
	form.Add("db_database", utils.Params.GetString("DB_DATABASE"))
	form.Add("db_connection", utils.Params.GetString("DB_CONN"))
	form.Add("db_port", utils.Params.GetString("DB_PORT"))
	form.Add("project", "Tester")
	form.Add("username", "admin")
	form.Add("password", "Password123456789012345678901234567890")
	form.Add("sample_data", "on")
	form.Add("description", "This is an awesome test")
	form.Add("domain", "http://localhost:8080")
	form.Add("email", "info@example.com")

	badForm := url.Values{}
	badForm.Add("db_host", "badconnection")
	badForm.Add("db_user", utils.Params.GetString("DB_USER"))
	badForm.Add("db_password", utils.Params.GetString("DB_PASS"))
	badForm.Add("db_database", utils.Params.GetString("DB_DATABASE"))
	badForm.Add("db_connection", "mysql")
	badForm.Add("db_port", utils.Params.GetString("DB_PORT"))
	badForm.Add("project", "Tester")
	badForm.Add("username", "admin")
	badForm.Add("password", "password123")
	badForm.Add("sample_data", "on")
	badForm.Add("description", "This is an awesome test")
	badForm.Add("domain", "http://localhost:8080")
	badForm.Add("email", "info@example.com")

	tests := []HTTPTest{
		{
			Name:           "Statping Check",
			URL:            "/api",
			Method:         "GET",
			ExpectedStatus: 200,
			FuncTest: func(t *testing.T) error {
				if core.App.Setup {
					return errors.New("core has already been setup")
				}
				return nil
			},
		},
		{
			Name:             "Statping Error Setup",
			URL:              "/api/setup",
			Method:           "POST",
			Body:             badForm.Encode(),
			ExpectedStatus:   500,
			ExpectedContains: []string{BadJSONDatabase},
			HttpHeaders:      []string{"Content-Type=application/x-www-form-urlencoded"},
		},
		{
			Name:   "Statping Run Setup",
			URL:    "/api/setup",
			Method: "POST",
			Body:   form.Encode(),
			// ExpectedStatus: 200,
			HttpHeaders:   []string{"Content-Type=application/x-www-form-urlencoded"},
			ExpectedFiles: []string{utils.Directory + "/config.yml"},
			FuncTest: func(t *testing.T) error {
				if !core.App.Setup {
					return errors.New("core has not been setup")
				}
				if core.App.ApiSecret == "" {
					return errors.New("API Key has not been set")
				}
				if len(services.AllInOrder()) == 0 {
					return errors.New("no services where found")
				}
				if len(users.All()) == 0 {
					return errors.New("no users where found")
				}
				if len(groups.All()) == 0 {
					return errors.New("no groups where found")
				}
				return nil
			},
			AfterTest: StopServices,
		},
	}

	for _, v := range tests {
		t.Run(v.Name, func(t *testing.T) {
			_, t, err := RunHTTPTest(v, t)
			assert.Nil(t, err)
			if err != nil {
				t.FailNow()
			}
		})
	}
}

func TestMainApiRoutes(t *testing.T) {
	ensureHandlerSetup(t)
	date := time.Now().Format("2006-01")
	tests := []HTTPTest{
		{
			Name:             "Statping Details",
			URL:              "/api",
			Method:           "GET",
			ExpectedStatus:   200,
			ExpectedContains: []string{`"description":"This is an awesome test"`},
			FuncTest: func(t *testing.T) error {
				if !core.App.Setup {
					return errors.New("database is not setup")
				}
				return nil
			},
		},
		{
			Name:           "Statping Renew API Keys",
			URL:            "/api/renew",
			Method:         "POST",
			ExpectedStatus: 200,
			BeforeTest:     SetTestENV,
			SecureRoute:    true,
		},
		{
			Name:           "Update Core",
			URL:            "/api/core",
			Method:         "POST",
			ExpectedStatus: 200,
			Body: `{
					"name": "Updated Core"
				}`,
			AfterTest: func(t *testing.T) error {
				assert.Equal(t, "Updated Core", core.App.Name)
				return nil
			},
		},
		{
			Name:           "Update Core - All Fields",
			URL:            "/api/core",
			Method:         "POST",
			ExpectedStatus: 200,
			Body: `{
					"name": "Full Update Core",
					"description": "Updated description",
					"style": "dark",
					"footer": "Custom footer",
					"domain": "https://example.com",
					"language": "en",
					"allow_reports": false
				}`,
			AfterTest: func(t *testing.T) error {
				assert.Equal(t, "Full Update Core", core.App.Name)
				assert.Equal(t, "Updated description", core.App.Description)
				return nil
			},
		},
		{
			Name:             "Update Core - Invalid JSON",
			URL:              "/api/core",
			Method:           "POST",
			ExpectedStatus:   422,
			Body:             `{invalid json}`,
			ExpectedContains: []string{"error"},
		},
		{
			Name:           "Get OAuth Settings",
			URL:            "/api/oauth",
			Method:         "GET",
			ExpectedStatus: 200,
			BeforeTest:     SetTestENV,
		},
		{
			Name:           "Update OAuth Settings",
			URL:            "/api/oauth",
			Method:         "POST",
			ExpectedStatus: 200,
			Body: `{
					"gh_client_id": "test_github_client",
					"gh_client_secret": "test_github_secret",
					"google_client_id": "test_google_client",
					"google_client_secret": "test_google_secret"
				}`,
			BeforeTest: SetTestENV,
		},
		{
			Name:             "Update OAuth Settings - Invalid JSON",
			URL:              "/api/oauth",
			Method:           "POST",
			ExpectedStatus:   422,
			Body:             `{invalid json}`,
			ExpectedContains: []string{"error"},
			BeforeTest:       SetTestENV,
		},
		{
			Name:             "Health Check endpoint",
			URL:              "/health",
			Method:           "GET",
			ExpectedStatus:   200,
			ExpectedContains: []string{`"online":true`, `"setup":true`},
		},
		{
			Name:             "Logs endpoint",
			URL:              "/api/logs",
			Method:           "GET",
			ExpectedStatus:   200,
			GreaterThan:      20,
			ExpectedContains: []string{date},
		},
		{
			Name:             "Logs endpoint",
			URL:              "/api/logs",
			Method:           "GET",
			ExpectedStatus:   200,
			GreaterThan:      20,
			ExpectedContains: []string{date},
		},
		{
			Name:             "Logs Last Line endpoint",
			URL:              "/api/logs/last",
			Method:           "GET",
			ExpectedStatus:   200,
			ExpectedContains: []string{date},
		},
		{
			Name:           "Prometheus Export Metrics",
			URL:            "/metrics",
			Method:         "GET",
			BeforeTest:     SetTestENV,
			AfterTest:      UnsetTestENV,
			ExpectedStatus: 200,
			ExpectedContains: []string{
				`go_goroutines`,
				`go_memstats_alloc_bytes`,
				`go_threads`,
			},
		},
		{
			Name:           "Index Page",
			URL:            "/",
			Method:         "GET",
			ExpectedStatus: 200,
		},
		{
			Name:           "Export Settings",
			URL:            "/api/settings/export",
			Method:         "GET",
			ExpectedStatus: 200,
			BeforeTest:     SetTestENV,
			AfterTest:      UnsetTestENV,
			ResponseFunc: func(r *httptest.ResponseRecorder, t *testing.T, bytes []byte) error {
				var data ExportData
				err := json.Unmarshal(r.Body.Bytes(), &data)
				require.Nil(t, err)
				assert.Len(t, data.Services, len(services.All()))
				assert.Len(t, data.Groups, len(groups.All()))
				assert.Len(t, data.Notifiers, len(services.AllNotifiers()))
				assert.Len(t, data.Users, len(users.All()))
				assert.Len(t, data.Messages, len(messages.All()))
				assert.Len(t, data.Checkins, len(checkins.All()))
				return nil
			},
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

type HttpFuncTest func(*testing.T) error

type ResponseFunc func(*httptest.ResponseRecorder, *testing.T, []byte) error

// HTTPTest contains all the parameters for a HTTP Unit Test
type HTTPTest struct {
	Name                string
	URL                 string
	Method              string
	Body                string
	ExpectedStatus      int
	ExpectedContains    []string
	ExpectedNotContains []string
	HttpHeaders         []string
	ExpectedFiles       []string
	FuncTest            HttpFuncTest
	BeforeTest          HttpFuncTest
	AfterTest           HttpFuncTest
	ResponseFunc        ResponseFunc
	ResponseLen         int
	GreaterThan         int
	SecureRoute         bool
	Skip                bool
	NoAuth              bool // Skip auto-adding authentication and CSRF tokens
}

func logTest(t *testing.T, err error) error {
	return err
}

// RunHTTPTest accepts a HTTPTest type to execute the HTTP request
func RunHTTPTest(test HTTPTest, t *testing.T) (string, *testing.T, error) {
	if test.Skip {
		t.SkipNow()
	}
	if test.BeforeTest != nil {
		if err := test.BeforeTest(t); err != nil {
			return "", t, logTest(t, err)
		}
	}

	rr, err := Request(test)
	if err != nil {
		return "", t, logTest(t, err)
	}
	defer func() { _ = rr.Result().Body.Close() }()

	if test.ExpectedStatus != 0 {
		if test.ExpectedStatus != rr.Result().StatusCode {
			assert.Equal(t, test.ExpectedStatus, rr.Result().StatusCode)
			return "", t, fmt.Errorf("status code %v does not match %v", rr.Result().StatusCode, test.ExpectedStatus)
		}
	}

	body, err := io.ReadAll(rr.Result().Body)
	if err != nil {
		assert.Nil(t, err)
		return "", t, logTest(t, err)
	}

	stringBody := string(body)

	if len(test.ExpectedContains) != 0 {
		for _, v := range test.ExpectedContains {
			assert.Contains(t, stringBody, v)
		}
	}
	if len(test.ExpectedNotContains) != 0 {
		for _, v := range test.ExpectedNotContains {
			assert.NotContains(t, stringBody, v)
		}
	}
	if len(test.ExpectedFiles) != 0 {
		for _, v := range test.ExpectedFiles {
			assert.FileExists(t, v)
		}
	}
	if test.FuncTest != nil {
		err := test.FuncTest(t)
		assert.Nil(t, err)
	}
	if test.ResponseFunc != nil {
		err := test.ResponseFunc(rr, t, body)
		assert.Nil(t, err)
	}

	if test.ResponseLen != 0 {
		var respArray []interface{}
		err := json.Unmarshal(body, &respArray)
		assert.Nil(t, err)
		assert.Equal(t, test.ResponseLen, len(respArray))
	}

	if test.GreaterThan != 0 {
		var respArray []interface{}
		err := json.Unmarshal(body, &respArray)
		assert.Nil(t, err)
		assert.GreaterOrEqual(t, len(respArray), test.GreaterThan)
	}

	if test.AfterTest != nil {
		if err := test.AfterTest(t); err != nil {
			return "", t, logTest(t, err)
		}
	}
	return stringBody, t, logTest(t, err)
}


func Request(test HTTPTest) (*httptest.ResponseRecorder, error) {
	return RequestWithAuth(test, !test.NoAuth)
}

func RequestWithAuth(test HTTPTest, addAuth bool) (*httptest.ResponseRecorder, error) {
	req, err := http.NewRequest(test.Method, test.URL, strings.NewReader(test.Body))
	if err != nil {
		return nil, err
	}
	if len(test.HttpHeaders) != 0 {
		for _, v := range test.HttpHeaders {
			splits := strings.Split(v, "=")
			req.Header.Set(splits[0], splits[1])
		}
	}

	// Ensure Host is set for CSRF validation
	if req.Host == "" {
		req.Host = "localhost"
	}

	// For state-changing requests, always set Origin to pass CSRF validation
	// This allows auth tests (NoAuth=true) to properly test authentication
	// rather than being blocked by CSRF middleware first
	if test.Method != "GET" && test.Method != "HEAD" && test.Method != "OPTIONS" {
		if req.Header.Get("Origin") == "" && req.Header.Get("Referer") == "" {
			req.Header.Set("Origin", "http://"+req.Host)
		}
	}

	// Skip auth header if addAuth is false (unauthenticated route tests)
	if !addAuth {
		rr := httptest.NewRecorder()
		Router().ServeHTTP(rr, req)
		return rr, nil
	}

	// For authenticated requests, add auth header if not already set
	if core.App != nil && core.App.ApiSecret != "" && req.Header.Get("Authorization") == "" && !strings.Contains(test.URL, "api=") {
		req.Header.Set("Authorization", "Bearer "+core.App.ApiSecret)
	}

	rr := httptest.NewRecorder()
	Router().ServeHTTP(rr, req)
	return rr, err
}

func ensureHandlerSetup(t *testing.T) {
	testSetupMutex.Lock()
	defer testSetupMutex.Unlock()

	// If already set up (e.g., by TestSetupRoutes), skip
	if core.App != nil && core.App.Setup && len(services.Services()) >= 6 {
		return
	}

	// Use sync.Once to ensure setup runs exactly once per test run
	setupOnce.Do(func() {
		setupErr = doHandlerSetup(t)
	})

	if setupErr != nil {
		t.Fatalf("handler setup failed: %v", setupErr)
	}

	// Verify setup completed correctly
	if len(services.Services()) < 6 {
		t.Fatalf("handler setup incomplete: expected >= 6 services, got %d", len(services.Services()))
	}
}

func doHandlerSetup(t *testing.T) error {
	// Clear any stale cache state
	services.ClearCache()

	core.App.Setup = false
	form := url.Values{}
	form.Add("db_host", utils.Params.GetString("DB_HOST"))
	form.Add("db_user", utils.Params.GetString("DB_USER"))
	form.Add("db_password", utils.Params.GetString("DB_PASS"))
	form.Add("db_database", utils.Params.GetString("DB_DATABASE"))
	form.Add("db_connection", utils.Params.GetString("DB_CONN"))
	form.Add("db_port", utils.Params.GetString("DB_PORT"))
	form.Add("project", "Tester")
	form.Add("username", "admin")
	form.Add("password", "Password123456789012345678901234567890")
	form.Add("sample_data", "on")
	form.Add("description", "This is an awesome test")
	form.Add("domain", "http://localhost:8080")
	form.Add("email", "info@example.com")

	req, err := http.NewRequest("POST", "/api/setup", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	Router().ServeHTTP(rr, req)
	if rr.Code != 200 {
		return fmt.Errorf("setup returned status %d: %s", rr.Code, rr.Body.String())
	}
	return nil
}

func SetTestENV(t *testing.T) error {
	// No-op: tests use production behavior with proper CSRF handling
	return nil
}

func UnsetTestENV(t *testing.T) error {
	// No-op: tests use production behavior with proper CSRF handling
	return nil
}

func StopServices(t *testing.T) error {
	for _, s := range services.All() {
		s.Close()
	}
	return nil
}

var (
	Success = `"status":"success"`

	MethodCreate = `"method":"create"`
	MethodUpdate = `"method":"update"`
	MethodDelete = `"method":"delete"`

	BadJSON         = `{incorrect: JSON %%% formatting, [&]}`
	BadJSONResponse = `{"error":"could not decode incoming JSON"}`
	BadJSONDatabase = `{"error":"error connecting to database`
)
