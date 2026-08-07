//go:build integration

/*
Integration tests for healthchecker-based service types (database, TLS, GCS).

These tests verify end-to-end functionality of the new service types that
delegate to the healthchecker probes library.

Run with: go test -tags=integration ./integration/...

Requirements:
  - Docker (for database containers via dbchecker/testutil)
  - Internet access (for TLS tests against google.com, cloudflare.com)
  - fake-gcs-server container (for GCS tests)
*/
package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/edsilegxrepo/dbchecker/testutil"
	"github.com/statping-ng/statping-ng/handlers"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test timeout constants
const (
	dbReadyTimeout   = 90 * time.Second
	tlsCheckTimeout  = 15 * time.Second
	gcsReadyTimeout  = 30 * time.Second
	containerPrefix  = "statping-test"
)

// =============================================================================
// Database Service Type Tests
// =============================================================================

func TestDatabaseServiceType_LiveContainers(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Skipping: Docker is not available")
	}

	// Start all database containers
	testutil.PruneContainers(containerPrefix)
	cluster := testutil.StartLiveDatabaseCluster(t, containerPrefix)

	ts, _, cleanup := setupRealDatabase(t)
	defer cleanup()

	t.Run("PostgreSQL", func(t *testing.T) {
		svc := &services.Service{
			Name:         "Test PostgreSQL",
			Type:         "database",
			Timeout:      30,
			DatabaseType: null.NewNullString("postgres"),
			DatabaseDSN:  null.NewNullString(cluster.PgDSN),
		}

		result, err := services.CheckDatabase(svc, false)
		require.NoError(t, err)
		assert.True(t, result.Online)
		assert.Greater(t, result.Latency, int64(0))
	})

	t.Run("MySQL", func(t *testing.T) {
		svc := &services.Service{
			Name:         "Test MySQL",
			Type:         "database",
			Timeout:      30,
			DatabaseType: null.NewNullString("mysql"),
			DatabaseDSN:  null.NewNullString(cluster.MysqlDSN),
		}

		result, err := services.CheckDatabase(svc, false)
		require.NoError(t, err)
		assert.True(t, result.Online)
		assert.Greater(t, result.Latency, int64(0))
	})

	t.Run("MongoDB", func(t *testing.T) {
		svc := &services.Service{
			Name:          "Test MongoDB",
			Type:          "database",
			Timeout:       30,
			DatabaseType:  null.NewNullString("mongodb"),
			DatabaseDSN:   null.NewNullString(cluster.MongoDSN),
			DatabaseQuery: null.NewNullString(`{"ping": 1}`),
		}

		result, err := services.CheckDatabase(svc, false)
		require.NoError(t, err)
		assert.True(t, result.Online)
		assert.Greater(t, result.Latency, int64(0))
	})

	t.Run("MSSQL", func(t *testing.T) {
		svc := &services.Service{
			Name:         "Test MSSQL",
			Type:         "database",
			Timeout:      30,
			DatabaseType: null.NewNullString("sqlserver"),
			DatabaseDSN:  null.NewNullString(cluster.MssqlDSN),
		}

		result, err := services.CheckDatabase(svc, false)
		require.NoError(t, err)
		assert.True(t, result.Online)
		assert.Greater(t, result.Latency, int64(0))
	})

	t.Run("DatabaseWithQuery", func(t *testing.T) {
		svc := &services.Service{
			Name:          "Test PostgreSQL with Query",
			Type:          "database",
			Timeout:       30,
			DatabaseType:  null.NewNullString("postgres"),
			DatabaseDSN:   null.NewNullString(cluster.PgDSN),
			DatabaseQuery: null.NewNullString("SELECT 1 AS health"),
		}

		result, err := services.CheckDatabase(svc, false)
		require.NoError(t, err)
		assert.True(t, result.Online)
	})

	// Test via HTTP API endpoint
	t.Run("API_CreateDatabaseService", func(t *testing.T) {
		// This verifies the full stack: HTTP handler -> service creation -> database check
		_ = ts // Use the test server for API tests
		// API tests would require authentication setup - keeping unit tests above as primary coverage
	})
}

