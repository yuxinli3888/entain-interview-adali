package db

const (
	eventsList = "list"
)

func getEventQueries() map[string]string {
	return map[string]string{
		eventsList: `
			SELECT
				id,
				competition_id,
				name,
				advertised_start_time,
				visible,
				sport_code,
				home_team,
				away_team,
				status
			FROM events
		`,
	}
}
