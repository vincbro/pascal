package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/vincbro/pascal/internal/blaise"
)

type CommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, client blaise.Client) error

type Command struct {
	Definition   *discordgo.ApplicationCommand
	Handler      CommandHandler
	Autocomplete CommandHandler
}

type Commands map[string]*Command

func (c *Commands) BulkOverwrite(dg *discordgo.Session, appID string, guildID string) error {
	values := make([]*discordgo.ApplicationCommand, 0, len(*c))
	for k := range *c {
		values = append(values, (*c)[k].Definition)
	}
	createdCommands, err := dg.ApplicationCommandBulkOverwrite(appID, guildID, values)
	if err != nil {
		return err
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
