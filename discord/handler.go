package main

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/vincbro/pascal/state"
)

func InteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate, cmds Commands, state *state.State) {
	data := i.ApplicationCommandData()
	slog.Debug("Got interaction", "name", data.Name, "user", data.TargetID)
	if cmd, ok := cmds[data.Name]; ok {
		var err error
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			err = cmd.Handler(s, i, state)
		case discordgo.InteractionApplicationCommandAutocomplete:
			if cmd.Autocomplete != nil {
				err = cmd.Autocomplete(s, i, state)
			}
		}

		if err != nil {
			slog.Error("error failed to complete command", "command", data.Name, "error", err)
		}
	}
}
