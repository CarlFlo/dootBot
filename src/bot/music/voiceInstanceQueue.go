package music

import "errors"

var (
	errEmptyQueue = errors.New("the queue is empty")
	errNoNextSong = errors.New("there is no next song to play")
)

func (vi *VoiceInstance) GetFirstInQueue() (*Song, error) {
	vi.mu.RLock()
	defer vi.mu.RUnlock()

	if len(vi.queue) == 0 {
		return &Song{}, errEmptyQueue
	}

	if vi.queueIndex >= len(vi.queue) {
		return &Song{}, errNoNextSong
	}

	return vi.queue[vi.queueIndex], nil
}

func (vi *VoiceInstance) AddToQueue(s *Song) {
	vi.mu.Lock()
	vi.queue = append(vi.queue, s)
	vi.mu.Unlock()
}

func (vi *VoiceInstance) PurgeQueue() {
	vi.mu.Lock()
	vi.queueIndex = 0
	vi.queue = []*Song{}
	vi.mu.Unlock()
}

func (vi *VoiceInstance) ClearQueue() {
	vi.mu.Lock()
	defer vi.mu.Unlock()

	if vi.queueIndex >= len(vi.queue) {
		return
	}

	vi.queue = vi.queue[vi.queueIndex : vi.queueIndex+1]
	vi.queueIndex = 0
}

func (vi *VoiceInstance) ClearQueueAfter() {
	vi.mu.Lock()
	defer vi.mu.Unlock()

	if vi.queueIndex >= len(vi.queue) {
		return
	}

	vi.queue = vi.queue[:vi.queueIndex+1]
}

func (vi *VoiceInstance) ClearQueuePrev() {
	vi.mu.Lock()
	defer vi.mu.Unlock()

	if vi.queueIndex >= len(vi.queue) {
		return
	}

	vi.queue = vi.queue[vi.queueIndex:]
	vi.queueIndex = 0
}

func (vi *VoiceInstance) QueueIsEmpty() bool {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return len(vi.queue) == 0
}

func (vi *VoiceInstance) GetQueueIndex() int {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return vi.queueIndex
}

func (vi *VoiceInstance) GetQueueLength() int {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return len(vi.queue)
}

func (vi *VoiceInstance) GetQueueLengthRelative() int {
	vi.mu.RLock()
	defer vi.mu.RUnlock()

	length := len(vi.queue) - vi.queueIndex
	if length < 0 {
		return 0
	}
	return length
}

func (vi *VoiceInstance) GetSongByIndex(i int) *Song {
	vi.mu.RLock()
	defer vi.mu.RUnlock()

	if i < 0 || i >= len(vi.queue) {
		return nil
	}

	return vi.queue[i]
}

func (vi *VoiceInstance) GetNextInQueue() (*Song, bool) {
	vi.mu.RLock()
	defer vi.mu.RUnlock()

	nextIndex := vi.queueIndex + 1
	if nextIndex < 0 || nextIndex >= len(vi.queue) {
		return nil, false
	}

	return vi.queue[nextIndex], true
}
