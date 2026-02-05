package state

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/vincbro/pascal/blaise"
	"github.com/vincbro/pascal/database"
)

var alerts = []blaise.Time{60 * 60, 30 * 60, 15 * 60, 10 * 60, 5 * 60, 4 * 60, 3 * 60, 2 * 60, 1 * 60}

type State struct {
	DB      *database.Database
	BClient *blaise.Client

	gtfsUrl string

	handlers  []RequestHandler
	requests  chan Request
	kill      chan struct{}
	wg        sync.WaitGroup
	tripsMeta map[string]TripMeta
}

type TripMeta struct {
	AlertHistory []bool
	Muted        bool
	Edited       bool
}

func NewState(db *database.Database, bClient *blaise.Client, gtfsUrl string) *State {
	return &State{
		DB:      db,
		BClient: bClient,

		gtfsUrl: gtfsUrl,

		wg:        sync.WaitGroup{},
		handlers:  make([]RequestHandler, 0),
		requests:  make(chan Request, 128),
		kill:      make(chan struct{}),
		tripsMeta: make(map[string]TripMeta),
	}
}

type Request struct {
	UserID  string
	TripID  string
	Message string
}

type RequestHandler = func(s *State, request Request) error

func (s *State) AddHandler(handler RequestHandler) {
	s.handlers = append(s.handlers, handler)
}

func (s *State) SendRequest(request Request) {
	select {
	case s.requests <- request:
	default:
		slog.Warn("Request channel full, dropping request", "user", request.UserID)
	}
}

func (s *State) MuteTrip(tripID string) {
	meta, ok := s.tripsMeta[tripID]
	if !ok {
		meta = TripMeta{
			AlertHistory: make([]bool, len(alerts)),
			Muted:        true,
			Edited:       true,
		}
	} else {
		meta.Muted = true
		meta.Edited = true
	}
	s.tripsMeta[tripID] = meta
}

func (s *State) UnMuteTrip(tripID string) {
	meta, ok := s.tripsMeta[tripID]
	if !ok {
		meta = TripMeta{
			AlertHistory: make([]bool, len(alerts)),
			Muted:        false,
			Edited:       true,
		}
	} else {
		meta.Muted = false
		meta.Edited = true
	}
	s.tripsMeta[tripID] = meta
}

func (s *State) UpdateAllTrips() error {
	wg := sync.WaitGroup{}
	trips, err := s.DB.GetAllTrips()
	if err != nil {
		return err
	}

	for _, t := range trips {
		wg.Add(1)
		go func(trip *database.Trip) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			itinerary, err := s.BClient.Routing(ctx, trip.FromID, trip.ToID, trip.Time, trip.Departure)
			cancel()
			if err != nil {
				slog.Error("error while getting trip", "name", trip.Name, "error", err)
				return
			}
			trip.ExpectedItinerary = itinerary
			if err := s.DB.UpdateTrip(trip); err != nil {
				slog.Error("error while updating trip", "name", trip.Name, "error", err)
				return
			}
			slog.Info("Updated trip", "name", trip.Name)
		}(t)
	}
	wg.Wait()
	return nil
}

func (s *State) Start() {
	// Dispatch requests
	s.wg.Go(func() {
		for {
			select {
			case req := <-s.requests:
				for _, h := range s.handlers {
					go func(handler RequestHandler) {
						handler(s, req)
					}(h)
				}
			case <-s.kill:
				return
			}
		}
	})

	// Send notifications
	s.wg.Go(func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				now := time.Now()
				midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				currentSeconds := blaise.Time(now.Sub(midnight).Seconds())

				trips, err := s.DB.GetAllTrips()
				if err != nil {
					slog.Error("failed to fetch trips", "error", err)
					break
				}
				for _, trip := range trips {
					if !trip.ShouldRun(now.Weekday()) {
						continue
					}
					meta, ok := s.tripsMeta[trip.ID]
					// Create a new if we don't have one
					if !ok {
						meta = TripMeta{
							AlertHistory: make([]bool, len(alerts)),
							Muted:        false,
							Edited:       false,
						}
					}

					departure := trip.ExpectedItinerary.DepartureTime
					if currentSeconds > departure {
						if meta.Edited {
							for i := range meta.AlertHistory {
								meta.AlertHistory[i] = false
							}
							meta.Muted = false
							meta.Edited = false
							slog.Debug("Reset", "name", trip.Name, "meta", meta)
						}
						s.tripsMeta[trip.ID] = meta
						continue
					}

					if meta.Muted {
						slog.Debug("Was Muted", "name", trip.Name, "meta", meta)
						continue
					}

					diff := departure - currentSeconds
					shouldNotify := false
					minsLeft := blaise.Time(0)
					for i, alert := range alerts {
						if alert > diff && !meta.AlertHistory[i] {
							meta.AlertHistory[i] = true
							meta.Edited = true
							shouldNotify = true
							minsLeft = alert / 60
						}
					}
					if shouldNotify {
						slog.Debug("Updated", "name", trip.Name, "meta", meta)
						s.tripsMeta[trip.ID] = meta
						s.SendRequest(Request{
							UserID:  trip.UserID,
							TripID:  trip.ID,
							Message: fmt.Sprintf("🔔 **Depart Soon:** **%s** leaves in **%d** min!", trip.Name, minsLeft)})
					}
				}
			case <-s.kill:
				return
			}
		}
	})

	// Update data
	s.wg.Go(func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				if now.Hour() >= 6 && now.Hour() < 8 {
					ageSeconds, err := s.BClient.GetAge(context.Background())
					if err != nil {
						slog.Error("Data Manager: Failed to get data age", "error", err)
						continue
					}
					ageHours := ageSeconds / 3600
					slog.Debug("Data Manager check", "age_hours", ageHours, "current_hour", now.Hour())

					if ageHours > 23 {
						slog.Info("GTFS data is stale, triggering refresh", "age_hours", ageHours)
						err := s.BClient.TriggerRefresh(context.Background(), s.gtfsUrl)
						if err != nil {
							slog.Error("error while trying to trigger refresh", "error", err)
						} else {
							slog.Info("GTFS refreshed successfully")
							s.UpdateAllTrips()
						}
					}
				}
			case <-s.kill:
				return
			}
		}
	})
}
func (s *State) Stop() {
	close(s.kill)
	s.wg.Wait()
	slog.Info("Stopped trip watcher")
}
