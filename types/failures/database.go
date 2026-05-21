package failures

import (
	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/metrics"
	"gorm.io/gorm"
)

var db database.Database

func SetDB(database database.Database) {
	db = database
}

func DB() database.Database {
	return db
}

func (f *Failure) AfterFind(tx *gorm.DB) (err error) {
	metrics.Query("failure", "find")
	return nil
}

func (f *Failure) AfterUpdate(tx *gorm.DB) (err error) {
	metrics.Query("failure", "update")
	return nil
}

func (f *Failure) AfterDelete(tx *gorm.DB) (err error) {
	metrics.Query("failure", "delete")
	return nil
}

func (f *Failure) AfterCreate(tx *gorm.DB) (err error) {
	metrics.Query("failure", "create")
	return nil
}

func (f *Failure) Create() error {
	q := db.Create(f)
	return q.Error()
}

func (f *Failure) Update() error {
	q := db.Update(f)
	return q.Error()
}

func (f *Failure) Delete() error {
	q := db.Delete(f)
	return q.Error()
}
