package bot

import (
	"fmt"

	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func readyHandler(s *discordgo.Session, ready *discordgo.Ready) {

	malm.Info("Bot is connected and present on %d servers", len(s.State.Guilds))

	statusMessage := fmt.Sprintf("on %d servers", len(s.State.Guilds))

	// Shows up like the bot is streaming. Allows us to have a link.
	s.UpdateStreamingStatus(0, statusMessage, "https://www.youtube.com/watch?v=3glxLWVkbSs")
	// Normal message
	//s.UpdateGameStatus(0, statusMessage)
}
