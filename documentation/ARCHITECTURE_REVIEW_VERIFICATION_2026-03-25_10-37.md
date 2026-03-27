# Architecture Review — Validity Check

**Based on:** [ARCHITECTURE_REVIEW.md](ARCHITECTURE_REVIEW.md) (last updated March 20, 2026)  
**Verified on:** March 25, 2026 — 10:37  
**Codebase state:** Post-feature additions (Fahrerlaubnis distribution, Eingabeübersicht)

---

## Summary

| # | Finding | Severity | Previous Status | Current Status |
|---|---------|----------|-----------------|----------------|
| 1 | Global database connection variable | Critical | ✅ Resolved | ✅ Still resolved |
| 2 | Redundant tables (gruppe + rel\_tn\_grp) | High | ✅ Resolved | ✅ Still resolved |
| 3 | Missing database indexes | High | ✅ Resolved | ✅ Still resolved |
| 4 | No connection pool management | High | ✅ Resolved | ✅ Still resolved |
| 5 | Empty `backend/handlers/` directory | Low | ⚠️ Open | ✅ **Resolved (Mar 25)** |
| 6 | Hardcoded configuration values | Medium | ✅ Resolved | ✅ Still resolved |
| 7 | No error recovery / retry mechanisms | Medium | ⚠️ Open | ⚠️ **Still open** |
| 8 | Frontend DOM manipulation | Low | ⚠️ Open | ⚠️ **Escalated — see below** |
| 9 | Certificate layout hardcoded | Low | ⚠️ Open | ⚠️ **Still open** |
| 10 | Test coverage reporting (`[no statements]`) | Medium | ⚠️ Open | ⚠️ **Still open** |
| 11 | Hardcoded certificate year | Low | ✅ Resolved | ✅ Still resolved |
| 12 | Data loss window in `LoadFile` | High | ✅ Resolved | ✅ Still resolved |
| 13 | No backend score validation | Medium | ✅ Resolved | ✅ Still resolved |
| 14 | Duplicate button-state management | Low | ✅ Resolved | ✅ Still resolved |

All resolved findings confirmed still resolved. Four originally-open findings remain open. One (Finding 8) has increased in relevance due to frontend growth.

---

## Open Findings — Detailed Status

### 5. Empty `backend/handlers/` Directory — ✅ Resolved (March 25, 2026)

**Option B implemented.** `backend/handlers/` now contains 6 domain files (`admin.go`, `backup.go`, `certificates.go`, `files.go`, `queries.go`, `reports.go`). A new project-root file `app_handlers.go` provides the Wails-required `App` method wrappers. `main.go` is now bootstrap-only.

---

---

### 7. No Error Recovery Mechanisms — ⚠️ Still Open

**Verified:** No retry or backoff logic was added to `backend/io/input.go` or `backend/io/output.go`. Grep for `retry`, `Retry`, `backoff`, `Backoff` returns no results in those files.

File operations still fail immediately on first error:
- `ReadXLSXFile` — no retry on transient file lock
- PDF `OutputFileAndClose` — no retry or temp-file fallback

**Recommendation unchanged:** Add `RetryWithBackoff` utility for critical I/O operations (PDF save, Excel open). Medium priority; 1–2 day effort.

---

### 8. Frontend DOM Manipulation — ⚠️ Escalated (Low → Medium)

**Verified:** The frontend has grown substantially since the March 10 review.

| File | Lines |
|------|-------|
| `frontend/app.js` | 110 |
| `frontend/admin/config-editor.js` | 97 |
| `frontend/admin/file-handler.js` | 238 |
| `frontend/evaluations/evaluations.js` | 212 |
| `frontend/groups/groups.js` | 147 |
| `frontend/reports/pdf-handlers.js` | 82 |
| `frontend/shared/dom.js` | 42 |
| `frontend/shared/utils.js` | 22 |
| `frontend/stations/scores.js` | 184 |
| `frontend/stations/stations.js` | 505 |
| **Total (active)** | **1,639** |

There are **88 DOM manipulation calls** (`innerHTML`, `createElement`, `appendChild`, `insertAdjacentHTML`) across the active frontend files.  

