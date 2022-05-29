package music

import (
	"fmt"
	"strings"
	"sync"

	"github.com/CarlFlo/DiscordMoneyBot/src/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/src/config"
	"github.com/CarlFlo/DiscordMoneyBot/src/utils"
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

// TODO
// weird error when plaing music: 'Error parsing ffmpeg stats: strconv.ParseFloat: parsing "N": invalid syntax'

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

	malm.Info("Music initialized")
	youtubeAPIKeyPresent = true
	return nil
}

// Same as resume
func PlayMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	if !isMusicEnabled(s, m) {
		return
	}

	guildID := utils.GetGuild(s, m)
	vi := instances[guildID]
	if vi == nil {
		// Not initialized
		vi = joinVoice(vi, s, m)
		if vi == nil {
			malm.Error("Failed to join voice channel")
			return
		}
	}

	// Check if the user is in the voice channel before playing
	voiceChannelID := utils.FindVoiceChannel(s, m.Author.ID)
	if vi.voice.ChannelID != voiceChannelID {
		utils.SendMessageFailure(s, m, "You are not in the same voice channel as the bot")
		return
	}

	if input.NumberOfArgsAre(0) {
		// User want to resume a song
		if vi.paused {
			vi.paused = false
			vi.stream.SetPaused(vi.paused)
		}
		return
	}

	var song Song
	inputText := strings.Join(input.GetArgs(), " ")

	err := parseMusicInput(m, inputText, &song)
	if err != nil {
		malm.Error("%s", err)
		utils.SendMessageFailure(s, m, fmt.Sprintf("Something went wrong when getting the song. The maximum duration for a song is %d", config.CONFIG.Music.MaxSongLengthMinutes))
		return
	}

	// Add the song to the queue
	vi.AddToQueue(song)

	utils.SendMessageNeutral(s, m, fmt.Sprintf("%s added the song ``%s`` to the queue (%s)", m.Author.Username, song.Title, song.duration))

	complexMessage := &discordgo.MessageSend{}

	CreateMusicOverviewMessage(s, m, complexMessage, nil)

	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}

	// The bot is already playing music so we dont send the start signal
	if !vi.IsPlaying() {
		songSignal <- vi
	}
}

func StopMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled(s, m) {
		return
	}

	guildID := utils.GetGuild(s, m)
	vi := instances[guildID]

	if vi == nil {
		// Nothing is playing
		return
	}

	leaveVoice(vi, s, m)
}

func SkipMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled(s, m) {
		return
	}

	guildID := utils.GetGuild(s, m)
	vi := instances[guildID]

	if vi == nil {
		// Nothing is playing
		return
	}

	if vi.Skip() {
		utils.SendMessageSuccess(s, m, "Skipped the song")
	} else {
		utils.SendMessageFailure(s, m, "There is no song to skip")
	}
}

// ClearQueueMusic clears but the first song from the queue
func ClearQueueMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled(s, m) {
		return
	}

	guildID := utils.GetGuild(s, m)
	vi := instances[guildID]

	if vi == nil {
		// Nothing is playing
		return
	}

	vi.RemoveAllButFirstInQueue()
	vi.Stop() // Should it stop the bot?
}

// PauseMusic pauyses the music
func PauseMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled(s, m) {
		return
	}

	guildID := utils.GetGuild(s, m)
	vi := instances[guildID]

	if vi == nil {
		// Nothing is playing
		return
	}

	vi.Pause()
}
