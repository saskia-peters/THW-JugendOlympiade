package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// initDatabase creates the SQLite database and table
func initDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create table with 4 columns
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS ` + tableName + ` (
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

	// Clear existing data
	_, err = db.Exec("DELETE FROM " + tableName)
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

	return db, nil
}

// insertData inserts rows from XLSX into the database
func insertData(db *sql.DB, rows [][]string) error {
	// Start transaction for better performance
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare insert statement
	stmt, err := tx.Prepare("INSERT INTO " + tableName + " (teilnehmer_id, name, ortsverband, age, geschlecht) VALUES (?, ?, ?, ?, ?)")
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
			name = row[0]
		}
		if len(row) > 1 {
			ortsverband = row[1]
		}
		if len(row) > 2 {
			alter = row[2]
		}
		if len(row) > 3 {
			geschlecht = row[3]
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

// getAllTeilnehmers reads all participants from the database
func getAllTeilnehmers(db *sql.DB) ([]Teilnehmer, error) {
	rows, err := db.Query("SELECT id, teilnehmer_id, name, ortsverband, age, geschlecht FROM " + tableName + " ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teilnehmers []Teilnehmer
	for rows.Next() {
		var t Teilnehmer
		var alter sql.NullInt64
		err := rows.Scan(&t.ID, &t.TeilnehmerID, &t.Name, &t.Ortsverband, &alter, &t.Geschlecht)
		if err != nil {
			return nil, err
		}
		if alter.Valid {
			t.Alter = int(alter.Int64)
		}
		teilnehmers = append(teilnehmers, t)
	}

	return teilnehmers, rows.Err()
}

// saveGroups saves groups and relationships to the database
func saveGroups(db *sql.DB, groups []Group) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

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

// getGroupsForReport retrieves all groups with their participants from the database
func getGroupsForReport(db *sql.DB) ([]Group, error) {
	// Get all group IDs
	groupRows, err := db.Query("SELECT DISTINCT group_id FROM gruppe ORDER BY group_id")
	if err != nil {
		return nil, err
	}
	defer groupRows.Close()

	var groupIDs []int
	for groupRows.Next() {
		var groupID int
		if err := groupRows.Scan(&groupID); err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, groupID)
	}

	// For each group, get all participants
	var groups []Group
	for _, groupID := range groupIDs {
		query := `
			SELECT t.id, t.teilnehmer_id, t.name, t.ortsverband, t.age, t.geschlecht
			FROM teilnehmer t
			INNER JOIN rel_tn_grp r ON t.teilnehmer_id = r.teilnehmer_id
			WHERE r.group_id = ?
			ORDER BY t.name
		`

		rows, err := db.Query(query, groupID)
		if err != nil {
			return nil, err
		}

		group := Group{
			GroupID:      groupID,
			Teilnehmers:  make([]Teilnehmer, 0),
			Ortsverbands: make(map[string]int),
			Geschlechts:  make(map[string]int),
		}

		for rows.Next() {
			var t Teilnehmer
			var alter sql.NullInt64
			err := rows.Scan(&t.ID, &t.TeilnehmerID, &t.Name, &t.Ortsverband, &alter, &t.Geschlecht)
			if err != nil {
				rows.Close()
				return nil, err
			}
			if alter.Valid {
				t.Alter = int(alter.Int64)
			}
			group.Teilnehmers = append(group.Teilnehmers, t)
		}
		rows.Close()

		groups = append(groups, group)
	}

	return groups, nil
}
