package core

import (
	"testing"

	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	_ = utils.InitLogs()

	t.Run("creates new core with version and commit", func(t *testing.T) {
		New("1.0.0", "abc123")

		assert.NotNil(t, App)
		assert.Equal(t, "1.0.0", App.Version)
		assert.Equal(t, "abc123", App.Commit)
		assert.False(t, App.Started.IsZero())
	})
}

func TestCore_TableName(t *testing.T) {
	c := Core{}
	assert.Equal(t, "core", c.TableName())
}

func TestExample(t *testing.T) {
	_ = utils.InitLogs()

	t.Run("creates example core", func(t *testing.T) {
		core := Example()

		assert.NotNil(t, App)
		assert.NotNil(t, core)
		assert.Equal(t, "Statping Testing", App.Name)
		assert.Equal(t, "exampleapisecret", App.ApiSecret)
		assert.Equal(t, "http://localhost:8080", App.Domain)
	})
}
