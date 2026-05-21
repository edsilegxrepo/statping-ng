package hits

import (
	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/metrics"
	"github.com/statping-ng/statping-ng/utils"
	"gorm.io/gorm"
)

var log = utils.Log

var db database.Database

func SetDB(database database.Database) {
	db = database
}

func (h *Hit) AfterFind(tx *gorm.DB) (err error) {
	metrics.Query("hit", "find")
	return nil
}

func (h *Hit) AfterUpdate(tx *gorm.DB) (err error) {
	metrics.Query("hit", "update")
	return nil
}

func (h *Hit) AfterDelete(tx *gorm.DB) (err error) {
	metrics.Query("hit", "delete")
	return nil
}

func (h *Hit) AfterCreate(tx *gorm.DB) (err error) {
	metrics.Query("hit", "create")
	return nil
}

func (h *Hit) Create() error {
	q := db.Create(h)
	return q.Error()
}

func (h *Hit) Update() error {
	q := db.Update(h)
	return q.Error()
}

func (h *Hit) Delete() error {
	q := db.Delete(h)
	return q.Error()
}
