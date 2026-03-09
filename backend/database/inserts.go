package database

import (
	"database/sql"
	"fmt"
	"strings"

	"experiment1/backend/models"
)

// trimSpace removes leading, trailing, and internal excessive whitespace
func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

// InsertData inserts participant data from XLSX rows into the database
func InsertData(db *sql.DB, rows [][]string) error {
	// Start transaction for better performance
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare insert statement
	stmt, err := tx.Prepare("INSERT INTO " + models.TableName + " (teilnehmer_id, name, ortsverband, age, geschlecht) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Skip header row (first row) and insert data
	for i, row := range rows {
		if i == 0 {
			// Skip header row
			continue
		}

		// Ensure we have exactly 4 columns (pad with empty strings if needed)
		name, ortsverband, alter, geschlecht := "", "", "", ""
		if len(row) > 0 {
			name = trimSpace(row[0])
		}
		if len(row) > 1 {
			ortsverband = trimSpace(row[1])
		}
		if len(row) > 2 {
			alter = trimSpace(row[2])
		}
		if len(row) > 3 {
			geschlecht = trimSpace(row[3])
		}

		// Skip empty rows (rows where name is empty or whitespace only)
		// This prevents accidentally inserting station names or other invalid data
		if name == "" {
			continue
		}

		// Use row number as teilnehmer_id (i represents row number)
		_, err = stmt.Exec(i, name, ortsverband, alter, geschlecht)
		if err != nil {
			return fmt.Errorf("failed to insert row %d: %w", i, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// InsertStations inserts station rows from XLSX into the database
func InsertStations(db *sql.DB, rows [][]string) error {
	if len(rows) == 0 {
		return nil // No stations to insert
	}

	// Start transaction for better performance
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare insert statement
	stmt, err := tx.Prepare("INSERT INTO stations (station_name) VALUES (?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Skip header row (first row) and insert data
	for i, row := range rows {
		if i == 0 {
			// Skip header row
			continue
		}

		// Get station name from first column
		stationName := ""
		if len(row) > 0 {
			stationName = row[0]
		}

		// Trim whitespace and skip empty rows
		stationName = trimSpace(stationName)
		if stationName == "" {
			continue
		}

		_, err = stmt.Exec(stationName)
		if err != nil {
			return fmt.Errorf("failed to insert station row %d: %w", i, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SaveGroups saves groups and relationships to the database
func SaveGroups(db *sql.DB, groups []models.Group) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing groups first
	_, err = tx.Exec("DELETE FROM rel_tn_grp")
	if err != nil {
		return fmt.Errorf("failed to clear rel_tn_grp: %w", err)
	}

	_, err = tx.Exec("DELETE FROM gruppe")
	if err != nil {
		return fmt.Errorf("failed to clear gruppe: %w", err)
	}

	// Insert into gruppe table
	gruppeStmt, err := tx.Prepare("INSERT INTO gruppe (group_id, teilnehmer_id) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare gruppe statement: %w", err)
	}
	defer gruppeStmt.Close()

	// Insert into rel_tn_grp table
	relStmt, err := tx.Prepare("INSERT INTO rel_tn_grp (teilnehmer_id, group_id) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare rel_tn_grp statement: %w", err)
	}
	defer relStmt.Close()

	for _, group := range groups {
		for _, teilnehmer := range group.Teilnehmers {
			// Insert into gruppe
			_, err = gruppeStmt.Exec(group.GroupID, teilnehmer.TeilnehmerID)
			if err != nil {
				return fmt.Errorf("failed to insert into gruppe: %w", err)
			}

			// Insert into rel_tn_grp
			_, err = relStmt.Exec(teilnehmer.TeilnehmerID, group.GroupID)
			if err != nil {
				return fmt.Errorf("failed to insert into rel_tn_grp: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
