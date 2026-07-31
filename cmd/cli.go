package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/statping-ng/statping-ng/handlers"
	"github.com/statping-ng/statping-ng/source"
	"github.com/statping-ng/statping-ng/types/configs"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

func initCli() error {
	fmt.Println("Checking dependencies...")

	// Try to initialize master key from existing sources
	err := utils.InitMasterKey()
	if err == nil {
		// Key found and initialized
		fmt.Println("✓ Master key: configured")
		fmt.Println("\nReady to start.")
		return nil
	}

	// Key not found
	if errors.Is(err, utils.ErrMasterKeyNotFound) {
		fmt.Println("✗ Master key: NOT FOUND in STATPING_MASTER_KEY or STATPING_MASTER_KEY_FILE")

		// If --key-file specified, generate and write
		if initKeyFile != "" {
			return generateAndWriteKey(initKeyFile)
		}

		fmt.Println("\nNo master key configured. To generate one:")
		fmt.Println("  statping init --key-file /path/to/master.key")
		fmt.Println("\nOr set STATPING_MASTER_KEY environment variable directly.")
		return errors.New("master key required")
	}

	// Other error (e.g., permission issue)
	return err
}

func generateAndWriteKey(keyFile string) error {
	// Check if file exists
	if utils.FileExists(keyFile) && !initForce {
		return fmt.Errorf("%s already exists. Use --force to overwrite (DANGER: will invalidate all encrypted data)", keyFile)
	}

	// Generate new key
	fmt.Println("\nGenerating new master key...")
	keyHex, err := utils.GenerateMasterKey()
	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}
	fmt.Println("✓ Generated 256-bit key")

	// Ensure parent directory exists
	dir := filepath.Dir(keyFile)
	if !utils.FolderExists(dir) {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write key to file
	if err := os.WriteFile(keyFile, []byte(keyHex), 0400); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}
	fmt.Printf("✓ Written to %s\n", keyFile)

	// Set permissions (Unix only, Windows ignores)
	if err := os.Chmod(keyFile, 0400); err != nil {
		fmt.Printf("Warning: could not set permissions to 0400: %v\n", err)
	} else {
		fmt.Println("✓ Permissions set to 0400")
	}

	fmt.Println("\nAdd to your environment:")
	fmt.Printf("  export STATPING_MASTER_KEY_FILE=\"%s\"\n", keyFile)
	fmt.Println("\nThen run 'statping init' again to verify.")

	return nil
}

func assetsCli() error {
	if err := utils.InitLogs(); err != nil {
		return err
	}
	if err := source.Assets(); err != nil {
		return err
	}
	fmt.Println("Assets initialized. Use the web UI to upload custom CSS.")
	return nil
}

func systemctlCli(dir string, uninstall bool, port int64) error {
	location := "/etc/systemd/system/statping.service"

	if uninstall {
		fmt.Println("systemctl stop statping")
		if _, _, err := utils.Command("systemctl", "stop", "statping"); err != nil {
			log.Errorln(err)
		}
		fmt.Println("systemctl disable statping")
		if _, _, err := utils.Command("systemctl", "disable", "statping"); err != nil {
			log.Errorln(err)
		}
		fmt.Println("Deleting systemctl: ", location)
		if err := utils.DeleteFile(location); err != nil {
			log.Errorln(err)
		}
		return nil
	}
	if ok := utils.FolderExists(dir); !ok {
		return errors.New("directory does not exist: " + dir)
	}

	binPath, err := os.Executable()
	if err != nil {
		return err
	}

	config := []byte(`[Unit]
Description=Statping Server
After=network.target
After=systemd-user-sessions.service
After=network-online.target

[Service]
Type=simple
Restart=always
Environment="STATPING_DIR=` + dir + `"
Environment="ALLOW_REPORTS=false"
ExecStart=` + binPath + ` --port=` + utils.ToString(port) + `
WorkingDirectory=` + dir + `

[Install]
WantedBy=multi-user.target"
`)
	fmt.Println("Saving systemctl service to: ", location)
	fmt.Printf("Using directory %s for Statping data\n", dir)
	fmt.Printf("Running on port %d\n", port)
	fmt.Printf("\n\n%s\n\n", string(config))
	if err := utils.SaveFile(location, config); err != nil {
		return err
	}
	fmt.Println("systemctl daemon-reload")
	if _, _, err := utils.Command("systemctl", "daemon-reload"); err != nil {
		return err
	}
	fmt.Println("systemctl enable statping")
	if _, _, err := utils.Command("systemctl", "enable", "statping.service"); err != nil {
		return err
	}
	fmt.Println("systemctl start statping")
	if _, _, err := utils.Command("systemctl", "start", "statping"); err != nil {
		return err
	}
	fmt.Println("Statping was will auto start on reboots")
	fmt.Println("systemctl service: ", location)

	return nil
}

