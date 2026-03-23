package services

import (
	"database/sql"
	"fmt"
	"math"
	"sort"

	"THW-JugendOlympiade/backend/database"
	"THW-JugendOlympiade/backend/models"
)

// CreateBalancedGroups creates groups with balanced distribution.
// maxGroupSize controls the maximum number of participants per group.
func CreateBalancedGroups(db *sql.DB, maxGroupSize int) error {
	// Read all participants from database
	teilnehmende, err := database.GetAllTeilnehmende(db)
	if err != nil {
		return fmt.Errorf("failed to read teilnehmende: %w", err)
	}

	if len(teilnehmende) == 0 {
		return nil // No participants to group
	}

	// Create balanced groups using the distribution algorithm
	groups := distributeIntoGroups(teilnehmende, maxGroupSize)

	// Save groups to database
	if err := database.SaveGroups(db, groups); err != nil {
		return fmt.Errorf("failed to save groups: %w", err)
	}

	// Distribute betreuende across groups (Ortsverband-matched, then round-robin)
	betreuende, err := database.GetAllBetreuende(db)
	if err != nil {
		return fmt.Errorf("failed to read betreuende: %w", err)
	}
	if len(betreuende) > 0 {
		distributeBetreuende(groups, betreuende)
		if err := database.SaveGroupBetreuende(db, groups); err != nil {
			return fmt.Errorf("failed to save group betreuende: %w", err)
		}
	}

	fmt.Printf("Created %d groups with balanced distribution\n", len(groups))
	for i, group := range groups {
		fmt.Printf("  Group %d: %d participants\n", i+1, len(group.Teilnehmende))
	}

	return nil
}

// distributeIntoGroups distributes participants into balanced groups
func distributeIntoGroups(teilnehmende []models.Teilnehmende, maxGroupSize int) []models.Group {
	if len(teilnehmende) == 0 {
		return nil
	}

	// Step 1: Separate participants with and without PreGroup
	preGroupMap := make(map[string][]models.Teilnehmende)
	var unassignedParticipants []models.Teilnehmende

	for _, t := range teilnehmende {
		if t.PreGroup != "" {
			preGroupMap[t.PreGroup] = append(preGroupMap[t.PreGroup], t)
		} else {
			unassignedParticipants = append(unassignedParticipants, t)
		}
	}

	// Step 2: Calculate number of groups needed
	// Count pre-formed groups
	numPreGroups := len(preGroupMap)
	// Calculate how many additional groups needed for unassigned participants
	numAdditionalGroups := int(math.Ceil(float64(len(unassignedParticipants)) / float64(maxGroupSize)))
	numGroups := numPreGroups + numAdditionalGroups

	// Ensure we have at least one group
	if numGroups == 0 {
		numGroups = 1
	}

	// Step 3: Initialize groups
	groups := make([]models.Group, numGroups)
	for i := range groups {
		groups[i] = models.Group{
			GroupID:      i + 1,
			Teilnehmende:  make([]models.Teilnehmende, 0, maxGroupSize),
			Ortsverbands: make(map[string]int),
			Geschlechts:  make(map[string]int),
		}
	}

	// Step 4: Assign pre-grouped participants to the first groups
	groupIdx := 0
	for _, preGroupMembers := range preGroupMap {
		// Add all members of this pre-group to the current group
		for _, t := range preGroupMembers {
			addTeilnehmendeToGroup(&groups[groupIdx], t)
		}
		groupIdx++
	}

	// Step 5: Sort unassigned participants for better distribution
	// First by Ortsverband, then by Geschlecht, then by Alter
	sort.Slice(unassignedParticipants, func(i, j int) bool {
		if unassignedParticipants[i].Ortsverband != unassignedParticipants[j].Ortsverband {
			return unassignedParticipants[i].Ortsverband < unassignedParticipants[j].Ortsverband
		}
		if unassignedParticipants[i].Geschlecht != unassignedParticipants[j].Geschlecht {
			return unassignedParticipants[i].Geschlecht < unassignedParticipants[j].Geschlecht
		}
		return unassignedParticipants[i].Alter < unassignedParticipants[j].Alter
	})

	// Step 6: Distribute unassigned participants using round-robin with diversity scoring
	for _, tn := range unassignedParticipants {
		bestGroupIdx := findBestGroup(groups, tn, maxGroupSize)
		addTeilnehmendeToGroup(&groups[bestGroupIdx], tn)
	}

	return groups
}

// findBestGroup finds the best group for a participant based on diversity
func findBestGroup(groups []models.Group, tn models.Teilnehmende, maxGroupSize int) int {
	bestIdx := 0
	bestScore := math.MaxFloat64

	for i, group := range groups {
		// Skip if group is full
		if len(group.Teilnehmende) >= maxGroupSize {
			continue
		}

		// Calculate diversity score (lower is better)
		score := calculateDiversityScore(group, tn)

		// Prefer groups with fewer members
		sizeBonus := float64(len(group.Teilnehmende)) * 0.5

		totalScore := score + sizeBonus

		if totalScore < bestScore {
			bestScore = totalScore
			bestIdx = i
		}
	}

	return bestIdx
}

// calculateDiversityScore calculates how well a participant fits in a group
// Lower score means better diversity
func calculateDiversityScore(group models.Group, tn models.Teilnehmende) float64 {
	if len(group.Teilnehmende) == 0 {
		return 0
	}

	score := 0.0

	// Penalize if Ortsverband is already common in the group
	ortsverbandCount := group.Ortsverbands[tn.Ortsverband]
	score += float64(ortsverbandCount) * 2.0

	// Penalize if Geschlecht is already common in the group
	geschlechtCount := group.Geschlechts[tn.Geschlecht]
	score += float64(geschlechtCount) * 1.5

	// Penalize if Alter is too similar to group average
	if len(group.Teilnehmende) > 0 && tn.Alter > 0 {
		avgAlter := float64(group.AlterSum) / float64(len(group.Teilnehmende))
		alterDiff := math.Abs(float64(tn.Alter) - avgAlter)
		if alterDiff < 2 {
			score += 1.0
		}
	}

	return score
}

// addTeilnehmendeToGroup adds a participant to the group and updates statistics
func addTeilnehmendeToGroup(g *models.Group, t models.Teilnehmende) {
	g.Teilnehmende = append(g.Teilnehmende, t)
	g.Ortsverbands[t.Ortsverband]++
	g.Geschlechts[t.Geschlecht]++
	g.AlterSum += t.Alter
}

// distributeBetreuende assigns caretakers to groups.
// Each caretaker is first matched to the group that has the most participants from
// the same Ortsverband. Unmatched caretakers are assigned round-robin.
func distributeBetreuende(groups []models.Group, betreuende []models.Betreuende) {
	if len(groups) == 0 {
		return
	}
	// Reset betreuende lists
	for i := range groups {
		groups[i].Betreuende = nil
	}
	// Track how many betreuende each group already has (for round-robin tiebreak)
	assigned := make([]int, len(groups))

	for _, b := range betreuende {
		bestIdx := -1
		bestOVCount := 0
		bestAssigned := math.MaxInt64

		for i, g := range groups {
			ovCount := g.Ortsverbands[b.Ortsverband]
			if ovCount > bestOVCount || (ovCount == bestOVCount && assigned[i] < bestAssigned) {
				bestIdx = i
				bestOVCount = ovCount
				bestAssigned = assigned[i]
			}
		}
		if bestIdx < 0 {
			bestIdx = 0
		}
		groups[bestIdx].Betreuende = append(groups[bestIdx].Betreuende, b)
		assigned[bestIdx]++
	}
}
