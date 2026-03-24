package test

import (
	"math"
	"testing"

	"THW-JugendOlympiade/backend/database"
	"THW-JugendOlympiade/backend/models"
	"THW-JugendOlympiade/backend/services"
)

// testMaxGroupSize is the fixed group size used by the distribution tests.
// It is intentionally independent of the config default so that test behaviour
// does not change when the operator adjusts their configuration.
const testMaxGroupSize = 8

// mockDistributeIntoGroups mimics the distribution logic for testing without DB
// This is a test helper that tests the algorithm logic independently
func mockDistributeIntoGroups(teilnehmende []models.Teilnehmende) []models.Group {
	if len(teilnehmende) == 0 {
		return nil
	}

	// Calculate number of groups needed
	numGroups := int(math.Ceil(float64(len(teilnehmende)) / float64(testMaxGroupSize)))

	// Initialize groups
	groups := make([]models.Group, numGroups)
	for i := range groups {
		groups[i] = models.Group{
			GroupID:      i + 1,
			Teilnehmende: make([]models.Teilnehmende, 0, testMaxGroupSize),
			Ortsverbands: make(map[string]int),
			Geschlechts:  make(map[string]int),
		}
	}

	// Simple round-robin distribution for testing
	for i, tn := range teilnehmende {
		groupIdx := i % numGroups
		group := &groups[groupIdx]
		group.Teilnehmende = append(group.Teilnehmende, tn)
		group.Ortsverbands[tn.Ortsverband]++
		group.Geschlechts[tn.Geschlecht]++
		group.AlterSum += tn.Alter
	}

	return groups
}

// TestDistribution_EmptyInput tests distribution with no participants
func TestDistribution_EmptyInput(t *testing.T) {
	teilnehmende := []models.Teilnehmende{}
	groups := mockDistributeIntoGroups(teilnehmende)

	if groups != nil {
		t.Errorf("Expected nil for empty input, got %d groups", len(groups))
	}
}

// TestDistribution_SingleParticipant tests distribution with one participant
func TestDistribution_SingleParticipant(t *testing.T) {
	teilnehmende := []models.Teilnehmende{
		{ID: 1, TeilnehmendeID: 1, Name: "Max Mustermann", Ortsverband: "Berlin", Alter: 25, Geschlecht: "M"},
	}

	groups := mockDistributeIntoGroups(teilnehmende)

	if len(groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups))
	}

	if len(groups[0].Teilnehmende) != 1 {
		t.Errorf("Expected 1 participant in group, got %d", len(groups[0].Teilnehmende))
	}

	if groups[0].GroupID != 1 {
		t.Errorf("Expected GroupID 1, got %d", groups[0].GroupID)
	}
}

// TestDistribution_ExactlyMaxGroupSize tests distribution with exactly testMaxGroupSize participants
func TestDistribution_ExactlyMaxGroupSize(t *testing.T) {
	teilnehmende := make([]models.Teilnehmende, testMaxGroupSize)
	for i := 0; i < testMaxGroupSize; i++ {
		teilnehmende[i] = models.Teilnehmende{
			ID:             i + 1,
			TeilnehmendeID: i + 1,
			Name:           "Participant " + string(rune(i)),
			Ortsverband:    "Berlin",
			Alter:          20 + i,
			Geschlecht:     "M",
		}
	}

	groups := mockDistributeIntoGroups(teilnehmende)

	if len(groups) != 1 {
		t.Errorf("Expected 1 group for %d participants, got %d", testMaxGroupSize, len(groups))
	}

	if len(groups[0].Teilnehmende) != testMaxGroupSize {
		t.Errorf("Expected %d participants in group, got %d", testMaxGroupSize, len(groups[0].Teilnehmende))
	}
}

// TestDistribution_MoreThanMaxGroupSize tests distribution requiring multiple groups
func TestDistribution_MoreThanMaxGroupSize(t *testing.T) {
	numParticipants := testMaxGroupSize + 1
	teilnehmende := make([]models.Teilnehmende, numParticipants)
	for i := 0; i < numParticipants; i++ {
		teilnehmende[i] = models.Teilnehmende{
			ID:             i + 1,
			TeilnehmendeID: i + 1,
			Name:           "Participant " + string(rune(i)),
			Ortsverband:    "Berlin",
			Alter:          20 + i,
			Geschlecht:     "M",
		}
	}

	groups := mockDistributeIntoGroups(teilnehmende)

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups for %d participants, got %d", numParticipants, len(groups))
	}

	// Check that all participants are distributed
	totalParticipants := 0
	for _, group := range groups {
		totalParticipants += len(group.Teilnehmende)
	}

	if totalParticipants != numParticipants {
		t.Errorf("Expected %d total participants, got %d", numParticipants, totalParticipants)
	}
}

