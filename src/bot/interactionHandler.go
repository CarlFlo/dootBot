package bot

import (
	"strings"

	"github.com/CarlFlo/dootBot/src/bot/commands"
	"github.com/CarlFlo/dootBot/src/bot/commands/daily"
	"github.com/CarlFlo/dootBot/src/bot/commands/farming"
	"github.com/CarlFlo/dootBot/src/bot/commands/work"
	"github.com/CarlFlo/dootBot/src/bot/music"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {

	// Check if the user that clicked the button is allowed to interact. e.i. the user that "created" the message

	msgEdit := &discordgo.MessageEdit{Channel: i.ChannelID, ID: i.Message.ID}
	var response string
	var responseEmbed []*discordgo.MessageEmbed
	var commandIssuerID string

	// Some messages, like music, does not have a user thumbnail (with their ID)
	if strings.Contains(i.Message.Embeds[0].Thumbnail.URL, "#") {

		commandIssuerID = strings.Split(i.Message.Embeds[0].Thumbnail.URL, "#")[1]
		if interactionValidateInteractor(i, commandIssuerID) {
			response = "You cannot interact with this message!"
			interactionResponse(s, i, &response, &responseEmbed)
			return
		}
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
	case "toggleSong":
		music.PlayMusicInteraction(i.GuildID, i.Interaction.Member.User, &response)
	case "stopSong":
		music.StopMusicInteraction(i.GuildID, i.Interaction.Member.User, &response)
	case "clearQueue":
		music.ClearMusicQueue(i.GuildID, i.Interaction.Member.User, &response)
	case "nextSong":
		malm.Info("Next song")
	default:
		malm.Error("Invalid interaction: '%s'", i.MessageComponentData().CustomID)
		return
	}

	if msgEdit.Embeds != nil {
		if _, err := s.ChannelMessageEditComplex(msgEdit); err != nil {
			malm.Error("cannot create message edit, error: %s", err)
		}
	}

	// -1 meaning that a response should not be sent. Most cases because the original message was or is going to be deleted
	if response == "-1" {
		return
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

	interactionResponse(s, i, &response, &responseEmbed)
}

func interactionResponse(s *discordgo.Session, i *discordgo.InteractionCreate, response *string, responseEmbed *[]*discordgo.MessageEmbed) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: *response,
			Flags:   1 << 6, // Makes it so only the 'clicker' can see the message
			Embeds:  *responseEmbed,
		},
	}); err != nil {
		malm.Error("Could not respond to the interaction! %s", err)
	}
}

// Returns true if the message interaction is from the same user that promted the bot to created the original message
func interactionValidateInteractor(i *discordgo.InteractionCreate, commandIssuerID string) bool {
	return i.User != nil && i.User.ID != commandIssuerID || i.Member != nil && i.Member.User.ID != commandIssuerID
}
