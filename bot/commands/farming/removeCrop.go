package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/bwmarrin/discordgo"
)

// Removes all crops
func farmRemoveCrops(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	// Check for input (that a plant has been specified)
	if !input.NumberOfArgsAre(2) {
		return
	}

	cropName := input.GetArgs()[1]

	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	var farm database.Farm
	defer farm.Save()

	farm.QueryUserFarmData(&user)
	farm.QueryFarmPlots()

	for _, plot := range farm.Plots {
		plot.QueryCropInfo()
		if plot.Crop.Name == cropName {

			farm.DeletePlot(plot)
			break
		}
	}

	// Send message to the user
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Remove command executed for the crop: '%s'", cropName))
}
