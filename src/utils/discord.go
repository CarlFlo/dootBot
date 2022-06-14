package utils

import (
	"github.com/CarlFlo/DiscordMoneyBot/src/config"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

// SendDirectMessage will send a direct messag to a user
func SendDirectMessage(s *discordgo.Session, m *discordgo.MessageCreate, content string) (*discordgo.Message, error) {
	ch, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		return nil, err
	}
	return s.ChannelMessageSend(ch.ID, content)
}

func SendMessageSuccess(s *discordgo.Session, m *discordgo.MessageCreate, content string) (*discordgo.Message, error) {
	return sendMessageEmbed(s, m, content, config.CONFIG.Colors.Success)
}

func SendMessageFailure(s *discordgo.Session, m *discordgo.MessageCreate, content string) (*discordgo.Message, error) {
	return sendMessageEmbed(s, m, content, config.CONFIG.Colors.Failure)
}

func SendMessageNeutral(s *discordgo.Session, m *discordgo.MessageCreate, content string) (*discordgo.Message, error) {
	return sendMessageEmbed(s, m, content, config.CONFIG.Colors.Neutral)
}

func sendMessageEmbed(s *discordgo.Session, m *discordgo.MessageCreate, content string, color int) (*discordgo.Message, error) {
	return s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
		Description: content,
		Color:       color,
	})
}

/* */

// GetGuild returns the guild ID from a channel ID
func GetGuild(s *discordgo.Session, channelID string) string {
	channel, err := s.Channel(channelID)
	if err != nil {
		malm.Warn("Failed to get channel: %s", err)
	}
	return channel.GuildID
}

// FindVoiceChannel finds the voice channel containing a specific user by their discord ID
func FindVoiceChannel(s *discordgo.Session, userID string) string {

	for _, g := range s.State.Guilds {
		for _, v := range g.VoiceStates {
			if v.UserID == userID {
				return v.ChannelID
			}
		}
	}
	return ""
}
