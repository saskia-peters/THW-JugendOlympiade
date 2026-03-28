package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

const ConfigFile = "config.toml"

// Config holds all user-configurable settings for the application.
type Config struct {
	Veranstaltung VeranstaltungConfig `toml:"veranstaltung"`
	Gruppen       GruppenConfig       `toml:"gruppen"`
	Ergebnisse    ErgebnisseConfig    `toml:"ergebnisse"`
	Ausgabe       AusgabeConfig       `toml:"ausgabe"`
}

type VeranstaltungConfig struct {
	Name string `toml:"name"`
	Jahr int    `toml:"jahr"`
}

type GruppenConfig struct {
	MaxGroesse int `toml:"max_groesse"`
}

type ErgebnisseConfig struct {
	MinPunkte int `toml:"min_punkte"`
	MaxPunkte int `toml:"max_punkte"`
}

type AusgabeConfig struct {
	PDFOrdner string `toml:"pdf_ordner"`
	DBName    string `toml:"db_name"`
	// UrkunderStil controls the participant certificate style.
	// "text" (default): group members table. "picture": group photo.
	UrkunderStil string `toml:"urkunden_stil"`
	BilderOrdner string `toml:"bilder_ordner"`
}

// Default returns the factory default configuration.
func Default() Config {
	return Config{
		Veranstaltung: VeranstaltungConfig{
			Name: "THW-JugendOlympiade",
			Jahr: 2026,
		},
		Gruppen: GruppenConfig{
			MaxGroesse: 8,
		},
		Ergebnisse: ErgebnisseConfig{
			MinPunkte: 100,
			MaxPunkte: 1200,
		},
		Ausgabe: AusgabeConfig{
			PDFOrdner:    "pdfdocs",
			DBName:       "data.db",
			UrkunderStil: "text",
			BilderOrdner: "pictures",
		},
	}
}

// LoadOrCreate reads config.toml if it exists; otherwise creates it with
// defaults and returns those defaults.
func LoadOrCreate() (Config, error) {
	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		cfg := Default()
		return cfg, writeDefault()
	}
	var cfg Config
	if _, err := toml.DecodeFile(ConfigFile, &cfg); err != nil {
		return Default(), fmt.Errorf("config.toml konnte nicht geladen werden: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Default(), fmt.Errorf("config.toml enthält ungültige Werte: %w", err)
	}
	return cfg, nil
}

// defaultTOML is written on first launch so the user can inspect and edit it.
const defaultTOML = `# Jugendolympiade - Konfiguration
# Diese Datei kann mit einem einfachen Texteditor (z. B. Notepad) bearbeitet werden.
# Zeilen, die mit # beginnen, sind Kommentare und werden ignoriert.

[veranstaltung]
# Name der Veranstaltung (erscheint auf Urkunden und PDFs)
name = "THW-JugendOlympiade"
# Jahreszahl der Veranstaltung
jahr = 2026

[gruppen]
# Maximale Anzahl Teilnehmende pro Gruppe
max_groesse = 8

[ergebnisse]
# Kleinstes erlaubtes Ergebnis pro Station
min_punkte = 100
# Groesstes erlaubtes Ergebnis pro Station
max_punkte = 1200

[ausgabe]
# Unterordner, in dem erzeugte PDFs gespeichert werden
pdf_ordner = "pdfdocs"
# Dateiname der SQLite-Datenbank (Standard: data.db)
# Nützlich, um verschiedene Veranstaltungen getrennt zu halten.
db_name = "data.db"
# Stil der Teilnehmer-Urkunden: "text" (Mitgliedertabelle) oder "picture" (Gruppenfoto)
urkunden_stil = "text"
# Unterordner mit Gruppenfotos (nur relevant bei urkunden_stil = "picture")
# Dateinamen: group_picture_001.jpg, group_picture_002.jpg, ...
bilder_ordner = "pictures"
`

func writeDefault() error {
	return os.WriteFile(ConfigFile, []byte(defaultTOML), 0644)
}

// ReadRaw returns the raw text content of config.toml.
func ReadRaw() (string, error) {
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return "", fmt.Errorf("Konfiguration konnte nicht gelesen werden: %w", err)
	}
	return string(data), nil
}

// Validate checks semantic constraints that TOML parsing cannot enforce.
func (c Config) Validate() error {
	if c.Gruppen.MaxGroesse < 1 {
		return fmt.Errorf("max_groesse muss mindestens 1 sein (aktuell: %d)", c.Gruppen.MaxGroesse)
	}
	if c.Ergebnisse.MaxPunkte <= c.Ergebnisse.MinPunkte {
		return fmt.Errorf("max_punkte (%d) muss größer als min_punkte (%d) sein", c.Ergebnisse.MaxPunkte, c.Ergebnisse.MinPunkte)
	}
	return nil
}

// ValidateAndSave parses content as TOML, validates semantic constraints,
// writes it to config.toml, and returns the resulting Config so the caller
// can update its in-memory copy.
func ValidateAndSave(content string) (Config, error) {
	var cfg Config
	if _, err := toml.Decode(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("ungültige TOML-Syntax: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	if err := os.WriteFile(ConfigFile, []byte(content), 0644); err != nil {
		return Config{}, fmt.Errorf("Konfiguration konnte nicht gespeichert werden: %w", err)
	}
	return cfg, nil
}
