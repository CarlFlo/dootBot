package commands

import (
	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/utils"
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
