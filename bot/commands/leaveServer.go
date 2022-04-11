package commands

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/bot/commands/cmdutils"
	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"

	"github.com/bwmarrin/discordgo"
)

// LeaveServer - Leaves the server of the guild ID provided
func LeaveServer(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput) {

	if len(input.GetArgs()) == 0 {
		cmdutils.SendDirectMessage(s, m, "No guild ID provided")
		return
	}

	g, _ := s.Guild(input.GetArgs()[0])

	if err := s.GuildLeave(input.GetArgs()[0]); err != nil {
		cmdutils.SendDirectMessage(s, m, fmt.Sprintf("Error leaving the server! %s", err))
		return
	}

	cmdutils.SendDirectMessage(s, m, fmt.Sprintf("Successfully left '%s' (guildID: %s)", g.Name, input.GetArgs()[0]))
}
