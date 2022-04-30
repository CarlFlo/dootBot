package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
)

// printFarm button component is turned off for now

func BuyFarmPlotInteraction(discordID string, response *string) bool {

	var user database.User
	user.QueryUserByDiscordID(discordID)

	// Valiadate again that the user have enough money
	if user.Money < uint64(config.CONFIG.Farm.FarmPlotPrice) {

		*response = fmt.Sprintf("You don't have enough money to buy a farm plot!\nYou have: %s %s", user.PrettyPrintMoney(), config.CONFIG.Economy.Name)
		return false
	}

	user.Money -= uint64(config.CONFIG.Farm.FarmPlotPrice)

	var farm database.Farm
	farm.QueryUserFarmData(&user)

	farm.OwnedPlots++

	user.Save()
	farm.Save()

	*response = "You successfully bought another plot!"

	return false
}
