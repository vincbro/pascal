package main

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/vincbro/pascal/database"
	"github.com/vincbro/pascal/state"
)

type CommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, state *state.State) error

type Command struct {
	Definition   *discordgo.ApplicationCommand
	Handler      CommandHandler
	Autocomplete CommandHandler
}

type Commands map[string]*Command

func (c *Commands) BulkOverwrite(dg *discordgo.Session, appID string, guildID string) error {
	values := make([]*discordgo.ApplicationCommand, 0, len(*c))
	for k := range *c {
		def := (*c)[k].Definition

		def.IntegrationTypes = &[]discordgo.ApplicationIntegrationType{
			discordgo.ApplicationIntegrationUserInstall,
		}
		def.Contexts = &[]discordgo.InteractionContextType{
			discordgo.InteractionContextBotDM,
		}

		values = append(values, def)
	}

	createdCommands, err := dg.ApplicationCommandBulkOverwrite(appID, guildID, values)
	if err != nil {
		return fmt.Errorf("failed to bulk overwrite global commands: %w", err)
	}

	for _, command := range createdCommands {
		if cmd, ok := (*c)[command.Name]; ok {
			cmd.Definition = command
		}
	}
	return nil
}

func (c *Commands) Add(command Command) {
	(*c)[command.Definition.Name] = &command
}

func GetCommands() Commands {
	c := make(Commands)
	c.Add(CreateNewTripCommand())
	return c
}

type OptionMap map[string]*discordgo.ApplicationCommandInteractionDataOption

func ParseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) OptionMap {
	om := make(OptionMap)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return om
}

func GetUser(dUser *discordgo.User, channelID string, state *state.State) (*database.User, error) {
	user, err := state.DB.GetOrCreateUser(dUser.ID, dUser.Username)
	if err != nil {
		return nil, err
	}
	if channelID != "" && channelID != user.ChannelID {
		slog.Debug("Updating user channel id", "old", user.ChannelID, "new", channelID)
		user.ChannelID = channelID
		if err = state.DB.UpdateUser(user); err != nil {
			return nil, err
		}
	}
	return user, nil
}
