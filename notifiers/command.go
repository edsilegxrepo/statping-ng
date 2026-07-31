package notifiers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/notifications"
	"github.com/statping-ng/statping-ng/types/notifier"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

var _ notifier.Notifier = (*commandLine)(nil)

type commandLine struct {
	*notifications.Notification
}

func (c *commandLine) Select() *notifications.Notification {
	return c.Notification
}

func (c *commandLine) Valid(values notifications.Values) error {
	// Var1 is used for test command - validate if provided
	if values.Var1 != "" {
		if err := validateCommand(values.Var1); err != nil {
			return fmt.Errorf("invalid test command: %w", err)
		}
	}
	return nil
}

// getTrustedScriptDir returns the trusted script directory path
// Scripts in this directory are allowed to execute
func getTrustedScriptDir() string {
	dir := utils.Params.GetString("TRUSTED_SCRIPT_DIR")
	if dir == "" {
		// Default to $STATPING_DIR/scripts
		dir = filepath.Join(utils.Directory, "scripts")
	}
	return dir
}

// isInTrustedDir checks if a path is inside the trusted scripts directory
// Resolves symlinks to prevent bypass attacks
func isInTrustedDir(cmdPath string) bool {
	trustedDir := getTrustedScriptDir()

	// Check if trusted dir exists
	if _, err := os.Stat(trustedDir); os.IsNotExist(err) {
		return false
	}

	// Resolve both paths to absolute
	absCmd, err := filepath.Abs(cmdPath)
	if err != nil {
		return false
	}

	absTrusted, err := filepath.Abs(trustedDir)
	if err != nil {
		return false
	}

	// Resolve symlinks to prevent bypass via symlink inside trusted dir pointing outside
	realCmd, err := filepath.EvalSymlinks(absCmd)
	if err != nil {
		// File doesn't exist yet or can't resolve - reject
		return false
	}

	realTrusted, err := filepath.EvalSymlinks(absTrusted)
	if err != nil {
		return false
	}

	// Clean paths and check prefix
	realCmd = filepath.Clean(realCmd)
	realTrusted = filepath.Clean(realTrusted)

	// Must be inside trusted dir (not just prefixed)
	rel, err := filepath.Rel(realTrusted, realCmd)
	if err != nil {
		return false
	}

	// If relative path starts with "..", it's outside trusted dir
	return !strings.HasPrefix(rel, "..")
}

// validateCommand checks if a command is safe to execute
// Only scripts in TRUSTED_SCRIPT_DIR are allowed
func validateCommand(cmd string) error {
	if cmd == "" {
		return nil
	}

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmdPath := parts[0]

	// Check if command is in trusted scripts directory
	if isInTrustedDir(cmdPath) {
		// Verify the script file exists and is a regular file
		info, err := os.Stat(cmdPath)
		if err != nil {
			return fmt.Errorf("script not found: %s", cmdPath)
		}
		if info.IsDir() {
			return fmt.Errorf("script path is a directory: %s", cmdPath)
		}
		return nil
	}

	return fmt.Errorf("command %q not in trusted scripts directory (%s)", cmdPath, getTrustedScriptDir())
}

var Command = &commandLine{&notifications.Notification{
	Method:      "command",
	Title:       "Command",
	Description: "Execute scripts from the trusted scripts directory (TRUSTED_SCRIPT_DIR or $STATPING_DIR/scripts).",
	Author:      "Hunter Long",
	AuthorUrl:   "https://github.com/hunterlong",
	Delay:       time.Duration(1 * time.Second),
	Icon:        "fas fa-terminal",
	SuccessData: null.NewNullString("scripts/notify-success.sh"),
	FailureData: null.NewNullString("scripts/notify-failure.sh"),
	DataType:    "text",
	Limits:      60,
}}

func runCommand(cmd string, env []string) (string, string, error) {
	// Validate command before execution
	if err := validateCommand(cmd); err != nil {
		return "", "", err
	}

	utils.Log.Infof("Command notifier executing: %s", cmd)
	cmdApp := strings.Fields(cmd)
	if len(cmdApp) == 0 {
		return "", "", fmt.Errorf("you need at least 1 command")
	}
	var cmdArgs []string
	if len(cmdApp) > 1 {
		cmdArgs = cmdApp[1:]
	}
	outStr, errStr, err := utils.CommandWithEnv(cmdApp[0], env, cmdArgs...)
	return outStr, errStr, err
}

// buildServiceEnv creates environment variables from service and failure data
// This prevents argument injection by passing data as env vars instead of string substitution
func buildServiceEnv(s *services.Service, f failures.Failure) []string {
	env := []string{
		fmt.Sprintf("STATPING_SERVICE_NAME=%s", s.Name),
		fmt.Sprintf("STATPING_SERVICE_ID=%d", s.Id),
		fmt.Sprintf("STATPING_SERVICE_DOMAIN=%s", s.Domain),
		fmt.Sprintf("STATPING_SERVICE_TYPE=%s", s.Type),
		fmt.Sprintf("STATPING_SERVICE_STATUS_CODE=%d", s.LastStatusCode),
		fmt.Sprintf("STATPING_SERVICE_LATENCY=%d", s.Latency),
		fmt.Sprintf("STATPING_SERVICE_ONLINE=%t", s.Online),
	}
	if f.Issue != "" {
		env = append(env, fmt.Sprintf("STATPING_FAILURE_ISSUE=%s", f.Issue))
		env = append(env, fmt.Sprintf("STATPING_FAILURE_ID=%d", f.Id))
	}
	return env
}

// OnSuccess for commandLine will trigger successful service
func (c *commandLine) OnSuccess(s *services.Service) (string, error) {
	// Pass service data as environment variables to prevent argument injection
	env := buildServiceEnv(s, failures.Failure{})
	out, _, err := runCommand(c.SuccessData.String, env)
	return out, err
}

// OnFailure for commandLine will trigger failing service
func (c *commandLine) OnFailure(s *services.Service, f failures.Failure) (string, error) {
	// Pass service/failure data as environment variables to prevent argument injection
	env := buildServiceEnv(s, f)
	out, _, err := runCommand(c.FailureData.String, env)
	return out, err
}

// OnTest for commandLine triggers when this notifier has been saved
func (c *commandLine) OnTest() (string, error) {
	example := services.Example(true)
	env := buildServiceEnv(example, failures.Example())
	in, out, err := runCommand(c.Var1.String, env)
	utils.Log.Infoln(in)
	utils.Log.Infoln(out)
	return out, err
}

// OnSave will trigger when this notifier is saved
func (c *commandLine) OnSave() (string, error) {
	return "", nil
}
