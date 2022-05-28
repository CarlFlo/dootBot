package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/src/config"
	"github.com/CarlFlo/DiscordMoneyBot/src/database"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

// printFarm button component is turned off for now
// Implement limit on how many plots a user can own

func BuyFarmPlotInteraction(discordID string, response *string, s *discordgo.Session, me *discordgo.MessageEdit) {

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

	if farm.HasMaxAmountOfPlots() {
		*response = fmt.Sprintf("You already have the maximum amount of farm plots!\nYou can only own %d farm plots", config.CONFIG.Farm.MaxPlots)
		return
	}

	user.DeductMoney(uint64(cost))

	farm.OwnedPlots++

	discordUser, err := s.User(discordID)
	if err != nil {
		malm.Error("Error getting user: %s", err)
	}

	farm.UpdateInteractionOverview(discordUser, me)

	user.Save()
	farm.Save()
}
