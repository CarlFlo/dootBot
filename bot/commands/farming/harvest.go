package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func HarvestInteraction(discordID string, response *string, disableButton *bool) {

	var user database.User
	user.QueryUserByDiscordID(discordID)

	var farm database.Farm
	defer farm.Save()

	farm.QueryUserFarmData(&user)
	farm.QueryFarmPlots()

	perishedCrops := farm.CropsPerishedCheck()

	result := farm.HarvestPlots()

	if len(result) == 0 && len(perishedCrops) == 0 {
		*response = "There is currently nothing ready to be harvested!"
		return
	}

	*response = "Your harvest:\n"

	for _, e := range result {
		*response += fmt.Sprintf("%s %s\n", e.Emoji, e.Name)
	}

	for _, name := range perishedCrops {
		*response += fmt.Sprintf("%s %s perished\n", config.CONFIG.Emojis.PerishedCrop, name)
	}

	if farm.SuccessfulHarvest() {
		*response += fmt.Sprintf("\nYou earned %s", utils.HumanReadableNumber(farm.HarvestEarnings))
		user.Money += uint64(farm.HarvestEarnings)
		user.Save()
	}

	*disableButton = true
}

func farmHarvestCrops(s *discordgo.Session, m *discordgo.MessageCreate) {

	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	var farm database.Farm
	defer farm.Save()

	farm.QueryUserFarmData(&user)
	farm.QueryFarmPlots()

	fields := createFieldsForHarvest(&farm)

	color := config.CONFIG.Colors.Failure

	if farm.SuccessfulHarvest() {
		color = config.CONFIG.Colors.Success
		user.Money += uint64(farm.HarvestEarnings)
		user.Save()
	}

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Color:       color,
			Title:       fmt.Sprintf("%s#%s's harvest", m.Author.Username, m.Author.Discriminator),
			Description: "",
			Fields:      fields,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Crops will perish if not watered everyday!",
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

func createFieldsForHarvest(f *database.Farm) []*discordgo.MessageEmbedField {

	var embed []*discordgo.MessageEmbedField

	perishedCrops := f.CropsPerishedCheck()

	result := f.HarvestPlots()

	for _, e := range result {
		embed = append(embed, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s %s", e.Emoji, e.Name),
			Value:  fmt.Sprintf("You earned %d", e.Earning),
			Inline: true,
		})
	}

	for _, name := range perishedCrops {
		embed = append(embed, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s %s perished", config.CONFIG.Emojis.PerishedCrop, name),
			Value:  "You didn't water it in time!",
			Inline: true,
		})
	}

	if len(embed) == 0 {
		embed = append(embed, &discordgo.MessageEmbedField{
			Name:   "Harvest information",
			Value:  "There is currently nothing ready to be harvested",
			Inline: true,
		})
	}

	return embed
}
