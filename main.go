package main

import (
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/CarlFlo/dootBot/src/bot"
	"github.com/CarlFlo/dootBot/src/bot/context"
	"github.com/CarlFlo/dootBot/src/bot/music"
	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/database"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/CarlFlo/malm"
)

const CurrentVersion = "2023-08-16"

func init() {

	malm.SetLogVerboseBitmask(39) // Turns of verbose for debug and info log messages
	rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	utils.Clear()
	malm.Debug("Running on %s", runtime.GOOS)

	config.Load()
	database.Connect()
	music.Initialize()

	go utils.CheckVersion(CurrentVersion)
}

func main() {

	session := bot.StartBot()
	context.SESSION = session

	// Waits for a CTRL-C
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	malm.Info("Shutting down!")

	// Run cleanup code here
	close(sc)
	session.Close() // Stops the discord bot
}

// Invite bot
// https://discordapp.com/oauth2/authorize?&client_id=<ID_GOES_HERE>&scope=bot&permissions=37211200
