package music

import (
	stdcontext "context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v4/disgolink"
	"github.com/disgoorg/disgolink/v4/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type MusicManager struct {
	mu                 sync.RWMutex
	client             *disgolink.Client
	node               *disgolink.Node
	session            *discordgo.Session
	lastConnectAttempt time.Time
	lastConnectErr     error
}

func NewMusicManager() *MusicManager {
	return &MusicManager{}
}

func (m *MusicManager) AttachSession(session *discordgo.Session) {
	if session == nil {
		return
	}

	m.mu.Lock()
	m.session = session
	m.mu.Unlock()

	ctx, cancel := stdcontext.WithTimeout(stdcontext.Background(), 10*time.Second)
	defer cancel()

	if err := m.EnsureReady(ctx); err != nil {
		malm.Warn("Lavalink unavailable during startup: %s", err)
	}
}

func (m *MusicManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client != nil {
		m.client.Close()
	}
	m.client = nil
	m.node = nil
	m.lastConnectErr = nil
}

func (m *MusicManager) EnsureReady(ctx stdcontext.Context) error {
	if !isMusicEnabled() {
		return errors.New("music is disabled")
	}

	m.mu.RLock()
	session := m.session
	node := m.node
	client := m.client
	lastAttempt := m.lastConnectAttempt
	lastErr := m.lastConnectErr
	m.mu.RUnlock()

	if node != nil && node.Status() == disgolink.StatusConnected {
		return nil
	}

	if session == nil || session.State == nil || session.State.User == nil {
		return errors.New("discord session is not ready yet")
	}

	if time.Since(lastAttempt) < 5*time.Second && lastErr != nil {
		return lastErr
	}

	cfg, err := loadLavalinkConfig()
	if err != nil {
		m.mu.Lock()
		m.lastConnectAttempt = time.Now()
		m.lastConnectErr = err
		m.mu.Unlock()
		return err
	}

	botID, err := snowflake.Parse(session.State.User.ID)
	if err != nil {
		return fmt.Errorf("invalid bot id: %w", err)
	}

	m.mu.Lock()
	m.lastConnectAttempt = time.Now()
	if client != nil && m.node != nil {
		client.RemoveNode(m.node.Config.Name)
	}
	lavalinkClient := disgolink.New(botID, disgolink.WithListeners(m))
	m.client = lavalinkClient
	m.node = nil
	m.mu.Unlock()

	nodeCtx, cancel := stdcontext.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	node, err = lavalinkClient.AddNode(nodeCtx, disgolink.NodeConfig{
		Name:     "main",
		Address:  net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Password: cfg.Password,
		Secure:   cfg.Secure,
	})
	if err != nil {
		m.mu.Lock()
		m.lastConnectErr = fmt.Errorf("unable to connect to Lavalink at %s:%d: %w", cfg.Host, cfg.Port, err)
		m.mu.Unlock()
		return m.lastConnectErr
	}

	m.mu.Lock()
	m.node = node
	m.lastConnectErr = nil
	m.mu.Unlock()

	malm.Info("Connected to Lavalink at %s:%d", cfg.Host, cfg.Port)
	return nil
}

func (m *MusicManager) loadTrack(ctx stdcontext.Context, identifier string) (lavalink.Track, error) {
	if err := m.EnsureReady(ctx); err != nil {
		return lavalink.Track{}, err
	}

	node := m.bestNode()
	if node == nil {
		return lavalink.Track{}, errMusicBackendUnavailable
	}

	result, err := node.Rest.LoadTracks(ctx, identifier)
	if err != nil {
		return lavalink.Track{}, fmt.Errorf("failed to load tracks from lavalink: %w", err)
	}
	if result == nil {
		return lavalink.Track{}, errEmptyTrackResult
	}

	switch data := result.Data.(type) {
	case lavalink.Track:
		return data, nil
	case lavalink.Playlist:
		if len(data.Tracks) == 0 {
			return lavalink.Track{}, errEmptyTrackResult
		}
		return data.Tracks[0], nil
	case lavalink.Search:
		if len(data) == 0 {
			return lavalink.Track{}, errEmptyTrackResult
		}
		return data[0], nil
	case lavalink.Empty:
		return lavalink.Track{}, errEmptyTrackResult
	case lavalink.Exception:
		return lavalink.Track{}, fmt.Errorf("lavalink load failed: %s", data.Message)
	default:
		return lavalink.Track{}, errEmptyTrackResult
	}
}

