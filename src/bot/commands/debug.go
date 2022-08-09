package commands

import (
	"fmt"
	"runtime"

	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/dootBot/src/utils"

	"github.com/bwmarrin/discordgo"
)

// Debug - prints some debug information
func Debug(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	currentOS := runtime.GOOS

	utils.SendMessageNeutral(m, fmt.Sprintf("OS: %s", currentOS))
}
