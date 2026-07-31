package notifiers

import (
	"testing"

	"github.com/statping-ng/statping-ng/types/notifier"
	"github.com/stretchr/testify/assert"
)

func TestTeamsNotifier(t *testing.T) {
	t.Run("implements Notifier interface", func(t *testing.T) {
		var _ notifier.Notifier = (*teams)(nil)
	})

	t.Run("implements DigestNotifier interface", func(t *testing.T) {
		var _ notifier.DigestNotifier = (*teams)(nil)
	})
}

func TestTeamsSelect(t *testing.T) {
	n := Teams.Select()
	assert.Equal(t, "teams", n.Method)
	assert.Equal(t, "Microsoft Teams", n.Title)
	assert.NotEmpty(t, n.Description)
	assert.Equal(t, "fab fa-microsoft", n.Icon)
}

func TestTeamsFormFields(t *testing.T) {
	n := Teams.Select()
	assert.Len(t, n.Form, 1, "Teams notifier should have 1 form field")

	// Webhook URL field
	assert.Equal(t, "text", n.Form[0].Type)
	assert.Equal(t, "Webhook URL", n.Form[0].Title)
	assert.Equal(t, "Host", n.Form[0].DbField)
	assert.True(t, n.Form[0].Required)
}

func TestTeamsMethod(t *testing.T) {
	assert.Equal(t, "teams", teamsMethod)
}

func TestTeamsAdaptiveCardStructure(t *testing.T) {
	// Test struct can be created
	card := TeamsAdaptiveCard{
		Type: "message",
		Attachments: []TeamsCardAttachment{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content: AdaptiveCardContent{
					Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
					Type:    "AdaptiveCard",
					Version: "1.2",
					Body:    []map[string]interface{}{},
				},
			},
		},
	}

	assert.Equal(t, "message", card.Type)
	assert.Len(t, card.Attachments, 1)
	assert.Equal(t, "application/vnd.microsoft.card.adaptive", card.Attachments[0].ContentType)
}

func TestAdaptiveCardContentStructure(t *testing.T) {
	content := AdaptiveCardContent{
		Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
		Type:    "AdaptiveCard",
		Version: "1.2",
		Body: []map[string]interface{}{
			{"type": "TextBlock", "text": "Test"},
		},
		Actions: []map[string]interface{}{
			{"type": "Action.OpenUrl", "title": "View", "url": "https://example.com"},
		},
	}

	assert.Equal(t, "AdaptiveCard", content.Type)
	assert.Equal(t, "1.2", content.Version)
	assert.Len(t, content.Body, 1)
	assert.Len(t, content.Actions, 1)
}
