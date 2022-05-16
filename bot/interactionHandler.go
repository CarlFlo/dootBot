package bot

import (
	"fmt"
	"strings"

	"github.com/CarlFlo/DiscordMoneyBot/bot/commands"
	"github.com/CarlFlo/DiscordMoneyBot/bot/commands/daily"
	"github.com/CarlFlo/DiscordMoneyBot/bot/commands/farming"
	"github.com/CarlFlo/DiscordMoneyBot/bot/commands/work"
	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// Check if the user that clicked the button is allowed to interact. e.i. the user that "created" the message

	var btnData []structs.ButtonData

	var response string
	var embeds []*discordgo.MessageEmbed

	commandIssuerID := strings.Split(i.Message.Embeds[0].Thumbnail.URL, "#")[1]

	if i.User != nil && i.User.ID != commandIssuerID || i.Member != nil && i.Member.User.ID != commandIssuerID {
		response = "You cannot interact with this message!"
		goto sendInteraction
	}

	switch i.MessageComponentData().CustomID {
	case "BWT": // BWT: Buy Work Tool
		work.BuyToolInteraction(commandIssuerID, &response, &btnData, i.Interaction)
	case "BFP": // BFP: Buy Farm Plot
		farming.BuyFarmPlotInteraction(commandIssuerID, &response, &btnData, i.Interaction)
	case "FPC": // FPC: Farm Plant Crop - Plants a crop from the farm message using the menu
		farming.FarmPlantInteraction(commandIssuerID, &response, i.Interaction)
	case "FH": // FH: Farm Harvest
		farming.HarvestInteraction(commandIssuerID, &response, &btnData, i.Interaction)
	case "FW": // FW: Farm Water
		farming.WaterInteraction(commandIssuerID, &response, &btnData, i.Interaction)
	case "FHELP":
		embeds = farming.FarmHelpInteraction(commandIssuerID, &response)
	case "RP": // RP: Refresh Profile
		commands.ProfileRefreshInteraction(commandIssuerID, i.Interaction, &btnData)
	case "PW": // PW: Profile Work - User worked from the profile message
		work.DoWorkInteraction(commandIssuerID, &response, i.Interaction, &btnData)
	case "PD": // PD: Profile Daily - User did their daily from the profile message
		daily.DoDailyInteraction(commandIssuerID, &response, i.Interaction, &btnData)
	default:
		malm.Error("Invalid interaction: '%s'", i.MessageComponentData().CustomID)
		return
	}

sendInteraction:

	// Updates the button on the original message
	if err := updateButtonComponent(s, i.Interaction, &btnData); err != nil {
		malm.Error("editMsgComponentsRemoved, error: %w", err)
	}

	// Nothing to reply with
	if len(response) == 0 && len(embeds) == 0 {
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
		}); err != nil {
			malm.Error("Could not respond to the interaction! %w", err)
		}
		return
	}

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

func updateButtonComponent(s *discordgo.Session, i *discordgo.Interaction, btnData *[]structs.ButtonData) error {

	for _, v := range i.Message.Components[0].(*discordgo.ActionsRow).Components {

		for _, btn := range *btnData {
			if v.(*discordgo.Button).CustomID == btn.CustomID {
				v.(*discordgo.Button).Disabled = btn.Disabled

				if len(btn.Label) != 0 {
					v.(*discordgo.Button).Label = btn.Label
				}
			}
		}
	}

	// Edits the message
	msgEdit := &discordgo.MessageEdit{
		Content: &i.Message.Content,
		Embeds:  i.Message.Embeds,
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
