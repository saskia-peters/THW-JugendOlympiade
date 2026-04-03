# Datenbankschema

Die Anwendung verwendet **SQLite** über den `modernc.org/sqlite`-Pure-Go-Treiber (kein CGO).

## Entity-Relationship-Diagramm

```mermaid
erDiagram
    teilnehmende ||--o{ gruppe : "zugeordnet"
    betreuende ||--o{ group_betreuende : "zugeordnet"
    gruppe ||--o{ group_betreuende : "hat"
    gruppe ||--o{ group_station_scores : "bewertet bei"
    stations ||--o{ group_station_scores : "bewertet durch"

    teilnehmende {
        INTEGER id PK "Auto-Inkrement"
        INTEGER teilnehmer_id "Fortlaufende ID (UNIQUE)"
        TEXT name "Vollständiger Name"
        TEXT ortsverband "Lokale Gliederung"
        INTEGER age "Alter"
        TEXT geschlecht "Geschlecht"
        TEXT pregroup "Optionaler Gruppierschlüssel"
    }

    betreuende {
        INTEGER id PK "Auto-Inkrement"
        TEXT name "Name"
        TEXT ortsverband "Lokale Gliederung"
        INTEGER fahrerlaubnis "1=ja / 0=nein"
    }

    gruppe {
        INTEGER id PK "Auto-Inkrement"
        INTEGER group_id "Logische Gruppennummer"
        INTEGER teilnehmer_id FK,UNIQUE "Ref. teilnehmende.teilnehmer_id"
    }

    group_betreuende {
        INTEGER id PK "Auto-Inkrement"
        INTEGER group_id FK "Ref. gruppe.group_id"
        INTEGER betreuenden_id FK "Ref. betreuende.id"
    }

    stations {
        INTEGER station_id PK "Auto-Inkrement"
        TEXT station_name "Stationsname"
    }

    group_station_scores {
        INTEGER id PK "Auto-Inkrement"
        INTEGER group_id FK "Ref. gruppe.group_id"
        INTEGER station_id FK "Ref. stations.station_id"
        INTEGER score "Punktzahl"
    }
```

## Tabellendetails

### `teilnehmende` — Teilnehmende

```sql
CREATE TABLE teilnehmende (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    teilnehmer_id  INTEGER UNIQUE NOT NULL,
    name           TEXT,
    ortsverband    TEXT,
    age            INTEGER,
    geschlecht     TEXT,
    pregroup       TEXT
);
CREATE INDEX idx_teilnehmende_id ON teilnehmende(teilnehmer_id);
```

`teilnehmer_id` ist eine fortlaufende ID aus dem Excel-Import (1, 2, 3, …). Die `UNIQUE`-Constraint ist für den Foreign Key in `gruppe` erforderlich.

---

### `betreuende` — Betreuende

```sql
CREATE TABLE betreuende (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    name           TEXT,
    ortsverband    TEXT,
    fahrerlaubnis  INTEGER   -- 1 = ja, 0 = nein
);
```

---

### `gruppe` — Gruppenverteilung

```sql
CREATE TABLE gruppe (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id       INTEGER NOT NULL,
    teilnehmer_id  INTEGER UNIQUE NOT NULL,
    FOREIGN KEY (teilnehmer_id) REFERENCES teilnehmende(teilnehmer_id)
);
CREATE INDEX idx_gruppe_group_id      ON gruppe(group_id);
CREATE INDEX idx_gruppe_teilnehmer_id ON gruppe(teilnehmer_id);
```

Mehrere Zeilen teilen dieselbe `group_id` — eine Zeile je Teilnehmenden pro Gruppe. `UNIQUE(teilnehmer_id)` erzwingt genau eine Gruppe pro Person.

---

### `group_betreuende` — Gruppen-Betreuenden-Zuordnung

```sql
CREATE TABLE group_betreuende (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id        INTEGER NOT NULL,
    betreuenden_id  INTEGER NOT NULL,
    FOREIGN KEY (group_id)       REFERENCES gruppe(group_id),
    FOREIGN KEY (betreuenden_id) REFERENCES betreuende(id)
);
```

---

### `stations` — Stationen

```sql
CREATE TABLE stations (
    station_id    INTEGER PRIMARY KEY AUTOINCREMENT,
    station_name  TEXT NOT NULL
);
```

---

### `group_station_scores` — Ergebnisse

```sql
CREATE TABLE group_station_scores (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id    INTEGER NOT NULL,
    station_id  INTEGER NOT NULL,
    score       INTEGER,
    FOREIGN KEY (group_id)   REFERENCES gruppe(group_id),
    FOREIGN KEY (station_id) REFERENCES stations(station_id),
    UNIQUE(group_id, station_id)
);
CREATE INDEX idx_scores_group_id   ON group_station_scores(group_id);
CREATE INDEX idx_scores_station_id ON group_station_scores(station_id);
```

`UNIQUE(group_id, station_id)` verhindert doppelte Einträge. `INSERT OR REPLACE` aktualisiert bestehende Ergebnisse.

## Pragmas

Jede Verbindung öffnet mit:

```sql
PRAGMA foreign_keys = ON;
```

SQLite deaktiviert FK-Enforcement standardmäßig — ohne diesen Pragma werden Constraints lautlos ignoriert.

## Backup & Wiederherstellung

Backups werden in `dbbackups/` gespeichert — einfache SQLite-Dateien mit Zeitstempel-Suffix. Der Wiederherstellungs-Dialog zeigt sie neueste-zuerst sortiert. Nach der Wiederherstellung wird eine neue Verbindung mit aktiviertem FK-Enforcement geöffnet.
