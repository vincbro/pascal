package main

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"

	"github.com/vincbro/pascal/blaise"
	"github.com/vincbro/pascal/database"
	"github.com/vincbro/pascal/state"
)

func CreateAddTripCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "add",
			Description: "Create a new trip",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "name",
					Description: "The name of your trip",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
				{
					Name:         "from",
					Description:  "The departure point of you trip",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
				{
					Name:         "to",
					Description:  "The destination point of you trip",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
				},
				{
					Name:        "type",
					Description: "Is this the time you want to leave or the time you want to arrive?",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Arrive By", Value: "arrive"},
						{Name: "Depart At", Value: "depart"},
					},
				},
				{
					Name:         "time",
					Description:  "The time you want to departe or arrive at",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
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
		Handler:      addTripHandler,
		Autocomplete: addTripAutocomplete,
	}
}

func addTripHandler(s *discordgo.Session, i *discordgo.InteractionCreate, state *state.State) error {
	user, err := GetUser(i.User, i.ChannelID, state)
	if err != nil {
		return err
	}
	opts := ParseOptions(i.ApplicationCommandData().Options)

	hasDay := func(day string) bool {
		val, ok := opts[day]
		if ok {
			return val.BoolValue()
		} else {
			return true
		}
	}

	name := opts["name"].StringValue()
	from := opts["from"].StringValue()
	to := opts["to"].StringValue()
	time := opts["time"].StringValue()
	departure := opts["type"].StringValue() == "depart"

	monday := hasDay("monday")
	tuesday := hasDay("tuesday")
	wednesday := hasDay("wednesday")
	thursday := hasDay("thursday")
	friday := hasDay("friday")
	saturday := hasDay("saturday")
	sunday := hasDay("sunday")

	itenirary, err := state.BClient.Routing(context.Background(), from, to, time, departure)
	if err != nil {
		return err
	}

	trip := database.Trip{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Name:      name,
		From:      itenirary.From.Name,
		FromID:    itenirary.From.ID,
		To:        itenirary.To.Name,
		ToID:      itenirary.To.ID,
		Time:      time,
		Departure: departure,

		Monday:    monday,
		Tuesday:   tuesday,
		Wednesday: wednesday,
		Thursday:  thursday,
		Friday:    friday,
		Saturday:  saturday,
		Sunday:    sunday,

		ExpectedItinerary: itenirary,
	}

	if err = state.DB.AddTrip(&trip); err != nil {
		return err
	}

	user.AddHistory(itenirary.From)
	user.AddHistory(itenirary.To)
	if err = state.DB.UpdateUser(user); err != nil {
		return err
	}

	scheduleType := "Depart at"
	if !departure {
		scheduleType = "Arrive by"
	}

	// 4. Create the Embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("✅ Saved: %s", name),
		Description: "I've added this trip to my database. I'll alert you before you need to leave.",
		Color:       0x57F287,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Route",
				Value:  fmt.Sprintf("From: **%s**\nTo: **%s**", itenirary.From.Name, itenirary.To.Name),
				Inline: true,
			},
			{
				Name:   "Schedule",
				Value:  fmt.Sprintf("%s **%s**", scheduleType, time),
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

func addTripAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, state *state.State) error {
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

	choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
		Name:  fmt.Sprintf("🕒 Suggestions for %s", time.Now().Format("15:04")),
		Value: "REFRESH_HEADER",
	})

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})

	return err
}

func timeSuggestions(input string) []time.Time {
	now := time.Now()
	y, m, d, loc := now.Year(), now.Month(), now.Day(), now.Location()
	s := strings.ReplaceAll(input, ":", "")
	var choices []time.Time

	date := func(h, min int) time.Time {
		return time.Date(y, m, d, h, min, 0, 0, loc)
	}
	now = date(now.Hour(), now.Minute())
	// Internal helper to reduce repeated formatting loops
	add := func(h, min, count int) {
		start := date(h, min)
		for i := range count {
			choices = append(choices, start.Add(time.Duration(i*15)*time.Minute))
		}
	}

	v, _ := strconv.Atoi(s) // Parse clean input as int for shared use

	switch len(s) {
	case 0:
		start := now.Add(time.Duration(15-(now.Minute()%15)) * time.Minute)
		for i := range 8 {
			choices = append(choices, start.Add(time.Duration(i*15)*time.Minute))
		}
	case 1:
		if v >= 0 && v <= 23 {
			add(v, 0, 4)
		}
	case 2:
		if v >= 0 && v <= 23 {
			add(v, 0, 5) // Valid hour (e.g. "12")
		} else if len(input) == 2 {
			// Fallback for non-hour inputs (e.g. "25" -> 02:05)
			h, min := int(s[0]-'0'), int(s[1]-'0')
			if h < 10 && min < 10 {
				choices = append(choices, time.Date(y, m, d, h, min, 0, 0, loc))
				if min <= 5 {
					add(h, min*10, 5)
				}
			}
		}
	case 3:
		d1, d2, d3 := int(s[0]-'0'), int(s[1]-'0'), int(s[2]-'0')
		candidates := []struct{ h, m int }{{d1, d2*10 + d3}, {d1*10 + d2, d3}, {d1*10 + d2, d3 * 10}}
		for _, t := range candidates {
			if t.h < 24 && t.m < 60 {
				choices = append(choices, time.Date(y, m, d, t.h, t.m, 0, 0, loc))
			}
		}
	case 4:
		h, min := v/100, v%100
		if h < 24 && min < 60 {
			add(h, min, 1)
		}
	}

	return choices
}
