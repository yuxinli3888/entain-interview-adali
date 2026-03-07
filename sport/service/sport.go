package service

import (
	"context"
	"errors"
	"log"
	"os"

	"git.neds.sh/matty/entain/sport/db"
	"git.neds.sh/matty/entain/sport/proto/sport"
)

var defaultLogger = log.New(os.Stdout, "[sport-service] ", log.LstdFlags)

type Sport interface {
	// ListEvents returns a list of sport events.
	ListEvents(ctx context.Context, in *sport.ListEventsRequest) (*sport.ListEventsResponse, error)
}

// sportService implements the Sport interface.
type sportService struct {
	eventsRepo db.EventsRepo
	logger     *log.Logger
}

// NewSportService instantiates and returns a new sportService.
func NewSportService(eventsRepo db.EventsRepo) Sport {
	return NewSportServiceWithLogger(eventsRepo, defaultLogger)
}

// NewSportServiceWithLogger instantiates a new sportService with a custom logger.
func NewSportServiceWithLogger(eventsRepo db.EventsRepo, logger *log.Logger) Sport {
	if logger == nil {
		logger = defaultLogger
	}

	return &sportService{
		eventsRepo: eventsRepo,
		logger:     logger,
	}
}

// ListEvents returns a list of sport events.
func (s *sportService) ListEvents(ctx context.Context, in *sport.ListEventsRequest) (*sport.ListEventsResponse, error) {
	if in == nil {
		err := errors.New("list events request is required")
		s.logger.Printf("ListEvents failed: %v", err)
		return nil, err
	}

	competitionIDCount := 0
	onlyVisible := false
	sortBy := sport.SortBy_ADVERTISED_START_TIME
	sortOrder := sport.Order_ASC
	if in.Filter != nil {
		competitionIDCount = len(in.Filter.CompetitionIds)
		onlyVisible = in.Filter.GetOnlyVisible()
		sortBy = in.Filter.GetSortBy()
		sortOrder = in.Filter.GetSortOrder()
	}
	s.logger.Printf("ListEvents called (competition_ids=%d only_visible=%t sort_by=%v sort_order=%v)", competitionIDCount, onlyVisible, sortBy, sortOrder)

	events, err := s.eventsRepo.List(in.Filter)
	if err != nil {
		s.logger.Printf("ListEvents failed while listing events: %v", err)
		return nil, err
	}

	s.logger.Printf("ListEvents succeeded (count=%d)", len(events))

	return &sport.ListEventsResponse{Events: events}, nil
}
