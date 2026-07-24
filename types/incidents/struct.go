package incidents

import "time"

// Incident is the main struct for Incidents
type Incident struct {
	Id          int64             `gorm:"primary_key;column:id" json:"id"`
	Title       string            `gorm:"column:title" json:"title,omitempty"`
	Description string            `gorm:"column:description" json:"description,omitempty"`
	ServiceId   int64             `gorm:"index:idx_incidents_service_created_at,priority:1;index;column:service" json:"service"`
	CreatedAt   time.Time         `gorm:"index:idx_incidents_service_created_at,priority:2;column:created_at" json:"created_at"`
	UpdatedAt   time.Time         `gorm:"column:updated_at" json:"updated_at"`
	Updates     []*IncidentUpdate `gorm:"foreignKey:IncidentId;references:Id" json:"updates,omitempty"`
}

// IncidentUpdate contains updates based on a Incident
type IncidentUpdate struct {
	Id         int64     `gorm:"primary_key;column:id" json:"id"`
	IncidentId int64     `gorm:"index:idx_incident_updates_incident_created_at,priority:1;index;column:incident" json:"-"`
	Message    string    `gorm:"column:message" json:"message,omitempty"`
	Type       string    `gorm:"column:type" json:"type,omitempty"`
	CreatedAt  time.Time `gorm:"index:idx_incident_updates_incident_created_at,priority:2;column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}