func exportCli(args []string) error {
	filename := filepath.Join(utils.Directory, time.Now().Format("01-02-2006-1504")+".json")
	if len(args) == 1 {
		filename = fmt.Sprintf("%s/%s", utils.Directory, args)
	}
	var data *handlers.ExportData
	if err := utils.InitLogs(); err != nil {
		return err
	}
	if err := source.Assets(); err != nil {
		return err
	}
	config, err := configs.LoadConfigs(configFile)
	if err != nil {
		return err
	}
	if err = configs.ConnectConfigs(config, false); err != nil {
		return err
	}
	if _, err := services.SelectAllServices(false); err != nil {
		return err
	}
	if data, err = handlers.ExportSettings(); err != nil {
		return fmt.Errorf("could not export settings: %v", err.Error())
	}
	if err = utils.SaveFile(filename, data.JSON()); err != nil {
		return fmt.Errorf("could not write file statping-export.json: %v", err.Error())
	}
	log.Infoln("Statping export file saved to ", filename)
	return nil
}

func resetCli() error {
	d := utils.Directory
	fmt.Println("Statping directory: ", d)
	assets := d + "/assets"
	if utils.FolderExists(assets) {
		fmt.Printf("Deleting %s folder.\n", assets)
		if err := utils.DeleteDirectory(assets); err != nil {
			return err
		}
	} else {
		fmt.Printf("Assets folder does not exist %s\n", assets)
	}

	logDir := d + "/logs"
	if utils.FolderExists(logDir) {
		// Close log file handles before deleting (required on Windows)
		utils.CloseLogs()
		fmt.Printf("Deleting %s directory.\n", logDir)
		if err := utils.DeleteDirectory(logDir); err != nil {
			return err
		}
	} else {
		fmt.Printf("Logs folder does not exist %s\n", logDir)
	}

	c := d + "/config.yml"
	if utils.FileExists(c) {
		fmt.Printf("Deleting %s file.\n", c)
		if err := utils.DeleteFile(c); err != nil {
			return err
		}
	} else {
		fmt.Printf("Config file does not exist %s\n", c)
	}

	dbFile := d + "/statping.db"
	if utils.FileExists(dbFile) {
		fmt.Printf("Backuping up %s file.\n", dbFile)
		if err := utils.RenameDirectory(dbFile, d+"/statping.db.backup"); err != nil {
			return err
		}
	} else {
		fmt.Printf("Statping SQL Database file does not exist %s\n", dbFile)
	}

	fmt.Println("Statping has been reset")
	return nil
}

func envCli() error {
	fmt.Println("Statping Configuration")
	fmt.Printf("Process ID:          %d\n", os.Getpid())
	fmt.Printf("Running as user id:  %d\n", os.Getuid())
	fmt.Printf("Running as group id: %d\n", os.Getgid())
	fmt.Printf("Statping Directory:  %s\n", utils.Directory)
	for k, v := range utils.Params.AllSettings() {
		fmt.Printf("%s=%v\n", strings.ToUpper(k), v)
	}
	return nil
}

func onceCli() error {
	if err := utils.InitLogs(); err != nil {
		return err
	}
	if err := source.Assets(); err != nil {
		return err
	}
	log.Infoln("Running 1 time and saving to database...")
	if err := runOnce(); err != nil {
		return err
	}
	// core.CloseDB()
	fmt.Println("Check is complete.")
	return nil
}

