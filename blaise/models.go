package blaise

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Time uint32

func (t Time) ToHMSString() string {
	h := t / 3600
	m := (t % 3600) / 60
	s := t % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

type Location struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Name       string     `json:"name"`
	Coordinate Coordinate `json:"coordinate"`
}

type Coordinate struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

type Itinerary struct {
	From          Location `json:"from"`
	To            Location `json:"to"`
	DepartureTime Time     `json:"departure_time"`
	ArrivalTime   Time     `json:"arrival_time"`
	Legs          []Leg    `json:"legs"`
}

type Leg struct {
	From          Location `json:"from"`
	To            Location `json:"to"`
	DepartureTime Time     `json:"departure_time"`
	ArrivalTime   Time     `json:"arrival_time"`
	Stops         []Stop   `json:"stops"`
	Shapes        []Shape  `json:"shapes"`
	Mode          string   `json:"mode"`
	HeadSign      *string  `json:"head_sign"`
	LongName      *string  `json:"long_name"`
	ShortName     *string  `json:"short_name"`
}

type Stop struct {
	Location        Location `json:"location"`
	DepartureTime   Time     `json:"departure_time"`
	ArrivalTime     Time     `json:"arrival_time"`
	DistanceTraveld float32  `json:"distance_traveld"`
}

type Shape struct {
	Location        Location `json:"location"`
	Sequence        uint     `json:"sequence"`
	DistanceTraveld float32  `json:"distance_traveld"`
}

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

func IteniraryToEmbedFields(itinerary Itinerary) []*discordgo.MessageEmbedField {
	fields := make([]*discordgo.MessageEmbedField, 0, len(itinerary.Legs))
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

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   legTitle,
			Value:  value,
			Inline: false,
		})
	}
	return fields
}
