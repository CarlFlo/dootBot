package bot

import (
	"github.com/CarlFlo/DiscordMoneyBot/bot/commands"
	"github.com/bwmarrin/discordgo"
)

func messageUpdateHandler(s *discordgo.Session, mu *discordgo.MessageUpdate) {

	// For handeling pings
	commands.Pong(s, mu)
}
