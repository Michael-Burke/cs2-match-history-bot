## Overview

Discord bot for Lurker Gaming that tracks specified FACEIT CS2 players and posts weekly/current‑week summaries to a Discord channel. Provides slash commands for manual refresh and for listing tracked players.

## Features

- **Hourly refresh**: pulls FACEIT stats and updates a pinned/rolling status message
- **Slash commands**: `/refresh`, `/list-players`
- **Configurable window/time zone**: week is Monday→Monday, `TIME_ZONE` supported
- **Optional team filter**: set `TEAM_NAME` to exclude league/team matches
- **Player list file**: `data/faceit_player_names.json`

## Quick start

1) Create `.env` (see Environment) and populate Discord/FACEIT values
2) Add FACEIT nicknames to `data/faceit_player_names.json`
3) Run via Docker or locally

Docker
```bash
docker build -t lg-cs2-bot:latest /home/sedare/coding/lurker-gaming-cs2-bot
docker compose -f /home/sedare/coding/lurker-gaming-cs2-bot/docker-compose.yaml up -d
```

Local
```bash
go run /home/sedare/coding/lurker-gaming-cs2-bot/main.go
```

## Environment

- `DISCORD_BOT_TOKEN`
- `DISCORD_APP_ID`
- `DISCORD_GUILD_ID` (register per‑guild for instant commands)
- `DISCORD_UPDATE_CHANNEL_ID` (channel to post summaries)
- `FACEIT_GAME_ID` (e.g., `CS2`)
- `FACEIT_APP_ID`
- `FACEIT_API_KEY`
- `TIME_ZONE` (default `US/Eastern`)
- `TEAM_NAME` (optional: exclude this team’s matches)

## Commands

- `/refresh`: refreshes current and last week and posts/updates summaries
- `/list-players`: lists tracked players and resolved FACEIT IDs

## Player list file

`data/faceit_player_names.json`
```json
{ "players": ["Sedare", "AnotherPlayer"] }
```