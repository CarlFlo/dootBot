package bot

import (
	"fmt"
	"strings"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// Check if the user that clicked the button is allowed to interact. e.i. the user that "created" the message

	cData := strings.Split(i.MessageComponentData().CustomID, "-")

	var response string

	commandIssuerID := strings.Split(i.Message.Embeds[0].Thumbnail.URL, "#")[1]

	if i.User != nil && i.User.ID != commandIssuerID || i.Member != nil && i.Member.User.ID != commandIssuerID {

		/*
			var youID string

			if i.User != nil {
				youID = i.User.ID
			} else if i.Member != nil {
				youID = i.Member.User.ID
			}
		*/

		response = "You cannot interact with this message!"
		goto sendInteraction
	}

	switch cData[0] {
	case "BWT": // BWT: Buy Work Tool
		buyWorkTool(cData, &response)
	default:
		malm.Error("Invalid interaction: '%s'", i.MessageComponentData().CustomID)
	}

sendInteraction:

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    response,
			Flags:      1 << 6, // Makes it so only the clicker can see the message
			Components: []discordgo.MessageComponent{},
		},
	}); err != nil {
		malm.Error("Could not respond to the interaction! %s", err)
	}

}

func buyWorkTool(cData []string, response *string) {
	malm.Info("Interaction: '%s' item: '%s' cost: '%s'", cData[0], cData[1], cData[2])

	*response = fmt.Sprintf("You tried to buy '%s' for %s %s", cData[1], cData[2], config.CONFIG.Economy.Name)
}
