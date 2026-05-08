package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"THW-JugendOlympiade/backend/config"
	"THW-JugendOlympiade/backend/database"
	"THW-JugendOlympiade/backend/io"
	"THW-JugendOlympiade/backend/models"
	"THW-JugendOlympiade/backend/services"
)

// GetParticipantsWithGroups returns all participants with their group assignments
// and a flag indicating whether score entry has started (which locks adds/removes).
func GetParticipantsWithGroups(db *sql.DB, groupNames []string) map[string]interface{} {
	if db == nil {
		return participantErr("Bitte zuerst eine Excel-Datei laden.")
	}
	locked, err := database.AnyScoreExists(db)
	if err != nil {
		return participantErr(fmt.Sprintf("Sperrstatus konnte nicht geprüft werden: %v", err))
	}

	rows, err := db.Query(`
		SELECT t.teilnehmer_id, t.name, t.ortsverband, COALESCE(t.age,0), t.geschlecht, COALESCE(g.group_id,0)
		FROM teilnehmende t
		LEFT JOIN gruppe g ON g.teilnehmer_id = t.teilnehmer_id
		ORDER BY COALESCE(g.group_id,0), t.name
	`)
	if err != nil {
		return participantErr(fmt.Sprintf("Teilnehmende konnten nicht abgerufen werden: %v", err))
	}
	defer rows.Close()

	var participants []models.ParticipantRow
	for rows.Next() {
		var p models.ParticipantRow
		if err := rows.Scan(&p.ID, &p.Name, &p.Ortsverband, &p.Alter, &p.Geschlecht, &p.GroupID); err != nil {
			return participantErr(fmt.Sprintf("Lesefehler: %v", err))
		}
		if p.GroupID > 0 {
			p.GroupName = config.GetGroupName(p.GroupID, groupNames)
		}
		participants = append(participants, p)
	}
	if err := rows.Err(); err != nil {
		return participantErr(fmt.Sprintf("Datenbankfehler: %v", err))
	}

	return map[string]interface{}{
		"status":       "success",
		"rows":         participants,
		"scoresLocked": locked,
	}
}

