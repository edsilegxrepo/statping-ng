package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build compiled Statping-ng binary once for end-to-end binary CLI testing
	tmpDir, err := os.MkdirTemp("", "statping-bin-build")
	if err != nil {
		os.Exit(1)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	binFile := filepath.Join(tmpDir, "statping-ng")
	buildCmd := exec.Command("go", "build", "-o", binFile, "../cmd")
	if err := buildCmd.Run(); err != nil {
		os.Exit(1)
	}
	binaryPath = binFile

	os.Exit(m.Run())
}

func TestBinaryCLI_Version(t *testing.T) {
	require.FileExists(t, binaryPath)

	cmd := exec.Command(binaryPath, "version")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "stderr: %s", stderr.String())
	assert.Contains(t, stderr.String()+stdout.String(), "dev")
}

func TestBinaryCLI_VersionFlag(t *testing.T) {
	require.FileExists(t, binaryPath)

	cmd := exec.Command(binaryPath, "--version")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "stderr: %s", stderr.String())
	assert.Contains(t, stderr.String()+stdout.String(), "dev")
}

func TestBinaryCLI_Help(t *testing.T) {
	require.FileExists(t, binaryPath)

	cmd := exec.Command(binaryPath, "help")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "stderr: %s", stderr.String())
	assert.Contains(t, stdout.String(), "Usage:")
	assert.Contains(t, stdout.String(), "Flags:")
}

func TestBinaryCLI_Env(t *testing.T) {
	require.FileExists(t, binaryPath)

	cmd := exec.Command(binaryPath, "env")
	cmd.Env = append(os.Environ(), "STATPING_DIR=/tmp/statping-test-env")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "stderr: %s", stderr.String())
	assert.Contains(t, stdout.String(), "STATPING_DIR")
}

func TestBinaryCLI_PortFlag(t *testing.T) {
	require.FileExists(t, binaryPath)

	cmd := exec.Command(binaryPath, "env", "--port", "9090")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "stderr: %s", stderr.String())
	assert.Contains(t, stdout.String(), "9090")
}

func TestBinaryCLI_IPFlag(t *testing.T) {
	require.FileExists(t, binaryPath)

	cmd := exec.Command(binaryPath, "env", "--ip", "127.0.0.1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "stderr: %s", stderr.String())
	assert.Contains(t, stdout.String(), "127.0.0.1")
}

func TestBinaryCLI_VerboseFlag(t *testing.T) {
	require.FileExists(t, binaryPath)

	cmd := exec.Command(binaryPath, "env", "-v", "4")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "stderr: %s", stderr.String())
	assert.Contains(t, stdout.String(), "4")
}

func TestBinaryCLI_ConfigFlag(t *testing.T) {
	require.FileExists(t, binaryPath)

	cmd := exec.Command(binaryPath, "env", "--config", "/tmp/custom-statping-config.yml")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "stderr: %s", stderr.String())
	assert.Contains(t, stdout.String(), "/tmp/custom-statping-config.yml")
}

func TestBinaryCLI_GlobalFlags(t *testing.T) {
	require.FileExists(t, binaryPath)

	cmd := exec.Command(binaryPath, "--help")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "stderr: %s", stderr.String())
	assert.Contains(t, stdout.String(), "Application Status Monitor")
}

func TestBinaryCLI_AssetsExport(t *testing.T) {
	require.FileExists(t, binaryPath)

	tmpDir, err := os.MkdirTemp("", "statping-assets-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cmd := exec.Command(binaryPath, "assets")
	cmd.Env = append(os.Environ(), "STATPING_DIR="+tmpDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	require.NoError(t, err, "stderr: %s", stderr.String())
	assert.DirExists(t, filepath.Join(tmpDir, "assets"))
}
