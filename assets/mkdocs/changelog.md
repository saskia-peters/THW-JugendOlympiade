# Changelog

Alle wesentlichen Änderungen an diesem Projekt sind hier dokumentiert.

Das Format basiert auf [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), das Projekt folgt [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.2] — 2026-03-23

### Geändert

- Alle Vorkommen von „Teilnehmer" in „Teilnehmende" umbenannt — Anzeigetexte, Schaltflächenbeschriftungen, Code-Bezeichner, Go-Structs und Funktionsnamen.
- Datenbanktabelle `teilnehmer` in `teilnehmende` umbenannt; alle SQL-Abfragen und FK-Referenzen aktualisiert.
- Excel-Import-Blattname von `Teilnehmer` auf `Teilnehmende` geändert.
- Schaltflächenbeschriftung „Teilnehmer zu Gruppen" in „Gruppen zusammenstellen" umbenannt.
- Ergebniseingabe-Modus blendet jetzt alle drei Button-Bereichsspalten aus.

---

## [0.1.1] — 2026-03-13

### Geändert

- Redundante `rel_tn_grp`-Tabelle in `gruppe` konsolidiert — alle Abfragen, Inserts und Indizes verwenden ausschließlich `gruppe`.
- `gruppe`-Tabellendefinition verschärft: `group_id` und `teilnehmer_id` sind jetzt `NOT NULL`; `teilnehmer_id` hat einen `UNIQUE`-Constraint; der Foreign Key referenziert korrekt `teilnehmer(teilnehmer_id)`.
- Teilnehmer-Urkunden: „Jugendolympiade"-Überschrift 1,5 cm tiefer; Abstand zwischen Überschrift und Jahr reduziert.
- Teilnehmer-Urkunden: Platzierungstext vergrößert (Größe 22, fett) und gold hervorgehoben.
- Teilnehmer-Urkunden: Abstand zwischen Platzierung und Mitgliedertabelle reduziert.

### Entfernt

- `rel_tn_grp`-Tabelle (Duplikat von `gruppe`); doppeltes Schreiben beim Gruppen-Speichern eliminiert.

### Behoben

- Foreign Key auf `gruppe.teilnehmer_id` referenzierte die falsche Spalte (`teilnehmer.id` statt `teilnehmer.teilnehmer_id`).
- FK-Enforcement war nie aktiv — `PRAGMA foreign_keys = ON` wird jetzt bei jeder Datenbankverbindung gesetzt.
- `teilnehmer.teilnehmer_id` fehlte `UNIQUE`-Constraint, was FK-Fehler beim Excel-Reload verursachte.
- Ungültiger FK auf `group_station_scores.group_id` entfernt.
- Ortsverband-Auswertung: Teilnehmendenzahl durch Stationsanzahl aufgebläht — behoben mit `COUNT(DISTINCT teilnehmer_id)`.
- Teilnehmer-Urkunden: „Gruppenmitglieder"-Label hatte überschüssigen Doppelpunkt und war linksbündig statt zentriert.
- Teilnehmer-Urkunden: Tabellenzeilen drifteten links aus dem Inhaltsbereich — `SetX(contentLeft)` jetzt je Zeile gesetzt.
- Teilnehmer-Urkunden: Linker Rand des Inhaltsbereichs von 5 mm auf 10 mm angepasst.

---

## [0.1.0] — 2026-03-13

### Hinzugefügt

- Erstveröffentlichung von Jugendolympiade Verwaltung.
- Gruppenverwaltung (Gruppen) mit Tab-Ansicht.
- Ergebniseingabe mit gruppen-erstem Arbeitsablauf — Gruppe auswählen, alle Stationen in Tabelle eingeben.
- Einzel- und Massenspeicherung für Stationsergebnisse; bestehende Ergebnisse vorausgefüllt.
- Schnellnavigations-Schaltfläche in Gruppen-Ansicht direkt zur Ergebniseingabe.
- Gruppen-Auswertung (Gruppenwertung) und Ortsverband-Auswertung.
- PDF-Generierung für Gruppen-Auswertung, OV-Auswertung und Urkunden.
- Datenbank-Backup-Funktion.
- Datenbank-Wiederherstellung aus Backup (neueste zuerst sortiert).
- Frontend nach Features gegliedert (shared, admin, groups, stations, evaluations, reports).
