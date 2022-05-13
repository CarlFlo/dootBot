package daily

import (
	"fmt"
	"math/rand"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func Daily(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	var daily database.Daily
	daily.GetDailyInfo(&user)

	// Reset streak if user hasn't done their daily in a specified amount of time (set in config)
	daily.CheckStreak()

	complexMessage := &discordgo.MessageSend{}
	dailyMessageBuilder(complexMessage, m, &user, &daily)

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

func dailyMessageBuilder(msg *discordgo.MessageSend, m *discordgo.MessageCreate, user *database.User, daily *database.Daily) {

	if daily.CanDoDaily() {

		daily.UpdateStreakAndTime()

		// Calculates the income
		moneyEarned := generateDailyIncome(daily)
		user.Money += uint64(moneyEarned)

		moneyEarnedString := utils.HumanReadableNumber(moneyEarned)

		extraRewardValue, percentage := generateDailyStreakMessage(daily.Streak, true)

		description := fmt.Sprintf("%sYou earned ``%s`` %s! Your new balance is ``%s`` %s!\nYou will be able to get your daily again %s\nCurrent streak: ``%d``",
			config.CONFIG.Emojis.Economy,
			moneyEarnedString,
			config.CONFIG.Economy.Name,
			user.PrettyPrintMoney(),
			config.CONFIG.Economy.Name,
			daily.CanDoDailyAt(),
			daily.ConsecutiveStreaks)

		msg.Embeds = []*discordgo.MessageEmbed{
			{
				Type:        discordgo.EmbedTypeRich,
				Color:       config.CONFIG.Colors.Success,
				Title:       "Daily Bonus",
				Description: description,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  fmt.Sprintf("Extra Reward Progress (%s)", percentage),
						Value: extraRewardValue,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Completing your streak will earn you an extra %d %s!\nThe streak resets after %d hours of inactivity.",
						config.CONFIG.Daily.StreakBonus,
						config.CONFIG.Economy.Name,
						config.CONFIG.Daily.StreakResetHours),
				},
			},
		}
	} else {

		description := fmt.Sprintf("You can get your next daily again %s", daily.CanDoDailyAt())
		extraRewardValue, percentage := generateDailyStreakMessage(daily.Streak, false)

		msg.Embeds = []*discordgo.MessageEmbed{
			{
				Type:        discordgo.EmbedTypeRich,
				Color:       config.CONFIG.Colors.Failure,
				Title:       fmt.Sprintf("%s Slow down!", config.CONFIG.Emojis.Failure),
				Description: description,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  fmt.Sprintf("Extra Reward Progress (%s)", percentage),
						Value: extraRewardValue,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("You can get your daily once every %d hours!", int(config.CONFIG.Daily.Cooldown)),
				},
			},
		}
	}

	msg.Embeds[0].Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: fmt.Sprintf("%s#%s", m.Author.AvatarURL("256"), m.Author.ID),
	}
}

func generateDailyStreakMessage(streak uint16, addStreakMessage bool) (string, string) {

	percentage := float64(streak) / float64(len(config.CONFIG.Daily.StreakOutput))
	upTo := int(float64(len(config.CONFIG.Daily.StreakOutput)) * percentage)

	// Append to a string values in config.CONFIG.Daily.StreakOutput up to the index of upTo
	var visualStreakProgress string

	for i := 0; i < upTo; i++ {
		visualStreakProgress += fmt.Sprintf("%s ", config.CONFIG.Daily.StreakOutput[i])
	}
	for i := upTo; i < len(config.CONFIG.Daily.StreakOutput); i++ {
		visualStreakProgress += "- "
	}

	percentageText := fmt.Sprintf("%d%%", int(percentage*100))

	var streakMessage string
	if addStreakMessage && streak == uint16(len(config.CONFIG.Daily.StreakOutput)) {
		streakMessage = fmt.Sprintf("An additional ``%s`` %s were added to your daily earnings!", utils.HumanReadableNumber(config.CONFIG.Daily.StreakBonus), config.CONFIG.Economy.Name)
	}

	return fmt.Sprintf("%s %s", visualStreakProgress, streakMessage), percentageText
}

func generateDailyIncome(daily *database.Daily) int {

	// Generate a random int between config.CONFIG.Daily.MinMoney and config.CONFIG.Daily.MaxMoney
	moneyEarned := rand.Intn(config.CONFIG.Daily.MaxMoney-config.CONFIG.Daily.MinMoney) + config.CONFIG.Daily.MinMoney

	// Adds the streak bonus to the amount
	if daily.Streak == uint16(len(config.CONFIG.Daily.StreakOutput)) {
		moneyEarned += config.CONFIG.Daily.StreakBonus
	}

	return moneyEarned
}
