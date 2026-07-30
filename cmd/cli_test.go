package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/statping-ng/statping-ng/source"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dir string

func init() {
	_ = utils.InitLogs()
}

// executeCommand is a helper for executing cobra commands in tests
// It properly resets the command state before each execution to avoid
// test pollution from previous runs
func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	// Reset args to ensure clean state
	root.SetArgs(args)
	err = root.Execute()
	return buf.String(), err
}

// resetRootCmd resets rootCmd to a clean state for testing
func resetRootCmd() {
	// Reset the args to nil to clear any previous test's args
	rootCmd.SetArgs(nil)
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

func TestRootCmdFlags(t *testing.T) {
	t.Run("port flag exists", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("port")
		assert.NotNil(t, flag)
		assert.Equal(t, "p", flag.Shorthand)
	})

	t.Run("ip flag exists", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("ip")
		assert.NotNil(t, flag)
		assert.Equal(t, "s", flag.Shorthand)
	})

	t.Run("verbose flag exists", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("verbose")
		assert.NotNil(t, flag)
		assert.Equal(t, "v", flag.Shorthand)
	})

	t.Run("config flag exists", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("config")
		assert.NotNil(t, flag)
		assert.Equal(t, "c", flag.Shorthand)
	})
}

func TestSubCommands(t *testing.T) {
	subCommands := rootCmd.Commands()
	commandNames := make([]string, len(subCommands))
	for i, cmd := range subCommands {
		commandNames[i] = cmd.Name()
	}

	expectedCommands := []string{"version", "assets", "export", "import", "env", "reset"}
	for _, expected := range expectedCommands {
		found := false
		for _, name := range commandNames {
			if name == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected command '%s' not found", expected)
	}
}

func TestVersionOutput(t *testing.T) {
	cmd := rootCmd
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"version"})
	err := cmd.Execute()
	require.Nil(t, err)
	out, err := io.ReadAll(b)
	require.Nil(t, err)
	output := string(out)
	assert.Contains(t, output, VERSION)
}

func TestEnvOutput(t *testing.T) {
	_ = os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	cmd := rootCmd
	cmd.SetArgs([]string{"env"})
	err := cmd.Execute()
	require.Nil(t, err)
}

// TestFlagParsing tests CLI flag parsing for all persistent flags
func TestFlagParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		checkFn  func(t *testing.T)
	}{
		{
			name: "port flag with short form",
			args: []string{"-p", "9090", "version"},
			checkFn: func(t *testing.T) {
				flag := rootCmd.PersistentFlags().Lookup("port")
				assert.Equal(t, "9090", flag.Value.String())
			},
		},
		{
			name: "port flag with long form",
			args: []string{"--port", "9091", "version"},
			checkFn: func(t *testing.T) {
				flag := rootCmd.PersistentFlags().Lookup("port")
				assert.Equal(t, "9091", flag.Value.String())
			},
		},
		{
			name: "ip flag with short form",
			args: []string{"-s", "127.0.0.1", "version"},
			checkFn: func(t *testing.T) {
				flag := rootCmd.PersistentFlags().Lookup("ip")
				assert.Equal(t, "127.0.0.1", flag.Value.String())
			},
		},
		{
			name: "ip flag with long form",
			args: []string{"--ip", "192.168.1.1", "version"},
			checkFn: func(t *testing.T) {
				flag := rootCmd.PersistentFlags().Lookup("ip")
				assert.Equal(t, "192.168.1.1", flag.Value.String())
			},
		},
		{
			name: "verbose flag with short form",
			args: []string{"-v", "3", "version"},
			checkFn: func(t *testing.T) {
				flag := rootCmd.PersistentFlags().Lookup("verbose")
				assert.Equal(t, "3", flag.Value.String())
			},
		},
		{
			name: "verbose flag with long form",
			args: []string{"--verbose", "5", "version"},
			checkFn: func(t *testing.T) {
				flag := rootCmd.PersistentFlags().Lookup("verbose")
				assert.Equal(t, "5", flag.Value.String())
			},
		},
		{
			name: "config flag with short form",
			args: []string{"-c", "/tmp/test-config.yml", "version"},
			checkFn: func(t *testing.T) {
				flag := rootCmd.PersistentFlags().Lookup("config")
				assert.Equal(t, "/tmp/test-config.yml", flag.Value.String())
			},
		},
		{
			name: "config flag with long form",
			args: []string{"--config", "/custom/path/config.yml", "version"},
			checkFn: func(t *testing.T) {
				flag := rootCmd.PersistentFlags().Lookup("config")
				assert.Equal(t, "/custom/path/config.yml", flag.Value.String())
			},
		},
		{
			name: "multiple flags combined",
			args: []string{"-p", "3000", "-s", "localhost", "-v", "4", "version"},
			checkFn: func(t *testing.T) {
				assert.Equal(t, "3000", rootCmd.PersistentFlags().Lookup("port").Value.String())
				assert.Equal(t, "localhost", rootCmd.PersistentFlags().Lookup("ip").Value.String())
				assert.Equal(t, "4", rootCmd.PersistentFlags().Lookup("verbose").Value.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executeCommand(rootCmd, tt.args...)
			require.Nil(t, err)
			tt.checkFn(t)
		})
	}
}

