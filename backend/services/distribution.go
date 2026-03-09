package services

import (
	"database/sql"
	"fmt"
	"math"
	"sort"

	"experiment1/backend/database"
	"experiment1/backend/models"
)

// CreateBalancedGroups creates groups with balanced distribution
func CreateBalancedGroups(db *sql.DB) error {
	// Read all participants from database
	teilnehmers, err := database.GetAllTeilnehmers(db)
	if err != nil {
		return fmt.Errorf("failed to read teilnehmers: %w", err)
	}

	if len(teilnehmers) == 0 {
		return nil // No participants to group
	}

	// Create balanced groups using the distribution algorithm
	groups := distributeIntoGroups(teilnehmers)

	// Save groups to database
	if err := database.SaveGroups(db, groups); err != nil {
		return fmt.Errorf("failed to save groups: %w", err)
	}

	fmt.Printf("Created %d groups with balanced distribution\n", len(groups))
	for i, group := range groups {
		fmt.Printf("  Group %d: %d participants\n", i+1, len(group.Teilnehmers))
	}

	return nil
}

// distributeIntoGroups distributes participants into balanced groups
func distributeIntoGroups(teilnehmers []models.Teilnehmer) []models.Group {
	if len(teilnehmers) == 0 {
		return nil
	}

	// Calculate number of groups needed
	numGroups := int(math.Ceil(float64(len(teilnehmers)) / float64(models.MaxGroupSize)))

	// Initialize groups
	groups := make([]models.Group, numGroups)
	for i := range groups {
		groups[i] = models.Group{
			GroupID:      i + 1,
			Teilnehmers:  make([]models.Teilnehmer, 0, models.MaxGroupSize),
			Ortsverbands: make(map[string]int),
			Geschlechts:  make(map[string]int),
		}
	}

	// Sort participants for better distribution
	// First by Ortsverband, then by Geschlecht, then by Alter
	sort.Slice(teilnehmers, func(i, j int) bool {
		if teilnehmers[i].Ortsverband != teilnehmers[j].Ortsverband {
			return teilnehmers[i].Ortsverband < teilnehmers[j].Ortsverband
		}
		if teilnehmers[i].Geschlecht != teilnehmers[j].Geschlecht {
			return teilnehmers[i].Geschlecht < teilnehmers[j].Geschlecht
		}
		return teilnehmers[i].Alter < teilnehmers[j].Alter
	})

	// Distribute participants using round-robin with diversity scoring
	for _, teilnehmer := range teilnehmers {
		bestGroupIdx := findBestGroup(groups, teilnehmer)
		addTeilnehmerToGroup(&groups[bestGroupIdx], teilnehmer)
	}

	return groups
}

// findBestGroup finds the best group for a participant based on diversity
func findBestGroup(groups []models.Group, teilnehmer models.Teilnehmer) int {
	bestIdx := 0
	bestScore := math.MaxFloat64

	for i, group := range groups {
		// Skip if group is full
		if len(group.Teilnehmers) >= models.MaxGroupSize {
			continue
		}

		// Calculate diversity score (lower is better)
		score := calculateDiversityScore(group, teilnehmer)

		// Prefer groups with fewer members
		sizeBonus := float64(len(group.Teilnehmers)) * 0.5

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
func calculateDiversityScore(group models.Group, teilnehmer models.Teilnehmer) float64 {
	if len(group.Teilnehmers) == 0 {
		return 0
	}

	score := 0.0

	// Penalize if Ortsverband is already common in the group
	ortsverbandCount := group.Ortsverbands[teilnehmer.Ortsverband]
	score += float64(ortsverbandCount) * 2.0

	// Penalize if Geschlecht is already common in the group
	geschlechtCount := group.Geschlechts[teilnehmer.Geschlecht]
	score += float64(geschlechtCount) * 1.5

	// Penalize if Alter is too similar to group average
	if len(group.Teilnehmers) > 0 && teilnehmer.Alter > 0 {
		avgAlter := float64(group.AlterSum) / float64(len(group.Teilnehmers))
		alterDiff := math.Abs(float64(teilnehmer.Alter) - avgAlter)
		if alterDiff < 2 {
			score += 1.0
		}
	}

	return score
}

// addTeilnehmerToGroup adds a participant to the group and updates statistics
func addTeilnehmerToGroup(g *models.Group, t models.Teilnehmer) {
	g.Teilnehmers = append(g.Teilnehmers, t)
	g.Ortsverbands[t.Ortsverband]++
	g.Geschlechts[t.Geschlecht]++
	g.AlterSum += t.Alter
}
