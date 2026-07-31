package core

import (
	"github.com/pkg/errors"
	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/metrics"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/utils"
	"gorm.io/gorm"
)

var db database.Database

func SetDB(database database.Database) {
	db = database
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
	// Check if there are secrets that need encryption
	hasSecrets := c.GithubClientSecret != "" && !utils.IsEncrypted(c.GithubClientSecret) ||
		c.GoogleClientSecret != "" && !utils.IsEncrypted(c.GoogleClientSecret) ||
		c.SlackClientSecret != "" && !utils.IsEncrypted(c.SlackClientSecret) ||
		c.CustomClientSecret != "" && !utils.IsEncrypted(c.CustomClientSecret) ||
		c.LdapBindPassword != "" && !utils.IsEncrypted(c.LdapBindPassword)

	if c.EncryptionKey == "" {
		if hasSecrets {
			return errors.New("encryption key not available - cannot save secrets")
		}
		return nil
	}
	key := c.EncryptionKey

	// Encrypt OAuth secrets - fail on any encryption error to prevent plaintext storage
	if c.GithubClientSecret != "" && !utils.IsEncrypted(c.GithubClientSecret) {
		encrypted, err := utils.Encrypt(c.GithubClientSecret, key)
		if err != nil {
			return errors.Wrap(err, "failed to encrypt GithubClientSecret")
		}
		c.GithubClientSecret = encrypted
	}
	if c.GoogleClientSecret != "" && !utils.IsEncrypted(c.GoogleClientSecret) {
		encrypted, err := utils.Encrypt(c.GoogleClientSecret, key)
		if err != nil {
			return errors.Wrap(err, "failed to encrypt GoogleClientSecret")
		}
		c.GoogleClientSecret = encrypted
	}
	if c.SlackClientSecret != "" && !utils.IsEncrypted(c.SlackClientSecret) {
		encrypted, err := utils.Encrypt(c.SlackClientSecret, key)
		if err != nil {
			return errors.Wrap(err, "failed to encrypt SlackClientSecret")
		}
		c.SlackClientSecret = encrypted
	}
	if c.CustomClientSecret != "" && !utils.IsEncrypted(c.CustomClientSecret) {
		encrypted, err := utils.Encrypt(c.CustomClientSecret, key)
		if err != nil {
			return errors.Wrap(err, "failed to encrypt CustomClientSecret")
		}
		c.CustomClientSecret = encrypted
	}

	// LDAP bind password
	if c.LdapBindPassword != "" && !utils.IsEncrypted(c.LdapBindPassword) {
		encrypted, err := utils.Encrypt(c.LdapBindPassword, key)
		if err != nil {
			return errors.Wrap(err, "failed to encrypt LdapBindPassword")
		}
		c.LdapBindPassword = encrypted
	}
	return nil
}

func (c *Core) decryptSecrets() {
	if c.EncryptionKey == "" {
		return
	}
	key := c.EncryptionKey

	// Decrypt OAuth secrets
	if c.GithubClientSecret != "" && utils.IsEncrypted(c.GithubClientSecret) {
		if decrypted, err := utils.Decrypt(c.GithubClientSecret, key); err == nil {
			c.GithubClientSecret = decrypted
		} else {
			utils.Log.Warnf("Failed to decrypt GithubClientSecret: %v", err)
		}
	}
	if c.GoogleClientSecret != "" && utils.IsEncrypted(c.GoogleClientSecret) {
		if decrypted, err := utils.Decrypt(c.GoogleClientSecret, key); err == nil {
			c.GoogleClientSecret = decrypted
		} else {
			utils.Log.Warnf("Failed to decrypt GoogleClientSecret: %v", err)
		}
	}
	if c.SlackClientSecret != "" && utils.IsEncrypted(c.SlackClientSecret) {
		if decrypted, err := utils.Decrypt(c.SlackClientSecret, key); err == nil {
			c.SlackClientSecret = decrypted
		} else {
			utils.Log.Warnf("Failed to decrypt SlackClientSecret: %v", err)
		}
	}
	if c.CustomClientSecret != "" && utils.IsEncrypted(c.CustomClientSecret) {
		if decrypted, err := utils.Decrypt(c.CustomClientSecret, key); err == nil {
			c.CustomClientSecret = decrypted
		} else {
			utils.Log.Warnf("Failed to decrypt CustomClientSecret: %v", err)
		}
	}

	// LDAP bind password
	if c.LdapBindPassword != "" && utils.IsEncrypted(c.LdapBindPassword) {
		if decrypted, err := utils.Decrypt(c.LdapBindPassword, key); err == nil {
			c.LdapBindPassword = decrypted
		} else {
			utils.Log.Warnf("Failed to decrypt LdapBindPassword: %v", err)
		}
	}
}

func Select() (*Core, error) {
	var c Core
	sqlDB, err := db.DB()
	if err != nil || sqlDB.Ping() != nil {
		return nil, errors.New("database has not been initiated yet.")
	}
	exists := db.HasTable("core")
	if !exists {
		return nil, errors.New("core database has not been setup yet.")
	}
	q := db.First(&c)
	if q.Error() != nil {
		return nil, q.Error()
	}
	App = &c

	if utils.Params.GetBool("USE_CDN") {
		App.UseCdn = null.NewNullBool(true)
	}
	// ALLOW_REPORTS controls digest feature availability (default: true)
	App.AllowReports = null.NewNullBool(utils.Params.GetBool("ALLOW_REPORTS"))
	if utils.Params.GetString("LANGUAGE") != "" {
		App.Language = utils.Params.GetString("LANGUAGE")
	}
	if utils.Params.GetString("API_SECRET") != "" {
		App.ApiSecret = utils.Params.GetString("API_SECRET")
	}
	App.Version = utils.Params.GetString("VERSION")
	App.Commit = utils.Params.GetString("COMMIT")
	return App, q.Error()
}

func (c *Core) Create() error {
	if c.ApiSecret == "" {
		c.ApiSecret = utils.RandomString(32)
		apiEnv := utils.Params.GetString("API_SECRET")
		if apiEnv != "" {
			c.ApiSecret = apiEnv
		}
	}
	// Generate a separate encryption key - never exposed via API
	if c.EncryptionKey == "" {
		c.EncryptionKey = utils.NewSHA256Hash()
	}
	q := db.Create(c)
	utils.Log.Infof("API Key created: %s", c.ApiSecret)
	return q.Error()
}

func (c *Core) Update() error {
	q := db.Model(&Core{}).Where("1 = 1").UpdateColumns(c)
	return q.Error()
}

func (c *Core) Delete() error {
	return nil
}
