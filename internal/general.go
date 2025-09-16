package internal

import (
	"log"
	"os"

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
	teamName     = os.Getenv("TEAM_NAME")
)

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
		teamName = os.Getenv("TEAM_NAME")

	} else {
		log.Println("Environment variables loaded")
	}
}
