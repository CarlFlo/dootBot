package music

import "errors"

var (
	errEmptyQueue = errors.New("the queue is empty")
	errNoNextSong = errors.New("there is no next song to play")
)

func (vi *VoiceInstance) GetFirstInQueue() (*Song, error) {
	vi.mu.Lock()
	defer vi.mu.Unlock()
	if vi.GetQueueLength() == 0 {
		return &Song{}, errEmptyQueue
	} else if vi.isEndOfQueue() {
		return &Song{}, errNoNextSong
	}

	return vi.queue[vi.queueIndex], nil
}

// AddToQueue - adds the song to the queue, and also prepares the song and caches it
func (vi *VoiceInstance) AddToQueue(s *Song) {
	vi.mu.Lock()
	vi.queue = append(vi.queue, s)
	vi.mu.Unlock()
}

// Removes all songs in the queue after the current song.
func (vi *VoiceInstance) ClearQueue() {
	vi.mu.Lock()
	defer vi.mu.Unlock()
	vi.queue = vi.queue[:vi.queueIndex+1]
}

// Removes all songs in the queue before the current song.
func (vi *VoiceInstance) ClearQueuePrev() {
	vi.mu.Lock()
	defer vi.mu.Unlock()
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
func (vi *VoiceInstance) GetSongByIndex(i int) *Song {
	return vi.queue[i]
}
