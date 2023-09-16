package music

import (
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/CarlFlo/dootBot/src/bot/context"
	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
	"github.com/jung-m/dca"
)

var (
	errEmptyQueue = errors.New("the queue is empty")
	errNoNextSong = errors.New("there is no next song to play")
)

type VoiceInstance struct {
	voice      *discordgo.VoiceConnection
	encoder    *dca.EncodeSession
	stream     *dca.StreamingSession
	queueMutex sync.Mutex
	queue      []Song
	guildID    string
	done       chan error // Used to interrupt the stream
	messageID  string
	channelID  string
	DJ
}

// The variables keeping track of the playback state
type DJ struct {
	playing    bool
	loading    bool
	stop       bool // stopping the music bot, removing the queue etc
	looping    bool
	queueIndex int
}

type songStreamCacheWrapper struct {
	mu        sync.Mutex
	songCache map[string]songStreamCache
}

type songStreamCache struct {
	streamURL string
	expires   time.Time
}

var songCache = songStreamCacheWrapper{
	songCache: make(map[string]songStreamCache),
}

// Adding a duplicate will overwrite the old one
func (c *songStreamCacheWrapper) Add(song *Song) {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.songCache[song.YoutubeVideoID] = songStreamCache{
		streamURL: song.StreamURL,
		expires:   time.Now().Add(time.Minute * config.CONFIG.Music.MaxCacheAgeMin), // Valid for 90 minutes, 1h 30 min
	}
}

func (c *songStreamCacheWrapper) Check(ytURL string) string {

	ssc := c.songCache[ytURL]
	if time.Now().After(ssc.expires) {
		// remove from map
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.songCache, ytURL)
		return ""
	}
	return ssc.streamURL
}

// New creates a new VoiceInstance. Remember to call Close() when before deleting the object
func (vi *VoiceInstance) New(guildID string) error {
	vi.guildID = guildID
	return nil
}

// Close acts as the destructor for the object
func (vi *VoiceInstance) Close() {

	// Delete the interaction buttons from the message

	// For now will delete the message
	context.SESSION.ChannelMessageDelete(vi.GetChannelID(), vi.GetMessageID())
	malm.Info("Ending music session. Deleted the bot message")
}

func (vi *VoiceInstance) playbackStarted() {
	vi.stop = false
	vi.playing = true
	vi.loading = true
}
func (vi *VoiceInstance) playbackStopped() {
	vi.stop = false
	vi.playing = false
	vi.loading = false
}

// Plays the Queue
func (vi *VoiceInstance) PlayQueue() {

	// This suppresses the warning from dca:
	// 'Error parsing ffmpeg stats: strconv.ParseFloat: parsing "N": invalid syntax'
	dca.Logger = log.New(io.Discard, "", 0)

	defer vi.playbackStopped()

	for {
		vi.playbackStarted()

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
		vi.FinishedPlayingSong()

		vi.playbackStopped()

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

	// song.StreamURL contains the URL to the stream.

	if streamURL := songCache.Check(song.YoutubeVideoID); len(streamURL) == 0 {
		// This function is slow. At least 2 seconds
		err = execYoutubeDL(song)
		if err != nil {
			return fmt.Errorf("[Youtube Downloader] %v", err)
		}

		songCache.Add(song)
	} else {
		song.StreamURL = streamURL
	}

	vi.encoder, err = dca.EncodeFile(song.StreamURL, settings)
	if err != nil {
		return err
	}

	vi.done = make(chan error)
	vi.stream = dca.NewStream(vi.encoder, vi.voice, vi.done)

	// Update the message to reflect that the song is playing

	vi.loading = false
	msgEdit := &discordgo.MessageEdit{
		Channel: vi.GetChannelID(),
		ID:      vi.GetMessageID(),
	}

	CreateMusicOverviewMessage(vi.GetChannelID(), msgEdit)

	if _, err := context.SESSION.ChannelMessageEditComplex(msgEdit); err != nil {
		malm.Error("cannot create message edit, error: %s", err)
	}

	// This code streams the audio
	// Once the song is finished, stopped or skipped so will this function return
	for { // (Do not use a range here as it does not work for this purpose)
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

// Call once the song has finished playing, or you want to skip to the next song
func (vi *VoiceInstance) FinishedPlayingSong() {

	if vi.IsLooping() {
		return
	}
	vi.IncrementQueueIndex()
}

// TODO: When at the end of queue. Should increment one more
func (vi *VoiceInstance) IncrementQueueIndex() bool {

	// Do not increment past the end of the queue
	if vi.isEndOfQueue() {
		return false
	}
	vi.queueIndex++
	return true
}

// Returns true if the index could be decremented
func (vi *VoiceInstance) DecrementQueueIndex() bool {

	if vi.queueIndex == 0 {
		return false
	}
	vi.queueIndex--
	return true
}

func (vi *VoiceInstance) isEndOfQueue() bool {
	return vi.GetQueueLength() == vi.queueIndex
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
// Will also disable looping (if enabled)
func (vi *VoiceInstance) Skip() bool {

	if !vi.playing {
		return false
	}

	vi.looping = false

	// This will interupt and stop the stream
	vi.done <- nil

	return true
}

func (vi *VoiceInstance) Prev() bool {
	if !vi.playing || !vi.DecrementQueueIndex() {
		// Music is not playing or there is no song to go back to
		return false
	}

	// This will interupt and stop the stream
	vi.done <- nil

	return true
}

func (vi *VoiceInstance) IsLoading() bool {
	return vi.loading
}

func (vi *VoiceInstance) IsPlaying() bool {
	return vi.playing
}

func (vi *VoiceInstance) IsLooping() bool {
	return vi.looping
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

func (vi *VoiceInstance) ToggleLooping() {

	vi.looping = !vi.looping
}

// Toggles between play and pause
func (vi *VoiceInstance) PauseToggle() {

	vi.playing = !vi.playing
	vi.stream.SetPaused(vi.playing)
}

func (vi *VoiceInstance) GetGuildID() string {
	return vi.guildID
}

func (vi *VoiceInstance) SetMessageID(id string) {
	vi.messageID = id
}

func (vi *VoiceInstance) GetMessageID() string {
	return vi.messageID
}

func (vi *VoiceInstance) SetChannelID(id string) {
	vi.channelID = id
}

func (vi *VoiceInstance) GetChannelID() string {
	return vi.channelID
}
