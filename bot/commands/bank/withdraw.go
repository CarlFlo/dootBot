package bank

import (
	"fmt"
	"strconv"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/bwmarrin/discordgo"
)

func Withdraw(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput) {

	if len(input.GetArgs()) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No amount specified!")
		return
	}

	// convert string to int
	amount, err := strconv.Atoi(input.GetArgs()[0])
	if err != nil {
		// Invalid input
		s.ChannelMessageSend(m.ChannelID, "Invalid withdraw amount!")
		return
	} else if amount < 1 {
		// Zero or negative value
		s.ChannelMessageSend(m.ChannelID, "Cannot withdraw zero or negative amounts!")
		return
	}

	var user database.User
	user.GetUserByDiscordID(m.Author.ID)

	var bank database.Bank
	bank.GetBankInfo(&user)

	err = bank.Withdraw(&user, uint64(amount))

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Withdraw failed! %s", err))
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Withdraw successful!")
}
