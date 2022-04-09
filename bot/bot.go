package bot

import (
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

// StartBot starts the bot and returns any errors that might occu
func StartBot() *discordgo.Session {

	// Checks if bot token is present
	if len(config.CONFIG.Token) == 0 {
		malm.Fatal("ERROR! No token provided!")
	}

	// Creates the bot/session
	session, err := discordgo.New("Bot " + config.CONFIG.Token)
	if err != nil {
		return nil
	}

	// Loads all the valid commands into a map
	mapValidCommands()

	// Adds message handler (https://github.com/bwmarrin/discordgo/blob/37088aefec2241139e59b9b804f193b539be25d6/eventhandlers.go#L937)
	session.AddHandler(messageHandler)
	session.AddHandler(readyHandler)
	session.AddHandler(messageUpdateHandler)

	// Attempts to open connection
	err = session.Open()
	if err != nil {
		malm.Fatal("%v", err)
	}

	// Returns session
	return session
}
