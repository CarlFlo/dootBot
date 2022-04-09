package commands

import (
	"fmt"

	"github.com/CarlFlo/discordBotTemplate/bot/commands/cmdutils"
	"github.com/CarlFlo/discordBotTemplate/bot/structs"
	"github.com/CarlFlo/discordBotTemplate/config"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

// BotInvite - Sends back the invite link to the bot
func BotInvite(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput) {

	if len(config.CONFIG.BotInfo.ClientID) == 0 {
		malm.Warn("ClientID not set in config file")
		cmdutils.SendDirectMessage(s, m, "Unable to create bot invite. Contact the administrator")
		return
	}

	inviteLink := fmt.Sprintf("https://discordapp.com/oauth2/authorize?&client_id=%s&scope=bot&permissions=%d", config.CONFIG.BotInfo.ClientID, config.CONFIG.BotInfo.Permission)

	cmdutils.SendDirectMessage(s, m, inviteLink)
}
