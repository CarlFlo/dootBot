package commands

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func Profile(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	var work database.Work
	work.GetWorkInfo(&user)

	var daily database.Daily
	daily.GetDailyInfo(&user)

	// You have this much money
	//netWorth := utils.HumanReadableNumber(user.Money + bank.Money)

	// The statuses on the cooldown's
	workStatus := config.CONFIG.Emojis.Success
	if !work.CanDoWork() {
		workStatus = fmt.Sprintf("%s Available %s", config.CONFIG.Emojis.Failure, work.CanDoWorkAt())
	}

	dailyStatus := config.CONFIG.Emojis.Success
	if !daily.CanDoDaily() {
		dailyStatus = fmt.Sprintf("%s Available %s", config.CONFIG.Emojis.Failure, daily.CanDoDailyAt())
	}

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		&discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Color:       config.CONFIG.Colors.Neutral,
			Title:       fmt.Sprintf("%s#%s profile", m.Author.Username, m.Author.Discriminator),
			Description: "",
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("Wallet %s", config.CONFIG.Emojis.Wallet),
					Value:  fmt.Sprintf("%s %s", config.CONFIG.Emojis.Economy, user.PrettyPrintMoney()),
					Inline: true,
				},
				/*
					&discordgo.MessageEmbedField{
						Name:   fmt.Sprintf("Net Worth %s", config.CONFIG.Emojis.NetWorth),
						Value:  fmt.Sprintf("%s %s", config.CONFIG.Emojis.Economy, netWorth),
						Inline: true,
					},
				*/
				&discordgo.MessageEmbedField{
					Name:   "Daily",
					Value:  dailyStatus,
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Work",
					Value:  workStatus,
					Inline: true,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Profile footer",
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("%s#%s", m.Author.AvatarURL("256"), m.Author.ID),
			},
		},
	}}

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}

}
