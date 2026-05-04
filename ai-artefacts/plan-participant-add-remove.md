# Plan: Add / Remove Participants Without Redistribution

**Date:** 2026-05-04  
**Status:** Draft — approved through Q&A before implementation

---

## 1. Overview

Two new operations are added to the admin UI:

| Operation | Trigger | Redistribution? | Modes |
|-----------|---------|-----------------|-------|
| **Remove** | Participant cannot attend | ❌ None | All modes |
| **Add** | Late/replacement participant | ❌ None | All modes |

Both operations are surgical: they touch only the affected rows in the database
and leave the group structure, CarGroups, vehicles and caretakers untouched.

---

## 2. Constraints & Rules

### 2.1 Remove Participant

- The participant is **deleted entirely** from the event:
  - `teilnehmende` row deleted
  - `gruppe` row deleted (removes them from their group)
- **Blocked as soon as ANY score exists** anywhere in `group_station_scores`,
  regardless of which group it belongs to (same global rule as Add — see §2.2 and §7).
  The rationale: once the competition has started (first score entered for any group),
  the entire participant list is considered frozen.
- **Scores are per group, not per participant** (see §7 for detail). Removing a
  participant does **not** touch `group_station_scores`. The group keeps its scores.
- The CarGroup pool assignment is **not changed** — the pool simply has one more
  free seat after removal.
- No redistribution of remaining participants is triggered.

### 2.2 Add Participant

- A new participant is created in `teilnehmende` with a **system-generated**
  `teilnehmer_id` (next available number), then assigned to a chosen group.
- Required fields: **Name**, **Ortsverband**, **Alter** (age), **Geschlecht** (gender).
  `PreGroup` is always left empty for added participants.
- Addition is **only allowed when `group_station_scores` is completely empty**
  (i.e., no score has been entered for any group yet).
- The app suggests the **group with the fewest current members** that still has a
  free slot; the user can override and choose any eligible group.

#### Free-slot rules by mode

| Mode | "Free slot" definition |
|------|------------------------|
| `FixGroupSize` + CarGroups = `ja` | Group member count < `fixedGroupSize` **AND** the CarGroup pool total seats > current total headcount (participants + caretakers in the pool) |
| `FixGroupSize` + CarGroups = `nein` | Group member count < `fixedGroupSize` |
| `Klassisch` / `Fahrzeuge` | Group member count < seats in the vehicle assigned to that group (`gruppe_fahrzeuge` → `fahrzeuge.sitzplaetze`). If no vehicle is assigned, fall back to `MaxGroesse`. |

---

## 3. UI / UX Design

### 3.1 New Admin Tab: "Teilnehmer verwalten"

A new tab is added to the admin section (next to "Gruppen", "Stationen", etc.).
It contains two sections:

#### Section A — Remove Participant

1. A **search field** (live filter by name or Ortsverband).
2. A **table** listing all participants with columns:
   - Name | Ortsverband | Alter | Geschlecht | Gruppe
3. Each row has a **"Entfernen"** button (red, with trash icon).
   - If any score exists anywhere in the system, the button is **disabled**
     and shows a tooltip: "Es wurden bereits Wertungen eingetragen."
4. Clicking an enabled "Entfernen" shows a **confirmation modal**:
   > "Teilnehmer/in **[Name]** aus Gruppe **[Gruppenname]** wirklich entfernen?
   > Diese Aktion kann nicht rückgängig gemacht werden."
   > [Abbrechen] [Entfernen]
5. On confirmation: the backend removes the participant, regenerates affected
   PDFs, and a **result modal** appears automatically (see §3.2).

#### Section B — Add Participant

1. A **form** with fields: Name\*, Ortsverband, Alter\*, Geschlecht\* (\* = required).
2. A **group selector** (dropdown):
   - Only shows groups with at least one free slot (as defined in §2.2).
   - Pre-selects the group with the fewest current members.
   - Each option shows: `Gruppe X — [N/max] Plätze`
   - Groups with no free slots are **not shown at all** in the dropdown.
3. A **"Hinzufügen"** button.
4. On submit:
   - Validate that all required fields are filled.
   - Re-validate the free-slot rule server-side before inserting.
   - Insert into `teilnehmende`, then into `gruppe`, regenerate affected PDFs.
   - A **result modal** appears automatically (see §3.2).
5. The participant list in Section A and the group dropdown both refresh
   automatically after the modal is dismissed.

---

### 3.2 Result Modal (No Button Required)

