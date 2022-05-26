package music

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type voiceInstance struct {
	Voice      *discordgo.VoiceConnection
	Session    *discordgo.Session
	Encoder    interface{} // temp
	Stream     interface{} // temp
	queueMutex sync.Mutex
	Queue      []song // First song in the queue is the current song
	GuildID    string
	Playing    bool
	Paused     bool // Pause and stop maybe the same thing here
	Stop       bool
	Skip       bool // Maybe not needed
}

type song struct {
	VoiceChannelID string
	User           string // Who requested the song
	Title          string
	VideoURL       string
}

func (vi *voiceInstance) AddToQueue(s song) {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	vi.Queue = append(vi.Queue, s)
}

func (vi *voiceInstance) ClearQueue(s song) {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	vi.Queue = []song{}
}

func (vi *voiceInstance) GetFirstInQueue(s song) (song, error) {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	if len(vi.Queue) == 0 {
		return song{}, fmt.Errorf("queue is empty")
	}
	return vi.Queue[0], nil
}

func (vi *voiceInstance) RemoveFirstInQueue() {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()

	// Only one entry in the queue, so clear it
	if len(vi.Queue) == 1 {
		vi.Queue = []song{}
		return
	}
	vi.Queue = vi.Queue[1:]
}
