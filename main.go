package main

import (
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/bot"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"github.com/CarlFlo/malm"
)

// https://discordapp.com/oauth2/authorize?&client_id=643191140849549333&scope=bot&permissions=37211200

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

	// Handles checking if there is an update available for the bot
	upToDate, githubVersion, err := utils.BotVersonHandler(CurrentVersion)
	if err != nil {
		malm.Error("%s", err)
	}

	if !upToDate {
		malm.Info("New version available! New version: '%s'; Your version: '%s'", githubVersion, CurrentVersion)
	} else {
		malm.Debug("Version %s", CurrentVersion)
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
