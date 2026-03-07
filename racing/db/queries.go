package db

const (
	racesList       = "list"
	racesListSingle = "listSingle"
)

func getRaceQueries() map[string]string {
	return map[string]string{
		racesList: `
			SELECT 
				id, 
				meeting_id, 
				name, 
				number, 
				visible, 
				advertised_start_time 
			FROM races
		`,
		racesListSingle: `
			SELECT 
				id, 
				meeting_id, 
				name, 
				number, 
				visible, 
				advertised_start_time 
			FROM races 
			WHERE id = ?
		`,
	}
}
