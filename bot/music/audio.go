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
	Voice      *discordgo.VoiceConnection
	Session    *discordgo.Session
	encoder    *dca.EncodeSession
	stream     *dca.StreamingSession
	queueMutex sync.Mutex
	Queue      []Song // First song in the queue is the current song
	GuildID    string
	Playing    bool
	Paused     bool
	stop       bool // Means clearing the queue
	Skip       bool // Maybe not needed
}

type Song struct {
	ChannelID   string
	User        string // Who requested the song
	Thumbnail   string
	ChannelName string
	Title       string
	YoutubeURL  string
	StreamURL   string
}

func (vi *VoiceInstance) playingStarted() {
	vi.Playing = true
	vi.Paused = false
	vi.Skip = false
}
func (vi *VoiceInstance) playingStopped() {
	vi.stop = false
	vi.Skip = false
	vi.Playing = false
}

// Plays the Queue
func (vi *VoiceInstance) PlayQueue() {

	for {
		malm.Debug("Starting a song")
		vi.playingStarted()

		if err := vi.Voice.Speaking(true); err != nil {
			malm.Error("%s", err)
			return
		}

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

		if vi.QueueIsEmpty() {
			return
		}

		vi.playingStopped()

		err = vi.Voice.Speaking(false)
		if err != nil {
			malm.Error("%s", err)
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

	// Be aware. This function is slow. Can take up to 2 seconds
	err = execYoutubeDL(&song)
	if err != nil {
		return err
	}

	vi.encoder, err = dca.EncodeFile(song.StreamURL, settings)
	if err != nil {
		return err
	}
	done := make(chan error)
	vi.stream = dca.NewStream(vi.encoder, vi.Voice, done)

	for err := range done {
		if err != nil && err != io.EOF {
			return err
		}

		vi.encoder.Cleanup()
	}

	return nil
}

func (vi *VoiceInstance) AddToQueue(s Song) {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	vi.Queue = append(vi.Queue, s)
}

func (vi *VoiceInstance) ClearQueue() {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	vi.Queue = []Song{}
}

func (vi *VoiceInstance) RemoveFirstInQueue() {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()

	// Only one entry in the queue, so clear it
	if len(vi.Queue) == 1 {
		vi.Queue = []Song{}
		return
	}
	vi.Queue = vi.Queue[1:]
}

func (vi *VoiceInstance) GetFirstInQueue() (Song, error) {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	if len(vi.Queue) == 0 {

		return Song{}, errors.New("the queue is empty")
	}
	return vi.Queue[0], nil
}

func (vi *VoiceInstance) QueueIsEmpty() bool {
	return len(vi.Queue) == 0
}

// Stops the current song
func (vi *VoiceInstance) Stop() {
	vi.stop = true
	if vi.encoder != nil {
		vi.encoder.Cleanup()
	}
}

// Disconnect dissconnects the bot from the voice connection
func (vi *VoiceInstance) Disconnect() {
	vi.Stop()
	time.Sleep(200 * time.Millisecond)

	err := vi.Voice.Disconnect()
	if err != nil {
		malm.Error("%s", err)
		return
	}
}
