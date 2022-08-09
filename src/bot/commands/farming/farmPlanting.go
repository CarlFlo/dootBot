package farming

import (
	"fmt"

	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/database"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

// Runs when the farm plant <crop> is run
func farmPlant(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	if input.NumberOfArgsAre(1) {
		// only ,farm plant. Missing plant name. Give some help
		utils.SendMessageFailure(m, fmt.Sprintf("You need to specify which crop to plant. Use the command '%sfarm [c | crops]' to see a list of available crops.", config.CONFIG.BotPrefix))
		return
	}

	// Check for input (that a plant has been specified)
	if !input.NumberOfArgsAre(2) {
		return
	}

	cropName := input.GetArgsLowercase()[1]

	// Check if the user has enough money to buy seeds
	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	var farm database.Farm
	farm.QueryUserFarmData(&user)

	var response string
	ok := farmPlantShared(&user, &farm, cropName, &response, true)

	// Send message to the user
	if ok {
		utils.SendMessageSuccess(m, response)
	} else {
		utils.SendMessageFailure(m, response)
	}

	// Update the database
	user.Save()
	farm.Save()

}

func FarmPlantInteraction(discordID string, response *string, i *discordgo.Interaction, s *discordgo.Session, me *discordgo.MessageEdit) {

	cropName := i.Data.(discordgo.MessageComponentInteractionData).Values[0]

	var user database.User
	user.QueryUserByDiscordID(discordID)

	var farm database.Farm
	farm.QueryUserFarmData(&user)

	farmPlantShared(&user, &farm, cropName, response, false)

	discordUser, err := s.User(discordID)
	if err != nil {
		malm.Error("Error getting user: %s", err)
	}

	// Update the message
	farm.UpdateInteractionOverview(discordUser, me)

	user.Save()
	farm.Save()
}

// farmPlantShared is the shared code for planting crops
// Returns true if success, else false
func farmPlantShared(user *database.User, farm *database.Farm, cropName string, response *string, outputCrop bool) bool {

	if !user.CanAfford(uint64(config.CONFIG.Farm.CropSeedPrice)) {
		*response = "You don't have enough money to plant a seed!"
		return false
	}

	var crop database.FarmCrop
	if ok := crop.GetCropByName(cropName); !ok {
		*response = fmt.Sprintf("The crop '%s' is not valid!", cropName)
		return false
	}

	// Check if the user have unlocked the crop
	if !farm.HasUserUnlocked(&crop) {
		*response = "You have not unlocked this crop!"
		return false
	}

	if !farm.HasFreePlot() {
		*response = "You don't have a free farm plot to plant in!"
		return false
	}

	user.DeductMoney(uint64(config.CONFIG.Farm.CropSeedPrice))

	fp := &database.FarmPlot{
		Farm: *farm,
		Crop: crop,
	}

	// Create a userFarmPlots entry with the data
	database.DB.Create(fp)

	// This is to ensure that the crop wont instantly perish once planted if they user haven't watered in a while
	if farm.MissedWaterDeadline() {
		farm.ResetLastWatered()
	}

	if outputCrop {
		*response = fmt.Sprintf("The crop %s %s was planted!", crop.Emoji, crop.Name)
	}

	if uint(farm.HighestPlantedCropIndex) == crop.ID {
		farm.HighestPlantedCropIndex++
		*response += "\n``You have unlocked a new crop!``"
	}
	return true
}