// TestDistribution_TwentyFourParticipants tests a realistic scenario
func TestDistribution_TwentyFourParticipants(t *testing.T) {
	// 24 participants should create 3 groups of 8
	teilnehmende := make([]models.Teilnehmende, 24)
	ortsverbands := []string{"Berlin", "Hamburg", "München", "Köln"}
	geschlechts := []string{"M", "W"}

	for i := 0; i < 24; i++ {
		teilnehmende[i] = models.Teilnehmende{
			ID:             i + 1,
			TeilnehmendeID: i + 1,
			Name:           "Participant " + string(rune(i)),
			Ortsverband:    ortsverbands[i%len(ortsverbands)],
			Alter:          18 + (i % 10),
			Geschlecht:     geschlechts[i%len(geschlechts)],
		}
	}

	groups := mockDistributeIntoGroups(teilnehmende)

	if len(groups) != 3 {
		t.Errorf("Expected 3 groups for 24 participants, got %d", len(groups))
	}

	// Check that each group has exactly 8 participants
	for i, group := range groups {
		if len(group.Teilnehmende) != 8 {
			t.Errorf("Group %d: Expected 8 participants, got %d", i+1, len(group.Teilnehmende))
		}
	}

	// Verify statistics are tracked
	for i, group := range groups {
		if len(group.Ortsverbands) == 0 {
			t.Errorf("Group %d: Ortsverbands map is empty", i+1)
		}
		if len(group.Geschlechts) == 0 {
			t.Errorf("Group %d: Geschlechts map is empty", i+1)
		}
		if group.AlterSum == 0 {
			t.Errorf("Group %d: AlterSum is 0", i+1)
		}
	}
}

// TestDistribution_GroupSizeLimit tests that no group exceeds MaxGroupSize
func TestDistribution_GroupSizeLimit(t *testing.T) {
	// Test with 50 participants
	teilnehmende := make([]models.Teilnehmende, 50)
	for i := 0; i < 50; i++ {
		teilnehmende[i] = models.Teilnehmende{
			ID:             i + 1,
			TeilnehmendeID: i + 1,
			Name:           "Participant " + string(rune(i)),
			Ortsverband:    "Berlin",
			Alter:          20 + (i % 20),
			Geschlecht:     "M",
		}
	}

	groups := mockDistributeIntoGroups(teilnehmende)

	// 50 participants / 8 max = 7 groups (with last group having 2)
	expectedGroups := int(math.Ceil(50.0 / float64(testMaxGroupSize)))
	if len(groups) != expectedGroups {
		t.Errorf("Expected %d groups, got %d", expectedGroups, len(groups))
	}

	// Verify no group exceeds max size
	for i, group := range groups {
		if len(group.Teilnehmende) > testMaxGroupSize {
			t.Errorf("Group %d exceeds max size: has %d participants (max %d)",
				i+1, len(group.Teilnehmende), testMaxGroupSize)
		}
	}

	// Verify all participants are assigned
	totalAssigned := 0
	for _, group := range groups {
		totalAssigned += len(group.Teilnehmende)
	}
	if totalAssigned != 50 {
		t.Errorf("Expected 50 participants assigned, got %d", totalAssigned)
	}
}

// TestDistribution_StatisticsTracking tests that group statistics are correctly maintained
func TestDistribution_StatisticsTracking(t *testing.T) {
	teilnehmende := []models.Teilnehmende{
		{ID: 1, TeilnehmendeID: 1, Name: "Anna", Ortsverband: "Berlin", Alter: 20, Geschlecht: "W"},
		{ID: 2, TeilnehmendeID: 2, Name: "Max", Ortsverband: "Berlin", Alter: 25, Geschlecht: "M"},
		{ID: 3, TeilnehmendeID: 3, Name: "Lisa", Ortsverband: "Hamburg", Alter: 22, Geschlecht: "W"},
		{ID: 4, TeilnehmendeID: 4, Name: "Tom", Ortsverband: "Hamburg", Alter: 24, Geschlecht: "M"},
	}

	groups := mockDistributeIntoGroups(teilnehmende)

	if len(groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups))
	}

	group := groups[0]

	// Check Ortsverband counts
	if group.Ortsverbands["Berlin"] != 2 {
		t.Errorf("Expected 2 from Berlin, got %d", group.Ortsverbands["Berlin"])
	}
	if group.Ortsverbands["Hamburg"] != 2 {
		t.Errorf("Expected 2 from Hamburg, got %d", group.Ortsverbands["Hamburg"])
	}

	// Check Geschlecht counts
	if group.Geschlechts["M"] != 2 {
		t.Errorf("Expected 2 male, got %d", group.Geschlechts["M"])
	}
	if group.Geschlechts["W"] != 2 {
		t.Errorf("Expected 2 female, got %d", group.Geschlechts["W"])
	}

	// Check age sum
	expectedAlterSum := 20 + 25 + 22 + 24
	if group.AlterSum != expectedAlterSum {
		t.Errorf("Expected AlterSum %d, got %d", expectedAlterSum, group.AlterSum)
	}
}

