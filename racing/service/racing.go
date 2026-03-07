package service

import (
	"errors"
	"log"
	"os"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"golang.org/x/net/context"
)

var defaultLogger = log.New(os.Stdout, "[racing-service] ", log.LstdFlags)

type Racing interface {
	// ListRaces will return a collection of races.
	ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error)
}

// racingService implements the Racing interface.
type racingService struct {
	racesRepo db.RacesRepo
	logger    *log.Logger
}

// NewRacingService instantiates and returns a new racingService.
func NewRacingService(racesRepo db.RacesRepo) Racing {
	return NewRacingServiceWithLogger(racesRepo, defaultLogger)
}

// NewRacingServiceWithLogger instantiates a new racingService with a custom logger.
func NewRacingServiceWithLogger(racesRepo db.RacesRepo, logger *log.Logger) Racing {
	if logger == nil {
		logger = defaultLogger
	}

	return &racingService{
		racesRepo: racesRepo,
		logger:    logger,
	}
}

func (s *racingService) ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error) {
	if in == nil {
		err := errors.New("list races request is required")
		s.logger.Printf("ListRaces failed: %v", err)
		return nil, err
	}

	meetingIDCount := 0
	onlyVisible := false
	raceOrder := racing.Order_ASC
	orderAttribute := racing.OrderAttribute_ADVERTISED_START_TIME
	if in.Filter != nil {
		meetingIDCount = len(in.Filter.MeetingIds)
		onlyVisible = in.Filter.GetOnlyVisible()
		raceOrder = in.Filter.GetRaceOrder()
		orderAttribute = in.Filter.GetOrderAttribute()
	}
	s.logger.Printf("ListRaces called with filter: meeting_id_count=%d, only_visible=%t, race_order=%v, order by attribute %v", meetingIDCount, onlyVisible, raceOrder, orderAttribute)

	races, err := s.racesRepo.List(in.Filter)
	if err != nil {
		s.logger.Printf("ListRaces failed while listing races: %v", err)
		return nil, err
	}

	s.logger.Printf("ListRaces succeeded (count=%d)", len(races))

	return &racing.ListRacesResponse{Races: races}, nil
}
