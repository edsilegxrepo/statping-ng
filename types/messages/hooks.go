package messages

import (
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/metrics"
	"gorm.io/gorm"
)

func (m *Message) Validate() error {
	if m.Title == "" {
		return errors.New("missing message title")
	}
	return nil
}

func (m *Message) BeforeUpdate(tx *gorm.DB) (err error) {
	return m.Validate()
}

func (m *Message) BeforeCreate(tx *gorm.DB) (err error) {
	return m.Validate()
}

func (m *Message) AfterFind(tx *gorm.DB) (err error) {
	metrics.Query("message", "find")
	return nil
}

func (m *Message) AfterCreate(tx *gorm.DB) (err error) {
	metrics.Query("message", "create")
	return nil
}

func (m *Message) AfterUpdate(tx *gorm.DB) (err error) {
	metrics.Query("message", "update")
	return nil
}

func (m *Message) AfterDelete(tx *gorm.DB) (err error) {
	metrics.Query("message", "delete")
	return nil
}
