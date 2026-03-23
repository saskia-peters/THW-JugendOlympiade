package main

import (
	"fmt"
	iolib "io"
	"os"
	"path/filepath"
	"time"

	"THW-JugendOlympiade/backend/database"
	"THW-JugendOlympiade/backend/io"
	"THW-JugendOlympiade/backend/models"
	"THW-JugendOlympiade/backend/services"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// CheckStartup reports whether the configured database file already exists.
// The frontend calls this on app-ready to decide whether to prompt the user.
func (a *App) CheckStartup() map[string]interface{} {
	_, err := os.Stat(models.DbFile)
	exists := err == nil
	return map[string]interface{}{
		"exists": exists,
		"dbName": models.DbFile,
	}
}

// UseExistingDB opens the already-existing database without wiping it.
// Call this when the user chooses to keep the existing data.
func (a *App) UseExistingDB() map[string]interface{} {
	if a.db != nil {
		a.db.Close()
		a.db = nil
	}
	db, err := database.OpenExistingDB()
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Datenbank konnte nicht geöffnet werden: %v", err),
		}
	}
	a.db = db
	// Count participants so the frontend can show the number
	var count int
	_ = a.db.QueryRow("SELECT COUNT(*) FROM teilnehmende").Scan(&count)
	return map[string]interface{}{
		"status": "ok",
		"count":  count,
	}
}

// ResetToFreshDB backs up the existing database file, then initialises a
// brand-new empty database.  Call this when the user chooses a clean start.
func (a *App) ResetToFreshDB() map[string]interface{} {
	if a.db != nil {
		a.db.Close()
		a.db = nil
	}

	// Backup existing file before wiping
	backupPath := ""
	if _, statErr := os.Stat(models.DbFile); statErr == nil {
		backupDir := "dbbackups"
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Backup-Verzeichnis konnte nicht erstellt werden: %v", err),
			}
		}
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		backupFilename := fmt.Sprintf("startup_backup_%s.db", timestamp)
		backupPath = filepath.Join(backupDir, backupFilename)

		src, err := os.Open(models.DbFile)
		if err != nil {
			return map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Bestehende Datenbank konnte nicht geöffnet werden: %v", err),
			}
		}
		dst, err := os.Create(backupPath)
		if err != nil {
			src.Close()
			return map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Backup-Datei konnte nicht erstellt werden: %v", err),
			}
		}
		_, copyErr := iolib.Copy(dst, src)
		dst.Sync()
		dst.Close()
		src.Close()
		if copyErr != nil {
			return map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Backup konnte nicht geschrieben werden: %v", copyErr),
			}
		}

		// Remove the original so InitDatabase creates a fresh one
		if err := os.Remove(models.DbFile); err != nil {
			return map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Alte Datenbank konnte nicht entfernt werden: %v", err),
			}
		}
	}

	db, err := database.InitDatabase()
	if err != nil {
		// InitDatabase failed — restore file from backup if we made one
		if backupPath != "" {
			_ = os.Rename(backupPath, models.DbFile)
		}
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Neue Datenbank konnte nicht erstellt werden: %v", err),
		}
	}
	a.db = db
	return map[string]interface{}{
		"status":     "ok",
		"backupPath": backupPath,
	}
}

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
