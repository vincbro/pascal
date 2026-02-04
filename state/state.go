package state

import (
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

	handlers     []RequestHandler
	requests     chan Request
	kill         chan struct{}
	wg           sync.WaitGroup
	alertHistory map[string][]bool
	mutedTrips   map[string]bool
}

func NewState(db *database.Database, bClient *blaise.Client) *State {
	return &State{
		DB:      db,
		BClient: bClient,

		wg:           sync.WaitGroup{},
		handlers:     make([]RequestHandler, 0),
		requests:     make(chan Request, 128),
		kill:         make(chan struct{}),
		alertHistory: make(map[string][]bool),
		mutedTrips:   make(map[string]bool),
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
	s.mutedTrips[tripID] = true
}

func (s *State) UnMuteTrip(tripID string) {
	s.mutedTrips[tripID] = false
}

func (s *State) Start() {
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
					history, ok := s.alertHistory[trip.ID]
					if !ok {
						history = make([]bool, len(alerts))
					}

					departure := trip.ExpectedItinerary.DepartureTime
					if currentSeconds > departure {
						for i := range history {
							history[i] = false
						}
						s.alertHistory[trip.ID] = history
						s.UnMuteTrip(trip.ID)
						slog.Debug("Reset", "name", trip.Name, "history", history)
						continue
					}

					muted, ok := s.mutedTrips[trip.ID]
					if !ok {
						muted = false
					}
					if muted {
						slog.Debug("Was Muted", "name", trip.Name, "history", history)
						continue
					}

					diff := departure - currentSeconds
					shouldNotify := false
					minsLeft := blaise.Time(0)
					for i, alert := range alerts {
						if alert > diff && !history[i] {
							history[i] = true
							shouldNotify = true
							minsLeft = alert / 60
						}
					}
					if shouldNotify {
						slog.Debug("Updated", "name", trip.Name, "history", history)
						s.alertHistory[trip.ID] = history
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
}
func (s *State) Stop() {
	close(s.kill)
	s.wg.Wait()
	slog.Info("Stopped trip watcher")
}
