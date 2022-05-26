package music

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
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
	musicMutex sync.Mutex
	songSignal chan *VoiceInstance
)

func InitializeMusic() {

	songSignal = make(chan *VoiceInstance)

	go func() {
		for vi := range songSignal {
			go vi.PlayQueue()
		}
	}()
	malm.Info("Music initialized")
}

func JoinVoice(vi *VoiceInstance, s *discordgo.Session, m *discordgo.MessageCreate) *VoiceInstance {

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

// TODO: Problem: the input is by default set to lowercase breaking the youtube URL's

// Same as resume
func PlayMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	guildID := utils.GetGuild(s, m)
	vi := instances[guildID]
	if vi == nil {
		// Not initialized
		vi = JoinVoice(vi, s, m)
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

	parsedURL, err := url.Parse(input.GetArgs()[0])
	if err != nil {
		malm.Error("%s", err)
		s.ChannelMessageSend(m.ChannelID, "Something went wrong when parsing the Youtube url")
		return
	}

	query := parsedURL.Query()
	videoId := query.Get("v")

	song, err := youtubeDL(m, videoId)

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

func StopMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

}

func SkipMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

}
