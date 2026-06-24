package music

import "github.com/bwmarrin/discordgo"

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

	if err := validateSameVoiceChannel(vi, author.ID); err != nil {
		*response = err.Error()
		return
	}

	if vi.IsLoading() {
		*response = "Hold on! The bot is loading the song"
		return
	}

	if !vi.PauseToggle() {
		*response = "There is no active song to pause or resume"
		return
	}

	if err := vi.refreshOverviewMessage(); err != nil {
		*response = "Unable to refresh the music message"
	}
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

	if err := validateSameVoiceChannel(vi, author.ID); err != nil {
		*response = err.Error()
		return
	}

	*response = "-1"
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

	if err := validateSameVoiceChannel(vi, author.ID); err != nil {
		*response = err.Error()
		return
	}

	vi.ClearQueueAfter()
	if err := vi.refreshOverviewMessage(); err != nil {
		*response = "Unable to refresh the music message"
	}
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

	if err := validateSameVoiceChannel(vi, author.ID); err != nil {
		*response = err.Error()
		return
	}

	vi.ToggleLooping()
	if err := vi.refreshOverviewMessage(); err != nil {
		*response = "Unable to refresh the music message"
	}
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

	if err := validateSameVoiceChannel(vi, author.ID); err != nil {
		*response = err.Error()
		return
	}

	if !vi.Prev() {
		*response = "There is no song to restart"
		return
	}
}

func NextSongInteraction(guildID string, author *discordgo.User, response *string) {
	if !isMusicEnabled() {
		*response = "Music is currently disabled"
		return
	}

	vi := instances[guildID]
	if vi == nil {
		*response = "No music is currently playing"
		return
	}

	if err := validateSameVoiceChannel(vi, author.ID); err != nil {
		*response = err.Error()
		return
	}

	if !vi.Skip() {
		*response = "There is no song to skip"
		return
	}
}
