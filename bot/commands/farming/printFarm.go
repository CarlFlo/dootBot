package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func printFarm(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	var farm database.Farm
	farm.QueryUserFarmData(&user)

	// TODO: Create a farm method that creates that description or the entire message

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Color:       config.CONFIG.Colors.Neutral,
			Title:       fmt.Sprintf("%s#%s's Farm", m.Author.Username, m.Author.Discriminator),
			Description: farm.CreateEmbedDescription(),
			Fields:      farm.CreateEmbedFields(),
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Crops will perish if not watered everyday!\nYou can own up to %d farm plots!", config.CONFIG.Farm.MaxPlots),
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("%s#%s", m.Author.AvatarURL("256"), m.Author.ID),
			},
		},
	}}

	// Adds the button(s)
	// Buttons are disabled if the actions are unavailable to be performed
	if components := createButtonComponent(&user, &farm); components != nil {
		complexMessage.Components = components
	}

	// Buttons for harvesting, watering and buying new plots (and items)

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}
}

func createButtonComponent(user *database.User, farm *database.Farm) []discordgo.MessageComponent {

	output := []discordgo.MessageComponent{}
	btnComponents := []discordgo.MessageComponent{}

	farm.QueryFarmPlots()

	// Harvest and water buttons
	btnComponents = append(btnComponents, &discordgo.Button{
		Label:    "Harvest",
		Disabled: !(farm.CanHarvest() && farm.HasPlantedPlots()), // !farm.CanHarvest()
		CustomID: "FH",                                           // 'FH' is code for 'Farm Harvest'
	})
	btnComponents = append(btnComponents, &discordgo.Button{
		Label:    "Water",
		Disabled: !(farm.CanWater() && farm.HasPlantedPlots()), // Disable if nothing is planted
		CustomID: "FW",                                         // 'FW' is code for 'Farm Water'
	})

	// For buying an additional plot

	plotPrice := farm.CalcFarmPlotPrice()

	canAffordPlot := user.Money >= uint64(plotPrice)

	// Add limit to the number of plots a user can buy

	btnComponents = append(btnComponents, &discordgo.Button{
		Label:    fmt.Sprintf("Buy Farm Plot (%s)", utils.HumanReadableNumber(plotPrice)),
		Style:    3, // Green color style
		Disabled: !canAffordPlot,
		Emoji: discordgo.ComponentEmoji{
			Name: config.CONFIG.Emojis.ComponentEmojiNames.MoneyBag,
		},
		CustomID: "BFP", // 'BFP' is code for 'Buy Farm Plot'
	})

	btnComponents = append(btnComponents, &discordgo.Button{
		Label:    "Help",
		Style:    2, // Gray color style
		Disabled: false,
		Emoji: discordgo.ComponentEmoji{
			Name: config.CONFIG.Emojis.ComponentEmojiNames.Help,
		},
		CustomID: "FHELP", // 'FHELP' is code for 'Farm Help'; Provies commands and information regarding farming
	})

	output = append(output, discordgo.ActionsRow{
		Components: btnComponents,
	})

	// Ability to select plantable crop from list. Only if it makes sense to add it
	if cropMenu := createPlantCropMenu(user, farm); cropMenu != nil {

		cropComponents := []discordgo.MessageComponent{cropMenu}
		output = append(output, discordgo.ActionsRow{
			Components: cropComponents,
		})
	}

	return output
}

func createPlantCropMenu(user *database.User, farm *database.Farm) *discordgo.SelectMenu {

	// User can't afford to plant so no need to create the menu
	// TODO: Needs to be checked on interaction as well
	if !user.CanAfford(uint64(config.CONFIG.Farm.CropSeedPrice)) {
		return nil
	}

	output := &discordgo.SelectMenu{
		CustomID:    "PC", // 'PC' is code for 'Plant Crop'
		Placeholder: "Select a crop to plant",
		MaxValues:   1,
		Options:     createCropOptions(farm),
	}
	return output
}

func createCropOptions(farm *database.Farm) []discordgo.SelectMenuOption {

	options := []discordgo.SelectMenuOption{}

	var crops []database.FarmCrop
	database.DB.Order("id asc").Limit(int(farm.HighestPlantedCropIndex)).Find(&crops)

	for _, crop := range crops {

		options = append(options, discordgo.SelectMenuOption{
			Label: crop.Name,
			Value: crop.Name,
			Emoji: discordgo.ComponentEmoji{
				Name: crop.Emoji,
			},
		})
	}

	return options
}
