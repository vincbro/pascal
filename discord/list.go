package main

import (
	"fmt"
	"log/slog"
	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/vincbro/pascal/blaise"
	"github.com/vincbro/pascal/state"
	"github.com/vincbro/suddig"
)

func CreateListCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "list",
			Description: "List out your trips",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "name",
					Description:  "Enter if you only care about a single trip",
					Type:         discordgo.ApplicationCommandOptionString,
					Autocomplete: true,
				},
			},
		},
		Handler:      listHandler,
		Autocomplete: listAutocomplete,
	}
}

func listHandler(s *discordgo.Session, i *discordgo.InteractionCreate, state *state.State) error {
	user, err := GetUser(i.User, i.ChannelID, state)
	if err != nil {
		return err
	}
	opts := ParseOptions(i.ApplicationCommandData().Options)

	var embed *discordgo.MessageEmbed
	if value, ok := opts["name"]; ok {
		id := value.StringValue()
		trip, err := state.DB.GetTrip(user.ID, id)
		if err != nil {
			return err
		}

		embed = &discordgo.MessageEmbed{
			Title: trip.Name,
			Description: fmt.Sprintf("From **%s** to **%s**\nDeparting **%s** and arriving **%s**\n(Travel time: %d min)",
				trip.From,
				trip.To,
				trip.ExpectedItinerary.DepartureTime.ToHMSString(),
				trip.ExpectedItinerary.ArrivalTime.ToHMSString(),
				(trip.ExpectedItinerary.ArrivalTime-trip.ExpectedItinerary.DepartureTime)/60,
			),
			Color:  0x57F287,
			Fields: blaise.IteniraryToEmbedFields(trip.ExpectedItinerary),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Pascal • Watching your commute",
			},
		}
	} else {
		trips, err := state.DB.GetAllTrips(user.ID)
		if err != nil {
			return err
		}

		fields := make([]*discordgo.MessageEmbedField, 0, len(trips))
		for _, trip := range trips {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name: trip.Name,
				Value: fmt.Sprintf("From **%s** to **%s**\nDeparting **%s** and arriving **%s**\n(Travel time: %d min)",
					trip.From,
					trip.To,
					trip.ExpectedItinerary.DepartureTime.ToHMSString(),
					trip.ExpectedItinerary.ArrivalTime.ToHMSString(),
					(trip.ExpectedItinerary.ArrivalTime-trip.ExpectedItinerary.DepartureTime)/60,
				),
			})
		}

		// 4. Create the Embed
		embed = &discordgo.MessageEmbed{
			Title:       "Your trips",
			Description: "Here is all the trips you have registered",
			Color:       0x57F287,
			Fields:      fields,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Pascal • Watching your commute",
			},
		}
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	return err
}

func listAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, state *state.State) error {
	user, err := GetUser(i.User, i.ChannelID, state)
	if err != nil {
		return err
	}

	data := i.ApplicationCommandData()
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, 20)

	for _, option := range data.Options {
		if !option.Focused {
			continue
		}
		switch option.Name {
		case "name":
			trips, err := state.DB.GetAllTrips(user.ID)
			if err != nil {
				slog.Error("error failed to get trips for user", "user", user.ID, "error", err)
				return err
			}
			haystack := make([]string, len(trips))
			for i, trip := range trips {
				haystack[i] = trip.Name
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  trip.Name,
					Value: trip.ID,
				})
			}
			scores := suddig.RankMatches(option.StringValue(), haystack)
			sort.Slice(choices, func(i, j int) bool {
				return scores[i] < scores[j]
			})
		}
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})

	return err
}
