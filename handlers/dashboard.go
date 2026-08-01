package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/statping-ng/statping-ng/source"
	"github.com/statping-ng/statping-ng/types/checkins"
	"github.com/statping-ng/statping-ng/types/configs"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/groups"
	"github.com/statping-ng/statping-ng/types/incidents"
	"github.com/statping-ng/statping-ng/types/messages"
	"github.com/statping-ng/statping-ng/types/notifications"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/types/users"
	"github.com/statping-ng/statping-ng/utils"
	"gopkg.in/yaml.v2"
)

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	removeJwtToken(w, r)

	// Check if forward auth logout URL is configured
	logoutURL := ForwardAuthLogoutURL()
	if logoutURL != "" {
		out := make(map[string]string)
		out["status"] = "success"
		out["redirect"] = logoutURL
		returnJson(out, w, r)
		return
	}

	out := make(map[string]string)
	out["status"] = "success"
	returnJson(out, w, r)
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	utils.LockLines.Lock()
	logs := make([]string, 0)
	length := len(utils.LastLines)
	// We need string log lines from end to start.
	for i := length - 1; i >= 0; i-- {
		logs = append(logs, utils.LastLines[i].FormatForHtml()+"\r\n")
	}
	utils.LockLines.Unlock()
	returnJson(logs, w, r)
}

// customCSSResponse is the response for custom CSS API
type customCSSResponse struct {
	CSS     string `json:"css"`
	Enabled bool   `json:"enabled"`
}

// apiThemeViewHandler returns the current custom CSS
func apiThemeViewHandler(w http.ResponseWriter, r *http.Request) {
	css, _ := source.LoadCustomCSS()
	resp := customCSSResponse{
		CSS:     css,
		Enabled: source.HasCustomCSS(),
	}
	returnJson(resp, w, r)
}

