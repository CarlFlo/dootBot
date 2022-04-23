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

var (
	farmCommands = [][]string{
		{"Plant a crop", "p", "plant"},
		{"Get info about available crops", "c", "crop", "crops"},
		{"Get help on farming", "h", "help"},
	}

	defaultFooter = fmt.Sprintf("Crops will perish if not watered everyday!\nUse command '%sfarm [h | help]' for assistance", config.CONFIG.BotPrefix)
)

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
	defer user.Save()
	user.GetUserByDiscordID(m.Author.ID)

	var farm database.Farm
	defer farm.Save()
	farm.GetFarmInfo(&user)

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Color:       config.CONFIG.Colors.Neutral,
			Title:       fmt.Sprintf("%s#%s farm", m.Author.Username, m.Author.Discriminator),
			Description: "",
			Fields:      createFieldsForPlots(&farm),
			Footer: &discordgo.MessageEmbedFooter{
				Text: defaultFooter,
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
}

func createFieldsForPlots(f *database.Farm) []*discordgo.MessageEmbedField {

	/*
		&discordgo.MessageEmbedField{
							Name:   fmt.Sprintf("Wallet %s", config.CONFIG.Emojis.Wallet),
							Value:  fmt.Sprintf("%s %s", config.CONFIG.Emojis.Economy, user.PrettyPrintMoney()),
							Inline: true,
						},
	*/

	return nil
}

func farmPlant(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	// Check for input (that a plant has been specified)

	// Check if the user has a free plot

	// Check if the user has enough money to buy seeds

	// parse the input plant (check the database)

	// Plant the seed

	// Update the database

	// Send message to the user

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
				Text: defaultFooter,
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
				Text: defaultFooter,
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
