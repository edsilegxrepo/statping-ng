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
	_ notifier.Notifier       = (*slack)(nil)
	_ notifier.DigestNotifier = (*slack)(nil)
)

const (
	slackMethod = "slack"
)

type slack struct {
	*notifications.Notification
}

func (s *slack) Select() *notifications.Notification {
	return s.Notification
}

var slacker = &slack{
	&notifications.Notification{
		Method:      slackMethod,
		Title:       "Slack",
		Description: "Send notifications to your slack channel when a service is offline. Insert your Incoming webhook URL for your channel to receive notifications. Based on the <a href=\"https://api.slack.com/incoming-webhooks\">Slack API</a>.",
		Author:      "Hunter Long",
		AuthorUrl:   "https://github.com/hunterlong",
		Delay:       time.Duration(10 * time.Second),
		Icon:        "fab fa-slack",
		SuccessData: null.NewNullString(`{ "blocks": [ { "type": "section", "text": { "type": "mrkdwn", "text": "The service {{.Service.Name}} is back online." } }, { "type": "actions", "elements": [ { "type": "button", "text": { "type": "plain_text", "text": "View Service", "emoji": true }, "style": "primary", "url": "{{.Core.Domain}}/service/{{.Service.Id}}" }, { "type": "button", "text": { "type": "plain_text", "text": "Go to Statping", "emoji": true }, "url": "{{.Core.Domain}}" } ] } ] }`),
		FailureData: null.NewNullString(`{ "blocks": [ { "type": "section", "text": { "type": "mrkdwn", "text": ":warning: The service {{.Service.Name}} is currently offline! :warning:" } }, { "type": "divider" }, { "type": "section", "fields": [ { "type": "mrkdwn", "text": "*Service:*\n{{.Service.Name}}" }, { "type": "mrkdwn", "text": "*URL:*\n{{.Service.Domain}}" }, { "type": "mrkdwn", "text": "*Status Code:*\n{{.Service.LastStatusCode}}" }, { "type": "mrkdwn", "text": "*When:*\n{{.Failure.CreatedAt}}" }, { "type": "mrkdwn", "text": "*Downtime:*\n{{.Service.Downtime.Human}}" }, { "type": "plain_text", "text": "*Error:*\n{{.Failure.Issue}}" } ] }, { "type": "divider" }, { "type": "actions", "elements": [ { "type": "button", "text": { "type": "plain_text", "text": "View Offline Service", "emoji": true }, "style": "danger", "url": "{{.Core.Domain}}/service/{{.Service.Id}}" }, { "type": "button", "text": { "type": "plain_text", "text": "Go to Statping", "emoji": true }, "url": "{{.Core.Domain}}" } ] } ] }`),
		DataType:    "json",
		RequestInfo: "Slack allows you to customize your own messages with many complex components. Checkout the <a target=\"_blank\" href=\"https://api.slack.com/reference/surfaces/formatting\">Slack Message API</a> to learn how you can create your own.",
		Limits:      60,
		Form: []notifications.NotificationForm{{
			Type:        "text",
			Title:       "Incoming Webhook Url",
			Placeholder: "https://hooks.slack.com/services/ETJ1B87WE/H76D6G8S30/H4d97R4EcZ40SpfyqPlAHr",
			SmallText:   "Incoming Webhook URL from <a href=\"https://api.slack.com/apps\" target=\"_blank\">Slack Apps</a>",
			DbField:     "Host",
			Required:    true,
		}},
	},
}

