package database

import (
	"database/sql"
	"fmt"
	"strings"

	"experiment1/backend/models"

	_ "modernc.org/sqlite"
)

// TrimSpace removes leading and trailing whitespace
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// InitDatabase creates the SQLite database and tables
func InitDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", models.DbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create teilnehmer table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS ` + models.TableName + ` (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		teilnehmer_id INTEGER,
		name TEXT,
		ortsverband TEXT,
		age INTEGER,
		geschlecht TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	// Create gruppe table
	createGruppeTableSQL := `
	CREATE TABLE IF NOT EXISTS gruppe (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_id INTEGER,
		teilnehmer_id INTEGER,
		FOREIGN KEY (teilnehmer_id) REFERENCES teilnehmer(id)
	);`

	_, err = db.Exec(createGruppeTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create gruppe table: %w", err)
	}

	// Create rel_tn_grp relationship table
	createRelTableSQL := `
	CREATE TABLE IF NOT EXISTS rel_tn_grp (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		teilnehmer_id INTEGER UNIQUE NOT NULL,
		group_id INTEGER NOT NULL,
		FOREIGN KEY (teilnehmer_id) REFERENCES teilnehmer(teilnehmer_id),
		FOREIGN KEY (group_id) REFERENCES gruppe(group_id)
	);`

	_, err = db.Exec(createRelTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create rel_tn_grp table: %w", err)
	}

	// Create stations table
	createStationsTableSQL := `
	CREATE TABLE IF NOT EXISTS stations (
		station_id INTEGER PRIMARY KEY AUTOINCREMENT,
		station_name TEXT NOT NULL
	);`

	_, err = db.Exec(createStationsTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create stations table: %w", err)
	}

	// Create group_station_scores relation table
	createGroupStationScoresTableSQL := `
	CREATE TABLE IF NOT EXISTS group_station_scores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_id INTEGER NOT NULL,
		station_id INTEGER NOT NULL,
		score INTEGER,
		FOREIGN KEY (group_id) REFERENCES gruppe(group_id),
		FOREIGN KEY (station_id) REFERENCES stations(station_id),
		UNIQUE(group_id, station_id)
	);`

	_, err = db.Exec(createGroupStationScoresTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create group_station_scores table: %w", err)
	}

	// Clear existing data
	_, err = db.Exec("DELETE FROM " + models.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to clear table: %w", err)
	}

	_, err = db.Exec("DELETE FROM gruppe")
	if err != nil {
		return nil, fmt.Errorf("failed to clear gruppe table: %w", err)
	}

	_, err = db.Exec("DELETE FROM rel_tn_grp")
	if err != nil {
		return nil, fmt.Errorf("failed to clear rel_tn_grp table: %w", err)
	}

	_, err = db.Exec("DELETE FROM stations")
	if err != nil {
		return nil, fmt.Errorf("failed to clear stations table: %w", err)
	}

	_, err = db.Exec("DELETE FROM group_station_scores")
	if err != nil {
		return nil, fmt.Errorf("failed to clear group_station_scores table: %w", err)
	}

	return db, nil
}
