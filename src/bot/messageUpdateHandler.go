package bot

import (
	"github.com/CarlFlo/dootBot/src/bot/commands"
	"github.com/bwmarrin/discordgo"
)

func messageUpdateHandler(s *discordgo.Session, mu *discordgo.MessageUpdate) {

	// For handeling pings
	commands.Pong(s, mu)
}
