package commands

import (
	"fmt"
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

// Debug - prints some debug information
func Work(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput) {

	// Check if user can work

	// Get the current time
	currentTime := time.Now()
	// Add six hours
	currentTime = currentTime.Add(time.Hour * -6)

	// convert to unix time
	untilYouCanWorkAgain := currentTime.Unix()

	//menuComponent := []discordgo.MessageComponent{}

	moneyEarned, currentStreak := 0, 0

	_, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("You performed some manual labour and earned some money.\nYou earned **%d** money.\nYou will be able to work again <t:%d:R>\nCurrent streak **%d**\n\nBuying additional tools will allow you to earn more money", moneyEarned, untilYouCanWorkAgain, currentStreak),
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{

					&discordgo.Button{
						Label:    "Buy Axe",
						Disabled: false,
						CustomID: "buyAxe",
					},
					&discordgo.Button{
						Label:    "Buy Pickaxe",
						Disabled: false,
						CustomID: "buyPickaxe",
					},
					&discordgo.Button{
						Label:    "Buy Shovel",
						Disabled: false,
						CustomID: "buyShovel",
					},
					&discordgo.Button{
						Label:    "Buy Hammer",
						Disabled: false,
						CustomID: "buyHammer",
					},
				},
			},
		},
	})

	if err != nil {
		malm.Error("Could not send message! %s", err)
	}

}

/*
const lib = require('lib')({token: process.env.STDLIB_SECRET_TOKEN});

await lib.discord.channels['@0.3.0'].messages.create({
  "channel_id": `${context.params.event.channel_id}`,
  "content": "",
  "tts": false,
  "components": [
    {
      "type": 1,
      "components": [
        {
          "style": 1,
          "label": `Buy Axe`,
          "custom_id": `row_0_button_0`,
          "disabled": false,
          "type": 2
        },
        {
          "style": 1,
          "label": `Buy Pickaxe`,
          "custom_id": `row_0_button_1`,
          "disabled": false,
          "type": 2
        }
      ]
    }
  ]
});
*/
