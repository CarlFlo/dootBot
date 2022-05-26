package commands

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/utils"

	"github.com/bwmarrin/discordgo"
)

// LeaveServer - Leaves the server of the guild ID provided
func LeaveServer(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	if len(input.GetArgsLowercase()) == 0 {
		utils.SendDirectMessage(s, m, "No guild ID provided")
		return
	}

	g, _ := s.Guild(input.GetArgsLowercase()[0])

	if err := s.GuildLeave(input.GetArgsLowercase()[0]); err != nil {
		utils.SendDirectMessage(s, m, fmt.Sprintf("Error leaving the server! %s", err))
		return
	}

	utils.SendDirectMessage(s, m, fmt.Sprintf("Successfully left '%s' (guildID: %s)", g.Name, input.GetArgsLowercase()[0]))
}
