package commands

import (
	"fmt"
	"runtime"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"

	"github.com/bwmarrin/discordgo"
)

// Debug - prints some debug information
func Debug(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput) {

	currentOS := runtime.GOOS

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("OS: %s", currentOS))
}
