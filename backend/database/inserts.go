package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
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
			if b.ID == 0 {
				// Skip synthetic entries (external drivers not in the Betreuende DB table).
				continue
			}
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

// InsertFahrzeuge inserts vehicle rows from XLSX into the database
func InsertFahrzeuge(db *sql.DB, rows [][]string) error {
	if len(rows) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO fahrzeuge (bezeichnung, ortsverband, funkrufname, fahrer_name, sitzplaetze) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}
		bezeichnung, ortsverband, funkrufname, fahrerName := "", "", "", ""
		sitzplaetze := 1
		if len(row) > 0 {
			bezeichnung = trimSpace(row[0])
		}
		if len(row) > 1 {
			ortsverband = trimSpace(row[1])
		}
		if len(row) > 2 {
			funkrufname = trimSpace(row[2])
		}
		if len(row) > 3 {
			fahrerName = trimSpace(row[3])
		}
		if len(row) > 4 {
			if s := trimSpace(row[4]); s != "" {
				n, err := strconv.Atoi(s)
				if err == nil && n >= 1 {
					sitzplaetze = n
				}
			}
		}
		if bezeichnung == "" {
			continue
		}
		if _, err = stmt.Exec(bezeichnung, ortsverband, funkrufname, fahrerName, sitzplaetze); err != nil {
			return fmt.Errorf("failed to insert fahrzeuge row %d: %w", i, err)
		}
	}

	return tx.Commit()
}

// SaveGroupFahrzeuge saves the vehicle-to-group assignments to the database
// and persists any FahrerName changes that were made in memory during
// distribution (e.g. the Phase 3b fallback driver assignment).
func SaveGroupFahrzeuge(db *sql.DB, groups []models.Group) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err = tx.Exec("DELETE FROM gruppe_fahrzeuge"); err != nil {
		return fmt.Errorf("failed to clear gruppe_fahrzeuge: %w", err)
	}

	linkStmt, err := tx.Prepare("INSERT INTO gruppe_fahrzeuge (group_id, fahrzeug_id) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare gruppe_fahrzeuge statement: %w", err)
	}
	defer linkStmt.Close()

	updateStmt, err := tx.Prepare("UPDATE fahrzeuge SET fahrer_name = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare fahrzeuge update statement: %w", err)
	}
	defer updateStmt.Close()

	for _, group := range groups {
		for _, f := range group.Fahrzeuge {
			if _, err = linkStmt.Exec(group.GroupID, f.ID); err != nil {
				return fmt.Errorf("failed to insert gruppe_fahrzeuge: %w", err)
			}
			if _, err = updateStmt.Exec(f.FahrerName, f.ID); err != nil {
				return fmt.Errorf("failed to update fahrer_name for fahrzeug %d: %w", f.ID, err)
			}
		}
	}

	return tx.Commit()
}

// SaveCarGroups persists the in-memory CarGroup pool assignments so that a
// backup/restore cycle can fully reconstruct the distribution result.
//
// It writes two junction tables:
//   - cargroup_groups    (pool_id, group_id)   — which participant group is in which pool
//   - cargroup_fahrzeuge (pool_id, fahrzeug_id) — which vehicle belongs to which pool
//
// It also updates fahrer_name in the fahrzeuge table to reflect any driver
// assigned by Phase 3 of the pool algorithm.
func SaveCarGroups(db *sql.DB, carGroups []*models.CarGroup) error {
	if len(carGroups) == 0 {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err = tx.Exec("DELETE FROM cargroup_fahrzeuge"); err != nil {
		return fmt.Errorf("failed to clear cargroup_fahrzeuge: %w", err)
	}
	if _, err = tx.Exec("DELETE FROM cargroup_groups"); err != nil {
		return fmt.Errorf("failed to clear cargroup_groups: %w", err)
	}

	groupStmt, err := tx.Prepare("INSERT INTO cargroup_groups (pool_id, group_id) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare cargroup_groups statement: %w", err)
	}
	defer groupStmt.Close()

	carStmt, err := tx.Prepare("INSERT INTO cargroup_fahrzeuge (pool_id, fahrzeug_id) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare cargroup_fahrzeuge statement: %w", err)
	}
	defer carStmt.Close()

	updateStmt2, err := tx.Prepare("UPDATE fahrzeuge SET fahrer_name = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare fahrzeuge update statement: %w", err)
	}
	defer updateStmt2.Close()

	for _, cg := range carGroups {
		for _, g := range cg.Groups {
			if _, err = groupStmt.Exec(cg.ID, g.GroupID); err != nil {
				return fmt.Errorf("failed to insert cargroup_groups: %w", err)
			}
		}
		for _, f := range cg.Cars {
			if _, err = carStmt.Exec(cg.ID, f.ID); err != nil {
				return fmt.Errorf("failed to insert cargroup_fahrzeuge: %w", err)
			}
			if _, err = updateStmt2.Exec(f.FahrerName, f.ID); err != nil {
				return fmt.Errorf("failed to update fahrer_name for fahrzeug %d: %w", f.ID, err)
			}
		}
	}

	return tx.Commit()
}

