package hits

import (
	"time"

	"gorm.io/gorm"
)

// Hit struct is a 'successful' ping or web response entry for a service.
type Hit struct {
	Id        int64     `gorm:"primary_key;column:id" json:"id"`
	Service   int64     `gorm:"index:idx_hits_service_created_at,priority:1;index;column:service" json:"-"`
	Latency   int64     `gorm:"column:latency" json:"latency"`
	PingTime  int64     `gorm:"column:ping_time" json:"ping_time"`
	CreatedAt time.Time `gorm:"index:idx_hits_service_created_at,priority:2;column:created_at" json:"created_at"`
}

// BeforeCreate for Hit will set CreatedAt to UTC
func (h *Hit) BeforeCreate(tx *gorm.DB) (err error) {
	if h.CreatedAt.IsZero() {
		h.CreatedAt = time.Now().UTC()
	}
	return nil
}
