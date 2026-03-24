package test

import (
	"database/sql"
	"testing"

	"THW-JugendOlympiade/backend/database"
	"THW-JugendOlympiade/backend/models"
)

// setupFullTestDB extends setupTestDB with the betreuende and gruppe_betreuende tables.
// These tables exist in production but are not part of the minimal setupTestDB fixture.
func setupFullTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db := setupTestDB(t)

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS betreuende (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		ortsverband TEXT,
		fahrerlaubnis INTEGER NOT NULL DEFAULT 0
	)`); err != nil {
		t.Fatalf("Failed to create betreuende table: %v", err)
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS gruppe_betreuende (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		group_id INTEGER NOT NULL,
		betreuende_id INTEGER NOT NULL,
		UNIQUE(group_id, betreuende_id)
	)`); err != nil {
		t.Fatalf("Failed to create gruppe_betreuende table: %v", err)
	}

	return db
}

// ---- GetAllTeilnehmende ----

func TestGetAllTeilnehmende_EmptyDB(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	result, err := database.GetAllTeilnehmende(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d items", len(result))
	}
}

func TestGetAllTeilnehmende_WithParticipants(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "28", "W", ""},
		{"Bob", "Hamburg", "22", "M", "Alpha"},
	}
	if err := database.InsertData(db, rows); err != nil {
		t.Fatalf("InsertData failed: %v", err)
	}

	result, err := database.GetAllTeilnehmende(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(result))
	}
	if result[0].Name != "Alice" {
		t.Errorf("expected first participant 'Alice', got '%s'", result[0].Name)
	}
	if result[0].Alter != 28 {
		t.Errorf("expected Alter 28, got %d", result[0].Alter)
	}
	if result[0].Geschlecht != "W" {
		t.Errorf("expected Geschlecht 'W', got '%s'", result[0].Geschlecht)
	}
	if result[1].Ortsverband != "Hamburg" {
		t.Errorf("expected Ortsverband 'Hamburg', got '%s'", result[1].Ortsverband)
	}
	if result[1].PreGroup != "Alpha" {
		t.Errorf("expected PreGroup 'Alpha', got '%s'", result[1].PreGroup)
	}
}

func TestGetAllTeilnehmers_OrderedByAutoIncrementID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Insert with non-sequential teilnehmer_ids; order should be by autoincrement id (insertion order)
	for _, ins := range []struct {
		tid  int
		name string
	}{{10, "Charlie"}, {20, "Alice"}, {30, "Bob"}} {
		if _, err := db.Exec(
			"INSERT INTO teilnehmende (teilnehmer_id, name, ortsverband, age, geschlecht, pregroup) VALUES (?, ?, '', NULL, '', '')",
			ins.tid, ins.name,
		); err != nil {
			t.Fatalf("insert failed: %v", err)
		}
	}

	result, err := database.GetAllTeilnehmende(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 participants, got %d", len(result))
	}
	// ORDER BY id (autoincrement = insertion order): Charlie, Alice, Bob
	names := []string{"Charlie", "Alice", "Bob"}
	for i, want := range names {
		if result[i].Name != want {
			t.Errorf("position %d: expected '%s', got '%s'", i, want, result[i].Name)
		}
	}
}

func TestGetAllTeilnehmers_NullAge(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if _, err := db.Exec(
		"INSERT INTO teilnehmende (teilnehmer_id, name, ortsverband, age, geschlecht, pregroup) VALUES (1, 'NoAge', 'Berlin', NULL, 'M', '')",
	); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	result, err := database.GetAllTeilnehmende(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 participant, got %d", len(result))
	}
	if result[0].Alter != 0 {
		t.Errorf("expected Alter 0 for NULL age, got %d", result[0].Alter)
	}
}

// ---- GetGroupsForReport ----

func TestGetGroupsForReport_EmptyDB(t *testing.T) {
	db := setupFullTestDB(t)
	defer teardownTestDB(t, db)

	result, err := database.GetGroupsForReport(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d groups", len(result))
	}
}