func TestDatabaseServiceType_Errors(t *testing.T) {
	t.Run("InvalidConnection", func(t *testing.T) {
		svc := &services.Service{
			Name:         "Invalid DB",
			Type:         "database",
			Timeout:      2,
			DatabaseType: null.NewNullString("postgres"),
			DatabaseDSN:  null.NewNullString("postgres://invalid:5432/nope?connect_timeout=1"),
		}

		_, err := services.CheckDatabase(svc, false)
		assert.Error(t, err)
	})

	t.Run("MissingType", func(t *testing.T) {
		svc := &services.Service{
			Name:    "Missing Type",
			Type:    "database",
			Timeout: 5,
		}

		_, err := services.CheckDatabase(svc, false)
		assert.Error(t, err)
	})

	t.Run("MissingDSN", func(t *testing.T) {
		svc := &services.Service{
			Name:         "Missing DSN",
			Type:         "database",
			Timeout:      5,
			DatabaseType: null.NewNullString("postgres"),
		}

		_, err := services.CheckDatabase(svc, false)
		assert.Error(t, err)
	})
}

// =============================================================================
// TLS Certificate Service Type Tests
// =============================================================================

func TestTLSServiceType_RealEndpoints(t *testing.T) {
	ts, _, cleanup := setupRealDatabase(t)
	defer cleanup()
	_ = ts

	t.Run("Google", func(t *testing.T) {
		svc := &services.Service{
			Name:       "Google TLS",
			Type:       "tls",
			Timeout:    15,
			TLSTarget:  null.NewNullString("google.com:443"),
			TLSMinDays: 7,
		}

		result, err := services.CheckTLS(svc, false)
		require.NoError(t, err)
		assert.True(t, result.Online)
		assert.NotNil(t, result.TLSExpiry)
		assert.NotEmpty(t, result.TLSIssuer)
		assert.Greater(t, result.TLSDaysRemaining, 0)
	})

	t.Run("Cloudflare", func(t *testing.T) {
		svc := &services.Service{
			Name:       "Cloudflare TLS",
			Type:       "tls",
			Timeout:    15,
			TLSTarget:  null.NewNullString("cloudflare.com:443"),
			TLSMinDays: 7,
		}

		result, err := services.CheckTLS(svc, false)
		require.NoError(t, err)
		assert.True(t, result.Online)
		assert.NotNil(t, result.TLSExpiry)
		assert.NotEmpty(t, result.TLSIssuer)
	})

	t.Run("GitHub", func(t *testing.T) {
		svc := &services.Service{
			Name:       "GitHub TLS",
			Type:       "tls",
			Timeout:    15,
			TLSTarget:  null.NewNullString("github.com:443"),
			TLSMinDays: 7,
		}

		result, err := services.CheckTLS(svc, false)
		require.NoError(t, err)
		assert.True(t, result.Online)
	})

	t.Run("DefaultPort443", func(t *testing.T) {
		svc := &services.Service{
			Name:       "AWS TLS Default Port",
			Type:       "tls",
			Timeout:    15,
			TLSTarget:  null.NewNullString("aws.amazon.com"), // No port - should default to 443
			TLSMinDays: 7,
		}

		result, err := services.CheckTLS(svc, false)
		require.NoError(t, err)
		assert.True(t, result.Online)
	})

	t.Run("DomainFallback", func(t *testing.T) {
		svc := &services.Service{
			Name:       "Domain Fallback",
			Type:       "tls",
			Domain:     "microsoft.com", // Uses Domain when TLSTarget is empty
			Timeout:    15,
			TLSMinDays: 7,
		}

		result, err := services.CheckTLS(svc, false)
		require.NoError(t, err)
		assert.True(t, result.Online)
	})
}

func TestTLSServiceType_Errors(t *testing.T) {
	t.Run("InvalidHost", func(t *testing.T) {
		svc := &services.Service{
			Name:      "Invalid TLS Host",
			Type:      "tls",
			Timeout:   2,
			TLSTarget: null.NewNullString("invalid.nonexistent.domain.local:443"),
		}

		_, err := services.CheckTLS(svc, false)
		assert.Error(t, err)
	})

	t.Run("MissingTarget", func(t *testing.T) {
		svc := &services.Service{
			Name:    "Missing Target",
			Type:    "tls",
			Timeout: 5,
		}

		_, err := services.CheckTLS(svc, false)
		assert.Error(t, err)
	})

	t.Run("ConnectionRefused", func(t *testing.T) {
		svc := &services.Service{
			Name:      "Connection Refused",
			Type:      "tls",
			Timeout:   2,
			TLSTarget: null.NewNullString("localhost:59999"), // Unlikely to have anything listening
		}

		_, err := services.CheckTLS(svc, false)
		assert.Error(t, err)
	})
}

// =============================================================================
// GCS Storage Service Type Tests
// =============================================================================

// gcsEmulatorPort is the port for fake-gcs-server
const gcsEmulatorPort = "4443"

