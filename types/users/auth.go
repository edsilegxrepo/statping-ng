package users

import (
	"fmt"
	"time"

	"github.com/statping-ng/statping-ng/utils"
)

// AuthUser will return the User and a boolean if authentication was correct.
// accepts username, and password as a string
func AuthUser(username, passwordHash string) (*User, bool) {
	user, err := FindByUsername(username)
	if err != nil {
		log.Warnln(fmt.Errorf("user %v not found", username))
		return nil, false
	}
	if utils.CheckHash(passwordHash, user.Password) {
		user.UpdatedAt = time.Now().UTC()
		if err := user.Update(); err != nil {
			log.Error(err)
		}
		return user, true
	}
	return nil, false
}
