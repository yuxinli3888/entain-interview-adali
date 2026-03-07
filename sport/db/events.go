package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/sport/proto/sport"
)

var (
	eventOrderRulesMap = map[sport.Order]string{
		sport.Order_ASC:  "ASC",
		sport.Order_DESC: "DESC",
	}

	eventSortByAttributeMap = map[sport.SortBy]string{
		sport.SortBy_ADVERTISED_START_TIME: "advertised_start_time",
		sport.SortBy_NAME:                  "name",
		sport.SortBy_VISIBLE:               "visible",
		sport.SortBy_SPORT_CODE:            "sport_code",
		sport.SortBy_COMPETITION_ID:        "competition_id",
		sport.SortBy_ID:                    "id",
		sport.SortBy_HOME_TEAM:             "home_team",
		sport.SortBy_AWAY_TEAM:             "away_team",
		sport.SortBy_EVENT_STATUS:          "status",
	}
)

// EventsRepo provides repository access to events.
type EventsRepo interface {
	// Init will initialise our events repository.
	Init() error

	// List will return a list of events.
	List(filter *sport.ListEventsRequestFilter) ([]*sport.Event, error)
}

type eventsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewEventsRepo creates a new events repository.
func NewEventsRepo(db *sql.DB) EventsRepo {
	return &eventsRepo{db: db}
}

// Init prepares the event repository dummy data.
func (r *eventsRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy events.
		err = r.seed()
	})

	return err
}

func (r *eventsRepo) List(filter *sport.ListEventsRequestFilter) ([]*sport.Event, error) {
	var (
		query string
		args  []interface{}
	)

	query = getEventQueries()[eventsList]

	query, args = r.applyFilter(query, filter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanEvents(rows)
}

// applyFilter applies the provided filter to the query and returns the modified query and its arguments.
// It supports filtering by competition IDs and visibility, as well as sorting by various attributes and order.
// The function constructs the WHERE clause based on the filter criteria and appends the appropriate ORDER BY clause.
// Can be extended in the future to support additional filtering and sorting options as needed.
func (r *eventsRepo) applyFilter(query string, filter *sport.ListEventsRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	orderAttribute := eventSortByAttributeMap[sport.SortBy_ADVERTISED_START_TIME]
	orderRules := eventOrderRulesMap[sport.Order_ASC]

	if filter != nil {
		if len(filter.CompetitionIds) > 0 {
			clauses = append(clauses, "competition_id IN ("+strings.Repeat("?,", len(filter.CompetitionIds)-1)+"?)")

			for _, competitionID := range filter.CompetitionIds {
				args = append(args, competitionID)
			}
		}

		if filter.GetOnlyVisible() {
			clauses = append(clauses, "visible == 1")
		}

		if candidate, ok := eventSortByAttributeMap[filter.GetSortBy()]; ok {
			orderAttribute = candidate
		}

		if candidate, ok := eventOrderRulesMap[filter.GetSortOrder()]; ok {
			orderRules = candidate
		}
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY " + orderAttribute + " " + orderRules

	return query, args
}

func (r *eventsRepo) scanEvents(rows *sql.Rows) ([]*sport.Event, error) {
	defer rows.Close()

	var events []*sport.Event

	for rows.Next() {
		var event sport.Event
		var advertisedStart time.Time
		var status int64

		if err := rows.Scan(
			&event.Id,
			&event.CompetitionId,
			&event.Name,
			&advertisedStart,
			&event.Visible,
			&event.SportCode,
			&event.HomeTeam,
			&event.AwayTeam,
			&status,
		); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		event.AdvertisedStartTime = ts
		if _, ok := sport.EventStatus_name[int32(status)]; ok {
			event.Status = sport.EventStatus(status)
		}

		events = append(events, &event)
	}

	return events, rows.Err()
}
