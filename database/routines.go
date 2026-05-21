package database

import (
	"fmt"
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

// deleteAllSince will delete a specific table's records based on a time using batches.
func deleteAllSince(table string, date time.Time) {
	formattedDate := database.FormatTime(date)
	for {
		sql := fmt.Sprintf("DELETE FROM %s WHERE created_at < '%s' LIMIT 5000", table, formattedDate)
		if database.DbType() == "postgres" || database.DbType() == "sqlite3" || database.DbType() == "sqlite" {
			sql = fmt.Sprintf("DELETE FROM %s WHERE id IN (SELECT id FROM %s WHERE created_at < '%s' LIMIT 5000)", table, table, formattedDate)
		}
		q := database.Exec(sql)
		if err := q.Error(); err != nil {
			log.WithField("query", sql).Errorln(err)
			break
		}
		if q.RowsAffected() == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}
