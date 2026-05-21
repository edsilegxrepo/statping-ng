package users

import (
	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/metrics"
	"github.com/statping-ng/statping-ng/utils"
	"gorm.io/gorm"
)

var (
	db  database.Database
	log = utils.Log.WithField("type", "user")
)

func SetDB(database database.Database) {
	db = database
}

func (u *User) AfterFind(tx *gorm.DB) (err error) {
	metrics.Query("user", "find")
	return nil
}

func (u *User) AfterCreate(tx *gorm.DB) (err error) {
	metrics.Query("user", "create")
	return nil
}

func (u *User) AfterUpdate(tx *gorm.DB) (err error) {
	metrics.Query("user", "update")
	return nil
}

func (u *User) AfterDelete(tx *gorm.DB) (err error) {
	metrics.Query("user", "delete")
	return nil
}

func Find(id int64) (*User, error) {
	var user User
	q := db.Where("id = ?", id).First(&user)
	return &user, q.Error()
}

func FindByUsername(username string) (*User, error) {
	var user User
	q := db.Where("username = ?", username).First(&user)
	return &user, q.Error()
}

func FindByAPIKey(key string) (*User, error) {
	var user User
	q := db.Where("api_key = ?", key).First(&user)
	return &user, q.Error()
}

func All() []*User {
	var users []*User
	db.Find(&users)
	return users
}

func (u *User) Create() error {
	q := db.Create(u)
	return q.Error()
}

func (u *User) Update() error {
	q := db.Update(u)
	return q.Error()
}

func (u *User) Delete() error {
	q := db.Delete(u)
	if db.Error() == nil {
		log.Warnf("User #%d (%s) has been deleted", u.Id, u.Username)
	}
	return q.Error()
}
