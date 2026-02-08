package main

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/vincbro/pascal/blaise"
	"github.com/vincbro/pascal/state"
	"github.com/vincbro/suddig"
)

func CreateEditTripCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "edit",
			Description: "Edit a trip",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "name",
					Description:  "The name of trip you want to edit",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
				{
					Name:        "new_name",
					Description: "The new name of trip",
					Type:        discordgo.ApplicationCommandOptionString,
				},
				{
					Name:         "from",
					Description:  "The departure point of you trip",
					Type:         discordgo.ApplicationCommandOptionString,
					Autocomplete: true,
				},
				{
					Name:         "to",
					Description:  "The destination point of you trip",
					Type:         discordgo.ApplicationCommandOptionString,
					Autocomplete: true,
				},
				{
					Name:        "type",
					Description: "Is this the time you want to leave or the time you want to arrive?",
					Type:        discordgo.ApplicationCommandOptionString,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Arrive By", Value: "arrive"},
						{Name: "Depart At", Value: "depart"},
					},
				},
				{
					Name:         "time",
					Description:  "The time you want to departe or arrive at",
					Type:         discordgo.ApplicationCommandOptionString,
					Autocomplete: true,
				},
				{
					Name:        "monday",
					Description: "Should the trip run on Monday",
					Type:        discordgo.ApplicationCommandOptionBoolean,
				},
				{
					Name:        "tuesday",
					Description: "Should the trip run on Tuesday",
					Type:        discordgo.ApplicationCommandOptionBoolean,
				},
				{
					Name:        "wednesday",
					Description: "Should the trip run on Wednesday",
					Type:        discordgo.ApplicationCommandOptionBoolean,
				},
				{
					Name:        "thursday",
					Description: "Should the trip run on Thursday",
					Type:        discordgo.ApplicationCommandOptionBoolean,
				},
				{
					Name:        "friday",
					Description: "Should the trip run on Friday",
					Type:        discordgo.ApplicationCommandOptionBoolean,
				},
				{
					Name:        "saturday",
					Description: "Should the trip run on Saturday",
					Type:        discordgo.ApplicationCommandOptionBoolean,
				},
				{
					Name:        "sunday",
					Description: "Should the trip run on Sunday",
					Type:        discordgo.ApplicationCommandOptionBoolean,
				},
			},
		},
		Handler:      editTripHandler,
		Autocomplete: editTripAutocomplete,
	}
}

func editTripHandler(s *discordgo.Session, i *discordgo.InteractionCreate, state *state.State) error {
	user, err := GetUser(i.User, i.ChannelID, state)
	if err != nil {
		return err
	}
	opts := ParseOptions(i.ApplicationCommandData().Options)

	hasDay := func(day string) (bool, bool) {
		val, ok := opts[day]
		if ok {
			return val.BoolValue(), true
		} else {
			return false, false
		}
	}

	hasString := func(key string) (string, bool) {
		val, ok := opts[key]
		if ok {
			return val.StringValue(), true
		} else {
			return "", false
		}
	}

	name := opts["name"].StringValue()
	trip, err := state.DB.GetTrip(user.ID, name)
	if err != nil {
		return err
	}
	if new_name, has := hasString("new_name"); has {
		trip.Name = new_name
	}
	if from, has := hasString("from"); has {
		trip.FromID = from
	}
	if to, has := hasString("to"); has {
		trip.ToID = to
	}
	if time, has := hasString("time"); has {
		trip.Time = time
	}
	if ttype, has := hasString("type"); has {
		trip.Departure = ttype == "depart"
	}

	if monday, has := hasDay("monday"); has {
		trip.Monday = monday
	}
	if tuesday, has := hasDay("tuesday"); has {
		trip.Tuesday = tuesday
	}
	if wednesday, has := hasDay("wednesday"); has {
		trip.Wednesday = wednesday
	}
	if thursday, has := hasDay("thursday"); has {
		trip.Thursday = thursday
	}
	if friday, has := hasDay("friday"); has {
		trip.Friday = friday
	}
	if saturday, has := hasDay("saturday"); has {
		trip.Saturday = saturday
	}
	if sunday, has := hasDay("sunday"); has {
		trip.Sunday = sunday
	}

	itenirary, err := state.BClient.Routing(context.Background(), trip.FromID, trip.ToID, trip.Time, trip.Departure)
	if err != nil {
		return err
	}

	trip.FromID = itenirary.From.ID
	trip.From = itenirary.From.Name
	trip.ToID = itenirary.To.ID
	trip.To = itenirary.To.Name
	trip.ExpectedItinerary = itenirary

	if err = state.DB.UpdateTrip(trip); err != nil {
		return err
	}

	user.AddHistory(itenirary.From)
	user.AddHistory(itenirary.To)
	if err = state.DB.UpdateUser(user); err != nil {
		return err
	}

	scheduleType := "Depart at"
	if !trip.Departure {
		scheduleType = "Arrive by"
	}

	// 4. Create the Embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("✅Updated: %s", trip.Name),
		Description: "I've updated this trip in my database. I'll alert you before you need to leave.",
		Color:       0x57F287,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Route",
				Value:  fmt.Sprintf("From: **%s**\nTo: **%s**", itenirary.From.Name, itenirary.To.Name),
				Inline: true,
			},
			{
				Name:   "Schedule",
				Value:  fmt.Sprintf("%s **%s**", scheduleType, trip.Time),
				Inline: true,
			},
			{
				Name: "Possible trips",
				Value: fmt.Sprintf("Found one departing **%s** and arriving **%s**\n(Travel time: %d min)",
					itenirary.DepartureTime.ToHMSString(),
					itenirary.ArrivalTime.ToHMSString(),
					(itenirary.ArrivalTime-itenirary.DepartureTime)/60,
				),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Pascal • Watching your commute",
		},
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	return err
}

func editTripAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, state *state.State) error {
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
			trips, err := state.DB.GetAllUsersTrips(user.ID)
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
				return scores[i] > scores[j]
			})
		case "from", "to":
			input := option.StringValue()
			var results []blaise.Location
			if len(input) == 0 {
				results = user.Locations
			} else {
				results, err = state.BClient.SearchAreas(context.Background(), option.StringValue(), 10)
				if err != nil {
					fmt.Println("error failed to search for area", err)
					return err
				}
			}
			for _, area := range results {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  area.Name,
					Value: area.ID,
				})
			}
		case "time":
			slog.Debug("Asking for time", "q", option.StringValue())
			for _, choice := range timeSuggestions(option.StringValue()) {
				slog.Debug("Got time suggestion", "time", choice)
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  choice.Format("15:04") + ":00",
					Value: choice.Format("15:04") + ":00",
				})
			}
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
