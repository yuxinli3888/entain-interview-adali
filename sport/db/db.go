package db

import (
	"time"

	"syreclabs.com/go/faker"
)

var sportCodes = []string{"soccer", "tennis", "basketball", "cricket"}

func (r *eventsRepo) seed() error {
	statement, err := r.db.Prepare(`CREATE TABLE IF NOT EXISTS events (id INTEGER PRIMARY KEY, competition_id INTEGER, name TEXT, advertised_start_time DATETIME, visible INTEGER, sport_code TEXT, home_team TEXT, away_team TEXT, status INTEGER)`)
	if err == nil {
		_, err = statement.Exec()
	}

	for i := 1; i <= 100; i++ {
		statement, err = r.db.Prepare(`INSERT OR IGNORE INTO events(id, competition_id, name, advertised_start_time, visible, sport_code, home_team, away_team, status) VALUES (?,?,?,?,?,?,?,?,?)`)
		if err == nil {
			homeTeam := faker.Name().FirstName()
			awayTeam := faker.Name().FirstName()
			_, err = statement.Exec(
				i,
				faker.Number().Between(1, 20),
				homeTeam+" v "+awayTeam,
				faker.Time().Between(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 2)).Format(time.RFC3339),
				faker.Number().Between(0, 1),
				sportCodes[i%len(sportCodes)],
				homeTeam,
				awayTeam,
				faker.Number().Between(0, 3),
			)
		}
	}

	return err
}