func (m *MusicManager) playCurrentSongWithSession(session *discordgo.Session, vi *VoiceInstance) error {
	if session != nil {
		m.AttachSession(session)
	}
	return m.playCurrentSong(stdcontext.Background(), vi)
}

func (m *MusicManager) playCurrentSong(ctx stdcontext.Context, vi *VoiceInstance) error {
	if vi == nil {
		return errors.New("music instance is nil")
	}
	if err := m.EnsureReady(ctx); err != nil {
		return err
	}
	if err := vi.ensureQueuePlayable(); err != nil {
		return err
	}

	player, err := m.playerForGuild(vi.GetGuildID())
	if err != nil {
		return err
	}

	song, err := vi.currentSong()
	if err != nil {
		return err
	}

	vi.mu.Lock()
	vi.workerRunning = true
	vi.loading = false
	vi.paused = false
	vi.mu.Unlock()

	if err := player.Update(ctx, disgolink.WithTrack(song.Track), disgolink.WithPaused(false)); err != nil {
		return fmt.Errorf("unable to start track: %w", err)
	}

	return nil
}

func (m *MusicManager) stopPlayback(ctx stdcontext.Context, guildID string) error {
	player, err := m.playerForGuild(guildID)
	if err != nil {
		return err
	}

	return player.Update(ctx, disgolink.WithNullTrack(), disgolink.WithPaused(false))
}

func (m *MusicManager) setPaused(guildID string, paused bool) error {
	player, err := m.playerForGuild(guildID)
	if err != nil {
		return err
	}

	return player.Update(stdcontext.Background(), disgolink.WithPaused(paused))
}

func (m *MusicManager) disconnectVoice(ctx stdcontext.Context, guildID string) {
	player, err := m.playerForGuild(guildID)
	if err == nil {
		if err := player.Destroy(ctx); err != nil {
			malm.Warn("unable to destroy lavalink player for guild %s: %s", guildID, err)
		}
	} else {
		malm.Warn("unable to destroy lavalink player for guild %s: %s", guildID, err)
	}

	m.mu.RLock()
	session := m.session
	m.mu.RUnlock()
	if session != nil {
		// This only asks Discord to clear the bot's voice state.
		// Audio transport and DAVE/E2EE negotiation stay on Lavalink.
		if err := session.ChannelVoiceJoinManual(guildID, "", false, true); err != nil {
			malm.Warn("unable to leave voice for guild %s: %s", guildID, err)
		}
	}
}

func (m *MusicManager) ForwardVoiceStateUpdate(vs *discordgo.VoiceStateUpdate) {
	if vs == nil || vs.VoiceState == nil {
		return
	}

	m.mu.RLock()
	session := m.session
	client := m.client
	m.mu.RUnlock()
	if session == nil || session.State == nil || session.State.User == nil || client == nil {
		return
	}

	if vs.UserID != session.State.User.ID {
		return
	}

	guildID, err := snowflake.Parse(vs.GuildID)
	if err != nil {
		return
	}

	var channelID *snowflake.ID
	if vs.ChannelID != "" {
		parsedChannelID, parseErr := snowflake.Parse(vs.ChannelID)
		if parseErr == nil {
			channelID = &parsedChannelID
		}
	}

	client.OnVoiceStateUpdate(stdcontext.Background(), guildID, channelID, vs.SessionID)
}

func (m *MusicManager) ForwardVoiceServerUpdate(vs *discordgo.VoiceServerUpdate) {
	if vs == nil {
		return
	}

	m.mu.RLock()
	client := m.client
	m.mu.RUnlock()
	if client == nil {
		return
	}

	guildID, err := snowflake.Parse(vs.GuildID)
	if err != nil {
		return
	}

	client.OnVoiceServerUpdate(stdcontext.Background(), guildID, vs.Token, vs.Endpoint)
}

