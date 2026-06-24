package music

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

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
	refreshTicker    *time.Ticker
	refreshStop      chan struct{}
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
	queueIndex := vi.queueIndex
	queueLen := len(vi.queue)
	running := vi.workerRunning
	vi.mu.RUnlock()

	if queueLen == 0 {
		return false
	}

	if !running && !(queueIndex >= queueLen && queueLen > 0) {
		return false
	}

	vi.mu.Lock()
	if queueIndex >= queueLen {
		vi.queueIndex = queueLen - 1
	} else if queueIndex > 0 {
		vi.queueIndex--
	} else {
		vi.mu.Unlock()
		return false
	}
	vi.mu.Unlock()

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

	vi.stopOverviewRefreshLoop()
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
	applyMusicOverviewMessage(vi, msgEdit)
	return manager.editOverviewMessage(msgEdit)
}

func (vi *VoiceInstance) deleteOverviewMessage() {
	vi.stopOverviewRefreshLoop()

	vi.mu.Lock()
	channelID := vi.messageChannelID
	messageID := vi.messageID
	vi.messageChannelID = ""
	vi.messageID = ""
	vi.mu.Unlock()

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
	vi.startOverviewRefreshLoop()
}

func (vi *VoiceInstance) handleTrackEnded(reason lavalink.TrackEndReason) (bool, error) {
	vi.mu.Lock()
	stopRefresh := false

	if vi.looping && reason.MayStartNext() {
		vi.mu.Unlock()
		return true, nil
	}

	if !reason.MayStartNext() {
		stopRefresh = true
		vi.mu.Unlock()
		if stopRefresh {
			go vi.stopOverviewRefreshLoop()
		}
		return false, nil
	}

	if vi.queueIndex < len(vi.queue) {
		vi.queueIndex++
	}

	if vi.queueIndex >= len(vi.queue) {
		vi.workerRunning = false
		vi.loading = false
		vi.paused = false
		stopRefresh = true
		vi.mu.Unlock()
		if stopRefresh {
			go vi.stopOverviewRefreshLoop()
		}
		return false, nil
	}

	vi.mu.Unlock()
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

func (vi *VoiceInstance) currentSongElapsed() (time.Duration, error) {
	player, err := manager.playerForGuild(vi.guildID)
	if err != nil {
		return 0, err
	}

	// Lavalink reports player position in milliseconds.
	elapsed := time.Duration(player.Position().Milliseconds()) * time.Millisecond
	if elapsed < 0 {
		elapsed = 0
	}

	song, err := vi.GetFirstInQueue()
	if err != nil {
		return 0, err
	}
	if elapsed > song.Duration {
		elapsed = song.Duration
	}

	return elapsed, nil
}

func (vi *VoiceInstance) currentSongProgress() (time.Duration, time.Duration, error) {
	elapsed, err := vi.currentSongElapsed()
	if err != nil {
		return 0, 0, err
	}

	song, err := vi.GetFirstInQueue()
	if err != nil {
		return elapsed, 0, err
	}

	total := song.Duration
	remaining := total - elapsed
	if remaining < 0 {
		remaining = 0
	}
	return elapsed, remaining, nil
}

func (vi *VoiceInstance) startOverviewRefreshLoop() {
	vi.mu.Lock()
	if vi.refreshTicker != nil {
		vi.mu.Unlock()
		return
	}

	ticker := time.NewTicker(10 * time.Second) // How often the music player updates its time elapsed for the current song
	stop := make(chan struct{})
	vi.refreshTicker = ticker
	vi.refreshStop = stop
	vi.mu.Unlock()

	go func() {
		for {
			select {
			case <-ticker.C:
				_ = vi.refreshOverviewMessage()
			case <-stop:
				return
			}
		}
	}()
}

func (vi *VoiceInstance) stopOverviewRefreshLoop() {
	vi.mu.Lock()
	ticker := vi.refreshTicker
	stop := vi.refreshStop
	vi.refreshTicker = nil
	vi.refreshStop = nil
	vi.mu.Unlock()

	if ticker != nil {
		ticker.Stop()
	}
	if stop != nil {
		close(stop)
	}
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
