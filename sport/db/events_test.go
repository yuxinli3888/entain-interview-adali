package db

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"git.neds.sh/matty/entain/sport/proto/sport"
	_ "github.com/mattn/go-sqlite3"
)

func TestEventsApplyFilter(t *testing.T) {
	repo := &eventsRepo{}
	baseQuery := getEventQueries()[eventsList]

	tests := []struct {
		name              string
		filter            *sport.ListEventsRequestFilter
		expectWhere       bool
		expectedFragments []string
		expectedArgs      []interface{}
	}{
		{
			name:        "nil filter",
			filter:      nil,
			expectWhere: false,
			expectedFragments: []string{
				"ORDER BY advertised_start_time ASC",
			},
		},
		{
			name:              "competition ids only",
			filter:            &sport.ListEventsRequestFilter{CompetitionIds: []int64{1, 2, 3}},
			expectWhere:       true,
			expectedFragments: []string{"competition_id IN (?,?,?)", "ORDER BY advertised_start_time ASC"},
			expectedArgs:      []interface{}{int64(1), int64(2), int64(3)},
		},
		{
			name:              "only visible",
			filter:            &sport.ListEventsRequestFilter{OnlyVisible: true},
			expectWhere:       true,
			expectedFragments: []string{"visible == 1", "ORDER BY advertised_start_time ASC"},
		},
		{
			name: "custom order desc",
			filter: &sport.ListEventsRequestFilter{
				SortOrder: orderPtr(sport.Order_DESC),
			},
			expectWhere:       false,
			expectedFragments: []string{"ORDER BY advertised_start_time DESC"},
		},
		{
			name: "custom order attribute and direction",
			filter: &sport.ListEventsRequestFilter{
				SortBy:    sortByPtr(sport.SortBy_NAME),
				SortOrder: orderPtr(sport.Order_ASC),
			},
			expectWhere:       false,
			expectedFragments: []string{"ORDER BY name ASC"},
		},
		{
			name:              "competition ids and visible",
			filter:            &sport.ListEventsRequestFilter{CompetitionIds: []int64{7, 9}, OnlyVisible: true},
			expectWhere:       true,
			expectedFragments: []string{"competition_id IN (?,?)", "visible == 1", " AND ", "ORDER BY advertised_start_time ASC"},
			expectedArgs:      []interface{}{int64(7), int64(9)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args := repo.applyFilter(baseQuery, tt.filter)

			if tt.expectWhere != strings.Contains(query, " WHERE ") {
				t.Fatalf("unexpected WHERE clause presence, query=%q", query)
			}

			for _, fragment := range tt.expectedFragments {
				if !strings.Contains(query, fragment) {
					t.Fatalf("expected query fragment %q in query %q", fragment, query)
				}
			}

			if !reflect.DeepEqual(args, tt.expectedArgs) {
				t.Fatalf("unexpected args, got=%v want=%v", args, tt.expectedArgs)
			}
		})
	}
}

