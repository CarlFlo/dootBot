package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/bwmarrin/discordgo"
)

func farmWaterCrops(s *discordgo.Session, m *discordgo.MessageCreate) {

	var user database.User
	user.GetUserByDiscordID(m.Author.ID)

	var farm database.Farm
	farm.GetUserFarmData(&user)

	// Check if user can water their plot
	if !farm.CanWater() {
		msg := fmt.Sprintf("You can't water your farm right now! You can water again %s", farm.CanWaterAt())
		s.ChannelMessageSend(m.ChannelID, msg)
		return
	}

	farm.GetFarmPlots()
	if len(farm.Plots) == 0 {
		s.ChannelMessageSend(m.ChannelID, "You don't have any plots to water, plant a crop first!")
		return
	}

	// Check for perished crops
	preishedCrops := farm.CropsPerishedCheck()

	// Decrease the wait time for each crop on the users plots
	farm.WaterPlots()
	message := "You watered your plots!"

	if len(preishedCrops) > 0 {
		message += fmt.Sprintf("\nHowever, the following crops perished: %v!\nRemember to water your crops daily!", preishedCrops)
	}

	s.ChannelMessageSend(m.ChannelID, message)

	farm.Save()
}
