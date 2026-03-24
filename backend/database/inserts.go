package database

import (
	"database/sql"
	"fmt"
	"strings"

	"THW-JugendOlympiade/backend/models"
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
	stmt, err := tx.Prepare("INSERT INTO teilnehmende (teilnehmer_id, name, ortsverband, age, geschlecht, pregroup) VALUES (?, ?, ?, ?, ?, ?)")
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

		// Ensure we have exactly 5 columns (pad with empty strings if needed)
		name, ortsverband, alter, geschlecht, pregroup := "", "", "", "", ""
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
		if len(row) > 4 {
			pregroup = trimSpace(row[4])
		}

		// Skip empty rows (rows where name is empty or whitespace only)
		// This prevents accidentally inserting station names or other invalid data
		if name == "" {
			continue
		}

		// Use row number as teilnehmer_id (i represents row number)
		_, err = stmt.Exec(i, name, ortsverband, alter, geschlecht, pregroup)
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

// InsertBetreuende inserts caretaker rows from XLSX into the database
func InsertBetreuende(db *sql.DB, rows [][]string) error {
	if len(rows) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO betreuende (name, ortsverband, fahrerlaubnis) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}
		name, ortsverband := "", ""
		fahrerlaubnis := 0
		if len(row) > 0 {
			name = trimSpace(row[0])
		}
		if len(row) > 1 {
			ortsverband = trimSpace(row[1])
		}
		if len(row) > 2 && strings.EqualFold(trimSpace(row[2]), "ja") {
			fahrerlaubnis = 1
		}
		if name == "" {
			continue
		}
		if _, err = stmt.Exec(name, ortsverband, fahrerlaubnis); err != nil {
			return fmt.Errorf("failed to insert betreuende row %d: %w", i, err)
		}
	}

	return tx.Commit()
}

// SaveGroupBetreuende saves the betreuende-to-group assignments to the database
func SaveGroupBetreuende(db *sql.DB, groups []models.Group) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err = tx.Exec("DELETE FROM gruppe_betreuende"); err != nil {
		return fmt.Errorf("failed to clear gruppe_betreuende: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO gruppe_betreuende (group_id, betreuende_id) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare gruppe_betreuende statement: %w", err)
	}
	defer stmt.Close()

	for _, group := range groups {
		for _, b := range group.Betreuende {
			if _, err = stmt.Exec(group.GroupID, b.ID); err != nil {
				return fmt.Errorf("failed to insert gruppe_betreuende: %w", err)
			}
		}
	}

	return tx.Commit()
}

// SaveGroups saves groups and relationships to the database
func SaveGroups(db *sql.DB, groups []models.Group) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing groups first
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

	for _, group := range groups {
		for _, tn := range group.Teilnehmende {
			_, err = gruppeStmt.Exec(group.GroupID, tn.TeilnehmendeID)
			if err != nil {
				return fmt.Errorf("failed to insert into gruppe: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
