package music

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
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
)

func InitializeMusic() {

	if len(config.CONFIG.Music.YoutubeAPIKey) == 0 {
		malm.Info("Music disabled. No Youtube API key provided in config")
		youtubeAPIKeyPresent = false
		return
	}

	songSignal = make(chan *VoiceInstance)

	go func() {
		for vi := range songSignal {
			go vi.PlayQueue()
		}
	}()

	malm.Info("Music initialized")
	youtubeAPIKeyPresent = true
}

func joinVoice(vi *VoiceInstance, s *discordgo.Session, m *discordgo.MessageCreate) *VoiceInstance {

	voiceChannelID := utils.FindVoiceChannel(s, m.Author.ID)
	if len(voiceChannelID) == 0 {
		s.ChannelMessageSend(m.ChannelID, "You are not in a voice channel") // Temporary
		return nil
	}

	if vi == nil {
		// Instance alreay initialized
		musicMutex.Lock()
		vi = &VoiceInstance{}
		guildID := utils.GetGuild(s, m)
		instances[guildID] = vi
		vi.GuildID = guildID
		vi.Session = s
		musicMutex.Unlock()
	}

	var err error
	vi.Voice, err = s.ChannelVoiceJoin(vi.GuildID, voiceChannelID, false, true)

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Failed to join voice channel")
		vi.Stop()
		return nil
	}

	err = vi.Voice.Speaking(false)
	if err != nil {
		malm.Error("%s", err)
		return nil
	}

	return vi
}

func LeaveVoice(vi *VoiceInstance, s *discordgo.Session, m *discordgo.MessageCreate) {

	if vi == nil {
		// Not in a voice channel
		return
	}

	vi.Disconnect()

	musicMutex.Lock()
	delete(instances, vi.GuildID)
	musicMutex.Unlock()
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
	if vi.Voice.ChannelID != voiceChannelID {
		s.ChannelMessageSend(m.ChannelID, "You are not in the same voice channel as the bot")
		return
	}

	if input.NumberOfArgsAre(0) {
		// User want to resume a song
		return
	}

	// If input is a youtube link

	song, err := parseMusicInput(m, strings.Join(input.GetArgs(), " "))
	if err != nil {
		malm.Error("%s", err)
		return
	}

	// This function is very slow. Takes up to 2 seconds

	if err != nil {
		malm.Error("%s", err)
		s.ChannelMessageSend(m.ChannelID, "Something went wrong when getting the song")
		return
	}

	// Add the song to the queue
	vi.AddToQueue(song)

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s added the song ``%s`` to the queue", m.Author.Username, song.Title))

	songSignal <- vi
}

func parseMusicInput(m *discordgo.MessageCreate, input string) (Song, error) {

	// youtubeBaseURL
	// youtubePattern

	var videoId string

	pattern := regexp.MustCompile(youtubePattern)
	if pattern.MatchString(input) {
		// Youtube link

		parsedURL, err := url.Parse(input)
		if err != nil {
			return Song{}, err
		}

		query := parsedURL.Query()
		videoId = query.Get("v")

		title, thumbnail, channelName, err := youtubeFindByVideoID(videoId)

		if err != nil {
			return Song{}, err
		}

		fmt.Printf("%s %s %s\n", title, thumbnail, channelName)

	} else {
		// Presumably a song name

	}
	song, err := youtubeDL(m, videoId)
	if err != nil {
		return Song{}, err
	}

	return song, nil
}

func StopMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled(s, m) {
		return
	}

}

func SkipMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled(s, m) {
		return
	}

}
