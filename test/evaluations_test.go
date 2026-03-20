package test

import (
	"testing"

	"THW-JugendOlympiade/backend/database"
	"THW-JugendOlympiade/backend/models"
)

// ---- GetGroupEvaluations edge cases ----
// Note: TestGetGroupEvaluations (multi-group ranking) is already covered in scores_test.go.
// These tests cover additional scenarios: empty DB, StationCount field, and tie-breaking.

func TestGetGroupEvaluations_EmptyDB(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	result, err := database.GetGroupEvaluations(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d evaluations", len(result))
	}
}

func TestGetGroupEvaluations_StationCount(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	if err := database.InsertData(db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
	}); err != nil {
		t.Fatalf("InsertData failed: %v", err)
	}
	if err := database.InsertStations(db, [][]string{
		{"Station"},
		{"Weitsprung"},
		{"Ballwurf"},
		{"Sprint"},
	}); err != nil {
		t.Fatalf("InsertStations failed: %v", err)
	}
	if err := database.SaveGroups(db, []models.Group{
		{GroupID: 1, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 1}}},
	}); err != nil {
		t.Fatalf("SaveGroups failed: %v", err)
	}
	// Score only 2 of 3 stations
	if err := database.AssignGroupStationScore(db, 1, 1, 70); err != nil {
		t.Fatalf("failed: %v", err)
	}
	if err := database.AssignGroupStationScore(db, 1, 2, 80); err != nil {
		t.Fatalf("failed: %v", err)
	}

	result, err := database.GetGroupEvaluations(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 evaluation, got %d", len(result))
	}
	if result[0].TotalScore != 150 {
		t.Errorf("expected TotalScore 150, got %d", result[0].TotalScore)
	}
	if result[0].StationCount != 2 {
		t.Errorf("expected StationCount 2 (only scored stations), got %d", result[0].StationCount)
	}
}

func TestGetGroupEvaluations_TieBrokenByGroupID(t *testing.T) {
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
		{"Station"},
		{"Weitsprung"},
	}); err != nil {
		t.Fatalf("InsertStations failed: %v", err)
	}
	if err := database.SaveGroups(db, []models.Group{
		{GroupID: 1, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 1}}},
		{GroupID: 2, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 2}}},
	}); err != nil {
		t.Fatalf("SaveGroups failed: %v", err)
	}
	// Both groups score equally
	if err := database.AssignGroupStationScore(db, 1, 1, 85); err != nil {
		t.Fatalf("failed: %v", err)
	}
	if err := database.AssignGroupStationScore(db, 2, 1, 85); err != nil {
		t.Fatalf("failed: %v", err)
	}

	result, err := database.GetGroupEvaluations(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 evaluations, got %d", len(result))
	}
	// With equal total scores, ORDER BY group_id ASC: group 1 before group 2
	if result[0].GroupID != 1 {
		t.Errorf("tie-break: expected group 1 first (lower group_id), got %d", result[0].GroupID)
	}
	if result[1].GroupID != 2 {
		t.Errorf("tie-break: expected group 2 second, got %d", result[1].GroupID)
	}
}

// ---- GetOrtsverbandEvaluations edge cases ----
// Note: TestGetOrtsverbandEvaluations (multi-ortsverband ranking) is already in scores_test.go.

func TestGetOrtsverbandEvaluations_EmptyDB(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	result, err := database.GetOrtsverbandEvaluations(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d evaluations", len(result))
	}
}

func TestGetOrtsverbandEvaluations_NoScoresAssigned(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Participants exist and are grouped, but no scores have been assigned yet.
	// The query uses INNER JOINs, so ortsverbands without scores must not appear.
	if err := database.InsertData(db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
	}); err != nil {
		t.Fatalf("InsertData failed: %v", err)
	}
	if err := database.SaveGroups(db, []models.Group{
		{GroupID: 1, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 1}}},
	}); err != nil {
		t.Fatalf("SaveGroups failed: %v", err)
	}

	result, err := database.GetOrtsverbandEvaluations(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 ortsverbands (no scores yet, inner join), got %d", len(result))
	}
}

func TestGetOrtsverbandEvaluations_AverageCalculation(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Single participant from Berlin, 1 station, score 90
	// average should be 90.0 / 1 = 90.0
	if err := database.InsertData(db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
	}); err != nil {
		t.Fatalf("InsertData failed: %v", err)
	}
	if err := database.InsertStations(db, [][]string{
		{"Station"},
		{"Weitsprung"},
	}); err != nil {
		t.Fatalf("InsertStations failed: %v", err)
	}
	if err := database.SaveGroups(db, []models.Group{
		{GroupID: 1, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 1}}},
	}); err != nil {
		t.Fatalf("SaveGroups failed: %v", err)
	}
	if err := database.AssignGroupStationScore(db, 1, 1, 90); err != nil {
		t.Fatalf("AssignGroupStationScore failed: %v", err)
	}

	result, err := database.GetOrtsverbandEvaluations(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 ortsverband, got %d", len(result))
	}
	if result[0].Ortsverband != "Berlin" {
		t.Errorf("expected 'Berlin', got '%s'", result[0].Ortsverband)
	}
	if result[0].TotalScore != 90 {
		t.Errorf("expected TotalScore 90, got %d", result[0].TotalScore)
	}
	if result[0].ParticipantCount != 1 {
		t.Errorf("expected ParticipantCount 1, got %d", result[0].ParticipantCount)
	}
	if result[0].AverageScore != 90.0 {
		t.Errorf("expected AverageScore 90.0, got %f", result[0].AverageScore)
	}
}

func TestGetOrtsverbandEvaluations_RankedByAverageDescending(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Berlin: 1 participant in group 1 → score 90, avg = 90
	// Hamburg: 2 participants in group 2 → score 60, avg = 60*2/2 = 60
	if err := database.InsertData(db, [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "25", "W", ""},
		{"Bob", "Hamburg", "22", "M", ""},
		{"Carol", "Hamburg", "30", "W", ""},
	}); err != nil {
		t.Fatalf("InsertData failed: %v", err)
	}
	if err := database.InsertStations(db, [][]string{
		{"Station"},
		{"Weitsprung"},
	}); err != nil {
		t.Fatalf("InsertStations failed: %v", err)
	}
	if err := database.SaveGroups(db, []models.Group{
		{GroupID: 1, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 1}}},
		{GroupID: 2, Teilnehmers: []models.Teilnehmer{{TeilnehmerID: 2}, {TeilnehmerID: 3}}},
	}); err != nil {
		t.Fatalf("SaveGroups failed: %v", err)
	}
	if err := database.AssignGroupStationScore(db, 1, 1, 90); err != nil {
		t.Fatalf("failed: %v", err)
	}
	if err := database.AssignGroupStationScore(db, 2, 1, 60); err != nil {
		t.Fatalf("failed: %v", err)
	}

	result, err := database.GetOrtsverbandEvaluations(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 ortsverbands, got %d", len(result))
	}
	// Ranked by average score DESC: Berlin (90) before Hamburg (60)
	if result[0].Ortsverband != "Berlin" {
		t.Errorf("expected 'Berlin' first, got '%s'", result[0].Ortsverband)
	}
	if result[0].AverageScore != 90.0 {
		t.Errorf("expected Berlin average 90.0, got %f", result[0].AverageScore)
	}
	if result[1].Ortsverband != "Hamburg" {
		t.Errorf("expected 'Hamburg' second, got '%s'", result[1].Ortsverband)
	}
	if result[1].AverageScore != 60.0 {
		t.Errorf("expected Hamburg average 60.0, got %f", result[1].AverageScore)
	}
}
