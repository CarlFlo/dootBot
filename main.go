package main

import (
	"context"
	"math/rand"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/CarlFlo/dootBot/src/bot"
	botcontext "github.com/CarlFlo/dootBot/src/bot/context"
	"github.com/CarlFlo/dootBot/src/bot/music"
	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/database"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/CarlFlo/malm"
)

const CurrentVersion = "2026-06-24"

// https://discordapp.com/oauth2/authorize?&client_id=239142763315200001&scope=bot&permissions=139690691648

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
	botcontext.SESSION = session

	// Wait for a console interrupt such as Ctrl+C.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	malm.Info("Shutting down!")

	// Run cleanup code here
	music.Close()

	session.Close() // Stops the discord bot
}

// Invite bot
// https://discordapp.com/oauth2/authorize?&client_id=<ID_GOES_HERE>&scope=bot&permissions=37211200
