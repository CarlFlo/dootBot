package music

func (vi *VoiceInstance) GetFirstInQueue() (*Song, error) {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	if vi.GetQueueLength() == 0 {
		return &Song{}, errEmptyQueue
	} else if vi.isEndOfQueue() {
		return &Song{}, errNoNextSong
	}

	return &vi.queue[vi.queueIndex], nil
}

func (vi *VoiceInstance) AddToQueue(s Song) {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	vi.queue = append(vi.queue, s)
}

// Removes all songs in the queue after the current song.
func (vi *VoiceInstance) ClearQueue() {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	vi.queue = vi.queue[:vi.queueIndex+1]
}

// Removes all songs in the queue before the current song.
func (vi *VoiceInstance) ClearQueuePrev() {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	vi.queue = vi.queue[vi.queueIndex:]
	vi.queueIndex = 0
}

func (vi *VoiceInstance) QueueIsEmpty() bool {
	return vi.GetQueueLength() == 0
}

func (vi *VoiceInstance) GetQueueIndex() int {
	return vi.queueIndex
}

func (vi *VoiceInstance) GetQueueLength() int {
	return len(vi.queue)
}

// Takes into account the current queue index
func (vi *VoiceInstance) GetQueueLengthRelative() int {
	return len(vi.queue) - vi.queueIndex
}

// Returns the song from the queue with the given index
func (vi *VoiceInstance) GetSongByIndex(i int) Song {
	return vi.queue[i]
}
