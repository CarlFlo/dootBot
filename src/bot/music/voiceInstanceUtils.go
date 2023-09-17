package music

func (vi *VoiceInstance) playbackStarted() {
	vi.stop = false
	vi.playing = true
	vi.loading = true
}
func (vi *VoiceInstance) playbackStopped() {
	vi.playing = false
	vi.loading = false
}

func (vi *VoiceInstance) ToggleLooping() {

	vi.looping = !vi.looping
}

// Toggles between play and pause
func (vi *VoiceInstance) PauseToggle() {

	vi.playing = !vi.playing
	vi.stream.SetPaused(vi.playing)
}

func (vi *VoiceInstance) GetGuildID() string {
	return vi.guildID
}

func (vi *VoiceInstance) SetMessageID(id string) {
	vi.messageID = id
}

func (vi *VoiceInstance) GetMessageID() string {
	return vi.messageID
}

func (vi *VoiceInstance) SetChannelID(id string) {
	vi.channelID = id
}

func (vi *VoiceInstance) GetChannelID() string {
	return vi.channelID
}

func (vi *VoiceInstance) IsLoading() bool {
	return vi.loading
}

func (vi *VoiceInstance) IsPlaying() bool {
	return vi.playing
}

func (vi *VoiceInstance) IsLooping() bool {
	return vi.looping
}
