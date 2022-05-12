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

func printFarm(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	var farm database.Farm
	farm.QueryUserFarmData(&user)

	description := fmt.Sprintf("You currently own %d plot", farm.OwnedPlots)

	// Pluralize the word "plot"
	if farm.OwnedPlots > 1 {
		description += "s"
	}

	// TODO: Check if any of the crops perished

	// TODO: Create a farm method that creates that description or the entire message

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Color:       config.CONFIG.Colors.Neutral,
			Title:       fmt.Sprintf("%s#%s's Farm", m.Author.Username, m.Author.Discriminator),
			Description: description,
			Fields:      createFieldsForPlots(&farm),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Crops will perish if not watered everyday!",
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

func createFieldsForPlots(f *database.Farm) []*discordgo.MessageEmbedField {

	var embed []*discordgo.MessageEmbedField

	f.QueryFarmPlots()

	for i, p := range f.Plots {

		p.QueryCropInfo()

		embed = append(embed, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d) %s %s", i+1, p.Crop.Emoji, p.Crop.Name),
			Value:  p.HarvestableAt(),
			Inline: true,
		})
	}

	unusedPlots := f.OwnedPlots - uint8(len(f.Plots))

	emptyPlotValue := strings.Repeat(config.CONFIG.Emojis.EmptyPlot, 5)

	for i := 0; i < int(unusedPlots); i++ {
		embed = append(embed, &discordgo.MessageEmbedField{
			Name: fmt.Sprintf("%d) Empty Plot ", i+1+len(f.Plots)),
			//Value:  "⠀",
			Value:  emptyPlotValue,
			Inline: true,
		})
	}

	return embed
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
