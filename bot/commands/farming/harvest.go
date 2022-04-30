package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func farmHarvestCrops(s *discordgo.Session, m *discordgo.MessageCreate) {

	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	var farm database.Farm
	defer farm.Save()

	farm.QueryUserFarmData(&user)
	farm.QueryFarmPlots()

	description := "<no description>"

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
			Description: description,
			Fields:      fields,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Crops will perish if not watered everyday!\nUse command '%sfarm help' for assistance", config.CONFIG.BotPrefix),
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
			Name:   fmt.Sprintf("%s perished", name),
			Value:  "You didn't water it in time!",
			Inline: true,
		})
	}

	if len(embed) == 0 {
		embed = append(embed, &discordgo.MessageEmbedField{
			Name:   "Harvest information",
			Value:  "There is currently nothing to harvest",
			Inline: true,
		})
	}

	return embed
}
