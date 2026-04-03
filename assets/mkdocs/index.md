# Jugendolympiade Verwaltung

<div class="hero-banner" markdown>
Eine plattformübergreifende Desktop-Anwendung für die Organisation von Jugendolympiaden — gebaut mit [Wails v2](https://wails.io/) (Go-Backend + Web-Frontend).
</div>

## Was die Anwendung leistet

<div class="grid cards" markdown>

-   :material-file-excel: **Datenverwaltung**

    ---

    Import von Teilnehmenden und Betreuenden aus Excel. Automatische Validierung sichert saubere Daten.

    [:octicons-arrow-right-24: Excel-Import](user-guide/excel-import.md)

-   :material-account-group: **Gruppenverteilung**

    ---

    Erstellt automatisch ausgewogene Gruppen nach Ortsverband, Alter und Geschlecht – mit Betreuenden-Zuweisung.

    [:octicons-arrow-right-24: Gruppen](user-guide/groups.md)

-   :material-scoreboard: **Ergebniseingabe**

    ---

    Ergebnisse je Gruppe und Station erfassen. Echtzeit-Aktualisierung mit konfigurierbaren Grenzwerten.

    [:octicons-arrow-right-24: Ergebniseingabe](user-guide/scoring.md)

-   :material-trophy: **Auswertung & Urkunden**

    ---

    Gruppen- und Ortsverband-Rankings. PDF-Urkunden für alle Teilnehmenden.

    [:octicons-arrow-right-24: Auswertung](user-guide/evaluations.md)

</div>

## Schnellstart

1. Excel-Datei vorbereiten (Blätter: `Teilnehmende`, `Betreuende`, `Stationen`)
2. `THW-JugendOlympiade.exe` starten
3. **📝 Daten → "Excel einlesen"** — Datei auswählen
4. **"Gruppen zusammenstellen"** klicken
5. Ergebnisse über **"Ergebniseingabe"** erfassen
6. PDFs über **📊 Ausgabe** erzeugen

[:octicons-arrow-right-24: Vollständige Anleitung](getting-started.md)

## Plattformen

| Plattform | Status |
|-----------|--------|
| Windows | ✅ Vollständig unterstützt |
| macOS | ✅ Vollständig unterstützt |
| Linux | ✅ Vollständig unterstützt |

## Tech Stack

| Komponente | Technologie |
|-----------|------------|
| Backend | Go 1.21+ |
| Desktop-Framework | Wails v2 |
| Datenbank | SQLite (modernc pure-Go) |
| PDF-Generierung | fpdf |
| Excel-Verarbeitung | excelize v2 |
| Frontend | Vanilla JS (ES6-Module) |
| Konfiguration | TOML |
