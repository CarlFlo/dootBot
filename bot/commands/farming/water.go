package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

/*
	The correct water reduce amount is not applied to the database when watering
*/

func farmWaterCrops(s *discordgo.Session, m *discordgo.MessageCreate) {

	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	var farm database.Farm
	farm.QueryUserFarmData(&user)

	// Check if user can water their plot
	if !config.CONFIG.Debug.IgnoreWaterCooldown && !farm.CanWater() {
		msg := fmt.Sprintf("You can't water your farm right now! You can water again %s", farm.CanWaterAt())
		s.ChannelMessageSend(m.ChannelID, msg)
		return
	}

	farm.QueryFarmPlots()
	if len(farm.Plots) == 0 {
		s.ChannelMessageSend(m.ChannelID, "You don't have any plots to water, plant a crop first!")
		return
	}

	// Check for perished crops
	preishedCrops := farm.CropsPerishedCheck()

	malm.Debug("Before watering")
	for i, plot := range farm.Plots {
		malm.Debug("%d - %v", i, plot.PlantedAt)
	}

	// Decrease the wait time for each crop on the users plots
	farm.WaterPlots()

	malm.Debug("After watering")
	for i, plot := range farm.Plots {
		malm.Debug("%d - %v", i, plot.PlantedAt)
	}

	message := "You watered your plots and reduced the growth time"

	if len(preishedCrops) > 0 {
		message += fmt.Sprintf("\nHowever, the following crops perished: %v!\nRemember to water your crops daily!", preishedCrops)
	}

	s.ChannelMessageSend(m.ChannelID, message)

	farm.Save()
}
