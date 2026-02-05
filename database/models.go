package database

import (
	"fmt"
	"time"

	"github.com/vincbro/pascal/blaise"
)

type User struct {
	ID        string `grom:"primaryKey"`
	Username  string
	ChannelID string
	Trips     []Trip
	Locations []blaise.Location `gorm:"serializer:json"`
}

func (u *User) AddHistory(newLoc blaise.Location) {
	cleaned := make([]blaise.Location, 0, len(u.Locations))
	for _, loc := range u.Locations {
		if loc.ID != newLoc.ID {
			cleaned = append(cleaned, loc)
		}
	}
	u.Locations = cleaned
	u.Locations = append(u.Locations, newLoc)

	count := len(u.Locations)
	if count > 10 {
		u.Locations = u.Locations[count-10 : count]
	}
}

type Trip struct {
	ID        string `grom:"primaryKey"`
	UserID    string `grom:"index"`
	Name      string
	FromID    string
	From      string
	ToID      string
	To        string
	Time      string
	Departure bool

	Monday    bool
	Tuesday   bool
	Wednesday bool
	Thursday  bool
	Friday    bool
	Saturday  bool
	Sunday    bool

	ExpectedItinerary blaise.Itinerary `gorm:"serializer:json"`
}

func (t Trip) FormatSchedule() string {
	return fmt.Sprintf("**Monday**: %t\n**Tuesday**: %t\n**Wednesday**: %t\n**Thursday**: %t\n**Friday**: %t\n**Saturday**: %t\n**Sunday**: %t",
		t.Monday,
		t.Tuesday,
		t.Wednesday,
		t.Thursday,
		t.Friday,
		t.Saturday,
		t.Sunday,
	)
}

func (t Trip) ShouldRun(weekday time.Weekday) bool {
	switch weekday {
	case time.Sunday:
		return t.Sunday
	case time.Monday:
		return t.Monday
	case time.Tuesday:
		return t.Tuesday
	case time.Wednesday:
		return t.Wednesday
	case time.Thursday:
		return t.Thursday
	case time.Friday:
		return t.Friday
	case time.Saturday:
		return t.Saturday
	}
	return false
}
