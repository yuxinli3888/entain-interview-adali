package db

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"git.neds.sh/matty/entain/racing/proto/racing"
	_ "github.com/mattn/go-sqlite3"
)

func TestRacesApplyFilter(t *testing.T) {
	repo := &racesRepo{}
	baseQuery := getRaceQueries()[racesList]

	tests := []struct {
		name              string
		filter            *racing.ListRacesRequestFilter
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
			name:              "meeting ids only",
			filter:            &racing.ListRacesRequestFilter{MeetingIds: []int64{1, 2, 3}},
			expectWhere:       true,
			expectedFragments: []string{"meeting_id IN (?,?,?)", "ORDER BY advertised_start_time ASC"},
			expectedArgs:      []interface{}{int64(1), int64(2), int64(3)},
		},
		{
			name:              "only visible",
			filter:            &racing.ListRacesRequestFilter{OnlyVisible: true},
			expectWhere:       true,
			expectedFragments: []string{"visible == 1", "ORDER BY advertised_start_time ASC"},
		},
		{
			name: "custom race order desc",
			filter: &racing.ListRacesRequestFilter{
				RaceOrder: orderPtr(racing.Order_DESC),
			},
			expectWhere:       false,
			expectedFragments: []string{"ORDER BY advertised_start_time DESC"},
		},
		{
			name: "custom order attribute and direction",
			filter: &racing.ListRacesRequestFilter{
				OrderAttribute: orderAttributePtr(racing.OrderAttribute_NAME),
				RaceOrder:      orderPtr(racing.Order_ASC),
			},
			expectWhere:       false,
			expectedFragments: []string{"ORDER BY name ASC"},
		},
		{
			name:              "meeting ids and visible",
			filter:            &racing.ListRacesRequestFilter{MeetingIds: []int64{7, 9}, OnlyVisible: true},
			expectWhere:       true,
			expectedFragments: []string{"meeting_id IN (?,?)", "visible == 1", " AND ", "ORDER BY advertised_start_time ASC"},
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

func TestRacesList(t *testing.T) {
	tests := []struct {
		name        string
		seedData    bool
		filter      *racing.ListRacesRequestFilter
		expectedIDs []int64
		expectErr   error
	}{
		{
			name:        "no filter returns all races",
			seedData:    true,
			filter:      nil,
			expectedIDs: []int64{3, 2, 1},
		},
		{
			name:        "meeting id filter",
			seedData:    true,
			filter:      &racing.ListRacesRequestFilter{MeetingIds: []int64{11}},
			expectedIDs: []int64{3, 1},
		},
		{
			name:        "only visible filter",
			seedData:    true,
			filter:      &racing.ListRacesRequestFilter{OnlyVisible: true},
			expectedIDs: []int64{3, 1},
		},
		{
			name:        "combined filter",
			seedData:    true,
			filter:      &racing.ListRacesRequestFilter{MeetingIds: []int64{12}, OnlyVisible: true},
			expectedIDs: []int64{},
		},
		{
			name:     "default order attribute with desc direction",
			seedData: true,
			filter: &racing.ListRacesRequestFilter{
				RaceOrder: orderPtr(racing.Order_DESC),
			},
			expectedIDs: []int64{1, 2, 3},
		},
		{
			name:     "order by name asc",
			seedData: true,
			filter: &racing.ListRacesRequestFilter{
				OrderAttribute: orderAttributePtr(racing.OrderAttribute_NAME),
			},
			expectedIDs: []int64{2, 1, 3},
		},
		{
			name:     "order by id desc",
			seedData: true,
			filter: &racing.ListRacesRequestFilter{
				OrderAttribute: orderAttributePtr(racing.OrderAttribute_ID),
				RaceOrder:      orderPtr(racing.Order_DESC),
			},
			expectedIDs: []int64{3, 2, 1},
		},
		{
			name:      "query error when table missing",
			seedData:  false,
			filter:    nil,
			expectErr: errors.New("no such table: races"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbConn := setupListTestDB(t, tt.seedData)
			t.Cleanup(func() { _ = dbConn.Close() })

			repo := &racesRepo{db: dbConn}
			races, err := repo.List(tt.filter)

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

			gotIDs := raceIDsInOrder(races)
			if !reflect.DeepEqual(gotIDs, tt.expectedIDs) {
				t.Fatalf("unexpected race ids, got=%v want=%v", gotIDs, tt.expectedIDs)
			}

			for _, race := range races {
				if race.AdvertisedStartTime == nil {
					t.Fatalf("expected advertised_start_time to be set for race id=%d", race.Id)
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

	_, err = dbConn.Exec(`CREATE TABLE races (
		id INTEGER PRIMARY KEY,
		meeting_id INTEGER,
		name TEXT,
		number INTEGER,
		visible INTEGER,
		advertised_start_time DATETIME
	)`)
	if err != nil {
		t.Fatalf("failed to create races table: %v", err)
	}

	start := time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)
	rows := []struct {
		id        int64
		meetingID int64
		name      string
		number    int64
		visible   bool
		startTime time.Time
	}{
		{id: 1, meetingID: 11, name: "Bravo", number: 2, visible: true, startTime: start.Add(2 * time.Hour)},
		{id: 2, meetingID: 12, name: "Alpha", number: 3, visible: false, startTime: start.Add(time.Hour)},
		{id: 3, meetingID: 11, name: "Charlie", number: 1, visible: true, startTime: start},
	}

	for _, row := range rows {
		_, err = dbConn.Exec(
			`INSERT INTO races(id, meeting_id, name, number, visible, advertised_start_time) VALUES (?,?,?,?,?,?)`,
			row.id,
			row.meetingID,
			row.name,
			row.number,
			row.visible,
			row.startTime,
		)
		if err != nil {
			t.Fatalf("failed to seed races row %+v: %v", row, err)
		}
	}

	return dbConn
}

func raceIDs(races []*racing.Race) []int64 {
	ids := make([]int64, 0, len(races))
	for _, race := range races {
		ids = append(ids, race.Id)
	}
	return ids
}

func raceIDsInOrder(races []*racing.Race) []int64 {
	return raceIDs(races)
}

func orderPtr(order racing.Order) *racing.Order {
	return &order
}

func orderAttributePtr(attr racing.OrderAttribute) *racing.OrderAttribute {
	return &attr
}
