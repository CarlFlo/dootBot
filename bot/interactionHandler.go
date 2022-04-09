package bot

import (
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// Check if the user that clicked the button is allowed to interact. e.i. the user that "created" the message

	component := i.MessageComponentData()

	malm.Info("Interaction: %s", component.CustomID)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "You clicked a button!",
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.Button{
							Label:    "X",
							Disabled: true,
							CustomID: component.CustomID,
						},
					},
				},
			},
		},
	})

	if err != nil {
		malm.Error("Could not respond to the interaction! %s", err)
	}
}
