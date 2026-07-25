package configs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/statping-ng/statping-ng/utils"
)

var log = utils.Log.WithField("type", "configs")

// ConnectConfigs will connect to the database and save the config.yml file
func ConnectConfigs(configs *DbConfig, retry bool) error {
	err := Connect(configs, retry)
	if err != nil {
		return errors.Wrap(err, "error connecting to database")
	}
	if err := configs.Save(utils.Directory); err != nil {
		return errors.Wrap(err, "error saving configuration")
	}
	return nil
}

// findDbFile will attempt to find the "statping.db" database file in the current
// working directory, or from STATPING_DIR env.
func findDbFile(configs *DbConfig) (string, error) {
	// Use Location from config if set, otherwise fall back to utils.Directory
	baseDir := utils.Directory
	if configs != nil && configs.Location != "" {
		baseDir = configs.Location
	}

	// If SqlFile is explicitly set, use it directly
	if configs != nil && configs.SqlFile != "" {
		return configs.SqlFile, nil
	}

	// Build the default location
	dbFilename := SqliteFilename
	if configs != nil && configs.DbData != "" {
		dbFilename = configs.DbData
	}
	location := baseDir + "/" + dbFilename

	// If no config provided, try to find existing db file
	if configs == nil {
		file, err := findSQLin(baseDir)
		if err != nil {
			log.Errorln(err)
			return location, nil
		}
		return file, nil
	}

	return location, nil
}

// findSQLin walks the current walking directory for statping.db
func findSQLin(path string) (string, error) {
	filename := SqliteFilename
	var found []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".db" {
			filename = info.Name()
			found = append(found, filename)
		}
		return nil
	})
	if err != nil {
		return filename, err
	}
	if len(found) > 1 {
		return filename, errors.Errorf("found multiple database files: %s", strings.Join(found, ", "))
	}
	return filename, nil
}
