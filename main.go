package main

import (
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/src/bot"
	"github.com/CarlFlo/DiscordMoneyBot/src/bot/music"
	"github.com/CarlFlo/DiscordMoneyBot/src/config"
	"github.com/CarlFlo/DiscordMoneyBot/src/database"
	"github.com/CarlFlo/DiscordMoneyBot/src/notifyManager"
	"github.com/CarlFlo/DiscordMoneyBot/src/utils"
	"github.com/CarlFlo/malm"
)

const CurrentVersion = "2022-06-12"

func init() {

	malm.SetLogVerboseBitmask(39) // Turns of verbose for debug and info log messages
	rand.Seed(time.Now().UTC().UnixNano())

	utils.Clear()
	malm.Debug("Running on %s", runtime.GOOS)

	config.Load()
	database.Connect()
	music.Initialize()
	notifyManager.Initialize()

	go utils.CheckVersion(CurrentVersion)
}

func main() {

	session := bot.StartBot()

	// Waits for a CTRL-C
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	malm.Info("Shutting down!")

	// Run cleanup code here
	close(sc)
	notifyManager.Stop()
	session.Close() // Stops the discord bot
}

// Invite bot
// https://discordapp.com/oauth2/authorize?&client_id=<ID_GOES_HERE>&scope=bot&permissions=37211200
