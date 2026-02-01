package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/vincbro/pascal/cmd/bot/commands"
	"github.com/vincbro/pascal/internal/blaise"
)

func InteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate, cmds commands.Commands, client blaise.Client) {
	data := i.ApplicationCommandData()
	cmd := cmds[data.Name]
	var err error
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		err = cmd.Handler(s, i, client)
	case discordgo.InteractionApplicationCommandAutocomplete:
		if cmd.Autocomplete != nil {
			err = cmd.Autocomplete(s, i, client)
		}
	}

	if err != nil {
		fmt.Printf("error failed to complete command %s: %s\n", data.Name, err)
	}
}