func importCli(args []string) error {
	var err error
	var data []byte
	if len(args) < 1 {
		return errors.New("invalid command arguments")
	}
	if data, err = os.ReadFile(args[0]); err != nil {
		return err
	}
	var exportData handlers.ExportData
	if err = json.Unmarshal(data, &exportData); err != nil {
		return err
	}
	log.Printf("=== %s ===\n", exportData.Core.Name)
	if exportData.Config != nil {
		log.Printf("Configs:     %s\n", exportData.Config.DbConn)
		if exportData.Config.DbUser != "" {
			log.Printf("   - Host:   %s\n", exportData.Config.DbHost)
			log.Printf("   - User:   %s\n", exportData.Config.DbUser)
		}
	}
	if len(exportData.Services) > 0 {
		log.Printf("Services:   %d\n", len(exportData.Services))
	}
	if len(exportData.Checkins) > 0 {
		log.Printf("Checkins:   %d\n", len(exportData.Checkins))
	}
	if len(exportData.Groups) > 0 {
		log.Printf("Groups:     %d\n", len(exportData.Groups))
	}
	if len(exportData.Messages) > 0 {
		log.Printf("Messages:   %d\n", len(exportData.Messages))
	}
	if len(exportData.Incidents) > 0 {
		log.Printf("Incidents:  %d\n", len(exportData.Incidents))
	}
	if len(exportData.Users) > 0 {
		log.Printf("Users:      %d\n", len(exportData.Users))
	}

	if exportData.Config != nil {
		if ask("Create config.yml file from Configs?") {
			log.Printf("Database Host:   	%s\n", exportData.Config.DbHost)
			log.Printf("Database Port:   	%d\n", exportData.Config.DbPort)
			log.Printf("Database User:   	%s\n", exportData.Config.DbUser)
			log.Printf("Database Password:   %s\n", exportData.Config.DbPass)
			if err := exportData.Config.Save(utils.Directory); err != nil {
				return err
			}
		}
	}

	config, err := configs.LoadConfigs(configFile)
	if err != nil {
		return err
	}
	if err = configs.ConnectConfigs(config, false); err != nil {
		return err
	}
	if ask("Create database rows and sample data?") {
		if err := config.ResetCore(); err != nil {
			return err
		}
	}
	if ask("Import Core settings?") {
		c := exportData.Core
		if err := c.Update(); err != nil {
			log.Errorln(err)
		}
	}
	for _, s := range exportData.Groups {
		if ask(fmt.Sprintf("Import Group '%s'?", s.Name)) {
			s.Id = 0
			if err := s.Create(); err != nil {
				log.Errorln(err)
			}
		}
	}
	for _, s := range exportData.Services {
		if ask(fmt.Sprintf("Import Service '%s'?", s.Name)) {
			s.Id = 0
			if err := s.Create(); err != nil {
				log.Errorln(err)
			}
		}
	}
	for _, s := range exportData.Checkins {
		if ask(fmt.Sprintf("Import Checkin '%s'?", s.Name)) {
			s.Id = 0
			if err := s.Create(); err != nil {
				log.Errorln(err)
			}
		}
	}
	for _, s := range exportData.Messages {
		if ask(fmt.Sprintf("Import Message '%s'?", s.Title)) {
			s.Id = 0
			if err := s.Create(); err != nil {
				log.Errorln(err)
			}
		}
	}
	for _, s := range exportData.Users {
		if ask(fmt.Sprintf("Import User '%s'?", s.Username)) {
			s.Id = 0
			if err := s.Create(); err != nil {
				log.Errorln(err)
			}
		}
	}
	log.Infof("Import complete")
	return nil
}

func ask(format string) bool {
	fmt.Printf("%s", format+" [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.ReplaceAll(text, "\n", "")
	return strings.ToLower(text) == "y"
}

// runOnce will initialize the Statping application and check each service 1 time, will not run HTTP server
func runOnce() error {
	config, err := configs.LoadConfigs(configFile)
	if err != nil {
		return errors.Wrap(err, "config.yml file not found")
	}
	err = configs.ConnectConfigs(config, false)
	if err != nil {
		return errors.Wrap(err, "issue connecting to database")
	}
	c, err := core.Select()
	if err != nil {
		return errors.Wrap(err, "core database was not found or setup")
	}

	core.App = c

	_, err = services.SelectAllServices(true)
	if err != nil {
		return errors.Wrap(err, "could not select all services")
	}
	for _, srv := range services.Services() {
		srv.CheckService(true)
	}
	return nil
}

// ExportChartsJs renders the charts for the index page

//type ExportData struct {
//	Config    *configs.DbConfig   `json:"config"`
//	Core      *core.Core          `json:"core"`
//	Services  []services.Service  `json:"services"`
//	Messages  []*messages.Message `json:"messages"`
//	Checkins  []*checkins.Checkin `json:"checkins"`
//	Users     []*users.User       `json:"users"`
//	Groups    []*groups.Group     `json:"groups"`
//	Notifiers []core.AllNotifiers `json:"notifiers"`
//}

// ExportSettings will export a JSON file containing all of the settings below:
// - Core
// - Notifiers
// - Checkins
// - Users
// - Services
// - Groups
// - Messages
//func ExportSettings() ([]byte, error) {
//	c, err := core.Select()
//	if err != nil {
//		return nil, err
//	}
//	var srvs []services.Service
//	for _, s := range services.AllInOrder() {
//		s.Failures = nil
//		srvs = append(srvs, s)
//	}
//
//	cfg, err := configs.LoadConfigs(configFile)
//	if err != nil {
//		return nil, err
//	}
//
//	data := ExportData{
//		Config:    cfg,
//		Core:      c,
//		Notifiers: core.App.Notifications,
//		Checkins:  checkins.All(),
//		Users:     users.All(),
//		Services:  srvs,
//		Groups:    groups.All(),
//		Messages:  messages.All(),
//	}
//	export, err := json.Marshal(data)
//	return export, err
//}

// ExportIndexHTML returns the HTML of the index page as a string
//func ExportIndexHTML() []byte {
//	source.Assets()
//	core.CoreApp.Connect(core.CoreApp., utils.Directory)
//	core.SelectAllServices(false)
//	for _, srv := range core.Services() {
//		core.CheckService(srv, true)
//	}
//	w := httptest.NewRecorder()
//	r := httptest.NewRequest("GET", "/", nil)
//	handlers.ExecuteResponse(w, r, "index.gohtml", nil, nil)
//	return w.Body.Bytes()
//}
