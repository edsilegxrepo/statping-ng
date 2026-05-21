package incidents

import (
	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/errors"
	"github.com/statping-ng/statping-ng/types/metrics"
	"github.com/statping-ng/statping-ng/utils"
	"gorm.io/gorm"
)

var (
	db       database.Database
	dbUpdate database.Database
	log      = utils.Log.WithField("type", "service")
)

func SetDB(database database.Database) {
	db = database
	dbUpdate = database
}

func (i *Incident) Validate() error {
	if i.Title == "" {
		return errors.New("missing title")
	}
	return nil
}

func (i *Incident) BeforeUpdate(tx *gorm.DB) (err error) {
	return i.Validate()
}

func (i *Incident) BeforeCreate(tx *gorm.DB) (err error) {
	return i.Validate()
}

func (i *Incident) AfterFind(tx *gorm.DB) (err error) {
	tx.Where("incident = ?", i.Id).Order("id DESC").Find(&i.Updates)
	metrics.Query("incident", "find")
	return nil
}

func (i *Incident) AfterCreate(tx *gorm.DB) (err error) {
	metrics.Query("incident", "create")
	return nil
}

func (i *Incident) AfterUpdate(tx *gorm.DB) (err error) {
	metrics.Query("incident", "update")
	return nil
}

func (i *Incident) AfterDelete(tx *gorm.DB) (err error) {
	metrics.Query("incident", "delete")
	return nil
}

func (i *IncidentUpdate) Validate() error {
	if i.Message == "" {
		return errors.New("missing incident update title")
	}
	return nil
}

func (i *IncidentUpdate) BeforeUpdate(tx *gorm.DB) (err error) {
	return i.Validate()
}

func (i *IncidentUpdate) BeforeCreate(tx *gorm.DB) (err error) {
	return i.Validate()
}

func (i *IncidentUpdate) AfterFind(tx *gorm.DB) (err error) {
	metrics.Query("incident_update", "find")
	return nil
}

func (i *IncidentUpdate) AfterCreate(tx *gorm.DB) (err error) {
	metrics.Query("incident_update", "create")
	return nil
}

func (i *IncidentUpdate) AfterUpdate(tx *gorm.DB) (err error) {
	metrics.Query("incident_update", "update")
	return nil
}

func (i *IncidentUpdate) AfterDelete(tx *gorm.DB) (err error) {
	metrics.Query("incident_update", "delete")
	return nil
}

func FindUpdate(uid int64) (*IncidentUpdate, error) {
	var update IncidentUpdate
	q := dbUpdate.Where("id = ?", uid).First(&update)
	return &update, q.Error()
}

func Find(id int64) (*Incident, error) {
	var incident Incident
	q := db.Where("id = ?", id).First(&incident)
	return &incident, q.Error()
}

func FindByService(id int64) []*Incident {
	var incidents []*Incident
	db.Where("service = ?", id).Find(&incidents)
	return incidents
}

func All() []*Incident {
	var incidents []*Incident
	db.Find(&incidents)
	return incidents
}

func (i *Incident) Create() error {
	return db.Create(i).Error()
}

func (i *Incident) Update() error {
	return db.Update(i).Error()
}

func (i *Incident) Delete() error {
	for _, u := range i.Updates {
		if err := u.Delete(); err != nil {
			return err
		}
	}
	return db.Delete(i).Error()
}
