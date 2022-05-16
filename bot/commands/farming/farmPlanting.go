package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/bwmarrin/discordgo"
)

// Make prettier to match the style of the other messages

func farmPlant(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	if input.NumberOfArgsAre(1) {
		// only ,farm plant. Missing plant name. Give some help
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You need to specify which crop to plant. Use the command '%sfarm [c | crops]' to see a list of available crops.", config.CONFIG.BotPrefix))
		return
	}

	// Check for input (that a plant has been specified)
	if !input.NumberOfArgsAre(2) {
		return
	}

	cropName := input.GetArgs()[1]

	// Check if the user has enough money to buy seeds
	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	if !user.CanAfford(uint64(config.CONFIG.Farm.CropSeedPrice)) {
		s.ChannelMessageSend(m.ChannelID, "You don't have enough money to plant a seed!")
		return
	}

	// Parse the input plant (checks the database)
	var crop database.FarmCrop
	if ok := crop.GetCropByName(cropName); !ok {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("The crop '%s' is not valid!", cropName))
		return
	}

	// Check if the user has a free plot
	var farm database.Farm
	farm.QueryUserFarmData(&user)

	// Check if the user unlocked this crop
	if crop.ID > uint(farm.HighestPlantedCropIndex) {
		s.ChannelMessageSend(m.ChannelID, "You haven't unlocked this crop yet!\nYou have to plant all the previous crops at least once to unlock this one.")
		return
	}

	farm.QueryFarmPlots()

	if !farm.HasFreePlot() {
		s.ChannelMessageSend(m.ChannelID, "You don't have a free farm plot to plant in!")
		return
	}

	// Deduct the money from the user
	user.DeductMoney(uint64(config.CONFIG.Farm.CropSeedPrice))

	// Create a userFarmPlots entry with the data
	database.DB.Create(&database.FarmPlot{
		Farm: farm,
		Crop: crop,
	})

	message := fmt.Sprintf("The crop %s %s was planted!", crop.Emoji, crop.Name)

	// Increment the highestPlantedCropIndex
	if uint(farm.HighestPlantedCropIndex) == crop.ID {
		farm.HighestPlantedCropIndex++
		message = fmt.Sprintf("%s\n``You have unlocked a new crop!``", message)
	}

	// Send message to the user
	s.ChannelMessageSend(m.ChannelID, message)

	// Update the database
	user.Save()
	farm.Save()

}
