package configs

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/types/checkins"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/groups"
	"github.com/statping-ng/statping-ng/types/hits"
	"github.com/statping-ng/statping-ng/types/incidents"
	"github.com/statping-ng/statping-ng/types/messages"
	"github.com/statping-ng/statping-ng/types/notifications"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/types/users"
	"github.com/statping-ng/statping-ng/utils"
	"gopkg.in/yaml.v2"
)

type SamplerFunc func() error

type Sampler interface {
	Samples() []database.DbObject
}

func TriggerSamples() error {
	return createSamples(
		core.Samples,
		services.Samples,
		messages.Samples,
		checkins.Samples,
		checkins.SamplesChkHits,
		failures.Samples,
		groups.Samples,
		hits.Samples,
		incidents.Samples,
	)
}

func createSamples(sm ...SamplerFunc) error {
	for _, v := range sm {
		if err := v(); err != nil {
			return err
		}
	}
	return nil
}

// Migrate function
func (d *DbConfig) Update() error {
	var err error
	config, err := os.Create(utils.Directory + "/config.yml")
	if err != nil {
		return err
	}
	defer func() { _ = config.Close() }()

	data, err := yaml.Marshal(d) // #nosec G117
	if err != nil {
		log.Errorln(err)
		return err
	}
	if _, err := config.WriteString(string(data)); err != nil {
		return err
	}
	return nil
}

// DropDatabase will DROP each table Statping created
func (d *DbConfig) DropDatabase() error {
	DbModels := []interface{}{&services.Service{}, &users.User{}, &hits.Hit{}, &failures.Failure{}, &messages.Message{}, &groups.Group{}, &checkins.Checkin{}, &checkins.CheckinHit{}, &notifications.Notification{}, &incidents.Incident{}, &incidents.IncidentUpdate{}}
	log.Infoln("Dropping Database Tables...")
	for _, t := range DbModels {
		if err := d.Db.DropTableIfExists(t); err != nil {
			return err.Error()
		}
		log.Infof("Dropped table: %T\n", t)
	}
	return nil
}

func (d *DbConfig) Close() {
	if d == nil {
		return
	}
	if d.Db != nil {
		if err := d.Db.Close(); err != nil {
			log.Error(err)
		}
	}
}

// CreateDatabase will CREATE TABLES for each of the Statping elements
func (d *DbConfig) CreateDatabase() error {
	var err error

	DbModels := []interface{}{&services.Service{}, &users.User{}, &hits.Hit{}, &failures.Failure{}, &messages.Message{}, &groups.Group{}, &checkins.Checkin{}, &checkins.CheckinHit{}, &notifications.Notification{}, &incidents.Incident{}, &incidents.IncidentUpdate{}}

	log.Infoln("Creating Database Tables...")
	for _, table := range DbModels {
		if err := d.Db.CreateTable(table).Error(); err != nil {
			return errors.Wrap(err, fmt.Sprintf("error creating '%T' table", table))
		}
	}
	if err := d.Db.Table("core").CreateTable(&core.Core{}).Error(); err != nil {
		return errors.Wrap(err, "error creating 'core' table")
	}
	log.Infoln("Statping Database Created")

	return err
}
