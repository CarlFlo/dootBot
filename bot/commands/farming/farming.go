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

func Farming(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	// Handle farm arguments

	if input.ArgsContains([]string{"p", "plant"}) {
		// User wants to plant some seeds
		farmPlant(s, m, input)
		return
	} else if input.ArgsContains([]string{"c", "crop", "crops"}) {
		// User wants info about crops/seeds
		farmCrops(s, m)
		return
	} else if input.ArgsContains([]string{"h", "help"}) {
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
				Text: fmt.Sprintf("Crops will perish if not watered everyday!\nUse command '%sfarm [help | h]' for assistance", config.CONFIG.BotPrefix),
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

}

func farmCrops(s *discordgo.Session, m *discordgo.MessageCreate) {

	var crops []database.FarmCrop
	database.DB.Find(&crops)

	Description := fmt.Sprintf("Type ``%sfarm [plant | p] <crop>`` to plant a crop!\nAll seeds cost %s ``%s``",
		config.CONFIG.BotPrefix,
		config.CONFIG.Emojis.Economy,
		utils.HumanReadableNumber(config.CONFIG.Farm.CropSeedPrice))

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Color:       config.CONFIG.Colors.Neutral,
			Title:       "Crops",
			Description: Description,
			Fields:      createFieldsForCrops(&crops),
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Crops will perish if not watered everyday!\nUse command '%sfarm [help | h]' for assistance", config.CONFIG.BotPrefix),
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

}
