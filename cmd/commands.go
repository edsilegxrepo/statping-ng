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

var resetCmd = &cobra.Command{
	Use:     "reset",
	Example: "statping reset",
	Short:   "Start a fresh copy of Statping",
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

func init() {
	initCmd.Flags().StringVar(&initKeyFile, "key-file", "", "path to write generated master key")
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing key file (DANGER)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		exit(err)
	}
}
