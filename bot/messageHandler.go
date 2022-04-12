package bot

import (
	"strings"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore messages from bots
	if m.Author.Bot {
		return
	}

	createUserIfNoitExists(m)

	// Check for prefix
	if strings.HasPrefix(m.Message.Content, config.CONFIG.BotPrefix) {
		// Message is a command

		// Checks that the origin of the message is valid
		if !validateMessageOrigin(m.GuildID, m.ChannelID) {
			return
		}

		// Trim length if required. Ignores if 0
		if config.CONFIG.MessageProcessing.MaxIncommingMsgLength != 0 && len(m.Message.Content) > config.CONFIG.MessageProcessing.MaxIncommingMsgLength {
			m.Message.Content = m.Message.Content[:config.CONFIG.MessageProcessing.MaxIncommingMsgLength]
		}

		// Turns the input string to a struct
		data := structs.CmdInput{}
		data.ParseInput(m.Message.Content, isOwner(m.Author.ID))

		// validCommands is a map containing all commands
		if command, ok := validCommands[data.GetCommand()]; ok {

			// Checks if the user has permission to run the command
			if command.requiredPermission == enumAdmin && !data.IsAdmin() {
				malm.Info("(%s) '%s' tried to run command: '%s'", m.Author.ID, m.Author.Username, data.GetCommand())
				return
			}

			// Executes the command
			command.function(s, m, data)
		}
	} else {
		// Message is not a command

	}
}

// Will only allowed messages from bound channels, if any are specified.
// If no bound channels are specified will all channels be allowed
// Does not handle direct messages
func fromBoundChannel(channelID string) bool {

	// If list is empty then allow everything
	if len(config.CONFIG.BoundChannels) == 0 {
		return true
	}

	// Iterate over all bound channels
	for _, allowedID := range config.CONFIG.BoundChannels {
		if channelID == allowedID {
			return true
		}
	}

	return false
}

// Checks if the message is a direct (private) message
func isDirectMessage(guildID string) bool {
	return len(guildID) == 0
}

// Checks where the message comes from and checks it against rules to
// allow or discard messages. Example: If it is a direct message while
// direct messaging in turned off. Or from an unbound channel, if any exists.
func validateMessageOrigin(guildID, channelID string) bool {

	// Check if it is a private message
	if isDirectMessage(guildID) {
		// Is a private message
		if !config.CONFIG.AllowDirectMessages {
			// Direct messages not allowed
			return false
		}
	} else {
		// Check if message is from a bound channel or if bound channels are used
		if !fromBoundChannel(channelID) {
			// Not from a bound channel
			return false
		}
	}

	return true
}

// createUserIfNoitExists - will create a database entry for the user if it does not exist
func createUserIfNoitExists(m *discordgo.MessageCreate) {

	// Checks if the user is in the database
	var user database.User
	if exists := user.UserExists(m.Author.ID); !exists {
		// User does not exist, add them
		//malm.Debug("User %s (%s) not found in database. Creating new entry for user", m.Author.Username, m.Author.ID)
		database.InitializeNewUser(m.Author.ID)
	}

}
