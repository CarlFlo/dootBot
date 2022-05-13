package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"github.com/bwmarrin/discordgo"
)

// printFarm button component is turned off for now
// Implement limit on how many plots a user can own

func BuyFarmPlotInteraction(discordID string, response *string, bdm *utils.ButtonDataWrapper, i *discordgo.Interaction) {

	var user database.User
	user.QueryUserByDiscordID(discordID)

	var farm database.Farm
	farm.QueryUserFarmData(&user)

	cost := farm.CalcFarmPlotPrice()

	// Valiadate again that the user have enough money
	if user.Money < uint64(cost) {
		*response = fmt.Sprintf("You don't have enough money to buy a farm plot!\nYou have: %s %s", user.PrettyPrintMoney(), config.CONFIG.Economy.Name)
		return
	}

	user.Money -= uint64(cost)

	farm.OwnedPlots++

	i.Message.Embeds[0].Description = farm.CreateEmbedDescription()
	i.Message.Embeds[0].Fields = farm.CreateEmbedFields()

	user.Save()
	farm.Save()

	//*response = "You successfully bought another plot!"

	bdm.ButtonData = append(bdm.ButtonData, utils.ButtonData{
		CustomID: "BFP",
		Disabled: false,
		Label:    fmt.Sprintf("Buy Farm Plot (%s)", utils.HumanReadableNumber(farm.CalcFarmPlotPrice())),
	})
}
