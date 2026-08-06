package core

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/metrics"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/utils"
	"gorm.io/gorm"
)

var (
	db   database.Database
	dbMu sync.RWMutex
)

func getDB() database.Database {
	dbMu.RLock()
	defer dbMu.RUnlock()
	return db
}

func SetDB(database database.Database) {
	dbMu.Lock()
	db = database
	dbMu.Unlock()
	c, err := Select()
	if err != nil {
		utils.Log.Errorln(err)
		return
	}
	apiEnv := utils.Params.GetString("API_SECRET")
	if c.ApiSecret != apiEnv && apiEnv != "" {
		c.ApiSecret = apiEnv
		if err := c.Update(); err != nil {
			utils.Log.Errorln(err)
		}
	}
}

func (c *Core) AfterFind(tx *gorm.DB) (err error) {
	metrics.Query("core", "find")
	// Decrypt OAuth and LDAP secrets after loading
	c.decryptSecrets()
	return nil
}

func (c *Core) BeforeSave(tx *gorm.DB) (err error) {
	// Encrypt OAuth and LDAP secrets before saving
	if err := c.encryptSecrets(); err != nil {
		return err
	}
	return nil
}

func (c *Core) encryptSecrets() error {
	// Master key must be initialized before encrypting secrets
	if !utils.MasterKeyInitialized() {
		// Check if there are secrets that need encryption
		hasSecrets := c.GithubClientSecret != "" && !utils.IsEncrypted(c.GithubClientSecret) ||
			c.GoogleClientSecret != "" && !utils.IsEncrypted(c.GoogleClientSecret) ||
			c.SlackClientSecret != "" && !utils.IsEncrypted(c.SlackClientSecret) ||
			c.OidcClientSecret != "" && !utils.IsEncrypted(c.OidcClientSecret) ||
			c.LdapBindPassword != "" && !utils.IsEncrypted(c.LdapBindPassword) ||
			c.LogShipToken != "" && !utils.IsEncrypted(c.LogShipToken)

		if hasSecrets {
			return errors.New("master key not initialized - cannot save secrets")
		}
		return nil
	}

	// Encrypt OAuth secrets - fail on any encryption error to prevent plaintext storage
	if c.GithubClientSecret != "" && !utils.IsEncrypted(c.GithubClientSecret) {
		encrypted, err := utils.Encrypt(c.GithubClientSecret)
		if err != nil {
			return errors.Wrap(err, "failed to encrypt GithubClientSecret")
		}
		c.GithubClientSecret = encrypted
	}
	if c.GoogleClientSecret != "" && !utils.IsEncrypted(c.GoogleClientSecret) {
		encrypted, err := utils.Encrypt(c.GoogleClientSecret)
		if err != nil {
			return errors.Wrap(err, "failed to encrypt GoogleClientSecret")
		}
		c.GoogleClientSecret = encrypted
	}
	if c.SlackClientSecret != "" && !utils.IsEncrypted(c.SlackClientSecret) {
		encrypted, err := utils.Encrypt(c.SlackClientSecret)
		if err != nil {
			return errors.Wrap(err, "failed to encrypt SlackClientSecret")
		}
		c.SlackClientSecret = encrypted
	}
	if c.OidcClientSecret != "" && !utils.IsEncrypted(c.OidcClientSecret) {
		encrypted, err := utils.Encrypt(c.OidcClientSecret)
		if err != nil {
			return errors.Wrap(err, "failed to encrypt OidcClientSecret")
		}
		c.OidcClientSecret = encrypted
	}

	// LDAP bind password
	if c.LdapBindPassword != "" && !utils.IsEncrypted(c.LdapBindPassword) {
		encrypted, err := utils.Encrypt(c.LdapBindPassword)
		if err != nil {
			return errors.Wrap(err, "failed to encrypt LdapBindPassword")
		}
		c.LdapBindPassword = encrypted
	}

	// Log shipping token
	if c.LogShipToken != "" && !utils.IsEncrypted(c.LogShipToken) {
		encrypted, err := utils.Encrypt(c.LogShipToken)
		if err != nil {
			return errors.Wrap(err, "failed to encrypt LogShipToken")
		}
		c.LogShipToken = encrypted
	}
	return nil
}

