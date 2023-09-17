package music

import (
	"github.com/CarlFlo/dootBot/src/utils"
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

func StopMusicInteraction(guildID string, author *discordgo.User, response *string) {

	if !isMusicEnabled() {
		*response = "Music is currently disabled"
		return
	}

	vi := instances[guildID]
	if vi == nil {
		*response = "No music is currently playing"
		return
	}

	// Check if the user is in the voice channel before playing
	voiceChannelID := utils.FindVoiceChannel(author.ID)
	if vi.voice.ChannelID != voiceChannelID {
		*response = "You are not in the same voice channel as the bot"
		return
	}

	*response = "-1" // Meaning do not send a response
	leaveVoice(vi)
}

func ClearMusicQueueInteraction(guildID string, author *discordgo.User, response *string) {
	if !isMusicEnabled() {
		*response = "Music is currently disabled"
		return
	}

	vi := instances[guildID]
	if vi == nil {
		*response = "No music is currently playing"
		return
	}

	// Check if the user is in the voice channel before playing
	voiceChannelID := utils.FindVoiceChannel(author.ID)
	if vi.voice.ChannelID != voiceChannelID {
		*response = "You are not in the same voice channel as the bot"
		return
	}

	vi.ClearQueue()

	// Todo update the 'Music Player' message
	// music.CreateMusicOverviewMessage(vi.channelID, i)
}

func SongLoopInteraction(guildID string, author *discordgo.User, response *string) {
	if !isMusicEnabled() {
		*response = "Music is currently disabled"
		return
	}

	vi := instances[guildID]
	if vi == nil {
		*response = "No music is currently playing"
		return
	}

	// Check if the user is in the voice channel before playing
	voiceChannelID := utils.FindVoiceChannel(author.ID)
	if vi.voice.ChannelID != voiceChannelID {
		*response = "You are not in the same voice channel as the bot"
		return
	}

	vi.ToggleLooping()
}

func PreviousSongInteraction(guildID string, author *discordgo.User, response *string) {
	if !isMusicEnabled() {
		*response = "Music is currently disabled"
		return
	}

	vi := instances[guildID]
	if vi == nil {
		*response = "No music is currently playing"
		return
	}

	// Check if the user is in the same voice channel as the bot
	voiceChannelID := utils.FindVoiceChannel(author.ID)
	if vi.voice.ChannelID != voiceChannelID {
		*response = "You are not in the same voice channel as the bot"
		return
	}

	vi.Prev()
}
