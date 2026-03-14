package database

import (
	"database/sql"

	"experiment1/backend/models"
)

// GetAllTeilnehmers reads all participants from the database
func GetAllTeilnehmers(db *sql.DB) ([]models.Teilnehmer, error) {
	rows, err := db.Query("SELECT id, teilnehmer_id, name, ortsverband, age, geschlecht, pregroup FROM teilnehmer ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teilnehmers []models.Teilnehmer
	for rows.Next() {
		var t models.Teilnehmer
		var alter sql.NullInt64
		err := rows.Scan(&t.ID, &t.TeilnehmerID, &t.Name, &t.Ortsverband, &alter, &t.Geschlecht, &t.PreGroup)
		if err != nil {
			return nil, err
		}
		if alter.Valid {
			t.Alter = int(alter.Int64)
		}
		teilnehmers = append(teilnehmers, t)
	}

	return teilnehmers, rows.Err()
}

// GetGroupsForReport retrieves all groups with their participants from the database
func GetGroupsForReport(db *sql.DB) ([]models.Group, error) {
	// Single query to get all groups and participants with JOIN
	query := `
		SELECT r.group_id, t.id, t.teilnehmer_id, t.name, t.ortsverband, t.age, t.geschlecht, t.pregroup
		FROM gruppe r
		INNER JOIN teilnehmer t ON t.teilnehmer_id = r.teilnehmer_id
		ORDER BY r.group_id, t.name
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map to build groups efficiently
	groupMap := make(map[int]*models.Group)
	var groupOrder []int

	for rows.Next() {
		var groupID int
		var t models.Teilnehmer
		var alter sql.NullInt64

		err := rows.Scan(&groupID, &t.ID, &t.TeilnehmerID, &t.Name, &t.Ortsverband, &alter, &t.Geschlecht, &t.PreGroup)
		if err != nil {
			return nil, err
		}

		if alter.Valid {
			t.Alter = int(alter.Int64)
		}

		// Get or create group
		group, exists := groupMap[groupID]
		if !exists {
			group = &models.Group{
				GroupID:      groupID,
				Teilnehmers:  make([]models.Teilnehmer, 0),
				Ortsverbands: make(map[string]int),
				Geschlechts:  make(map[string]int),
			}
			groupMap[groupID] = group
			groupOrder = append(groupOrder, groupID)
		}

		// Add participant to group
		group.Teilnehmers = append(group.Teilnehmers, t)

		// Update group statistics
		group.Ortsverbands[t.Ortsverband]++
		group.Geschlechts[t.Geschlecht]++
		group.AlterSum += t.Alter
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Convert map to slice in correct order
	groups := make([]models.Group, 0, len(groupMap))
	for _, groupID := range groupOrder {
		groups = append(groups, *groupMap[groupID])
	}

	return groups, nil
}

// GetStationsForReport retrieves all stations with group scores from the database
func GetStationsForReport(db *sql.DB) ([]models.Station, error) {
	// Single query to get all stations with their scores using LEFT JOIN
	// (LEFT JOIN ensures we get stations even if they have no scores yet)
	query := `
		SELECT 
			s.station_id, 
			s.station_name,
			gss.group_id,
			gss.score
		FROM stations s
		LEFT JOIN group_station_scores gss ON s.station_id = gss.station_id
		ORDER BY s.station_name, gss.score DESC, gss.group_id ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map to build stations efficiently
	stationMap := make(map[int]*models.Station)
	var stationOrder []int

	for rows.Next() {
		var stationID int
		var stationName string
		var groupID sql.NullInt64
		var score sql.NullInt64

		err := rows.Scan(&stationID, &stationName, &groupID, &score)
		if err != nil {
			return nil, err
		}

		// Get or create station
		station, exists := stationMap[stationID]
		if !exists {
			station = &models.Station{
				StationID:   stationID,
				StationName: stationName,
				GroupScores: make([]models.GroupScore, 0),
			}
			stationMap[stationID] = station
			stationOrder = append(stationOrder, stationID)
		}

		// Add score if exists (groupID will be NULL if no scores for this station)
		if groupID.Valid {
			gs := models.GroupScore{
				GroupID: int(groupID.Int64),
			}
			if score.Valid {
				gs.Score = int(score.Int64)
			}
			station.GroupScores = append(station.GroupScores, gs)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Convert map to slice in correct order
	stations := make([]models.Station, 0, len(stationMap))
	for _, stationID := range stationOrder {
		stations = append(stations, *stationMap[stationID])
	}

	return stations, nil
}

// GetAllGroupIDs retrieves all group IDs from the database
func GetAllGroupIDs(db *sql.DB) ([]int, error) {
	rows, err := db.Query("SELECT DISTINCT group_id FROM gruppe ORDER BY group_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groupIDs []int
	for rows.Next() {
		var groupID int
		if err := rows.Scan(&groupID); err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, groupID)
	}

	return groupIDs, rows.Err()
}
