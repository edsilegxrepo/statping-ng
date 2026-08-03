package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/users"
	"github.com/statping-ng/statping-ng/utils"
)

func findUser(r *http.Request) (*users.User, int64, error) {
	vars := mux.Vars(r)
	if utils.NotNumber(vars["id"]) {
		return nil, 0, errors.NotNumber
	}
	num := utils.ToInt(vars["id"])
	user, err := users.Find(num)
	if err != nil {
		return nil, num, errors.Missing(&users.User{}, num)
	}
	return user, num, nil
}

func apiUserHandler(w http.ResponseWriter, r *http.Request) {
	user, _, err := findUser(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	user.Password = ""
	returnJson(user, w, r)
}

func apiUserUpdateHandler(w http.ResponseWriter, r *http.Request) {
	user, _, err := findUser(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	// Store original state for audit logging and last-admin check
	wasAdmin := user.Admin.Bool
	wasEnabled := user.Enabled.Bool
	originalUsername := user.Username

	err = DecodeJSON(r, &user)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	// Protect against demoting or disabling the last admin
	if wasAdmin && wasEnabled {
		beingDemoted := !user.Admin.Bool
		beingDisabled := !user.Enabled.Bool
		if (beingDemoted || beingDisabled) && users.IsLastEnabledAdmin(user.Id) {
			sendErrorJson(errors.New("cannot demote or disable the last enabled admin"), w, r)
			return
		}
	}

	passwordChanged := false
	if user.Password != "" {
		user.Password = utils.HashPassword(user.Password)
		passwordChanged = true
	}

	err = user.Update()
	if err != nil {
		sendErrorJson(fmt.Errorf("issue updating user #%d: %s", user.Id, err), w, r)
		return
	}

	// Audit log security-relevant changes
	if wasAdmin && !user.Admin.Bool {
		AuditLog(AuditAdminDemoted, r, map[string]interface{}{"username": originalUsername})
	} else if !wasAdmin && user.Admin.Bool {
		AuditLog(AuditAdminPromoted, r, map[string]interface{}{"username": originalUsername})
	}
	if wasEnabled && !user.Enabled.Bool {
		AuditLog(AuditUserDisabled, r, map[string]interface{}{"username": originalUsername})
	} else if !wasEnabled && user.Enabled.Bool {
		AuditLog(AuditUserEnabled, r, map[string]interface{}{"username": originalUsername})
	}
	if passwordChanged {
		AuditLog(AuditPasswordChanged, r, map[string]interface{}{"username": originalUsername})
	}

	AuditLog(AuditUserUpdated, r, map[string]interface{}{"username": originalUsername})

	// Clear sensitive data before returning
	user.Password = ""
	sendJsonAction(user, "update", w, r)
}

func apiUserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	allUsers := users.All()
	if len(allUsers) == 1 {
		sendErrorJson(errors.New("cannot delete the last user"), w, r)
		return
	}
	user, _, err := findUser(r)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	// Protect against deleting the last enabled admin
	if users.IsLastEnabledAdmin(user.Id) {
		sendErrorJson(errors.New("cannot delete the last enabled admin"), w, r)
		return
	}

	username := user.Username
	if err := user.Delete(); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	AuditLog(AuditUserDeleted, r, map[string]interface{}{"username": username})
	sendJsonAction(user, "delete", w, r)
}

func apiAllUsersHandler(r *http.Request) interface{} {
	allUsers := users.All()
	return allUsers
}

func apiAuthProvidersHandler(w http.ResponseWriter, r *http.Request) {
	returnJson(users.GetAuthProviders(), w, r)
}

func apiCheckUserTokenHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		sendErrorJson(err, w, r)
		return
	}
	token := r.PostForm.Get("token")
	if token == "" {
		sendErrorJson(errors.New("missing token parameter"), w, r)
		return
	}

	claim, err := parseToken(token)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	returnJson(claim, w, r)
}

func apiGetUserTokenHandler(w http.ResponseWriter, r *http.Request) {
	claim, err := getJwtToken(r)
	if err != nil {
		// Not logged in - return empty response
		returnJson(nil, w, r)
		return
	}
	returnJson(claim, w, r)
}

func apiCreateUsersHandler(w http.ResponseWriter, r *http.Request) {
	var user *users.User
	err := DecodeJSON(r, &user)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	err = user.Create()
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	AuditLog(AuditUserCreated, r, map[string]interface{}{
		"username": user.Username,
		"is_admin": user.Admin.Bool,
	})

	// Clear sensitive data before returning
	user.Password = ""
	sendJsonAction(user, "create", w, r)
}
