package commands

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

// Balance - Output the users balance to the chat
func Balance(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput) {

	var user database.User
	var bank database.Bank
	user.GetUserByDiscordID(m.Author.ID)

	//timestamp := fmt.Sprintf("Timestamp: <t:%d:R>", time.Now().Unix())

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		&discordgo.MessageEmbed{
			Type:  discordgo.EmbedTypeRich,
			Title: "Here is your financial information",
			Author: &discordgo.MessageEmbedAuthor{
				Name:    fmt.Sprintf("%s#%s", m.Author.Username, m.Author.Discriminator),
				IconURL: m.Author.AvatarURL(""),
			},
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "Wallet",
					Value:  fmt.Sprintf("%s %s", config.CONFIG.Economy.Emoji, user.PrettyPrintMoney()),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Bank",
					Value:  fmt.Sprintf("%s %s", config.CONFIG.Economy.Emoji, bank.PrettyPrintMoney()),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Net Worth",
					Value:  fmt.Sprintf("%s %s", config.CONFIG.Economy.Emoji, user.PrettyPrintMoney()),
					Inline: true,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Depositing money in the bank will earn you interest.",
			},
		},
	}}

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}

}
