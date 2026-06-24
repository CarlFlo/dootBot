package music

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/CarlFlo/dootBot/src/bot/context"
	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/database"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

var (
	errDetectedNonYTURL        = errors.New("non youtube URL detected")
	errEmptyYTResult           = errors.New("empty youtube search result")
	errSongLengthLimitExceeded = "song duration exceeds the limit (%s min) in the config file"
	errStatusYTSearchQuery     = "youtube search error - status code: %d - query: %s"
	errStatusYTSearchVideoID   = "youtube search error - status code: %d - videoID: %s"
	httpClient                 = &http.Client{Timeout: 15 * time.Second}
)

const (
	youtubeFindEndpoint     = "https://www.googleapis.com/youtube/v3/videos?part=snippet,contentDetails&key=%s&id=%s"
	youtubeSearchEndpoint   = "https://www.googleapis.com/youtube/v3/search?part=snippet&type=video&key=%s&q=%s&fields=items(id)"
	youtubePlaylistEndpoint = "https://www.googleapis.com/youtube/v3/playlistItems?part=snippet,contentDetails&key=%s&playlistId=%s&maxResults=50&fields=items(snippet)"
)

type youtubeResponseFind struct {
	Items []itemsFind
}

type itemsFind struct {
	Snippet        snippet
	ID             string
	ContentDetails contentDetails
}

type youtubeResponseSearch struct {
	Items []itemsSearch
}

type itemsSearch struct {
	ID id
}

type id struct {
	VideoId string
}

type snippet struct {
	Title        string
	Thumbnails   thumbnails
	ChannelTitle string
}

type thumbnails struct {
	Standard standard
}

type standard struct {
	Url    string
	Width  int
	Height int
}

type contentDetails struct {
	Duration string
}

func isMusicEnabled() bool {
	return youtubeAPIKeysValid
}

func parseMusicInput(m *discordgo.MessageCreate, input string, song *Song) error {
	var title, thumbnail, channelName, videoID, streamURL string
	var duration time.Duration
	var err error

	input = strings.TrimSpace(input)

	if parsedURL, ok := parseURLInput(input); ok {
		if !isYoutubeURL(parsedURL) {
			return errDetectedNonYTURL
		}

		videoID, err = extractVideoID(parsedURL)
		if err != nil {
			return err
		}

		var cache database.YoutubeCache
		exists := cache.Check(videoID, &title, &thumbnail, &channelName, &streamURL, &duration)
		if !exists {
			title, thumbnail, channelName, duration, err = youtubeFindByVideoID(videoID)
			if err != nil {
				return err
			}
			cache.Cache(videoID, title, thumbnail, channelName, duration)
		}
	} else {
		title, thumbnail, channelName, videoID, duration, err = youtubeFindBySearch(input)
		if err != nil {
			return err
		}
	}

	song.ChannelID = m.ChannelID
	song.User = m.Author.ID
	song.Title = title
	song.Thumbnail = thumbnail
	song.ChannelName = channelName
	song.YoutubeVideoID = videoID
	song.Duration = duration
	song.StreamURL = streamURL

	return checkDurationCompliance(song.Duration)
}

func youtubeFindByVideoID(videoID string) (string, string, string, time.Duration, error) {
	var emptyDuration time.Duration

	apiKey := utils.GetYoutubeAPIKey()
	res, err := httpClient.Get(fmt.Sprintf(youtubeFindEndpoint, apiKey, videoID))
	if err != nil {
		return "", "", "", emptyDuration, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", "", "", emptyDuration, fmt.Errorf(errStatusYTSearchVideoID, res.StatusCode, videoID)
	}

	var page youtubeResponseFind
	if err := json.NewDecoder(res.Body).Decode(&page); err != nil {
		return "", "", "", emptyDuration, err
	}

	if len(page.Items) == 0 {
		return "", "", "", emptyDuration, errEmptyYTResult
	}

	title := page.Items[0].Snippet.Title
	thumbnail := page.Items[0].Snippet.Thumbnails.Standard.Url
	channelName := page.Items[0].Snippet.ChannelTitle
	duration := youtubeTimeToDuration(page.Items[0].ContentDetails.Duration)

	return title, thumbnail, channelName, duration, nil
}

