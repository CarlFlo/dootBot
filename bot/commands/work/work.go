package work

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

func Work(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput) {

	var user database.User
	user.GetUserByDiscordID(m.Author.ID)

	var work database.Work
	work.GetWorkInfo(&user)

	// Reset streak if user hasn't worked in a specified amount of time (set in config)
	work.CheckStreak()

	complexMessage := &discordgo.MessageSend{}
	workMessageBuilder(complexMessage, m, &user, &work)

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! (Data not saved) %s", err)
		return
	}

	// Wrap around the streak
	work.Streak %= uint16(len(config.CONFIG.Work.StreakOutput))

	// Save the new streak, time and money to the user
	user.Save()
	work.Save()
}

func workMessageBuilder(msg *discordgo.MessageSend, m *discordgo.MessageCreate, user *database.User, work *database.Work) {

	toolsTooltip := generateToolTooltip(work)

	if work.CanDoWork() {

		// Calculates the cooldown
		nextWorkTime := time.Now().Add(time.Hour * config.CONFIG.Work.Cooldown)

		work.UpdateStreakAndTime()

		// Calculates the income
		moneyEarned := generateWorkIncome(work)
		user.Money += uint64(moneyEarned)

		moneyEarnedString := utils.HumanReadableNumber(moneyEarned)

		extraRewardValue, percentage := generateWorkStreakMessage(work.Streak, true)

		description := fmt.Sprintf("%sYou earned ``%s`` %s and your new balance is ``%s`` %s!\nYou will be able to work again <t:%d:R>\nCurrent streak: ``%d``\n\n%s", config.CONFIG.Emojis.Economy, moneyEarnedString, config.CONFIG.Economy.Name, user.PrettyPrintMoney(), config.CONFIG.Economy.Name, nextWorkTime.Unix(), work.ConsecutiveStreaks, toolsTooltip)

		msg.Embeds = []*discordgo.MessageEmbed{
			&discordgo.MessageEmbed{
				Type:        discordgo.EmbedTypeRich,
				Color:       config.CONFIG.Colors.Success,
				Title:       "Pay Check",
				Description: description,
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{
						Name:  fmt.Sprintf("Extra Reward Progress (%s)", percentage),
						Value: extraRewardValue,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Completing your streak will earn you an extra %d %s!\nThe streak resets after %d hours of inactivity.", config.CONFIG.Work.StreakBonus, config.CONFIG.Economy.Name, config.CONFIG.Work.StreakResetHours),
				},
			},
		}
	} else {

		description := fmt.Sprintf("You can work again <t:%d:R>\n\n%s", work.LastWorkedAt.Add(time.Hour*6).Unix(), toolsTooltip)
		extraRewardValue, percentage := generateWorkStreakMessage(work.Streak, false)

		msg.Embeds = []*discordgo.MessageEmbed{
			&discordgo.MessageEmbed{
				Type:        discordgo.EmbedTypeRich,
				Color:       config.CONFIG.Colors.Failure,
				Title:       fmt.Sprintf("%s Slow down!", config.CONFIG.Emojis.Failure),
				Description: description,
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{
						Name:  fmt.Sprintf("Extra Reward Progress (%s)", percentage),
						Value: extraRewardValue,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("You can only work once every %d hours!", int(config.CONFIG.Work.Cooldown)),
				},
			},
		}
	}

	msg.Embeds[0].Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: fmt.Sprintf("%s#%s", m.Author.AvatarURL("256"), m.Author.ID),
	}

	// Adds the button
	if components := createButtonComponent(work); components != nil {
		msg.Components = components
	}

}

func generateToolTooltip(work *database.Work) string {
	numOfBoughtTools := numberOfBoughtTools(work)
	if numOfBoughtTools > 0 {
		wordFormat := "tool"
		if numOfBoughtTools > 1 {
			wordFormat = "tools"
		}

		return fmt.Sprintf(":tools: You have %d %s, giving you an additional %d %s", numOfBoughtTools, wordFormat, numOfBoughtTools*config.CONFIG.Work.ToolBonus, config.CONFIG.Economy.Name)
	}

	return fmt.Sprintf("Buying additional tools will add an extra income of **%d** %s", config.CONFIG.Work.ToolBonus, config.CONFIG.Economy.Name)
}

func generateWorkStreakMessage(streak uint16, addStreakMessage bool) (string, string) {

	percentage := float64(streak) / float64(len(config.CONFIG.Work.StreakOutput))
	upTo := int(float64(len(config.CONFIG.Work.StreakOutput)) * percentage)

	// Append to a string values in config.CONFIG.Work.StreakOutput up to the index of upTo
	var visualStreakProgress string

	for i := 0; i < upTo; i++ {
		visualStreakProgress += fmt.Sprintf("%s ", config.CONFIG.Work.StreakOutput[i])
	}
	for i := upTo; i < len(config.CONFIG.Work.StreakOutput); i++ {
		visualStreakProgress += "- "
	}

	percentageText := fmt.Sprintf("%d%%", int(percentage*100))

	var streakMessage string
	if addStreakMessage && streak == uint16(len(config.CONFIG.Work.StreakOutput)) {
		streakMessage = fmt.Sprintf("An additional ``%s`` %s were added to your earnings!", utils.HumanReadableNumber(config.CONFIG.Work.StreakBonus), config.CONFIG.Economy.Name)
	}

	return fmt.Sprintf("%s %s", visualStreakProgress, streakMessage), percentageText
}

func generateWorkIncome(work *database.Work) int {

	// Generate a random int between config.CONFIG.Work.MinMoney and config.CONFIG.Work.MaxMoney
	moneyEarned := rand.Intn(config.CONFIG.Work.MaxMoney-config.CONFIG.Work.MinMoney) + config.CONFIG.Work.MinMoney

	// Adds the streak bonus to the amount
	if work.Streak == uint16(len(config.CONFIG.Work.StreakOutput)) {
		moneyEarned += config.CONFIG.Work.StreakBonus
	}

	moneyEarned += numberOfBoughtTools(work) * config.CONFIG.Work.ToolBonus

	return moneyEarned
}

func numberOfBoughtTools(work *database.Work) int {
	// Factor in the number of bought tools
	// Count the numbers of bits set in the variable work.Tools

	numBoughtTools := 0
	for i := 0; i < 8; i++ {
		if work.Tools&(1<<uint8(i)) != 0 {
			numBoughtTools++
		}
	}
	return numBoughtTools
}

func createButtonComponent(work *database.Work) []discordgo.MessageComponent {

	components := []discordgo.MessageComponent{}

	// Adds each tool present in the config file
	for i, v := range config.CONFIG.Work.Tools {
		if work.Tools&(1<<i) == 0 {
			components = append(components, &discordgo.Button{
				Label:    fmt.Sprintf("Buy %s (%d)", v.Name, v.Price),
				Disabled: false,
				CustomID: fmt.Sprintf("BWT-%s", v.Name), // 'BWT' is code for 'Buy Work Tool'
			})
		}
	}

	if len(components) == 0 {
		return nil
	}

	return []discordgo.MessageComponent{discordgo.ActionsRow{Components: components}}
}
