package notifiers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/notifications"
	"github.com/statping-ng/statping-ng/types/notifier"
	"github.com/statping-ng/statping-ng/types/null"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
)

var _ notifier.Notifier = (*ntfyNotifier)(nil)

type ntfyNotifier struct {
	*notifications.Notification
}

func (n *ntfyNotifier) Select() *notifications.Notification {
	return n.Notification
}

func (n *ntfyNotifier) Valid(values notifications.Values) error {
	if values.Host != "" {
		if err := utils.ValidateExternalURL(values.Host); err != nil {
			return fmt.Errorf("invalid ntfy server URL: %w", err)
		}
	}
	return nil
}

var Ntfy = &ntfyNotifier{
	&notifications.Notification{
		Method:      "ntfy",
		Title:       "ntfy.sh",
		Description: "Send push notifications via <a href=\"https://ntfy.sh\">ntfy.sh</a> (self-hostable). Create a topic and subscribe on your phone.",
		Author:      "Statping-ng",
		AuthorUrl:   "https://github.com/statping-ng",
		Icon:        "fas fa-bell",
		Delay:       time.Duration(5 * time.Second),
		Limits:      60,
		SuccessData: null.NewNullString(`Service '{{.Service.Name}}' is back online`),
		FailureData: null.NewNullString(`Service '{{.Service.Name}}' is offline: {{.Failure.Issue}}`),
		DataType:    "text",
		Form: []notifications.NotificationForm{
			{
				Type:        "text",
				Title:       "Server URL",
				Placeholder: "https://ntfy.sh",
				SmallText:   "Default: https://ntfy.sh (or your self-hosted instance)",
				DbField:     "Host",
			},
			{
				Type:        "text",
				Title:       "Topic",
				Placeholder: "my-statping-alerts",
				SmallText:   "The topic name to publish to (subscribers need this)",
				DbField:     "Var1",
				Required:    true,
			},
			{
				Type:        "text",
				Title:       "Access Token",
				Placeholder: "tk_xxxxxxxxx",
				SmallText:   "Optional: for authenticated topics",
				DbField:     "api_key",
			},
			{
				Type:        "list",
				Title:       "Priority",
				Placeholder: "default",
				DbField:     "Var2",
				ListOptions: []string{"min", "low", "default", "high", "urgent"},
			},
		},
	},
}

type ntfyMessage struct {
	Topic    string   `json:"topic"`
	Title    string   `json:"title,omitempty"`
	Message  string   `json:"message"`
	Priority int      `json:"priority,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

func (n *ntfyNotifier) sendNtfy(title, message string, isFailure bool) error {
	server := strings.TrimSpace(n.Host.String)
	if server == "" {
		server = "https://ntfy.sh"
	}
	server = strings.TrimSuffix(server, "/")

	topic := strings.TrimSpace(n.Var1.String)
	if topic == "" {
		return fmt.Errorf("ntfy topic is required")
	}

	// Validate and resolve URL to prevent DNS rebinding attacks
	resolvedIPs, err := utils.ValidateAndResolveURL(server)
	if err != nil {
		return fmt.Errorf("invalid ntfy server URL: %w", err)
	}

	priority := priorityToInt(n.Var2.String)

	var tags []string
	if isFailure {
		tags = []string{"rotating_light", "warning"}
	} else {
		tags = []string{"white_check_mark"}
	}

	msg := ntfyMessage{
		Topic:    topic,
		Title:    title,
		Message:  message,
		Priority: priority,
		Tags:     tags,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", server, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Add auth token if provided
	if token := strings.TrimSpace(n.ApiKey.String); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Use safe client with pinned DNS to prevent rebinding
	client := utils.SafeHTTPClient(resolvedIPs, 10*time.Second)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ntfy returned status %d", resp.StatusCode)
	}

	return nil
}

func priorityToInt(p string) int {
	switch strings.ToLower(p) {
	case "min":
		return 1
	case "low":
		return 2
	case "default", "":
		return 3
	case "high":
		return 4
	case "urgent":
		return 5
	default:
		return 3
	}
}

func (n *ntfyNotifier) OnFailure(s *services.Service, f failures.Failure) (string, error) {
	msg := ReplaceTemplate(n.FailureData.String, makeReplacer(s, f))
	title := fmt.Sprintf("🔴 %s Offline", s.Name)
	return msg, n.sendNtfy(title, msg, true)
}

func (n *ntfyNotifier) OnSuccess(s *services.Service) (string, error) {
	msg := ReplaceTemplate(n.SuccessData.String, makeReplacer(s, failures.Failure{}))
	title := fmt.Sprintf("✅ %s Online", s.Name)
	return msg, n.sendNtfy(title, msg, false)
}

func (n *ntfyNotifier) OnTest() (string, error) {
	msg := "This is a test notification from Statping"
	return msg, n.sendNtfy("Statping Test", msg, false)
}

func (n *ntfyNotifier) OnSave() (string, error) {
	return "", nil
}
