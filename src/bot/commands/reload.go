package commands

import (
	"github.com/CarlFlo/DiscordMoneyBot/src/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/src/config"
	"github.com/CarlFlo/DiscordMoneyBot/src/utils"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

// Reload - Reloads the configuration without restarting the application
func Reload(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	if err := config.ReloadConfig(); err != nil {
		malm.Error("Could not reload config! %s", err)
		return
	}

	utils.SendDirectMessage(m, "Config reloaded")
}
