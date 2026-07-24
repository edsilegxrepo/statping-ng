package configs

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/source"
	"github.com/statping-ng/statping-ng/types/notifications"
	"github.com/statping-ng/statping-ng/utils"
	_ "gorm.io/driver/mysql"
	_ "gorm.io/driver/postgres"

	"github.com/statping-ng/statping-ng/types/checkins"
	"github.com/statping-ng/statping-ng/types/core"
	"github.com/statping-ng/statping-ng/types/failures"
	"github.com/statping-ng/statping-ng/types/groups"
	"github.com/statping-ng/statping-ng/types/hits"
	"github.com/statping-ng/statping-ng/types/incidents"
	"github.com/statping-ng/statping-ng/types/messages"
	"github.com/statping-ng/statping-ng/types/services"
	"github.com/statping-ng/statping-ng/types/users"
)

func (d *DbConfig) ResetCore() error {
	if d.Db.HasTable("core") {
		return nil
	}
	var srvs int64
	if d.Db.HasTable(&services.Service{}) {
		d.Db.Model(&services.Service{}).Count(&srvs)
		if srvs > 0 {
			return errors.New("there are already services setup.")
		}
	}
	if err := d.DropDatabase(); err != nil {
		return errors.Wrap(err, "error dropping database")
	}
	if err := d.CreateDatabase(); err != nil {
		return errors.Wrap(err, "error creating database")
	}
	// Only collect secrets if they were randomly generated (not provided in config/env)
	allSecrets := make(map[string]string)
	if utils.Params.GetString("ADMIN_PASSWORD") == "" {
		adminPass, err := CreateAdminUser()
		if err != nil {
			return errors.Wrap(err, "error creating default admin user")
		}
		allSecrets[utils.Params.GetString("ADMIN_USER")] = adminPass
	} else {
		if _, err := CreateAdminUser(); err != nil {
			return errors.Wrap(err, "error creating default admin user")
		}
	}

	if err := core.Samples(); err != nil {
		return errors.Wrap(err, "error added core details")
	}

	if utils.Params.GetBool("SAMPLE_DATA") {
		log.Infoln("Adding Sample Data")
		creds, err := TriggerSamples()
		if err != nil {
			return errors.Wrap(err, "error adding sample data")
		}
		// Always collect sample data passwords as they are always random
		for k, v := range creds {
			allSecrets[k] = v
		}
		utils.Params.Set("SAMPLE_DATA", false)
	}

	if len(allSecrets) > 0 {
		if err := SaveSecrets(allSecrets); err != nil {
			log.Errorf("failed to save statping.secrets: %v", err)
		}
	}
	return nil
}

func (d *DbConfig) DatabaseChanges() error {
	var cr core.Core
	d.Db.Model(&core.Core{}).Find(&cr)

	if latestMigration > cr.MigrationId {
		log.Infof("Statping database is out of date, migrating to: %d", latestMigration)

		switch d.Db.DbType() {
		case "mysql":
			if err := d.genericMigration("MODIFY", false); err != nil {
				return err
			}
		case "postgres":
			if err := d.genericMigration("ALTER", true); err != nil {
				return err
			}
		default:
			if err := d.sqliteMigration(); err != nil {
				return err
			}
		}

		if err := d.Db.Exec("UPDATE core SET migration_id = ?", latestMigration).Error(); err != nil {
			return err
		}

		if err := d.BackupAssets(); err != nil {
			return err
		}
	}
	return nil
}

// BackupAssets is a temporary function (to version 0.90.*) to backup your customized theme
// to a new folder called 'assets_backup'.
func (d *DbConfig) BackupAssets() error {
	if source.UsingAssets(utils.Directory) {
		log.Infof("Backing up 'assets' folder to 'assets_backup'")
		if err := utils.RenameDirectory(utils.Directory+"/assets", utils.Directory+"/assets_backup"); err != nil {
			return err
		}
		log.Infof("%s", "Old assets are now stored in: "+utils.Directory+"/assets_backup")
	}
	return nil
}

