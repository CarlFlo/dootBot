<h1 align="center">
  DootBot
</h1>

<p align="center">
  <a href="#about">About</a>
  •
  <a href="#features">Features</a>
  •
  <a href="#commands">Commands</a>
  •
  <a href="#setup">Setup</a>
  •
  <a href="#todo">Todo</a>
  
</p>

![Tests](https://github.com/CarlFlo/dootBot/actions/workflows/go.yml/badge.svg)

## About

This project is made for and is intended to be a fun learning exercise.

The purpose of the bot is to allow members of a Discord channel to engage with the bot and earn *money*. This bot is inspired by and is very similar to idle games. The user can interact with the bot to earn and spend fake money. Actions as well as rewards are locked behind a cooldown/timer, some of which require the user to interact with the bot daily to receive their reward.

## Features

- Easy to use! Uses buttons, emojis and color-coding in conjunction with updating messages to provide a clear and user-friendly experience 

- Engaging the users by providing daily tasks 

- User activities with progression

- Music player with queue

## Commands

- Work - Allows the user to earn a random amount of money [6 hour cooldown]
- Daily - Gives the user a random amount of money daily [24 hour cooldown]
- Farm - Allows the user to plant crops with can be harvested for a monetary reward. Crops must be watered within a timeframe for them to not perish. New crops can be unlocked by planting.
- Mine - Your own dwarven keep where your dwarfs will mine for ore and other precious gems
- Play - Plays a youtube song in the voice channel. Provide an url or search for a song.

## Setup

The first time the bot is run, a config.json file is created. This JSON file requires some information in order to make the bot function.

1. The bot token [Token]
2. The owners (your) Discord ID [OwnerID]
3. The bots Discord ID [AppID]
4. (optinal) Youtube API key for music [youtubeAPIKey]

### Requirements 

Needs to be available in the PATH

**For the music**
* [**ffmpeg**](https://ffmpeg.org/download.html)
* [**yt-dlp**](https://github.com/yt-dlp/yt-dlp/releases)

**Additional requirements**
gcc

### Configuration

The bot is fully customizable through the config.json file, where most if not all variables can be customised.

### Running

You're able to build and run the bot with the included `makefile`.

Run `make build` or `make b` to build the bot

Run `make run` or `make r` to just run the bot

The `makefile` can also be used to build/run the bot for different operating systems such as **Windows**, **Mac**, and **Linux**

It is also possible to build or run it yourself with the `go run main.go` and `go build main.go` command, respectively.

## Todo

