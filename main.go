package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	iolib "io"
	"os"
	"path/filepath"
	"time"

	"THW-JugendOlympiade/backend/config"
	"THW-JugendOlympiade/backend/database"
	"THW-JugendOlympiade/backend/io"
	"THW-JugendOlympiade/backend/models"
	"THW-JugendOlympiade/backend/services"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	_ "modernc.org/sqlite"
)

//go:embed all:frontend
var assets embed.FS

// App struct
type App struct {
	ctx context.Context
	db  *sql.DB
	cfg config.Config
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	cfg, err := config.LoadOrCreate()
	if err != nil {
		fmt.Printf("Konfiguration konnte nicht geladen werden: %v\n", err)
	}
	a.cfg = cfg
	io.SetPDFOutputDir(cfg.Ausgabe.PDFOrdner)
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		a.db.Close()
	}
}

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Jugendolympiade Verwaltung",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// GetConfig returns the user-facing configuration values needed by the frontend.
func (a *App) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"scoreMin":     a.cfg.Ergebnisse.MinPunkte,
		"scoreMax":     a.cfg.Ergebnisse.MaxPunkte,
		"maxGroupSize": a.cfg.Gruppen.MaxGroesse,
		"eventName":    a.cfg.Veranstaltung.Name,
		"eventYear":    a.cfg.Veranstaltung.Jahr,
	}
}

// GetConfigRaw returns the raw text content of config.toml for in-app editing.
func (a *App) GetConfigRaw() map[string]interface{} {
	content, err := config.ReadRaw()
	if err != nil {
		return map[string]interface{}{"status": "error", "message": err.Error()}
	}
	return map[string]interface{}{"status": "ok", "content": content}
}

// SaveConfigRaw validates content as TOML, writes config.toml, and reloads the
// in-memory config so changes take effect immediately (where possible).
func (a *App) SaveConfigRaw(content string) map[string]interface{} {
	cfg, err := config.ValidateAndSave(content)
	if err != nil {
		return map[string]interface{}{"status": "error", "message": err.Error()}
	}
	a.cfg = cfg
	io.SetPDFOutputDir(cfg.Ausgabe.PDFOrdner)
	return map[string]interface{}{
		"status":  "ok",
		"message": "Konfiguration gespeichert. Einige Änderungen (z. B. Gruppen, Ergebnisse) werden erst nach einem Neustart der App wirksam.",
	}
}

// CheckDB checks if the database has any data
func (a *App) CheckDB() map[string]interface{} {
	hasData := false
	count := 0

	if a.db != nil {
		var rowCount int
		err := a.db.QueryRow("SELECT COUNT(*) FROM teilnehmer").Scan(&rowCount)
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

// LoadFile opens a file dialog and loads the selected Excel file
func (a *App) LoadFile() map[string]interface{} {
	// Open file dialog
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
			"message": fmt.Sprintf("Failed to open file dialog: %v", err),
		}
	}

	if filePath == "" {
		// User cancelled
		return map[string]interface{}{
			"status":  "cancelled",
			"message": "File selection cancelled",
		}
	}

	// Read XLSX file
	rows, err := io.ReadXLSXFile(filePath)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to read XLSX file: %v", err),
		}
	}

	// Close previous database if any
	if a.db != nil {
		a.db.Close()
		a.db = nil
	}

	// Remove old database file to start fresh
	os.Remove(models.DbFile)

	// Initialize SQLite database
	db, err := database.InitDatabase()
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to initialize database: %v", err),
		}
	}
	a.db = db

	// Insert data into database
	if err := database.InsertData(a.db, rows); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to insert data: %v", err),
		}
	}

	// Read and insert stations from Stationen sheet
	stationRows, err := io.ReadStationsFromXLSX(filePath)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to read stations: %v", err),
		}
	}

	if err := database.InsertStations(a.db, stationRows); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to insert stations: %v", err),
		}
	}

	participantCount := len(rows) - 1

	return map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Successfully loaded %d participants", participantCount),
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

// ShowGroups retrieves and returns groups from the database
func (a *App) ShowGroups() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Please load an Excel file first",
		}
	}

	// Retrieve groups from database
	groups, err := database.GetGroupsForReport(a.db)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to retrieve groups: %v", err),
		}
	}

	if len(groups) == 0 {
		return map[string]interface{}{
			"status":  "error",
			"message": "No groups found. Please load a file first.",
		}
	}

	return map[string]interface{}{
		"status": "success",
		"count":  len(groups),
		"groups": groups,
	}
}

