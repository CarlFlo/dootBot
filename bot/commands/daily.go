package commands

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

/*
	Fix: streak bonus not paying out
*/

func Daily(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput) {

	var daily database.Daily
	var user database.User

	daily.GetDailyByDiscordID(m.Author.ID)
	user.GetUserByDiscordID(m.Author.ID)

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

	// Save the new streak, time and money to the user
	database.DB.Save(&user)
	database.DB.Save(&daily)
}

func dailyMessageBuilder(msg *discordgo.MessageSend, m *discordgo.MessageCreate, user *database.User, daily *database.Daily) {

	if daily.CanDoDaily() {

		// Calculates the cooldown
		nextDailyTime := time.Now().Add(time.Hour * config.CONFIG.Daily.Cooldown)

		daily.UpdateStreakAndTime()

		// Calculates the income
		moneyEarned := generateDailyIncome(daily)
		user.Money += uint64(moneyEarned)

		extraRewardValue, percentage := generateDailyStreakMessage(daily.Streak, true)

		description := fmt.Sprintf("%sYou earned ``%d`` %s and your new balance is ``%d`` !\nYou will be able to get your daily again <t:%d:R>\nCurrent streak: ``%d``", config.CONFIG.Economy.Emoji, moneyEarned, config.CONFIG.Economy.Name, user.Money, nextDailyTime.Unix(), daily.ConsecutiveStreaks)

		msg.Embeds = []*discordgo.MessageEmbed{
			&discordgo.MessageEmbed{
				Type:        discordgo.EmbedTypeRich,
				Title:       "Daily Bonus",
				Description: description,
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{
						Name:  fmt.Sprintf("Extra Reward Progress (%s)", percentage),
						Value: extraRewardValue,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Completing your streak will earn you an extra %d %s!\nThe streak resets after %d hours of inactivity.", config.CONFIG.Daily.StreakBonus, config.CONFIG.Economy.Name, config.CONFIG.Daily.StreakResetHours),
				},
			},
		}
	} else {

		description := fmt.Sprintf("You can get your next daily again <t:%d:R>", daily.LastDailyAt.Add(time.Hour*6).Unix())
		extraRewardValue, percentage := generateDailyStreakMessage(daily.Streak, false)

		msg.Embeds = []*discordgo.MessageEmbed{
			&discordgo.MessageEmbed{
				Type:        discordgo.EmbedTypeRich,
				Title:       ":x: Slow down!",
				Description: description,
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{
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
		streakMessage = fmt.Sprintf("An additional ``%d`` %s were added to your daily earnings!", config.CONFIG.Daily.StreakBonus, config.CONFIG.Economy.Name)
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
