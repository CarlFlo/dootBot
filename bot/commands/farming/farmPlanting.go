package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/bwmarrin/discordgo"
)

func farmPlant(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	// Check for input (that a plant has been specified)
	if !input.NumberOfArgsAre(2) {
		return
	}

	cropName := input.GetArgs()[1]

	// Check if the user has enough money to buy seeds
	var user database.User
	user.GetUserByDiscordID(m.Author.ID)

	if uint64(config.CONFIG.Farm.CropSeedPrice) > user.Money {
		s.ChannelMessageSend(m.ChannelID, "You don't have enough money to plant a seed!")
		return
	}

	// Check if the user has a free plot
	var farm database.Farm
	farm.GetUserFarmData(&user)
	farm.GetFarmPlots()

	if !farm.HasFreePlot() {
		s.ChannelMessageSend(m.ChannelID, "You don't have a free farm plot to plant in!")
		return
	}

	// Parse the input plant (checks the database)
	var crop database.FarmCrop
	if ok := crop.GetCropByName(cropName); !ok {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("The crop '%s' is not valid!", cropName))
		return
	}

	// Deduct the money from the user
	user.Money -= uint64(config.CONFIG.Farm.CropSeedPrice)

	// Create a userFarmPlots entry with the data
	database.DB.Create(&database.FarmPlot{
		Farm: farm,
		Crop: crop,
	})

	// Send message to the user
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("The crop %s %s was planted!", crop.Emoji, crop.Name))

	// Update the database
	user.Save()
	farm.Save()

}
