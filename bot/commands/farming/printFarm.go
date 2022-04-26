package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

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

	for i, p := range f.Plots {

		crop := p.GetCropInfo()

		embed = append(embed, &discordgo.MessageEmbedField{
			Name: fmt.Sprintf("Plot %d growing %s", i+1, crop.Name),
			/*TODO: Change to discord formatted time*/
			Value:  fmt.Sprintf("%s in %s", crop.Name, crop.GetDuration()),
			Inline: true,
		})
	}

	unusedPlots := f.OwnedPlots - uint8(len(f.Plots))

	for i := 0; i < int(unusedPlots); i++ {
		embed = append(embed, &discordgo.MessageEmbedField{
			Name:   "Empty Plot",
			Value:  "⠀",
			Inline: true,
		})
	}

	return embed
}
