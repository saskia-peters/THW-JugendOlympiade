package database

import (
	"database/sql"
	"fmt"
	"strings"

	"THW-JugendOlympiade/backend/models"

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

	if _, err = db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Create teilnehmende table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS teilnehmende (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		teilnehmer_id INTEGER UNIQUE,
		name TEXT,
		ortsverband TEXT,
		age INTEGER,
		geschlecht TEXT,
		pregroup TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	// Create gruppe table
	createGruppeTableSQL := `
	CREATE TABLE IF NOT EXISTS gruppe (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_id INTEGER NOT NULL,
		teilnehmer_id INTEGER UNIQUE NOT NULL,
		FOREIGN KEY (teilnehmer_id) REFERENCES teilnehmende(teilnehmer_id)
	);`

	_, err = db.Exec(createGruppeTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create gruppe table: %w", err)
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
	// Note: group_id has no FK constraint because gruppe.group_id is not unique
	// (multiple participants share the same group_id); integrity is enforced at application level
	createGroupStationScoresTableSQL := `
	CREATE TABLE IF NOT EXISTS group_station_scores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_id INTEGER NOT NULL,
		station_id INTEGER NOT NULL,
		score INTEGER,
		FOREIGN KEY (station_id) REFERENCES stations(station_id),
		UNIQUE(group_id, station_id)
	);`

	_, err = db.Exec(createGroupStationScoresTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create group_station_scores table: %w", err)
	}

	// Create betreuende table
	createBetreuerTableSQL := `
	CREATE TABLE IF NOT EXISTS betreuende (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		ortsverband TEXT
	);`
	_, err = db.Exec(createBetreuerTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create betreuende table: %w", err)
	}

	// Create gruppe_betreuende relation table
	createGruppeBetreuendeSQL := `
	CREATE TABLE IF NOT EXISTS gruppe_betreuende (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_id INTEGER NOT NULL,
		betreuende_id INTEGER NOT NULL,
		FOREIGN KEY (betreuende_id) REFERENCES betreuende(id),
		UNIQUE(group_id, betreuende_id)
	);`
	_, err = db.Exec(createGruppeBetreuendeSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create gruppe_betreuende table: %w", err)
	}

	// Clear existing data
	_, err = db.Exec("DELETE FROM teilnehmende")
	if err != nil {
		return nil, fmt.Errorf("failed to clear table: %w", err)
	}

	_, err = db.Exec("DELETE FROM gruppe")
	if err != nil {
		return nil, fmt.Errorf("failed to clear gruppe table: %w", err)
	}

	_, err = db.Exec("DELETE FROM stations")
	if err != nil {
		return nil, fmt.Errorf("failed to clear stations table: %w", err)
	}

	_, err = db.Exec("DELETE FROM group_station_scores")
	if err != nil {
		return nil, fmt.Errorf("failed to clear group_station_scores table: %w", err)
	}

	_, err = db.Exec("DELETE FROM gruppe_betreuende")
	if err != nil {
		return nil, fmt.Errorf("failed to clear gruppe_betreuende table: %w", err)
	}

	_, err = db.Exec("DELETE FROM betreuende")
	if err != nil {
		return nil, fmt.Errorf("failed to clear betreuende table: %w", err)
	}

	// Create indexes for better query performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_gruppe_group_id ON gruppe(group_id)",
		"CREATE INDEX IF NOT EXISTS idx_gruppe_teilnehmer_id ON gruppe(teilnehmer_id)",
		"CREATE INDEX IF NOT EXISTS idx_scores_group_id ON group_station_scores(group_id)",
		"CREATE INDEX IF NOT EXISTS idx_scores_station_id ON group_station_scores(station_id)",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			return nil, fmt.Errorf("failed to create index: %w", err)
		}
	}

	return db, nil
}
