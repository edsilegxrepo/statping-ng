package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/statping-ng/statping-ng/notifiers"
	"github.com/statping-ng/statping-ng/types/configs"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

func setupRedirectHandler(w http.ResponseWriter, r *http.Request) {
	if core.App.Setup {
		http.Redirect(w, r, basePath, http.StatusFound)
		return
	}
	baseHandler(w, r)
}

func processSetupHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	if core.App.Setup {
		sendErrorJson(errors.New("statping has already been setup"), w, r)
		return
	}

	confgs, err := configs.LoadConfigForm(r)
	if err != nil {
		log.Errorln(err)
		sendErrorJson(err, w, r)
		return
	}

	project := r.PostForm.Get("project")
	description := r.PostForm.Get("description")
	domain := r.PostForm.Get("domain")
	sendReports, _ := strconv.ParseBool(r.PostForm.Get("send_reports"))
	sampleData := r.PostForm.Get("sample_data") == "on" || r.PostForm.Get("sample_data") == "true" || r.PostForm.Get("sample_data") == "1"

	log.WithFields(utils.ToFields(core.App, confgs)).Debugln("new configs posted")

	if err = configs.ConnectConfigs(confgs, false); err != nil {
		log.Errorln(err)
		sendErrorJson(err, w, r)
		return
	}

	if err := confgs.Save(utils.Directory); err != nil {
		log.Errorln(err)
		sendErrorJson(err, w, r)
		return
	}

	var adminPass string
	var samplePasses map[string]string

	if !core.App.Setup {
		if err := confgs.DropDatabase(); err != nil {
			sendErrorJson(err, w, r)
			return
		}

		if err := confgs.CreateDatabase(); err != nil {
			sendErrorJson(err, w, r)
			return
		}

		if err = confgs.MigrateDatabase(); err != nil {
			sendErrorJson(err, w, r)
			return
		}

		adminPass, err = configs.CreateAdminUser()
		if err != nil {
			sendErrorJson(err, w, r)
			return
		}

		log.Infoln("Migrating Notifiers...")
		notifiers.InitNotifiers()

		if sampleData {
			samplePasses, err = configs.TriggerSamples()
			if err != nil {
				sendErrorJson(err, w, r)
				return
			}
		}
		confgs.SampleData = false
	}

	c := &core.Core{
		Name:         project,
		Description:  description,
		ApiSecret:    utils.Params.GetString("API_SECRET"),
		Domain:       domain,
		Version:      core.App.Version,
		Started:      utils.Now(),
		CreatedAt:    utils.Now(),
		UseCdn:       null.NewNullBool(false),
		Footer:       null.NewNullString(""),
		Language:     confgs.Language,
		AllowReports: null.NewNullBool(sendReports),
	}

	log.Infoln("Creating new Core")
	if err := c.Create(); err != nil {
		log.Errorln(err)
		sendErrorJson(err, w, r)
		return
	}

	core.App = c

	log.Infoln("Initializing new Statping instance")

	if _, err := services.SelectAllServices(true); err != nil {
		log.Errorln(err)
		sendErrorJson(err, w, r)
		return
	}

	services.CheckServices()

	core.App.Setup = true

	resetCookies()

	out := struct {
		Message string            `json:"message"`
		Config  *configs.DbConfig `json:"config"`
		Admin   string            `json:"admin_password,omitempty"`
		Samples map[string]string `json:"sample_passwords,omitempty"`
	}{
		"success",
		confgs,
		adminPass,
		samplePasses,
	}
	returnJson(out, w, r)
}
