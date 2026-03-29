package database_test

import (
	"testing"

	"THW-JugendOlympiade/backend/database"
	"THW-JugendOlympiade/backend/models"
)

// ---------------------------------------------------------------------------
// InsertData
// ---------------------------------------------------------------------------

func TestInsertData_Basic(t *testing.T) {
	db := newTestDB(t)

	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
		{"Bob", "Hamburg", "30", "M", "Alpha"},
		{"Carol", "München", "22", "W", "Alpha"},
	}
	if err := database.InsertData(db, rows); err != nil {
		t.Fatalf("InsertData: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM teilnehmende").Scan(&count)
	if count != 3 {
		t.Errorf("expected 3 participants, got %d", count)
	}
}

func TestInsertData_HeaderOnly_InsertsNothing(t *testing.T) {
	db := newTestDB(t)

	if err := database.InsertData(db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
	}); err != nil {
		t.Fatalf("InsertData: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM teilnehmende").Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 participants, got %d", count)
	}
}

func TestInsertData_EmptyInput_IsNoOp(t *testing.T) {
	db := newTestDB(t)
	if err := database.InsertData(db, [][]string{}); err != nil {
		t.Fatalf("InsertData with empty input: %v", err)
	}
	var count int
	db.QueryRow("SELECT COUNT(*) FROM teilnehmende").Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 participants, got %d", count)
	}
}

func TestInsertData_SkipsRowsWithEmptyName(t *testing.T) {
	db := newTestDB(t)

	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
		{"", "Hamburg", "30", "M", ""},   // empty name — should be skipped
		{"  ", "München", "22", "W", ""}, // whitespace-only — should be skipped
	}
	if err := database.InsertData(db, rows); err != nil {
		t.Fatalf("InsertData: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM teilnehmende").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 participant, got %d", count)
	}
}

func TestInsertData_RowIndexBecomesToeilnehmerID(t *testing.T) {
	db := newTestDB(t)

	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
		{"Bob", "Hamburg", "30", "M", ""},
	}
	if err := database.InsertData(db, rows); err != nil {
		t.Fatalf("InsertData: %v", err)
	}

	// Alice is at row index 1 → teilnehmer_id = 1
	var name string
	db.QueryRow("SELECT name FROM teilnehmende WHERE teilnehmer_id = 1").Scan(&name)
	if name != "Alice" {
		t.Errorf("expected teilnehmer_id=1 → Alice, got %q", name)
	}
	// Bob is at row index 2 → teilnehmer_id = 2
	db.QueryRow("SELECT name FROM teilnehmende WHERE teilnehmer_id = 2").Scan(&name)
	if name != "Bob" {
		t.Errorf("expected teilnehmer_id=2 → Bob, got %q", name)
	}
}

func TestInsertData_TrimsWhitespace(t *testing.T) {
	db := newTestDB(t)

	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"  Alice  ", "  Berlin  ", "25", "W", ""},
	}
	if err := database.InsertData(db, rows); err != nil {
		t.Fatalf("InsertData: %v", err)
	}

	var name, ov string
	db.QueryRow("SELECT name, ortsverband FROM teilnehmende").Scan(&name, &ov)
	if name != "Alice" {
		t.Errorf("name: want %q, got %q", "Alice", name)
	}
	if ov != "Berlin" {
		t.Errorf("ortsverband: want %q, got %q", "Berlin", ov)
	}
}

func TestInsertData_ShortRowsPaddedWithEmpty(t *testing.T) {
	db := newTestDB(t)

	// Row has only name — other fields default to empty
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice"},
	}
	if err := database.InsertData(db, rows); err != nil {
		t.Fatalf("InsertData: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM teilnehmende WHERE name='Alice'").Scan(&count)
	if count != 1 {
		t.Errorf("expected Alice to be inserted even with short row")
	}
}

// ---------------------------------------------------------------------------
// InsertStations
// ---------------------------------------------------------------------------

