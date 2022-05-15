package work

import (
	"fmt"
	"regexp"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/bwmarrin/discordgo"
)

func BuyToolInteraction(authorID string, response *string, btnData *[]structs.ButtonData, i *discordgo.Interaction) {

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
		*response = fmt.Sprintf("You have reached the maximum number of tools you can buy! Max %d",
			config.CONFIG.Work.MaxTools)
		return
	}

	user.DeductMoney(uint64(price))

	work.Tools += 1

	// Update the message as well to reflect that a new tool was bought.
	patternString := fmt.Sprintf(`%s .+ \d+ tool.+`, config.CONFIG.Emojis.Tools)

	pattern := regexp.MustCompile(patternString)
	modifiedMsg := pattern.ReplaceAllString(i.Message.Embeds[0].Description, generateToolTooltip(&work))

	i.Message.Embeds[0].Description = modifiedMsg

	// Calculate new cost
	_, newPriceStr := work.CalcBuyToolPrice()

	*btnData = append(*btnData, structs.ButtonData{
		CustomID: "BWT",
		Disabled: work.HasHitMaxToolLimit(),
		Label:    fmt.Sprintf("Buy Tool (%s)", newPriceStr),
	})

	user.Save()
	work.Save()
}

func DoWorkInteraction(authorID string, response *string, btnData *[]structs.ButtonData) {

}
