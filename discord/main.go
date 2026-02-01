package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/vincbro/pascal/blaise"
	"github.com/vincbro/pascal/database"
	"github.com/vincbro/pascal/state"
)

func main() {
	_ = slog.SetLogLoggerLevel(slog.LevelDebug)
	slog.Info("Loading env variables")
	err := godotenv.Load()
	if err != nil {
		slog.Error("error loading .env file", "error", err)
		os.Exit(1)
	}
	discordKey := os.Getenv("DISCORD_KEY")
	appID := os.Getenv("APP_ID")
	guildID := os.Getenv("GUILD_ID")
	blaiseUrl := os.Getenv("BLAISE_URL")
	slog.Debug("Settings: ", "Guild ID", guildID)

	slog.Info("Creating blaise client")
	bClient := blaise.NewClient(blaiseUrl)
	slog.Info("Connecting to database")
	db, err := database.NewDatabase("main.db")
	if err != nil {
		slog.Error("error connection to database", "error", err)
		os.Exit(2)
	}
	state := state.NewState(db, bClient)

	slog.Info("Created discord session")
	dg, err := discordgo.New("Bot " + discordKey)
	if err != nil {
		slog.Error("error creating Discord session", "error", err)
		os.Exit(3)
	}

	slog.Info("Sending commands to discord")
	cmds := GetCommands()
	err = cmds.BulkOverwrite(dg, appID, guildID)
	if err != nil {
		slog.Error("error creating commands", "error", err)
		os.Exit(4)
	}

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		slog.Info("Pascal is now online!", "user", r.User.Username)
		s.UpdateGameStatus(0, "Watching your commute 🚆")
	})

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		InteractionCreate(s, i, cmds, state)
	})

	slog.Info("Opening discord bot connection")
	err = dg.Open()
	if err != nil {
		slog.Error("error opening connection", "error", err)
		os.Exit(5)
	}
	defer dg.Close()

	// Wait here until CTRL-C or other term signal is received.
	slog.Info("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
