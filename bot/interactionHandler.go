package bot

import (
	"fmt"
	"strings"

	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// Check if the user that clicked the button is allowed to interact. e.i. the user that "created" the message

	cData := strings.Split(i.MessageComponentData().CustomID, "-")

	var response string

	switch cData[0] {
	case "BWT": // BWT: Buy Work Tool
		buyWorkTool(cData, &response)
	default:
		malm.Error("Invalid interaction: '%s'", i.MessageComponentData().CustomID)
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    response,
			Components: []discordgo.MessageComponent{},
		},
	}); err != nil {
		malm.Error("Could not respond to the interaction! %s", err)
	}

}

func buyWorkTool(cData []string, response *string) {
	malm.Info("Interaction: '%s' value '%s'", cData[0], cData[1])

	*response = fmt.Sprintf("You tried to buy '%s'", cData[1])
}