func TestGetGroupsForReport_GroupsWithParticipants(t *testing.T) {
	db := setupFullTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertData(db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "28", "W", ""},
		{"Bob", "Berlin", "22", "M", ""},
		{"Carol", "Hamburg", "25", "W", ""},
	}); err != nil {
		t.Fatalf("InsertData failed: %v", err)
	}

	// teilnehmer_id: Alice=1, Bob=2, Carol=3
	if err := database.SaveGroups(db, []models.Group{
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 1}, {TeilnehmendeID: 2}}},
		{GroupID: 2, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 3}}},
	}); err != nil {
		t.Fatalf("SaveGroups failed: %v", err)
	}

	result, err := database.GetGroupsForReport(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(result))
	}

	g1 := result[0]
	if g1.GroupID != 1 {
		t.Errorf("expected GroupID 1, got %d", g1.GroupID)
	}
	if len(g1.Teilnehmende) != 2 {
		t.Errorf("expected 2 participants in group 1, got %d", len(g1.Teilnehmende))
	}
	if g1.Ortsverbands["Berlin"] != 2 {
		t.Errorf("expected 2 from Berlin in group 1, got %d", g1.Ortsverbands["Berlin"])
	}
	if g1.AlterSum != 50 { // 28 + 22
		t.Errorf("expected AlterSum 50 in group 1, got %d", g1.AlterSum)
	}

	g2 := result[1]
	if g2.GroupID != 2 {
		t.Errorf("expected GroupID 2, got %d", g2.GroupID)
	}
	if len(g2.Teilnehmende) != 1 {
		t.Errorf("expected 1 participant in group 2, got %d", len(g2.Teilnehmende))
	}
}

func TestGetGroupsForReport_GeschlechtStatistics(t *testing.T) {
	db := setupFullTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertData(db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Anna", "München", "20", "W", ""},
		{"Tom", "München", "30", "M", ""},
		{"Lena", "Köln", "25", "W", ""},
	}); err != nil {
		t.Fatalf("InsertData failed: %v", err)
	}

	if err := database.SaveGroups(db, []models.Group{
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 1}, {TeilnehmendeID: 2}, {TeilnehmendeID: 3}}},
	}); err != nil {
		t.Fatalf("SaveGroups failed: %v", err)
	}

	result, err := database.GetGroupsForReport(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 group, got %d", len(result))
	}

	g := result[0]
	if g.Geschlechts["W"] != 2 {
		t.Errorf("expected 2 female, got %d", g.Geschlechts["W"])
	}
	if g.Geschlechts["M"] != 1 {
		t.Errorf("expected 1 male, got %d", g.Geschlechts["M"])
	}
	if g.AlterSum != 75 { // 20 + 30 + 25
		t.Errorf("expected AlterSum 75, got %d", g.AlterSum)
	}
}

// ---- GetAllBetreuende ----

