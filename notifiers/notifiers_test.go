package notifiers

import (
	"testing"

	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
)

func TestReplaceTemplate(t *testing.T) {
	t.Parallel()
	temp := `{"id":{{.Service.Id}},"name":"{{.Service.Name}}"}`
	replaced := ReplaceTemplate(temp, makeReplacer(services.Example(true), failures.Failure{}))
	assert.Equal(t, `{"id":6283,"name":"Statping Example"}`, replaced)

	temp = `{"id":{{.Service.Id}},"name":"{{.Service.Name}}","failure":"{{.Failure.Issue}}"}`
	replaced = ReplaceTemplate(temp, makeReplacer(services.Example(false), failures.Example()))
	assert.Equal(t, `{"id":6283,"name":"Statping Example","failure":"Response did not response a 200 status code"}`, replaced)
}

func TestReplaceTemplateInvalid(t *testing.T) {
	t.Parallel()
	temp := `{"id":{{.Invalid.Field}}`
	replaced := ReplaceTemplate(temp, makeReplacer(services.Example(true), failures.Failure{}))
	assert.Contains(t, replaced, "template")
}

func TestReplaceVars(t *testing.T) {
	_ = utils.InitLogs()
	core.Example()

	svc := services.Example(false)
	fail := failures.Example()

	result := ReplaceVars("Service {{.Service.Name}} has issue: {{.Failure.Issue}}", svc, fail)
	assert.Contains(t, result, "Statping Example")
	assert.Contains(t, result, "did not response")
}

func TestPushover_Select(t *testing.T) {
	tests := []struct {
		Value    string
		Expected string
	}{
		{
			"lowest",
			"-2",
		},
		{
			"low",
			"-1",
		},
		{
			"normal",
			"0",
		},
		{
			"high",
			"1",
		},
		{
			"emergency",
			"2",
		},
		{
			"",
			"0",
		},
	}

	for _, v := range tests {
		assert.Equal(t, v.Expected, priority(v.Value))
	}
}