// GetEligibleGroups returns the groups that currently have at least one free
// participant slot, factoring in carpool seat capacity for FixGroupSize+CarGroups mode.
// No redistribution of cars, drivers, or cargroups is ever performed.
func GetEligibleGroups(db *sql.DB, cfg config.Config, groupNames []string) map[string]interface{} {
	if db == nil {
		return participantErr("Bitte zuerst eine Excel-Datei laden.")
	}

	memberCounts, err := database.GetGroupMemberCounts(db)
	if err != nil {
		return participantErr(fmt.Sprintf("Gruppenbesetzung konnte nicht gelesen werden: %v", err))
	}

	mode := cfg.Verteilung.Verteilungsmodus
	carGroupsEnabled := strings.EqualFold(cfg.Verteilung.CarGroups, "ja")

	var eligible []models.EligibleGroup

	if mode == "FixGroupSize" && carGroupsEnabled {
		fixedSize := cfg.Verteilung.FixGroupSize
		carGroups := services.GetLastCarGroups()

		caretakerCounts, err := database.GetGroupCaretakerCounts(db)
		if err != nil {
			return participantErr(fmt.Sprintf("Betreuenden-Besetzung konnte nicht gelesen werden: %v", err))
		}

		if len(carGroups) == 0 {
			// No pools in memory — fall back to group-size-only check.
			for gid, cnt := range memberCounts {
				if cnt < fixedSize {
					eligible = append(eligible, models.EligibleGroup{
						GroupID:      gid,
						GroupName:    config.GetGroupName(gid, groupNames),
						CurrentCount: cnt,
						MaxSlots:     fixedSize,
					})
				}
			}
		} else {
			// Check group size AND carpool seat availability.
			// Cars/drivers/cargroups are never redistributed.
			for _, pool := range carGroups {
				poolSeats := 0
				for _, car := range pool.Cars {
					poolSeats += car.Sitzplaetze
				}
				// Current headcount from DB (participants + caretakers) is authoritative.
				poolHeadcount := 0
				for _, g := range pool.Groups {
					poolHeadcount += memberCounts[g.GroupID] + caretakerCounts[g.GroupID]
				}
				poolFreeSeats := poolSeats - poolHeadcount

				for _, g := range pool.Groups {
					cnt := memberCounts[g.GroupID]
					if cnt < fixedSize && poolFreeSeats > 0 {
						// Cap MaxSlots at whichever limit is tighter: the per-group
						// fixedSize or what the pool can still physically accommodate.
						effectiveMax := fixedSize
						if cnt+poolFreeSeats < fixedSize {
							effectiveMax = cnt + poolFreeSeats
						}
						eligible = append(eligible, models.EligibleGroup{
							GroupID:      g.GroupID,
							GroupName:    config.GetGroupName(g.GroupID, groupNames),
							CurrentCount: cnt,
							MaxSlots:     effectiveMax,
						})
					}
				}
			}
		}
	} else {
		// Klassisch / Fahrzeuge: check group size against vehicle seats (or MaxGroesse).
		vehicleSeats, err := database.GetGroupVehicleSeats(db)
		if err != nil {
			return participantErr(fmt.Sprintf("Fahrzeug-Sitzplätze konnten nicht gelesen werden: %v", err))
		}
		maxGroesse := cfg.Gruppen.MaxGroesse
		for gid, cnt := range memberCounts {
			seats, hasVehicle := vehicleSeats[gid]
			maxSlots := maxGroesse
			if hasVehicle {
				maxSlots = seats
			}
			if cnt < maxSlots {
				eligible = append(eligible, models.EligibleGroup{
					GroupID:      gid,
					GroupName:    config.GetGroupName(gid, groupNames),
					CurrentCount: cnt,
					MaxSlots:     maxSlots,
				})
			}
		}
	}

	// Sort fewest-members-first, break ties by GroupID.
	sort.Slice(eligible, func(i, j int) bool {
		if eligible[i].CurrentCount != eligible[j].CurrentCount {
			return eligible[i].CurrentCount < eligible[j].CurrentCount
		}
		return eligible[i].GroupID < eligible[j].GroupID
	})

	return map[string]interface{}{
		"status": "success",
		"groups": eligible,
	}
}