func youtubeFindBySearch(query string) (string, string, string, string, time.Duration, error) {
	var emptyDuration time.Duration

	query = url.QueryEscape(query)
	apiKey := utils.GetYoutubeAPIKey()
	res, err := httpClient.Get(fmt.Sprintf(youtubeSearchEndpoint, apiKey, query))
	if err != nil {
		return "", "", "", "", emptyDuration, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", "", "", "", emptyDuration, fmt.Errorf(errStatusYTSearchQuery, res.StatusCode, query)
	}

	var page youtubeResponseSearch
	if err := json.NewDecoder(res.Body).Decode(&page); err != nil {
		return "", "", "", "", emptyDuration, err
	}

	if len(page.Items) == 0 {
		return "", "", "", "", emptyDuration, errEmptyYTResult
	}

	videoID := page.Items[0].ID.VideoId
	title, thumbnail, channelName, duration, err := youtubeFindByVideoID(videoID)
	if err != nil {
		return "", "", "", "", emptyDuration, err
	}

	return title, thumbnail, channelName, videoID, duration, nil
}

func joinVoice(vi *VoiceInstance, authorID, channelID string) (*VoiceInstance, error) {
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

	vi.mu.RLock()
	alreadyConnected := vi.voice != nil && vi.voice.ChannelID == voiceChannelID
	vi.mu.RUnlock()
	if alreadyConnected {
		return vi, nil
	}

	voice, err := context.SESSION.ChannelVoiceJoin(guildID, voiceChannelID, false, true)
	if err != nil {
		return nil, errors.New("failed to join voice channel")
	}

	if err := voice.Speaking(false); err != nil {
		return nil, err
	}

	vi.mu.Lock()
	vi.voice = voice
	vi.mu.Unlock()

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

	vi.mu.RLock()
	currentChannelID := ""
	if vi.voice != nil {
		currentChannelID = vi.voice.ChannelID
	}
	vi.mu.RUnlock()

	if currentChannelID == "" || currentChannelID != voiceChannelID {
		return errors.New("you are not in the same voice channel as the bot")
	}

	return nil
}

func parseURLInput(input string) (*url.URL, bool) {
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return nil, false
	}

	parsedURL, err := url.Parse(input)
	if err != nil {
		return nil, false
	}

	return parsedURL, true
}

func isYoutubeURL(parsedURL *url.URL) bool {
	host := strings.ToLower(strings.TrimPrefix(parsedURL.Host, "www."))
	return host == "youtube.com" || host == "m.youtube.com" || host == "youtu.be"
}

func extractVideoID(parsedURL *url.URL) (string, error) {
	host := strings.ToLower(strings.TrimPrefix(parsedURL.Host, "www."))

	switch host {
	case "youtu.be":
		videoID := strings.Trim(parsedURL.Path, "/")
		if videoID == "" {
			return "", errors.New("youtube url is missing a video id")
		}
		return videoID, nil
	case "youtube.com", "m.youtube.com":
		videoID := parsedURL.Query().Get("v")
		if videoID == "" {
			return "", errors.New("youtube url is missing a video id")
		}
		return videoID, nil
	default:
		return "", errDetectedNonYTURL
	}
}

func youtubeTimeToDuration(input string) time.Duration {
	var duration time.Duration

	input = strings.TrimPrefix(input, "PT")

	split := strings.Split(input, "H")
	if len(split) != 1 {
		if val, err := strconv.Atoi(split[0]); err == nil {
			duration += time.Hour * time.Duration(val)
		} else {
			malm.Warn("Unable to format time for input %v. Reason: %s", split, err)
		}
		input = split[1]
	}

	split = strings.Split(input, "M")
	if len(split) != 1 {
		if val, err := strconv.Atoi(split[0]); err == nil {
			duration += time.Minute * time.Duration(val)
		} else {
			malm.Warn("Unable to format time for input %v. Reason: %s", split, err)
		}
		input = split[1]
	}

	split = strings.Split(input, "S")
	if len(split) != 1 {
		if val, err := strconv.Atoi(split[0]); err == nil {
			duration += time.Second * time.Duration(val)
		} else {
			malm.Warn("Unable to format time for input %v. Reason: %s", split, err)
		}
	}

	return duration
}

func checkDurationCompliance(duration time.Duration) error {
	maxSongDuration := time.Minute * config.CONFIG.Music.MaxSongLengthMinutes
	if duration > maxSongDuration {
		return fmt.Errorf(errSongLengthLimitExceeded, config.CONFIG.Music.MaxSongLengthMinutes)
	}

	return nil
}
