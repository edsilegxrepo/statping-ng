package handlers

import (
	"net/http"

	"github.com/statping-ng/statping-ng/notifiers"
	"github.com/statping-ng/statping-ng/types/core"
)

func apiDigestSettingsHandler(w http.ResponseWriter, r *http.Request) {
	c := core.App
	settings := map[string]interface{}{
		"digest_enabled": c.DigestEnabled.Bool,
		"digest_emails":  c.DigestEmails,
		"digest_hour":    c.DigestHour,
	}
	returnJson(settings, w, r)
}

func apiDigestSaveHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DigestEnabled bool   `json:"digest_enabled"`
		DigestEmails  string `json:"digest_emails"`
		DigestHour    int    `json:"digest_hour"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	// Validate digest hour (0-23)
	if req.DigestHour < 0 || req.DigestHour > 23 {
		req.DigestHour = 8 // default to 8 AM
	}

	c := core.App
	c.DigestEnabled.Bool = req.DigestEnabled
	c.DigestEnabled.Valid = true
	c.DigestEmails = req.DigestEmails
	c.DigestHour = req.DigestHour

	if err := c.Update(); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	// Restart scheduler if settings changed
	notifiers.StopDigestScheduler()
	if req.DigestEnabled {
		notifiers.StartDigestScheduler()
	}

	returnJson(map[string]string{"status": "success"}, w, r)
}

func apiDigestTestHandler(w http.ResponseWriter, r *http.Request) {
	if err := notifiers.SendTestDigest(); err != nil {
		returnJson(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		}, w, r)
		return
	}

	returnJson(map[string]interface{}{
		"success": true,
		"message": "Test digest sent successfully",
	}, w, r)
}

func apiDigestSmtpTestHandler(w http.ResponseWriter, r *http.Request) {
	result := notifiers.TestSMTPConnection()
	returnJson(result, w, r)
}
