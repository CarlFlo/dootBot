package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
)

// printFarm button component is turned off for now
// Implement limit on how many plots a user can own

func BuyFarmPlotInteraction(discordID string, response *string, disableButton *bool) {

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

	user.Save()
	farm.Save()

	*response = "You successfully bought another plot!"
}