func TestGetAllBetreuende_EmptyDB(t *testing.T) {
	db := setupFullTestDB(t)
	defer teardownTestDB(t, db)

	result, err := database.GetAllBetreuende(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

func TestGetAllBetreuende_WithData(t *testing.T) {
	db := setupFullTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertBetreuende(db, [][]string{
		{"Name", "Ortsverband"},
		{"Trainer Müller", "Berlin"},
		{"Trainer Schmidt", "Hamburg"},
	}); err != nil {
		t.Fatalf("InsertBetreuende failed: %v", err)
	}

	result, err := database.GetAllBetreuende(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 betreuende, got %d", len(result))
	}
	if result[0].Name != "Trainer Müller" {
		t.Errorf("expected 'Trainer Müller', got '%s'", result[0].Name)
	}
	if result[0].Ortsverband != "Berlin" {
		t.Errorf("expected Ortsverband 'Berlin', got '%s'", result[0].Ortsverband)
	}
	if result[1].Name != "Trainer Schmidt" {
		t.Errorf("expected 'Trainer Schmidt', got '%s'", result[1].Name)
	}
	// Verify IDs are assigned
	if result[0].ID == 0 {
		t.Error("expected non-zero ID for betreuende")
	}
}

// ---- GetStationsForReport ----

func TestGetStationsForReport_EmptyDB(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	result, err := database.GetStationsForReport(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d stations", len(result))
	}
}

func TestGetStationsForReport_StationsWithoutScores(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertStations(db, [][]string{
		{"Station Name"},
		{"Weitsprung"},
		{"Sprint"},
	}); err != nil {
		t.Fatalf("InsertStations failed: %v", err)
	}

	result, err := database.GetStationsForReport(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 stations, got %d", len(result))
	}
	// LEFT JOIN — stations with no scores should have empty GroupScores
	for _, station := range result {
		if len(station.GroupScores) != 0 {
			t.Errorf("station %q: expected no group scores (LEFT JOIN), got %d",
				station.StationName, len(station.GroupScores))
		}
	}
}

func TestGetStationsForReport_WithScores(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertData(db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
		{"Bob", "Hamburg", "22", "M", ""},
	}); err != nil {
		t.Fatalf("InsertData failed: %v", err)
	}
	if err := database.InsertStations(db, [][]string{
		{"Station Name"},
		{"Ballwurf"},
	}); err != nil {
		t.Fatalf("InsertStations failed: %v", err)
	}
	if err := database.SaveGroups(db, []models.Group{
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 1}}},
		{GroupID: 2, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 2}}},
	}); err != nil {
		t.Fatalf("SaveGroups failed: %v", err)
	}
	if err := database.AssignGroupStationScore(db, 1, 1, 80); err != nil {
		t.Fatalf("AssignGroupStationScore failed: %v", err)
	}
	if err := database.AssignGroupStationScore(db, 2, 1, 60); err != nil {
		t.Fatalf("AssignGroupStationScore failed: %v", err)
	}

	result, err := database.GetStationsForReport(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 station, got %d", len(result))
	}
	if result[0].StationName != "Ballwurf" {
		t.Errorf("expected station name 'Ballwurf', got '%s'", result[0].StationName)
	}
	// Scores ordered by score DESC (80, 60)
	if len(result[0].GroupScores) != 2 {
		t.Fatalf("expected 2 group scores, got %d", len(result[0].GroupScores))
	}
	if result[0].GroupScores[0].Score != 80 {
		t.Errorf("expected highest score 80 first, got %d", result[0].GroupScores[0].Score)
	}
	if result[0].GroupScores[1].Score != 60 {
		t.Errorf("expected score 60 second, got %d", result[0].GroupScores[1].Score)
	}
}

func TestGetStationsForReport_MultipleStations(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertData(db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
	}); err != nil {
		t.Fatalf("InsertData failed: %v", err)
	}
	if err := database.InsertStations(db, [][]string{
		{"Station Name"},
		{"Weitsprung"},
		{"Sprint"},
		{"Ballwurf"},
	}); err != nil {
		t.Fatalf("InsertStations failed: %v", err)
	}
	if err := database.SaveGroups(db, []models.Group{
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 1}}},
	}); err != nil {
		t.Fatalf("SaveGroups failed: %v", err)
	}
	// Only score two of three stations
	if err := database.AssignGroupStationScore(db, 1, 1, 70); err != nil {
		t.Fatalf("score assignment failed: %v", err)
	}
	if err := database.AssignGroupStationScore(db, 1, 2, 85); err != nil {
		t.Fatalf("score assignment failed: %v", err)
	}

	result, err := database.GetStationsForReport(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 stations (including unscored), got %d", len(result))
	}

	scored := 0
	for _, s := range result {
		scored += len(s.GroupScores)
	}
	if scored != 2 {
		t.Errorf("expected 2 scored stations total, got %d", scored)
	}
}

// ---- GetAllGroupIDs ----

func TestGetAllGroupIDs_EmptyDB(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	result, err := database.GetAllGroupIDs(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestGetAllGroupIDs_DistinctAndSorted(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertData(db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
		{"Bob", "Hamburg", "22", "M", ""},
		{"Carol", "München", "30", "W", ""},
	}); err != nil {
		t.Fatalf("InsertData failed: %v", err)
	}

	// Save groups in non-sequential order to verify sorting
	if err := database.SaveGroups(db, []models.Group{
		{GroupID: 3, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 1}}},
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 2}}},
		{GroupID: 2, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 3}}},
	}); err != nil {
		t.Fatalf("SaveGroups failed: %v", err)
	}

	result, err := database.GetAllGroupIDs(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 group IDs, got %d", len(result))
	}
	// Must be sorted ascending: 1, 2, 3
	for i, expected := range []int{1, 2, 3} {
		if result[i] != expected {
			t.Errorf("position %d: expected group ID %d, got %d", i, expected, result[i])
		}
	}
}