// MigrateDatabase will migrate the database structure to current version.
// This function will NOT remove previous records, tables or columns from the database.
// If this function has an issue, it will ROLLBACK to the previous state.
func (d *DbConfig) MigrateDatabase() error {
	DbModels := []interface{}{&services.Service{}, &users.User{}, &hits.Hit{}, &failures.Failure{}, &messages.Message{}, &groups.Group{}, &checkins.Checkin{}, &checkins.CheckinHit{}, &notifications.Notification{}, &incidents.Incident{}, &incidents.IncidentUpdate{}}

	log.Infoln("Migrating Database Tables...")
	tx := d.Db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	for _, table := range DbModels {
		if err := tx.AutoMigrate(table).Error(); err != nil {
			tx.Rollback()
			log.Errorln(err)
			return err
		}
	}

	log.Infof("Migrating App to version: %s (%s)", utils.Params.GetString("VERSION"), utils.Params.GetString("COMMIT"))
	if err := tx.Table("core").AutoMigrate(&core.Core{}).Error(); err != nil {
		tx.Rollback()
		log.Errorln(fmt.Sprintf("Statping Database could not be migrated: %v", err))
		return err
	}

	if err := tx.Commit().Error(); err != nil {
		return err
	}

	d.Db.Table("core").Model(&core.Core{}).Where("migration_id >= 0").Update("version", utils.Params.GetString("VERSION"))

	log.Infoln("Statping Database Tables Migrated")

	if !d.Db.HasIndex(&hits.Hit{}, "idx_service_hit_created_at") {
		if err := d.Db.Model(&hits.Hit{}).AddIndex("idx_service_hit_created_at", "service", "created_at").Error(); err != nil {
			log.Errorln(err)
		}
	}

	if !d.Db.HasIndex(&failures.Failure{}, "idx_service_fail_created_at") {
		if err := d.Db.Model(&failures.Failure{}).AddIndex("idx_service_fail_created_at", "service", "created_at").Error(); err != nil {
			log.Errorln(err)
		}
	}

	if !d.Db.HasIndex(&checkins.CheckinHit{}, "idx_checkin_hit_created_at") {
		if err := d.Db.Model(&checkins.CheckinHit{}).AddIndex("idx_checkin_hit_created_at", "checkin", "created_at").Error(); err != nil {
			log.Errorln(err)
		}
	}

	// High-volume created_at indexes for maintenance cleanup performance
	if !d.Db.HasIndex(&hits.Hit{}, "idx_hits_created_at") {
		if err := d.Db.Model(&hits.Hit{}).AddIndex("idx_hits_created_at", "created_at").Error(); err != nil {
			log.Errorln(err)
		}
	}
	if !d.Db.HasIndex(&failures.Failure{}, "idx_failures_created_at") {
		if err := d.Db.Model(&failures.Failure{}).AddIndex("idx_failures_created_at", "created_at").Error(); err != nil {
			log.Errorln(err)
		}
	}
	if !d.Db.HasIndex(&checkins.CheckinHit{}, "idx_checkin_hits_created_at") {
		if err := d.Db.Model(&checkins.CheckinHit{}).AddIndex("idx_checkin_hits_created_at", "created_at").Error(); err != nil {
			log.Errorln(err)
		}
	}
	log.Infoln("Database Indexes Created")

	if database.DbType() == "postgres" || database.DbType() == "mysql" {
		log.Infoln("Adding Foreign Key Constraints...")
		if !d.Db.HasConstraint(&hits.Hit{}, "fk_hits_service") {
			d.Db.Exec("ALTER TABLE hits ADD CONSTRAINT fk_hits_service FOREIGN KEY (service) REFERENCES services(id) ON DELETE CASCADE")
		}
		if !d.Db.HasConstraint(&failures.Failure{}, "fk_failures_service") {
			d.Db.Exec("ALTER TABLE failures ADD CONSTRAINT fk_failures_service FOREIGN KEY (service) REFERENCES services(id) ON DELETE CASCADE")
		}
		if !d.Db.HasConstraint(&failures.Failure{}, "fk_failures_checkin") {
			d.Db.Exec("ALTER TABLE failures ADD CONSTRAINT fk_failures_checkin FOREIGN KEY (checkin) REFERENCES checkins(id) ON DELETE CASCADE")
		}
		if !d.Db.HasConstraint(&checkins.CheckinHit{}, "fk_checkin_hits_checkin") {
			d.Db.Exec("ALTER TABLE checkin_hits ADD CONSTRAINT fk_checkin_hits_checkin FOREIGN KEY (checkin) REFERENCES checkins(id) ON DELETE CASCADE")
		}
		if !d.Db.HasConstraint(&checkins.Checkin{}, "fk_checkins_service") {
			d.Db.Exec("ALTER TABLE checkins ADD CONSTRAINT fk_checkins_service FOREIGN KEY (service) REFERENCES services(id) ON DELETE CASCADE")
		}
		if !d.Db.HasConstraint(&incidents.Incident{}, "fk_incidents_service") {
			d.Db.Exec("ALTER TABLE incidents ADD CONSTRAINT fk_incidents_service FOREIGN KEY (service) REFERENCES services(id) ON DELETE CASCADE")
		}
		if !d.Db.HasConstraint(&incidents.IncidentUpdate{}, "fk_incidents_updates_incident") {
			d.Db.Exec("ALTER TABLE incident_updates ADD CONSTRAINT fk_incidents_updates_incident FOREIGN KEY (incident) REFERENCES incidents(id) ON DELETE CASCADE")
		}
		if !d.Db.HasConstraint(&messages.Message{}, "fk_messages_service") {
			d.Db.Exec("ALTER TABLE messages ADD CONSTRAINT fk_messages_service FOREIGN KEY (service) REFERENCES services(id) ON DELETE CASCADE")
		}
	}

	return nil
}
