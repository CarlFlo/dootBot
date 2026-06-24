package music

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	botcontext "github.com/CarlFlo/dootBot/src/bot/context"
	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v4/lavalink"
)

var (
	errMusicBackendUnavailable = errors.New("music backend is currently unavailable")
	errEmptyTrackResult        = errors.New("no tracks matched that query")
	errSongLengthLimitExceeded = "song duration exceeds the limit (%s min) in the config file"
)

func isMusicEnabled() bool {
	return config.CONFIG != nil && config.CONFIG.Music.EnableMusic
}

func parseMusicInput(m *discordgo.MessageCreate, input string) (*Song, error) {
	ctx, cancel := contextWithTimeout(15 * time.Second)
	defer cancel()

	if err := manager.EnsureReady(ctx); err != nil {
		return nil, err
	}

	identifier := buildTrackIdentifier(input)
	track, err := manager.loadTrack(ctx, identifier)
	if err != nil {
		return nil, err
	}

	song := NewSongFromTrack(track, m.ChannelID, m.Author.ID)
	if err := checkDurationCompliance(song.Duration); err != nil {
		return nil, err
	}

	return song, nil
}

func joinVoice(vi *VoiceInstance, authorID, channelID string) (*VoiceInstance, error) {
	if err := manager.EnsureReady(context.Background()); err != nil {
		return nil, err
	}

	voiceChannelID := utils.FindVoiceChannel(authorID)
	if voiceChannelID == "" {
		return nil, errors.New("you are not in a voice channel")
	}

	guildID, err := utils.GetGuild(channelID)
	if err != nil {
		return nil, errors.New("internal error")
	}

	musicMutex.Lock()
	defer musicMutex.Unlock()

	if vi == nil {
		vi = &VoiceInstance{}
		if err := vi.New(guildID); err != nil {
			return nil, err
		}
		instances[guildID] = vi
	}

	if vi.VoiceChannelID() == voiceChannelID {
		return vi, nil
	}

	// This only sends Discord's voice state update over the gateway.
	// Lavalink/disgolink handles the actual DAVE-capable voice session.
	if err := botcontext.SESSION.ChannelVoiceJoinManual(guildID, voiceChannelID, false, true); err != nil {
		delete(instances, guildID)
		return nil, fmt.Errorf("failed to join voice channel: %w", err)
	}

	vi.setVoiceChannelID(voiceChannelID)
	return vi, nil
}

func leaveVoice(vi *VoiceInstance) {
	if vi == nil {
		return
	}

	vi.Disconnect()
	vi.Close()

	musicMutex.Lock()
	delete(instances, vi.GetGuildID())
	musicMutex.Unlock()
}

func validateSameVoiceChannel(vi *VoiceInstance, authorID string) error {
	voiceChannelID := utils.FindVoiceChannel(authorID)
	if voiceChannelID == "" {
		return errors.New("you are not in a voice channel")
	}

	if vi.VoiceChannelID() == "" || vi.VoiceChannelID() != voiceChannelID {
		return errors.New("you are not in the same voice channel as the bot")
	}

	return nil
}

func buildTrackIdentifier(input string) string {
	input = strings.TrimSpace(input)

	if looksLikeURL(input) {
		return input
	}

	return lavalink.SearchTypeYouTube.Apply(input)
}

func looksLikeURL(input string) bool {
	parsedURL, err := url.ParseRequestURI(strings.TrimSpace(input))
	if err != nil {
		return false
	}
	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

func checkDurationCompliance(duration time.Duration) error {
	maxSongDuration := time.Minute * config.CONFIG.Music.MaxSongLengthMinutes
	if duration > maxSongDuration {
		return fmt.Errorf(errSongLengthLimitExceeded, config.CONFIG.Music.MaxSongLengthMinutes)
	}

	return nil
}

func contextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