// apiThemeSaveHandler saves custom CSS
func apiThemeSaveHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CSS string `json:"css"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	if err := source.SaveCustomCSS(req.CSS); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	resetRouter()
	sendJsonAction(req, "saved", w, r)
}

// apiThemeCreateHandler is now a no-op (CSS doesn't need initialization)
func apiThemeCreateHandler(w http.ResponseWriter, r *http.Request) {
	sendJsonAction("custom CSS ready", "created", w, r)
}

// apiThemeRemoveHandler deletes custom CSS
func apiThemeRemoveHandler(w http.ResponseWriter, r *http.Request) {
	if err := source.DeleteCustomCSS(); err != nil {
		log.Errorln(fmt.Errorf("error deleting custom CSS: %v", err))
	}
	resetRouter()
	sendJsonAction("custom.css", "deleted", w, r)
}

type ExportData struct {
	Config          *configs.DbConfig            `json:"config,omitempty"`
	Core            *core.Core                   `json:"core"`
	Services        []*services.Service          `json:"services"`
	Messages        []*messages.Message          `json:"messages"`
	Incidents       []*incidents.Incident        `json:"incidents"`
	IncidentUpdates []*incidents.IncidentUpdate  `json:"incident_updates"`
	Checkins        []*checkins.Checkin          `json:"checkins"`
	Users           []*users.User                `json:"users"`
	Groups          []*groups.Group              `json:"groups"`
	Notifiers       []notifications.Notification `json:"notifiers"`
}

func (e *ExportData) JSON() []byte {
	d, _ := json.Marshal(e)
	return d
}

func ExportSettings() (*ExportData, error) {
	var notifiers []notifications.Notification
	for _, n := range services.AllNotifiers() {
		notifiers = append(notifiers, *n.Select())
	}

	data := &ExportData{
		Core:      core.App,
		Notifiers: notifiers,
		Checkins:  checkins.All(),
		Users:     users.All(),
		Services:  services.AllInOrder(),
		Groups:    groups.All(),
		Messages:  messages.All(),
	}
	return data, nil
}

func settingsImportHandler(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	defer func() { _ = r.Body.Close() }()

	var exportData *ExportData
	if err := json.Unmarshal(data, &exportData); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	if exportData.Core != nil {
		core.App = exportData.Core
		if err := core.App.Update(); err != nil {
			sendErrorJson(err, w, r)
			return
		}
	}

	if exportData.Groups != nil {
		for _, s := range exportData.Groups {
			s.Id = 0
			if err := s.Create(); err != nil {
				sendErrorJson(err, w, r)
				return
			}
		}
	}

	if exportData.Services != nil {
		for _, s := range exportData.Services {
			s.Id = 0
			if err := s.Create(); err != nil {
				sendErrorJson(err, w, r)
				return
			}
		}
	}

	if exportData.Users != nil {
		for _, s := range exportData.Users {
			s.Id = 0
			if err := s.Create(); err != nil {
				sendErrorJson(err, w, r)
				return
			}
		}
	}

	if exportData.Notifiers != nil {
		for _, s := range exportData.Notifiers {
			notif := services.ReturnNotifier(s.Method)
			n := notif.Select().UpdateFields(&s)
			if err := n.Update(); err != nil {
				sendErrorJson(err, w, r)
				return
			}
		}
	}

	sendJsonAction(exportData, "import", w, r)
}

func configsSaveHandler(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	defer func() { _ = r.Body.Close() }()

	var cfg *configs.DbConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	oldCfg, err := configs.LoadConfigs(utils.Directory + "/config.yml")
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	newCfg := cfg.Merge(oldCfg)
	if err := newCfg.Save(utils.Directory); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	sendJsonAction(newCfg.Clean(), "updated", w, r)
}

func configsViewHandler(w http.ResponseWriter, r *http.Request) {
	db, err := configs.LoadConfigs(utils.Directory + "/config.yml")
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}
	_, _ = w.Write(db.Clean().ToYAML())
}

func settingsExportHandler(w http.ResponseWriter, r *http.Request) {
	exported, err := ExportSettings()
	if err != nil {
		sendErrorJson(err, w, r)
		return
	}

	file := bytes.NewBuffer(exported.JSON())

	w.Header().Set("Content-Disposition", "attachment; filename=statping.json")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", utils.ToString(len(exported.JSON())))

	_, _ = io.Copy(w, file)
}

func logsLineHandler(w http.ResponseWriter, r *http.Request) {
	if lastLine := utils.GetLastLine(); lastLine != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(lastLine.FormatForHtml()))
	}
}

func apiLoginHandler(w http.ResponseWriter, r *http.Request) {
	form := parseForm(r)
	username := form.Get("username")
	password := form.Get("password")

	var user *users.User
	var auth bool

	// Try LDAP authentication first if enabled
	if core.App.LdapEnabled.Bool {
		ldapUser, err := processLDAPLogin(username, password)
		if err != nil {
			log.Warnln(fmt.Sprintf("LDAP auth error for user %v: %v", username, err))
			// Check if it's a group membership error
			if err.Error() == "user is not a member of the authorized group" {
				returnJson(struct {
					Error string `json:"error"`
				}{"not authorized to access this application"}, w, r)
				return
			}
		}
		if ldapUser != nil {
			user = ldapUser
			auth = true
		}
	}

	// Fallback to local authentication
	if !auth {
		user, auth = users.AuthUser(username, password)
	}

	if auth {
		// Check if user is enabled
		if !user.Enabled.Bool {
			log.Infoln(fmt.Sprintf("User %v login rejected - account pending approval", user.Username))
			returnJson(struct {
				Error string `json:"error"`
			}{"account pending approval - please contact an administrator"}, w, r)
			return
		}

		log.Infoln(fmt.Sprintf("User %v logged in from IP %v", user.Username, r.RemoteAddr))
		claim, token := setJwtToken(user, w, r)
		resp := struct {
			Token   string `json:"token"`
			IsAdmin bool   `json:"admin"`
		}{
			token,
			claim.Admin,
		}
		returnJson(resp, w, r)
	} else {
		resp := struct {
			Error string `json:"error"`
		}{
			"incorrect authentication",
		}
		returnJson(resp, w, r)
	}
}