// NextTeilnehmerIDConn returns the next available teilnehmer_id.
// Must be called on a conn that is already inside a transaction.
func NextTeilnehmerIDConn(conn *sql.Conn, ctx context.Context) (int64, error) {
	var id int64
	err := conn.QueryRowContext(ctx, "SELECT COALESCE(MAX(teilnehmer_id),0)+1 FROM teilnehmende").Scan(&id)
	return id, err
}

// AddTeilnehmerToGroupConn inserts a participant and assigns them to a group
// atomically. Must be called on a conn that is already inside a transaction.
func AddTeilnehmerToGroupConn(conn *sql.Conn, ctx context.Context, t models.Teilnehmende, groupID int) error {
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO teilnehmende (teilnehmer_id, name, ortsverband, age, geschlecht, pregroup) VALUES (?, ?, ?, ?, ?, '')",
		t.TeilnehmendeID, t.Name, t.Ortsverband, t.Alter, t.Geschlecht,
	); err != nil {
		return fmt.Errorf("teilnehmende insert failed: %w", err)
	}
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO gruppe (group_id, teilnehmer_id) VALUES (?, ?)",
		groupID, t.TeilnehmendeID,
	); err != nil {
		return fmt.Errorf("gruppe insert failed: %w", err)
	}
	return nil
}

// DeleteTeilnehmerConn removes a participant (by teilnehmer_id) and their
// gruppe row. Returns the groupID they belonged to and whether that group is
// now empty (triggering orphaned-row cleanup).
// Must be called on a conn that is already inside a transaction.
func DeleteTeilnehmerConn(conn *sql.Conn, ctx context.Context, teilnehmerID int64) (groupID int, groupEmpty bool, err error) {
	err = conn.QueryRowContext(ctx, "SELECT group_id FROM gruppe WHERE teilnehmer_id = ?", teilnehmerID).Scan(&groupID)
	if err == sql.ErrNoRows {
		groupID = 0
	} else if err != nil {
		return 0, false, fmt.Errorf("gruppe lookup failed: %w", err)
	}

	if _, err = conn.ExecContext(ctx, "DELETE FROM gruppe WHERE teilnehmer_id = ?", teilnehmerID); err != nil {
		return 0, false, fmt.Errorf("gruppe delete failed: %w", err)
	}
	if _, err = conn.ExecContext(ctx, "DELETE FROM teilnehmende WHERE teilnehmer_id = ?", teilnehmerID); err != nil {
		return 0, false, fmt.Errorf("teilnehmende delete failed: %w", err)
	}

	if groupID > 0 {
		var cnt int
		if err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM gruppe WHERE group_id = ?", groupID).Scan(&cnt); err != nil {
			return groupID, false, fmt.Errorf("group count check failed: %w", err)
		}
		groupEmpty = cnt == 0
		if groupEmpty {
			if _, err = conn.ExecContext(ctx, "DELETE FROM gruppe_fahrzeuge WHERE group_id = ?", groupID); err != nil {
				return groupID, true, fmt.Errorf("gruppe_fahrzeuge cleanup failed: %w", err)
			}
			if _, err = conn.ExecContext(ctx, "DELETE FROM cargroup_groups WHERE group_id = ?", groupID); err != nil {
				return groupID, true, fmt.Errorf("cargroup_groups cleanup failed: %w", err)
			}
		}
	}
	return groupID, groupEmpty, nil
}
