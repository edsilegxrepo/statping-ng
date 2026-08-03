package users

import (
	"strings"

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

// FindByUsername finds a user by username (case-insensitive).
// Normalizes input to lowercase since usernames are stored lowercase.
func FindByUsername(username string) (*User, error) {
	var user User
	q := db.Where("username = ?", strings.ToLower(username)).First(&user)
	return &user, q.Error()
}

func FindByAPIKey(key string) (*User, error) {
	var user User
	q := db.Where("api_key = ?", key).First(&user)
	return &user, q.Error()
}

// FindByEmail finds a user by email (case-insensitive).
// Normalizes input to lowercase since emails are stored lowercase.
func FindByEmail(email string) (*User, error) {
	var user User
	q := db.Where("email = ?", strings.ToLower(email)).First(&user)
	return &user, q.Error()
}

func All() []*User {
	var users []*User
	db.Find(&users)
	return users
}

// CountEnabledAdmins returns the number of enabled admin users
func CountEnabledAdmins() int {
	var count int64
	db.Model(&User{}).Where("administrator = ?", true).Where("enabled = ?", true).Count(&count)
	return int(count)
}

// IsLastEnabledAdmin checks if the given user is the last enabled admin
func IsLastEnabledAdmin(userID int64) bool {
	user, err := Find(userID)
	if err != nil {
		return false
	}
	if !user.Admin.Bool || !user.Enabled.Bool {
		return false
	}
	return CountEnabledAdmins() == 1
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
