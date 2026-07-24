package users

import (
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/utils"
)

func Samples() (map[string]string, error) {
	log.Infoln("Inserting Sample Users...")
	pass1 := utils.RandomString(28) + "Aa1!"
	pass2 := utils.RandomString(28) + "Bb2!"

	u2 := &User{
		Username: "testadmin",
		Password: pass1,
		Email:    "info@betatude.com",
		Scopes:   "admin",
		Admin:    null.NewNullBool(true),
	}

	if err := u2.Create(); err != nil {
		return nil, err
	}

	u3 := &User{
		Username: "testadmin2",
		Password: pass2,
		Email:    "info@adminhere.com",
		Scopes:   "admin",
		Admin:    null.NewNullBool(true),
	}

	if err := u3.Create(); err != nil {
		return nil, err
	}

	return map[string]string{
		"testadmin":  pass1,
		"testadmin2": pass2,
	}, nil
}