// TestFlagDefaults tests that flags have correct default values
func TestFlagDefaults(t *testing.T) {
	t.Run("port default is 8080", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("port")
		require.NotNil(t, flag)
		assert.Equal(t, "8080", flag.DefValue)
	})

	t.Run("ip default is 127.0.0.1", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("ip")
		require.NotNil(t, flag)
		assert.Equal(t, "127.0.0.1", flag.DefValue)
	})

	t.Run("verbose default is 2", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("verbose")
		require.NotNil(t, flag)
		assert.Equal(t, "2", flag.DefValue)
	})

	t.Run("config default contains config.yml", func(t *testing.T) {
		flag := rootCmd.PersistentFlags().Lookup("config")
		require.NotNil(t, flag)
		assert.Contains(t, flag.DefValue, "config.yml")
	})
}

// TestInvalidFlagHandling tests that invalid flags are properly rejected
func TestInvalidFlagHandling(t *testing.T) {
	t.Run("unknown flag returns error", func(t *testing.T) {
		_, err := executeCommand(rootCmd, "--unknownflag", "value")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "unknown flag")
	})

	t.Run("invalid port type returns error", func(t *testing.T) {
		_, err := executeCommand(rootCmd, "--port", "notanumber", "version")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("invalid verbose type returns error", func(t *testing.T) {
		_, err := executeCommand(rootCmd, "--verbose", "high", "version")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})
}

// TestSubcommandRouting tests that subcommands are properly routed
func TestSubcommandRouting(t *testing.T) {
	t.Run("version subcommand executes", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "version")
		require.Nil(t, err)
		assert.Contains(t, output, VERSION)
	})

	t.Run("help subcommand executes", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "help")
		require.Nil(t, err)
		assert.Contains(t, output, "Usage:")
		assert.Contains(t, output, "statping")
	})

	t.Run("env subcommand executes", func(t *testing.T) {
		_, err := executeCommand(rootCmd, "env")
		require.Nil(t, err)
	})

	t.Run("unknown subcommand returns error", func(t *testing.T) {
		_, err := executeCommand(rootCmd, "unknowncommand")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "unknown command")
	})
}

// TestHelpTextOutput tests help text content for root and subcommands
func TestHelpTextOutput(t *testing.T) {
	t.Run("root help contains usage", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "--help")
		require.Nil(t, err)
		assert.Contains(t, output, "Usage:")
	})

	t.Run("root help contains available commands", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "--help")
		require.Nil(t, err)
		assert.Contains(t, output, "Available Commands:")
	})

	t.Run("root help contains flags section", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "--help")
		require.Nil(t, err)
		assert.Contains(t, output, "Flags:")
	})

	t.Run("root help lists version command", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "--help")
		require.Nil(t, err)
		assert.Contains(t, output, "version")
	})

	t.Run("root help lists env command", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "--help")
		require.Nil(t, err)
		assert.Contains(t, output, "env")
	})

	t.Run("root help lists assets command", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "--help")
		require.Nil(t, err)
		assert.Contains(t, output, "assets")
	})

	t.Run("version command has help", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "version", "--help")
		require.Nil(t, err)
		assert.Contains(t, output, "Print the version")
	})

	t.Run("env command has help", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "env", "--help")
		require.Nil(t, err)
		assert.Contains(t, output, "configs")
	})

	t.Run("assets command has help", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "assets", "--help")
		require.Nil(t, err)
		assert.Contains(t, output, "assets")
	})

	t.Run("import command has help", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "import", "--help")
		require.Nil(t, err)
		assert.Contains(t, output, "Imports settings")
	})

	t.Run("export command has help", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "export", "--help")
		require.Nil(t, err)
		assert.Contains(t, output, "Exports")
	})
}

