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

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Color:       config.CONFIG.Colors.Neutral,
			Title:       fmt.Sprintf("%s#%s profile", m.Author.Username, m.Author.Discriminator),
			Description: "",
			Fields:      generateProfileFields(m.Author.ID),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Profile footer",
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("%s#%s", m.Author.AvatarURL("256"), m.Author.ID),
			},
		},
	}}

	if components := createButtonComponent(); components != nil {
		complexMessage.Components = components
	}

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}
}

func generateProfileFields(authorID string) []*discordgo.MessageEmbedField {

	var user database.User
	user.QueryUserByDiscordID(authorID)

	var work database.Work
	work.GetWorkInfo(&user)

	var daily database.Daily
	daily.GetDailyInfo(&user)

	// The statuses on the cooldown's
	workStatus := config.CONFIG.Emojis.Success
	if !work.CanDoWork() {
		workStatus = fmt.Sprintf("%s Available %s", config.CONFIG.Emojis.Failure, work.CanDoWorkAt())
	}

	dailyStatus := config.CONFIG.Emojis.Success
	if !daily.CanDoDaily() {
		dailyStatus = fmt.Sprintf("%s Available %s", config.CONFIG.Emojis.Failure, daily.CanDoDailyAt())
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   fmt.Sprintf("Wallet %s", config.CONFIG.Emojis.Wallet),
			Value:  fmt.Sprintf("%s %s", config.CONFIG.Emojis.Economy, user.PrettyPrintMoney()),
			Inline: true,
		},
		/*
			{
				Name:   fmt.Sprintf("Net Worth %s", config.CONFIG.Emojis.NetWorth),
				Value:  fmt.Sprintf("%s %s", config.CONFIG.Emojis.Economy, netWorth),
				Inline: true,
			},
		*/
		{
			Name:   "Daily",
			Value:  dailyStatus,
			Inline: true,
		},
		{
			Name:   "Work",
			Value:  workStatus,
			Inline: true,
		},
	}

	return fields
}

func createButtonComponent() []discordgo.MessageComponent {

	components := []discordgo.MessageComponent{}

	components = append(components, &discordgo.Button{
		Label:    "",
		Style:    1, // Default purple
		Disabled: false,
		Emoji: discordgo.ComponentEmoji{
			Name: config.CONFIG.Emojis.ComponentEmojiNames.Refresh,
		},
		CustomID: "RP", // 'RP' is code for 'Refresh Profile'
	})

	if len(components) == 0 {
		return nil
	}

	return []discordgo.MessageComponent{discordgo.ActionsRow{Components: components}}
}

func ProfileRefreshInteraction(authorID string, i *discordgo.Interaction) {

	i.Message.Embeds[0].Fields = generateProfileFields(authorID)
}
