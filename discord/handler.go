package main

import (
	"fmt"
	"log/slog"
	"strings"

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

func MessageReactionRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove, cmds Commands, st *state.State) {
	if r.UserID == s.State.User.ID {
		return
	}

	if r.Emoji.Name != "🤫" {
		return
	}

	user, err := st.DB.GetUser(r.UserID)
	if err != nil {
		slog.Error("error while getting user", "error", err)
		return
	}

	msg, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		slog.Error("error while getting msg", "error", err)
		return
	}

	if len(msg.Embeds) > 0 && msg.Embeds[0].Footer != nil {
		footerText := msg.Embeds[0].Footer.Text
		words := strings.Fields(footerText)
		if len(words) > 0 {
			tripID := words[len(words)-1]
			trip, err := st.DB.GetTrip(user.ID, tripID)
			if err != nil {
				slog.Error("error while getting trip", "error", err)
				return
			}
			st.UnMuteTrip(trip.ID)
			_, err = s.ChannelMessageSend(r.ChannelID, fmt.Sprintf("🔔 **%s unmuted.**", trip.Name))
			if err != nil {
				slog.Error("error while sending msg", "error", err)
				return
			}
		}
	}
}

func MessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd, cmds Commands, st *state.State) {
	if r.UserID == s.State.User.ID {
		return
	}

	if r.Emoji.Name != "🤫" {
		return
	}

	user, err := st.DB.GetUser(r.UserID)
	if err != nil {
		slog.Error("error while getting user", "error", err)
		return
	}

	msg, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		slog.Error("error while getting msg", "error", err)
		return
	}

	if len(msg.Embeds) > 0 && msg.Embeds[0].Footer != nil {
		footerText := msg.Embeds[0].Footer.Text
		words := strings.Fields(footerText)
		if len(words) > 0 {
			tripID := words[len(words)-1]
			trip, err := st.DB.GetTrip(user.ID, tripID)
			if err != nil {
				slog.Error("error while getting trip", "error", err)
				return
			}
			st.MuteTrip(trip.ID)
			_, err = s.ChannelMessageSend(r.ChannelID, fmt.Sprintf("🔕 **%s muted.**", trip.Name))
			if err != nil {
				slog.Error("error while sending msg", "error", err)
				return
			}
		}
	}

}
