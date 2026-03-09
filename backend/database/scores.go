package database

import (
	"database/sql"
)

// AssignGroupStationScore inserts or updates a score for a group at a station
func AssignGroupStationScore(db *sql.DB, groupID int, stationID int, score int) error {
	// Check if score already exists
	var existingID int
	err := db.QueryRow("SELECT id FROM group_station_scores WHERE group_id = ? AND station_id = ?", groupID, stationID).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Insert new score
		_, err = db.Exec("INSERT INTO group_station_scores (group_id, station_id, score) VALUES (?, ?, ?)", groupID, stationID, score)
		return err
	} else if err != nil {
		return err
	} else {
		// Update existing score
		_, err = db.Exec("UPDATE group_station_scores SET score = ? WHERE id = ?", score, existingID)
		return err
	}
}
