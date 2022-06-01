package music

import "testing"

func validate(t *testing.T, got interface{}, expected interface{}, msg string) {
	if got != expected {
		t.Error("Expected", expected, ", got", got, ":", msg)
	}
}

func TestClearQueue(t *testing.T) {
	vi := VoiceInstance{}
	vi.AddToQueue(Song{Title: "song 1"})
	vi.AddToQueue(Song{Title: "song 2"})
	vi.ClearQueue()
	validate(t, vi.GetQueueLength(), 1, "QueueLength() should be 1 after ClearQueue()")

	vi.AddToQueue(Song{Title: "song 3"})
	vi.AddToQueue(Song{Title: "song 4"})
	vi.AddToQueue(Song{Title: "song 5"})

	// Simulate playing two songs
	vi.FinishedPlayingSong()
	vi.FinishedPlayingSong()

	vi.ClearQueuePrev()
	validate(t, vi.GetQueueLength(), 2, "QueueLength() should be 2 after calling ClearQueuePrev()")
}

func TestQueueNextSong(t *testing.T) {

	vi := VoiceInstance{}
	vi.AddToQueue(Song{Title: "song 1"})
	vi.AddToQueue(Song{Title: "song 2"})

	s1, _ := vi.GetFirstInQueue()
	validate(t, s1.Title, "song 1", "GetFirstInQueue() should return song 1")

	vi.FinishedPlayingSong()

	s2, _ := vi.GetFirstInQueue()
	validate(t, s2.Title, "song 2", "GetFirstInQueue() should return song 2")

	vi.AddToQueue(Song{Title: "song 3"})
	vi.SetLooping(true)
	vi.FinishedPlayingSong()

	s3, _ := vi.GetFirstInQueue()

	validate(t, s3.Title, "song 2", "The two songs should be the same")
}

func TestEndOfQueue(t *testing.T) {
	vi := VoiceInstance{}
	vi.AddToQueue(Song{Title: "song 1"})
	vi.AddToQueue(Song{Title: "song 2"})
	vi.FinishedPlayingSong()
	vi.FinishedPlayingSong()
	_, err := vi.GetFirstInQueue()
	if err == nil {
		t.Error("GetFirstInQueue() should return an error indicating that there is no song to get")
	}
}

func TestQueueAdd(t *testing.T) {
	vi := VoiceInstance{}
	validate(t, vi.GetQueueLength(), 0, "QueueLength() just initialized, should be 0")
	vi.AddToQueue(Song{Title: "song 1"})
	vi.AddToQueue(Song{Title: "song 2"})
	validate(t, vi.GetQueueLength(), 2, "QueueLength() Added two songs, so should be 2")
}

func TestQueueIncrementDecrement(t *testing.T) {
	vi := VoiceInstance{}
	vi.AddToQueue(Song{Title: "song 1"})
	vi.AddToQueue(Song{Title: "song 2"})
	vi.AddToQueue(Song{Title: "song 3"})
	validate(t, vi.GetQueueIndex(), 0, "QueueIndex() should be 0")

	vi.FinishedPlayingSong()
	vi.FinishedPlayingSong()
	vi.FinishedPlayingSong()
	validate(t, vi.GetQueueIndex(), 3, "QueueIndex() should be 3")

	vi.DecrementQueueIndex()
	vi.DecrementQueueIndex()
	vi.DecrementQueueIndex() // Index is 0 here
	vi.DecrementQueueIndex() // Should not go below zero
	vi.DecrementQueueIndex()
	validate(t, vi.GetQueueIndex(), 0, "QueueIndex() should be 0")
}
