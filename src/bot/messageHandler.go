package bot

import (
	"fmt"
	"strings"

	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/database"
	"github.com/CarlFlo/dootBot/src/permissions"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore messages from bots
	if m.Author.Bot {
		return
	}

	validateUserExistance(m)

	// Check for prefix
	if strings.HasPrefix(m.Message.Content, config.CONFIG.BotPrefix) {
		// Message is a command

		// Trim length if required. Ignores if 0
		if config.CONFIG.MessageProcessing.MaxIncommingMsgLength != 0 && len(m.Message.Content) > config.CONFIG.MessageProcessing.MaxIncommingMsgLength {
			m.Message.Content = m.Message.Content[:config.CONFIG.MessageProcessing.MaxIncommingMsgLength]
		}

		// Turns the input string to a struct
		permissionCtx := permissions.ResolveMessage(s, m)
		data := &structs.CmdInput{}
		data.ParseInput(m.Message.Content, permissionCtx)

		// Checks that the origin of the message is valid for this command
		if !validateMessageOrigin(m.GuildID, m.ChannelID, data.GetCommand()) {
			return
		}

		// validCommands is a map containing all commands
		if command, ok := validCommands[data.GetCommand()]; ok {

			// Checks if the user has permission to run the command
			if !hasCommandPermission(command.requiredPermission, data) {
				malm.Info("(%s) '%s' tried to run command: '%s'", m.Author.ID, m.Author.Username, data.GetCommand())
				utils.SendMessageFailure(m, fmt.Sprintf("You do not have permission to run `%s` here", data.GetCommand()))
				return
			}

			// Executes the command
			command.function(s, m, data)
		}
		return
	}
	// Message is not a command

}

// If no guild music channels are bound then all channels are allowed.
func fromBoundChannel(guildID, channelID string) bool {
	allowed, err := database.IsGuildMusicChannelAllowed(guildID, channelID)
	if err != nil {
		malm.Error("Error checking guild music channels: %s", err)
		return true
	}

	return allowed
}

// Checks if the message is a direct (private) message
func isDirectMessage(guildID string) bool {
	return len(guildID) == 0
}

// Checks where the message comes from and checks it against rules to
// allow or discard messages. Example: If it is a direct message while
// direct messaging in turned off. Or from an unbound channel, if any exists.
func validateMessageOrigin(guildID, channelID, command string) bool {

	// Check if it is a private message
	if isDirectMessage(guildID) {
		// Is a private message
		if !config.CONFIG.AllowDirectMessages {
			// Direct messages not allowed
			return false
		}
	} else {
		if command != "" && !fromBoundChannel(guildID, channelID) {
			return false
		}
	}

	return true
}

// validateUserExistance - will create a database entry for the user if it does not exist
func validateUserExistance(m *discordgo.MessageCreate) {

	// Checks if the user is in the database
	var user database.User
	if exists := user.DoesUserExist(m.Author.ID); !exists {
		// User does not exist, add them
		//malm.Debug("User %s (%s) not found in database. Creating new entry for user", m.Author.Username, m.Author.ID)
		database.InitializeNewUser(m.Author.ID)
	}

}