// ShowStations retrieves and returns stations with group scores from the database
func (a *App) ShowStations() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Please load an Excel file first",
		}
	}

	// Retrieve stations from database
	stations, err := database.GetStationsForReport(a.db)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to retrieve stations: %v", err),
		}
	}

	if len(stations) == 0 {
		return map[string]interface{}{
			"status":  "error",
			"message": "No stations found. Please ensure your Excel file has a 'Stationen' sheet.",
		}
	}

	return map[string]interface{}{
		"status":   "success",
		"count":    len(stations),
		"stations": stations,
	}
}

// GetAllGroups retrieves all group IDs from the database
func (a *App) GetAllGroups() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Please load an Excel file first",
		}
	}

	groupIDs, err := database.GetAllGroupIDs(a.db)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to retrieve groups: %v", err),
		}
	}

	return map[string]interface{}{
		"status": "success",
		"groups": groupIDs,
	}
}

// AssignScore assigns a score to a group at a station
func (a *App) AssignScore(groupID int, stationID int, score int) map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Please load an Excel file first",
		}
	}

	err := database.AssignGroupStationScore(a.db, groupID, stationID, score)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to assign score: %v", err),
		}
	}

	return map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Score %d assigned to Group %d", score, groupID),
	}
}

// GetGroupEvaluations retrieves all groups with their total scores ranked from high to low
func (a *App) GetGroupEvaluations() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Please load an Excel file first",
		}
	}

	evaluations, err := database.GetGroupEvaluations(a.db)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to retrieve group evaluations: %v", err),
		}
	}

	if len(evaluations) == 0 {
		return map[string]interface{}{
			"status":  "error",
			"message": "No groups found with scores.",
		}
	}

	return map[string]interface{}{
		"status":      "success",
		"evaluations": evaluations,
	}
}

// GetOrtsverbandEvaluations retrieves all ortsverbands with their total scores ranked from high to low
func (a *App) GetOrtsverbandEvaluations() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Please load an Excel file first",
		}
	}

	evaluations, err := database.GetOrtsverbandEvaluations(a.db)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to retrieve ortsverband evaluations: %v", err),
		}
	}

	if len(evaluations) == 0 {
		return map[string]interface{}{
			"status":  "error",
			"message": "No ortsverbands found with scores.",
		}
	}

	return map[string]interface{}{
		"status":      "success",
		"evaluations": evaluations,
	}
}

// GeneratePDF generates a PDF report
func (a *App) GeneratePDF() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Please load an Excel file first",
		}
	}

	// Generate PDF report
	if err := io.GeneratePDFReport(a.db); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to generate PDF report: %v", err),
		}
	}

	absPath, _ := os.Getwd()

	return map[string]interface{}{
		"status":  "success",
		"message": "PDF report generated successfully",
		"file":    "groups_report.pdf",
		"path":    absPath + string(os.PathSeparator) + "pdfdocs" + string(os.PathSeparator) + "groups_report.pdf",
	}
}

// GenerateGroupEvaluationPDF generates a PDF report for group rankings
func (a *App) GenerateGroupEvaluationPDF() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Please load an Excel file first",
		}
	}

	// Generate PDF report
	if err := io.GenerateGroupEvaluationPDF(a.db); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to generate PDF report: %v", err),
		}
	}

	absPath, _ := os.Getwd()

	return map[string]interface{}{
		"status":  "success",
		"message": "Group evaluation PDF generated successfully",
		"file":    "group_evaluations.pdf",
		"path":    absPath + string(os.PathSeparator) + "pdfdocs" + string(os.PathSeparator) + "group_evaluations.pdf",
	}
}

// GenerateOrtsverbandEvaluationPDF generates a PDF report for ortsverband rankings
func (a *App) GenerateOrtsverbandEvaluationPDF() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Please load an Excel file first",
		}
	}

	// Generate PDF report
	if err := io.GenerateOrtsverbandEvaluationPDF(a.db); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to generate PDF report: %v", err),
		}
	}

	absPath, _ := os.Getwd()

	return map[string]interface{}{
		"status":  "success",
		"message": "Ortsverband evaluation PDF generated successfully",
		"file":    "ortsverband_evaluations.pdf",
		"path":    absPath + string(os.PathSeparator) + "pdfdocs" + string(os.PathSeparator) + "ortsverband_evaluations.pdf",
	}
}

