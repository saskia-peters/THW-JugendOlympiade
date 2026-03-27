# Developer Documentation - Jugendolympiade Verwaltung

Technical documentation for developers working on the Jugendolympiade Verwaltung application.

**User Documentation**: See [README.md](README.md) for end-user instructions.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![Wails](https://img.shields.io/badge/Wails-v2-red)
![SQLite](https://img.shields.io/badge/SQLite-3-003B57?logo=sqlite)

## Table of Contents

- [Development Requirements](#development-requirements)
- [Project Structure](#project-structure)
- [Architecture](#architecture)
- [Database Schema](#database-schema)
- [Algorithms](#algorithms)
- [Testing](#testing)
- [Security & Performance](#security--performance)
- [Development Setup](#development-setup)
- [Building](#building)
- [Configuration](#configuration)
- [Contributing](#contributing)

## Development Requirements

### Core Dependencies
- **Go**: 1.21 or later
- **Wails CLI**: v2.x (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- **GCC**: For CGO/SQLite compilation
  - Windows: [MinGW](https://www.mingw-w64.org/) or [TDM-GCC](https://jmeubank.github.io/tdm-gcc/)
  - macOS: Xcode Command Line Tools (`xcode-select --install`)
  - Linux: build-essential (`sudo apt-get install build-essential`)
- **Node.js**: Not required (frontend is vanilla JS)

### Go Modules
```bash
go mod download
```

Key dependencies automatically installed:
- `github.com/wailsapp/wails/v2` - Desktop framework
- `github.com/xuri/excelize/v2` - Excel processing
- `github.com/jung-kurt/gofpdf` - PDF generation
- `modernc.org/sqlite` - Pure Go SQLite driver
- `github.com/BurntSushi/toml` - TOML config parsing

## Project Structure

```
THW-JugendOlympiade/
├── backend/                  # Go backend code
│   ├── config/              # Configuration management
│   │   └── config.go        # TOML config struct, LoadOrCreate(), ReadRaw(), ValidateAndSave()
│   ├── database/            # Database layer
│   │   ├── db.go           # DB initialization and connection
│   │   ├── queries.go      # Optimized read queries
│   │   ├── inserts.go      # Write operations (incl. SaveGroupBetreuende)
│   │   ├── evaluations.go  # Ranking and evaluation queries
│   │   └── scores.go       # Score management
│   ├── io/                 # Input/Output operations
│   │   ├── input.go        # Excel import (Teilnehmende + Betreuende + Stationen)
│   │   ├── output.go       # Shared PDF helpers (enc(), directory setup)
│   │   ├── pdf_groups.go   # Groups + Betreuende report PDF
│   │   ├── pdf_styles.go   # Shared PDF style helpers
│   │   ├── pdf_evaluations.go         # Evaluation PDFs (group + ortsverband)
│   │   ├── pdf_cert_teilnehmende.go   # Participant certificates (one per person)
│   │   └── pdf_cert_ortsverbaende.go  # OV certificates (Siegerurkunde + Urkunde)
│   ├── models/             # Data models
│   │   └── types.go        # Structs: Teilnehmende, Betreuende (incl. Fahrerlaubnis), Group, Station …
│   └── services/           # Business logic
│       └── distribution.go # Participant + Betreuende distribution algorithm
├── frontend/               # Web frontend (vanilla ES6 modules)
│   ├── index.html         # Main UI structure
│   ├── app.js             # Orchestrator: imports modules, loads config, exposes to window
│   ├── admin/
│   │   ├── file-handler.js # File load, backup, restore, group distribution
│   │   └── config-editor.js # In-app TOML config editor modal
│   ├── groups/
│   │   ├── groups.js       # Group display with tabs, Betreuende + statistics
│   │   └── groups.css
│   ├── stations/
│   │   ├── stations.js     # Results entry, dirty-tracking, Eingabeübersicht matrix
│   │   ├── scores.js       # Legacy per-station score assignment helpers
│   │   └── stations.css
│   ├── evaluations/
│   │   ├── evaluations.js  # Group and ortsverband rankings
│   │   └── evaluations.css
│   ├── reports/
│   │   └── pdf-handlers.js # PDF generation wrappers
│   └── shared/
│       ├── dom.js          # DOM refs, setStatus(), clearAllTabs(), setEvalButtonsEnabled()
│       ├── utils.js        # escapeHtml(), switchTab()
│       ├── styles.css      # Global styles
│       └── components.css  # Shared component styles
├── build/                  # Build outputs
│   ├── bin/               # Compiled executables
│   └── windows/           
│       └── icon.ico       # Windows icon (ICO)
├── dev_utils/              # Development utilities
│   ├── convert_icon.ps1   # PowerShell icon converter
│   ├── convert_icon.py    # Python icon converter
│   └── README.md          # Utility documentation
├── pdfdocs/                # PDF outputs (auto-created at runtime)
├── test/                   # Unit and integration tests
│   ├── database_test.go   # Database operation tests
│   ├── distribution_test.go # Group distribution tests
│   ├── input_test.go      # Excel import validation tests
│   ├── models_test.go     # Data model tests
│   ├── scores_test.go     # Score assignment tests
│   └── README.md          # Test documentation
├── main.go                 # Application entry point & Wails setup
├── config.toml             # Runtime configuration (auto-created on first launch)
├── wails.json             # Wails build configuration
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── README.md              # User documentation
└── documentation/         # Extended documentation
```

### Backend Architecture

**Layered Architecture:**
1. **main.go**: Entry point, App struct, startup/shutdown hooks, Wails bootstrap
2. **app_handlers.go**: Thin `App` method wrappers — delegates every bound method to `backend/handlers/`
3. **backend/handlers**: Domain handler functions (no Wails dependency)
4. **backend/database**: Data access layer (DAL)
5. **backend/services**: Business logic layer
6. **backend/io**: File I/O operations
7. **backend/models**: Shared data structures

**Key Design Patterns:**
- **App-scoped DB Connection**: `a.db` on the App struct (thread-safe, testable)
- **Transaction Wrapping**: Bulk operations use transactions
- **Error Propagation**: Explicit error handling throughout

## Architecture

### Application Flow

```
User Interaction (Frontend)
    ↓
Wails Runtime Binding
    ↓
Go Backend Methods (main.go)
    ↓
Service Layer (business logic)
    ↓
Database Layer (queries, inserts)
    ↓
SQLite Database (data.db)
```

### Frontend ↔ Backend Communication

The frontend communicates with Go backend via Wails runtime:

```javascript
// Frontend (app.js)
const result = await window.go.main.App.LoadFile();
```

```go
// Backend (app_handlers.go) — thin wrapper
func (a *App) LoadFile() map[string]interface{} {
    return handlers.LoadFile(a.ctx, &a.db)
}

// Backend (backend/handlers/files.go) — actual logic
func LoadFile(ctx context.Context, db **sql.DB) map[string]interface{} {
    // Implementation
}
```

All public methods on `App` struct are automatically bound.

## Database Schema

The application uses SQLite with four main tables:

### Entity-Relationship Diagram

```mermaid
erDiagram
    teilnehmende ||--o{ gruppe : "has"
    betreuende ||--o{ group_betreuende : "assigned to"
    gruppe ||--o{ group_betreuende : "has"
    gruppe ||--o{ group_station_scores : "visits"
    stations ||--o{ group_station_scores : "scored by"

    teilnehmende {
        INTEGER id PK "Auto-increment"
        INTEGER teilnehmer_id "Sequential participant ID"
        TEXT name "Participant name"
        TEXT ortsverband "Location/District"
        INTEGER age "Age"
        TEXT geschlecht "Gender"
        TEXT pregroup "Optional pre-group code"
    }

    betreuende {
        INTEGER id PK "Auto-increment"
        TEXT name "Supervisor name"
        TEXT ortsverband "Location/District"
        INTEGER fahrerlaubnis "1 = ja, 0 = nein"
    }

    gruppe {
        INTEGER id PK "Auto-increment"
        INTEGER group_id "Group identifier"
        INTEGER teilnehmer_id FK "References teilnehmende.teilnehmer_id"
    }

    group_betreuende {
        INTEGER id PK "Auto-increment"
        INTEGER group_id FK "References gruppe.group_id"
        INTEGER betreuenden_id FK "References betreuende.id"
    }

    stations {
        INTEGER station_id PK "Auto-increment"
        TEXT station_name "Station name"
    }

    group_station_scores {
        INTEGER id PK "Auto-increment"
        INTEGER group_id FK "References gruppe.group_id"
        INTEGER station_id FK "References stations.station_id"
        INTEGER score "Score value"
    }
```

### Table Details

#### teilnehmende (Participants)
Primary participant data table.

```sql
CREATE TABLE teilnehmende (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    teilnehmer_id INTEGER,
    name TEXT,
    ortsverband TEXT,
    age INTEGER,
    geschlecht TEXT
);
```

- `id`: Internal auto-increment key
- `teilnehmer_id`: Sequential ID based on import order (1, 2, 3, ...)
- `name`, `ortsverband`, `age`, `geschlecht`: Data from Excel import

#### betreuende (Supervisors)
Supervisor data imported from the `Betreuende` sheet.

```sql
CREATE TABLE betreuende (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    ortsverband TEXT,
    fahrerlaubnis INTEGER  -- 1 = ja, 0 = nein
);
```

#### group_betreuende (Group–Supervisor Assignment)
Links supervisors to groups.

```sql
CREATE TABLE group_betreuende (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    betreuenden_id INTEGER NOT NULL,
    FOREIGN KEY (group_id) REFERENCES gruppe(group_id),
    FOREIGN KEY (betreuenden_id) REFERENCES betreuende(id)
);
```

- Cleared and re-written each time `DistributeGroups()` is called.

#### gruppe (Groups)
Group-participant assignment table.

```sql
CREATE TABLE gruppe (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    teilnehmer_id INTEGER UNIQUE NOT NULL,
    FOREIGN KEY (teilnehmer_id) REFERENCES teilnehmende(teilnehmer_id)
);
```

- `group_id`: Logical group number (multiple rows share the same `group_id` for one group)
- `teilnehmer_id`: UNIQUE — enforces one group per participant
- Cleared and re-written each time `DistributeGroups()` is called

#### stations (Activity Stations)
List of stations/activities for scoring.

```sql
CREATE TABLE stations (
    station_id INTEGER PRIMARY KEY AUTOINCREMENT,
    station_name TEXT NOT NULL
);
```

#### group_station_scores (Performance Tracking)
Records group performance at each station.

```sql
CREATE TABLE group_station_scores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    station_id INTEGER NOT NULL,
    score INTEGER,
    FOREIGN KEY (group_id) REFERENCES gruppe(group_id),
    FOREIGN KEY (station_id) REFERENCES stations(station_id),
    UNIQUE(group_id, station_id)
);
```

- UNIQUE constraint prevents duplicate scores for same group-station combination
- Used for rankings and evaluations

## Algorithms

### Group Distribution Algorithm

**Location**: `backend/services/distribution.go`

**Objective**: Create balanced groups with maximum diversity across Ortsverband, gender, and age, then assign Betreuende observing Fahrerlaubnis and OV-cohesion rules.

#### Participant Distribution

1. **Validate Pre-Groups** — any `PreGroup` that exceeds `maxGroupSize` is rejected with a clear error listing the offending group names.

2. **Calculate Group Count**
   ```go
   numPreGroups + ceil(len(unassigned) / maxGroupSize)
   ```

3. **Place Pre-Groups** — participants sharing a `PreGroup` code are placed into dedicated groups first.

4. **Diversity Scoring** — remaining participants are sorted and assigned to the group with the best diversity score:
   ```go
   score += float64(ortsverbandCount) * 2.0  // penalise same OV
   score += float64(geschlechtCount)  * 1.5  // penalise same gender
   score += 1.0  // penalise if age close to group average
   ```
   Prefer the group with the fewest members (size bonus 0.5 per member).

#### Betreuende Distribution (four-phase)

1. **Phase 1 — Spread licensed drivers (Fahrerlaubnis=ja) one-per-group.**  
   Uses `findGroupForLicensed()`: picks the group with the fewest licensed drivers already assigned (OV participant count breaks ties). This guarantees no group receives a second licensed driver before every group has at least one.

2. **Phase 2 — Unlicensed Betreuende follow their OV.**  
   Uses `findGroupForUnlicensed()`: priority order:
   - Group that already has a licensed driver from the same OV.
   - Group that has any Betreuende from the same OV.
   - Group with the fewest Betreuende overall.

3. **Phase 2b — Rebalance unlicensed members evenly.**  
   Iteratively moves one unlicensed member from the most-loaded group to the least-loaded until the difference in total Betreuende count is ≤ 1. Only unlicensed members are moved to preserve the Phase 1 guarantee.

4. **Phase 3 — Safety net.**  
   Any group still without a Betreuende steals one from the group with the largest surplus (unlicensed preferred, to keep licensed drivers where possible).

**Warnings** (returned as `string` from `CreateBalancedGroups`):  
- Groups with no Betreuende at all (shows group numbers).  
- Groups with no licensed driver (shows group numbers).

**Time Complexity**: O(n·g) for participants; O(b·g) for Betreuende (b = Betreuende count, g = group count).

### Evaluation Queries

**Group Evaluation** (`backend/database/evaluations.go`):
- Sums scores across all stations per group
- Orders by total score descending
- Uses single JOIN query (optimized)

**Ortsverband Evaluation**:
- Calculates average score per participant by ortsverband
- Original query had N+1 issue (fixed)
- Current: Single query with aggregation

## Testing

### Test Organization

Tests are located in `test/` directory:

```
test/
├── database_test.go       # Database operation tests
├── distribution_test.go   # Group distribution (8 tests)
├── evaluations_test.go    # Evaluation query tests
├── input_test.go          # Excel import validation (10 tests)
├── inserts_test.go        # Insert operation tests
├── models_test.go         # Data model tests
├── queries_test.go        # Query tests
├── scores_test.go         # Score assignment tests
├── services_test.go       # Service layer tests
└── README.md              # Detailed test documentation
```

### Running Tests

```bash
cd test
go test -v                  # Run all tests with verbose output
go test -run TestName       # Run specific test
go test -cover              # Run with coverage report
```

### Test Suites

#### Excel Import Tests (`input_test.go`)
Tests for `backend/io/input.go`:

1. `TestReadXLSXFile_ValidFile` - Valid file import
2. `TestReadXLSXFile_InvalidPath` - Non-existent file handling
3. `TestReadXLSXFile_InvalidHeaders` - Wrong column headers
4. `TestReadXLSXFile_MissingRequiredField` - Empty required fields
5. `TestReadXLSXFile_InvalidAge` - Age validation (1-100 range)
6. `TestReadXLSXFile_NonNumericAge` - Age type validation
7. `TestReadXLSXFile_EmptySheet` - Empty sheet handling
8. `TestValidateHeaders_Valid` - Header validation (positive)
9. `TestValidateHeaders_Invalid` - Header validation (negative)
10. `TestValidateParticipantRow` - Row validation logic

**Coverage**: ~85% of `input.go`

#### Distribution Tests (`distribution_test.go`)
Tests for `backend/services/distribution.go`:

1. `TestDistribution_BasicFunctionality` - Basic distribution works
2. `TestDistribution_EmptyParticipants` - Empty input handling
3. `TestDistribution_GroupSizeLimit` - Max 8 per group
4. `TestDistribution_SingleParticipant` - Edge case: 1 participant
5. `TestDistribution_ExactlyMaxSize` - Edge case: exactly 8 participants
6. `TestDistribution_StatisticsTracking` - Stats accuracy
7. `TestDistribution_DiversityScoring` - Diversity algorithm validation
8. `TestDistribution_ConsistentOutput` - Deterministic results

**Coverage**: ~90% of `distribution.go`

### Writing New Tests

**Test File Template:**
```go
package test

import (
    "testing"
    "THW-JugendOlympiade/backend/services"
)

func TestNewFeature(t *testing.T) {
    // Arrange
    input := prepareTestData()
    
    // Act
    result := services.NewFeature(input)
    
    // Assert
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

**Best Practices:**
- Use table-driven tests for multiple scenarios
- Mock database connections when possible
- Clean up test files/databases after tests
- Test both happy path and error cases

## Security & Performance

### Security Features

#### ✅ SQL Injection Protection
All database queries use parameterized statements:

```go
// ❌ VULNERABLE (old code)
query := fmt.Sprintf("SELECT * FROM teilnehmende WHERE id = %d", id)

// ✅ SAFE (current code)
query := "SELECT * FROM teilnehmende WHERE id = ?"
db.Query(query, id)
```

**Fixed in PR**: SQL injection vulnerabilities resolved in 5 locations.

#### ✅ Input Validation
Comprehensive validation in `backend/io/input.go`:

```go
func validateParticipantRow(row []string, rowIndex int) error {
    // Name validation
    if strings.TrimSpace(row[0]) == "" {
        return fmt.Errorf("row %d: name is required", rowIndex)
    }
    
    // Age validation
    age, err := strconv.Atoi(strings.TrimSpace(row[2]))
    if err != nil {
        return fmt.Errorf("row %d: age must be a number", rowIndex)
    }
    if age < 1 || age > 100 {
        return fmt.Errorf("row %d: age must be between 1 and 100", rowIndex)
    }
    
    // ... more validation
}
```

#### ✅ Type Safety
Strong typing throughout Go codebase prevents type confusion attacks.

#### ✅ Error Handling
Proper error propagation with context:

```go
if err != nil {
    return fmt.Errorf("failed to initialize database: %w", err)
}
```

### Performance Optimizations

#### ✅ N+1 Query Prevention

**Problem**: Original code made multiple queries in loops.

**Solution**: Use JOINs and map-based aggregation.

**Example - GetGroupsForReport():**

Before (N+1 pattern):
```go
// 1 query for groups
for _, group := range groups {
    // N queries for participants (32 queries for 32 groups)
    participants := getParticipantsForGroup(group.ID)
}
```

After (optimized):
```go
// 1 query with JOIN
rows := db.Query(`
    SELECT g.group_id, t.* 
    FROM gruppe g
    JOIN teilnehmende t ON g.teilnehmer_id = t.teilnehmer_id
    ORDER BY g.group_id
`)

// Map-based aggregation in memory
groupMap := make(map[int][]Participant)
for rows.Next() {
    // ... scan and aggregate
}
```

**Impact**: 32 queries → 2 queries (93% reduction)

#### ✅ Transaction Usage

Bulk inserts use transactions:

```go
tx, _ := db.Begin()
for _, participant := range participants {
    tx.Exec("INSERT INTO ...")
}
tx.Commit()
```

**Impact**: 10x faster for large datasets

#### ✅ Efficient Algorithms

- Group distribution: O(n·g) = O(n²/8) ≈ O(n) for fixed max group size
- Sorting preprocessing: O(n log n)
- Memory usage: O(g) for group statistics

#### ✅ Resource Management

- Proper cleanup with `defer db.Close()`
- File handles closed after use
- PDF streams flushed and closed

### Known Performance Issues

1. **No Database Indexes on custom queries**: Base indexes exist (`idx_gruppe_group_id`, `idx_gruppe_teilnehmer_id`, `idx_scores_group_id`, `idx_scores_station_id`); add more if queries are extended.

## Development Setup

### 1. Install Prerequisites

**Go:**
```bash
# Download from https://go.dev/dl/
# Verify installation
go version  # Should show 1.21 or later
```

**Wails CLI:**
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor  # Verify installation
```

**GCC (Windows):**
```bash
# Install MinGW or TDM-GCC
# Add to PATH
gcc --version  # Verify installation
```

### 2. Clone Repository

```bash
git clone <repository-url>
cd THW-JugendOlympiade
go mod download
```

### 3. Development Mode

Launch with hot reload:

```bash
wails dev
```

This starts the app with:
- ✅ Frontend hot reload (HTML/CSS/JS changes)
- ✅ Backend recompilation on .go file changes
- ✅ DevTools enabled (F12)
- ✅ Console logging
- ✅ Faster startup (no packaging)

**Dev Mode Shortcuts:**
- `F5` - Reload frontend
- `F12` - Open DevTools
- `Ctrl+C` - Stop dev server

### 4. Make Changes

**Backend Changes:**
1. Edit Go files in `backend/` or `main.go`
2. Wails detects changes and recompiles
3. App restarts automatically

**Frontend Changes:**
1. Edit HTML/CSS/JS in `frontend/`
2. Browser reloads automatically
3. No restart needed

### 5. Test Changes

```bash
cd test
go test -v
```

Run specific test:
```bash
go test -v -run TestValidateHeaders
```

### 6. Build for Testing

```bash
wails build
./build/bin/THW-JugendOlympiade.exe  # Windows
./build/bin/THW-JugendOlympiade.app  # macOS
./build/bin/THW-JugendOlympiade      # Linux
```

## Building

### Production Builds

**Windows:**
```bash
wails build
# Output: build/bin/THW-JugendOlympiade.exe
```

**Windows (No Console):**
```bash
wails build -ldflags "-H windowsgui"
# Hides console window
```

**macOS (Universal Binary):**
```bash
wails build -platform darwin/universal
# Output: build/bin/THW-JugendOlympiade.app
# Supports Intel and Apple Silicon
```

**Linux:**
```bash
wails build -platform linux/amd64
# Output: build/bin/THW-JugendOlympiade
```

### Cross-Compilation

Build for multiple platforms:

```bash
# Build for Windows from macOS/Linux
wails build -platform windows/amd64

# Build for macOS from Windows/Linux
wails build -platform darwin/universal

# Build for Linux from Windows/macOS
wails build -platform linux/amd64
```

**Note**: Some platforms require CGO cross-compilation setup.

### Build Options

**Debug Build:**
```bash
wails build -debug
# Includes debug symbols, enables logging
```

**Compressed Build:**
```bash
wails build -upx
# Compresses with UPX (smaller executable)
```

**Custom Output:**
```bash
wails build -o myapp.exe
# Custom executable name
```

See `wails build --help` for all options.

## Configuration

### Runtime Configuration (`config.toml`)

A `config.toml` file is auto-created next to the executable on first launch. It is loaded by `backend/config/config.go` via `LoadOrCreate()` and stored in the `App.cfg` struct at startup.

```toml
[veranstaltung]
name = "THW-JugendOlympiade"  # Appears on PDFs and certificates
jahr = 2026

[gruppen]
max_groesse = 8  # Maximum participants per group

[ergebnisse]
min_punkte = 100   # Minimum score per station
max_punkte = 1200  # Maximum score per station

[ausgabe]
pdf_ordner = "pdfdocs"  # Output directory for generated PDFs
db_name = "data.db"     # SQLite database filename
urkunden_stil = "text"  # Participant certificate style: "text" (members table) or "picture" (group photo)
bilder_ordner = "pictures"  # Directory containing group photos (only used when urkunden_stil = "picture")
```

The in-app editor (Admin → "Konfiguration bearbeiten") calls `GetConfigRaw()` / `SaveConfigRaw()` to read and write this file with server-side TOML validation.

### Application Configuration (`wails.json`)
```json
{
  "name": "THW-JugendOlympiade",
  "outputfilename": "THW-JugendOlympiade",
  "frontend:install": "",
  "frontend:build": "",
  "info": {
    "companyName": "",
    "productName": "Jugendolympiade Verwaltung",
    "productVersion": "1.0.0",
    "copyright": "",
    "comments": ""
  }
}
```

**Modifiable:**
- `outputfilename`: Executable name
- `productName`: Shown in title bar
- `productVersion`: Update for releases
- `copyright`, `comments`: Metadata

### Code Configuration

**Database filename** (`backend/models/types.go`):
```go
var DbFile = "data.db"
```

All other previously hardcoded values (group size, score bounds, PDF directory, event name, certificate style) are now read from `config.toml` at runtime.

### Icons and Branding

**Icon Files:**
- `build/appicon.png` - Used by Wails for multiple platforms
- `build/windows/icon.ico` - Windows-specific icon

**Regenerate Icons:**

From custom logo (`logo_jo26_spiele.png`):

```bash
# Windows (PowerShell)
powershell -ExecutionPolicy Bypass -File dev_utils\convert_icon.ps1

# Cross-platform (Python with Pillow)
pip install Pillow
python dev_utils/convert_icon.py
```

After updating icons:
```bash
wails build  # Icons embedded in new build
```

See [dev_utils/README.md](dev_utils/README.md) for details.

## Contributing

### Workflow

1. **Fork and Clone**
   ```bash
   git clone <your-fork-url>
   cd THW-JugendOlympiade
   ```

2. **Create Feature Branch**
   ```bash
   git checkout -b feature/my-new-feature
   ```

3. **Make Changes**
   - Follow Go conventions
   - Add tests for new features
   - Update documentation

4. **Run Tests**
   ```bash
   cd test
   go test -v
   ```

5. **Format Code**
   ```bash
   go fmt ./...
   ```

6. **Commit**
   ```bash
   git commit -m "Add new feature: description"
   ```

7. **Push and PR**
   ```bash
   git push origin feature/my-new-feature
   # Create Pull Request on GitHub
   ```

### Code Style

**Go Conventions:**
- Use `gofmt` for formatting
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Add comments for exported functions
- Handle all errors explicitly

**Example:**
```go
// DoSomething performs an operation and returns an error if it fails.
// The input parameter must be non-nil.
func DoSomething(input *Data) error {
    if input == nil {
        return fmt.Errorf("input cannot be nil")
    }
    
    // Implementation
    
    return nil
}
```

### Testing Requirements

- **New Features**: Add tests in `test/` directory
- **Bug Fixes**: Add regression test
- **Coverage**: Aim for >80% coverage on new code
- **All Tests Pass**: `go test -v` must pass before PR

### Documentation

- Update [README.md](README.md) for user-facing changes
- Update this file for architectural changes
- Add inline comments for complex logic
- Update [test/README.md](test/README.md) for new tests

### Pull Request Checklist

- [ ] Code follows Go conventions
- [ ] All tests pass (`go test -v`)
- [ ] New tests added for new features
- [ ] Documentation updated
- [ ] No security vulnerabilities introduced
- [ ] Performance impact considered

## Additional Resources

- **Wails Documentation**: https://wails.io/docs/
- **Go Documentation**: https://go.dev/doc/
- **excelize Documentation**: https://xuri.me/excelize/
- **gofpdf Documentation**: https://github.com/jung-kurt/gofpdf

## Future Improvements

### Medium Priority
1. **Add Integration Tests**: Test full workflows end-to-end
2. **Add retry mechanisms**: For transient file/PDF failures (see architecture review)
3. **Add Logging Framework**: Better debugging

### Low Priority
4. **TypeScript Frontend**: Type safety for frontend
5. **Consider frontend framework**: Only if UI complexity grows significantly
6. **Remove empty `backend/handlers/` directory**: Minor cleanup

---

**Questions?** Open an issue or contact the maintainers.
