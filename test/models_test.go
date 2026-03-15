package test

import (
	"testing"

	"THW-JugendOlympiade/backend/models"
)

// TestMaxGroupSizeConstant tests the MaxGroupSize constant
func TestMaxGroupSizeConstant(t *testing.T) {
	if models.MaxGroupSize != 8 {
		t.Errorf("Expected MaxGroupSize to be 8, got %d", models.MaxGroupSize)
	}
}

// TestDbFileConstant tests the DbFile constant
func TestDbFileConstant(t *testing.T) {
	expected := "data.db"
	if models.DbFile != expected {
		t.Errorf("Expected DbFile to be %q, got %q", expected, models.DbFile)
	}
}

// TestSheetNameConstant tests the SheetName constant
func TestSheetNameConstant(t *testing.T) {
	expected := "Teilnehmer"
	if models.SheetName != expected {
		t.Errorf("Expected SheetName to be %q, got %q", expected, models.SheetName)
	}
}

// TestTeilnehmerStructure tests the Teilnehmer struct fields
func TestTeilnehmerStructure(t *testing.T) {
	teilnehmer := models.Teilnehmer{
		ID:           1,
		TeilnehmerID: 100,
		Name:         "Max Mustermann",
		Ortsverband:  "Berlin",
		Alter:        25,
		Geschlecht:   "M",
	}

	if teilnehmer.ID != 1 {
		t.Errorf("Expected ID 1, got %d", teilnehmer.ID)
	}
	if teilnehmer.TeilnehmerID != 100 {
		t.Errorf("Expected TeilnehmerID 100, got %d", teilnehmer.TeilnehmerID)
	}
	if teilnehmer.Name != "Max Mustermann" {
		t.Errorf("Expected Name 'Max Mustermann', got %s", teilnehmer.Name)
	}
	if teilnehmer.Ortsverband != "Berlin" {
		t.Errorf("Expected Ortsverband 'Berlin', got %s", teilnehmer.Ortsverband)
	}
	if teilnehmer.Alter != 25 {
		t.Errorf("Expected Alter 25, got %d", teilnehmer.Alter)
	}
	if teilnehmer.Geschlecht != "M" {
		t.Errorf("Expected Geschlecht 'M', got %s", teilnehmer.Geschlecht)
	}
}

// TestGroupStructure tests the Group struct fields and initialization
func TestGroupStructure(t *testing.T) {
	group := models.Group{
		GroupID:      1,
		Teilnehmers:  make([]models.Teilnehmer, 0),
		Ortsverbands: make(map[string]int),
		Geschlechts:  make(map[string]int),
		AlterSum:     0,
	}

	if group.GroupID != 1 {
		t.Errorf("Expected GroupID 1, got %d", group.GroupID)
	}

	if len(group.Teilnehmers) != 0 {
		t.Errorf("Expected empty Teilnehmers slice, got length %d", len(group.Teilnehmers))
	}

	if len(group.Ortsverbands) != 0 {
		t.Errorf("Expected empty Ortsverbands map, got length %d", len(group.Ortsverbands))
	}

	if len(group.Geschlechts) != 0 {
		t.Errorf("Expected empty Geschlechts map, got length %d", len(group.Geschlechts))
	}

	if group.AlterSum != 0 {
		t.Errorf("Expected AlterSum 0, got %d", group.AlterSum)
	}
}

