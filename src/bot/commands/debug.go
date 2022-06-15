package commands

import (
	"fmt"
	"runtime"

	"github.com/CarlFlo/DiscordMoneyBot/src/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/src/utils"

	"github.com/bwmarrin/discordgo"
)

// Debug - prints some debug information
func Debug(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	currentOS := runtime.GOOS

	utils.SendMessageNeutral(m, fmt.Sprintf("OS: %s", currentOS))
}
