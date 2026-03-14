package database

import (
	"database/sql"

	"experiment1/backend/models"
)

// GetGroupEvaluations retrieves all groups with their total scores, ranked from high to low
func GetGroupEvaluations(db *sql.DB) ([]models.GroupEvaluation, error) {
	// Query directly from group_station_scores and aggregate by group
	query := `
		SELECT 
			group_id,
			COALESCE(SUM(score), 0) as total_score,
			COUNT(score) as station_count
		FROM group_station_scores
		GROUP BY group_id
		ORDER BY total_score DESC, group_id ASC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evaluations []models.GroupEvaluation
	for rows.Next() {
		var eval models.GroupEvaluation
		var totalScore sql.NullInt64
		var stationCount sql.NullInt64

		err := rows.Scan(&eval.GroupID, &totalScore, &stationCount)
		if err != nil {
			return nil, err
		}

		if totalScore.Valid {
			eval.TotalScore = int(totalScore.Int64)
		}
		if stationCount.Valid {
			eval.StationCount = int(stationCount.Int64)
		}

		evaluations = append(evaluations, eval)
	}

	return evaluations, rows.Err()
}

// GetOrtsverbandEvaluations retrieves all ortsverbands with their average scores, ranked from high to low
// For each ortsverband, sums all participant scores (based on their group's total) and divides by participant count
func GetOrtsverbandEvaluations(db *sql.DB) ([]models.OrtsverbandEvaluation, error) {
	// Query: For each ortsverband:
	// 1. Get each participant and their group's total score
	// 2. Sum all participant scores
	// 3. Divide by the number of participants to get average
	query := `
				with base as (
					select    t.ortsverband
							, sum(gss.score) as total_score
							, count(distinct t.teilnehmer_id) as participant_count
					from teilnehmer t
					join gruppe r on t.teilnehmer_id = r.teilnehmer_id
					join group_station_scores gss on r.group_id = gss.group_id        
					group by t.ortsverband
				)
				select   ortsverband
				       , total_score
					   , participant_count
					   , case when participant_count > 0 
					          then total_score * 1.0 / participant_count 
							  else 0 
							  end as average_score
				  from base
				order by average_score desc, ortsverband asc;
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evaluations []models.OrtsverbandEvaluation
	for rows.Next() {
		var eval models.OrtsverbandEvaluation
		var totalScore sql.NullInt64
		var participantCount sql.NullInt64
		var averageScore sql.NullFloat64

		err := rows.Scan(&eval.Ortsverband, &totalScore, &participantCount, &averageScore)
		if err != nil {
			return nil, err
		}

		if totalScore.Valid {
			eval.TotalScore = int(totalScore.Int64)
		}
		if participantCount.Valid {
			eval.ParticipantCount = int(participantCount.Int64)
		}
		if averageScore.Valid {
			eval.AverageScore = averageScore.Float64
		}

		evaluations = append(evaluations, eval)
	}

	return evaluations, rows.Err()
}
