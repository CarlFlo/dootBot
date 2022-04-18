package commands

import (
	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

// Reload - Reloads the configuration without restarting the application
func Reload(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	if err := config.ReloadConfig(); err != nil {
		malm.Error("Could not reload config! %s", err)
		return
	}

	utils.SendDirectMessage(s, m, "Config reloaded")
}
