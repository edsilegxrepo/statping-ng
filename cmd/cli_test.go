package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/statping-ng/statping-ng/source"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dir string

func init() {
	_ = utils.InitLogs()
}

func TestStatpingDirectory(t *testing.T) {
	dir = utils.Params.GetString("STATPING_DIR")
	require.NotEmpty(t, dir)
}

func TestEnvCLI(t *testing.T) {
	_ = os.Setenv("API_SECRET", "demoapisecret123")
	_ = os.Setenv("SASS", "/usr/local/bin/sass")

	cmd := rootCmd
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"env"})
	err := cmd.Execute()
	require.Nil(t, err)

	_ = os.Unsetenv("API_SECRET")
	_ = os.Unsetenv("SASS")
}

func TestVersionCLI(t *testing.T) {
	cmd := rootCmd
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"version"})
	err := cmd.Execute()
	require.Nil(t, err)
	out, err := io.ReadAll(b)
	require.Nil(t, err)
	assert.Contains(t, strings.TrimSpace(string(out)), VERSION)
}

func TestAssetsCLI(t *testing.T) {
	cmd := rootCmd
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"assets"})
	err := cmd.Execute()
	// Skip test if sass executable is not available (build-time dependency)
	if err != nil && strings.Contains(err.Error(), "sass") {
		t.Skip("sass executable not found in PATH - skipping assets CLI test")
	}
	require.Nil(t, err)
	for _, f := range source.RequiredFiles {
		assert.FileExists(t, utils.Directory+"/assets/"+f)
	}
}

func TestUpdateCLI(t *testing.T) {
	t.Skip("Skipping network-dependent update CLI test")
}

func TestHelpCLI(t *testing.T) {
	cmd := rootCmd
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"help"})
	err := cmd.Execute()
	require.Nil(t, err)
	out, err := io.ReadAll(b)
	require.Nil(t, err)
	assert.Contains(t, string(out), "Usage:")
}

func TestResetCLI(t *testing.T) {
	err := utils.SaveFile(utils.Directory+"/statping.db", []byte("test data"))
	require.Nil(t, err)

	cmd := rootCmd
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"reset"})
	err = cmd.Execute()
	require.Nil(t, err)

	_ = utils.DeleteDirectory(utils.Directory + "/assets")
	_ = utils.DeleteDirectory(utils.Directory + "/logs")

	assert.NoFileExists(t, utils.Directory+"/config.yml")
	assert.NoFileExists(t, utils.Directory+"/statping.db")
	assert.FileExists(t, utils.Directory+"/statping.db.backup")

	err = utils.DeleteFile(utils.Directory + "/statping.db.backup")
	require.Nil(t, err)
}
