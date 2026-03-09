# XLSX to SQLite Importer

A Go application that reads an XLSX file with a "Teilnehmer" sheet (4 columns) and imports the data into a SQLite database.

## Features

- Reads XLSX files using the excelize library
- Imports data from a sheet named "Teilnehmer"
- Stores data in a SQLite database
- Handles 4 columns from the Excel sheet
- Uses transactions for efficient bulk inserts
- **Automatically creates balanced groups** with at most 8 participants per group
- **Smart distribution algorithm** that balances groups by:
  - Ortsverband (location/district)
  - Alter (age)
  - Geschlecht (gender)
- **Generates PDF report** in A4 portrait format with one group per page
  - Professional table layout
  - Group statistics (distribution by Ortsverband, Geschlecht, and average age)
  - Alternating row colors for readability

## Requirements

- Go 1.21 or later
- GCC (for building SQLite driver on Windows, use MinGW or TDM-GCC)

## Installation

1. Initialize the Go module and download dependencies:
```bash
go mod download
```

## Usage

1. Place your XLSX file named `data.xlsx` in the same directory as the application
2. The XLSX file must contain a sheet named "Teilnehmer" with 4 columns
3. Run the application:
```bash
go run main.go
```

The application will:
- Read the `data.xlsx` file
- Create/update a SQLite database file named `data.db`
- Import all rows from the "Teilnehmer" sheet into the `teilnehmer` table
- The first row is assumed to be the header and will be skipped
- Create balanced groups using the distribution algorithm
- Generate a PDF report named `groups_report.pdf` with one group per page

## Database Schema

The application creates five tables with the following structure:

```mermaid
erDiagram
    teilnehmer ||--o{ gruppe : "has"
    teilnehmer ||--o| rel_tn_grp : "assigned to"
    gruppe ||--o{ rel_tn_grp : "contains"
    gruppe ||--o{ group_station_scores : "visits"
    stations ||--o{ group_station_scores : "scored by"

    teilnehmer {
        INTEGER id PK "Auto-increment"
        INTEGER teilnehmer_id "Sequential participant ID"
        TEXT name "Participant name"
        TEXT ortsverband "Location/District"
        INTEGER age "Age"
        TEXT geschlecht "Gender"
    }

    gruppe {
        INTEGER id PK "Auto-increment"
        INTEGER group_id "Group identifier"
        INTEGER teilnehmer_id FK "References teilnehmer.id"
    }

    rel_tn_grp {
        INTEGER id PK "Auto-increment"
        INTEGER teilnehmer_id UK "References teilnehmer.teilnehmer_id"
        INTEGER group_id FK "References gruppe.group_id"
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

**teilnehmer table:**
```sql
CREATE TABLE teilnehmer (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    teilnehmer_id INTEGER,
    name TEXT,
    ortsverband TEXT,
    age INTEGER,
    geschlecht TEXT
);
```

The columns map to:
- **id**: Auto-incremented internal database ID
- **teilnehmer_id**: Sequential participant ID (based on row number)
- **name**: NAME (Column 1) - Text
- **ortsverband**: ORTSVERBAND (Column 2) - Text
- **age**: ALTER (Column 3) - Integer
- **geschlecht**: GESCHLECHT (Column 4) - Text

**gruppe table:**
```sql
CREATE TABLE gruppe (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER,
    teilnehmer_id INTEGER,
    FOREIGN KEY (teilnehmer_id) REFERENCES teilnehmer(id)
);
```

The gruppe table links groups to participants (Teilnehmer).

**rel_tn_grp table:**
```sql
CREATE TABLE rel_tn_grp (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    teilnehmer_id INTEGER UNIQUE NOT NULL,
    group_id INTEGER NOT NULL,
    FOREIGN KEY (teilnehmer_id) REFERENCES teilnehmer(teilnehmer_id),
    FOREIGN KEY (group_id) REFERENCES gruppe(group_id)
);
```

The rel_tn_grp table is a relationship table that connects teilnehmer to gruppe:
- Each teilnehmer can appear only once (UNIQUE constraint on teilnehmer_id)
- Each group_id can appear multiple times (many participants can be in the same group)

**stations table:**
```sql
CREATE TABLE stations (
    station_id INTEGER PRIMARY KEY AUTOINCREMENT,
    station_name TEXT NOT NULL
);
```

The stations table stores information about different stations/locations.

**group_station_scores table:**
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

The group_station_scores table records scores for each group at each station:
- Each group-station combination can only have one score (UNIQUE constraint)
- Enables ranking by group and by ortsverband

## Grouping Algorithm

The application automatically creates balanced groups after importing the data. The algorithm:

1. **Calculates optimal group count**: Divides participants into groups of at most 8 members
2. **Sorts participants** by ortsverband, geschlecht, and alter for better initial distribution
3. **Uses diversity scoring** to assign each participant to the most suitable group:
   - Penalizes groups that already have many participants from the same Ortsverband
   - Penalizes groups that already have many participants of the same Geschlecht
   - Considers age (Alter) distribution to avoid clustering similar ages
   - Prefers groups with fewer members to balance group sizes
4. **Populates both tables**: Inserts records into both `gruppe` and `rel_tn_grp` tables

This ensures that groups are diverse and balanced across all three criteria.

## Output Files

After running the application, you will have:

1. **data.db** - SQLite database containing:
   - `teilnehmer` table with all participant data
   - `gruppe` table with group assignments
   - `rel_tn_grp` relationship table

2. **groups_report.pdf** - PDF report showing:
   - One group per page in A4 portrait format
   - Participant list with Name, Ortsverband, Alter, and Geschlecht
   - Group statistics showing distribution across the three criteria
   - Average age for each group

## Configuration

You can modify these constants in `main.go`:

- `dbFile`: SQLite database filename (default: "data.db")
- `xlsxFile`: Input XLSX filename (default: "data.xlsx")
- `sheetName`: Sheet name to read (default: "Teilnehmer")
- `tableName`: Database table name (default: "teilnehmer")
- `maxGroupSize`: Maximum participants per group (default: 8)

## Building

To build a standalone executable:
```bash
go build -o xlsx-importer.exe
```
