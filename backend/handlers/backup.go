package handlers

import (
	"database/sql"
	"fmt"
	iolib "io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"THW-JugendOlympiade/backend/models"
)

// BackupDatabase creates a timestamped backup of the database.
func BackupDatabase(db *sql.DB) map[string]interface{} {
	if db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Keine Datenbank vorhanden. Bitte zuerst eine Excel-Datei laden.",
		}
	}
	backupDir := "dbbackups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Backup-Verzeichnis konnte nicht erstellt werden: %v", err),
		}
	}
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupFilename := fmt.Sprintf("data_backup_%s.db", timestamp)
	backupPath := filepath.Join(backupDir, backupFilename)

	sourceFile, err := os.Open(models.DbFile)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Datenbankdatei konnte nicht geöffnet werden: %v", err),
		}
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Backup-Datei konnte nicht erstellt werden: %v", err),
		}
	}
	defer destFile.Close()

	bytesWritten, err := iolib.Copy(destFile, sourceFile)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Datenbank konnte nicht kopiert werden: %v", err),
		}
	}
	if err := destFile.Sync(); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Backup-Datei konnte nicht synchronisiert werden: %v", err),
		}
	}

	absPath, _ := os.Getwd()
	fullPath := filepath.Join(absPath, backupPath)
	return map[string]interface{}{
		"status":    "success",
		"message":   fmt.Sprintf("Datenbank erfolgreich gesichert (%d Bytes)", bytesWritten),
		"file":      backupFilename,
		"path":      fullPath,
		"size":      bytesWritten,
		"timestamp": timestamp,
	}
}

// ListBackups returns a list of available database backups.
func ListBackups() map[string]interface{} {
	backupDir := "dbbackups"
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return map[string]interface{}{
			"status":  "success",
			"backups": []map[string]interface{}{},
			"count":   0,
		}
	}
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Backup-Verzeichnis konnte nicht gelesen werden: %v", err),
		}
	}
	backups := []map[string]interface{}{}
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".db" {
			continue
		}
		info, err := file.Info()
		if err != nil {
			continue
		}
		backups = append(backups, map[string]interface{}{
			"name":     file.Name(),
			"size":     info.Size(),
			"modified": info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}
	return map[string]interface{}{
		"status":  "success",
		"backups": backups,
		"count":   len(backups),
	}
}

// RestoreDatabase restores the database from a backup file.
func RestoreDatabase(db **sql.DB, backupFilename string) map[string]interface{} {
	backupDir := "dbbackups"

	// Guard against path traversal (e.g. "../config.toml").
	absDir, err := filepath.Abs(backupDir)
	if err != nil {
		return map[string]interface{}{"status": "error", "message": "Interner Fehler beim Prüfen des Backup-Verzeichnisses."}
	}
	absFile, err := filepath.Abs(filepath.Join(backupDir, backupFilename))
	if err != nil || !strings.HasPrefix(absFile, absDir+string(filepath.Separator)) {
		return map[string]interface{}{"status": "error", "message": "Ungültiger Dateiname."}
	}
	backupPath := absFile

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return map[string]interface{}{
			"status":  "error",
			"message": "Backup-Datei nicht gefunden.",
		}
	}

	// Integrity-check the backup BEFORE touching the live database.
	checkDB, err := sql.Open("sqlite", backupPath)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Backup-Datei konnte nicht geprüft werden: %v", err),
		}
	}
	var quickCheckResult string
	if err := checkDB.QueryRow("PRAGMA quick_check").Scan(&quickCheckResult); err != nil {
		checkDB.Close()
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Integritätsprüfung fehlgeschlagen: %v", err),
		}
	}
	checkDB.Close()
	if quickCheckResult != "ok" {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Backup ist beschädigt (PRAGMA quick_check: %s). Wiederherstellung abgebrochen.", quickCheckResult),
		}
	}

	if *db != nil {
		(*db).Close()
		*db = nil
	}

	if err := os.Remove(models.DbFile); err != nil && !os.IsNotExist(err) {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Aktuelle Datenbank konnte nicht entfernt werden: %v", err),
		}
	}

	sourceFile, err := os.Open(backupPath)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Backup-Datei konnte nicht geöffnet werden: %v", err),
		}
	}
	defer sourceFile.Close()

	destFile, err := os.Create(models.DbFile)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Datenbankdatei konnte nicht erstellt werden: %v", err),
		}
	}
	defer destFile.Close()

	bytesWritten, err := iolib.Copy(destFile, sourceFile)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Datenbank konnte nicht wiederhergestellt werden: %v", err),
		}
	}
	if err := destFile.Sync(); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Datenbankdatei konnte nicht synchronisiert werden: %v", err),
		}
	}

	newDB, err := sql.Open("sqlite", models.DbFile)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Wiederhergestellte Datenbank konnte nicht geöffnet werden: %v", err),
		}
	}
	if _, err = newDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Fremdschlüssel konnten nicht aktiviert werden: %v", err),
		}
	}
	*db = newDB

	return map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Datenbank erfolgreich aus %s wiederhergestellt (%d Bytes)", backupFilename, bytesWritten),
		"file":    backupFilename,
		"size":    bytesWritten,
	}
}
