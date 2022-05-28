package daily

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/src/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/src/config"
	"github.com/CarlFlo/DiscordMoneyBot/src/database"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func Daily(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	var daily database.Daily
	daily.GetDailyInfo(&user)

	ok, earnedMoney, streakReward, streakPercentage, titleText, footerText := daily.DoDaily(&user)

	var description string
	var color int

	if ok {
		color = config.CONFIG.Colors.Success
		description = fmt.Sprintf("%sYou earned ``%s`` %s! Your new balance is ``%s`` %s!\nYou will be able to get your daily again %s\nCurrent streak: ``%d``",
			config.CONFIG.Emojis.Economy,
			earnedMoney,
			config.CONFIG.Economy.Name,
			user.PrettyPrintMoney(),
			config.CONFIG.Economy.Name,
			daily.CanDoDailyAt(),
			daily.ConsecutiveStreaks)
	} else {
		color = config.CONFIG.Colors.Failure
		description = fmt.Sprintf("You can get your next daily again %s", daily.CanDoDailyAt())
	}

	complexMessage := &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Color:       color,
			Title:       titleText,
			Description: description,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  fmt.Sprintf("Extra Reward Progress (%s)", streakPercentage),
					Value: streakReward,
				},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("%s#%s", m.Author.AvatarURL("256"), m.Author.ID),
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: footerText,
			},
		},
	}

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! (Data not saved) %s", err)
		return
	}

	// Wrap around the streak
	daily.Streak %= uint16(len(config.CONFIG.Daily.StreakOutput))

	user.Save()
	daily.Save()
}
