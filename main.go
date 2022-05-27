package main

import (
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/bot"
	"github.com/CarlFlo/DiscordMoneyBot/bot/music"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"github.com/CarlFlo/malm"
)

const CurrentVersion = "2022-05-27"

func init() {

	utils.Clear()

	rand.Seed(time.Now().UTC().UnixNano())

	malm.SetLogVerboseBitmask(39) // Turns of verbose for debug and info log messages
	malm.Debug("Running on %s", runtime.GOOS)

	if err := config.LoadConfiguration(); err != nil {
		malm.Fatal("Error loading configuration: %v", err)
	}

	if err := database.SetupDatabase(); err != nil {
		malm.Fatal("Database initialization error: %s", err)
	}

	if err := music.InitializeMusic(); err != nil {
		malm.Info("Music disabled. %s", err.Error())
	}

	// Handles checking if there is an update available for the bot
	upToDate, githubVersion, err := utils.BotVersonHandler(CurrentVersion)
	if err != nil {
		malm.Error("%s", err)
	}

	if upToDate {
		malm.Debug("Version %s", CurrentVersion)
	} else {
		malm.Info("New version available! New version: '%s'; Your version: '%s'", githubVersion, CurrentVersion)
	}
}

func main() {

	session := bot.StartBot()

	time.Sleep(500 * time.Millisecond) // Added this sleep so the messages below will come last
	// Keeps bot from closing. Waits for CTRL-C
	malm.Info("Press CTRL-C to initiate shutdown")
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