// AddTeilnehmer adds a new participant to the specified group.
// Pre-validates free slots (including carpool seat capacity) before the atomic
// insert. Under no circumstances are cars, drivers, or cargroups redistributed.
func AddTeilnehmer(db *sql.DB, name, ortsverband string, alter int, geschlecht string, groupID int, cfg config.Config, groupNames []string) map[string]interface{} {
	if db == nil {
		return participantErr("Bitte zuerst eine Excel-Datei laden.")
	}

	// --- Pre-validation (outside transaction, safe because SetMaxOpenConns(1)) ---

	locked, err := database.AnyScoreExists(db)
	if err != nil {
		return participantErr(fmt.Sprintf("Sperrstatus konnte nicht geprüft werden: %v", err))
	}
	if locked {
		return participantErr("Hinzufügen ist nicht möglich, da bereits Wertungen eingetragen wurden.")
	}

	// Server-side eligibility check: verify the target group still has a free
	// slot AND (in FixGroupSize+CarGroups mode) the pool has a free seat.
	eligibleResp := GetEligibleGroups(db, cfg, groupNames)
	if eligibleResp["status"] != "success" {
		return eligibleResp
	}
	groupsIface := eligibleResp["groups"]
	eligibleGroups, _ := groupsIface.([]models.EligibleGroup)
	found := false
	for _, eg := range eligibleGroups {
		if eg.GroupID == groupID {
			found = true
			break
		}
	}
	if !found {
		return participantErr(fmt.Sprintf(
			"Gruppe %s hat keinen freien Platz (inkl. Fahrgemeinschaft). Bitte eine andere Gruppe wählen.",
			config.GetGroupName(groupID, groupNames),
		))
	}

	// --- Atomic insert via dedicated conn with BEGIN IMMEDIATE ---

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return participantErr(fmt.Sprintf("Datenbankverbindung fehlgeschlagen: %v", err))
	}

	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		conn.Close()
		return participantErr(fmt.Sprintf("Transaktion konnte nicht gestartet werden: %v", err))
	}
	rollback := func() { conn.ExecContext(ctx, "ROLLBACK"); conn.Close() } //nolint:errcheck

	// Definitive score check inside the transaction.
	var scoreCount int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM group_station_scores").Scan(&scoreCount); err != nil {
		rollback()
		return participantErr(fmt.Sprintf("Sperrstatus konnte nicht geprüft werden: %v", err))
	}
	if scoreCount > 0 {
		rollback()
		return participantErr("Hinzufügen ist nicht möglich, da bereits Wertungen eingetragen wurden.")
	}

	newID, err := database.NextTeilnehmerIDConn(conn, ctx)
	if err != nil {
		rollback()
		return participantErr(fmt.Sprintf("Neue Teilnehmer-ID konnte nicht ermittelt werden: %v", err))
	}

	t := models.Teilnehmende{
		TeilnehmendeID: int(newID),
		Name:           name,
		Ortsverband:    ortsverband,
		Alter:          alter,
		Geschlecht:     geschlecht,
	}
	if err := database.AddTeilnehmerToGroupConn(conn, ctx, t, groupID); err != nil {
		rollback()
		return participantErr(fmt.Sprintf("Teilnehmer/in konnte nicht hinzugefügt werden: %v", err))
	}

	if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
		rollback()
		return participantErr(fmt.Sprintf("Transaktion konnte nicht abgeschlossen werden: %v", err))
	}
	// Release the pinned connection before PDF generation — with SetMaxOpenConns(1)
	// the connection must be back in the pool before db.* calls in the PDF generators.
	conn.Close()

	// Refresh in-memory CarGroups so the CarGroup PDF reflects the new headcount.
	if cgs, err := database.LoadCarGroups(db); err == nil && len(cgs) > 0 {
		services.SetLastCarGroups(cgs)
	}

	groupName := config.GetGroupName(groupID, groupNames)
	pdfResults := regenerateMembershipPDFs(db, cfg, groupNames)

	return map[string]interface{}{
		"status":     "success",
		"message":    fmt.Sprintf("%s wurde Gruppe %s hinzugefügt.", name, groupName),
		"name":       name,
		"groupName":  groupName,
		"newID":      newID,
		"pdfResults": pdfResults,
	}
}

