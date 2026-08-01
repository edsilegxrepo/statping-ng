package utils

import (
	"io"
	"os"
	"time"

	"github.com/spf13/viper"
)

var Params *viper.Viper

func InitEnvs() {
	if Params != nil {
		return
	}
	Params = viper.New()
	Params.AutomaticEnv()

	var err error
	defaultDir, err := os.Getwd()
	if err != nil {
		Log.Errorln(err)
		defaultDir = "."
	}
	Params.SetDefault("DISABLE_HTTP", false)
	Params.SetDefault("STATPING_DIR", defaultDir)
	Params.SetDefault("GO_ENV", "production")
	Params.SetDefault("DEBUG", false)
	Params.SetDefault("ADMIN_LOCK", false)
	Params.SetDefault("DB_CONN", "")
	Params.SetDefault("DB_DSN", "")
	Params.SetDefault("DISABLE_LOGS", false)
	Params.SetDefault("USE_ASSETS", false)
	Params.SetDefault("BASE_PATH", "")
	Params.SetDefault("ADMIN_USER", "admin")
	Params.SetDefault("ADMIN_PASSWORD", "")
	Params.SetDefault("ADMIN_EMAIL", "info@admin.com")
	Params.SetDefault("MAX_OPEN_CONN", 25)
	Params.SetDefault("MAX_IDLE_CONN", 25)
	Params.SetDefault("MAX_LIFE_CONN", 5*time.Minute)
	Params.SetDefault("SAMPLE_DATA", false)
	Params.SetDefault("ALLOW_REPORTS", true)
	Params.SetDefault("POSTGRES_SSLMODE", "disable")
	Params.SetDefault("NAME", "Platform Monitoring")
	Params.SetDefault("DOMAIN", "http://localhost:8080")
	Params.SetDefault("DESCRIPTION", "Enterprise Status Page")
	Params.SetDefault("REMOVE_AFTER", 365*24*time.Hour) // Default 12 Months (for 1 year uptime calculation)
	Params.SetDefault("CLEANUP_INTERVAL", 1*time.Hour)
	Params.SetDefault("LANGUAGE", "en")
	Params.SetDefault("READ_ONLY", false)
	Params.SetDefault("LOGS_MAX_COUNT", 5)
	Params.SetDefault("LOGS_MAX_AGE", 28)
	Params.SetDefault("LOGS_MAX_SIZE", 16)
	Params.SetDefault("DISABLE_COLORS", false)
	// Log shipping: send logs to external systems (Loki, Elasticsearch, Splunk, Cribl, webhook)
	Params.SetDefault("LOG_SHIP_TYPE", "")       // loki, elasticsearch, splunk, cribl, webhook
	Params.SetDefault("LOG_SHIP_ENDPOINT", "")   // URL to send logs to
	Params.SetDefault("LOG_SHIP_TOKEN", "")      // Bearer token (or Splunk HEC token)
	Params.SetDefault("LOG_SHIP_LABELS", "")     // Additional labels (key=value,key2=value2)
	Params.SetDefault("LOG_SHIP_INDEX", "main")  // Splunk index
	Params.SetDefault("LOG_SHIP_SOURCETYPE", "") // Splunk/Cribl sourcetype (default: statping)

	dbConn := Params.GetString("DB_CONN")
	dbInt := Params.GetInt("DB_PORT")
	if dbInt == 0 && dbConn != "sqlite" && dbConn != "sqlite3" {
		if dbConn == "postgres" {
			Params.SetDefault("DB_PORT", 5432)
		}
		if dbConn == "mysql" {
			Params.SetDefault("DB_PORT", 3306)
		}
	}

	Directory = Params.GetString("STATPING_DIR")
	// Params.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	Params.SetConfigName("config")
	Params.SetConfigType("yml")
	Params.AddConfigPath(Directory)
	if err := Params.ReadInConfig(); err != nil {
		// Config file not found is acceptable (will use defaults/env vars)
		// but log other errors
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			Log.Warnln("Error reading config.yml:", err)
		}
	}

	Params.AddConfigPath(Directory)
	Params.SetConfigFile(".env")
	if err := Params.ReadInConfig(); err != nil {
		// .env file not found is acceptable - check both viper error and OS error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			if !os.IsNotExist(err) {
				Log.Warnln("Error reading .env file:", err)
			}
		}
	}

	// check if logs are disabled
	if Params.GetBool("DISABLE_LOGS") {
		Log.Out = io.Discard
		return
	}
	Log.Debugln("current working directory: ", Directory)
	Log.AddHook(new(hook))
	Log.SetNoLock()
	checkVerboseMode()
}
