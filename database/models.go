package database

import "github.com/vincbro/pascal/blaise"

type User struct {
	ID        string `grom:"primaryKey"`
	Username  string
	ChannelID string
	Trips     []Trip
}

type Trip struct {
	ID        string `grom:"primaryKey"`
	UserID    string `grom:"index"`
	Name      string
	From      string
	To        string
	Time      string
	Departure bool

	ExpectedDeparture blaise.Time
	ExpectedArrival   blaise.Time
}
