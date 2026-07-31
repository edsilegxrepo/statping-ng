package database

import (
	"time"

	"github.com/statping-ng/statping-ng/utils"

	_ "github.com/mattn/go-sqlite3"
	_ "gorm.io/driver/mysql"
	_ "gorm.io/driver/postgres"
)

var log = utils.Log.WithField("type", "database")

// Maintenance will automatically delete old records from 'failures' and 'hits'
// this function is currently set to delete records 365+ days old every 60 minutes
// env: REMOVE_AFTER - golang duration parsed time for deleting records older than REMOVE_AFTER duration from now
// env: CLEANUP_INTERVAL - golang duration parsed time for checking old records routine
func Maintenance() {
	dur := utils.Params.GetDuration("REMOVE_AFTER")
	interval := utils.Params.GetDuration("CLEANUP_INTERVAL")

	log.Infof("Database Cleanup runs every %s and will remove records older than %s", interval.String(), dur.String())
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		deleteAfter := utils.Now().Add(-dur)

		log.Infof("Deleting failures older than %s", deleteAfter.String())
		deleteAllSince("failures", deleteAfter)

		log.Infof("Deleting hits older than %s", deleteAfter.String())
		deleteAllSince("hits", deleteAfter)

		log.Infof("Deleting checkin hits older than %s", deleteAfter.String())
		deleteAllSince("checkin_hits", deleteAfter)
	}
}

// allowedTables is a whitelist of tables that can be cleaned up
var allowedTables = map[string]bool{
	"failures":     true,
	"hits":         true,
	"checkin_hits": true,
}

// deleteAllSince will delete a specific table's records based on a time using batches.
// Uses parameterized queries to prevent SQL injection.
func deleteAllSince(table string, date time.Time) {
	if !allowedTables[table] {
		log.Errorf("deleteAllSince: invalid table name %q", table)
		return
	}

	for {
		var rowsAffected int64
		var err error

		switch database.DbType() {
		case "postgres":
			result := database.Exec(
				"DELETE FROM "+table+" WHERE id IN (SELECT id FROM "+table+" WHERE created_at < $1 LIMIT 5000)",
				date,
			)
			rowsAffected = result.RowsAffected()
			err = result.Error()
		case "sqlite3", "sqlite":
			result := database.Exec(
				"DELETE FROM "+table+" WHERE id IN (SELECT id FROM "+table+" WHERE created_at < ? LIMIT 5000)",
				date,
			)
			rowsAffected = result.RowsAffected()
			err = result.Error()
		default: // mysql
			result := database.Exec(
				"DELETE FROM "+table+" WHERE created_at < ? LIMIT 5000",
				date,
			)
			rowsAffected = result.RowsAffected()
			err = result.Error()
		}

		if err != nil {
			log.WithField("table", table).Errorln(err)
			break
		}
		if rowsAffected == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}
