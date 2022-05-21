package bot

import (
	"strings"

	"github.com/CarlFlo/DiscordMoneyBot/bot/commands"
	"github.com/CarlFlo/DiscordMoneyBot/bot/commands/daily"
	"github.com/CarlFlo/DiscordMoneyBot/bot/commands/farming"
	"github.com/CarlFlo/DiscordMoneyBot/bot/commands/work"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// Check if the user that clicked the button is allowed to interact. e.i. the user that "created" the message

	msgEdit := &discordgo.MessageEdit{Channel: i.ChannelID, ID: i.Message.ID}
	var response string
	var responseEmbed []*discordgo.MessageEmbed

	commandIssuerID := strings.Split(i.Message.Embeds[0].Thumbnail.URL, "#")[1]

	if i.User != nil && i.User.ID != commandIssuerID || i.Member != nil && i.Member.User.ID != commandIssuerID {
		response = "You cannot interact with this message!"
		goto sendInteractionResponse
	}

	switch i.MessageComponentData().CustomID {
	case "BWT": // BWT: Buy Work Tool
		work.BuyToolInteraction(commandIssuerID, &response, i.Interaction, msgEdit)
		// Farming
	case "BFP": // BFP: Buy Farm Plot
		farming.BuyFarmPlotInteraction(commandIssuerID, &response, s, msgEdit)
	case "FPC": // FPC: Farm Plant Crop - Plants a crop from the farm message using the menu
		farming.FarmPlantInteraction(commandIssuerID, &response, i.Interaction, s, msgEdit)
	case "FH": // FH: Farm Harvest
		farming.HarvestInteraction(commandIssuerID, &response, s, msgEdit)
	case "FW": // FW: Farm Water
		farming.WaterInteraction(commandIssuerID, &response, s, msgEdit)
	case "FHELP":
		farming.FarmHelpInteractionEmbedCreate(&responseEmbed)
		// Profile
	case "RP": // RP: Refresh Profile
		commands.ProfileRefreshInteraction(commandIssuerID, i.Interaction.Member.User, msgEdit)
	case "PW": // PW: Profile Work - User worked from the profile message
		work.DoWorkInteraction(commandIssuerID, &response, i.Interaction.Member.User, msgEdit)
	case "PD": // PD: Profile Daily - User did their daily from the profile message
		daily.DoDailyInteraction(commandIssuerID, &response, i.Interaction.Member.User, msgEdit)
	default:
		malm.Error("Invalid interaction: '%s'", i.MessageComponentData().CustomID)
		return
	}

	if msgEdit.Embeds != nil {
		if _, err := s.ChannelMessageEditComplex(msgEdit); err != nil {
			malm.Error("cannot create message edit, error: %s", err)
		}
	}

	// Nothing to reply with
	if len(response) == 0 && len(responseEmbed) == 0 {
		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
		}); err != nil {
			malm.Error("Could not respond to the interaction! %s", err)
		}
		return
	}

sendInteractionResponse:

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
			Flags:   1 << 6, // Makes it so only the 'clicker' can see the message
			Embeds:  responseEmbed,
		},
	}); err != nil {
		malm.Error("Could not respond to the interaction! %s", err)
	}
}
