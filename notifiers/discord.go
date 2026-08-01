package notifiers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/notifications"
	"github.com/statping-ng/statping-ng/types/notifier"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

var (
	_ notifier.Notifier       = (*discord)(nil)
	_ notifier.DigestNotifier = (*discord)(nil)
)

type discord struct {
	*notifications.Notification
}

var Discorder = &discord{
	&notifications.Notification{
		Method:      "discord",
		Title:       "Discord",
		Description: "Send notifications to your discord channel using discord webhooks. Insert your discord channel Webhook URL to receive notifications. Based on the <a href=\"https://discordapp.com/developers/docs/resources/Webhook\">discord webhooker API</a>.",
		Author:      "Hunter Long",
		AuthorUrl:   "https://github.com/hunterlong",
		Delay:       time.Duration(5 * time.Second),
		Icon:        "fab fa-discord",
		SuccessData: null.NewNullString(`{"content": "Your service '{{.Service.Name}}' is currently back online and was down for {{.Service.Downtime.Human}}."}`),
		FailureData: null.NewNullString(`{"content": "Your service '{{.Service.Name}}' is has been failing for {{.Service.Downtime.Human}}! Reason: {{.Failure.Issue}}"}`),
		DataType:    "json",
		Limits:      60,
		Form: []notifications.NotificationForm{{
			Type:        "text",
			Title:       "discord webhooker URL",
			Placeholder: "https://discordapp.com/api/webhooks/****/*****",
			DbField:     "host",
		}},
	},
}

// Send will send a HTTP Post to the discord API. It accepts type: []byte
func (d *discord) sendRequest(msg string) (string, error) {
	// Use SafeHttpRequest to prevent SSRF/DNS rebinding
	out, _, err := utils.SafeHttpRequest(d.Host.String, "POST", "application/json", nil, strings.NewReader(msg), time.Duration(10*time.Second))
	return string(out), err
}

func (d *discord) Select() *notifications.Notification {
	return d.Notification
}

func (d *discord) Valid(values notifications.Values) error {
	if values.Host != "" {
		if err := utils.ValidateExternalURL(values.Host); err != nil {
			return fmt.Errorf("invalid Discord webhook URL: %w", err)
		}
	}
	return nil
}

// OnFailure will trigger failing service
func (d *discord) OnFailure(s *services.Service, f failures.Failure) (string, error) {
	out, err := d.sendRequest(ReplaceVars(d.FailureData.String, s, f))
	return out, err
}

// OnSuccess will trigger successful service
func (d *discord) OnSuccess(s *services.Service) (string, error) {
	out, err := d.sendRequest(ReplaceVars(d.SuccessData.String, s, failures.Failure{}))
	return out, err
}

// OnSave triggers when this notifier has been saved
func (d *discord) OnTest() (string, error) {
	outError := errors.New("incorrect discord URL, please confirm URL is correct")
	message := `{"content": "Testing the discord notifier"}`
	// Use SafeHttpRequest to prevent SSRF/DNS rebinding
	contents, _, err := utils.SafeHttpRequest(Discorder.Host.String, "POST", "application/json", nil, bytes.NewBuffer([]byte(message)), time.Duration(10*time.Second))
	if string(contents) == "" {
		return "", nil
	}
	var dtt discordTestJson
	if err != nil {
		return "", err
	}
	if err = json.Unmarshal(contents, &dtt); err != nil {
		return string(contents), outError
	}
	if dtt.Code == 0 {
		return string(contents), outError
	}
	return string(contents), nil
}

// OnSave will trigger when this notifier is saved
func (d *discord) OnSave() (string, error) {
	return "", nil
}

type discordTestJson struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// OnDigest sends the daily digest to Discord
func (d *discord) OnDigest(data notifier.DigestData) (string, error) {
	// Build Discord embed
	var fields []map[string]interface{}

	fields = append(fields, map[string]interface{}{
		"name":   "Total Services",
		"value":  fmt.Sprintf("%d", data.TotalServices),
		"inline": true,
	})
	fields = append(fields, map[string]interface{}{
		"name":   "Healthy",
		"value":  fmt.Sprintf("%d", data.HealthyServices),
		"inline": true,
	})
	fields = append(fields, map[string]interface{}{
		"name":   "Currently Down",
		"value":  fmt.Sprintf("%d", data.FailedServices),
		"inline": true,
	})

	// Add service issues
	if data.HasFailures {
		var issueLines []string
		for _, svc := range data.ServiceSummary {
			icon := ":white_check_mark:"
			if svc.Status == "Offline" {
				icon = ":x:"
			}
			issueLines = append(issueLines, fmt.Sprintf("%s **%s** - %d failures", icon, svc.Name, svc.FailureCount))
		}
		fields = append(fields, map[string]interface{}{
			"name":  "Service Issues (Last 24h)",
			"value": strings.Join(issueLines, "\n"),
		})
	}

	// Add app errors count
	if data.HasAppErrors {
		fields = append(fields, map[string]interface{}{
			"name":  "Application Errors",
			"value": fmt.Sprintf(":warning: %d errors in the last 24 hours", len(data.AppErrors)),
		})
	}

	color := 0x28a745 // green
	if data.FailedServices > 0 {
		color = 0xdc3545 // red
	}

	description := ":tada: All services healthy - no issues in the last 24 hours!"
	if data.HasFailures {
		description = fmt.Sprintf(":warning: %d services had issues in the last 24 hours", len(data.ServiceSummary))
	}

	embed := map[string]interface{}{
		"title":       fmt.Sprintf(":bar_chart: %s Daily Digest", data.AppName),
		"description": description,
		"color":       color,
		"fields":      fields,
		"footer": map[string]interface{}{
			"text": data.Period,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{embed},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return d.sendRequest(string(jsonData))
}
