package farming

import (
	"fmt"
	"time"

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

	// Check if the user have any planted crops to water?

	// Update last watered at
	farm.LastWatered = time.Now()

	// Decrease the wait time for each crop on the users plots
	farm.GetFarmPlots()
	farm.WaterPlots()

	farm.Save()
}
