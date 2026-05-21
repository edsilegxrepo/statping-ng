package notifications

import (
	"github.com/statping-ng/statping-ng/types/metrics"
	"gorm.io/gorm"
)

func (n *Notification) AfterFind(tx *gorm.DB) (err error) {
	metrics.Query("notifier", "find")
	return nil
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
