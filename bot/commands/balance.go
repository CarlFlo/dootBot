package commands

import (
	"fmt"
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

// Balance - Output the users balance to the chat
func Balance(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	var user database.User

	user.QueryUserByDiscordID(m.Author.ID)

	description := fmt.Sprintf("As of <t:%d:R>", time.Now().Unix())

	//netWorth := utils.HumanReadableNumber(user.Money + bank.Money)

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Color:       config.CONFIG.Colors.Neutral,
			Title:       "Here is your financial information",
			Description: description,
			Author: &discordgo.MessageEmbedAuthor{
				Name:    fmt.Sprintf("%s#%s", m.Author.Username, m.Author.Discriminator),
				IconURL: m.Author.AvatarURL(""),
			},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   fmt.Sprintf("Wallet %s", config.CONFIG.Emojis.Wallet),
					Value:  fmt.Sprintf("%s %s", config.CONFIG.Emojis.Economy, user.PrettyPrintMoney()),
					Inline: true,
				},
				{
					Name:   fmt.Sprintf("Lifetime earnings %s", config.CONFIG.Emojis.NetWorth),
					Value:  fmt.Sprintf("%s %s", config.CONFIG.Emojis.Economy, user.PrettyPrintLifetimeEarnings()),
					Inline: true,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "",
			},
		},
	}}

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}

}
