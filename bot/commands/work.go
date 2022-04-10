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

	// if there has been 6 hours since last time the user worked
	if time.Since(work.LastUpdated).Hours() < 6 {

		message := fmt.Sprintf("You can only work once every %d hours.\nYou can work again <t:%d:R>", config.CONFIG.Work.WorkCooldown, work.LastUpdated.Add(time.Hour*6).Unix())
		s.ChannelMessageSend(m.ChannelID, message)
		// TODO: Make complex with componentes to user can buy tools
		return
	}

	// Reset streak if user hasnt worked in 24 hours
	if time.Since(work.LastUpdated).Hours() > 24 {
		work.Streak = 0
	}

	var user database.User
	database.DB.Table("Users").Where("discord_id = ?", m.Author.ID).First(&user)

	// Get the current time
	currentTime := time.Now()
	// Add six hours
	currentTime = currentTime.Add(time.Hour * 6)

	moneyEarned := generateWorkIncome(&work)

	// Updates the variables
	work.Streak += 1
	work.LastUpdated = time.Now()
	user.Money += uint64(moneyEarned)

	generateResponse(s, m, &work, moneyEarned, &currentTime)

	// Save the new streak, time and money to the user
	database.DB.Save(&user)
	database.DB.Save(&work)
}

func generateResponse(s *discordgo.Session, m *discordgo.MessageCreate, work *database.Work, moneyEarned int, currentTime *time.Time) {
	// Sends the message
	_, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("You performed some manual labour and earned some credits.\nYou earned **%d** credits.\nYou will be able to work again <t:%d:R>\nCurrent streak **%d**\n\nBuying additional tools will allow you to earn more money\n Each tool adds an extra income of %d credits", moneyEarned, currentTime.Unix(), work.Streak, config.CONFIG.Work.ToolBonus),
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{

					&discordgo.Button{
						Label:    "Buy Axe",
						Disabled: false,
						CustomID: "buyAxe",
					},
					&discordgo.Button{
						Label:    "Buy Pickaxe",
						Disabled: false,
						CustomID: "buyPickaxe",
					},
					&discordgo.Button{
						Label:    "Buy Shovel",
						Disabled: false,
						CustomID: "buyShovel",
					},
					&discordgo.Button{
						Label:    "Buy Hammer",
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
}

func generateWorkIncome(work *database.Work) int {
	// Generate a random int between config.CONFIG.Work.MinMoney and config.CONFIG.Work.MaxMoney
	moneyEarned := rand.Intn(config.CONFIG.Work.MaxMoney-config.CONFIG.Work.MinMoney) + config.CONFIG.Work.MinMoney

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