// Send will send a HTTP Post to the slack webhooker API. It accepts type: string
func (s *slack) sendSlack(msg string) (string, error) {
	// Use SafeHttpRequest to prevent SSRF/DNS rebinding
	resp, _, err := utils.SafeHttpRequest(s.Host.String, "POST", "application/json", nil, strings.NewReader(msg), time.Duration(10*time.Second))
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (s *slack) OnTest() (string, error) {
	example := services.Example(true)
	testMsg := ReplaceVars(s.SuccessData.String, example, failures.Failure{})
	// Use SafeHttpRequest to prevent SSRF/DNS rebinding
	contents, _, err := utils.SafeHttpRequest(s.Host.String, "POST", "application/json", nil, bytes.NewBuffer([]byte(testMsg)), time.Duration(10*time.Second))
	if err != nil {
		return "", err
	}
	if string(contents) != "ok" {
		return string(contents), errors.New("the slack response was incorrect, check the URL")
	}
	return string(contents), nil
}

// OnFailure will trigger failing service
func (s *slack) OnFailure(srv *services.Service, f failures.Failure) (string, error) {
	msg := ReplaceVars(s.FailureData.String, srv, f)
	out, err := s.sendSlack(msg)
	return out, err
}

// OnSuccess will trigger successful service
func (s *slack) OnSuccess(srv *services.Service) (string, error) {
	msg := ReplaceVars(s.SuccessData.String, srv, failures.Failure{})
	out, err := s.sendSlack(msg)
	return out, err
}

// OnSave will trigger when this notifier is saved
func (s *slack) OnSave() (string, error) {
	return "", nil
}

func (s *slack) Valid(values notifications.Values) error {
	if values.Host != "" {
		if err := utils.ValidateExternalURL(values.Host); err != nil {
			return fmt.Errorf("invalid Slack webhook URL: %w", err)
		}
	}
	return nil
}

// OnDigest sends the daily digest to Slack
func (s *slack) OnDigest(data notifier.DigestData) (string, error) {
	// Build blocks for Slack Block Kit
	blocks := []map[string]interface{}{
		{
			"type": "header",
			"text": map[string]interface{}{
				"type":  "plain_text",
				"text":  fmt.Sprintf(":bar_chart: %s Daily Digest", data.AppName),
				"emoji": true,
			},
		},
		{
			"type": "context",
			"elements": []map[string]interface{}{
				{
					"type": "plain_text",
					"text": data.Period,
				},
			},
		},
		{
			"type": "divider",
		},
		{
			"type": "section",
			"fields": []map[string]interface{}{
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Total Services*\n%d", data.TotalServices),
				},
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Healthy*\n:white_check_mark: %d", data.HealthyServices),
				},
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Currently Down*\n:x: %d", data.FailedServices),
				},
			},
		},
	}

	// Add service issues if any
	if data.HasFailures {
		blocks = append(blocks, map[string]interface{}{
			"type": "divider",
		})
		blocks = append(blocks, map[string]interface{}{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": "*Service Issues (Last 24h)*",
			},
		})

		for _, svc := range data.ServiceSummary {
			statusIcon := ":white_check_mark:"
			if svc.Status == "Offline" {
				statusIcon = ":x:"
			}
			blocks = append(blocks, map[string]interface{}{
				"type": "section",
				"text": map[string]interface{}{
					"type": "mrkdwn",
					"text": fmt.Sprintf("%s *%s* - %d failures", statusIcon, svc.Name, svc.FailureCount),
				},
			})
		}
	} else {
		blocks = append(blocks, map[string]interface{}{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": ":tada: *All services healthy* - no issues in the last 24 hours!",
			},
		})
	}

	// Add app errors summary if any
	if data.HasAppErrors {
		blocks = append(blocks, map[string]interface{}{
			"type": "divider",
		})
		blocks = append(blocks, map[string]interface{}{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": fmt.Sprintf(":warning: *%d application errors* in the last 24 hours", len(data.AppErrors)),
			},
		})
	}

	// Add action button
	blocks = append(blocks, map[string]interface{}{
		"type": "divider",
	})
	blocks = append(blocks, map[string]interface{}{
		"type": "actions",
		"elements": []map[string]interface{}{
			{
				"type": "button",
				"text": map[string]interface{}{
					"type": "plain_text",
					"text": "Open Dashboard",
				},
				"url": data.Domain,
			},
		},
	})

	payload := map[string]interface{}{
		"blocks": blocks,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return s.sendSlack(string(jsonData))
}
