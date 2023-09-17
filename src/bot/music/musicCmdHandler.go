package music

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/CarlFlo/dootBot/src/bot/context"
	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

// GuildID is the key
var instances = map[string]*VoiceInstance{}

/*
	Play songs in a voice channel
	Commands:
	play (plays a song or adds the song to the queue if something is playing), resume, skip, stop, pause, playlist (ability to create a personal playlist, adds songs with buttons etc)

	playlist: dropdown menu with selections of playlists in the guild

	Save stats in DB for songs played, skiped
	Only save:

	https://www.youtube.com/watch?v=5qap5aO4i9A -> 5qap5aO4i9A
	To save storage, in DB
*/

var (
	musicMutex           sync.Mutex
	songSignal           chan *VoiceInstance
	youtubeAPIKeyPresent bool
)

const (
	youtubePattern string = `(youtube\.com\/watch\?v=)`
	urlPattern     string = `[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`
)

func Initialize() {
	if err := InitializeMusic(); err != nil {
		malm.Info("Music disabled. %s", err.Error())
		return
	}
	malm.Info("Music initialized")
}

// InitializeMusic initializes the music goroutine and channel signal
func InitializeMusic() error {

	if err := utils.ValidateYoutubeAPIKey(); err != nil {
		youtubeAPIKeyPresent = false
		return err
	}

	songSignal = make(chan *VoiceInstance)

	go func() {
		for vi := range songSignal {
			go vi.PlayQueue()
		}
	}()

	youtubeAPIKeyPresent = true
	return nil
}

// Same as resume
func PlayMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	if !isMusicEnabled() {
		utils.SendMessageNeutral(m, "Music is currently disabled")
		return
	}

	guildID, err := utils.GetGuild(m.ChannelID)
	if err != nil {
		malm.Error("Error getting guild ID: %s", err)
		return
	}

	vi := instances[guildID]
	var errStr string
	if vi == nil {
		// Not initialized
		vi, errStr = joinVoice(vi, m.Author.ID, m.ChannelID)
		if vi == nil {
			utils.SendMessageFailure(m, errStr)
			return
		}
	}

	// Check if the user is in the voice channel before playing
	voiceChannelID := utils.FindVoiceChannel(m.Author.ID)
	if vi.voice.ChannelID != voiceChannelID {
		utils.SendMessageFailure(m, "You are not in the same voice channel as the bot")
		return
	}

	if input.NumberOfArgsAre(0) {
		// User want to resume a song
		if !vi.playing {
			vi.playing = true
			vi.stream.SetPaused(vi.playing)
		}
		return
	}

	var song Song
	inputText := strings.Join(input.GetArgs(), " ")

	err = parseMusicInput(m, inputText, &song)
	if err != nil {
		malm.Error("%s", err)
		utils.SendMessageFailure(m, fmt.Sprintf("Something went wrong when getting the song.\nNote: the maximum duration for a song is currently set at %d minutes in the configuration file", config.CONFIG.Music.MaxSongLengthMinutes))
		return
	}

	// Add the song to the queue
	vi.AddToQueue(&song)
	go song.FetchStreamURL()

	addedSongMsg, _ := utils.SendMessageNeutral(m, fmt.Sprintf("%s added the song ``%s`` to the queue (%s)", m.Author.Username, song.Title, song.duration))

	go func() {
		for range time.After(time.Second * 5) {
			context.SESSION.ChannelMessageDelete(m.ChannelID, addedSongMsg.ID)
		}
	}()

	complexMessage := &discordgo.MessageSend{}

	if !vi.IsPlaying() {
		vi.loading = true
	}

	CreateMusicOverviewMessage(m.ChannelID, complexMessage)

	msg, err := context.SESSION.ChannelMessageSendComplex(m.ChannelID, complexMessage)
	if err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}
	vi.SetMessageID(msg.ID)
	vi.SetChannelID(msg.ChannelID)

	// The bot is already playing music so we dont send the start signal
	if !vi.IsPlaying() {
		songSignal <- vi
	}
}

func StopMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled() {
		utils.SendMessageNeutral(m, "Music is currently disabled")
		return
	}

	guildID, err := utils.GetGuild(m.ChannelID)
	if err != nil {
		malm.Error("Error getting guild ID: %s", err.Error())
		return
	}

	vi := instances[guildID]

	if vi == nil {
		// Nothing is playing
		return
	}

	leaveVoice(vi)
}

func SkipMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled() {
		utils.SendMessageNeutral(m, "Music is currently disabled")
		return
	}

	guildID, err := utils.GetGuild(m.ChannelID)
	if err != nil {
		malm.Error("Error getting guild ID: %s", err.Error())
		return
	}
	vi := instances[guildID]

	if vi == nil {
		// Nothing is playing
		return
	}

	if vi.Skip() {
		utils.SendMessageSuccess(m, "Skipped the song")
	} else {
		utils.SendMessageFailure(m, "There is no song to skip")
	}
}

// ClearQueueMusic clears the queue. Does not include the current song or previus songs
func ClearQueueMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled() {
		utils.SendMessageNeutral(m, "Music is currently disabled")
		return
	}

	guildID, err := utils.GetGuild(m.ChannelID)
	if err != nil {
		malm.Error("Error getting guild ID: %s", err.Error())
		return
	}
	vi := instances[guildID]

	if vi == nil {
		// Nothing is playing
		return
	}

	vi.ClearQueue()
	//vi.Stop() // Should it stop the bot?
}

// PauseMusic toggles the music from playing to pausing
func PauseMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled() {
		utils.SendMessageNeutral(m, "Music is currently disabled")
		return
	}

	guildID, err := utils.GetGuild(m.ChannelID)
	if err != nil {
		malm.Error("Error getting guild ID: %s", err.Error())
		return
	}
	vi := instances[guildID]

	if vi == nil {
		// Nothing is playing
		return
	}

	vi.PauseToggle()
}

func MusicPrevious(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	if !isMusicEnabled() {
		utils.SendMessageNeutral(m, "Music is currently disabled")
		return
	}

	guildID, err := utils.GetGuild(m.ChannelID)
	if err != nil {
		malm.Error("Error getting guild ID: %s", err.Error())
		return
	}

	vi := instances[guildID]

	if vi == nil { // Nothing is playing
		return
	}

	if !vi.Prev() {
		utils.SendMessageNeutral(m, "You are at the start of the queue")
	}
}
