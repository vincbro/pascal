package main

import (
	"fmt"
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
	st := state.NewState(db, bClient)

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
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		InteractionCreate(s, i, cmds, st)
	})
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
		MessageReactionAdd(s, r, cmds, st)
	})
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
		MessageReactionRemove(s, r, cmds, st)
	})

	st.AddHandler(func(s *state.State, request state.Request) error {
		user, err := s.DB.GetUser(request.UserID)
		if err != nil {
			return err
		}
		trip, err := s.DB.GetTrip(request.UserID, request.TripID)
		if err != nil {
			return err
		}
		embed := &discordgo.MessageEmbed{
			Title:  request.Message,
			Fields: blaise.IteniraryToEmbedFields(trip.ExpectedItinerary),
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Pascal • TripID: %s", trip.ID),
			},
		}
		msg, err := dg.ChannelMessageSendEmbed(user.ChannelID, embed)
		if err != nil {
			return err
		}
		return dg.MessageReactionAdd(msg.ChannelID, msg.ID, "🤫")
	})

	slog.Info("Opening discord bot connection")
	err = dg.Open()
	if err != nil {
		slog.Error("error opening connection", "error", err)
		os.Exit(5)
	}
	defer dg.Close()

	slog.Info("Starting trip watcher")
	st.Start()
	defer st.Stop()

	// Wait here until CTRL-C or other term signal is received.
	slog.Info("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
