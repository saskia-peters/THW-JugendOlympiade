package test

import (
	"testing"

	"THW-JugendOlympiade/backend/database"
	"THW-JugendOlympiade/backend/models"
)

// TestAssignGroupStationScore_NewScore tests assigning a new score
func TestAssignGroupStationScore_NewScore(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Insert test data
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
	}

	err := database.InsertData(db, rows)
	if err != nil {
		t.Fatalf("Failed to insert participants: %v", err)
	}

	stationRows := [][]string{
		{"Station Name"},
		{"Weitsprung"},
	}

	err = database.InsertStations(db, stationRows)
	if err != nil {
		t.Fatalf("Failed to insert stations: %v", err)
	}

	groups := []models.Group{
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 1}}},
	}

	err = database.SaveGroups(db, groups)
	if err != nil {
		t.Fatalf("Failed to save groups: %v", err)
	}

	// Assign score
	err = database.AssignGroupStationScore(db, 1, 1, 85)
	if err != nil {
		t.Fatalf("AssignGroupStationScore failed: %v", err)
	}

	// Verify score was inserted
	var score int
	err = db.QueryRow("SELECT score FROM group_station_scores WHERE group_id = 1 AND station_id = 1").Scan(&score)
	if err != nil {
		t.Fatalf("Failed to query score: %v", err)
	}

	if score != 85 {
		t.Errorf("Expected score 85, got %d", score)
	}
}

// TestAssignGroupStationScore_UpdateScore tests updating an existing score
func TestAssignGroupStationScore_UpdateScore(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Insert test data
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
	}

	err := database.InsertData(db, rows)
	if err != nil {
		t.Fatalf("Failed to insert participants: %v", err)
	}

	stationRows := [][]string{
		{"Station Name"},
		{"Weitsprung"},
	}

	err = database.InsertStations(db, stationRows)
	if err != nil {
		t.Fatalf("Failed to insert stations: %v", err)
	}

	groups := []models.Group{
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 1}}},
	}

	err = database.SaveGroups(db, groups)
	if err != nil {
		t.Fatalf("Failed to save groups: %v", err)
	}

	// Assign initial score
	err = database.AssignGroupStationScore(db, 1, 1, 85)
	if err != nil {
		t.Fatalf("Failed to assign initial score: %v", err)
	}

	// Update score
	err = database.AssignGroupStationScore(db, 1, 1, 92)
	if err != nil {
		t.Fatalf("Failed to update score: %v", err)
	}

	// Verify score was updated
	var score int
	err = db.QueryRow("SELECT score FROM group_station_scores WHERE group_id = 1 AND station_id = 1").Scan(&score)
	if err != nil {
		t.Fatalf("Failed to query score: %v", err)
	}

	if score != 92 {
		t.Errorf("Expected updated score 92, got %d", score)
	}

	// Verify only one record exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM group_station_scores WHERE group_id = 1 AND station_id = 1").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count scores: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 score record, got %d", count)
	}
}

// TestAssignGroupStationScore_MultipleStations tests scores for multiple stations
func TestAssignGroupStationScore_MultipleStations(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Insert test data
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
	}

	err := database.InsertData(db, rows)
	if err != nil {
		t.Fatalf("Failed to insert participants: %v", err)
	}

	stationRows := [][]string{
		{"Station Name"}, // Header
		{"Weitsprung"},
		{"Ballwurf"},
		{"Sprint"},
	}

	err = database.InsertStations(db, stationRows)
	if err != nil {
		t.Fatalf("Failed to insert stations: %v", err)
	}

	groups := []models.Group{
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 1}}},
	}

	err = database.SaveGroups(db, groups)
	if err != nil {
		t.Fatalf("Failed to save groups: %v", err)
	}

	// Assign scores for different stations
	scores := map[int]int{
		1: 85,
		2: 90,
		3: 78,
	}

	for stationID, score := range scores {
		err = database.AssignGroupStationScore(db, 1, stationID, score)
		if err != nil {
			t.Fatalf("Failed to assign score for station %d: %v", stationID, err)
		}
	}

	// Verify all scores were inserted
	for stationID, expectedScore := range scores {
		var score int
		err = db.QueryRow("SELECT score FROM group_station_scores WHERE group_id = 1 AND station_id = ?", stationID).Scan(&score)
		if err != nil {
			t.Fatalf("Failed to query score for station %d: %v", stationID, err)
		}

		if score != expectedScore {
			t.Errorf("Station %d: expected score %d, got %d", stationID, expectedScore, score)
		}
	}
}