func TestGCSServiceType_FakeServer(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Skipping: Docker is not available")
	}

	// Start fake-gcs-server container
	containerName := containerPrefix + "-gcs"
	stopGCS := startFakeGCSServer(t, containerName)
	defer stopGCS()

	// Wait for emulator to be ready
	emulatorHost := getDockerHost() + ":" + gcsEmulatorPort
	waitForHTTP(t, "http://"+emulatorHost+"/storage/v1/b", gcsReadyTimeout)

	// Set emulator host for GCS client
	os.Setenv("STORAGE_EMULATOR_HOST", "http://"+emulatorHost)
	defer os.Unsetenv("STORAGE_EMULATOR_HOST")

	// Create a test bucket via the emulator API
	createTestBucket(t, emulatorHost, "test-bucket")

	t.Run("GCSBucketCheck", func(t *testing.T) {
		svc := &services.Service{
			Name:           "Test GCS Bucket",
			Type:           "storage",
			Timeout:        10,
			StorageBackend: null.NewNullString("gcs"),
			StorageBucket:  null.NewNullString("test-bucket"),
			// No credentials needed with emulator
		}

		result, err := services.CheckStorage(svc, false)
		require.NoError(t, err)
		assert.True(t, result.Online)
		assert.Greater(t, result.Latency, int64(0))
	})
}

func TestGCSServiceType_Errors(t *testing.T) {
	t.Run("MissingBackend", func(t *testing.T) {
		svc := &services.Service{
			Name:    "Missing Backend",
			Type:    "storage",
			Timeout: 5,
		}

		_, err := services.CheckStorage(svc, false)
		assert.Error(t, err)
	})

	t.Run("MissingBucket", func(t *testing.T) {
		svc := &services.Service{
			Name:           "Missing Bucket",
			Type:           "storage",
			Timeout:        5,
			StorageBackend: null.NewNullString("gcs"),
		}

		_, err := services.CheckStorage(svc, false)
		assert.Error(t, err)
	})

	t.Run("UnsupportedBackend", func(t *testing.T) {
		svc := &services.Service{
			Name:           "Unsupported Backend",
			Type:           "storage",
			Timeout:        5,
			StorageBackend: null.NewNullString("azure"),
			StorageBucket:  null.NewNullString("my-bucket"),
		}

		_, err := services.CheckStorage(svc, false)
		assert.Error(t, err)
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// isDockerAvailable checks if Docker is installed and running.
func isDockerAvailable() bool {
	return testutil.IsDockerAvailable()
}

// getDockerHost returns the host IP for Docker containers.
func getDockerHost() string {
	return testutil.GetDockerHost()
}

// startFakeGCSServer starts the fake-gcs-server container and returns a cleanup function.
func startFakeGCSServer(t *testing.T, name string) func() {
	t.Helper()

	prefix := testutil.GetDockerPrefix()

	// Pull image
	pullArgs := append(prefix, "pull", "fsouza/fake-gcs-server")
	// #nosec G204 -- Test helper pulls Docker image
	pullCmd := exec.Command(pullArgs[0], pullArgs[1:]...)
	_ = pullCmd.Run() // Ignore error if already pulled

	// Run container
	runArgs := append(prefix, "run", "-d", "--rm",
		"--name", name,
		"-p", gcsEmulatorPort+":4443",
		"fsouza/fake-gcs-server",
		"-scheme", "http",
	)
	// #nosec G204 -- Test helper runs Docker container
	runCmd := exec.Command(runArgs[0], runArgs[1:]...)
	if err := runCmd.Run(); err != nil {
		t.Fatalf("Failed to start fake-gcs-server: %v", err)
	}

	cleanup := func() {
		stopArgs := append(prefix, "stop", name)
		// #nosec G204 -- Test helper stops Docker container
		stopCmd := exec.Command(stopArgs[0], stopArgs[1:]...)
		_ = stopCmd.Run()
	}

	t.Cleanup(cleanup)
	return cleanup
}

// createTestBucket creates a bucket in the fake-gcs-server.
func createTestBucket(t *testing.T, host, bucketName string) {
	t.Helper()

	url := "http://" + host + "/storage/v1/b?project=test-project"
	body := strings.NewReader(`{"name":"` + bucketName + `"}`)

	req, err := http.NewRequest("POST", url, body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Warning: Could not create test bucket: %v", err)
		return
	}
	defer resp.Body.Close()
}

// waitForHTTP polls until the URL returns a non-error response.
func waitForHTTP(t *testing.T, url string, timeout time.Duration) {
	t.Helper()

	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("HTTP endpoint %s did not become ready within %v", url, timeout)
}

// Compile-time check that handlers.Router exists
var _ = handlers.Router
var _ = httptest.NewServer