// TestGroupWithParticipants tests adding participants to a group
func TestGroupWithParticipants(t *testing.T) {
	group := models.Group{
		GroupID:      1,
		Teilnehmers:  make([]models.Teilnehmer, 0),
		Ortsverbands: make(map[string]int),
		Geschlechts:  make(map[string]int),
		AlterSum:     0,
	}

	// Add participants
	participants := []models.Teilnehmer{
		{ID: 1, Name: "Max", Ortsverband: "Berlin", Alter: 25, Geschlecht: "M"},
		{ID: 2, Name: "Anna", Ortsverband: "Berlin", Alter: 30, Geschlecht: "W"},
		{ID: 3, Name: "Tom", Ortsverband: "Hamburg", Alter: 22, Geschlecht: "M"},
	}

	for _, p := range participants {
		group.Teilnehmers = append(group.Teilnehmers, p)
		group.Ortsverbands[p.Ortsverband]++
		group.Geschlechts[p.Geschlecht]++
		group.AlterSum += p.Alter
	}

	// Verify counts
	if len(group.Teilnehmers) != 3 {
		t.Errorf("Expected 3 participants, got %d", len(group.Teilnehmers))
	}

	if group.Ortsverbands["Berlin"] != 2 {
		t.Errorf("Expected 2 from Berlin, got %d", group.Ortsverbands["Berlin"])
	}

	if group.Ortsverbands["Hamburg"] != 1 {
		t.Errorf("Expected 1 from Hamburg, got %d", group.Ortsverbands["Hamburg"])
	}

	if group.Geschlechts["M"] != 2 {
		t.Errorf("Expected 2 male, got %d", group.Geschlechts["M"])
	}

	if group.Geschlechts["W"] != 1 {
		t.Errorf("Expected 1 female, got %d", group.Geschlechts["W"])
	}

	expectedAlterSum := 25 + 30 + 22
	if group.AlterSum != expectedAlterSum {
		t.Errorf("Expected AlterSum %d, got %d", expectedAlterSum, group.AlterSum)
	}
}

// TestStationStructure tests the Station struct
func TestStationStructure(t *testing.T) {
	station := models.Station{
		StationID:   1,
		StationName: "Weitsprung",
		GroupScores: make([]models.GroupScore, 0),
	}

	if station.StationID != 1 {
		t.Errorf("Expected StationID 1, got %d", station.StationID)
	}

	if station.StationName != "Weitsprung" {
		t.Errorf("Expected StationName 'Weitsprung', got %s", station.StationName)
	}

	if len(station.GroupScores) != 0 {
		t.Errorf("Expected empty GroupScores, got length %d", len(station.GroupScores))
	}
}

// TestGroupScoreStructure tests the GroupScore struct
func TestGroupScoreStructure(t *testing.T) {
	score := models.GroupScore{
		GroupID: 1,
		Score:   85,
	}

	if score.GroupID != 1 {
		t.Errorf("Expected GroupID 1, got %d", score.GroupID)
	}

	if score.Score != 85 {
		t.Errorf("Expected Score 85, got %d", score.Score)
	}
}

// TestGroupEvaluationStructure tests the GroupEvaluation struct
func TestGroupEvaluationStructure(t *testing.T) {
	eval := models.GroupEvaluation{
		GroupID:    1,
		TotalScore: 350,
	}

	if eval.GroupID != 1 {
		t.Errorf("Expected GroupID 1, got %d", eval.GroupID)
	}

	if eval.TotalScore != 350 {
		t.Errorf("Expected TotalScore 350, got %d", eval.TotalScore)
	}
}

// TestOrtsverbandEvaluationStructure tests the OrtsverbandEvaluation struct
func TestOrtsverbandEvaluationStructure(t *testing.T) {
	eval := models.OrtsverbandEvaluation{
		Ortsverband: "Berlin",
		TotalScore:  450,
	}

	if eval.Ortsverband != "Berlin" {
		t.Errorf("Expected Ortsverband 'Berlin', got %s", eval.Ortsverband)
	}

	if eval.TotalScore != 450 {
		t.Errorf("Expected TotalScore 450, got %d", eval.TotalScore)
	}
}