// TestAssignGroupStationScore_MultipleGroups tests scores for multiple groups
func TestAssignGroupStationScore_MultipleGroups(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Insert test data
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
		{"Anna Schmidt", "Hamburg", "30", "W", ""},
		{"Tom Meyer", "München", "22", "M", ""},
	}

	err := database.InsertData(db, rows)
	if err != nil {
		t.Fatalf("Failed to insert participants: %v", err)
	}

	stationRows := [][]string{
		{"Station Name"},
		{"Weitsprung"},
	}

	err = database.InsertStations(db, stationRows)
	if err != nil {
		t.Fatalf("Failed to insert stations: %v", err)
	}

	groups := []models.Group{
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 1}}},
		{GroupID: 2, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 2}}},
		{GroupID: 3, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 3}}},
	}

	err = database.SaveGroups(db, groups)
	if err != nil {
		t.Fatalf("Failed to save groups: %v", err)
	}

	// Assign different scores for each group
	groupScores := map[int]int{
		1: 85,
		2: 90,
		3: 78,
	}

	for groupID, score := range groupScores {
		err = database.AssignGroupStationScore(db, groupID, 1, score)
		if err != nil {
			t.Fatalf("Failed to assign score for group %d: %v", groupID, err)
		}
	}

	// Verify all scores were inserted correctly
	for groupID, expectedScore := range groupScores {
		var score int
		err = db.QueryRow("SELECT score FROM group_station_scores WHERE group_id = ? AND station_id = 1", groupID).Scan(&score)
		if err != nil {
			t.Fatalf("Failed to query score for group %d: %v", groupID, err)
		}

		if score != expectedScore {
			t.Errorf("Group %d: expected score %d, got %d", groupID, expectedScore, score)
		}
	}
}

// TestAssignGroupStationScore_ZeroScore tests assigning a score of 0
func TestAssignGroupStationScore_ZeroScore(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Insert minimal test data
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
	}

	err := database.InsertData(db, rows)
	if err != nil {
		t.Fatalf("Failed to insert participants: %v", err)
	}

	stationRows := [][]string{
		{"Station Name"},
		{"Weitsprung"},
	}

	err = database.InsertStations(db, stationRows)
	if err != nil {
		t.Fatalf("Failed to insert stations: %v", err)
	}

	groups := []models.Group{
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 1}}},
	}

	err = database.SaveGroups(db, groups)
	if err != nil {
		t.Fatalf("Failed to save groups: %v", err)
	}

	// Assign score of 0
	err = database.AssignGroupStationScore(db, 1, 1, 0)
	if err != nil {
		t.Fatalf("Failed to assign zero score: %v", err)
	}

	// Verify score is 0
	var score int
	err = db.QueryRow("SELECT score FROM group_station_scores WHERE group_id = 1 AND station_id = 1").Scan(&score)
	if err != nil {
		t.Fatalf("Failed to query score: %v", err)
	}

	if score != 0 {
		t.Errorf("Expected score 0, got %d", score)
	}
}

