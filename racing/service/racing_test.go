package service

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"git.neds.sh/matty/entain/racing/proto/racing"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This file contains unit tests for the racing service. We use a stub repository to isolate the service logic from the database layer.
type stubRaces struct {
	listFunc               func(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error)
	listSingleRaceByIDFunc func(raceID int64) ([]*racing.Race, error)
}

// Init is a no-op for the stub repository.
func (s *stubRaces) Init() error {
	return nil
}

// List calls the configured listFunc
func (s *stubRaces) List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
	return s.listFunc(filter)
}

// ListSingleRaceByID calls the configured listSingleRaceByIDFunc
func (s *stubRaces) GetRace(raceID int64) ([]*racing.Race, error) {
	return s.listSingleRaceByIDFunc(raceID)
}

func TestRacingServiceListRaces(t *testing.T) {
	pastTime := time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)
	futureTime := time.Now().Add(1 * time.Hour)

	filter := &racing.ListRacesRequestFilter{
		MeetingIds:     []int64{42},
		OnlyVisible:    true,
		RaceOrder:      orderPtr(racing.Order_DESC),
		OrderAttribute: orderAttributePtr(racing.OrderAttribute_NAME),
	}

	tests := []struct {
		name           string
		listResult     []*racing.Race
		listErr        error
		request        *racing.ListRacesRequest
		expectedResp   *racing.ListRacesResponse
		expectErr      error
		expectListCall bool
	}{
		{
			name: "returns races from repository",
			listResult: []*racing.Race{
				{Id: 1, Name: "Alpha", AdvertisedStartTime: timestamppb.New(pastTime)},
				{Id: 2, Name: "Beta", AdvertisedStartTime: timestamppb.New(futureTime)},
				{Id: 3, Name: "No Time"},
			},
			request: &racing.ListRacesRequest{Filter: filter},
			expectedResp: &racing.ListRacesResponse{Races: []*racing.Race{
				{Id: 1, Name: "Alpha", AdvertisedStartTime: timestamppb.New(pastTime), Status: racing.RaceStatus_CLOSED},
				{Id: 2, Name: "Beta", AdvertisedStartTime: timestamppb.New(futureTime), Status: racing.RaceStatus_OPEN},
				{Id: 3, Name: "No Time", Status: racing.RaceStatus_CLOSED},
			}},
			expectListCall: true,
		},
		{
			name:           "propagates repository error",
			listErr:        errors.New("db unavailable"),
			request:        &racing.ListRacesRequest{Filter: filter},
			expectErr:      errors.New("db unavailable"),
			expectListCall: true,
		},
		{
			name:           "nil request returns validation error",
			request:        nil,
			expectErr:      errors.New("list races request is required"),
			expectListCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotFilter *racing.ListRacesRequestFilter
			listCalled := false
			stub := &stubRaces{listFunc: func(in *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
				listCalled = true
				gotFilter = in
				return tt.listResult, tt.listErr
			}}

			svc := NewRacingService(stub)
			resp, err := svc.ListRaces(context.Background(), tt.request)

			if tt.expectErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.expectErr.Error())
				}
				if err.Error() != tt.expectErr.Error() {
					t.Fatalf("unexpected error, got=%q want=%q", err.Error(), tt.expectErr.Error())
				}
				if resp != nil {
					t.Fatalf("expected nil response on error, got=%+v", resp)
				}
				if listCalled != tt.expectListCall {
					t.Fatalf("unexpected list invocation, got=%t want=%t", listCalled, tt.expectListCall)
				}
				return
			}

			if err != nil {
				t.Fatalf("ListRaces returned error: %v", err)
			}

			if !reflect.DeepEqual(resp, tt.expectedResp) {
				t.Fatalf("unexpected response, got=%+v want=%+v", resp, tt.expectedResp)
			}

			if gotFilter != tt.request.Filter {
				t.Fatalf("request filter not forwarded to repository")
			}

			if listCalled != tt.expectListCall {
				t.Fatalf("unexpected list invocation, got=%t want=%t", listCalled, tt.expectListCall)
			}
		})
	}
}

func TestRacingServiceGetRace(t *testing.T) {
	pastTime := time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)
	futureTime := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name              string
		request           *racing.GetRaceRequest
		getRaceResult     []*racing.Race
		getRaceErr        error
		expectedResp      *racing.ListRacesResponse
		expectErr         error
		expectGetRaceCall bool
		expectedRaceID    int64
	}{
		{
			name:    "returns single race from repository",
			request: &racing.GetRaceRequest{Id: 7},
			getRaceResult: []*racing.Race{
				{Id: 7, Name: "Past Race", AdvertisedStartTime: timestamppb.New(pastTime)},
				{Id: 8, Name: "Future Race", AdvertisedStartTime: timestamppb.New(futureTime)},
				{Id: 9, Name: "No Time"},
			},
			expectedResp: &racing.ListRacesResponse{
				Races: []*racing.Race{
					{Id: 7, Name: "Past Race", AdvertisedStartTime: timestamppb.New(pastTime), Status: racing.RaceStatus_CLOSED},
					{Id: 8, Name: "Future Race", AdvertisedStartTime: timestamppb.New(futureTime), Status: racing.RaceStatus_OPEN},
					{Id: 9, Name: "No Time", Status: racing.RaceStatus_CLOSED},
				},
			},
			expectGetRaceCall: true,
			expectedRaceID:    7,
		},
		{
			name:              "propagates repository error",
			request:           &racing.GetRaceRequest{Id: 99},
			getRaceErr:        errors.New("db failure"),
			expectErr:         errors.New("db failure"),
			expectGetRaceCall: true,
			expectedRaceID:    99,
		},
		{
			name:              "nil request returns validation error",
			request:           nil,
			expectErr:         errors.New("get race request is required"),
			expectGetRaceCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getRaceCalled := false
			var gotRaceID int64

			stub := &stubRaces{
				listFunc: func(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
					return nil, errors.New("unexpected List call in GetRace test")
				},
				listSingleRaceByIDFunc: func(raceID int64) ([]*racing.Race, error) {
					getRaceCalled = true
					gotRaceID = raceID
					return tt.getRaceResult, tt.getRaceErr
				},
			}

			svc := NewRacingService(stub)
			resp, err := svc.GetRace(context.Background(), tt.request)

			if tt.expectErr != nil {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tt.expectErr.Error())
				}
				if err.Error() != tt.expectErr.Error() {
					t.Fatalf("unexpected error, got=%q want=%q", err.Error(), tt.expectErr.Error())
				}
				if resp != nil {
					t.Fatalf("expected nil response on error, got=%+v", resp)
				}
				if getRaceCalled != tt.expectGetRaceCall {
					t.Fatalf("unexpected get-race invocation, got=%t want=%t", getRaceCalled, tt.expectGetRaceCall)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetRace returned error: %v", err)
			}

			if !reflect.DeepEqual(resp, tt.expectedResp) {
				t.Fatalf("unexpected response, got=%+v want=%+v", resp, tt.expectedResp)
			}

			if getRaceCalled != tt.expectGetRaceCall {
				t.Fatalf("unexpected get-race invocation, got=%t want=%t", getRaceCalled, tt.expectGetRaceCall)
			}

			if gotRaceID != tt.expectedRaceID {
				t.Fatalf("unexpected race ID passed to repository, got=%d want=%d", gotRaceID, tt.expectedRaceID)
			}
		})
	}
}

func orderPtr(order racing.Order) *racing.Order {
	return &order
}

func orderAttributePtr(attr racing.OrderAttribute) *racing.OrderAttribute {
	return &attr
}
