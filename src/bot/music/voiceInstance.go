package music

import (
	"errors"
	"io"
	"log"
	"sync"

	"github.com/CarlFlo/dootBot/src/bot/context"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
	"github.com/jung-m/dca"
)

type VoiceInstance struct {
	voice            *discordgo.VoiceConnection
	encoder          *dca.EncodeSession
	stream           *dca.StreamingSession
	mu               sync.RWMutex
	queue            []*Song
	guildID          string
	done             chan error
	messageID        string
	messageChannelID string
	PlaybackState
}

type PlaybackState struct {
	workerRunning bool
	loading       bool
	paused        bool
	looping       bool
	stopRequested bool
	previous      bool
	queueIndex    int
}

func (vi *VoiceInstance) New(guildID string) error {
	vi.guildID = guildID
	return nil
}

func (vi *VoiceInstance) Close() {
	vi.requestStop()
	vi.PurgeQueue()
	vi.deleteOverviewMessage()
}

func (vi *VoiceInstance) PlayQueue() {
	if !vi.startWorker() {
		return
	}

	dca.Logger = log.New(io.Discard, "", 0)
	defer vi.finishWorker()

	for {
		song, err := vi.prepareCurrentSong()
		if err != nil {
			if errors.Is(err, errEmptyQueue) || errors.Is(err, errNoNextSong) {
				return
			}
			malm.Error("music playback preparation failed: %s", err)
			return
		}

		if err := vi.setSpeaking(true); err != nil {
			malm.Error("%s", err)
			return
		}

		if err := vi.streamSong(song); err != nil {
			_ = vi.setSpeaking(false)
			malm.Error("music stream failed: %s", err)
			return
		}

		if err := vi.setSpeaking(false); err != nil {
			malm.Error("%s", err)
			return
		}

		if vi.shouldStop() {
			vi.PurgeQueue()
			return
		}

		vi.FinishedPlayingSong()
	}
}

func (vi *VoiceInstance) prepareCurrentSong() (*Song, error) {
	vi.setLoading(true)
	vi.setPaused(false)

	song, err := vi.GetFirstInQueue()
	if err != nil {
		vi.setLoading(false)
		_ = vi.refreshOverviewMessage()
		return nil, err
	}

	if err := song.FetchStreamURL(); err != nil {
		vi.setLoading(false)
		_ = vi.refreshOverviewMessage()
		return nil, err
	}

	vi.setLoading(false)
	if err := vi.refreshOverviewMessage(); err != nil {
		malm.Error("unable to refresh music overview: %s", err)
	}

	return song, nil
}

func (vi *VoiceInstance) streamSong(song *Song) error {
	settings := dca.StdEncodeOptions
	settings.RawOutput = true
	settings.Bitrate = 64
	settings.Application = "lowdelay"

	encoder, err := dca.EncodeFile(song.StreamURL, settings)
	if err != nil {
		return err
	}
	defer encoder.Cleanup()

	done := make(chan error, 1)
	stream := dca.NewStream(encoder, vi.voice, done)

	vi.mu.Lock()
	vi.encoder = encoder
	vi.stream = stream
	vi.done = done
	vi.mu.Unlock()

	defer vi.clearStreamSession()

	for {
		err := <-done
		if err != nil && err != io.EOF {
			return err
		}
		return nil
	}
}

func (vi *VoiceInstance) FinishedPlayingSong() {
	vi.mu.Lock()
	defer vi.mu.Unlock()

	if vi.previous {
		vi.previous = false
		return
	}

	if vi.looping {
		return
	}

	if vi.queueIndex < len(vi.queue) {
		vi.queueIndex++
	}
}

func (vi *VoiceInstance) IncrementQueueIndex() bool {
	vi.mu.Lock()
	defer vi.mu.Unlock()

	if vi.queueIndex >= len(vi.queue) {
		return false
	}

	vi.queueIndex++
	return true
}

func (vi *VoiceInstance) DecrementQueueIndex() bool {
	vi.mu.Lock()
	defer vi.mu.Unlock()

	if vi.queueIndex == 0 {
		return false
	}

	vi.queueIndex--
	return true
}

func (vi *VoiceInstance) isEndOfQueue() bool {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return vi.queueIndex >= len(vi.queue)
}

func (vi *VoiceInstance) Disconnect() {
	vi.requestStop()

	vi.mu.RLock()
	voice := vi.voice
	vi.mu.RUnlock()

	if voice == nil {
		return
	}

	if err := voice.Disconnect(); err != nil {
		malm.Error("%s", err)
	}
}

func (vi *VoiceInstance) Skip() bool {
	vi.mu.Lock()
	running := vi.workerRunning
	vi.looping = false
	vi.mu.Unlock()

	if !running {
		return false
	}

	vi.signalDone(nil)
	return true
}

func (vi *VoiceInstance) Prev() bool {
	vi.mu.Lock()
	if !vi.workerRunning {
		vi.mu.Unlock()
		return false
	}

	if vi.queueIndex > 0 {
		vi.queueIndex--
	}
	vi.previous = true
	vi.mu.Unlock()

	vi.signalDone(nil)
	return true
}

func (vi *VoiceInstance) Stop() bool {
	vi.mu.Lock()
	running := vi.workerRunning
	if running {
		vi.stopRequested = true
	}
	vi.mu.Unlock()

	if !running {
		return false
	}

	vi.signalDone(nil)
	return true
}

func (vi *VoiceInstance) refreshOverviewMessage() error {
	vi.mu.RLock()
	channelID := vi.messageChannelID
	messageID := vi.messageID
	vi.mu.RUnlock()

	if channelID == "" || messageID == "" {
		return nil
	}

	components := []discordgo.MessageComponent{}
	embeds := []*discordgo.MessageEmbed{}

	msgEdit := &discordgo.MessageEdit{
		Channel:    channelID,
		ID:         messageID,
		Components: &components,
		Embeds:     &embeds,
	}

	CreateMusicOverviewMessage(channelID, msgEdit)
	_, err := context.SESSION.ChannelMessageEditComplex(msgEdit)
	return err
}

func (vi *VoiceInstance) deleteOverviewMessage() {
	vi.mu.RLock()
	channelID := vi.messageChannelID
	messageID := vi.messageID
	vi.mu.RUnlock()

	if channelID == "" || messageID == "" {
		return
	}

	if err := context.SESSION.ChannelMessageDelete(channelID, messageID); err != nil {
		malm.Debug("unable to delete music overview message: %s", err)
	}
}

func (vi *VoiceInstance) setSpeaking(speaking bool) error {
	vi.mu.RLock()
	voice := vi.voice
	vi.mu.RUnlock()

	if voice == nil {
		return errors.New("voice connection is not initialized")
	}

	return voice.Speaking(speaking)
}

func (vi *VoiceInstance) clearStreamSession() {
	vi.mu.Lock()
	vi.encoder = nil
	vi.stream = nil
	vi.done = nil
	vi.mu.Unlock()
}