func (m *MusicManager) OnEvent(event disgolink.Event) {
	switch e := event.(type) {
	case *disgolink.PlayerTrackStartEvent:
		vi := m.voiceInstanceByGuildID(e.Player.GuildID.String())
		if vi == nil {
			return
		}
		vi.handleTrackStarted()
		if err := vi.refreshOverviewMessage(); err != nil {
			malm.Error("unable to refresh music overview: %s", err)
		}
	case *disgolink.PlayerTrackEndEvent:
		vi := m.voiceInstanceByGuildID(e.Player.GuildID.String())
		if vi == nil {
			return
		}

		shouldContinue, err := vi.handleTrackEnded(e.Reason)
		if err != nil {
			malm.Error("unable to advance music queue: %s", err)
			return
		}
		if shouldContinue {
			if err := m.playCurrentSong(stdcontext.Background(), vi); err != nil {
				malm.Error("unable to continue music playback: %s", err)
			}
		}
		if err := vi.refreshOverviewMessage(); err != nil {
			malm.Error("unable to refresh music overview: %s", err)
		}
	case *disgolink.PlayerTrackExceptionEvent:
		malm.Error("lavalink track exception in guild %s: %s", e.Player.GuildID, e.Exception.Message)
	case *disgolink.PlayerWebSocketClosedEvent:
		malm.Warn("lavalink websocket closed in guild %s with code %d", e.Player.GuildID, e.Code)
	}
}

func (m *MusicManager) playerForGuild(guildID string) (*disgolink.Player, error) {
	if err := m.EnsureReady(stdcontext.Background()); err != nil {
		return nil, err
	}

	parsedGuildID, err := snowflake.Parse(guildID)
	if err != nil {
		return nil, fmt.Errorf("invalid guild id: %w", err)
	}

	m.mu.RLock()
	client := m.client
	m.mu.RUnlock()
	if client == nil {
		return nil, errMusicBackendUnavailable
	}

	player := client.Player(parsedGuildID)
	if player == nil {
		return nil, errMusicBackendUnavailable
	}
	if player.Node == nil {
		player.Node = m.bestNode()
	}
	if player.Node == nil {
		return nil, errMusicBackendUnavailable
	}

	return player, nil
}

func (m *MusicManager) bestNode() *disgolink.Node {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.client == nil {
		return nil
	}
	return m.client.BestNode()
}

func (m *MusicManager) voiceInstanceByGuildID(guildID string) *VoiceInstance {
	musicMutex.Lock()
	defer musicMutex.Unlock()
	return instances[guildID]
}

func (m *MusicManager) editOverviewMessage(msgEdit *discordgo.MessageEdit) error {
	if msgEdit == nil || m.session == nil {
		return nil
	}
	_, err := m.session.ChannelMessageEditComplex(msgEdit)
	return err
}

func (m *MusicManager) deleteOverviewMessage(channelID, messageID string) {
	m.mu.RLock()
	session := m.session
	m.mu.RUnlock()
	if session == nil {
		return
	}
	if err := session.ChannelMessageDelete(channelID, messageID); err != nil {
		malm.Debug("unable to delete music overview message: %s", err)
	}
}

type lavalinkConfig struct {
	Host     string
	Port     int
	Password string
	Secure   bool
}

func loadLavalinkConfig() (lavalinkConfig, error) {
	cfg := lavalinkConfig{
		Host:     firstNonEmpty(os.Getenv("LAVALINK_HOST"), config.CONFIG.Music.Lavalink.Host),
		Port:     firstPositive(envInt("LAVALINK_PORT"), config.CONFIG.Music.Lavalink.Port),
		Password: firstNonEmpty(os.Getenv("LAVALINK_PASSWORD"), config.CONFIG.Music.Lavalink.Password),
		Secure:   envBool("LAVALINK_SECURE", config.CONFIG.Music.Lavalink.Secure),
	}

	if cfg.Host == "" {
		return lavalinkConfig{}, errors.New("lavalink host is not configured")
	}

	if cfg.Port <= 0 {
		return lavalinkConfig{}, errors.New("lavalink port must be greater than zero")
	}

	if cfg.Password == "" {
		return lavalinkConfig{}, errors.New("lavalink password is not configured")
	}

	return cfg, nil
}

func envInt(key string) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return value
}

func envBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func firstPositive(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
