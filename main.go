package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"THW-JugendOlympiade/backend/config"
	"THW-JugendOlympiade/backend/io"
	"THW-JugendOlympiade/backend/models"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

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

	// Apply configured database file name (non-empty; fall back to default)
	if cfg.Ausgabe.DBName != "" {
		models.DbFile = cfg.Ausgabe.DBName
	}
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
