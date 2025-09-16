package main

import (
	"log"
	"os"
	"os/signal"

	"lurker-gaming-cs2-bot/internal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	botToken        = os.Getenv("DISCORD_BOT_TOKEN")
	applicationID   = os.Getenv("DISCORD_APP_ID")   // needed to register slash commands
	guildID         = os.Getenv("DISCORD_GUILD_ID") // optional: register per-guild for instant availability
	updateChannelID = os.Getenv("DISCORD_UPDATE_CHANNEL_ID")

	gameName     = os.Getenv("FACEIT_GAME_ID")
	faceitAppID  = os.Getenv("FACEIT_APP_ID")
	faceitAPIKey = os.Getenv("FACEIT_API_KEY")
)
var s *discordgo.Session

func init() {
	loadEnv(true)
	var err error
	s, err = discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatal("Error creating Discord session: ", err)
	}
	log.Println("Discord session created")
	if botToken == "" || applicationID == "" {
		log.Fatal("Missing required env: DISCORD_BOT_TOKEN, DISCORD_APP_ID")
	}
}

func loadEnv(debug bool) {
	if debug {
		log.Println("Loading environment variables...")
		// load .env file from ./.env
		godotenv.Load("./.env")

		botToken = os.Getenv("DISCORD_BOT_TOKEN")
		applicationID = os.Getenv("DISCORD_APP_ID")
		guildID = os.Getenv("DISCORD_GUILD_ID")
		updateChannelID = os.Getenv("DISCORD_UPDATE_CHANNEL_ID")
		gameName = os.Getenv("FACEIT_GAME_ID")
		faceitAppID = os.Getenv("FACEIT_APP_ID")
		faceitAPIKey = os.Getenv("FACEIT_API_KEY")

	} else {
		log.Println("Environment variables loaded")
	}
}

func main() {
	// Open Discord session
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Discord bot connected")
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Error opening Discord session: ", err)
	}
	internal.BotInit(s)
	// Start FACEIT hourly refresher
	stopCh := make(chan struct{})
	go internal.StartFACEITRefresher(s, stopCh)

	defer s.Close()

	// Wait for Ctrl+C
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Bot running. Press Ctrl+C to exit.")
	<-stop
	close(stopCh)
	log.Println("Shutting down…")
}
