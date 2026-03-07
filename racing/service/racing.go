package service

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"git.neds.sh/matty/entain/racing/db"
	"git.neds.sh/matty/entain/racing/proto/racing"
)

var defaultLogger = log.New(os.Stdout, "[racing-service] ", log.LstdFlags)

type Racing interface {
	// ListRaces will return a collection of races.
	ListRaces(ctx context.Context, in *racing.ListRacesRequest) (*racing.ListRacesResponse, error)
	// ListSingleRaceByID will return a single race by its ID.
	GetRace(ctx context.Context, in *racing.GetRaceRequest) (*racing.ListRacesResponse, error)
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

	races = s.deriveRaceStatus(races)
	s.logger.Printf("ListRaces succeeded (count=%d)", len(races))

	return &racing.ListRacesResponse{Races: races}, nil
}

func (s *racingService) GetRace(ctx context.Context, in *racing.GetRaceRequest) (*racing.ListRacesResponse, error) {
	if in == nil {
		err := errors.New("get race request is required")
		s.logger.Printf("GetRace failed: %v", err)
		return nil, err
	}

	s.logger.Printf("GetRace called with race ID: %d", in.Id)
	races, err := s.racesRepo.GetRace(in.Id)
	if err != nil {
		s.logger.Printf("GetRace failed while listing single race: %v", err)
		return nil, err
	}

	races = s.deriveRaceStatus(races)
	s.logger.Printf("GetRace succeeded (count=%d)", len(races))

	return &racing.ListRacesResponse{Races: races}, nil
}

func (s *racingService) deriveRaceStatus(races []*racing.Race) []*racing.Race {
	logName := "deriveRaceStatus"
	now := time.Now()
	s.logger.Printf("Start Processing %s for %d races", logName, len(races))

	for _, race := range races {
		if race == nil {
			continue
		}
		if race.AdvertisedStartTime != nil {
			switch {
			case race.AdvertisedStartTime.AsTime().Before(now):
				race.Status = racing.RaceStatus_CLOSED
			default:
				race.Status = racing.RaceStatus_OPEN
			}
		} else {
			race.Status = racing.RaceStatus_CLOSED
		}
	}
	return races
}
