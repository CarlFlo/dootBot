package work

import (
	"fmt"
	"regexp"

	"github.com/CarlFlo/DiscordMoneyBot/bot/commands"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/bwmarrin/discordgo"
)

func BuyToolInteraction(authorID string, response *string, i *discordgo.Interaction, me *discordgo.MessageEdit) {

	// Check if the user has enough money
	var user database.User
	user.QueryUserByDiscordID(authorID)

	var work database.Work
	work.GetWorkInfo(&user)

	price, _ := work.CalcBuyToolPrice()

	if uint64(price) > user.Money {
		difference := uint64(price) - user.Money
		*response = fmt.Sprintf("You are lacking ``%d`` %s for this transaction.\nYour balance: ``%d`` %s", difference, config.CONFIG.Economy.Name, user.Money, config.CONFIG.Economy.Name)
		return
	}

	if work.HasHitMaxToolLimit() {
		*response = fmt.Sprintf("You have reached the maximum number of tools you can buy! Max %d", config.CONFIG.Work.MaxTools)
		return
	}

	user.DeductMoney(uint64(price))

	work.Tools += 1

	// Update the message as well to reflect that a new tool was bought.
	patternString := fmt.Sprintf(`%s .+ \d+ tool.+`, config.CONFIG.Emojis.Tools)

	pattern := regexp.MustCompile(patternString)
	modifiedMsg := pattern.ReplaceAllString(i.Message.Embeds[0].Description, generateToolTooltip(&work))

	me.Embeds = append(me.Embeds, &discordgo.MessageEmbed{
		Title:       i.Message.Embeds[0].Title,
		Description: modifiedMsg,
		Color:       i.Message.Embeds[0].Color,
		Fields:      i.Message.Embeds[0].Fields,
		Footer:      i.Message.Embeds[0].Footer,
		Thumbnail:   i.Message.Embeds[0].Thumbnail,
	})

	me.Components = work.CreateMessageComponents()
	/*
		if components := work.CreateMessageComponents(); components != nil {
			me.Components = components
		}
	*/

	user.Save()
	work.Save()
}

// From the profile message
func DoWorkInteraction(authorID string, response *string, author *discordgo.User, me *discordgo.MessageEdit) {

	var user database.User
	user.QueryUserByDiscordID(authorID)

	var work database.Work
	work.GetWorkInfo(&user)

	canDoWork := work.CanDoWork()

	work.StreakPreMsgAction()

	*response = createWorkMessageDescription(&user, &work, canDoWork)

	work.StreakPostMsgAction()

	user.Save()
	work.Save()

	var daily database.Daily
	daily.GetDailyInfo(&user)

	commands.ProfileUpdateMessageEdit(&user, &work, &daily, author, me)
}
