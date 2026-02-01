package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/vincbro/pascal/cmd/bot/commands"
	"github.com/vincbro/pascal/internal/blaise"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("error loading .env file")
	}
	discordKey := os.Getenv("DISCORD_KEY")
	appID := os.Getenv("APP_ID")
	guildID := os.Getenv("GUILD_ID")
	blaiseUrl := os.Getenv("BLAISE_URL")

	client := blaise.NewClient(blaiseUrl)

	dg, err := discordgo.New("Bot " + discordKey)
	if err != nil {
		fmt.Println("error creating Discord session", err)
	}

	cmds := commands.GetCommands()
	err = cmds.BulkOverwrite(dg, appID, guildID)
	if err != nil {
		fmt.Println("error creating App commands", err)
	}

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		InteractionCreate(s, i, cmds, *client)
	})

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	defer dg.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
