package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"THW-JugendOlympiade/backend/config"
	"THW-JugendOlympiade/backend/io"
	"THW-JugendOlympiade/backend/models"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	_ "modernc.org/sqlite"
)

//go:embed all:frontend
//go:embed assets
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

	// Seed templates/, example/ etc. on first run from embedded defaults.
	extractDefaultAssets(assets)
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

// extractDefaultAssets walks the embedded "assets/" tree and writes any file that
// does not yet exist on disk, stripping the leading "assets/" prefix so that
// "assets/templates/foo.png" is placed at "templates/foo.png" relative to the cwd.
// Existing files are never overwritten, preserving user customisations.
func extractDefaultAssets(fsys embed.FS) {
	_ = fs.WalkDir(fsys, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		dest := strings.TrimPrefix(path, "assets/")
		if _, statErr := os.Stat(dest); statErr == nil {
			return nil // already exists — never overwrite user's file
		}
		data, readErr := fsys.ReadFile(path)
		if readErr != nil {
			fmt.Printf("Standarddatei konnte nicht gelesen werden (%s): %v\n", path, readErr)
			return nil
		}
		if mkdirErr := os.MkdirAll(filepath.Dir(dest), 0755); mkdirErr != nil {
			fmt.Printf("Ordner konnte nicht erstellt werden (%s): %v\n", filepath.Dir(dest), mkdirErr)
			return nil
		}
		if writeErr := os.WriteFile(dest, data, 0644); writeErr != nil {
			fmt.Printf("Standarddatei konnte nicht extrahiert werden (%s): %v\n", dest, writeErr)
		}
		return nil
	})
}
