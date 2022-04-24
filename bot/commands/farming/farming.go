package farming

import (
	"fmt"
	"strings"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

// First index is the help, the rest is the commands
var farmCommands = [][]string{
	{"Plant a crop", "p", "plant"},
	{"Get info about available crops", "c", "crop", "crops"},
	{"Get help on farming", "h", "help"},
}

/*
	Create a "farmplot" entry in the database with a nil value for the crop type
*/

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
	}

	printFarm(s, m, input)
}

func printFarm(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	var user database.User
	user.GetUserByDiscordID(m.Author.ID)

	var farm database.Farm
	farm.GetUserFarmData(&user)

	description := fmt.Sprintf("You currently own %d plot", farm.OwnedPlots)

	// Pluralize the word "plot"
	if farm.OwnedPlots > 1 {
		description += "s"
	}

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Color:       config.CONFIG.Colors.Neutral,
			Title:       fmt.Sprintf("%s#%s's farm", m.Author.Username, m.Author.Discriminator),
			Description: description,
			Fields:      createFieldsForPlots(&farm),
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Crops will perish if not watered everyday!\nUse command '%sfarm [h | help]' for assistance", config.CONFIG.BotPrefix),
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("%s#%s", m.Author.AvatarURL("256"), m.Author.ID),
			},
		},
	}}

	// Buttons for harvesting and watering
	// Buttons are disabled if the actions are unavailable to be performed

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}

	//farm.Save()
}

func createFieldsForPlots(f *database.Farm) []*discordgo.MessageEmbedField {

	var embed []*discordgo.MessageEmbedField

	f.GetFarmPlots()

	unusedPlots := f.OwnedPlots - uint8(len(f.Plots))

	for i, p := range f.Plots {

		crop := p.GetCropInfo()

		embed = append(embed, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("Plot %d growing %s", i+1, crop.Name),
			Value:  fmt.Sprintf("%s in %s", crop.Name, crop.GetDuration()),
			Inline: true,
		})
	}

	for i := 0; i < int(unusedPlots); i++ {
		embed = append(embed, &discordgo.MessageEmbedField{
			Name:   "Empty Plot",
			Value:  "⠀",
			Inline: true,
		})
	}

	return embed
}

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

	// TODO: This does not seem to be working
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

	// Create a userFarmPlots entry with the data
	database.DB.Create(&database.FarmPlot{
		Farm: farm,
		Crop: crop,
	})

	// Update the database

	// Send message to the user

	user.Save()
	farm.Save()

}
func farmCrops(s *discordgo.Session, m *discordgo.MessageCreate) {

	var crops []database.FarmCrop
	database.DB.Find(&crops)

	description := fmt.Sprintf("Type ``%sfarm [p | plant] <crop>`` to plant a crop!\nAll seeds cost %s ``%s`` %s",
		config.CONFIG.BotPrefix,
		config.CONFIG.Emojis.Economy,
		utils.HumanReadableNumber(config.CONFIG.Farm.CropSeedPrice),
		config.CONFIG.Economy.Name)

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Color:       config.CONFIG.Colors.Neutral,
			Title:       "Crops",
			Description: description,
			Fields:      createFieldsForCrops(&crops),
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Crops will perish if not watered everyday!\nUse command '%sfarm [h | help]' for assistance", config.CONFIG.BotPrefix),
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("%s#%s", m.Author.AvatarURL("256"), m.Author.ID),
			},
		},
	}}

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}

}

func createFieldsForCrops(fc *[]database.FarmCrop) []*discordgo.MessageEmbedField {

	var embed []*discordgo.MessageEmbedField

	for _, crop := range *fc {
		embed = append(embed, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s %s", crop.Emoji, crop.Name),
			Value:  fmt.Sprintf("Takes %s\nEarns %s", crop.GetDuration(), utils.HumanReadableNumber(crop.HarvestReward)),
			Inline: true,
		})
	}
	return embed
}

func farmHelp(s *discordgo.Session, m *discordgo.MessageCreate) {

	title := "Farming Help"
	description := "These are the commands you can use with the farming system"

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Color:       config.CONFIG.Colors.Neutral,
			Title:       title,
			Description: description,
			Fields:      createHelpFields(),
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Crops will perish if not watered everyday!\nUse command '%sfarm [h | help]' for assistance", config.CONFIG.BotPrefix),
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("%s#%s", m.Author.AvatarURL("256"), m.Author.ID),
			},
		},
	}}

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}
}

func createHelpFields() []*discordgo.MessageEmbedField {

	var embed []*discordgo.MessageEmbedField

	command := fmt.Sprintf("%sfarm ", config.CONFIG.BotPrefix)

	for i, e := range farmCommands {
		embed = append(embed, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s [%v]", command, strings.Join(e[1:], " | ")),
			Value:  farmCommands[i][0],
			Inline: true,
		})
	}

	return embed
}
