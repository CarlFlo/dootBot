package music

import (
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/CarlFlo/dootBot/src/bot/context"
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
	paused     bool
	loading    bool
	stop       bool
	looping    bool
	queueIndex int
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

func (vi *VoiceInstance) playingStarted() {
	vi.playing = true
	vi.paused = false
	vi.loading = true
}
func (vi *VoiceInstance) playingStopped() {
	vi.stop = false
	vi.playing = false
	vi.loading = false
}

// Plays the Queue
func (vi *VoiceInstance) PlayQueue() {

	// This suppresses the warning from dca:
	// 'Error parsing ffmpeg stats: strconv.ParseFloat: parsing "N": invalid syntax'
	dca.Logger = log.New(io.Discard, "", 0)

	defer vi.playingStopped()

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
		vi.FinishedPlayingSong()

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

// #### Queue Code ####

func (vi *VoiceInstance) GetFirstInQueue() (Song, error) {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	if vi.GetQueueLength() == 0 {
		return Song{}, errEmptyQueue
	} else if vi.isEndOfQueue() {
		return Song{}, errNoNextSong
	}

	return vi.queue[vi.queueIndex], nil
}

func (vi *VoiceInstance) AddToQueue(s Song) {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	vi.queue = append(vi.queue, s)
}

// Removes all songs in the queue after the current song.
func (vi *VoiceInstance) ClearQueue() {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
	vi.queue = vi.queue[:vi.queueIndex+1]
}

// Removes all songs in the queue before the current song.
func (vi *VoiceInstance) ClearQueuePrev() {
	vi.queueMutex.Lock()
	defer vi.queueMutex.Unlock()
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
func (vi *VoiceInstance) GetSongByIndex(i int) Song {
	return vi.queue[i]
}

//////////////////////////// Queue code end ////////////////////////////

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
func (vi *VoiceInstance) Skip() bool {

	if !vi.playing {
		return false
	}

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

func (vi *VoiceInstance) IsPaused() bool {
	return vi.paused
}

func (vi *VoiceInstance) IsLooping() bool {
	return vi.looping
}

func (vi *VoiceInstance) SetLooping(loop bool) {
	vi.looping = loop
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

// Toggles between play and pause
func (vi *VoiceInstance) PauseToggle() {

	vi.paused = !vi.paused
	vi.stream.SetPaused(vi.paused)
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
