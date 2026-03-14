# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-03-13

### Changed
- Consolidated redundant `rel_tn_grp` table into `gruppe` — all queries, inserts, and indexes now use `gruppe` exclusively
- `gruppe` table definition tightened: `group_id` and `teilnehmer_id` are now `NOT NULL`, `teilnehmer_id` has a `UNIQUE` constraint, and the foreign key correctly references `teilnehmer(teilnehmer_id)`
- Participant certificates: "Jugendolympiade" heading moved 1.5cm lower; gap between heading and year reduced
- Participant certificates: rank text enlarged (size 22, bold) and highlighted in gold for better visibility
- Participant certificates: spacing between rank and group members table reduced

### Removed
- `rel_tn_grp` table (duplicate of `gruppe`); double-write on group save eliminated

### Fixed
- Foreign key on `gruppe.teilnehmer_id` previously referenced the wrong column (`teilnehmer.id` instead of `teilnehmer.teilnehmer_id`)
- Foreign key enforcement was never active — `PRAGMA foreign_keys = ON` is now set on every database connection (initial open and after restore)
- `teilnehmer.teilnehmer_id` missing `UNIQUE` constraint caused FK mismatch error on Excel reload with FK enforcement enabled
- Invalid FK on `group_station_scores.group_id` referencing non-unique `gruppe(group_id)` removed; integrity maintained at application level
- Ortsverband evaluation Teilnehmer count inflated by number of stations — fixed with `COUNT(DISTINCT teilnehmer_id)`
- Participant certificates: "Gruppenmitglieder" label had stray colon and was left-aligned instead of centered
- Participant certificates: table rows drifted left of content area — `SetX(contentLeft)` now applied per row
- Participant certificates: content area left margin adjusted from 5mm to 10mm

## [0.1.0] - 2026-03-13

### Added
- Initial release of Jugendolympiade Verwaltung
- Group management (Gruppen) with tabbed view
- Results entry (Ergebniseingabe) with group-first workflow — select a group to view all stations in a table with individual score inputs (100–1200)
- Individual and bulk save for station scores; existing scores pre-populated on load
- Quick navigation button in Gruppen view to jump directly to Ergebniseingabe for a selected group
- Group evaluation (Gruppenwertung) and Ortsverband evaluation views
- PDF generation for group evaluation, Ortsverband evaluation, and certificates
- Database backup functionality
- Database restore from backup with newest-first sorted list
- Frontend structured by feature (shared, admin, groups, stations, evaluations, reports)
