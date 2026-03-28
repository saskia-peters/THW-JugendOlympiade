package handlers

import (
	"context"
	"database/sql"
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
func CheckStartup() map[string]interface{} {
	_, err := os.Stat(models.DbFile)
	exists := err == nil
	return map[string]interface{}{
		"exists": exists,
		"dbName": models.DbFile,
	}
}

// UseExistingDB opens the already-existing database without wiping it.
func UseExistingDB(db **sql.DB) map[string]interface{} {
	if *db != nil {
		(*db).Close()
		*db = nil
	}
	newDB, err := database.OpenExistingDB()
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Datenbank konnte nicht geöffnet werden: %v", err),
		}
	}
	*db = newDB
	var count int
	if err := (*db).QueryRow("SELECT COUNT(*) FROM teilnehmende").Scan(&count); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Teilnehmerzahl konnte nicht gelesen werden: %v", err),
		}
	}
	return map[string]interface{}{
		"status": "ok",
		"count":  count,
	}
}

// ResetToFreshDB backs up the existing database file, then initialises a
// brand-new empty database.
func ResetToFreshDB(db **sql.DB) map[string]interface{} {
	if *db != nil {
		(*db).Close()
		*db = nil
	}

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
		if err := os.Remove(models.DbFile); err != nil {
			return map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Alte Datenbank konnte nicht entfernt werden: %v", err),
			}
		}
	}

	newDB, err := database.InitDatabase()
	if err != nil {
		if backupPath != "" {
			_ = os.Rename(backupPath, models.DbFile)
		}
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Neue Datenbank konnte nicht erstellt werden: %v", err),
		}
	}
	*db = newDB
	return map[string]interface{}{
		"status":     "ok",
		"backupPath": backupPath,
	}
}

// CheckDB checks if the database has any data.
func CheckDB(db *sql.DB) map[string]interface{} {
	hasData := false
	count := 0
	if db != nil {
		var rowCount int
		err := db.QueryRow("SELECT COUNT(*) FROM teilnehmende").Scan(&rowCount)
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
func LoadFile(ctx context.Context, db **sql.DB) map[string]interface{} {
	filePath, err := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
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

	if *db != nil {
		(*db).Close()
		*db = nil
	}

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

	newDB, err := database.InitDatabase()
	if err != nil {
		if dbBackup != "" {
			_ = os.Rename(dbBackup, models.DbFile)
		} else {
			_ = os.Remove(models.DbFile)
		}
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Datenbank konnte nicht initialisiert werden: %v", err),
		}
	}
	if dbBackup != "" {
		_ = os.Remove(dbBackup)
	}
	*db = newDB

	if err := database.InsertData(*db, rows); err != nil {
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
	if err := database.InsertStations(*db, stationRows); err != nil {
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
	if err := database.InsertBetreuende(*db, betreuendeRows); err != nil {
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
func HasScores(db *sql.DB) (bool, error) {
	if db == nil {
		return false, nil
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM group_station_scores WHERE score IS NOT NULL").Scan(&count); err != nil {
		return false, fmt.Errorf("Fehler beim Prüfen der Punkte: %w", err)
	}
	return count > 0, nil
}

// DistributeGroups creates balanced groups from the loaded participants.
func DistributeGroups(db *sql.DB, maxGroupSize int) map[string]interface{} {
	if db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Bitte zuerst eine Excel-Datei laden.",
		}
	}
	warning, err := services.CreateBalancedGroups(db, maxGroupSize)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Gruppen konnten nicht erstellt werden: %v", err),
		}
	}
	result := map[string]interface{}{
		"status":  "success",
		"message": "Ausgewogene Gruppen wurden erstellt.",
	}
	if warning != "" {
		result["warning"] = warning
	}
	return result
}
