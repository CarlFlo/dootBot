package daily

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/src/bot/commands"
	"github.com/CarlFlo/DiscordMoneyBot/src/config"
	"github.com/CarlFlo/DiscordMoneyBot/src/database"
	"github.com/bwmarrin/discordgo"
)

func DoDailyInteraction(authorID string, response *string, author *discordgo.User, me *discordgo.MessageEdit) {

	var user database.User
	user.QueryUserByDiscordID(authorID)

	var work database.Work
	work.GetWorkInfo(&user)

	var daily database.Daily
	daily.GetDailyInfo(&user)

	ok, earnedMoney, streakReward, streakPercentage, titleText, _ := daily.DoDaily(&user)

	if ok {
		*response = fmt.Sprintf("%s!\n%sYou earned ``%s`` %s! Your new balance is ``%s`` %s!\nYou will be able to get your daily again %s\nCurrent streak: ``%d``\n\nExtra Reward Progress (%s)\n%s",
			titleText,
			config.CONFIG.Emojis.Economy,
			earnedMoney,
			config.CONFIG.Economy.Name,
			user.PrettyPrintMoney(),
			config.CONFIG.Economy.Name,
			daily.CanDoDailyAt(),
			daily.ConsecutiveStreaks,
			streakPercentage,
			streakReward)
	} else {
		*response = fmt.Sprintf("%s!\nYou can get your next daily again %s",
			titleText,
			daily.CanDoDailyAt())
	}

	commands.ProfileUpdateMessageEdit(&user, &work, &daily, author, me)

	user.Save()
	daily.Save()
}
