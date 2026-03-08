package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

var (
	raceOrderRulesMap = map[racing.Order]string{
		racing.Order_ASC:  "ASC",
		racing.Order_DESC: "DESC",
	}

	orderAttributeMap = map[racing.OrderAttribute]string{
		racing.OrderAttribute_ADVERTISED_START_TIME: "advertised_start_time",
		racing.OrderAttribute_NAME:                  "name",
		racing.OrderAttribute_NUMBER:                "number",
		racing.OrderAttribute_ID:                    "id",
		racing.OrderAttribute_MEETING_ID:            "meeting_id",
		racing.OrderAttribute_VISIBLE:               "visible",
	}
)

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// List will return a list of races.
	List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error)

	// ListSingleRaceByID will return a single race by its ID.
	GetRace(raceID int64) ([]*racing.Race, error)
}

type racesRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewRacesRepo creates a new races repository.
func NewRacesRepo(db *sql.DB) RacesRepo {
	return &racesRepo{db: db}
}

// Init prepares the race repository dummy data.
func (r *racesRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})

	return err
}

// List returns a list of races, filtered by the provided criteria.
func (r *racesRepo) List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query, args = r.applyFilter(query, filter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows)
}

// GetRace returns a single race by its ID.
func (r *racesRepo) GetRace(raceID int64) ([]*racing.Race, error) {
	var (
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesListSingle]

	args = append(args, raceID)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows)
}

// applyFilter applies the provided filter criteria to the SQL query and returns the modified query along with the corresponding arguments.
func (r *racesRepo) applyFilter(query string, filter *racing.ListRacesRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	orderAttribute := orderAttributeMap[racing.OrderAttribute_ADVERTISED_START_TIME]
	orderRules := raceOrderRulesMap[racing.Order_ASC]

	if filter != nil {
		if len(filter.MeetingIds) > 0 {
			clauses = append(clauses, "meeting_id IN ("+strings.Repeat("?,", len(filter.MeetingIds)-1)+"?)")

			for _, meetingID := range filter.MeetingIds {
				args = append(args, meetingID)
			}
		}

		if filter.GetOnlyVisible() {
			clauses = append(clauses, "visible == 1")
		}

		if candidate, ok := orderAttributeMap[filter.GetOrderAttribute()]; ok {
			orderAttribute = candidate
		}

		if candidate, ok := raceOrderRulesMap[filter.GetRaceOrder()]; ok {
			orderRules = candidate
		}
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY " + orderAttribute + " " + orderRules

	return query, args
}

func (m *racesRepo) scanRaces(
	rows *sql.Rows,
) ([]*racing.Race, error) {
	var races []*racing.Race

	for rows.Next() {
		var race racing.Race
		var advertisedStart time.Time

		if err := rows.Scan(&race.Id, &race.MeetingId, &race.Name, &race.Number, &race.Visible, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		race.AdvertisedStartTime = ts

		races = append(races, &race)
	}

	return races, nil
}
