package farming

import (
	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/bwmarrin/discordgo"
)

// Redo this better later...

// First index is the help, the rest is the commands
var farmCommands = [][]string{
	{"Plant a crop", "p", "plant"},
	{"Get info about available crops", "c", "crop", "crops"},
	{"Get help on farming", "help"},
	{"Water your crops", "w", "water"},
	{"Harvest your crops", "h", "harvest"},
	{"Removes 1 crop with the matching name from the farm", "rm", "remove"},
}

func Farming(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	// Handle farm arguments

	if input.ArgsContains(farmCommands[0][1:]) {
		// User wants to plant some seeds
		farmPlant(s, m, input)
		return
	} else if input.ArgsContains(farmCommands[1][1:]) {
		// User wants info about crops/seeds
		farmCrops(s, m)
		return
	} else if input.ArgsContains(farmCommands[2][1:]) {
		farmHelp(s, m)
		return
	} else if input.ArgsContains(farmCommands[3][1:]) {
		// Water the crops
		farmWaterCrops(s, m)
		return
	} else if input.ArgsContains(farmCommands[4][1:]) {
		// Harvest the crops
		farmHarvestCrops(s, m)
		return
	} else if input.ArgsContains(farmCommands[5][1:]) {
		// Removes a crop
		farmRemoveCrops(s, m, input)
		return
	}

	printFarm(s, m, input)
}