The newly added Eingabeübersicht feature in `stations.js` builds the matrix using HTML string concatenation and then assigns it via `innerHTML`. Dynamic station names and titles are correctly escaped via `escapeHtml()`, so XSS risk is mitigated. However, the pattern is spreading.

**Additional note — dead code:**  
`frontend/app.old.js` (898 lines) is still present in the project. It appears to be the pre-modularization version of `app.js` and serves no runtime purpose.

**Recommendation (updated):** 
- **Immediate:** Delete `frontend/app.old.js` (dead code, 898 lines, potential source of confusion).
- **Short-term:** Continue using `escapeHtml()` consistently for all dynamic content inserted via `innerHTML` — this discipline is already in place in new code.
- **Medium-term:** As the frontend continues to grow, consider replacing HTML string concatenation with a lightweight template utility or a minimal reactive library (see original recommendation for options).

---

### 9. Certificate Layout Hardcoded — ⚠️ Still Open

**Verified:** Both certificate generators still use hardcoded layout coordinates:

| File | Hardcoded position calls (`SetXY`, `CellFormat`, `SetFont`) |
|------|-----|
| `backend/io/pdf_cert_teilnehmende.go` | 25 |
| `backend/io/pdf_cert_ortsverbaende.go` | 24 |

No JSON/external layout configuration was introduced.

**Recommendation unchanged:** External layout configuration remains optional. Acceptable as-is for the current use case; implement only if layout customization is explicitly requested.

---

### 10. Test Coverage Reporting — ⚠️ Still Open

**Verified:**

```
$ go test -cover ./test/...
ok  THW-JugendOlympiade/test   coverage: [no statements]
```

Per-package coverage shows the root cause:

```
$ go test -cover ./backend/...
THW-JugendOlympiade/backend/config     coverage: 0.0% of statements
THW-JugendOlympiade/backend/database   coverage: 0.0% of statements
THW-JugendOlympiade/backend/io         coverage: 0.0% of statements
THW-JugendOlympiade/backend/services   coverage: 0.0% of statements
```

The tests live in package `test` and import the backend packages, but because there are no `.go` source files in the `test/` package itself, the toolchain reports zero coverage.

**Positive development:** The test suite has grown significantly — from 18 tests at the initial review to **122 test runs** (including table-driven sub-tests) as of today. All pass.

**Recommendation unchanged:** Use `-coverpkg` to report coverage against the packages under test:

```bash
go test -coverpkg=THW-JugendOlympiade/backend/... ./test/...
```

Or generate a full HTML report:

```bash
go test -coverpkg=THW-JugendOlympiade/backend/... -coverprofile=coverage.out ./test/...
go tool cover -html=coverage.out -o coverage.html
```

Estimated effort: 1–2 hours.

---

## New Observations (Since March 20, 2026)

These were not in the original review but are noteworthy as of this verification.

### N1. Fahrerlaubnis Distribution — 4-Phase Algorithm Added ✅

A new `distributeBetreuende` function implements a 4-phase algorithm:
1. Licensed drivers (`Fahrerlaubnis=ja`) spread one-per-group (round-robin)
2. Unlicensed Betreuende placed with their OV's licensed driver
3. Rebalancing unlicensed Betreuende evenly across groups
4. Safety net — any group still without a Betreuende receives one

The `Fahrerlaubnis` field is typed as `bool` in `models.Betreuende` (converted from the Excel "ja"/"nein" string on import). Architecture is clean.

### N2. Eingabeübersicht Matrix View ✅

A station × group completeness matrix was added to `frontend/stations/stations.js`. Dynamic content uses `escapeHtml()` consistently, which is correct security practice. The feature is self-contained within the stations module.

### N3. Test Suite Growth ✅

Test count grew from 18 (March 10) to 122 (March 25). New tests cover the Fahrerlaubnis distribution logic, group creation edge cases, score assignment, and evaluations. This is a strong positive signal.

---

## Priority Actions

| Priority | Action | Effort | Finding |
|----------|--------|--------|--------|
| Low | Fix coverage reporting with `-coverpkg` flag | 1–2 h | #10 |
| Medium | Add retry logic to PDF and Excel I/O operations | 1–2 d | #7 |
| Low | Externalize certificate layout (only if customization needed) | 2–3 d | #9 |

---

*Verified by GitHub Copilot — March 25, 2026*