// TestGroupCapacity tests that groups can hold up to MaxGroupSize
func TestGroupCapacity(t *testing.T) {
	group := models.Group{
		GroupID:      1,
		Teilnehmers:  make([]models.Teilnehmer, 0, models.MaxGroupSize),
		Ortsverbands: make(map[string]int),
		Geschlechts:  make(map[string]int),
	}

	// Add exactly MaxGroupSize participants
	for i := 0; i < models.MaxGroupSize; i++ {
		participant := models.Teilnehmer{
			ID:    i + 1,
			Name:  "Participant",
			Alter: 20,
		}
		group.Teilnehmers = append(group.Teilnehmers, participant)
	}

	if len(group.Teilnehmers) != models.MaxGroupSize {
		t.Errorf("Expected %d participants, got %d", models.MaxGroupSize, len(group.Teilnehmers))
	}

	// Verify capacity was pre-allocated correctly
	if cap(group.Teilnehmers) < models.MaxGroupSize {
		t.Errorf("Expected capacity >= %d, got %d", models.MaxGroupSize, cap(group.Teilnehmers))
	}
}

// TestEmptyTeilnehmer tests creating an empty Teilnehmer
func TestEmptyTeilnehmer(t *testing.T) {
	var teilnehmer models.Teilnehmer

	// Verify zero values
	if teilnehmer.ID != 0 {
		t.Errorf("Expected zero ID, got %d", teilnehmer.ID)
	}
	if teilnehmer.Name != "" {
		t.Errorf("Expected empty Name, got %s", teilnehmer.Name)
	}
	if teilnehmer.Alter != 0 {
		t.Errorf("Expected zero Alter, got %d", teilnehmer.Alter)
	}
}

// TestEmptyGroup tests creating an empty Group
func TestEmptyGroup(t *testing.T) {
	var group models.Group

	// Verify zero values
	if group.GroupID != 0 {
		t.Errorf("Expected zero GroupID, got %d", group.GroupID)
	}
	if group.Teilnehmers != nil {
		t.Error("Expected nil Teilnehmers slice")
	}
	if group.Ortsverbands != nil {
		t.Error("Expected nil Ortsverbands map")
	}
	if group.AlterSum != 0 {
		t.Errorf("Expected zero AlterSum, got %d", group.AlterSum)
	}
}

// TestTeilnehmerWithMissingFields tests Teilnehmer with optional empty fields
func TestTeilnehmerWithMissingFields(t *testing.T) {
	teilnehmer := models.Teilnehmer{
		ID:           1,
		TeilnehmerID: 100,
		Name:         "Max Mustermann",
		Ortsverband:  "", // Empty ortsverband
		Alter:        0,  // Zero age
		Geschlecht:   "", // Empty geschlecht
	}

	if teilnehmer.Ortsverband != "" {
		t.Errorf("Expected empty Ortsverband, got %s", teilnehmer.Ortsverband)
	}
	if teilnehmer.Alter != 0 {
		t.Errorf("Expected zero Alter, got %d", teilnehmer.Alter)
	}
	if teilnehmer.Geschlecht != "" {
		t.Errorf("Expected empty Geschlecht, got %s", teilnehmer.Geschlecht)
	}
}

// TestMultipleGroupsWithSameParticipants tests that different groups can exist
func TestMultipleGroupsWithSameStructure(t *testing.T) {
	group1 := models.Group{
		GroupID:      1,
		Teilnehmers:  make([]models.Teilnehmer, 0),
		Ortsverbands: make(map[string]int),
		Geschlechts:  make(map[string]int),
	}

	group2 := models.Group{
		GroupID:      2,
		Teilnehmers:  make([]models.Teilnehmer, 0),
		Ortsverbands: make(map[string]int),
		Geschlechts:  make(map[string]int),
	}

	// Add same participant structure to both groups
	participant := models.Teilnehmer{
		ID:   1,
		Name: "Test",
	}

	group1.Teilnehmers = append(group1.Teilnehmers, participant)
	group2.Teilnehmers = append(group2.Teilnehmers, participant)

	// Verify groups are independent
	if group1.GroupID == group2.GroupID {
		t.Error("Expected different GroupIDs")
	}

	if len(group1.Teilnehmers) != 1 || len(group2.Teilnehmers) != 1 {
		t.Error("Both groups should have 1 participant each")
	}
}
