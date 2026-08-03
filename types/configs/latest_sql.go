package configs

import (
	"fmt"
	"os"

	"github.com/statping-ng/statping-ng/utils"
)

const latestMigration = 1722700000 // 2024-08-03: normalize usernames/emails to lowercase

func init() {
	_ = os.Setenv("MIGRATION_ID", utils.ToString(latestMigration))
}

func (d *DbConfig) genericMigration(alterStr string, isPostgres bool) error {
	var extra string
	extraType := "UNSIGNED INTEGER"
	if isPostgres {
		extra = " TYPE"
		extraType = "bigint"
	}
	if err := d.Db.Exec(fmt.Sprintf("ALTER TABLE hits %s COLUMN latency%s BIGINT;", alterStr, extra)).Error(); err != nil {
		return err
	}
	if err := d.Db.Exec(fmt.Sprintf("ALTER TABLE hits %s COLUMN ping_time%s BIGINT;", alterStr, extra)).Error(); err != nil {
		return err
	}
	if err := d.Db.Exec(fmt.Sprintf("ALTER TABLE failures %s COLUMN ping_time%s BIGINT;", alterStr, extra)).Error(); err != nil {
		return err
	}
	if err := d.Db.Exec(fmt.Sprintf("UPDATE hits SET latency = CAST(latency * 1000000 AS %s);", extraType)).Error(); err != nil {
		return err
	}
	if err := d.Db.Exec(fmt.Sprintf("UPDATE hits SET ping_time = CAST(ping_time * 1000000 AS %s);", extraType)).Error(); err != nil {
		return err
	}
	if err := d.Db.Exec(fmt.Sprintf("UPDATE failures SET ping_time = CAST(ping_time * 1000000 AS %s);", extraType)).Error(); err != nil {
		return err
	}
	return nil
}

// migrateAuthProvider sets auth_provider to 'local' for existing users where it's NULL or empty
func (d *DbConfig) migrateAuthProvider() error {
	log.Infoln("Migrating users: setting auth_provider to 'local' for existing users")
	if err := d.Db.Exec("UPDATE users SET auth_provider = 'local' WHERE auth_provider IS NULL OR auth_provider = ''").Error(); err != nil {
		return err
	}
	return nil
}

// migrateNormalizeUsernames normalizes existing usernames and emails to lowercase
func (d *DbConfig) migrateNormalizeUsernames() error {
	log.Infoln("Migrating users: normalizing usernames and emails to lowercase")
	if err := d.Db.Exec("UPDATE users SET username = LOWER(username), email = LOWER(email)").Error(); err != nil {
		return err
	}
	return nil
}

func (d *DbConfig) sqliteMigration() error {
	if err := d.Db.Exec(`ALTER TABLE hits RENAME TO hits_backup;`).Error(); err != nil {
		return err
	}
	if err := d.Db.Exec(`CREATE TABLE hits (id INTEGER PRIMARY KEY AUTOINCREMENT, service bigint, latency bigint, ping_time bigint, created_at datetime);`).Error(); err != nil {
		return err
	}
	if err := d.Db.Exec(`INSERT INTO hits (id, service, latency, ping_time, created_at) SELECT id, service, CAST(latency * 1000000 AS bigint), CAST(ping_time * 1000000 AS bigint), created_at FROM hits_backup;`).Error(); err != nil {
		return err
	}
	// failures table
	if err := d.Db.Exec(`ALTER TABLE failures RENAME TO failures_backup;`).Error(); err != nil {
		return err
	}
	if err := d.Db.Exec(`CREATE TABLE failures (id INTEGER PRIMARY KEY AUTOINCREMENT, issue varchar(255), method varchar(255), method_id bigint, service bigint, ping_time bigint, checkin bigint, error_code bigint, created_at datetime);`).Error(); err != nil {
		return err
	}
	if err := d.Db.Exec(`INSERT INTO failures (id, issue, method, method_id, service, ping_time, checkin, created_at) SELECT id, issue, method, method_id, service, CAST(ping_time * 1000000 AS bigint), checkin, created_at FROM failures_backup;`).Error(); err != nil {
		return err
	}
	if err := d.Db.Exec(`DROP TABLE hits_backup;`).Error(); err != nil {
		return err
	}
	if err := d.Db.Exec(`DROP TABLE failures_backup;`).Error(); err != nil {
		return err
	}
	return nil
}
