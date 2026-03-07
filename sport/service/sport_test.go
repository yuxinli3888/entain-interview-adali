package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"git.neds.sh/matty/entain/sport/proto/sport"
)

type stubEventsRepo struct {
	listFunc func(filter *sport.ListEventsRequestFilter) ([]*sport.Event, error)
}

func (s *stubEventsRepo) Init() error {
	return nil
}

func (s *stubEventsRepo) List(filter *sport.ListEventsRequestFilter) ([]*sport.Event, error) {
	return s.listFunc(filter)
}

func TestSportServiceListEvents(t *testing.T) {
	filter := &sport.ListEventsRequestFilter{CompetitionIds: []int64{10}, OnlyVisible: true}

	tests := []struct {
		name           string
		request        *sport.ListEventsRequest
		listResult     []*sport.Event
		listErr        error
		expectedResp   *sport.ListEventsResponse
		expectErr      error
		expectListCall bool
	}{
		{
			name:    "returns events from repository",
			request: &sport.ListEventsRequest{Filter: filter},
			listResult: []*sport.Event{
				{Id: 1, Name: "A"},
				{Id: 2, Name: "B"},
			},
			expectedResp: &sport.ListEventsResponse{Events: []*sport.Event{
				{Id: 1, Name: "A"},
				{Id: 2, Name: "B"},
			}},
			expectListCall: true,
		},
		{
			name:           "propagates repository error",
			request:        &sport.ListEventsRequest{Filter: filter},
			listErr:        errors.New("db unavailable"),
			expectErr:      errors.New("db unavailable"),
			expectListCall: true,
		},
		{
			name:           "nil request returns validation error",
			request:        nil,
			expectErr:      errors.New("list events request is required"),
			expectListCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listCalled := false
			var gotFilter *sport.ListEventsRequestFilter

			repo := &stubEventsRepo{listFunc: func(in *sport.ListEventsRequestFilter) ([]*sport.Event, error) {
				listCalled = true
				gotFilter = in
				return tt.listResult, tt.listErr
			}}

			svc := NewSportService(repo)
			resp, err := svc.ListEvents(context.Background(), tt.request)

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
				t.Fatalf("ListEvents returned error: %v", err)
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