func TestInsertStations_Basic(t *testing.T) {
	db := newTestDB(t)

	rows := [][]string{
		{"Stationsname"},
		{"Bogenschießen"},
		{"Sanitätsdienst"},
		{"Knotenkunde"},
	}
	if err := database.InsertStations(db, rows); err != nil {
		t.Fatalf("InsertStations: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM stations").Scan(&count)
	if count != 3 {
		t.Errorf("expected 3 stations, got %d", count)
	}
}

func TestInsertStations_EmptyInput_IsNoOp(t *testing.T) {
	db := newTestDB(t)
	if err := database.InsertStations(db, [][]string{}); err != nil {
		t.Fatalf("InsertStations with empty input: %v", err)
	}
	var count int
	db.QueryRow("SELECT COUNT(*) FROM stations").Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 stations, got %d", count)
	}
}

func TestInsertStations_SkipsEmptyNames(t *testing.T) {
	db := newTestDB(t)

	rows := [][]string{
		{"Stationsname"},
		{"Bogenschießen"},
		{""},   // empty — should be skipped
		{"  "}, // whitespace — should be skipped
		{"Sanitätsdienst"},
	}
	if err := database.InsertStations(db, rows); err != nil {
		t.Fatalf("InsertStations: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM stations").Scan(&count)
	if count != 2 {
		t.Errorf("expected 2 stations (empty names skipped), got %d", count)
	}
}

// ---------------------------------------------------------------------------
// InsertBetreuende
// ---------------------------------------------------------------------------

func TestInsertBetreuende_Basic(t *testing.T) {
	db := newTestDB(t)

	rows := [][]string{
		{"Name", "Ortsverband", "Fahrerlaubnis"},
		{"Max Muster", "Berlin", "ja"},
		{"Anna Klein", "Hamburg", "nein"},
		{"Tom Berg", "München", ""},
	}
	if err := database.InsertBetreuende(db, rows); err != nil {
		t.Fatalf("InsertBetreuende: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM betreuende").Scan(&count)
	if count != 3 {
		t.Errorf("expected 3 betreuende, got %d", count)
	}
}

func TestInsertBetreuende_FahrerlaubnisMapping(t *testing.T) {
	db := newTestDB(t)

	rows := [][]string{
		{"Name", "Ortsverband", "Fahrerlaubnis"},
		{"Licensed", "Berlin", "ja"},
		{"LicensedUpper", "Berlin", "JA"},
		{"LicensedMixed", "Berlin", "Ja"},
		{"Unlicensed", "Hamburg", "nein"},
		{"NoEntry", "München", ""},
	}
	if err := database.InsertBetreuende(db, rows); err != nil {
		t.Fatalf("InsertBetreuende: %v", err)
	}

	var licensed int
	db.QueryRow("SELECT COUNT(*) FROM betreuende WHERE fahrerlaubnis = 1").Scan(&licensed)
	if licensed != 3 {
		t.Errorf("expected 3 licensed drivers (all 'ja' variants), got %d", licensed)
	}

	var unlicensed int
	db.QueryRow("SELECT COUNT(*) FROM betreuende WHERE fahrerlaubnis = 0").Scan(&unlicensed)
	if unlicensed != 2 {
		t.Errorf("expected 2 non-licensed betreuende, got %d", unlicensed)
	}
}

func TestInsertBetreuende_EmptyInput_IsNoOp(t *testing.T) {
	db := newTestDB(t)
	if err := database.InsertBetreuende(db, [][]string{}); err != nil {
		t.Fatalf("InsertBetreuende with empty: %v", err)
	}
	var count int
	db.QueryRow("SELECT COUNT(*) FROM betreuende").Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 betreuende, got %d", count)
	}
}

func TestInsertBetreuende_SkipsEmptyNames(t *testing.T) {
	db := newTestDB(t)

	rows := [][]string{
		{"Name", "Ortsverband", "Fahrerlaubnis"},
		{"Max Muster", "Berlin", "ja"},
		{"", "Hamburg", "ja"},   // empty name — skipped
		{"  ", "München", "ja"}, // whitespace — skipped (if trimmed)
	}
	if err := database.InsertBetreuende(db, rows); err != nil {
		t.Fatalf("InsertBetreuende: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM betreuende").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 betreuende (empty names skipped), got %d", count)
	}
}

// ---------------------------------------------------------------------------
// SaveGroups
// ---------------------------------------------------------------------------

func TestSaveGroups_InsertsGroupMemberships(t *testing.T) {
	db := newTestDB(t)
	mustInsertParticipants(t, db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
		{"Bob", "Hamburg", "22", "M", ""},
	})

	groups := []models.Group{
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 1}}},
		{GroupID: 2, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 2}}},
	}

	if err := database.SaveGroups(db, groups); err != nil {
		t.Fatalf("SaveGroups: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM gruppe").Scan(&count)
	if count != 2 {
		t.Errorf("expected 2 gruppe rows, got %d", count)
	}
}

