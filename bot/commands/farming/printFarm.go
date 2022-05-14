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

	components := []discordgo.MessageComponent{}

	farm.QueryFarmPlots()

	// Harvest and water buttons
	components = append(components, &discordgo.Button{
		Label:    "Harvest",
		Disabled: !(farm.CanHarvest() && farm.HasPlantedPlots()), // !farm.CanHarvest()
		CustomID: "FH",                                           // 'FH' is code for 'Farm Harvest'
	})
	components = append(components, &discordgo.Button{
		Label:    "Water",
		Disabled: !(farm.CanWater() && farm.HasPlantedPlots()), // Disable if nothing is planted
		CustomID: "FW",                                         // 'FW' is code for 'Farm Water'
	})

	// For buying an additional plot

	plotPrice := farm.CalcFarmPlotPrice()

	canAffordPlot := user.Money >= uint64(plotPrice)

	// Add limit to the number of plots a user can buy

	components = append(components, &discordgo.Button{
		Label:    fmt.Sprintf("Buy Farm Plot (%s)", utils.HumanReadableNumber(plotPrice)),
		Style:    3, // Green color style
		Disabled: !canAffordPlot,
		Emoji: discordgo.ComponentEmoji{
			Name: config.CONFIG.Emojis.ComponentEmojiNames.MoneyBag,
		},
		CustomID: "BFP", // 'BFP' is code for 'Buy Farm Plot'
	})

	components = append(components, &discordgo.Button{
		Label:    "Help",
		Style:    2, // Gray color style
		Disabled: false,
		Emoji: discordgo.ComponentEmoji{
			Name: config.CONFIG.Emojis.ComponentEmojiNames.Help,
		},
		CustomID: "FHELP", // 'FHELP' is code for 'Farm Help'; Provies commands and information regarding farming
	})

	if len(components) == 0 {
		return nil
	}

	return []discordgo.MessageComponent{discordgo.ActionsRow{Components: components}}
}
