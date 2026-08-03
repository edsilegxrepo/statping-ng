package users

import (
	"strings"

	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/utils"
	"gorm.io/gorm"
)

func (u *User) Validate() error {
	if u.Username == "" {
		return errors.New("username is empty")
	} else if u.Password == "" {
		return errors.New("password is empty")
	} else if u.Email == "" {
		return errors.New("email is empty")
	}

	if !utils.IsHash(u.Password) {
		if !utils.ComplexityCheck(u.Password) {
			return errors.New("password must be at least 30 characters and include uppercase, lowercase, and digits")
		}
	}

	return nil
}

func (u *User) BeforeDelete(tx *gorm.DB) (err error) {
	if utils.Params.GetBool("ADMIN_LOCK") {
		if u.Username == "admin" {
			return errors.New("cannot delete admin in ADMIN_LOCK")
		}
	}
	return nil
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	// Normalize username and email to lowercase for case-insensitive lookups
	u.Username = strings.ToLower(u.Username)
	u.Email = strings.ToLower(u.Email)

	if err := u.Validate(); err != nil {
		return err
	}
	u.Password = utils.HashPassword(u.Password)
	u.ApiKey = utils.NewSHA256Hash()
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	return u.Validate()
}