// TestGetGroupEvaluations tests retrieving group rankings
func TestGetGroupEvaluations(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Insert test data
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
		{"Anna Schmidt", "Hamburg", "30", "W", ""},
		{"Tom Meyer", "München", "22", "M", ""},
	}

	err := database.InsertData(db, rows)
	if err != nil {
		t.Fatalf("Failed to insert participants: %v", err)
	}

	stationRows := [][]string{
		{"Station Name"}, {"Weitsprung"}, {"Ballwurf"},
		{"Weitsprung", "Ballwurf"},
	}

	err = database.InsertStations(db, stationRows)
	if err != nil {
		t.Fatalf("Failed to insert stations: %v", err)
	}

	groups := []models.Group{
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 1}}},
		{GroupID: 2, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 2}}},
		{GroupID: 3, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 3}}},
	}

	err = database.SaveGroups(db, groups)
	if err != nil {
		t.Fatalf("Failed to save groups: %v", err)
	}

	// Assign scores: Group 2 should win with 180 points
	// Group 1: 85 + 90 = 175
	database.AssignGroupStationScore(db, 1, 1, 85)
	database.AssignGroupStationScore(db, 1, 2, 90)

	// Group 2: 90 + 90 = 180 (highest)
	database.AssignGroupStationScore(db, 2, 1, 90)
	database.AssignGroupStationScore(db, 2, 2, 90)

	// Group 3: 70 + 80 = 150
	database.AssignGroupStationScore(db, 3, 1, 70)
	database.AssignGroupStationScore(db, 3, 2, 80)

	// Get evaluations
	evaluations, err := database.GetGroupEvaluations(db)
	if err != nil {
		t.Fatalf("GetGroupEvaluations failed: %v", err)
	}

	if len(evaluations) != 3 {
		t.Errorf("Expected 3 evaluations, got %d", len(evaluations))
	}

	// Verify ranking order (highest score first)
	if evaluations[0].GroupID != 2 {
		t.Errorf("Expected first place to be Group 2, got Group %d", evaluations[0].GroupID)
	}
	if evaluations[0].TotalScore != 180 {
		t.Errorf("Expected Group 2 total score 180, got %d", evaluations[0].TotalScore)
	}

	if evaluations[1].GroupID != 1 {
		t.Errorf("Expected second place to be Group 1, got Group %d", evaluations[1].GroupID)
	}
	if evaluations[1].TotalScore != 175 {
		t.Errorf("Expected Group 1 total score 175, got %d", evaluations[1].TotalScore)
	}

	if evaluations[2].GroupID != 3 {
		t.Errorf("Expected third place to be Group 3, got Group %d", evaluations[2].GroupID)
	}
	if evaluations[2].TotalScore != 150 {
		t.Errorf("Expected Group 3 total score 150, got %d", evaluations[2].TotalScore)
	}
}

// TestGetOrtsverbandEvaluations tests retrieving ortsverband rankings
func TestGetOrtsverbandEvaluations(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Insert test data with multiple ortsverbands
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Max Mustermann", "Berlin", "25", "M", ""},
		{"Anna Schmidt", "Berlin", "30", "W", ""},
		{"Tom Meyer", "Hamburg", "22", "M", ""},
		{"Lisa Weber", "Hamburg", "24", "W", ""},
		{"John Doe", "München", "28", "M", ""},
	}

	err := database.InsertData(db, rows)
	if err != nil {
		t.Fatalf("Failed to insert participants: %v", err)
	}

	stationRows := [][]string{
		{"Station Name"},
		{"Weitsprung"},
	}

	err = database.InsertStations(db, stationRows)
	if err != nil {
		t.Fatalf("Failed to insert stations: %v", err)
	}

	// Create groups with mixed ortsverbands
	groups := []models.Group{
		{GroupID: 1, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 1}, {TeilnehmendeID: 3}}}, // Berlin + Hamburg
		{GroupID: 2, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 2}, {TeilnehmendeID: 4}}}, // Berlin + Hamburg
		{GroupID: 3, Teilnehmende: []models.Teilnehmende{{TeilnehmendeID: 5}}},                    // München
	}

	err = database.SaveGroups(db, groups)
	if err != nil {
		t.Fatalf("Failed to save groups: %v", err)
	}

	// Assign scores
	database.AssignGroupStationScore(db, 1, 1, 90) // Berlin gets 45, Hamburg gets 45
	database.AssignGroupStationScore(db, 2, 1, 80) // Berlin gets 40, Hamburg gets 40
	database.AssignGroupStationScore(db, 3, 1, 60) // München gets 60

	// Get ortsverband evaluations
	evaluations, err := database.GetOrtsverbandEvaluations(db)
	if err != nil {
		t.Fatalf("GetOrtsverbandEvaluations failed: %v", err)
	}

	if len(evaluations) != 3 {
		t.Errorf("Expected 3 ortsverband evaluations, got %d", len(evaluations))
	}

	// Verify each ortsverband appears
	ortsverbandMap := make(map[string]int)
	for _, eval := range evaluations {
		ortsverbandMap[eval.Ortsverband] = eval.TotalScore
	}

	if _, exists := ortsverbandMap["Berlin"]; !exists {
		t.Error("Expected Berlin in evaluations")
	}
	if _, exists := ortsverbandMap["Hamburg"]; !exists {
		t.Error("Expected Hamburg in evaluations")
	}
	if _, exists := ortsverbandMap["München"]; !exists {
		t.Error("Expected München in evaluations")
	}
}
