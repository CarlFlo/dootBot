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

	// if there has been n hours since last time the user worked
	if time.Since(work.LastWorkedAt).Hours() < float64(config.CONFIG.Work.Cooldown) {

		message := fmt.Sprintf("You can only work once every %d hours.\nYou can work again <t:%d:R>", config.CONFIG.Work.Cooldown, work.LastWorkedAt.Add(time.Hour*6).Unix())
		s.ChannelMessageSend(m.ChannelID, message)
		// TODO: Make complex with componentes to user can buy tools
		return
	}

	// Reset streak if user hasn't worked in 24 hours
	if time.Since(work.LastWorkedAt).Hours() > 24 {
		work.Streak = 0
	}

	// TODO: Handle extra rewards for long streaks

	var user database.User
	database.DB.Table("Users").Where("discord_id = ?", m.Author.ID).First(&user)

	currentTime := time.Now()
	// Adds the cooldown hours
	currentTime = currentTime.Add(time.Hour * config.CONFIG.Work.Cooldown)

	// Updates the variables
	work.Streak += 1
	work.LastWorkedAt = time.Now()

	streakBonus := 1
	if work.Streak%6 == 0 {
		streakBonus = 2
	}
	moneyEarned := generateWorkIncome(&work, streakBonus)
	user.Money += uint64(moneyEarned)

	// TODO: Add ability to buy tools

	// Special message if user has a streak

	// Sends the message
	_, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("You performed some manual labour and earned some credits.\nYou earned **%d** credits.\nYou will be able to work again <t:%d:R>\nCurrent streak **%d**\n\nBuying additional tools will allow you to earn more money\n Each tool adds an extra income of %d credits", moneyEarned, currentTime.Unix(), work.Streak, config.CONFIG.Work.ToolBonus),
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{

					&discordgo.Button{
						Label:    "Buy Axe (500)",
						Disabled: false,
						CustomID: "buyAxe",
					},
					&discordgo.Button{
						Label:    "Buy Pickaxe (750)",
						Disabled: false,
						CustomID: "buyPickaxe",
					},
					&discordgo.Button{
						Label:    "Buy Shovel (850)",
						Disabled: false,
						CustomID: "buyShovel",
					},
					&discordgo.Button{
						Label:    "Buy Hammer (1000)",
						Disabled: false,
						CustomID: "buyHammer",
					},
				},
			},
		},
	})

	if err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}
	// Save the new streak, time and money to the user
	database.DB.Save(&user)
	database.DB.Save(&work)
}

func generateWorkIncome(work *database.Work, streakBonus int) int {
	// Generate a random int between config.CONFIG.Work.MinMoney and config.CONFIG.Work.MaxMoney
	moneyEarned := rand.Intn(config.CONFIG.Work.MaxMoney-config.CONFIG.Work.MinMoney) + config.CONFIG.Work.MinMoney

	moneyEarned *= streakBonus

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
