package bot

import (
	"fmt"
	"strings"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// Check if the user that clicked the button is allowed to interact. e.i. the user that "created" the message

	cData := strings.Split(i.MessageComponentData().CustomID, "-")

	var response string

	commandIssuerID := strings.Split(i.Message.Embeds[0].Thumbnail.URL, "#")[1]

	if i.User != nil && i.User.ID != commandIssuerID || i.Member != nil && i.Member.User.ID != commandIssuerID {
		response = "You cannot interact with this message!"
		goto sendInteraction
	}

	switch cData[0] {
	case "BWT": // BWT: Buy Work Tool
		buyWorkTool(cData, &response, commandIssuerID)
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

func buyWorkTool(cData []string, response *string, authorID string) {
	//malm.Info("Interaction: '%s' item: '%s'", cData[0], cData[1])

	// Find the item in config.CONFIG.Work.Tools
	index := -1
	for i, e := range config.CONFIG.Work.Tools {
		if e.Name == cData[1] {
			index = i
			break
		}
	}

	// We got nothing
	if index == -1 {
		malm.Error("Could not find the item '%s' '%s'", cData[0], cData[1])
		return
	}

	// Check if the user has enough money

	var user database.User
	database.DB.Table("Users").Where("discord_id = ?", authorID).First(&user)

	if config.CONFIG.Work.Tools[index].Price > int(user.Money) {
		*response = fmt.Sprintf("You do not have enough %s for this transaction\nYou have %d and you need %d", config.CONFIG.Economy.Name, user.Money, config.CONFIG.Work.Tools[index].Price)
		return
	}

	user.Money -= uint64(config.CONFIG.Work.Tools[index].Price)

	var work database.Work
	database.DB.Raw("select * from Works JOIN Users ON Works.ID = Users.ID WHERE Users.discord_id = ?", authorID).First(&work)
	work.Tools |= 1 << index

	database.DB.Save(&user)
	database.DB.Save(&work)

	*response = fmt.Sprintf("You succesfully bought the %s for %d %s", cData[1], config.CONFIG.Work.Tools[index].Price, config.CONFIG.Economy.Name)
}