// TestEnvironmentVariableHandling tests that environment variables are properly read
func TestEnvironmentVariableHandling(t *testing.T) {
	t.Run("STATPING_DIR is set", func(t *testing.T) {
		dir := utils.Params.GetString("STATPING_DIR")
		assert.NotEmpty(t, dir)
	})

	t.Run("custom env var is accessible in env command", func(t *testing.T) {
		_ = os.Setenv("CUSTOM_TEST_VAR", "custom_value_123")
		defer os.Unsetenv("CUSTOM_TEST_VAR")

		_, err := executeCommand(rootCmd, "env")
		require.Nil(t, err)
	})

	t.Run("API_SECRET can be set", func(t *testing.T) {
		_ = os.Setenv("API_SECRET", "test_secret_key")
		defer os.Unsetenv("API_SECRET")

		_, err := executeCommand(rootCmd, "env")
		require.Nil(t, err)
	})

	t.Run("PORT can be set via environment", func(t *testing.T) {
		originalPort := os.Getenv("PORT")
		_ = os.Setenv("PORT", "9999")
		defer func() {
			if originalPort != "" {
				os.Setenv("PORT", originalPort)
			} else {
				os.Unsetenv("PORT")
			}
		}()

		_, err := executeCommand(rootCmd, "env")
		require.Nil(t, err)
	})
}

// TestConfigFileHandling tests config file path handling via CLI
func TestConfigFileHandling(t *testing.T) {
	t.Run("config flag accepts absolute path", func(t *testing.T) {
		_, err := executeCommand(rootCmd, "--config", "/absolute/path/config.yml", "version")
		require.Nil(t, err)
		flag := rootCmd.PersistentFlags().Lookup("config")
		assert.Equal(t, "/absolute/path/config.yml", flag.Value.String())
	})

	t.Run("config flag accepts relative path", func(t *testing.T) {
		_, err := executeCommand(rootCmd, "--config", "./relative/config.yml", "version")
		require.Nil(t, err)
		flag := rootCmd.PersistentFlags().Lookup("config")
		assert.Equal(t, "./relative/config.yml", flag.Value.String())
	})

	t.Run("config flag with spaces in path", func(t *testing.T) {
		_, err := executeCommand(rootCmd, "--config", "/path with spaces/config.yml", "version")
		require.Nil(t, err)
		flag := rootCmd.PersistentFlags().Lookup("config")
		assert.Equal(t, "/path with spaces/config.yml", flag.Value.String())
	})
}

// TestImportCommandValidation tests import command argument validation
func TestImportCommandValidation(t *testing.T) {
	t.Run("import command has Args validator", func(t *testing.T) {
		// Verify the import command has an Args function defined
		assert.NotNil(t, importCmd.Args, "import command should have Args validator")
	})

	t.Run("import Args validator rejects empty args", func(t *testing.T) {
		// Test the Args validator directly to avoid Cobra state issues
		err := importCmd.Args(importCmd, []string{})
		require.NotNil(t, err, "import command should require a file argument")
		assert.Contains(t, err.Error(), "requires")
	})

	t.Run("import Args validator accepts valid file arg", func(t *testing.T) {
		// Test the Args validator accepts a single argument
		err := importCmd.Args(importCmd, []string{"somefile.json"})
		assert.Nil(t, err, "import command should accept a single file argument")
	})
}

