package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"

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

	if cfg.Ausgabe.BilderOrdner != "" {
		if err := os.MkdirAll(cfg.Ausgabe.BilderOrdner, 0755); err != nil {
			fmt.Printf("Bilderordner konnte nicht erstellt werden: %v\n", err)
		}
	}

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