func (c *Core) decryptSecrets() {
	// If master key not initialized, skip decryption (secrets stay encrypted)
	if !utils.MasterKeyInitialized() {
		return
	}

	// Decrypt OAuth secrets
	if c.GithubClientSecret != "" && utils.IsEncrypted(c.GithubClientSecret) {
		if decrypted, err := utils.Decrypt(c.GithubClientSecret); err == nil {
			c.GithubClientSecret = decrypted
		} else {
			utils.Log.Warnf("Failed to decrypt GithubClientSecret: %v", err)
		}
	}
	if c.GoogleClientSecret != "" && utils.IsEncrypted(c.GoogleClientSecret) {
		if decrypted, err := utils.Decrypt(c.GoogleClientSecret); err == nil {
			c.GoogleClientSecret = decrypted
		} else {
			utils.Log.Warnf("Failed to decrypt GoogleClientSecret: %v", err)
		}
	}
	if c.SlackClientSecret != "" && utils.IsEncrypted(c.SlackClientSecret) {
		if decrypted, err := utils.Decrypt(c.SlackClientSecret); err == nil {
			c.SlackClientSecret = decrypted
		} else {
			utils.Log.Warnf("Failed to decrypt SlackClientSecret: %v", err)
		}
	}
	if c.OidcClientSecret != "" && utils.IsEncrypted(c.OidcClientSecret) {
		if decrypted, err := utils.Decrypt(c.OidcClientSecret); err == nil {
			c.OidcClientSecret = decrypted
		} else {
			utils.Log.Warnf("Failed to decrypt OidcClientSecret: %v", err)
		}
	}

	// LDAP bind password
	if c.LdapBindPassword != "" && utils.IsEncrypted(c.LdapBindPassword) {
		if decrypted, err := utils.Decrypt(c.LdapBindPassword); err == nil {
			c.LdapBindPassword = decrypted
		} else {
			utils.Log.Warnf("Failed to decrypt LdapBindPassword: %v", err)
		}
	}

	// Log shipping token
	if c.LogShipToken != "" && utils.IsEncrypted(c.LogShipToken) {
		if decrypted, err := utils.Decrypt(c.LogShipToken); err == nil {
			c.LogShipToken = decrypted
		} else {
			utils.Log.Warnf("Failed to decrypt LogShipToken: %v", err)
		}
	}
}

func Select() (*Core, error) {
	var c Core
	d := getDB()
	if d == nil {
		return nil, errors.New("database has not been initiated yet.")
	}
	sqlDB, err := d.DB()
	if err != nil || sqlDB.Ping() != nil {
		return nil, errors.New("database has not been initiated yet.")
	}
	exists := d.HasTable("core")
	if !exists {
		return nil, errors.New("core database has not been setup yet.")
	}
	q := d.First(&c)
	if q.Error() != nil {
		return nil, q.Error()
	}

	// ALLOW_REPORTS controls digest feature availability (default: true)
	c.AllowReports = null.NewNullBool(utils.Params.GetBool("ALLOW_REPORTS"))
	if utils.Params.GetString("LANGUAGE") != "" {
		c.Language = utils.Params.GetString("LANGUAGE")
	}
	if utils.Params.GetString("API_SECRET") != "" {
		c.ApiSecret = utils.Params.GetString("API_SECRET")
	}
	c.Version = utils.Params.GetString("VERSION")
	c.Commit = utils.Params.GetString("COMMIT")
	c.Environment = utils.Params.GetString("STATPING_ENV")

	// Use thread-safe setter for global App
	SetApp(&c)
	return GetApp(), q.Error()
}

func (c *Core) Create() error {
	d := getDB()
	if d == nil {
		return errors.New("database has not been initiated yet.")
	}
	if c.ApiSecret == "" {
		c.ApiSecret = utils.RandomString(32)
		apiEnv := utils.Params.GetString("API_SECRET")
		if apiEnv != "" {
			c.ApiSecret = apiEnv
		}
	}
	// Note: EncryptionKey field is deprecated - external master key is now used
	// Field kept for backward compatibility with existing databases
	q := d.Create(c)
	utils.Log.Infof("API Key created: %s", c.ApiSecret)
	return q.Error()
}

func (c *Core) Update() error {
	d := getDB()
	if d == nil {
		return errors.New("database has not been initiated yet.")
	}
	q := d.Model(&Core{}).Where("1 = 1").UpdateColumns(c)
	return q.Error()
}

func (c *Core) Delete() error {
	return nil
}
