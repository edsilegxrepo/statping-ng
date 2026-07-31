package notifications

import (
	"fmt"

	"github.com/statping-ng/statping-ng/types/metrics"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/utils"
	"gorm.io/gorm"
)

func (n *Notification) AfterFind(tx *gorm.DB) (err error) {
	metrics.Query("notifier", "find")
	// Decrypt secrets after loading from database
	n.decryptSecrets()
	return nil
}

func (n *Notification) BeforeSave(tx *gorm.DB) (err error) {
	// Encrypt secrets before saving to database
	if err := n.encryptSecrets(); err != nil {
		return err
	}
	return nil
}

func (n *Notification) encryptSecrets() error {
	// Master key must be initialized before encrypting secrets
	if !utils.MasterKeyInitialized() {
		// Check if there are secrets that need encryption
		hasSecrets := (n.Password.String != "" && !utils.IsEncrypted(n.Password.String)) ||
			(n.ApiSecret.String != "" && !utils.IsEncrypted(n.ApiSecret.String))

		if hasSecrets {
			return fmt.Errorf("master key not initialized - cannot save notifier %s secrets", n.Method)
		}
		return nil
	}

	// Encrypt Password - fail on error to prevent plaintext storage
	if n.Password.String != "" && !utils.IsEncrypted(n.Password.String) {
		encrypted, err := utils.Encrypt(n.Password.String)
		if err != nil {
			return fmt.Errorf("failed to encrypt notifier %s password: %w", n.Method, err)
		}
		n.Password = null.NewNullString(encrypted)
	}

	// Encrypt ApiSecret - fail on error to prevent plaintext storage
	if n.ApiSecret.String != "" && !utils.IsEncrypted(n.ApiSecret.String) {
		encrypted, err := utils.Encrypt(n.ApiSecret.String)
		if err != nil {
			return fmt.Errorf("failed to encrypt notifier %s api_secret: %w", n.Method, err)
		}
		n.ApiSecret = null.NewNullString(encrypted)
	}
	return nil
}

func (n *Notification) decryptSecrets() {
	// If master key not initialized, skip decryption
	if !utils.MasterKeyInitialized() {
		return
	}

	// Decrypt Password
	if n.Password.String != "" && utils.IsEncrypted(n.Password.String) {
		if decrypted, err := utils.Decrypt(n.Password.String); err == nil {
			n.Password = null.NewNullString(decrypted)
		} else {
			log.Warnf("Failed to decrypt notifier %s password: %v", n.Method, err)
		}
	}

	// Decrypt ApiSecret
	if n.ApiSecret.String != "" && utils.IsEncrypted(n.ApiSecret.String) {
		if decrypted, err := utils.Decrypt(n.ApiSecret.String); err == nil {
			n.ApiSecret = null.NewNullString(decrypted)
		} else {
			log.Warnf("Failed to decrypt notifier %s api_secret: %v", n.Method, err)
		}
	}
}

func (n *Notification) AfterCreate(tx *gorm.DB) (err error) {
	metrics.Query("notifier", "create")
	return nil
}

func (n *Notification) AfterUpdate(tx *gorm.DB) (err error) {
	metrics.Query("notifier", "update")
	return nil
}

func (n *Notification) AfterDelete(tx *gorm.DB) (err error) {
	metrics.Query("notifier", "delete")
	return nil
}
