package notifiers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/notifications"
	"github.com/statping-ng/statping-ng/types/notifier"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

var _ notifier.DigestNotifier = (*teams)(nil)

var _ notifier.Notifier = (*teams)(nil)

const teamsMethod = "teams"

type teams struct {
	*notifications.Notification
}

func (t *teams) Select() *notifications.Notification {
	return t.Notification
}

var Teams = &teams{
	&notifications.Notification{
		Method:      teamsMethod,
		Title:       "Microsoft Teams",
		Description: "Send notifications to Microsoft Teams channels using Workflows webhooks (Power Automate). Create an Incoming Webhook workflow in your Teams channel to get started.",
		Author:      "Statping-ng",
		AuthorUrl:   "https://github.com/statping-ng/statping-ng",
		Delay:       time.Duration(5 * time.Second),
		Icon:        "fab fa-microsoft",
		Limits:      30,
		Form: []notifications.NotificationForm{{
			Type:        "text",
			Title:       "Webhook URL",
			Placeholder: "https://prod-xx.westus.logic.azure.com:443/workflows/...",
			SmallText:   "Create an Incoming Webhook workflow in Teams: Channel Settings → Connectors → Workflows → Post to a channel when a webhook request is received",
			DbField:     "Host",
			Required:    true,
		}},
	},
}

// TeamsAdaptiveCard represents a Microsoft Teams Adaptive Card
type TeamsAdaptiveCard struct {
	Type        string                 `json:"type"`
	Attachments []TeamsCardAttachment  `json:"attachments"`
}

type TeamsCardAttachment struct {
	ContentType string      `json:"contentType"`
	ContentUrl  *string     `json:"contentUrl"`
	Content     interface{} `json:"content"`
}

type AdaptiveCardContent struct {
	Schema  string                   `json:"$schema"`
	Type    string                   `json:"type"`
	Version string                   `json:"version"`
	Body    []map[string]interface{} `json:"body"`
	Actions []map[string]interface{} `json:"actions,omitempty"`
}

