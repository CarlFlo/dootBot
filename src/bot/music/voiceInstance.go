package music

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v4/lavalink"
)

type VoiceInstance struct {
	mu               sync.RWMutex
	queue            []*Song
	guildID          string
	voiceChannelID   string
	messageID        string
	messageChannelID string
	PlaybackState
}

type PlaybackState struct {
	workerRunning bool
	loading       bool
	paused        bool
	looping       bool
	queueIndex    int
}

func (vi *VoiceInstance) New(guildID string) error {
	vi.guildID = guildID
	return nil
}

func (vi *VoiceInstance) Close() {
	vi.PurgeQueue()
	vi.deleteOverviewMessage()
}

func (vi *VoiceInstance) FinishedPlayingSong() {
	vi.mu.Lock()
	defer vi.mu.Unlock()

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
	vi.mu.RLock()
	voiceChannelID := vi.voiceChannelID
	vi.mu.RUnlock()

	if voiceChannelID == "" {
		return
	}

	manager.disconnectVoice(context.Background(), vi.guildID)

	vi.mu.Lock()
	vi.voiceChannelID = ""
	vi.workerRunning = false
	vi.loading = false
	vi.paused = false
	vi.mu.Unlock()
}

func (vi *VoiceInstance) Skip() bool {
	vi.mu.RLock()
	running := vi.workerRunning
	vi.mu.RUnlock()
	if !running {
		return false
	}

	vi.mu.Lock()
	hasNext := vi.queueIndex+1 < len(vi.queue)
	if hasNext {
		vi.queueIndex++
	} else {
		vi.queueIndex = len(vi.queue)
		vi.workerRunning = false
		vi.paused = false
	}
	vi.mu.Unlock()

	if hasNext {
		return manager.playCurrentSong(context.Background(), vi) == nil
	}

	if err := manager.stopPlayback(context.Background(), vi.guildID); err != nil {
		return false
	}

	_ = vi.refreshOverviewMessage()
	return true
}

func (vi *VoiceInstance) Prev() bool {
	vi.mu.RLock()
	running := vi.workerRunning
	vi.mu.RUnlock()
	if !running {
		return false
	}

	vi.DecrementQueueIndex()
	return manager.playCurrentSong(context.Background(), vi) == nil
}

func (vi *VoiceInstance) Stop() bool {
	vi.mu.RLock()
	running := vi.workerRunning
	vi.mu.RUnlock()
	if !running {
		return false
	}

	if err := manager.stopPlayback(context.Background(), vi.guildID); err != nil {
		return false
	}

	vi.mu.Lock()
	vi.workerRunning = false
	vi.paused = false
	vi.loading = false
	vi.mu.Unlock()
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
	return manager.editOverviewMessage(msgEdit)
}

func (vi *VoiceInstance) deleteOverviewMessage() {
	vi.mu.RLock()
	channelID := vi.messageChannelID
	messageID := vi.messageID
	vi.mu.RUnlock()

	if channelID == "" || messageID == "" {
		return
	}

	manager.deleteOverviewMessage(channelID, messageID)
}

func (vi *VoiceInstance) handleTrackStarted() {
	vi.mu.Lock()
	vi.workerRunning = true
	vi.loading = false
	vi.paused = false
	vi.mu.Unlock()
}

func (vi *VoiceInstance) handleTrackEnded(reason lavalink.TrackEndReason) (bool, error) {
	vi.mu.Lock()
	defer vi.mu.Unlock()

	if vi.looping && reason.MayStartNext() {
		return true, nil
	}

	if !reason.MayStartNext() {
		return false, nil
	}

	if vi.queueIndex < len(vi.queue) {
		vi.queueIndex++
	}

	if vi.queueIndex >= len(vi.queue) {
		vi.workerRunning = false
		vi.loading = false
		vi.paused = false
		return false, nil
	}

	return true, nil
}

func (vi *VoiceInstance) currentSong() (*Song, error) {
	song, err := vi.GetFirstInQueue()
	if err != nil {
		return nil, err
	}
	if song.Track.Encoded == "" {
		return nil, errors.New("song is missing an encoded lavalink track")
	}
	return song, nil
}

func (vi *VoiceInstance) markLoading(loading bool) {
	vi.mu.Lock()
	vi.loading = loading
	vi.mu.Unlock()
}

func (vi *VoiceInstance) setVoiceChannelID(channelID string) {
	vi.mu.Lock()
	vi.voiceChannelID = channelID
	vi.mu.Unlock()
}

func (vi *VoiceInstance) VoiceChannelID() string {
	vi.mu.RLock()
	defer vi.mu.RUnlock()
	return vi.voiceChannelID
}

func (vi *VoiceInstance) ensureQueuePlayable() error {
	if _, err := vi.currentSong(); err != nil {
		return fmt.Errorf("there is no song to play: %w", err)
	}
	return nil
}
