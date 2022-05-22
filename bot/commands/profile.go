package commands

import (
	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
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

	complexMessage := &discordgo.MessageSend{
		Components: user.CreateProfileComponents(&work, &daily),
	}

	user.CreateProfileEmbeds(m.Author, &work, &daily, &complexMessage.Embeds)

	/*
		if components := user.CreateProfileComponents(&work, &daily); components != nil {
			complexMessage.Components = components
		}
	*/

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}
}

func ProfileRefreshInteraction(authorID string, author *discordgo.User, me *discordgo.MessageEdit) {

	var user database.User
	user.QueryUserByDiscordID(authorID)

	var work database.Work
	work.GetWorkInfo(&user)

	var daily database.Daily
	daily.GetDailyInfo(&user)

	ProfileUpdateMessageEdit(&user, &work, &daily, author, me)
}

func ProfileUpdateMessageEdit(user *database.User, work *database.Work, daily *database.Daily, author *discordgo.User, me *discordgo.MessageEdit) {
	user.CreateProfileEmbeds(author, work, daily, &me.Embeds)

	me.Components = user.CreateProfileComponents(work, daily)

	/*
		if components := user.CreateProfileComponents(work, daily); components != nil {
			me.Components = components
		}
	*/
}
