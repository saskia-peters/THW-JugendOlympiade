# Code Review: THW-JugendOlympiade

**Review Date:** 2026-03-24

---

## Summary

This is a well-structured local desktop app that covers the core competition-day workflow. The backend and PDF generation are solid. The main gaps fall into three practical tiers: things that could cause data loss or silent failures on competition day, UX friction points that would slow down operators, and smaller robustness improvements.

---

## Findings

### 🔴 Critical

**[handlers_files.go] No confirmation before re-distributing groups**
If groups already exist and the user clicks "Gruppen zusammenstellen" again (e.g. after re-loading the Excel), the existing distribution and all entered scores are silently wiped. A guard that checks whether scores exist and shows a destructive-action warning would prevent data loss.

**[backend/services/distribution.go] No overflow guard for oversized PreGroups**
If a PreGroup tag in the Excel file has more members than `max_groesse`, the algorithm will stuff them all into one group, breaking the size invariant silently with no error message. There should be a validation pass before distribution that rejects or splits oversized pre-groups.

**[backend/database/db.go] No SQLite integrity check after restore**
`RestoreDatabase()` copies the file byte-for-byte and immediately opens it, with no `PRAGMA integrity_check` or `PRAGMA quick_check`. A corrupted backup will appear to restore successfully but will fail on the first real query during competition.

**[backend/config/config.go] No semantic validation of config values**
`ValidateAndSave()` only validates TOML syntax. Values such as `max_groesse = 0`, `min_punkte = -500`, or `min_punkte > max_punkte` are silently accepted and will cause broken behavior at runtime (e.g. score entry rejects every legitimate score, or distribution produces zero-size groups).

**[backend/io/output.go] `pdf_ordner` not sanitised**
The `pdf_ordner` value from config is passed directly to `os.MkdirAll`. A malicious or accidental value like `../../Windows/System32/pdfs` would create unexpected directories. The value should be validated to be a relative path with no `..` components.

---

### 🟡 Suggestions — UX & Workflow

**[frontend/stations/stations.js] No client-side score bounds display**
The bounds (`min_punkte`, `max_punkte`) are available in `window.appConfig` but are not shown on the score input table. Operators should see the allowed range in the column header or as a placeholder, so they don't have to guess why a value is rejected.

**[frontend/stations/stations.js] No visual indicator for saved vs. unsaved scores**
After `AssignScore()` succeeds the row reverts to normal. On a busy competition day with many groups and stations, there is no persistent visual confirmation that a row has been saved (e.g. a green checkmark or row highlight that fades). This leads to re-entry of already-saved values.

**[frontend/stations/stations.js] Tab/Enter navigation between score inputs**
Score entry requires clicking each input individually. Pressing Enter or Tab should advance to the next station row. This is the most-used interaction on competition day and keyboard navigation would significantly speed up data entry.

**[frontend/evaluations/evaluations.js] No "all groups have scored" indicator**
There is no way to tell from the evaluation view whether every group has completed all stations. `GetGroupEvaluations()` returns a `station_count` per group — surfacing a warning when any group has fewer completed stations than the maximum would prevent publishing an incomplete ranking.

**[frontend/admin/file-handler.js] No preview of Excel data before import**
After selecting the XLSX file, data is immediately committed to the database. A preview step (participant count, detected headers, first 5 rows) before the final import would catch column mapping errors without having to reload.

**[handlers_backup.go] No automatic cleanup of old backups**
Backups accumulate indefinitely. A configurable retention count (e.g. keep the 10 most recent) or at least a "Delete backup" button in the restore dialog would prevent the backup folder growing without bound.

**[frontend/admin/config-editor.js] Config editor gives no hint about valid values**
The modal shows raw TOML with no inline documentation of allowed values or ranges. Either rendering the comments from `defaultTOML` (they explain each field) or showing a structured form editor would make this significantly safer for non-technical users.

---

### 🟡 Suggestions — PDF & Certificates

**[backend/io/pdf_cert_teilnehmende.go] No fallback message when template image is missing**
If `templates/background_urkunde_teilnehmende.png` is absent the programmatic layout is used silently. The user never learns that the background template was not applied. A status message ("Kein Zertifikat-Template gefunden, Layout wird automatisch erstellt.") would avoid confusion when the printed output looks different to expectations.

**[backend/io/pdf_evaluations.go] No "stations completed" column in group evaluation PDF**
The group ranking PDF currently shows Rank, Group ID, Station Count, and Total Score. For transparency it would be useful to flag groups where `station_count` is less than the total number of stations, so the printed result sheet clearly shows any incomplete entries.

**[backend/io/] PDF filenames hardcoded**
All output PDF names (`Gruppeneinteilung.pdf`, `Urkunden_Teilnehmende.pdf`, etc.) are hardcoded strings. If the user runs the app for two events (e.g. different years or locations using `db_name`), the PDFs overwrite each other. Prefixing with `veranstaltung.name` + `veranstaltung.jahr` from config would make the outputs distinguishable.

---

### 🟡 Suggestions — Data & Export

**No CSV / Excel export of rankings**
Rankings and group distributions can currently only be consumed as PDFs. Adding a simple CSV export for group evaluations and Ortsverband evaluations would allow results to be published on a website or merged into a spreadsheet for post-event reporting.

**No summary/statistics panel before PDF generation**
Before generating certificates, users would benefit from a summary confirming: total participants, number of groups, number of Ortsverbände, number of stations completed vs total. This acts as a final sanity check before the irreversible print action.

**[backend/database/evaluations.go] Tie-breaking rule not defined**
When two groups have equal total scores, the order between them is determined by the SQL `ORDER BY total_score DESC` alone. Ties could resolve to arbitrary row order between runs. A documented and consistent tie-breaking rule (e.g. secondary sort by group ID, or by fewest stations needed) should be applied.

---

### 🟢 Positive

- Clean separation of concerns: database, services, PDF I/O, HTTP handlers, and frontend modules are all well-isolated.
- The distribution algorithm is thoughtful — diversity scoring by Ortsverband, gender, and age is a sophisticated approach for a competition tool.
- Pre-group preservation in distribution is a strong feature for real-world use.
- Score entry dirty-tracking prevents accidental data loss when switching groups.
- The startup DB dialog (keep/reset with auto-backup) is a practical solution to the re-run scenario.
- Test coverage for the distribution algorithm and evaluation queries is solid.
- Input validation in `input.go` is comprehensive (header check, age range, alphanumeric PreGroup, score bounds).

---

## Test Assessment

Coverage: **Adequate for backend logic**, needs improvement for integration paths.

Missing tests:
- Full workflow test: load XLSX → distribute → assign scores → evaluate (verifies the entire pipeline end-to-end)
- Corrupt/malformed XLSX files (empty sheets, missing required columns, non-numeric scores in score column)
- Config validation edge cases (`min_punkte > max_punkte`, zero group size)
- Backup/restore cycle (backup → modify data → restore → verify data matches original)
- Distribution when PreGroup size exceeds `max_groesse`

---

## Verdict

🟡 **Functional for basic use — needs hardening before competition-day deployment**

The highest-priority items are:
1. The destructive-action confirmation on re-distribution (prevents silent data loss on competition day)
2. Config semantic validation (prevents a misconfigured `min_punkte`/`max_punkte` breaking score entry)
3. Score-entry UX improvements (bounds display, keyboard navigation, saved-state indicator) — the most heavily used part of the app
