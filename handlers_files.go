package main

import (
	"fmt"
	"os"

	"THW-JugendOlympiade/backend/database"
	"THW-JugendOlympiade/backend/io"
	"THW-JugendOlympiade/backend/models"
	"THW-JugendOlympiade/backend/services"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// CheckDB checks if the database has any data.
func (a *App) CheckDB() map[string]interface{} {
	hasData := false
	count := 0

	if a.db != nil {
		var rowCount int
		err := a.db.QueryRow("SELECT COUNT(*) FROM teilnehmende").Scan(&rowCount)
		if err == nil && rowCount > 0 {
			hasData = true
			count = rowCount
		}
	}

	return map[string]interface{}{
		"hasData": hasData,
		"count":   count,
	}
}

// LoadFile opens a file dialog and loads the selected Excel file.
func (a *App) LoadFile() map[string]interface{} {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Excel File",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Excel Files (*.xlsx)",
				Pattern:     "*.xlsx",
			},
		},
	})

	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Datei-Dialog konnte nicht geöffnet werden: %v", err),
		}
	}

	if filePath == "" {
		return map[string]interface{}{
			"status":  "cancelled",
			"message": "Dateiauswahl abgebrochen",
		}
	}

	rows, err := io.ReadXLSXFile(filePath)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Excel-Datei konnte nicht gelesen werden: %v", err),
		}
	}

	if a.db != nil {
		a.db.Close()
		a.db = nil
	}
	// If a database already exists, rename it to a backup so it can be restored if init fails.
	// On a fresh start there is no existing DB, so we skip the backup.
	var dbBackup string
	if _, statErr := os.Stat(models.DbFile); statErr == nil {
		dbBackup = models.DbFile + ".bak"
		if renameErr := os.Rename(models.DbFile, dbBackup); renameErr != nil {
			return map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Datenbank-Backup konnte nicht erstellt werden: %v", renameErr),
			}
		}
	}

	db, err := database.InitDatabase()
	if err != nil {
		if dbBackup != "" {
			// Restore the previous database so the user doesn't lose data
			_ = os.Rename(dbBackup, models.DbFile)
		} else {
			// No prior DB — remove any partial file that InitDatabase may have created
			_ = os.Remove(models.DbFile)
		}
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Datenbank konnte nicht initialisiert werden: %v", err),
		}
	}
	// New DB is healthy — discard the backup
	if dbBackup != "" {
		_ = os.Remove(dbBackup)
	}
	a.db = db

	if err := database.InsertData(a.db, rows); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Daten konnten nicht eingefügt werden: %v", err),
		}
	}

	stationRows, err := io.ReadStationsFromXLSX(filePath)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Stationen konnten nicht gelesen werden: %v", err),
		}
	}

	if err := database.InsertStations(a.db, stationRows); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Stationen konnten nicht eingefügt werden: %v", err),
		}
	}

	betreuendeRows, err := io.ReadBetreuendeFromXLSX(filePath)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Betreuende konnten nicht gelesen werden: %v", err),
		}
	}

	if err := database.InsertBetreuende(a.db, betreuendeRows); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Betreuende konnten nicht eingefügt werden: %v", err),
		}
	}

	participantCount := len(rows) - 1
	return map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Erfolgreich %d Teilnehmende geladen", participantCount),
		"count":   participantCount,
	}
}

// HasScores returns whether any score has been saved to the database.
func (a *App) HasScores() bool {
	if a.db == nil {
		return false
	}
	var count int
	_ = a.db.QueryRow("SELECT COUNT(*) FROM group_station_scores WHERE score IS NOT NULL").Scan(&count)
	return count > 0
}

// DistributeGroups creates balanced groups from the loaded participants.
func (a *App) DistributeGroups() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}
	if err := services.CreateBalancedGroups(a.db, a.cfg.Gruppen.MaxGroesse); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Gruppen konnten nicht erstellt werden: %v", err),
		}
	}
	return map[string]interface{}{
		"status":  "success",
		"message": "Ausgewogene Gruppen wurden erstellt.",
	}
}
