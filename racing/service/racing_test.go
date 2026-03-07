package service

import (
	"errors"
	"reflect"
	"testing"

	"git.neds.sh/matty/entain/racing/proto/racing"
	"golang.org/x/net/context"
)

// This file contains unit tests for the racing service. We use a stub repository to isolate the service logic from the database layer.
type stubRaces struct {
	listFunc func(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error)
}

// Init is a no-op for the stub repository.
func (s *stubRaces) Init() error {
	return nil
}

// List calls the configured listFunc
func (s *stubRaces) List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
	return s.listFunc(filter)
}

func TestRacingServiceListRaces(t *testing.T) {
	filter := &racing.ListRacesRequestFilter{
		MeetingIds: []int64{42},
		OnlyVisible: true,
		RaceOrder: orderPtr(racing.Order_DESC),
		OrderAttribute: orderAttributePtr(racing.OrderAttribute_NAME),
	}

	tests := []struct {
		name         string
		listResult   []*racing.Race
		listErr      error
		request      *racing.ListRacesRequest
		expectedResp *racing.ListRacesResponse
		expectErr    error
		expectListCall bool
	}{
		{
			name: "returns races from repository",
			listResult: []*racing.Race{
				{Id: 1, Name: "Alpha"},
				{Id: 2, Name: "Beta"},
			},
			request: &racing.ListRacesRequest{Filter: filter},
			expectedResp: &racing.ListRacesResponse{Races: []*racing.Race{
				{Id: 1, Name: "Alpha"},
				{Id: 2, Name: "Beta"},
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

func orderPtr(order racing.Order) *racing.Order {
	return &order
}

func orderAttributePtr(attr racing.OrderAttribute) *racing.OrderAttribute {
	return &attr
}
