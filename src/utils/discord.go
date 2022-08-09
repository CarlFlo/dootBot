package utils

import (
	"github.com/CarlFlo/dootBot/src/bot/context"
	"github.com/CarlFlo/dootBot/src/config"
	"github.com/bwmarrin/discordgo"
)

// SendDirectMessage will send a direct messag to a user
func SendDirectMessage(m *discordgo.MessageCreate, content string) (*discordgo.Message, error) {
	ch, err := context.SESSION.UserChannelCreate(m.Author.ID)
	if err != nil {
		return nil, err
	}
	return context.SESSION.ChannelMessageSend(ch.ID, content)
}

func SendMessageSuccess(m *discordgo.MessageCreate, content string) (*discordgo.Message, error) {
	return sendMessageEmbed(m, content, config.CONFIG.Colors.Success)
}

func SendMessageFailure(m *discordgo.MessageCreate, content string) (*discordgo.Message, error) {
	return sendMessageEmbed(m, content, config.CONFIG.Colors.Failure)
}

func SendMessageNeutral(m *discordgo.MessageCreate, content string) (*discordgo.Message, error) {
	return sendMessageEmbed(m, content, config.CONFIG.Colors.Neutral)
}

func sendMessageEmbed(m *discordgo.MessageCreate, content string, color int) (*discordgo.Message, error) {
	return context.SESSION.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Description: content,
		Color:       color,
	})
}

/* */

// GetGuild returns the guild ID from a channel ID
func GetGuild(channelID string) (string, error) {

	channel, err := context.SESSION.Channel(channelID)
	if err != nil {
		/*
			// Workaround for msg interactions when creating a new session
			// This can cause problems if the API requests fails for any other reason
			if strings.Contains(err.Error(), "Unknown Channel") {
				return channelID, nil
			}*/

		return "", err
	}
	return channel.GuildID, nil
}

// FindVoiceChannel finds the voice channel containing a specific user by their discord ID
func FindVoiceChannel(userID string) string {

	for _, g := range context.SESSION.State.Guilds {
		for _, v := range g.VoiceStates {
			if v.UserID == userID {
				return v.ChannelID
			}
		}
	}
	return ""
}
