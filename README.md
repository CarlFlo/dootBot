# DootBot

## About

DootBot is a Discord bot focused on progression-style commands like work, daily rewards, farming, and music playback.

## Features

- Economy commands with cooldowns and progression
- Button-driven interactions for profile, farming, and music controls
- Music playback with queue support powered by Lavalink

## Commands

- `work` earns a random amount of money every 6 hours
- `daily` gives a daily reward every 24 hours
- `farm` lets users manage crops and plots
- `play` plays a track from a URL or search query in the user's voice channel

## Setup

The first time the bot runs it creates `config.json`.

Required config values:

1. `token`
2. `ownerID`
3. `botInfo.appID`

Optional music config lives under `music.lavalink` and can also be overridden with environment variables:

- `LAVALINK_HOST`
- `LAVALINK_PORT`
- `LAVALINK_PASSWORD`
- `LAVALINK_SECURE`

Example config block:

```json
"music": {
  "enableMusic": true,
  "maxSongLengthMinutes": 120,
  "lavalink": {
    "host": "127.0.0.1",
    "port": 2333,
    "password": "youshallnotpass",
    "secure": false
  }
}
```

## Lavalink

The bot no longer streams audio directly. DiscordGo is only used for the normal gateway, commands, and text interactions. Voice playback is delegated to Lavalink and must run separately.

For Discord's current DAVE/E2EE voice requirement:

- Use Lavalink `4.2.0` or newer
- Use a DAVE-capable disgolink release (`3.1.0` or newer; this project currently uses a newer `v4` module build)
- Do not reintroduce DCA, `OpusSend`, `dgvoice`, or manual Opus streaming

The bot bridges Discord voice gateway events into disgolink/Lavalink, including the `channelId`-carrying voice state required by Lavalink `4.2.0+`.

Quick start:

1. Copy `.env.example` if you want environment-based config.
2. Start Lavalink `4.2.0+` with `docker compose up -d`.
3. Start the bot after Lavalink is healthy.

If you expect YouTube playback, enable a YouTube source plugin in Lavalink. A minimal example is included in `lavalink/application.yml`.

## Requirements

- Go
- GCC for `go-sqlite3`
- A running Lavalink `4.2.0+` server for music playback

## Running

- `make build` builds the bot
- `make run` runs the bot
- `go build ./...` and `go run main.go` also work
