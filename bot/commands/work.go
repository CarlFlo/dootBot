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

	// Check if user can work
	var work database.Work

	//database.DB.Table("Works").Where("user_id = ?", m.Author.ID).First(&work)
	// https://gorm.io/docs/query.html#Joins
	if err := database.DB.Joins("JOIN Works ON Works.user_ID = Users.ID").
		Where("Users.discord_ID = ?", m.Author.ID).
		First(&work); err != nil {

		malm.Error("DB error: %s", err)
	}

	/* This code above does not work. Broken join */

	// if there has been 6 hours since last time the user worked
	if time.Since(work.LastUpdated).Hours() < 6 {

		message := fmt.Sprintf("You can only work once every 6 hours.\nYou'll have to wait <t:%d:R>", work.LastUpdated.Add(time.Hour*6).Unix())
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

	// convert to unix time
	untilYouCanWorkAgain := currentTime.Unix()

	// Generate a random int between config.CONFIG.Work.MinMoney and config.CONFIG.Work.MaxMoney
	rand.Seed(time.Now().UnixNano())
	moneyEarned := rand.Intn(config.CONFIG.Work.MaxMoney-config.CONFIG.Work.MinMoney) + config.CONFIG.Work.MinMoney

	// Saves the variables
	work.Streak += 1
	work.LastUpdated = currentTime
	user.Money += uint64(moneyEarned)

	_, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("You performed some manual labour and earned some money.\nYou earned **%d** money.\nYou will be able to work again <t:%d:R>\nCurrent streak **%d**\n\nBuying additional tools will allow you to earn more money", moneyEarned, untilYouCanWorkAgain, work.Streak),
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

	// Save the new streak, time and money to the user
	database.DB.Save(&work)
	database.DB.Save(&user)

}
