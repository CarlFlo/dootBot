package mine

import (
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

// CreateDKOverviewMessage creates the Dwarvenkeep overview message
func CreateDKOverviewMessage(msg interface{}) {

	// Check the type of msg
	switch msg.(type) {
	case *discordgo.MessageSend:
		malm.Info("Message send")
	case *discordgo.MessageEdit:
		malm.Info("Editing message")
	default:
		malm.Error("Unknown message type")
	}
}