// GenerateParticipantCertificates generates participant certificates PDF
func (a *App) GenerateParticipantCertificates() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Please load an Excel file first",
		}
	}

	// Generate PDF report
	if err := io.GenerateParticipantCertificates(a.db); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to generate PDF certificates: %v", err),
		}
	}

	absPath, _ := os.Getwd()

	return map[string]interface{}{
		"status":  "success",
		"message": "Urkunden Teilnehmende erfolgreich erstellt",
		"file":    "Urkunden_Teilnehmende.pdf",
		"path":    absPath + string(os.PathSeparator) + "pdfdocs" + string(os.PathSeparator) + "Urkunden_Teilnehmende.pdf",
	}
}

// GenerateOrtsverbandCertificates generates ortsverband certificates PDF.
// The best-ranked Ortsverband receives a special Siegerurkunde; all others get
// an identical participation certificate with no ranking mentioned.
func (a *App) GenerateOrtsverbandCertificates() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "Please load an Excel file first",
		}
	}

	if err := io.GenerateOrtsverbandCertificates(a.db); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to generate ortsverband certificates: %v", err),
		}
	}

	absPath, _ := os.Getwd()

	return map[string]interface{}{
		"status":  "success",
		"message": "Urkunden Ortsverbände erfolgreich erstellt",
		"file":    "Urkunden_Ortsverbaende.pdf",
		"path":    absPath + string(os.PathSeparator) + "pdfdocs" + string(os.PathSeparator) + "Urkunden_Ortsverbaende.pdf",
	}
}

// BackupDatabase creates a timestamped backup of the database
func (a *App) BackupDatabase() map[string]interface{} {
	if a.db == nil {
		return map[string]interface{}{
			"status":  "error",
			"message": "No database to backup. Please load an Excel file first",
		}
	}

	// Create backup directory if it doesn't exist
	backupDir := "dbbackups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to create backup directory: %v", err),
		}
	}

	// Create timestamp for backup filename
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupFilename := fmt.Sprintf("data_backup_%s.db", timestamp)
	backupPath := filepath.Join(backupDir, backupFilename)

	// Copy database file
	sourceFile, err := os.Open(models.DbFile)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to open database file: %v", err),
		}
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to create backup file: %v", err),
		}
	}
	defer destFile.Close()

	// Copy the file contents
	bytesWritten, err := iolib.Copy(destFile, sourceFile)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to copy database: %v", err),
		}
	}

	// Ensure all data is written to disk
	if err := destFile.Sync(); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to sync backup file: %v", err),
		}
	}

	absPath, _ := os.Getwd()
	fullPath := filepath.Join(absPath, backupPath)

	return map[string]interface{}{
		"status":    "success",
		"message":   fmt.Sprintf("Database backed up successfully (%d bytes)", bytesWritten),
		"file":      backupFilename,
		"path":      fullPath,
		"size":      bytesWritten,
		"timestamp": timestamp,
	}
}

// ListBackups returns a list of available database backups
func (a *App) ListBackups() map[string]interface{} {
	backupDir := "dbbackups"

	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return map[string]interface{}{
			"status":  "success",
			"backups": []map[string]interface{}{},
			"count":   0,
		}
	}

	// Read backup directory
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to read backup directory: %v", err),
		}
	}

	// Collect backup files
	backups := []map[string]interface{}{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Only include .db files
		if filepath.Ext(file.Name()) != ".db" {
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

// RestoreDatabase restores the database from a backup file
func (a *App) RestoreDatabase(backupFilename string) map[string]interface{} {
	backupDir := "dbbackups"
	backupPath := filepath.Join(backupDir, backupFilename)

	// Validate backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return map[string]interface{}{
			"status":  "error",
			"message": "Backup file does not exist",
		}
	}

	// Close current database connection if open
	if a.db != nil {
		a.db.Close()
		a.db = nil
	}

	// Remove current database file
	if err := os.Remove(models.DbFile); err != nil && !os.IsNotExist(err) {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to remove current database: %v", err),
		}
	}

	// Copy backup file to database location
	sourceFile, err := os.Open(backupPath)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to open backup file: %v", err),
		}
	}
	defer sourceFile.Close()

	destFile, err := os.Create(models.DbFile)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to create database file: %v", err),
		}
	}
	defer destFile.Close()

	// Copy the file contents
	bytesWritten, err := iolib.Copy(destFile, sourceFile)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to restore database: %v", err),
		}
	}

	// Ensure all data is written to disk
	if err := destFile.Sync(); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to sync database file: %v", err),
		}
	}

	// Reopen the database
	db, err := sql.Open("sqlite", models.DbFile)
	if err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to open restored database: %v", err),
		}
	}
	if _, err = db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Failed to enable foreign keys: %v", err),
		}
	}
	a.db = db

	return map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Database restored successfully from %s (%d bytes)", backupFilename, bytesWritten),
		"file":    backupFilename,
		"size":    bytesWritten,
	}
}