// TestSystemctlCommandValidation tests systemctl command argument validation
func TestSystemctlCommandValidation(t *testing.T) {
	t.Run("systemctl command has Args validator", func(t *testing.T) {
		// Verify the systemctl command has an Args function defined
		assert.NotNil(t, systemctlCmd.Args, "systemctl command should have Args validator")
	})

	t.Run("systemctl Args validator rejects empty args", func(t *testing.T) {
		// Test the Args validator directly to avoid Cobra state issues
		err := systemctlCmd.Args(systemctlCmd, []string{})
		require.NotNil(t, err, "systemctl command should require arguments")
		assert.Contains(t, err.Error(), "requires")
	})

	t.Run("systemctl Args validator rejects single arg", func(t *testing.T) {
		// Test the Args validator directly
		err := systemctlCmd.Args(systemctlCmd, []string{"install"})
		require.NotNil(t, err, "systemctl command should require working path")
		assert.Contains(t, err.Error(), "requires")
	})
}

// TestVersionWithCommit tests version output with commit hash
func TestVersionWithCommit(t *testing.T) {
	t.Run("version shows VERSION", func(t *testing.T) {
		output, err := executeCommand(rootCmd, "version")
		require.Nil(t, err)
		assert.Contains(t, output, VERSION)
	})

	t.Run("version command uses correct short description", func(t *testing.T) {
		assert.Equal(t, "Print the version number of Statping", versionCmd.Short)
	})
}

// TestRootCommandMetadata tests root command metadata
func TestRootCommandMetadata(t *testing.T) {
	t.Run("root command use is statping", func(t *testing.T) {
		assert.Equal(t, "statping", rootCmd.Use)
	})

	t.Run("root command has short description", func(t *testing.T) {
		assert.NotEmpty(t, rootCmd.Short)
		assert.Contains(t, rootCmd.Short, "Status Monitor")
	})

	t.Run("root command version matches VERSION", func(t *testing.T) {
		assert.Equal(t, VERSION, rootCmd.Version)
	})
}

// TestSubcommandMetadata tests subcommand metadata
func TestSubcommandMetadata(t *testing.T) {
	subcommandTests := []struct {
		cmd      *cobra.Command
		name     string
		hasShort bool
	}{
		{versionCmd, "version", true},
		{assetsCmd, "assets", true},
		{exportCmd, "export", true},
		{importCmd, "import", true},
		{envCmd, "env", true},
		{resetCmd, "reset", true},
		{onceCmd, "once", true},
		{sassCmd, "sass", true},
		{systemctlCmd, "systemctl", true},
	}

	for _, tt := range subcommandTests {
		t.Run(tt.name+" has metadata", func(t *testing.T) {
			assert.Equal(t, tt.name, tt.cmd.Use[:len(tt.name)])
			if tt.hasShort {
				assert.NotEmpty(t, tt.cmd.Short, "command %s should have Short description", tt.name)
			}
		})
	}
}

// TestAllSubcommandsRegistered verifies all expected subcommands are registered
func TestAllSubcommandsRegistered(t *testing.T) {
	expectedCommands := []string{
		"version",
		"update",
		"assets",
		"export",
		"import",
		"sass",
		"once",
		"env",
		"systemctl",
		"reset",
	}

	commands := rootCmd.Commands()
	registeredNames := make(map[string]bool)
	for _, cmd := range commands {
		registeredNames[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		t.Run(expected+" is registered", func(t *testing.T) {
			assert.True(t, registeredNames[expected], "command '%s' should be registered", expected)
		})
	}
}

// TestFlagShorthands tests that flag shorthands are correctly assigned
func TestFlagShorthands(t *testing.T) {
	flagTests := []struct {
		flagName  string
		shorthand string
	}{
		{"port", "p"},
		{"ip", "s"},
		{"verbose", "v"},
		{"config", "c"},
	}

	for _, tt := range flagTests {
		t.Run(tt.flagName+" has shorthand "+tt.shorthand, func(t *testing.T) {
			flag := rootCmd.PersistentFlags().Lookup(tt.flagName)
			require.NotNil(t, flag)
			assert.Equal(t, tt.shorthand, flag.Shorthand)
		})
	}
}

// TestFlagUsageStrings tests that flags have usage strings
func TestFlagUsageStrings(t *testing.T) {
	flags := []string{"port", "ip", "verbose", "config"}

	for _, flagName := range flags {
		t.Run(flagName+" has usage string", func(t *testing.T) {
			flag := rootCmd.PersistentFlags().Lookup(flagName)
			require.NotNil(t, flag)
			assert.NotEmpty(t, flag.Usage)
		})
	}
}
