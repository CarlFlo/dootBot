package music

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/CarlFlo/dootBot/src/bot/context"
	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/dootBot/src/permissions"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

var instances = map[string]*VoiceInstance{}

var (
	musicMutex sync.Mutex
	manager    = NewMusicManager()
)

func Initialize() {
	if !isMusicEnabled() {
		malm.Info("Music disabled")
		return
	}

	malm.Info("Music configured for Lavalink")
}

func AttachSession(session *discordgo.Session) {
	if !isMusicEnabled() || session == nil {
		return
	}

	manager.AttachSession(session)
}

func ForwardVoiceStateUpdate(vs *discordgo.VoiceStateUpdate) {
	manager.ForwardVoiceStateUpdate(vs)
}

func ForwardVoiceServerUpdate(vs *discordgo.VoiceServerUpdate) {
	manager.ForwardVoiceServerUpdate(vs)
}

func Close() {
	musicMutex.Lock()
	defer musicMutex.Unlock()

	for guildID, vi := range instances {
		vi.Disconnect()
		vi.Close()
		delete(instances, guildID)
	}

	manager.Close()
}

func PlayMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled() {
		utils.SendMessageNeutral(m, "Music is currently disabled")
		return
	}

	vi, err := getOrCreateVoiceInstance(m.Author.ID, m.ChannelID)
	if err != nil {
		utils.SendMessageFailure(m, err.Error())
		return
	}

	if err := validateSameVoiceChannel(vi, m.Author.ID); err != nil {
		utils.SendMessageFailure(m, err.Error())
		return
	}

	if input.NumberOfArgsAre(0) {
		if !input.HasGuildPermission(permissions.LevelController) {
			utils.SendMessageFailure(m, "You need Controller permission to resume playback")
			return
		}
		if !vi.PauseToggle() {
			utils.SendMessageFailure(m, "There is no active song to resume")
			return
		}
		if err := vi.refreshOverviewMessage(); err != nil {
			malm.Error("unable to refresh music overview: %s", err)
		}
		return
	}

	vi.markLoading(true)
	song, err := parseMusicInput(m, strings.Join(input.GetArgs(), " "))
	vi.markLoading(false)
	if err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Something went wrong when getting the song.\nReason: %s", err))
		return
	}

	vi.AddToQueue(song)
	if err := updateOverviewMessageForQueue(m.ChannelID, vi); err != nil {
		malm.Error("unable to update music overview: %s", err)
	}

	addedSongMsg, _ := utils.SendMessageNeutral(m, fmt.Sprintf("%s added the song ``%s`` to the queue (%s)", m.Author.Username, song.Title, song.Duration))
	go func() {
		time.Sleep(5 * time.Second)
		if addedSongMsg != nil {
			_ = context.SESSION.ChannelMessageDelete(m.ChannelID, addedSongMsg.ID)
		}
	}()

	if !vi.IsWorkerRunning() {
		if err := manager.playCurrentSongWithSession(s, vi); err != nil {
			utils.SendMessageFailure(m, fmt.Sprintf("Unable to start playback.\nReason: %s", err))
			return
		}
		if err := vi.refreshOverviewMessage(); err != nil {
			malm.Error("unable to refresh music overview: %s", err)
		}
	}
}

func StopMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled() {
		utils.SendMessageNeutral(m, "Music is currently disabled")
		return
	}

	vi, err := getExistingVoiceInstanceByChannel(m.ChannelID)
	if err != nil || vi == nil {
		return
	}

	leaveVoice(vi)
}

func SkipMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled() {
		utils.SendMessageNeutral(m, "Music is currently disabled")
		return
	}

	vi, err := getExistingVoiceInstanceByChannel(m.ChannelID)
	if err != nil || vi == nil {
		return
	}

	if vi.Skip() {
		utils.SendMessageSuccess(m, "Skipped the song")
		return
	}

	utils.SendMessageFailure(m, "There is no song to skip")
}

func ClearQueueMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled() {
		utils.SendMessageNeutral(m, "Music is currently disabled")
		return
	}

	vi, err := getExistingVoiceInstanceByChannel(m.ChannelID)
	if err != nil || vi == nil {
		return
	}

	vi.ClearQueueAfter()
	if err := vi.refreshOverviewMessage(); err != nil {
		malm.Error("unable to refresh music overview: %s", err)
	}
}

func PauseMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled() {
		utils.SendMessageNeutral(m, "Music is currently disabled")
		return
	}

	vi, err := getExistingVoiceInstanceByChannel(m.ChannelID)
	if err != nil || vi == nil {
		return
	}

	if !vi.PauseToggle() {
		utils.SendMessageFailure(m, "There is no song to pause")
		return
	}

	if err := vi.refreshOverviewMessage(); err != nil {
		malm.Error("unable to refresh music overview: %s", err)
	}
}

func LoopMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled() {
		utils.SendMessageNeutral(m, "Music is currently disabled")
		return
	}

	vi, err := getExistingVoiceInstanceByChannel(m.ChannelID)
	if err != nil || vi == nil {
		return
	}

	vi.ToggleLooping()
	if err := vi.refreshOverviewMessage(); err != nil {
		malm.Error("unable to refresh music overview: %s", err)
	}
}

func MusicPrevious(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !isMusicEnabled() {
		utils.SendMessageNeutral(m, "Music is currently disabled")
		return
	}

	vi, err := getExistingVoiceInstanceByChannel(m.ChannelID)
	if err != nil || vi == nil {
		return
	}

	if !vi.Prev() {
		utils.SendMessageNeutral(m, "There is no song to restart")
		return
	}

	utils.SendMessageNeutral(m, "Restarted the current song")
}

func getExistingVoiceInstanceByChannel(channelID string) (*VoiceInstance, error) {
	guildID, err := utils.GetGuild(channelID)
	if err != nil {
		malm.Error("Error getting guild ID: %s", err)
		return nil, err
	}

	musicMutex.Lock()
	defer musicMutex.Unlock()
	return instances[guildID], nil
}

func getOrCreateVoiceInstance(authorID, channelID string) (*VoiceInstance, error) {
	guildID, err := utils.GetGuild(channelID)
	if err != nil {
		malm.Error("Error getting guild ID: %s", err)
		return nil, fmt.Errorf("internal error")
	}

	musicMutex.Lock()
	vi := instances[guildID]
	musicMutex.Unlock()

	return joinVoice(vi, authorID, channelID)
}

func updateOverviewMessageForQueue(channelID string, vi *VoiceInstance) error {
	complexMessage := &discordgo.MessageSend{}
	if !vi.IsWorkerRunning() {
		vi.setLoading(true)
	}

	CreateMusicOverviewMessage(channelID, complexMessage)

	msg, err := context.SESSION.ChannelMessageSendComplex(channelID, complexMessage)
	if err != nil {
		return err
	}

	if oldMessageID := vi.GetMessageID(); oldMessageID != "" {
		_ = context.SESSION.ChannelMessageDelete(vi.GetMessageChannelID(), oldMessageID)
	}

	vi.SetMessageID(msg.ID)
	vi.SetMessageChannelID(msg.ChannelID)
	return nil
}
