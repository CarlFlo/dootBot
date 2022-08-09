package commands

import (
	"fmt"

	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

// BotInvite - Sends back the invite link to the bot
func BotInvite(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	if len(config.CONFIG.BotInfo.AppID) == 0 {
		malm.Warn("ClientID not set in config file")
		utils.SendDirectMessage(m, "Unable to create bot invite. Contact the administrator")
		return
	}

	inviteLink := fmt.Sprintf("https://discordapp.com/oauth2/authorize?&client_id=%s&scope=bot&permissions=%d", config.CONFIG.BotInfo.AppID, config.CONFIG.BotInfo.Permission)

	utils.SendDirectMessage(m, inviteLink)
}
