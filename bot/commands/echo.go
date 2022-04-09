package commands

import (
	"strings"

	"github.com/CarlFlo/discordBotTemplate/bot/structs"

	"github.com/bwmarrin/discordgo"
)

// Echo - echoes the message
func Echo(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput) {

	content := strings.Join(input.GetArgs(), " ")

	s.ChannelMessageSend(m.ChannelID, content)
}
