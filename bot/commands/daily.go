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

func Daily(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput) {

	var daily database.Daily

	database.DB.Raw("select * from dalies JOIN Users ON dalies.ID = Users.ID WHERE Users.discord_id = ?", m.Author.ID).First(&daily)

	// if there has been n hours since last time the user worked
	if time.Since(daily.LastDailyAt).Hours() < float64(config.CONFIG.Daily.Cooldown) {

		message := fmt.Sprintf("You can only receive your daily once every %d hours.\nYou can receive it again <t:%d:R>", config.CONFIG.Work.Cooldown, daily.LastDailyAt.Add(time.Hour*6).Unix())
		s.ChannelMessageSend(m.ChannelID, message)
		// TODO: Make complex with componentes to user can buy tools
		return
	}

	// Reset streak if user hasn't gotten their daily in 48
	if time.Since(daily.LastDailyAt).Hours() > 48 {
		daily.Streak = 0
	}

	// TODO: Handle extra rewards for long streaks

	var user database.User
	database.DB.Table("Users").Where("discord_id = ?", m.Author.ID).First(&user)

	currentTime := time.Now()
	// Adds the cooldown hours
	currentTime = currentTime.Add(time.Hour * config.CONFIG.Daily.Cooldown)

	// Updates the variables
	daily.Streak += 1
	daily.LastDailyAt = time.Now()

	streakBonus := 1
	if daily.Streak%7 == 0 {
		streakBonus = 2
	}

	moneyEarned := (rand.Intn(config.CONFIG.Daily.MaxMoney-config.CONFIG.Daily.MinMoney) + config.CONFIG.Daily.MinMoney) * streakBonus
	user.Money += uint64(moneyEarned)

	// TODO: Add ability to buy tools

	// Special message if user has a streak

	// Sends the message
	_, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content:    fmt.Sprintf("You got some free credits.\nYou earned **%d** credits.\nYou will be able to get some more again <t:%d:R>\nCurrent streak **%d**\n\nBuying additional tools will allow you to earn more money", moneyEarned, currentTime.Unix(), daily.Streak),
		Components: []discordgo.MessageComponent{},
	})

	if err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}
	// Save the new streak, time and money to the user
	database.DB.Save(&user)
	database.DB.Save(&daily)

}
