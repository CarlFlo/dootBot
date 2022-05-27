package music

import (
	"errors"
	"io"
	"sync"
	"time"

	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
	"github.com/jung-m/dca"
)

type VoiceInstance struct {
	voice      *discordgo.VoiceConnection
	Session    *discordgo.Session
	encoder    *dca.EncodeSession
	stream     *dca.StreamingSession
	queueMutex sync.Mutex
	queue      []Song // First song in the queue is the current song
	GuildID    string
	playing    bool
	paused     bool
	stop       bool // Means clearing the queue
	done       chan error
}

type Song struct {
	ChannelID   string
	User        string // Who requested the song
	Thumbnail   string
	ChannelName string
	Title       string
	YoutubeURL  string
	StreamURL   string
	Duration    string
}

func (vi *VoiceInstance) playingStarted() {
	vi.playing = true
	vi.paused = false
}
func (vi *VoiceInstance) playingStopped() {
	vi.stop = false
	vi.playing = false
}

// Plays the Queue
func (vi *VoiceInstance) PlayQueue() {

	for {
		vi.playingStarted()

		if err := vi.voice.Speaking(true); err != nil {
			malm.Error("%s", err)
			return
		}

		// This is the function that streams the audio to the voice channel
		err := vi.StreamAudio()
		if err != nil {
			malm.Error("%s", err)
			return
		}

		if vi.stop {
			vi.ClearQueue()
			return
		}
		vi.RemoveFirstInQueue()

		vi.playingStopped()

		err = vi.voice.Speaking(false)
		if err != nil {
			malm.Error("%s", err)
			return
		}

		if vi.QueueIsEmpty() {
			return
		}
	}
}

func (vi *VoiceInstance) StreamAudio() error {

	settings := dca.StdEncodeOptions
	// Custom settings
	settings.RawOutput = true
	settings.Bitrate = 64
	//settings.Application = "lowdelay"

	song, err := vi.GetFirstInQueue()
	if err != nil {
		return err
	}

	// This function is slow. ~2 seconds
	err = execYoutubeDL(&song)
	if err != nil {
		return err
	}

	vi.encoder, err = dca.EncodeFile(song.StreamURL, settings)
	if err != nil {
		return err
	}

	vi.done = make(chan error)
	vi.stream = dca.NewStream(vi.encoder, vi.voice, vi.done)

	// Ignore this problem. Using a range here does not work properly for this purpose
	for {
		select {
		case err := <-vi.done:
			if err != nil && err != io.EOF {
				return err
			}
			vi.encoder.Cleanup()
			return nil
		}
	}
}

func (vi *VoiceInstance) AddToQueue(s Song) {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	vi.queue = append(vi.queue, s)
}

func (vi *VoiceInstance) ClearQueue() {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	vi.queue = []Song{}
}

func (vi *VoiceInstance) RemoveFirstInQueue() {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()

	// Only one entry in the queue, so clear it
	if len(vi.queue) == 1 {
		vi.queue = []Song{}
		return
	}
	vi.queue = vi.queue[1:]
}

func (vi *VoiceInstance) GetFirstInQueue() (Song, error) {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	if len(vi.queue) == 0 {

		return Song{}, errors.New("the queue is empty")
	}
	return vi.queue[0], nil
}

func (vi *VoiceInstance) QueueIsEmpty() bool {
	return len(vi.queue) == 0
}

// Disconnect dissconnects the bot from the voice connection
func (vi *VoiceInstance) Disconnect() {
	vi.Stop()
	time.Sleep(200 * time.Millisecond)

	err := vi.voice.Disconnect()
	if err != nil {
		malm.Error("%s", err)
		return
	}
}

// Skip skipps the song. returns true of success, else false
func (vi *VoiceInstance) Skip() bool {

	if !vi.playing {
		return false
	}

	// This will interupt and stop the stream
	vi.done <- nil

	return true
}

func (vi *VoiceInstance) IsPlaying() bool {
	return vi.playing
}

// Stops the current song and clears the queue. returns true of success, else false
func (vi *VoiceInstance) Stop() bool {

	if !vi.playing {
		return false
	}

	vi.stop = true

	// This will interupt and stop the stream
	vi.done <- nil

	return true
}

func (vi *VoiceInstance) Pause() {

	vi.paused = !vi.paused
	vi.stream.SetPaused(vi.paused)
}