// TestDistribution_GroupIDsSequential tests that GroupIDs are sequential starting from 1
func TestDistribution_GroupIDsSequential(t *testing.T) {
	teilnehmende := make([]models.Teilnehmende, 20)
	for i := 0; i < 20; i++ {
		teilnehmende[i] = models.Teilnehmende{
			ID:             i + 1,
			TeilnehmendeID: i + 1,
			Name:           "Participant " + string(rune(i)),
			Ortsverband:    "Berlin",
			Alter:          20,
			Geschlecht:     "M",
		}
	}

	groups := mockDistributeIntoGroups(teilnehmende)

	// Verify GroupIDs are sequential starting from 1
	for i, group := range groups {
		expectedID := i + 1
		if group.GroupID != expectedID {
			t.Errorf("Group %d: Expected GroupID %d, got %d", i, expectedID, group.GroupID)
		}
	}
}

// TestDistribution_PreGroupsStayTogether tests that participants with the same PreGroup value are grouped together
func TestDistribution_PreGroupsStayTogether(t *testing.T) {
	db := setupFullTestDB(t)
	defer teardownTestDB(t, db)

	// Insert participants with PreGroup values
	rows := [][]string{
		{"Name", "Ortsverband", "Alter", "Geschlecht", "PreGroup"},
		{"Alice", "Berlin", "20", "W", "TeamA"},
		{"Bob", "Hamburg", "22", "M", "TeamA"},
		{"Carol", "München", "21", "W", "TeamA"},
		{"Dave", "Köln", "23", "M", "TeamB"},
		{"Eve", "Berlin", "24", "W", "TeamB"},
		{"Frank", "Hamburg", "25", "M", ""}, // No PreGroup
		{"Grace", "München", "26", "W", ""}, // No PreGroup
		{"Henry", "Köln", "27", "M", ""},    // No PreGroup
		{"Ivy", "Berlin", "28", "W", ""},    // No PreGroup
		{"Jack", "Hamburg", "29", "M", ""},  // No PreGroup
	}

	err := database.InsertData(db, rows)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Run the distribution algorithm
	_, err = services.CreateBalancedGroups(db, 8)
	if err != nil {
		t.Fatalf("Failed to create balanced groups: %v", err)
	}

	// Retrieve groups
	groups, err := database.GetGroupsForReport(db)
	if err != nil {
		t.Fatalf("Failed to get groups: %v", err)
	}

	// Find which group TeamA members are in
	teamAGroupID := -1
	teamAMembers := []string{"Alice", "Bob", "Carol"}
	for _, group := range groups {
		for _, tn := range group.Teilnehmende {
			if tn.Name == "Alice" || tn.Name == "Bob" || tn.Name == "Carol" {
				if teamAGroupID == -1 {
					teamAGroupID = group.GroupID
				} else if teamAGroupID != group.GroupID {
					t.Fatalf("TeamA members are in different groups: expected all in group %d, but found %s in group %d",
						teamAGroupID, tn.Name, group.GroupID)
				}
			}
		}
	}

	// Verify all TeamA members are found and in the same group
	if teamAGroupID == -1 {
		t.Fatal("TeamA members not found in any group")
	}

	// Count TeamA members in the correct group
	teamACount := 0
	for _, group := range groups {
		if group.GroupID == teamAGroupID {
			for _, tn := range group.Teilnehmende {
				for _, name := range teamAMembers {
					if tn.Name == name {
						teamACount++
					}
				}
			}
		}
	}

	if teamACount != 3 {
		t.Errorf("Expected 3 TeamA members in group %d, got %d", teamAGroupID, teamACount)
	}

	// Find which group TeamB members are in
	teamBGroupID := -1
	teamBMembers := []string{"Dave", "Eve"}
	for _, group := range groups {
		for _, tn := range group.Teilnehmende {
			if tn.Name == "Dave" || tn.Name == "Eve" {
				if teamBGroupID == -1 {
					teamBGroupID = group.GroupID
				} else if teamBGroupID != group.GroupID {
					t.Fatalf("TeamB members are in different groups: expected all in group %d, but found %s in group %d",
						teamBGroupID, tn.Name, group.GroupID)
				}
			}
		}
	}

	// Verify all TeamB members are found and in the same group
	if teamBGroupID == -1 {
		t.Fatal("TeamB members not found in any group")
	}

	// Count TeamB members in the correct group
	teamBCount := 0
	for _, group := range groups {
		if group.GroupID == teamBGroupID {
			for _, tn := range group.Teilnehmende {
				for _, name := range teamBMembers {
					if tn.Name == name {
						teamBCount++
					}
				}
			}
		}
	}

	if teamBCount != 2 {
		t.Errorf("Expected 2 TeamB members in group %d, got %d", teamBGroupID, teamBCount)
	}

	// Verify TeamA and TeamB are in different groups
	if teamAGroupID == teamBGroupID {
		t.Errorf("TeamA and TeamB should be in different groups, but both are in group %d", teamAGroupID)
	}

	// Verify participants without PreGroup are distributed
	unassignedCount := 0
	for _, group := range groups {
		for _, tn := range group.Teilnehmende {
			if tn.Name == "Frank" || tn.Name == "Grace" || tn.Name == "Henry" || tn.Name == "Ivy" || tn.Name == "Jack" {
				unassignedCount++
			}
		}
	}

	if unassignedCount != 5 {
		t.Errorf("Expected 5 unassigned participants to be distributed, got %d", unassignedCount)
	}
}
