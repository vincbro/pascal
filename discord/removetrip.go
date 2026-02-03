package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/vincbro/pascal/state"
)

func CreateRemoveTripCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "remove",
			Description: "Remove a trip",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "name",
					Description:  "The trip you want to remove",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		Handler:      removeTripHandler,
		Autocomplete: listAutocomplete,
	}
}

func removeTripHandler(s *discordgo.Session, i *discordgo.InteractionCreate, state *state.State) error {
	user, err := GetUser(i.User, i.ChannelID, state)
	if err != nil {
		return err
	}
	opts := ParseOptions(i.ApplicationCommandData().Options)

	tripID := opts["name"].StringValue()

	trip, err := state.DB.GetTrip(user.ID, tripID)
	if err != nil {
		return err
	}

	// 4. Create the Embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Removed: %s", trip.Name),
		Description: "I've removed this trip from my database.",
		Color:       0x57F287,
		Fields:      []*discordgo.MessageEmbedField{},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Pascal • Watching your commute",
		},
	}

	if err := state.DB.RemoveTrip(user.ID, tripID); err != nil {
		return err
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	return err
}
