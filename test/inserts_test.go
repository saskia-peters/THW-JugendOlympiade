package test

import (
	"testing"

	"THW-JugendOlympiade/backend/database"
	"THW-JugendOlympiade/backend/models"
)

// ---- InsertStations ----

func TestInsertStations_EmptyRows(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertStations(db, [][]string{}); err != nil {
		t.Errorf("expected no error for empty rows, got: %v", err)
	}
}

func TestInsertStations_ValidData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	rows := [][]string{
		{"Station Name"},
		{"Weitsprung"},
		{"Sprint"},
		{"Ballwurf"},
	}
	if err := database.InsertStations(db, rows); err != nil {
		t.Fatalf("InsertStations failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM stations").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 stations, got %d", count)
	}

	var name string
	if err := db.QueryRow("SELECT station_name FROM stations WHERE station_id = 1").Scan(&name); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if name != "Weitsprung" {
		t.Errorf("expected first station 'Weitsprung', got '%s'", name)
	}
}

func TestInsertStations_SkipsEmptyNames(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	rows := [][]string{
		{"Station Name"},
		{"Weitsprung"},
		{""},   // empty name — should be skipped
		{"  "}, // whitespace only — should be skipped
		{"Ballwurf"},
	}
	if err := database.InsertStations(db, rows); err != nil {
		t.Fatalf("InsertStations failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM stations").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 stations (empty names skipped), got %d", count)
	}
}

func TestInsertStations_OnlyHeader(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertStations(db, [][]string{{"Station Name"}}); err != nil {
		t.Errorf("expected no error for header-only, got: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM stations").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 stations, got %d", count)
	}
}

// ---- InsertBetreuende ----

func TestInsertBetreuende_EmptyRows(t *testing.T) {
	db := setupFullTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertBetreuende(db, [][]string{}); err != nil {
		t.Errorf("expected no error for empty rows, got: %v", err)
	}
}

func TestInsertBetreuende_ValidData(t *testing.T) {
	db := setupFullTestDB(t)
	defer teardownTestDB(t, db)

	rows := [][]string{
		{"Name", "Ortsverband"},
		{"Hans Müller", "Berlin"},
		{"Maria Schmidt", "Hamburg"},
		{"Klaus Weber", "München"},
	}
	if err := database.InsertBetreuende(db, rows); err != nil {
		t.Fatalf("InsertBetreuende failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM betreuende").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 betreuende, got %d", count)
	}

	var name, ov string
	if err := db.QueryRow("SELECT name, ortsverband FROM betreuende WHERE id = 1").Scan(&name, &ov); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if name != "Hans Müller" {
		t.Errorf("expected 'Hans Müller', got '%s'", name)
	}
	if ov != "Berlin" {
		t.Errorf("expected 'Berlin', got '%s'", ov)
	}
}

func TestInsertBetreuende_SkipsEmptyNames(t *testing.T) {
	db := setupFullTestDB(t)
	defer teardownTestDB(t, db)

	rows := [][]string{
		{"Name", "Ortsverband"},
		{"Valid Trainer", "Berlin"},
		{"", "Hamburg"},   // empty name → skipped
		{"  ", "München"}, // whitespace-only name → skipped
	}
	if err := database.InsertBetreuende(db, rows); err != nil {
		t.Fatalf("InsertBetreuende failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM betreuende").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 betreuende (empty names skipped), got %d", count)
	}
}

func TestInsertBetreuende_OnlyHeader(t *testing.T) {
	db := setupFullTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertBetreuende(db, [][]string{{"Name", "Ortsverband"}}); err != nil {
		t.Errorf("expected no error for header-only, got: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM betreuende").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 betreuende, got %d", count)
	}
}

// ---- SaveGroups ----

func TestSaveGroups_Basic(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertData(db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
		{"Bob", "Hamburg", "22", "M", ""},
	}); err != nil {
		t.Fatalf("InsertData failed: %v", err)
	}

	groups := []models.Group{
		{GroupID: 1, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 1}}},
		{GroupID: 2, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 2}}},
	}
	if err := database.SaveGroups(db, groups); err != nil {
		t.Fatalf("SaveGroups failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM gruppe").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 gruppe rows, got %d", count)
	}

	var groupID int
	if err := db.QueryRow("SELECT group_id FROM gruppe WHERE teilnehmer_id = 1").Scan(&groupID); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if groupID != 1 {
		t.Errorf("expected teilnehmer_id 1 in group 1, got group %d", groupID)
	}
}

