package bot

import (
	"strings"

	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// Check if the user that clicked the button is allowed to interact. e.i. the user that "created" the message

	cData := strings.Split(i.MessageComponentData().CustomID, "-")

	switch cData[0] {
	case "BWT": // BWT: Buy Work Tool
		buyWorkTool(cData)
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    "You clicked a button!",
			Components: []discordgo.MessageComponent{},
		},
	}); err != nil {
		malm.Error("Could not respond to the interaction! %s", err)
	}

}

func buyWorkTool(cData []string) {
	malm.Info("Interaction: %s", cData[0])

}
