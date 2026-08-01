package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/statping-ng/statping-ng/utils"
)

var (
	initKeyFile string
	initForce   bool
)

var rootCmd = &cobra.Command{
	Use:     "statping",
	Version: VERSION,
	Short:   "A simple Application Status Monitor that is opensource and lightweight.",
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Example: "statping version",
	Short:   "Print the version number of Statping",
	Run: func(cmd *cobra.Command, args []string) {
		if COMMIT != "" {
			cmd.Printf("%s (%s)\n", VERSION, COMMIT)
		} else {
			cmd.Printf("%s\n", VERSION)
		}
	},
}

var initCmd = &cobra.Command{
	Use:     "init",
	Example: "statping init --key-file /etc/statping/master.key",
	Short:   "Initialize Statping and verify dependencies (master key, etc.)",
	Long: `Pre-flight check that verifies all dependencies are configured correctly.

If no master key is found and --key-file is specified, generates a new
256-bit master key and writes it to the specified file with secure permissions.

Examples:
  # Check if master key is configured
  statping init

  # Generate new master key and save to file
  statping init --key-file /etc/statping/master.key

  # Force overwrite existing key file (DANGER: invalidates encrypted data)
  statping init --key-file /etc/statping/master.key --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initCli()
	},
}

var systemctlCmd = &cobra.Command{
	Use:     "systemctl [install/uninstall]",
	Example: "statping systemctl install",
	Short:   "Install or Uninstall systemctl services",
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[1] == "install" {
			if len(args) < 3 {
				return errors.New("requires 'install <working_path> <port>'")
			}
		}
		port := utils.ToInt(args[2])
		if port == 0 {
			port = 80
		}
		if err := systemctlCli(args[1], args[0] == "uninstall", port); err != nil {
			return err
		}
		return nil
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("requires 'install <working_path>' or 'uninstall' as arguments")
		}
		return nil
	},
}

var assetsCmd = &cobra.Command{
	Use:     "assets",
	Example: "statping assets",
	Short:   "Dump all assets used locally to be edited",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := assetsCli(); err != nil {
			return err
		}
		return nil
	},
}

var exportCmd = &cobra.Command{
	Use:     "export",
	Example: "statping export",
	Short:   "Exports your Statping settings to a 'statping-export.json' file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := exportCli(args); err != nil {
			return err
		}
		return nil
	},
}

var envCmd = &cobra.Command{
	Use:     "env",
	Example: "statping env",
	Short:   "Return the configs that will be ran",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := envCli(); err != nil {
			return err
		}
		return nil
	},
}

var resetForce bool

var resetCmd = &cobra.Command{
	Use:     "reset",
	Example: "statping reset",
	Short:   "Start a fresh copy of Statping (DESTRUCTIVE)",
	Long: `Reset Statping to a fresh state by deleting assets, logs, and config.
The database is backed up to statping.db.backup.

This is a DESTRUCTIVE operation - use with caution.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := resetCli(); err != nil {
			return err
		}
		return nil
	},
}

var onceCmd = &cobra.Command{
	Use:     "once",
	Example: "statping once",
	Short:   "Check all services 1 time and then quit",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := onceCli(); err != nil {
			return err
		}
		return nil
	},
}

var importCmd = &cobra.Command{
	Use:     "import [.json file]",
	Example: "statping import backup.json",
	Short:   "Imports settings from a previously saved JSON file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := importCli(args); err != nil {
			return err
		}
		os.Exit(0)
		return nil
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires input file (.json)")
		}
		return nil
	},
}

var (
	resetAdminUsername string
	resetAdminPassword string
	resetAdminEmail    string
)

var resetAdminCmd = &cobra.Command{
	Use:   "reset-admin",
	Short: "Reset admin user password (emergency recovery)",
	Long: `Reset the password for an admin user. Requires master key to be configured.

This command connects to the database and updates the specified user's password.
Use this when you've lost access to the admin account.

Examples:
  statping reset-admin --password YourNewPassword
  statping reset-admin --user otheradmin --password YourNewPassword
  statping reset-admin --user admin --password NewPass --email new@example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := resetAdminCli(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initKeyFile, "key-file", "", "path to write generated master key")
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing key file (DANGER)")

	resetCmd.Flags().BoolVarP(&resetForce, "force", "f", false, "skip confirmation prompt (DANGER)")

	resetAdminCmd.Flags().StringVar(&resetAdminUsername, "user", "admin", "username to reset")
	resetAdminCmd.Flags().StringVar(&resetAdminPassword, "password", "", "new password (required)")
	resetAdminCmd.Flags().StringVar(&resetAdminEmail, "email", "", "update email address (optional)")
	_ = resetAdminCmd.MarkFlagRequired("password")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		exit(err)
	}
}