// RemoveTeilnehmer removes a participant (identified by teilnehmer_id) from the event.
// Under no circumstances are cars, drivers, or cargroups redistributed.
func RemoveTeilnehmer(db *sql.DB, teilnehmerID int64, cfg config.Config, groupNames []string) map[string]interface{} {
	if db == nil {
		return participantErr("Bitte zuerst eine Excel-Datei laden.")
	}

	// Fetch name and group before the transaction (for the result message).
	var pName string
	var pGroupID int
	err := db.QueryRow(`
		SELECT t.name, COALESCE(g.group_id,0)
		FROM teilnehmende t
		LEFT JOIN gruppe g ON g.teilnehmer_id = t.teilnehmer_id
		WHERE t.teilnehmer_id = ?`, teilnehmerID).Scan(&pName, &pGroupID)
	if err == sql.ErrNoRows {
		return participantErr("Teilnehmer/in nicht gefunden.")
	} else if err != nil {
		return participantErr(fmt.Sprintf("Teilnehmer/in konnte nicht gefunden werden: %v", err))
	}
	pGroupName := ""
	if pGroupID > 0 {
		pGroupName = config.GetGroupName(pGroupID, groupNames)
	}

	// --- Atomic delete via dedicated conn with BEGIN IMMEDIATE ---

	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return participantErr(fmt.Sprintf("Datenbankverbindung fehlgeschlagen: %v", err))
	}

	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		conn.Close()
		return participantErr(fmt.Sprintf("Transaktion konnte nicht gestartet werden: %v", err))
	}
	rollback := func() { conn.ExecContext(ctx, "ROLLBACK"); conn.Close() } //nolint:errcheck

	// Definitive score check inside the transaction.
	var scoreCount int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM group_station_scores").Scan(&scoreCount); err != nil {
		rollback()
		return participantErr(fmt.Sprintf("Sperrstatus konnte nicht geprüft werden: %v", err))
	}
	if scoreCount > 0 {
		rollback()
		return participantErr("Entfernen ist nicht möglich, da bereits Wertungen eingetragen wurden.")
	}

	_, _, err = database.DeleteTeilnehmerConn(conn, ctx, teilnehmerID)
	if err != nil {
		rollback()
		return participantErr(fmt.Sprintf("Teilnehmer/in konnte nicht entfernt werden: %v", err))
	}

	if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
		rollback()
		return participantErr(fmt.Sprintf("Transaktion konnte nicht abgeschlossen werden: %v", err))
	}
	// Release the pinned connection before any db.* calls below — with
	// SetMaxOpenConns(1) keeping it open deadlocks LoadCarGroups and the PDF generators.
	conn.Close()

	// Reload in-memory CarGroups from DB so the CarGroup PDF reflects the
	// updated headcount. No cars, drivers, or pools are redistributed.
	if cgs, err := database.LoadCarGroups(db); err == nil && len(cgs) > 0 {
		services.SetLastCarGroups(cgs)
	}

	pdfResults := regenerateMembershipPDFs(db, cfg, groupNames)

	return map[string]interface{}{
		"status":     "success",
		"message":    fmt.Sprintf("%s wurde aus Gruppe %s entfernt.", pName, pGroupName),
		"name":       pName,
		"groupName":  pGroupName,
		"pdfResults": pdfResults,
	}
}

// regenerateMembershipPDFs regenerates all PDFs that depend on group composition.
// PDF failures are non-fatal: they are collected as warnings and do not roll back
// the already-committed DB mutation.
func regenerateMembershipPDFs(db *sql.DB, cfg config.Config, groupNames []string) []models.PDFResult {
	carGroups := services.GetLastCarGroups()
	var results []models.PDFResult

	run := func(name string, fn func() error) {
		if err := fn(); err != nil {
			results = append(results, models.PDFResult{Name: name, Status: "error", Error: err.Error()})
		} else {
			results = append(results, models.PDFResult{Name: name, Status: "ok"})
		}
	}

	run("Gruppenaufteilung", func() error {
		return io.GeneratePDFReport(db, cfg.Veranstaltung.Name, cfg.Veranstaltung.Jahr, carGroups, groupNames)
	})
	run("Teilnehmendenkarten", func() error {
		return io.GenerateTeilnehmendeCardsPDF(db, cfg.Veranstaltung.Name, cfg.Veranstaltung.Jahr, groupNames)
	})
	run("Stationslaufzettel", func() error {
		return io.GenerateStationSheetsPDF(db, cfg.Veranstaltung.Name, cfg.Veranstaltung.Jahr, groupNames)
	})
	run("OV-Zuweisungen", func() error {
		return io.GenerateOVAssignmentsPDF(db, cfg.Veranstaltung.Name, cfg.Veranstaltung.Jahr, groupNames)
	})
	run("Übersicht", func() error {
		return io.GenerateOverviewPDF(db, cfg.Veranstaltung.Name, cfg.Veranstaltung.Jahr, carGroups)
	})
	if cfg.Verteilung.Verteilungsmodus == "FixGroupSize" &&
		strings.EqualFold(cfg.Verteilung.CarGroups, "ja") &&
		len(carGroups) > 0 {
		run("Fahrgemeinschaften", func() error {
			return io.GenerateCarGroupsPDF(carGroups, cfg.Veranstaltung.Name, cfg.Veranstaltung.Jahr, groupNames, cfg)
		})
	}
	return results
}

func participantErr(msg string) map[string]interface{} {
	return map[string]interface{}{"status": "error", "message": msg}
}
