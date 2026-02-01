package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/vincbro/pascal/blaise"
	"github.com/vincbro/pascal/state"
)

func CreateNewTripCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "new",
			Description: "Create a new trip",
			// Contexts: &[]discordgo.InteractionContextType{
			// 	discordgo.InteractionContextBotDM,
			// 	discordgo.InteractionContextGuild,
			// 	discordgo.InteractionContextPrivateChannel,
			// },

			// IntegrationTypes: &[]discordgo.ApplicationIntegrationType{
			// 	discordgo.ApplicationIntegrationUserInstall,
			// 	discordgo.ApplicationIntegrationGuildInstall,
			// },
			Options: []*discordgo.ApplicationCommandOption{
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
			},
		},
		Handler:      newTripHandler,
		Autocomplete: newTripAutocomplete,
	}
}

func newTripHandler(s *discordgo.Session, i *discordgo.InteractionCreate, state *state.State) error {
	opts := ParseOptions(i.ApplicationCommandData().Options)

	from := opts["from"].Value.(string)
	to := opts["to"].Value.(string)

	itinerary, err := state.BClient.Routing(context.Background(), from, to)
	if err != nil {
		return err
	}

	embed := formatRouteEmbed(itinerary)
	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
	return err
}

func newTripAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, state *state.State) error {
	data := i.ApplicationCommandData()
	var choices []*discordgo.ApplicationCommandOptionChoice

	for _, option := range data.Options {
		if !option.Focused {
			continue
		}
		switch option.Name {
		case "from", "to":
			// User is typing in the "from" field
			results, err := state.BClient.SearchAreas(context.Background(), option.StringValue(), 10)
			if err != nil {
				fmt.Println("error failed to search for area", err)
				return err
			}
			for _, area := range results {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  area.Name,
					Value: area.ID,
				})
			}

		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	return err
}

// TEMP
func getModeEmoji(mode string) string {
	switch mode {
	case "Tram":
		return "🚋"
	case "Subway":
		return "🚇"
	case "Rail":
		return "🚆"
	case "Bus":
		return "🚌"
	case "Ferry":
		return "⛴️"
	case "Walk":
		return "🚶"
	case "Transfer":
		return "🔄"
	default:
		return "❓"
	}
}

// TEMP
func formatRouteEmbed(itinerary blaise.Itenirary) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       "📍 Route Details",
		Color:       0x3498db, // Pascal Blue
		Description: fmt.Sprintf("**From:** %s\n**To:** %s", itinerary.From.Name, itinerary.To.Name),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Powered by Blaise Engine",
		},
	}

	for _, leg := range itinerary.Legs {
		emoji := getModeEmoji(leg.Mode)

		// 1. Format the title of the leg (e.g., "Bus 50 (Towards Centralen)")
		legTitle := fmt.Sprintf("%s %s", emoji, leg.Mode)
		if leg.ShortName != nil {
			legTitle = fmt.Sprintf("%s %s", emoji, *leg.ShortName)
		}
		if leg.HeadSign != nil {
			legTitle += fmt.Sprintf(" (Towards %s)", *leg.HeadSign)
		}

		// 2. Build the detailed "Value" using a strings.Builder for efficiency
		var sb strings.Builder

		// Departure time and location
		fmt.Fprintf(&sb, "`%s` ➔ `%s`\n", leg.DepartureTime.ToHMSString(), leg.ArrivalTime.ToHMSString())
		fmt.Fprintf(&sb, "**Start:** %s\n", leg.From.Name)

		// 3. Add the intermediate stops
		if len(leg.Stops) > 0 {
			for _, stop := range leg.Stops {
				// Avoid redundancy: skip the stop if it's the same as the leg's starting location
				if stop.Location.ID == leg.From.ID || stop.Location.ID == leg.To.ID {
					continue
				}
				// Format as a bullet point: • 12:05 Stop Name
				fmt.Fprintf(&sb, "• `%s` %s\n", stop.ArrivalTime.ToHMSString(), stop.Location.Name)
			}
		}
		fmt.Fprintf(&sb, "**End:** %s\n", leg.To.Name)

		// Ensure the field value does not exceed Discord's 1024-character limit
		value := sb.String()
		if len(value) > 1021 {
			value = value[:1018] + "..."
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   legTitle,
			Value:  value,
			Inline: false,
		})
	}

	return embed
}
