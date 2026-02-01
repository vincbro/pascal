package state

import (
	"github.com/vincbro/pascal/blaise"
	"github.com/vincbro/pascal/database"
)

type State struct {
	DB      *database.Database
	BClient *blaise.Client
}

func NewState(db *database.Database, bClient *blaise.Client) *State {
	return &State{
		DB:      db,
		BClient: bClient,
	}
}