func TestSaveGroups_ReplacesExistingAssignments(t *testing.T) {
	db := newTestDB(t)
	mustInsertParticipants(t, db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
		{"Bob", "Hamburg", "22", "M", ""},
	})

	first := []models.Group{
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 1}}},
	}
	second := []models.Group{
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 2}}},
	}

	if err := database.SaveGroups(db, first); err != nil {
		t.Fatalf("first SaveGroups: %v", err)
	}
	if err := database.SaveGroups(db, second); err != nil {
		t.Fatalf("second SaveGroups: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM gruppe").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 gruppe row after replace, got %d (old rows leaked)", count)
	}

	var memberID int
	db.QueryRow("SELECT teilnehmer_id FROM gruppe WHERE group_id = 1").Scan(&memberID)
	if memberID != 2 {
		t.Errorf("expected group 1 to contain teilnehmer_id=2, got %d", memberID)
	}
}

// ---------------------------------------------------------------------------
// SaveGroupBetreuende
// ---------------------------------------------------------------------------

func TestSaveGroupBetreuende_AssignsBetreuerToGroup(t *testing.T) {
	db := newTestDB(t)

	// Seed participants and groups
	mustInsertParticipants(t, db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
	})
	mustInsertBetreuende(t, db, [][]string{
		{"Name", "Ortsverband", "Fahrerlaubnis"},
		{"Trainer", "Berlin", "ja"},
	})

	// Fetch betreuer ID
	var betreuerID int
	db.QueryRow("SELECT id FROM betreuende WHERE name = 'Trainer'").Scan(&betreuerID)

	groups := []models.Group{
		{
			GroupID: 1,
			Betreuende: []models.Betreuende{
				{ID: betreuerID, Name: "Trainer"},
			},
		},
	}

	if err := database.SaveGroupBetreuende(db, groups); err != nil {
		t.Fatalf("SaveGroupBetreuende: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM gruppe_betreuende WHERE group_id = 1").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 gruppe_betreuende row, got %d", count)
	}
}

func TestSaveGroupBetreuende_ReplacesExistingAssignments(t *testing.T) {
	db := newTestDB(t)

	mustInsertBetreuende(t, db, [][]string{
		{"Name", "Ortsverband", "Fahrerlaubnis"},
		{"T1", "Berlin", "ja"},
		{"T2", "Hamburg", "ja"},
	})

	var id1, id2 int
	db.QueryRow("SELECT id FROM betreuende WHERE name = 'T1'").Scan(&id1)
	db.QueryRow("SELECT id FROM betreuende WHERE name = 'T2'").Scan(&id2)

	first := []models.Group{{GroupID: 1, Betreuende: []models.Betreuende{{ID: id1}}}}
	second := []models.Group{{GroupID: 1, Betreuende: []models.Betreuende{{ID: id2}}}}

	if err := database.SaveGroupBetreuende(db, first); err != nil {
		t.Fatalf("first SaveGroupBetreuende: %v", err)
	}
	if err := database.SaveGroupBetreuende(db, second); err != nil {
		t.Fatalf("second SaveGroupBetreuende: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM gruppe_betreuende").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 assignment after replace, got %d", count)
	}
}
