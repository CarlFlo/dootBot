package utils

import (
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

// GetGuild returns the guild ID from a channel ID
func GetGuild(s *discordgo.Session, m *discordgo.MessageCreate) string {
	channel, err := s.Channel(m.ChannelID)
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
