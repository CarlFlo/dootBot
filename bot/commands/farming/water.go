package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func farmWaterCrops(s *discordgo.Session, m *discordgo.MessageCreate) {

	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	var farm database.Farm
	farm.QueryUserFarmData(&user)

	var response string

	waterShared(&farm, &response, true)

	s.ChannelMessageSend(m.ChannelID, response)

	farm.Save()
}

func WaterInteraction(discordID string, response *string, s *discordgo.Session, me *discordgo.MessageEdit) {

	var user database.User
	user.QueryUserByDiscordID(discordID)

	var farm database.Farm
	farm.QueryUserFarmData(&user)

	waterShared(&farm, response, false)

	farm.Save()

	discordUser, err := s.User(discordID)
	if err != nil {
		malm.Error("Error getting user: %s", err)
	}

	// Update the message
	farm.UpdateInteractionOverview(discordUser, me)
}

func waterShared(farm *database.Farm, response *string, printSuccess bool) {

	// Check if user can water their plot
	if !config.CONFIG.Debug.IgnoreWaterCooldown && !farm.CanWater() {
		*response = fmt.Sprintf("You can't water your farm right now! You can water again %s", farm.CanWaterAt())
		return
	}

	farm.QueryFarmPlots()
	if len(farm.Plots) == 0 {
		*response = "You don't have any plots to water, plant a crop first!"
		return
	}

	perished := farm.Peek()

	// Decrease the wait time for each crop on the users plots
	farm.WaterPlots()

	if printSuccess {
		*response = "You watered your plots and reduced the growth time"
	}

	if perished {
		*response += "\nHowever, some crops perished!\nRemember to water your crops daily!"
	}

}
