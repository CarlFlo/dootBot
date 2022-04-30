package farming

import (
	"fmt"
	"strings"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/bwmarrin/discordgo"
)

func FarmHelpInteraction(discordID string, response *string) (bool, []*discordgo.MessageEmbed) {

	embeds := []*discordgo.MessageEmbed{{
		Title:       "Farming Help",
		Description: "These are the commands you can use with the farming system:",
		Fields:      createHelpFields(),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Use the command '%sfarm' and then the interaction buttons for assistance", config.CONFIG.BotPrefix),
		},
	},
	}

	return false, embeds
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
