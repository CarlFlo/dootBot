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

	noWaitCooldown := true

	var work database.Work

	database.DB.Raw("select * from Works JOIN Users ON Works.ID = Users.ID WHERE Users.discord_id = ?", m.Author.ID).First(&work)

	// Has there has been enough time since the last time the user worked?
	if !noWaitCooldown && time.Since(work.LastWorkedAt).Hours() < float64(config.CONFIG.Work.Cooldown) {

		message := fmt.Sprintf("You can only work once every %d hours.\nYou can work again <t:%d:R>", config.CONFIG.Work.Cooldown, work.LastWorkedAt.Add(time.Hour*6).Unix())
		s.ChannelMessageSend(m.ChannelID, message)
		// TODO: Make complex with componentes to user can buy tools
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
	currentTime := time.Now().Add(time.Hour * config.CONFIG.Work.Cooldown)

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

	// TODO: Add ability to buy tools

	complexMessage := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			&discordgo.MessageEmbed{
				Type:        discordgo.EmbedTypeRich,
				Title:       "Pay Check",
				Description: fmt.Sprintf(":coin:**%d** credits were deposited into your account!\nYou will be able to work again <t:%d:R>\nCurrent streak: **%d** (%d)\n\nBuying additional tools will add an extra income of **%d** credits", moneyEarned, currentTime.Unix(), work.ConsecutiveStreaks, work.Streak, config.CONFIG.Work.ToolBonus),
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{
						Name:  "Extra Reward",
						Value: generateStreakMessage(work.Streak),
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: fmt.Sprintf("Completing your streak will earn you an extra %d credits!\nThe streak resets after %d hours of inactivity", config.CONFIG.Work.StreakBonus, config.CONFIG.Work.StreakResetHours),
				},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: m.Author.AvatarURL("256"),
				},
			},
		},
	}

	if components := createButtonComponent(&work); components != nil {
		complexMessage.Components = components
	}

	// Sends the message
	_, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage)

	if err != nil {
		malm.Error("Could not send message! (Data not saved) %s", err)
		return
	}

	// Wrap around the streak
	work.Streak %= config.CONFIG.Work.StreakLength

	// Save the new streak, time and money to the user
	database.DB.Save(&user)
	database.DB.Save(&work)
}

func generateStreakMessage(streak uint16) string {

	// Array of strings
	streakMessages := []string{
		":regional_indicator_b:", ":regional_indicator_o:", ":regional_indicator_n:", ":regional_indicator_u:", ":regional_indicator_s:"}

	percentage := float64(streak) / float64(config.CONFIG.Work.StreakLength)
	upTo := int(float64(len(streakMessages)) * percentage)

	// Append to a string values in streakMessages upto the index of upTo
	var message string

	for i := 0; i < upTo; i++ {
		message += fmt.Sprintf("%s ", streakMessages[i])
	}
	for i := upTo; i < len(streakMessages); i++ {
		message += "- "
	}
	var streakMessage string
	if streak == config.CONFIG.Work.StreakLength {
		streakMessage = fmt.Sprintf("You earned an additional **%d** credits!", config.CONFIG.Work.StreakBonus)
	}
	return fmt.Sprintf("%s(%d%%) %s", message, int(percentage*100), streakMessage)
}

func generateWorkIncome(work *database.Work) int {

	// Generate a random int between config.CONFIG.Work.MinMoney and config.CONFIG.Work.MaxMoney
	moneyEarned := rand.Intn(config.CONFIG.Work.MaxMoney-config.CONFIG.Work.MinMoney) + config.CONFIG.Work.MinMoney

	// Adds the streak bonus to the amount
	if work.Streak == config.CONFIG.Work.StreakLength {
		moneyEarned += config.CONFIG.Work.StreakBonus
	}

	// Factor in the numBoughtTools
	// Count the numbers of bits set in the variable work.Tools
	numBoughtTools := 0
	for i := 0; i < 8; i++ {
		if work.Tools&(1<<uint8(i)) != 0 {
			numBoughtTools++
		}
	}

	moneyEarned += numBoughtTools * config.CONFIG.Work.ToolBonus

	return moneyEarned
}

func createButtonComponent(work *database.Work) []discordgo.MessageComponent {

	components := []discordgo.MessageComponent{}

	// Adds each tool present in the config file
	for i, v := range config.CONFIG.Work.Tools {
		if work.Tools&(1<<i) == 0 {
			components = append(components, &discordgo.Button{
				Label:    fmt.Sprintf("Buy %s (%d)", v.Name, v.Price),
				Disabled: false,
				CustomID: fmt.Sprintf("buy%s", v.Name),
			})
		}
	}

	if len(components) == 0 {
		return nil
	}

	return []discordgo.MessageComponent{discordgo.ActionsRow{Components: components}}
}
