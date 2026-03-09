package database

import (
	"database/sql"

	"experiment1/backend/models"
)

// GetAllTeilnehmers reads all participants from the database
func GetAllTeilnehmers(db *sql.DB) ([]models.Teilnehmer, error) {
	rows, err := db.Query("SELECT id, teilnehmer_id, name, ortsverband, age, geschlecht FROM " + models.TableName + " ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teilnehmers []models.Teilnehmer
	for rows.Next() {
		var t models.Teilnehmer
		var alter sql.NullInt64
		err := rows.Scan(&t.ID, &t.TeilnehmerID, &t.Name, &t.Ortsverband, &alter, &t.Geschlecht)
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
	// Get all group IDs
	groupRows, err := db.Query("SELECT DISTINCT group_id FROM gruppe ORDER BY group_id")
	if err != nil {
		return nil, err
	}
	defer groupRows.Close()

	var groupIDs []int
	for groupRows.Next() {
		var groupID int
		if err := groupRows.Scan(&groupID); err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, groupID)
	}

	// For each group, get all participants
	var groups []models.Group
	for _, groupID := range groupIDs {
		query := `
			SELECT t.id, t.teilnehmer_id, t.name, t.ortsverband, t.age, t.geschlecht
			FROM teilnehmer t
			INNER JOIN rel_tn_grp r ON t.teilnehmer_id = r.teilnehmer_id
			WHERE r.group_id = ?
			ORDER BY t.name
		`

		rows, err := db.Query(query, groupID)
		if err != nil {
			return nil, err
		}

		group := models.Group{
			GroupID:      groupID,
			Teilnehmers:  make([]models.Teilnehmer, 0),
			Ortsverbands: make(map[string]int),
			Geschlechts:  make(map[string]int),
		}

		for rows.Next() {
			var t models.Teilnehmer
			var alter sql.NullInt64
			err := rows.Scan(&t.ID, &t.TeilnehmerID, &t.Name, &t.Ortsverband, &alter, &t.Geschlecht)
			if err != nil {
				rows.Close()
				return nil, err
			}
			if alter.Valid {
				t.Alter = int(alter.Int64)
			}
			group.Teilnehmers = append(group.Teilnehmers, t)

			// Update group statistics
			group.Ortsverbands[t.Ortsverband]++
			group.Geschlechts[t.Geschlecht]++
			group.AlterSum += t.Alter
		}
		rows.Close()

		groups = append(groups, group)
	}

	return groups, nil
}

// GetStationsForReport retrieves all stations with group scores from the database
func GetStationsForReport(db *sql.DB) ([]models.Station, error) {
	// Get all stations
	stationRows, err := db.Query("SELECT station_id, station_name FROM stations ORDER BY station_name")
	if err != nil {
		return nil, err
	}
	defer stationRows.Close()

	var stations []models.Station
	for stationRows.Next() {
		var station models.Station
		if err := stationRows.Scan(&station.StationID, &station.StationName); err != nil {
			return nil, err
		}

		// Get group scores for this station, ordered by score descending
		scoreQuery := `
			SELECT group_id, score
			FROM group_station_scores
			WHERE station_id = ?
			ORDER BY score DESC, group_id ASC
		`

		scoreRows, err := db.Query(scoreQuery, station.StationID)
		if err != nil {
			return nil, err
		}

		station.GroupScores = make([]models.GroupScore, 0)
		for scoreRows.Next() {
			var gs models.GroupScore
			var score sql.NullInt64
			if err := scoreRows.Scan(&gs.GroupID, &score); err != nil {
				scoreRows.Close()
				return nil, err
			}
			if score.Valid {
				gs.Score = int(score.Int64)
			}
			station.GroupScores = append(station.GroupScores, gs)
		}
		scoreRows.Close()

		stations = append(stations, station)
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