func TestSaveGroups_ReplacesExistingAssignments(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertData(db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
		{"Bob", "Hamburg", "22", "M", ""},
	}); err != nil {
		t.Fatalf("InsertData failed: %v", err)
	}

	// First assignment: Alice → group 1
	if err := database.SaveGroups(db, []models.Group{
		{GroupID: 1, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 1}}},
	}); err != nil {
		t.Fatalf("first SaveGroups failed: %v", err)
	}

	// Regroup: Alice → group 2, Bob → group 1
	if err := database.SaveGroups(db, []models.Group{
		{GroupID: 1, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 2}}},
		{GroupID: 2, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 1}}},
	}); err != nil {
		t.Fatalf("second SaveGroups failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM gruppe").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 rows after replace, got %d", count)
	}

	// Alice (teilnehmer_id=1) should now be in group 2
	var groupID int
	if err := db.QueryRow("SELECT group_id FROM gruppe WHERE teilnehmer_id = 1").Scan(&groupID); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if groupID != 2 {
		t.Errorf("expected Alice in group 2 after reshuffle, got group %d", groupID)
	}
}

func TestSaveGroups_EmptyGroupList(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.SaveGroups(db, []models.Group{}); err != nil {
		t.Errorf("expected no error for empty group list, got: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM gruppe").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 gruppe rows, got %d", count)
	}
}

// ---- SaveGroupBetreuende ----

func TestSaveGroupBetreuende_Basic(t *testing.T) {
	db := setupFullTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertBetreuende(db, [][]string{
		{"Name", "Ortsverband"},
		{"Trainer A", "Berlin"},
		{"Trainer B", "Hamburg"},
	}); err != nil {
		t.Fatalf("InsertBetreuende failed: %v", err)
	}

	var idA, idB int
	if err := db.QueryRow("SELECT id FROM betreuende WHERE name = 'Trainer A'").Scan(&idA); err != nil {
		t.Fatalf("ID lookup failed: %v", err)
	}
	if err := db.QueryRow("SELECT id FROM betreuende WHERE name = 'Trainer B'").Scan(&idB); err != nil {
		t.Fatalf("ID lookup failed: %v", err)
	}

	groups := []models.Group{
		{GroupID: 1, Betreuende: []models.Betreuende{{ID: idA, Name: "Trainer A"}}},
		{GroupID: 2, Betreuende: []models.Betreuende{{ID: idB, Name: "Trainer B"}}},
	}
	if err := database.SaveGroupBetreuende(db, groups); err != nil {
		t.Fatalf("SaveGroupBetreuende failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM gruppe_betreuende").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 gruppe_betreuende rows, got %d", count)
	}

	var groupID int
	if err := db.QueryRow("SELECT group_id FROM gruppe_betreuende WHERE betreuende_id = ?", idA).Scan(&groupID); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if groupID != 1 {
		t.Errorf("expected Trainer A in group 1, got group %d", groupID)
	}
}

func TestSaveGroupBetreuende_ReplacesExistingAssignments(t *testing.T) {
	db := setupFullTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertBetreuende(db, [][]string{
		{"Name", "Ortsverband"},
		{"Trainer A", "Berlin"},
	}); err != nil {
		t.Fatalf("InsertBetreuende failed: %v", err)
	}

	var idA int
	if err := db.QueryRow("SELECT id FROM betreuende WHERE name = 'Trainer A'").Scan(&idA); err != nil {
		t.Fatalf("ID lookup failed: %v", err)
	}

	// First: Trainer A → group 1
	if err := database.SaveGroupBetreuende(db, []models.Group{
		{GroupID: 1, Betreuende: []models.Betreuende{{ID: idA}}},
	}); err != nil {
		t.Fatalf("first SaveGroupBetreuende failed: %v", err)
	}

	// Replace: Trainer A → group 2
	if err := database.SaveGroupBetreuende(db, []models.Group{
		{GroupID: 2, Betreuende: []models.Betreuende{{ID: idA}}},
	}); err != nil {
		t.Fatalf("second SaveGroupBetreuende failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM gruppe_betreuende").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row after replace, got %d", count)
	}

	var groupID int
	if err := db.QueryRow("SELECT group_id FROM gruppe_betreuende WHERE betreuende_id = ?", idA).Scan(&groupID); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if groupID != 2 {
		t.Errorf("expected Trainer A in group 2 after replace, got group %d", groupID)
	}
}

func TestSaveGroupBetreuende_EmptyGroupList(t *testing.T) {
	db := setupFullTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.SaveGroupBetreuende(db, []models.Group{}); err != nil {
		t.Errorf("expected no error for empty group list, got: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM gruppe_betreuende").Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows, got %d", count)
	}
}
