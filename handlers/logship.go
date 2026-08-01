package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/utils"
)

// apiLogShipSettingsHandler returns current log shipping settings
func apiLogShipSettingsHandler(w http.ResponseWriter, r *http.Request) {
	c := core.App

	// Check if env vars are set (they take precedence)
	envOverride := utils.Params.GetString("LOG_SHIP_TYPE") != "" ||
		utils.Params.GetString("LOG_SHIP_ENDPOINT") != ""

	settings := map[string]interface{}{
		"log_ship_enabled":    c.LogShipEnabled.Bool,
		"log_ship_type":       c.LogShipType,
		"log_ship_endpoint":   c.LogShipEndpoint,
		"log_ship_index":      c.LogShipIndex,
		"log_ship_sourcetype": c.LogShipSourcetype,
		"log_ship_labels":     c.LogShipLabels,
		"env_override":        envOverride, // Let UI know if env vars are in use
		"types": []map[string]string{
			{"value": "loki", "label": "Grafana Loki"},
			{"value": "elasticsearch", "label": "Elasticsearch"},
			{"value": "splunk", "label": "Splunk HEC"},
			{"value": "cribl", "label": "Cribl"},
			{"value": "webhook", "label": "Generic Webhook"},
		},
	}
	returnJson(settings, w, r)
}

// apiLogShipSaveHandler saves log shipping settings
func apiLogShipSaveHandler(w http.ResponseWriter, r *http.Request) {
	var settings core.LogShipping
	if err := DecodeJSON(r, &settings); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	c := core.App
	c.LogShipEnabled = settings.LogShipEnabled
	c.LogShipType = strings.ToLower(strings.TrimSpace(settings.LogShipType))
	c.LogShipEndpoint = strings.TrimSpace(settings.LogShipEndpoint)
	c.LogShipIndex = strings.TrimSpace(settings.LogShipIndex)
	c.LogShipSourcetype = strings.TrimSpace(settings.LogShipSourcetype)
	c.LogShipLabels = strings.TrimSpace(settings.LogShipLabels)

	// Only update token if provided (allows keeping existing encrypted token)
	if settings.LogShipToken != "" {
		c.LogShipToken = settings.LogShipToken
	}

	if err := c.Update(); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	// Reload log shipper with new config
	utils.ReloadLogShipper(&utils.LogShipConfig{
		Enabled:    c.LogShipEnabled.Bool,
		Type:       c.LogShipType,
		Endpoint:   c.LogShipEndpoint,
		Token:      c.LogShipToken,
		Index:      c.LogShipIndex,
		Sourcetype: c.LogShipSourcetype,
		Labels:     c.LogShipLabels,
	})

	returnJson(map[string]string{"status": "success"}, w, r)
}

// LogShipTestRequest is the request body for testing log shipping
type LogShipTestRequest struct {
	Type       string `json:"type"`
	Endpoint   string `json:"endpoint"`
	Token      string `json:"token"`
	Index      string `json:"index"`
	Sourcetype string `json:"sourcetype"`
}

// LogShipTestResponse is the response for log shipping test
type LogShipTestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// apiLogShipTestHandler tests log shipping connection
func apiLogShipTestHandler(w http.ResponseWriter, r *http.Request) {
	var req LogShipTestRequest
	if err := DecodeJSON(r, &req); err != nil {
		sendErrorJson(err, w, r)
		return
	}

	if req.Type == "" || req.Endpoint == "" {
		returnJson(LogShipTestResponse{
			Success: false,
			Message: "Type and endpoint are required",
		}, w, r)
		return
	}

	// Send a test log entry
	testEntry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     "info",
		"message":   "Statping log shipping test",
		"app":       "statping",
		"test":      true,
	}

	var err error
	switch strings.ToLower(req.Type) {
	case "loki":
		err = testLoki(req.Endpoint, req.Token, testEntry)
	case "elasticsearch":
		err = testElasticsearch(req.Endpoint, req.Token, testEntry)
	case "splunk":
		err = testSplunk(req.Endpoint, req.Token, req.Index, req.Sourcetype, testEntry)
	case "cribl":
		err = testCribl(req.Endpoint, req.Token, testEntry)
	case "webhook":
		err = testWebhook(req.Endpoint, req.Token, testEntry)
	default:
		err = fmt.Errorf("unknown log ship type: %s", req.Type)
	}

	if err != nil {
		returnJson(LogShipTestResponse{
			Success: false,
			Message: err.Error(),
		}, w, r)
		return
	}

	returnJson(LogShipTestResponse{
		Success: true,
		Message: "Test log sent successfully",
	}, w, r)
}

func testLoki(endpoint, token string, entry map[string]interface{}) error {
	payload := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": map[string]string{"app": "statping", "test": "true"},
				"values": [][]string{
					{fmt.Sprintf("%d", time.Now().UnixNano()), "Statping log shipping test"},
				},
			},
		},
	}
	return sendTestRequest(strings.TrimSuffix(endpoint, "/")+"/loki/api/v1/push", token, payload, false)
}

func testElasticsearch(endpoint, token string, entry map[string]interface{}) error {
	return sendTestRequest(strings.TrimSuffix(endpoint, "/")+"/statping-test/_doc", token, entry, false)
}

func testSplunk(endpoint, token, index, sourcetype string, entry map[string]interface{}) error {
	if sourcetype == "" {
		sourcetype = "statping"
	}
	if index == "" {
		index = "main"
	}
	payload := map[string]interface{}{
		"time":       float64(time.Now().Unix()),
		"host":       "statping",
		"source":     "statping",
		"sourcetype": sourcetype,
		"index":      index,
		"event":      entry,
	}
	url := strings.TrimSuffix(endpoint, "/")
	if !strings.HasSuffix(url, "/services/collector/event") {
		url += "/services/collector/event"
	}
	return sendTestRequest(url, token, payload, true)
}

func testCribl(endpoint, token string, entry map[string]interface{}) error {
	entry["_time"] = float64(time.Now().Unix())
	return sendTestRequest(endpoint, token, entry, false)
}

func testWebhook(endpoint, token string, entry map[string]interface{}) error {
	return sendTestRequest(endpoint, token, []interface{}{entry}, false)
}

func sendTestRequest(url, token string, payload interface{}, isSplunk bool) error {
	// SSRF protection: validate URL doesn't target internal resources
	if err := utils.ValidateExternalURL(url); err != nil {
		return fmt.Errorf("invalid endpoint URL: %w", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		if isSplunk {
			req.Header.Set("Authorization", "Splunk "+token)
		} else {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned HTTP %d", resp.StatusCode)
	}

	return nil
}
