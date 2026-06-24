package music

func (vi *VoiceInstance) startWorker() bool {
	vi.mu.Lock()
	defer vi.mu.Unlock()

	if vi.workerRunning {
		return false
	}

	vi.workerRunning = true
	vi.stopRequested = false
	vi.loading = false
	vi.paused = false
	return true
}

func (vi *VoiceInstance) finishWorker() {
	vi.mu.Lock()
	vi.workerRunning = false
	vi.loading = false
	vi.paused = false
	vi.stopRequested = false
	vi.previous = false
	vi.mu.Unlock()
}

func (vi *VoiceInstance) requestStop() {
	vi.mu.Lock()
	vi.stopRequested = true
	vi.mu.Unlock()
	vi.signalDone(nil)
}

func (vi *VoiceInstance) shouldStop() bool {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return vi.stopRequested
}

func (vi *VoiceInstance) ToggleLooping() {
	vi.mu.Lock()
	vi.looping = !vi.looping
	vi.mu.Unlock()
}

func (vi *VoiceInstance) PauseToggle() bool {
	vi.mu.Lock()
	defer vi.mu.Unlock()

	if vi.stream == nil || !vi.workerRunning {
		return false
	}

	vi.paused = !vi.paused
	vi.stream.SetPaused(vi.paused)
	return true
}

func (vi *VoiceInstance) GetGuildID() string {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return vi.guildID
}

func (vi *VoiceInstance) SetMessageID(id string) {
	vi.mu.Lock()
	vi.messageID = id
	vi.mu.Unlock()
}

func (vi *VoiceInstance) GetMessageID() string {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return vi.messageID
}

func (vi *VoiceInstance) SetMessageChannelID(id string) {
	vi.mu.Lock()
	vi.messageChannelID = id
	vi.mu.Unlock()
}

func (vi *VoiceInstance) GetMessageChannelID() string {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return vi.messageChannelID
}

func (vi *VoiceInstance) IsLoading() bool {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return vi.loading
}

func (vi *VoiceInstance) IsPlaying() bool {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return vi.workerRunning && !vi.loading && !vi.paused
}

func (vi *VoiceInstance) IsPaused() bool {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return vi.paused
}

func (vi *VoiceInstance) IsWorkerRunning() bool {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return vi.workerRunning
}

func (vi *VoiceInstance) IsStartOfQueue() bool {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return vi.queueIndex == 0
}

func (vi *VoiceInstance) IsLooping() bool {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return vi.looping
}

func (vi *VoiceInstance) setLoading(loading bool) {
	vi.mu.Lock()
	vi.loading = loading
	vi.mu.Unlock()
}

func (vi *VoiceInstance) setPaused(paused bool) {
	vi.mu.Lock()
	vi.paused = paused
	if vi.stream != nil {
		vi.stream.SetPaused(paused)
	}
	vi.mu.Unlock()
}

func (vi *VoiceInstance) signalDone(err error) {
	vi.mu.RLock()
	done := vi.done
	vi.mu.RUnlock()

	if done == nil {
		return
	}

	select {
	case done <- err:
	default:
	}
}
