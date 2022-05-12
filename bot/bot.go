package bot

import (
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

// StartBot starts the bot and returns any errors that might occu
func StartBot() *discordgo.Session {

	variableCheck()

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
	session.AddHandler(interactionHandler)

	// Attempts to open connection
	err = session.Open()
	if err != nil {
		malm.Fatal("%s", err)
	}

	// Returns session
	return session
}

func variableCheck() {

	// This function checks if some important variables are set in the config file
	problem := false

	if len(config.CONFIG.Token) == 0 {
		malm.Error("No bot Token provided in the config file!")
		problem = true
	}

	if len(config.CONFIG.BotInfo.AppID) == 0 {
		malm.Error("No AppID provided in the config file! (The bot's Discord ID)")
		problem = true
	}

	if len(config.CONFIG.OwnerID) == 0 {
		malm.Error("No OwnerID provided in the config file! (This should be your Discord ID)")
		problem = true
	}

	if problem {
		malm.Fatal("There are at least one variable missing in the configuration file. Please fix the above errors!")
	}

}
