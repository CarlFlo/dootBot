package music

import (
	"io"
	"log"
	"sync"
	"time"

	"github.com/CarlFlo/dootBot/src/bot/context"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
	"github.com/jung-m/dca"
)

/*
	todo
	Fix message update (rewrite)
	What to do once the last song has played
*/

type VoiceInstance struct {
	voice            *discordgo.VoiceConnection
	encoder          *dca.EncodeSession
	stream           *dca.StreamingSession
	mu               sync.Mutex
	queue            []*Song
	guildID          string
	done             chan error // Used to interrupt the stream
	messageID        string
	messageChannelID string
	PlaybackState
}

// The variables keeping track of the playback state
type PlaybackState struct {
	playing    bool
	loading    bool
	stop       bool // stopping the music bot, removing the queue etc
	looping    bool
	previous   bool // indicates the user wants to go back and play the previous song
	queueIndex int
}

// New creates a new VoiceInstance. Remember to call 'vi.Close()' before deleting the object
func (vi *VoiceInstance) New(guildID string) error {
	vi.guildID = guildID
	return nil
}

// Close acts as the destructor for the object
func (vi *VoiceInstance) Close() {

	// Clean-up here
	vi.PurgeQueue()
	vi.stop = true
	go func() {
		vi.done <- nil
	}()

	// For now will delete the message
	context.SESSION.ChannelMessageDelete(vi.GetMessageChannelID(), vi.GetMessageID())
}

// Plays the Queue
func (vi *VoiceInstance) PlayQueue() {

	// This suppresses the warning from dca:
	// 'Error parsing ffmpeg stats: strconv.ParseFloat: parsing "N": invalid syntax'
	dca.Logger = log.New(io.Discard, "", 0)

	defer vi.playbackStopped()

	for {

		//TODO: move check to see if there is a new song next in queue. If no song is in the queue. Then wait for n minutes until there is on. Else leave vc

		vi.playbackStarted()

		if err := vi.voice.Speaking(true); err != nil {
			malm.Error("%s", err)
			return
		}

		if err := vi.StreamAudioToVoiceChannel(); err != nil {
			malm.Error("%s", err)
			return
		}

		if vi.stop {
			vi.PurgeQueue()
			return
		}

		vi.FinishedPlayingSong()

		vi.playbackStopped()

		if err := vi.voice.Speaking(false); err != nil {
			malm.Error("%s", err)
			return
		}

		if vi.QueueIsEmpty() {
			return
		}
	}
}

func (vi *VoiceInstance) StreamAudioToVoiceChannel() error {

	settings := dca.StdEncodeOptions
	// Custom settings
	settings.RawOutput = true
	settings.Bitrate = 64
	settings.Application = "lowdelay"

	// TODO: Wait for next song. Currently will just exit the voice instance and will force the user to run the 'play' again
	song, err := vi.GetFirstInQueue()
	if err != nil {
		return err
	}

	// Waiting until the song has a streamURL
	for song.StreamURL == "" {
		time.Sleep(100 * time.Millisecond)
	}

	vi.encoder, err = dca.EncodeFile(song.StreamURL, settings)
	if err != nil {
		return err
	}
	defer vi.encoder.Cleanup()

	vi.done = make(chan error)
	vi.stream = dca.NewStream(vi.encoder, vi.voice, vi.done)

	// Update the message to reflect that the song is playing

	vi.loading = false
	msgEdit := &discordgo.MessageEdit{
		Channel: vi.GetMessageChannelID(),
		ID:      vi.GetMessageID(),
	}

	CreateMusicOverviewMessage(vi.GetMessageChannelID(), msgEdit)

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
			return nil
		}
	}

}

// Call once the song has finished playing, or you want to skip to the next song
// TODO: add check for if the user wants to play the song again. i.e. previous command
func (vi *VoiceInstance) FinishedPlayingSong() {

	// Indicates the user wants to go back and play the previous song
	if vi.previous {
		vi.previous = false
		return
	}

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

// Prev will go back to the previous song. If there is no song to go back to so will the song be restarted
// Running this command, if nothing is playing, should start playing the song
func (vi *VoiceInstance) Prev() bool {

	// TODO: Remove?
	if !vi.playing {
		// Music is not playing
		return false
	}

	vi.DecrementQueueIndex()

	vi.previous = true

	// This will interupt and stop the stream
	vi.done <- nil

	return true
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
