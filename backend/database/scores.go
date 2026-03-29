package database

import (
	"database/sql"
)

// AssignGroupStationScore inserts or updates a score for a group at a station.
// Uses a single atomic upsert exploiting the UNIQUE(group_id, station_id) constraint
// so there is no SELECT + INSERT/UPDATE race condition.
func AssignGroupStationScore(db *sql.DB, groupID int, stationID int, score int) error {
	_, err := db.Exec(`
		INSERT INTO group_station_scores (group_id, station_id, score)
		VALUES (?, ?, ?)
		ON CONFLICT(group_id, station_id) DO UPDATE SET score = excluded.score`,
		groupID, stationID, score)
	return err
}
