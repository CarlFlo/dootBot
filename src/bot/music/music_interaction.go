package music

import (
	"github.com/CarlFlo/DiscordMoneyBot/src/utils"
	"github.com/bwmarrin/discordgo"
)

func PlayMusicInteraction(guildID string, author *discordgo.User, response *string) {

	if !isMusicEnabled() {
		*response = "Music is currently disabled"
		return
	}

	vi := instances[guildID]
	if vi == nil {
		*response = "You cannot start the bot from an interaction. Run it normally first"
		return
	}

	if vi.IsLoading() {
		*response = "Hold on! The bot is loading the song"
		return
	}

	// Check if the user is in the voice channel before playing
	voiceChannelID := utils.FindVoiceChannel(author.ID)
	if vi.voice.ChannelID != voiceChannelID {
		*response = "You are not in the same voice channel as the bot"
		return
	}

	vi.PauseToggle()
}