func (t *teams) sendTeams(card *TeamsAdaptiveCard) (string, error) {
	data, err := json.Marshal(card)
	if err != nil {
		return "", err
	}

	// Use SafeHttpRequest to prevent SSRF/DNS rebinding
	resp, _, err := utils.SafeHttpRequest(t.Host.String, "POST", "application/json", nil, bytes.NewReader(data), time.Duration(15*time.Second))
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (t *teams) buildCard(title, message, color string, facts []map[string]interface{}, actions []map[string]interface{}) *TeamsAdaptiveCard {
	body := []map[string]interface{}{
		{
			"type":   "TextBlock",
			"size":   "Medium",
			"weight": "Bolder",
			"text":   title,
			"style":  "heading",
			"color":  color,
		},
		{
			"type":      "TextBlock",
			"text":      message,
			"wrap":      true,
			"separator": true,
		},
	}

	if len(facts) > 0 {
		body = append(body, map[string]interface{}{
			"type":      "FactSet",
			"facts":     facts,
			"separator": true,
		})
	}

	content := AdaptiveCardContent{
		Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
		Type:    "AdaptiveCard",
		Version: "1.4",
		Body:    body,
		Actions: actions,
	}

	return &TeamsAdaptiveCard{
		Type: "message",
		Attachments: []TeamsCardAttachment{{
			ContentType: "application/vnd.microsoft.card.adaptive",
			ContentUrl:  nil,
			Content:     content,
		}},
	}
}

func (t *teams) OnTest() (string, error) {
	card := t.buildCard(
		"Statping Test Notification",
		"This is a test message from Statping-ng. If you see this, your Teams integration is working correctly!",
		"Good",
		[]map[string]interface{}{
			{"title": "Status", "value": "Test Successful"},
			{"title": "Time", "value": time.Now().Format("2006-01-02 15:04:05")},
		},
		[]map[string]interface{}{
			{
				"type":  "Action.OpenUrl",
				"title": "Open Statping",
				"url":   core.App.Domain,
			},
		},
	)
	return t.sendTeams(card)
}

func (t *teams) OnFailure(srv *services.Service, f failures.Failure) (string, error) {
	card := t.buildCard(
		fmt.Sprintf("🔴 Service Offline: %s", srv.Name),
		fmt.Sprintf("The service **%s** is currently **offline** and not responding.", srv.Name),
		"Attention",
		[]map[string]interface{}{
			{"title": "Service", "value": srv.Name},
			{"title": "URL", "value": srv.Domain},
			{"title": "Error", "value": f.Issue},
			{"title": "Status Code", "value": fmt.Sprintf("%d", srv.LastStatusCode)},
			{"title": "Downtime", "value": srv.Downtime().Human()},
			{"title": "When", "value": f.CreatedAt.Format("2006-01-02 15:04:05")},
		},
		[]map[string]interface{}{
			{
				"type":  "Action.OpenUrl",
				"title": "View Service",
				"url":   fmt.Sprintf("%s/service/%d", core.App.Domain, srv.Id),
			},
		},
	)
	return t.sendTeams(card)
}

func (t *teams) OnSuccess(srv *services.Service) (string, error) {
	card := t.buildCard(
		fmt.Sprintf("🟢 Service Online: %s", srv.Name),
		fmt.Sprintf("The service **%s** is back **online** and responding normally.", srv.Name),
		"Good",
		[]map[string]interface{}{
			{"title": "Service", "value": srv.Name},
			{"title": "URL", "value": srv.Domain},
			{"title": "Response Time", "value": fmt.Sprintf("%.2f ms", float64(srv.Latency)/1000)},
			{"title": "Recovered At", "value": time.Now().Format("2006-01-02 15:04:05")},
		},
		[]map[string]interface{}{
			{
				"type":  "Action.OpenUrl",
				"title": "View Service",
				"url":   fmt.Sprintf("%s/service/%d", core.App.Domain, srv.Id),
			},
		},
	)
	return t.sendTeams(card)
}

func (t *teams) OnSave() (string, error) {
	return "", nil
}

func (t *teams) Valid(values notifications.Values) error {
	if values.Host == "" {
		return fmt.Errorf("webhook URL is required")
	}
	if !strings.HasPrefix(values.Host, "https://") {
		return fmt.Errorf("webhook URL must start with https://")
	}
	if err := utils.ValidateExternalURL(values.Host); err != nil {
		return fmt.Errorf("invalid Teams webhook URL: %w", err)
	}
	return nil
}

// OnDigest sends the daily digest to Teams
func (t *teams) OnDigest(data notifier.DigestData) (string, error) {
	// Build service summary facts
	var serviceFacts []map[string]interface{}
	for _, s := range data.ServiceSummary {
		statusIcon := "🟢"
		if s.Status == "Offline" {
			statusIcon = "🔴"
		}
		serviceFacts = append(serviceFacts, map[string]interface{}{
			"title": fmt.Sprintf("%s %s", statusIcon, s.Name),
			"value": fmt.Sprintf("%d failures", s.FailureCount),
		})
	}

	// Summary message
	var summaryParts []string
	summaryParts = append(summaryParts, fmt.Sprintf("**%d** total services", data.TotalServices))
	summaryParts = append(summaryParts, fmt.Sprintf("**%d** healthy", data.HealthyServices))
	if data.FailedServices > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("**%d** currently down", data.FailedServices))
	}

	summaryMessage := strings.Join(summaryParts, " | ")
	if !data.HasFailures {
		summaryMessage += "\n\n✅ All services healthy - no issues in the last 24 hours!"
	}

	body := []map[string]interface{}{
		{
			"type":   "TextBlock",
			"size":   "Large",
			"weight": "Bolder",
			"text":   fmt.Sprintf("📊 %s Daily Digest", data.AppName),
			"style":  "heading",
		},
		{
			"type": "TextBlock",
			"text": data.Period,
			"size": "Small",
			"isSubtle": true,
		},
		{
			"type":      "TextBlock",
			"text":      summaryMessage,
			"wrap":      true,
			"separator": true,
		},
	}

	// Add service issues if any
	if len(serviceFacts) > 0 {
		body = append(body, map[string]interface{}{
			"type":   "TextBlock",
			"text":   "**Service Issues (Last 24h)**",
			"weight": "Bolder",
			"separator": true,
		})
		body = append(body, map[string]interface{}{
			"type":  "FactSet",
			"facts": serviceFacts,
		})
	}

	// Add app errors summary if any
	if data.HasAppErrors {
		body = append(body, map[string]interface{}{
			"type":      "TextBlock",
			"text":      fmt.Sprintf("⚠️ **%d application errors** in the last 24 hours", len(data.AppErrors)),
			"color":     "Warning",
			"separator": true,
		})
	}

	content := AdaptiveCardContent{
		Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
		Type:    "AdaptiveCard",
		Version: "1.4",
		Body:    body,
		Actions: []map[string]interface{}{
			{
				"type":  "Action.OpenUrl",
				"title": "Open Dashboard",
				"url":   data.Domain,
			},
		},
	}

	card := &TeamsAdaptiveCard{
		Type: "message",
		Attachments: []TeamsCardAttachment{{
			ContentType: "application/vnd.microsoft.card.adaptive",
			ContentUrl:  nil,
			Content:     content,
		}},
	}

	return t.sendTeams(card)
}
