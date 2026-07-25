package hits

import (
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	_ = utils.InitLogs()
	m.Run()
}

func TestHit_BeforeCreate(t *testing.T) {
	t.Run("sets CreatedAt when zero", func(t *testing.T) {
		h := &Hit{
			Service:  1,
			Latency:  100,
			PingTime: 50,
		}
		err := h.BeforeCreate(&gorm.DB{})
		assert.Nil(t, err)
		assert.False(t, h.CreatedAt.IsZero())
		assert.True(t, time.Since(h.CreatedAt) < time.Second)
	})

	t.Run("preserves existing CreatedAt", func(t *testing.T) {
		existingTime := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
		h := &Hit{
			Service:   1,
			Latency:   100,
			PingTime:  50,
			CreatedAt: existingTime,
		}
		err := h.BeforeCreate(&gorm.DB{})
		assert.Nil(t, err)
		assert.Equal(t, existingTime, h.CreatedAt)
	})
}

func TestHit_CRUD(t *testing.T) {
	testDb, err := database.OpenTester()
	require.NoError(t, err)
	SetDB(testDb)

	if migrateErr := testDb.AutoMigrate(&Hit{}).Error(); migrateErr != nil {
		require.NoError(t, migrateErr)
	}

	t.Run("Create hit", func(t *testing.T) {
		h := &Hit{
			Service:  1,
			Latency:  150,
			PingTime: 75,
		}
		err := h.Create()
		assert.Nil(t, err)
		assert.True(t, h.Id > 0)
	})

	t.Run("Update hit", func(t *testing.T) {
		h := &Hit{
			Service:  1,
			Latency:  200,
			PingTime: 100,
		}
		err := h.Create()
		require.Nil(t, err)

		h.Latency = 250
		err = h.Update()
		assert.Nil(t, err)
	})

	t.Run("Delete hit", func(t *testing.T) {
		h := &Hit{
			Service:  1,
			Latency:  300,
			PingTime: 150,
		}
		err := h.Create()
		require.Nil(t, err)

		err = h.Delete()
		assert.Nil(t, err)
	})
}
