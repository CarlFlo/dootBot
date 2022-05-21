package farming

import (
	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func printFarm(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	var farm database.Farm
	farm.QueryUserFarmData(&user)

	complexMessage := &discordgo.MessageSend{}
	farm.CreateFarmOverview(complexMessage, m, &user)

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}
}