func TestEventsList(t *testing.T) {
	tests := []struct {
		name        string
		seedData    bool
		filter      *sport.ListEventsRequestFilter
		expectedIDs []int64
		expectErr   error
	}{
		{
			name:        "no filter returns all events",
			seedData:    true,
			filter:      nil,
			expectedIDs: []int64{2, 1, 3},
		},
		{
			name:        "competition id filter",
			seedData:    true,
			filter:      &sport.ListEventsRequestFilter{CompetitionIds: []int64{10}},
			expectedIDs: []int64{1},
		},
		{
			name:        "only visible filter",
			seedData:    true,
			filter:      &sport.ListEventsRequestFilter{OnlyVisible: true},
			expectedIDs: []int64{1, 3},
		},
		{
			name:        "combined filter",
			seedData:    true,
			filter:      &sport.ListEventsRequestFilter{CompetitionIds: []int64{11}, OnlyVisible: true},
			expectedIDs: []int64{},
		},
		{
			name:     "default order attribute with desc direction",
			seedData: true,
			filter: &sport.ListEventsRequestFilter{
				SortOrder: orderPtr(sport.Order_DESC),
			},
			expectedIDs: []int64{3, 1, 2},
		},
		{
			name:     "order by name asc",
			seedData: true,
			filter: &sport.ListEventsRequestFilter{
				SortBy: sortByPtr(sport.SortBy_NAME),
			},
			expectedIDs: []int64{2, 1, 3},
		},
		{
			name:     "order by id desc",
			seedData: true,
			filter: &sport.ListEventsRequestFilter{
				SortBy:    sortByPtr(sport.SortBy_ID),
				SortOrder: orderPtr(sport.Order_DESC),
			},
			expectedIDs: []int64{3, 2, 1},
		},
		{
			name:      "query error when table missing",
			seedData:  false,
			filter:    nil,
			expectErr: errors.New("no such table: events"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbConn := setupListTestDB(t, tt.seedData)
			t.Cleanup(func() { _ = dbConn.Close() })

			repo := &eventsRepo{db: dbConn}
			events, err := repo.List(tt.filter)

			if tt.expectErr != nil {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectErr.Error())
				}
				if !strings.Contains(err.Error(), tt.expectErr.Error()) {
					t.Fatalf("unexpected error, got=%q want substring=%q", err.Error(), tt.expectErr.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("List returned error: %v", err)
			}

			gotIDs := eventIDsInOrder(events)
			if !reflect.DeepEqual(gotIDs, tt.expectedIDs) {
				t.Fatalf("unexpected event ids, got=%v want=%v", gotIDs, tt.expectedIDs)
			}

			for _, event := range events {
				if event.AdvertisedStartTime == nil {
					t.Fatalf("expected advertised_start_time to be set for event id=%d", event.Id)
				}
			}
		})
	}
}

func setupListTestDB(t *testing.T, seed bool) *sql.DB {
	t.Helper()

	dbConn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite memory db: %v", err)
	}

	if !seed {
		return dbConn
	}

	_, err = dbConn.Exec(`CREATE TABLE events (
		id INTEGER PRIMARY KEY,
		competition_id INTEGER,
		name TEXT,
		advertised_start_time DATETIME,
		visible INTEGER,
		sport_code TEXT,
		home_team TEXT,
		away_team TEXT,
		status INTEGER
	)`)
	if err != nil {
		t.Fatalf("failed to create events table: %v", err)
	}

	start := time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)
	rows := []struct {
		id            int64
		competitionID int64
		name          string
		startTime     time.Time
		visible       bool
		sportCode     string
		homeTeam      string
		awayTeam      string
		status        sport.EventStatus
	}{
		{id: 1, competitionID: 10, name: "Beta v Gamma", startTime: start.Add(1 * time.Hour), visible: true, sportCode: "soccer", homeTeam: "Beta", awayTeam: "Gamma", status: sport.EventStatus_SCHEDULED},
		{id: 2, competitionID: 11, name: "Alpha v Delta", startTime: start, visible: false, sportCode: "tennis", homeTeam: "Alpha", awayTeam: "Delta", status: sport.EventStatus_LIVE},
		{id: 3, competitionID: 20, name: "Zeta v Epsilon", startTime: start.Add(2 * time.Hour), visible: true, sportCode: "soccer", homeTeam: "Zeta", awayTeam: "Epsilon", status: sport.EventStatus_FINISHED},
	}

	for _, row := range rows {
		_, err = dbConn.Exec(`INSERT INTO events(id, competition_id, name, advertised_start_time, visible, sport_code, home_team, away_team, status) VALUES (?,?,?,?,?,?,?,?,?)`,
			row.id,
			row.competitionID,
			row.name,
			row.startTime,
			row.visible,
			row.sportCode,
			row.homeTeam,
			row.awayTeam,
			row.status,
		)
		if err != nil {
			t.Fatalf("failed to insert row id=%d: %v", row.id, err)
		}
	}

	return dbConn
}

func eventIDsInOrder(events []*sport.Event) []int64 {
	ids := make([]int64, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.Id)
	}
	return ids
}

func orderPtr(order sport.Order) *sport.Order {
	return &order
}

func sortByPtr(attr sport.SortBy) *sport.SortBy {
	return &attr
}