After every successful add or remove the backend returns the operation summary
and a list of regenerated PDFs. The frontend shows this as a modal automatically
— the user does **not** need to press a "regenerate" button.

**Modal content:**

> ✅ **[Name]** wurde aus Gruppe **[Gruppenname]** entfernt.
> *(or: „… wurde Gruppe [Gruppenname] hinzugefügt.")*
>
> Folgende Dokumente wurden automatisch neu erstellt:
> • Gruppenaufteilung (PDF)
> • Teilnehmendenkarten (PDF)
> • Stationslaufzettel (PDF)
> • OV-Zuweisungen (PDF)
> • Übersicht (PDF)
> • Fahrgemeinschaften (PDF) *(only if CarGroups = ja)*
>
> Die Dateien liegen im konfigurierten PDF-Ordner.
>
> [Schließen]

If PDF regeneration fails for one or more documents, the modal still appears
but lists the failures in an orange warning block — the add/remove itself is
not rolled back.

---

## 4. Backend Changes Required

### 4.1 New Database Functions (`backend/database/`)

| Function | Description |
|----------|-------------|
| `DeleteTeilnehmer(db, teilnehmerID int) error` | DELETE FROM gruppe + teilnehmende in a transaction |
| `InsertTeilnehmer(db, t models.Teilnehmende) (int64, error)` | INSERT into teilnehmende, returns new id |
| `AssignTeilnehmerToGroup(db, teilnehmerID int, groupID int) error` | INSERT into gruppe |
| `GetGroupMemberCounts(db) (map[int]int, error)` | Returns map[group_id]count for all groups |
| `GetGroupSeatCapacity(db, mode string, fixedGroupSize int) (map[int]int, error)` | Returns map[group_id]maxSeats based on mode (vehicle seats or fixedGroupSize) |
| `NextTeilnehmerID(db) (int, error)` | SELECT MAX(teilnehmer_id)+1 from teilnehmende |
| `AnyScoreExists(db) (bool, error)` | Checks if `group_station_scores` has any row at all (global lock) |

### 4.2 New Handlers (`backend/handlers/admin.go` or new file `admin_participants.go`)

| Handler / Wails Export | Input | Output |
|------------------------|-------|--------|
| `RemoveTeilnehmer(id int)` | DB teilnehmer id | `{status, message, name, groupName, pdfResults []PDFResult}` |
| `GetParticipantsWithGroups()` | — | `[]ParticipantRow{ID, Name, Ortsverband, Alter, Geschlecht, GroupID, GroupName}` + top-level `scoresLocked bool` flag |
| `GetEligibleGroups()` | — | `[]EligibleGroup{GroupID, GroupName, CurrentCount, MaxSlots}` |
| `AddTeilnehmer(name, ortsverband string, alter int, geschlecht, groupID)` | — | `{status, message, newID, groupName, pdfResults []PDFResult}` |

`PDFResult` is a small struct: `{Name string, Status string, Error string}` —
one entry per PDF attempted, with `Status = "ok"` or `"error"`.

Each mutating handler (Add / Remove) runs the following sequence **before returning**:
1. DB transaction (delete or insert) — if this fails, return early with error.
2. Call each applicable `Generate*PDF` function in order (§8).
3. Collect successes and failures into `pdfResults`.
4. Return combined result — the frontend builds the result modal from this.

### 4.3 New / Updated Models

A lightweight response struct for the UI:

```go
type ParticipantRow struct {
    ID          int
    Name        string
    Ortsverband string
    Alter       int
    Geschlecht  string
    GroupID     int
    GroupName   string
}

type EligibleGroup struct {
    GroupID      int
    GroupName    string
    CurrentCount int
    MaxSlots     int   // -1 means unlimited (edge case: no vehicle assigned in Klassisch mode)
}
```

---

## 5. Frontend Changes Required

### 5.1 New file: `frontend/admin/participants.js`

Exports:
- `loadParticipantManagement()` — renders both sections into the tab container
- `handleRemoveParticipant(id, name, groupName)` — shows confirm modal, calls backend
- `handleAddParticipant(formData)` — validates form, calls backend
- `refreshEligibleGroups()` — re-fetches and updates the group dropdown

### 5.2 `frontend/shared/dom.js`

Add DOM reference for the new tab button and its container div.

### 5.3 `frontend/app.js`

- Import `loadParticipantManagement` from `participants.js`
- Expose relevant handlers to `window` for HTML `onclick` attributes
- Add tab-switch logic for the new tab

### 5.4 `frontend/index.html`

- Add new tab button: `<button id="btnParticipants">Teilnehmer verwalten</button>`
- Add new tab container section

### 5.5 `frontend/wailsjs/go/main/App.js` (auto-generated)

The three new Wails-exported functions will be auto-generated here when the Go
build runs: `RemoveTeilnehmer`, `GetParticipantsWithGroups`, `GetEligibleGroups`,
`AddTeilnehmer`.

---

## 6. Transaction Safety

Both Remove and Add must run inside a **single DB transaction** to avoid partial
state if the app crashes mid-operation:

- **Remove**: `DELETE FROM gruppe` + `DELETE FROM teilnehmende` — atomic.
- **Add**: `INSERT INTO teilnehmende` + `INSERT INTO gruppe` — atomic.
  The free-slot re-validation must happen inside the same transaction (using
  `SELECT COUNT` with `FOR UPDATE` semantics, or a SQLite `IMMEDIATE` transaction
  to prevent TOCTOU between the frontend check and the actual insert).

---

## 7. Important Clarification: Scores Are Per Group, Not Per Participant

`group_station_scores` stores one row per `(group_id, station_id)`. There are **no
participant-level score rows**. Consequences:

- Removing a participant **never deletes any score data** — the group's scores remain intact.
- The **score-gate rule applies globally**: the moment any single score row exists
  in `group_station_scores`, both Add and Remove are locked for the entire event.
- `AnyScoreExists(db) (bool, error)` is the single guard for both operations.
- Once locked, all "Entfernen" buttons are disabled and the Add form is hidden
  or replaced with an informational message:
  > "Hinzufügen und Entfernen ist nicht mehr möglich, da bereits Wertungen
  > eingetragen wurden."

---

## 8. PDF Auto-Regeneration

The following PDFs are regenerated automatically after every successful Add or
Remove. They all reflect group composition or participant headcount, so they
become stale after any membership change.

| PDF | Generator Function | Trigger |
|-----|-------------------|---------| 
| Gruppenaufteilung (group lists) | `GeneratePDFReport` | Always |
| Teilnehmendenkarten (participant cards) | `GenerateTeilnehmendeCardsPDF` | Always |
| Stationslaufzettel (station run sheets) | `GenerateStationSheetsPDF` | Always |
| OV-Zuweisungen (ortsverband assignment list) | `GenerateOVAssignmentsPDF` | Always |
| Übersicht (overview) | `GenerateOverviewPDF` | Always |
| Fahrgemeinschaften (carpool seating) | `GenerateCarGroupsPDF` | Only if CarGroups = `ja` |

PDFs that are **not** regenerated (unaffected by membership changes):
- `GenerateGroupEvaluationPDF` — scores-based, no participant list
- `GenerateOrtsverbandEvaluationPDF` — scores-based, no participant list
- `GenerateParticipantCertificates` — post-competition only
- `GenerateOrtsverbandCertificates` — post-competition only

### Regeneration order

PDFs are generated sequentially in the order listed above. Failure of one does
not abort the others. All outcomes are collected and returned in `pdfResults`.

---

## 9. Out of Scope

The following are explicitly **not** part of this plan:

- Re-assigning a participant to a different group (move, not add/remove)
- Modifying an existing participant's attributes (e.g., changing name or age)
- Modifying vehicle, caretaker, or CarGroup pool assignments
- Any redistribution or rebalancing after a removal
- Bulk import or undo history

---

## 10. Open Questions

| # | Question | Impact |
|---|----------|--------|
| 1 | In `Fahrzeuge` mode a group may have no vehicle assigned (e.g., a caretaker-only group). Should addition still be allowed using `MaxGroesse` as the cap, or should we block it? | `GetEligibleGroups` logic |
| 2 | Should the new "Teilnehmer verwalten" tab only be visible after a distribution has been run (groups exist), or always? | Tab visibility guard |
| 3 | For the Ortsverband field on add: free text, or a dropdown of existing Ortsverbände from the DB? | UX of add form |
| 4 | Should a removed participant be logged anywhere (audit trail) for on-site record-keeping? | New `audit_log` table or skip entirely |
| 5 | Should the result modal show a "Backup erstellen" shortcut button, since a membership change is a good moment to back up? | Post-action modal UX |
| 6 | If PDF regeneration is slow (e.g. >2 seconds), should the UI show a spinner while the backend processes, or is a blocking call acceptable? | Handler response time vs. UX |
