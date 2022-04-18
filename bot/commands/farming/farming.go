package farming

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func Farming(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	// Handle farm arguments
	if input.ArgsContains([]string{"p", "plant"}) {
		farmPlant(s, m, input)
		return
	}
	if input.ArgsContains([]string{"h", "help"}) {
		farmHelp(s, m)
		return
	}

	printFarm(s, m, input)
}

func printFarm(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	var user database.User

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		&discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Color:       config.CONFIG.Colors.Neutral,
			Title:       fmt.Sprintf("%s#%s farm", m.Author.Username, m.Author.Discriminator),
			Description: "-",
			Fields:      createFieldsForPlots(&user),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Crops will perish if not watered everyday!\nUse command 'Farm <help or h>' for assistance",
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

func createFieldsForPlots(user *database.User) []*discordgo.MessageEmbedField {

	return nil
}

func farmPlant(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

}

func farmHelp(s *discordgo.Session, m *discordgo.MessageCreate) {

}

/*
&discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("Wallet %s", config.CONFIG.Emojis.Wallet),
					Value:  fmt.Sprintf("%s %s", config.CONFIG.Emojis.Economy, user.PrettyPrintMoney()),
					Inline: true,
				},
*/