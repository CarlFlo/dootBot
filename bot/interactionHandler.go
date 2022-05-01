package bot

import (
	"fmt"
	"strings"

	"github.com/CarlFlo/DiscordMoneyBot/bot/commands/farming"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// Check if the user that clicked the button is allowed to interact. e.i. the user that "created" the message

	disableButton := false

	var response string
	var embeds []*discordgo.MessageEmbed

	commandIssuerID := strings.Split(i.Message.Embeds[0].Thumbnail.URL, "#")[1]

	if i.User != nil && i.User.ID != commandIssuerID || i.Member != nil && i.Member.User.ID != commandIssuerID {
		response = "You cannot interact with this message!"
		goto sendInteraction
	}

	/* TODO
	Update the original farming message with the updated information on successful interaction
	*/

	switch i.MessageComponentData().CustomID {
	case "BWT": // BWT: Buy Work Tool
		disableButton = buyWorkTool(&response, commandIssuerID)
	case "BFP": // BFP: Buy Farm Plot
		disableButton = farming.BuyFarmPlotInteraction(commandIssuerID, &response)
	case "FH": // FH: Farm Harvest
		disableButton = farming.HarvestInteraction(commandIssuerID, &response)
	case "FW": // FW: Farm Water
		disableButton = farming.WaterInteraction(commandIssuerID, &response)
	case "FHELP":
		disableButton, embeds = farming.FarmHelpInteraction(commandIssuerID, &response)

	default:
		malm.Error("Invalid interaction: '%s'", i.MessageComponentData().CustomID)
		return
	}

sendInteraction:

	// Disables the button
	if disableButton {
		if err := disableButtonComponent(s, i.Interaction, i.MessageComponentData().CustomID); err != nil {
			malm.Error("editMsgComponentsRemoved, error: %w", err)
		}
	}

	// Delete this after some seconds?
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    response,
			Flags:      1 << 6, // Makes it so only the 'clicker' can see the message
			Components: []discordgo.MessageComponent{},
			Embeds:     embeds,
		},
	}); err != nil {
		malm.Error("Could not respond to the interaction! %w", err)
	}

}

func buyWorkTool(response *string, authorID string) bool {

	// Check if the user has enough money
	var user database.User
	user.QueryUserByDiscordID(authorID)

	var work database.Work
	work.GetWorkInfo(&user)

	price, priceStr := work.CalcBuyToolPrice()

	if uint64(price) > user.Money {
		difference := uint64(price) - user.Money
		*response = fmt.Sprintf("You are lacking ``%d`` %s for this transaction.\nYour balance: ``%d`` %s", difference, config.CONFIG.Economy.Name, user.Money, config.CONFIG.Economy.Name)
		return false
	}

	user.Money -= uint64(price)

	work.Tools += 1

	user.Save()
	work.Save()

	// TODO: Update the original message with the updated price
	// TODO: SOme bug with calculating the new price.

	*response = fmt.Sprintf("You succesfully bought an additioanl tool for %s %s", priceStr, config.CONFIG.Economy.Name)
	return true
}

func disableButtonComponent(s *discordgo.Session, i *discordgo.Interaction, customID string) error {

	for _, v := range i.Message.Components[0].(*discordgo.ActionsRow).Components {
		if v.(*discordgo.Button).CustomID == customID {
			v.(*discordgo.Button).Disabled = true
			break
		}
	}

	// Edits the message to disable the pressed button
	msgEdit := &discordgo.MessageEdit{
		Content: &i.Message.Content,
		Embed:   i.Message.Embeds[0],
		ID:      i.Message.ID,
		Channel: i.ChannelID,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: i.Message.Components[0].(*discordgo.ActionsRow).Components,
			},
		},
	}

	_, err := s.ChannelMessageEditComplex(msgEdit)
	if err != nil {
		return fmt.Errorf("cannot create message edit, error: %w", err)
	}
	return nil
}
