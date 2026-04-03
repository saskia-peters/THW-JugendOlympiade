# Build & Entwicklung

## Voraussetzungen

| Tool | Version | Zweck |
|------|---------|-------|
| [Go](https://go.dev/dl/) | 1.21+ | Backend-Compiler |
| [Wails CLI](https://wails.io/docs/gettingstarted/installation) | v2.x | Desktop-Framework |
| GCC | aktuell | CGO (Windows: MinGW oder TDM-GCC) |
| [uv](https://astral.sh/uv) | aktuell | Python-Umgebung für Dokumentation |
| [just](https://just.systems) | aktuell | Task-Runner |

Node.js wird **nicht benötigt** — das Frontend verwendet native ES6-Module ohne Build-Schritt.

### Wails installieren

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor   # Abhängigkeiten prüfen
```

### GCC installieren (nur Windows)

[TDM-GCC](https://jmeubank.github.io/tdm-gcc/) oder [MinGW-w64](https://www.mingw-w64.org/) herunterladen und zu `PATH` hinzufügen.

```bash
gcc --version   # prüfen
```

## Entwicklungsmodus

```bash
wails dev
```

Oder mit `just`:

```bash
just dev
```

Hot-Reload-Entwicklungsserver:

- Frontend-Änderungen (HTML/CSS/JS) laden sofort neu.
- Go-Backend-Änderungen lösen automatische Neukompilierung aus.
- DevTools über **F12** verfügbar.

**Tastenkürzel im Dev-Modus:**

| Kürzel | Aktion |
|--------|--------|
| `F5` | Frontend neu laden |
| `F12` | DevTools öffnen |
| `Ctrl+C` | Dev-Server beenden |

## Tests ausführen

```bash
just test
# oder direkt:
cd test && go test -v ./...
```

## Produktions-Build

```bash
just build
# oder direkt:
wails build
```

Ausgabe in `build/bin/`:

| Plattform | Ausgabe |
|-----------|---------|
| Windows | `build/bin/THW-JugendOlympiade.exe` |
| macOS | `build/bin/THW-JugendOlympiade.app` |
| Linux | `build/bin/THW-JugendOlympiade` |

## Dokumentation generieren

```bash
just docs          # Dokumentation bauen (Ausgabe → docs/)
just docs-serve    # Lokalen Dev-Server starten (http://localhost:8000)
```

## Projektstruktur

```
THW-JugendOlympiade/
├── main.go                     Anwendungseinstiegspunkt
├── app_handlers.go             Dünne App-Wrapper (Wails-Bindings)
├── config.toml                 Laufzeit-Konfiguration (auto-erstellt)
├── go.mod / go.sum             Go-Moduldateien
├── wails.json                  Wails-Build-Konfiguration
├── pyproject.toml              Python-Projektdatei (uv/mkdocs)
├── mkdocs.yml                  MkDocs-Konfiguration
├── justfile                    Task-Runner-Definitionen
├── backend/
│   ├── config/                 TOML-Konfigurationsparsung
│   ├── database/               Datenzugriffsschicht
│   ├── handlers/               Domain-Handler-Funktionen
│   ├── io/                     Excel-Import + PDF-Generierung
│   ├── models/                 Gemeinsame Datenstrukturen
│   └── services/               Business-Logik
├── frontend/
│   ├── index.html              Haupt-UI
│   ├── app.js                  Orchestrator-Modul
│   ├── admin/                  Datei-Handler, Config-Editor
│   ├── evaluations/            Ranking-UI
│   ├── groups/                 Gruppen-Anzeige
│   ├── reports/                PDF-Wrapper
│   ├── shared/                 dom.js, utils.js, Styles
│   └── stations/               Ergebniseingabe-UI
├── assets/mkdocs/              Dokumentationsquellen (MkDocs)
├── docs/                       Gebaubte Dokumentation (GitHub Pages)
└── test/                       Integrationstests
```

## Neues Backend-Feature hinzufügen

1. Struct/Typ in `backend/models/types.go` ergänzen.
2. DB-Query/-Insert in `backend/database/` hinzufügen.
3. Business-Logik in `backend/services/` hinzufügen (optional).
4. Handler-Funktion in `backend/handlers/` ergänzen.
5. Dünnen Wrapper in `app_handlers.go` exponieren.
6. Frontend-Aufruf via `window.go.main.App.MeineMethode()`.
7. Tests in `test/` schreiben.
