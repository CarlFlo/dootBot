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

// Debug - prints some debug information
func Work(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput) {

	var work database.Work

	database.DB.Raw("select * from Works JOIN Users ON Works.ID = Users.ID WHERE Users.discord_id = ?", m.Author.ID).First(&work)

	// Has there has been enough time since the last time the user worked?
	if !config.CONFIG.Debug.IgnoreWorkCooldown && time.Since(work.LastWorkedAt).Hours() < float64(config.CONFIG.Work.Cooldown) {
		triedToWorkTooEarly(s, m, &work)
		return
	}

	// Reset streak if user hasn't worked in the default 24 hours
	if time.Since(work.LastWorkedAt).Hours() > float64(config.CONFIG.Work.StreakResetHours) {
		work.ConsecutiveStreaks = 0
		work.Streak = 0
	}

	var user database.User
	database.DB.Table("Users").Where("discord_id = ?", m.Author.ID).First(&user)

	// Adds the cooldown
	nextWorkTime := time.Now().Add(time.Hour * config.CONFIG.Work.Cooldown)

	// Updates the variables
	work.ConsecutiveStreaks += 1
	work.Streak += 1
	work.LastWorkedAt = time.Now()

	// The StreakLength changed, so we need to update the streak for the player to avoid a crash
	if work.Streak > config.CONFIG.Work.StreakLength {
		work.Streak = config.CONFIG.Work.StreakLength
	}

	moneyEarned := generateWorkIncome(&work)
	user.Money += uint64(moneyEarned)

	// Tools tooltip
	toolsTooltip := generateToolTooltip(&work)

	description := fmt.Sprintf("%sYou earned ``%d`` %s!\nYou will be able to work again <t:%d:R>\nCurrent streak: ``%d``\n\n%s", config.CONFIG.Economy.Emoji, moneyEarned, config.CONFIG.Economy.Name, nextWorkTime.Unix(), work.ConsecutiveStreaks, toolsTooltip)

	extraRewardValue, percentage := generateStreakMessage(work.Streak, false)

	complexMessage := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			&discordgo.MessageEmbed{
				Type:        discordgo.EmbedTypeRich,
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
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: fmt.Sprintf("%s#%s", m.Author.AvatarURL("256"), m.Author.ID),
				},
			},
		},
	}

	if components := createButtonComponent(&work); components != nil {
		complexMessage.Components = components
	}

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! (Data not saved) %s", err)
		return
	}

	// Wrap around the streak
	work.Streak %= config.CONFIG.Work.StreakLength

	// Save the new streak, time and money to the user
	database.DB.Save(&user)
	database.DB.Save(&work)
}

func triedToWorkTooEarly(s *discordgo.Session, m *discordgo.MessageCreate, work *database.Work) {

	toolsTooltip := generateToolTooltip(work)

	description := fmt.Sprintf("You can work again <t:%d:R>\n\n%s", work.LastWorkedAt.Add(time.Hour*6).Unix(), toolsTooltip)

	extraRewardValue, percentage := generateStreakMessage(work.Streak, false)

	complexMessage := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
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
					Text: fmt.Sprintf("You can only work once every %d hours!", int(config.CONFIG.Work.Cooldown)),
				},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: fmt.Sprintf("%s#%s", m.Author.AvatarURL("256"), m.Author.ID),
				},
			},
		},
	}

	if components := createButtonComponent(work); components != nil {
		complexMessage.Components = components
	}

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
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

func generateStreakMessage(streak uint16, addStreakMessage bool) (string, string) {

	// Array of strings
	streakMessages := []string{
		":regional_indicator_b:", ":regional_indicator_o:", ":regional_indicator_n:", ":regional_indicator_u:", ":regional_indicator_s:"}

	percentage := float64(streak) / float64(config.CONFIG.Work.StreakLength)
	upTo := int(float64(len(streakMessages)) * percentage)

	// Append to a string values in streakMessages upto the index of upTo
	var visualStreakProgress string

	for i := 0; i < upTo; i++ {
		visualStreakProgress += fmt.Sprintf("%s ", streakMessages[i])
	}
	for i := upTo; i < len(streakMessages); i++ {
		visualStreakProgress += "- "
	}

	percentageText := fmt.Sprintf("%d%%", int(percentage*100))

	var streakMessage string
	if addStreakMessage && streak == config.CONFIG.Work.StreakLength {
		streakMessage = fmt.Sprintf("You earned an additional **%d** %s!", config.CONFIG.Work.StreakBonus, config.CONFIG.Economy.Name)
	}

	return fmt.Sprintf("%s %s", visualStreakProgress, streakMessage), percentageText
}

func generateWorkIncome(work *database.Work) int {

	// Generate a random int between config.CONFIG.Work.MinMoney and config.CONFIG.Work.MaxMoney
	moneyEarned := rand.Intn(config.CONFIG.Work.MaxMoney-config.CONFIG.Work.MinMoney) + config.CONFIG.Work.MinMoney

	// Adds the streak bonus to the amount
	if work.Streak == config.CONFIG.Work.StreakLength {
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
