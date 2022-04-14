package commands

import (
	"fmt"
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

// Balance - Output the users balance to the chat
func Balance(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput) {

	var user database.User

	user.GetUserByDiscordID(m.Author.ID)

	var bank database.Bank
	bank.GetBankInfo(&user)

	description := fmt.Sprintf("As of <t:%d:R>", time.Now().Unix())

	netWorth := utils.HumanReadableNumber(user.Money + bank.Money)

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		&discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Here is your financial information",
			Description: description,
			Author: &discordgo.MessageEmbedAuthor{
				Name:    fmt.Sprintf("%s#%s", m.Author.Username, m.Author.Discriminator),
				IconURL: m.Author.AvatarURL(""),
			},
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("Wallet %s", config.CONFIG.Emojis.Wallet),
					Value:  fmt.Sprintf("%s %s", config.CONFIG.Emojis.Economy, user.PrettyPrintMoney()),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("Bank %s", config.CONFIG.Emojis.Bank),
					Value:  fmt.Sprintf("%s %s", config.CONFIG.Emojis.Economy, bank.PrettyPrintMoney()),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("Net Worth %s", config.CONFIG.Emojis.NetWorth),
					Value:  fmt.Sprintf("%s %s", config.CONFIG.Emojis.Economy, netWorth),
					Inline: true,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Depositing money at the bank gives you %.2f%% interest if the balance exceedes %d %s.", config.CONFIG.Bank.InterestRate*100, config.CONFIG.Bank.MinAmountForInterest, config.CONFIG.Economy.Name),
			},
		},
	}}

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}

}
