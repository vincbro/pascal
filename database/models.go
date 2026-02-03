package database

import "github.com/vincbro/pascal/blaise"

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

	ExpectedItinerary blaise.Itinerary `gorm:"serializer:json"`
}
