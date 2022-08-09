package farming

import (
	"fmt"

	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/database"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func farmCrops(s *discordgo.Session, m *discordgo.MessageCreate) {

	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	var farm database.Farm
	farm.QueryUserFarmData(&user)

	var crops []database.FarmCrop
	database.DB.Order("id asc").Limit(int(farm.HighestPlantedCropIndex)).Find(&crops)

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
				Text: "Crops will perish if not watered every day!\nYou unlock new crops by planting your newest crops on your farm.",
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
