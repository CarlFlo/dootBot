package farming

import (
	"fmt"
	"strings"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

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
