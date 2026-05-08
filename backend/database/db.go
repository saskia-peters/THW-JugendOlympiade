package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"THW-JugendOlympiade/backend/models"

	_ "modernc.org/sqlite"
)

// TrimSpace removes leading and trailing whitespace
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// OpenExistingDB opens an already-initialised database file (read/write, no DDL,
// no data wipe). Returns an error if the file does not exist or cannot be opened.
func OpenExistingDB() (*sql.DB, error) {
	if _, err := os.Stat(models.DbFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file not found: %s", models.DbFile)
	}
	db, err := sql.Open("sqlite", models.DbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open existing database: %w", err)
	}
	// Serialise all Wails-dispatched concurrent calls to prevent SQLITE_BUSY.
	db.SetMaxOpenConns(1)
	if _, err = db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	// Migrate older databases that pre-date the cargroup persistence tables.
	for _, ddl := range []string{
		`CREATE TABLE IF NOT EXISTS cargroup_groups (
			pool_id  INTEGER NOT NULL,
			group_id INTEGER NOT NULL,
			UNIQUE(group_id)
		)`,
		`CREATE TABLE IF NOT EXISTS cargroup_fahrzeuge (
			pool_id     INTEGER NOT NULL,
			fahrzeug_id INTEGER NOT NULL,
			UNIQUE(fahrzeug_id)
		)`,
	} {
		if _, err = db.Exec(ddl); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to migrate database schema: %w", err)
		}
	}
	return db, nil
}

// InitDatabase creates the SQLite database and tables
func InitDatabase() (retDB *sql.DB, retErr error) {
	db, err := sql.Open("sqlite", models.DbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer func() {
		if retErr != nil {
			db.Close()
		}
	}()
	// Serialise all Wails-dispatched concurrent calls to prevent SQLITE_BUSY.
	db.SetMaxOpenConns(1)

	// PRAGMA must be set outside a transaction (connection-scope setting).
	if _, err = db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Wrap all DDL + data-wipe in a single atomic transaction so that a
	// mid-way process kill never leaves the schema in a partial state.
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin init transaction: %w", err)
	}
	defer func() {
		if retErr != nil {
			tx.Rollback()
		}
	}()

	stmts := []string{
		// --- schema ---
		`CREATE TABLE IF NOT EXISTS teilnehmende (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			teilnehmer_id INTEGER UNIQUE,
			name TEXT,
			ortsverband TEXT,
			age INTEGER,
			geschlecht TEXT,
			pregroup TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS gruppe (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			group_id INTEGER NOT NULL,
			teilnehmer_id INTEGER UNIQUE NOT NULL,
			FOREIGN KEY (teilnehmer_id) REFERENCES teilnehmende(teilnehmer_id)
		)`,
		`CREATE TABLE IF NOT EXISTS stations (
			station_id INTEGER PRIMARY KEY AUTOINCREMENT,
			station_name TEXT NOT NULL
		)`,
		// Note: group_id has no FK constraint because gruppe.group_id is not unique
		// (multiple participants share the same group_id); integrity is enforced at application level.
		`CREATE TABLE IF NOT EXISTS group_station_scores (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			group_id INTEGER NOT NULL,
			station_id INTEGER NOT NULL,
			score INTEGER,
			FOREIGN KEY (station_id) REFERENCES stations(station_id),
			UNIQUE(group_id, station_id)
		)`,
		`CREATE TABLE IF NOT EXISTS betreuende (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			ortsverband TEXT,
			fahrerlaubnis INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS gruppe_betreuende (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			group_id INTEGER NOT NULL,
			betreuende_id INTEGER NOT NULL,
			FOREIGN KEY (betreuende_id) REFERENCES betreuende(id),
			UNIQUE(group_id, betreuende_id)
		)`,
		`CREATE TABLE IF NOT EXISTS fahrzeuge (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bezeichnung TEXT NOT NULL,
			ortsverband TEXT,
			funkrufname TEXT,
			fahrer_name TEXT,
			sitzplaetze INTEGER NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS gruppe_fahrzeuge (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			group_id INTEGER NOT NULL,
			fahrzeug_id INTEGER NOT NULL,
			FOREIGN KEY (fahrzeug_id) REFERENCES fahrzeuge(id),
			UNIQUE(fahrzeug_id)
		)`,
		// CarGroups pool persistence: pool → groups and pool → vehicles.
		// pool_id matches models.CarGroup.ID.
		`CREATE TABLE IF NOT EXISTS cargroup_groups (
			pool_id  INTEGER NOT NULL,
			group_id INTEGER NOT NULL,
			UNIQUE(group_id)
		)`,
		`CREATE TABLE IF NOT EXISTS cargroup_fahrzeuge (
			pool_id     INTEGER NOT NULL,
			fahrzeug_id INTEGER NOT NULL,
			UNIQUE(fahrzeug_id)
		)`,
		// --- wipe existing data ---
		`DELETE FROM cargroup_fahrzeuge`,
		`DELETE FROM cargroup_groups`,
		`DELETE FROM gruppe_fahrzeuge`,
		`DELETE FROM gruppe_betreuende`,
		`DELETE FROM group_station_scores`,
		`DELETE FROM stations`,
		`DELETE FROM gruppe`,
		`DELETE FROM betreuende`,
		`DELETE FROM fahrzeuge`,
		`DELETE FROM teilnehmende`,
		// --- indexes ---
		`CREATE INDEX IF NOT EXISTS idx_gruppe_group_id ON gruppe(group_id)`,
		`CREATE INDEX IF NOT EXISTS idx_gruppe_teilnehmer_id ON gruppe(teilnehmer_id)`,
		`CREATE INDEX IF NOT EXISTS idx_scores_group_id ON group_station_scores(group_id)`,
		`CREATE INDEX IF NOT EXISTS idx_scores_station_id ON group_station_scores(station_id)`,
	}

	for _, stmt := range stmts {
		if _, err = tx.Exec(stmt); err != nil {
			return nil, fmt.Errorf("init statement failed: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit init transaction: %w", err)
	}

	return db, nil
}
