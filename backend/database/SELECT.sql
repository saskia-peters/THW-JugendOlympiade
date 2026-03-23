		SELECT 
			t.ortsverband,
			COALESCE(SUM(gss.total_group_score), 0) as total_score,
			COUNT(DISTINCT t.teilnehmer_id) as participant_count,
			CASE 
				WHEN COUNT(DISTINCT t.teilnehmer_id) > 0 
				THEN COALESCE(SUM(gss.total_group_score), 0) * 1.0 / COUNT(DISTINCT t.teilnehmer_id)
				ELSE 0 
			END as average_score
		FROM teilnehmende t
		LEFT JOIN gruppe r ON t.teilnehmer_id = r.teilnehmer_id
		LEFT JOIN (
			SELECT group_id, SUM(score) as total_group_score
			FROM group_station_scores
			GROUP BY group_id
		) gss ON r.group_id = gss.group_id
		GROUP BY t.ortsverband
		ORDER BY average_score DESC, t.ortsverband ASC;

with base as (
	select t.ortsverband
	, sum(gss.score) as total_score
	, count(t.teilnehmer_id) as participant_count
	from teilnehmende t
	join gruppe r on t.teilnehmer_id = r.teilnehmer_id
	join group_station_scores gss on r.group_id = gss.group_id        
	group by t.ortsverband
)
select ortsverband, total_score, participant_count,
case when participant_count > 0 then total_score * 1.0 / participant_count else 0 end as average_score
from base
order by average_score desc, ortsverband asc;