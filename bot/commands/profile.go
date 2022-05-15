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

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Color:       config.CONFIG.Colors.Neutral,
			Title:       fmt.Sprintf("%s#%s profile", m.Author.Username, m.Author.Discriminator),
			Description: "",
			Fields:      generateProfileFields(&user, &work, &daily),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Profile footer",
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("%s#%s", m.Author.AvatarURL("256"), m.Author.ID),
			},
		},
	}}

	if components := createButtonComponent(&work, &daily); components != nil {
		complexMessage.Components = components
	}

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}
}

func generateProfileFields(user *database.User, work *database.Work, daily *database.Daily) []*discordgo.MessageEmbedField {

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

func createButtonComponent(work *database.Work, daily *database.Daily) []discordgo.MessageComponent {

	components := []discordgo.MessageComponent{}

	if daily.CanDoDaily() {
		// Only create if the daily can be done
		components = append(components, &discordgo.Button{
			Label:    "Collect Daily",
			Style:    1, // Default purple
			Disabled: false,
			CustomID: "PD", // 'PD' is code for 'Profile Daily'
		})
	}
	if work.CanDoWork() {
		// Only create if the work can be done
		components = append(components, &discordgo.Button{
			Label:    "Work",
			Style:    1, // Default purple
			Disabled: false,
			CustomID: "PW", // 'PW' is code for 'Profile Work'
		})
	}

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

	var user database.User
	user.QueryUserByDiscordID(authorID)

	var work database.Work
	work.GetWorkInfo(&user)

	var daily database.Daily
	daily.GetDailyInfo(&user)

	i.Message.Embeds[0].Fields = generateProfileFields(&user, &work, &daily)
	// Also update buttons

	//i.Message.Components = createButtonComponent(&work, &daily) // Causes crash

}
