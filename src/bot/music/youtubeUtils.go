package music

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

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
)

const (
	youtubeFindEndpoint     string = "https://www.googleapis.com/youtube/v3/videos?part=snippet,contentDetails&key=%s&id=%s"
	youtubeSearchEndpoint   string = "https://www.googleapis.com/youtube/v3/search?part=snippet&type=video&key=%s&q=%s&fields=items(id)"
	youtubePlaylistEndpoint string = "https://www.googleapis.com/youtube/v3/playlistItems?part=snippet,contentDetails&key=%s&playlistId=%s&maxResults=50&fields=items(snippet)"
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

	var title, thumbnail, channelName, videoID, duration, streamURL string
	var err error

	ytRegex := regexp.MustCompile(youtubePattern)
	urlRegex := regexp.MustCompile(urlPattern)

	if ytRegex.MatchString(input) {
		// Youtube link

		parsedURL, err := url.Parse(input)
		if err != nil {
			return err
		}

		query := parsedURL.Query()
		videoID = query.Get("v")

		// Cache
		var cache database.YoutubeCache
		// "Will attempt to load the values into the pointers"
		exists := cache.Check(videoID, &title, &thumbnail, &channelName, &duration, &streamURL)

		if !exists {
			title, thumbnail, channelName, duration, err = youtubeFindByVideoID(videoID)
			if err != nil {
				return err
			}
			// Save results to cache
			cache.Cache(videoID, title, thumbnail, channelName, duration)
		}

	} else if urlRegex.MatchString(input) {
		// URL from another source than Youtube
		parsedURL, err := url.Parse(input)
		if err != nil {
			return err
		}

		malm.Debug("%s", parsedURL.Host)
		return errDetectedNonYTURL
	} else {
		// Presumably a song name. Search Youtube
		title, thumbnail, channelName, videoID, duration, err = youtubeFindBySearch(input)
		if err != nil {
			return err
		}
	}

	// Update the song object
	song.ChannelID = m.ChannelID
	song.User = m.Author.ID
	song.Title = title
	song.Thumbnail = thumbnail
	song.ChannelName = channelName
	song.YoutubeVideoID = videoID
	song.duration = duration
	song.StreamURL = streamURL

	// Returns 'nil' if everything is ok
	return checkDurationCompliance(song.duration)
}

// Returns the title, thumbnail and channel of a youtube video
// error if there was any problem
func youtubeFindByVideoID(videoID string) (string, string, string, string, error) {

	apiKey := utils.GetYoutubeAPIKey()
	res, err := http.Get(fmt.Sprintf(youtubeFindEndpoint, apiKey, videoID))
	if err != nil {
		return "", "", "", "", err
	} else if res.StatusCode != 200 {
		return "", "", "", "", fmt.Errorf(errStatusYTSearchVideoID, res.StatusCode, videoID)
	}
	defer res.Body.Close()

	var page youtubeResponseFind

	err = json.NewDecoder(res.Body).Decode(&page)
	if err != nil {
		return "", "", "", "", err
	}

	if len(page.Items) == 0 {
		return "", "", "", "", errEmptyYTResult
	}

	title := page.Items[0].Snippet.Title
	thumbnail := page.Items[0].Snippet.Thumbnails.Standard.Url
	channelName := page.Items[0].Snippet.ChannelTitle
	duration := formatYoutubeDuration(page.Items[0].ContentDetails.Duration)

	return title, thumbnail, channelName, duration, nil
}

func youtubeFindBySearch(query string) (string, string, string, string, string, error) {

	query = url.QueryEscape(query)
	apiKey := utils.GetYoutubeAPIKey()
	res, err := http.Get(fmt.Sprintf(youtubeSearchEndpoint, apiKey, query))
	if err != nil {
		return "", "", "", "", "", err
	} else if res.StatusCode != 200 {
		return "", "", "", "", "", fmt.Errorf(errStatusYTSearchQuery, res.StatusCode, query)
	}
	defer res.Body.Close()

	var page youtubeResponseSearch

	err = json.NewDecoder(res.Body).Decode(&page)
	if err != nil {
		return "", "", "", "", "", err
	}

	if len(page.Items) == 0 {
		return "", "", "", "", "", errEmptyYTResult
	}

	videoID := page.Items[0].ID.VideoId

	title, thumbnail, channelName, duration, err := youtubeFindByVideoID(videoID)
	if err != nil {
		return "", "", "", "", "", err
	}

	return title, thumbnail, channelName, videoID, duration, nil
}

func joinVoice(vi *VoiceInstance, authorID, channelID string) (*VoiceInstance, string) {

	voiceChannelID := utils.FindVoiceChannel(authorID)
	if len(voiceChannelID) == 0 {
		return nil, "You are not in a voice channel"
	}

	if vi == nil {
		// Instance not initialized
		musicMutex.Lock()
		defer musicMutex.Unlock()

		vi = &VoiceInstance{}

		guildID, err := utils.GetGuild(channelID)
		if err != nil {
			vi = nil
			return nil, "Internal error"
		}

		if err := vi.New(guildID); err != nil {
			return nil, ""
		}

		instances[vi.guildID] = vi
	}

	var err error
	vi.voice, err = context.SESSION.ChannelVoiceJoin(vi.GetGuildID(), voiceChannelID, false, true)

	if err != nil {
		vi.Stop()
		return nil, "Failed to join voice channel"
	}

	err = vi.voice.Speaking(false)
	if err != nil {
		malm.Error("%s", err)
		return nil, ""
	}

	return vi, ""
}

func leaveVoice(vi *VoiceInstance) {

	vi.Disconnect()
	vi.Close()

	musicMutex.Lock()
	delete(instances, vi.GetGuildID())
	musicMutex.Unlock()
}

// formatYoutubeDuration formats the youtube duration string
// PT1H24M47S -> 1h 24m 47s
func formatYoutubeDuration(input string) string {
	// example string: PT1H24M47S

	var buffer bytes.Buffer

	// Removes the prefix
	input = strings.TrimPrefix(input, "PT")

	// Split the string into hours, minutes and seconds
	split := strings.Split(input, "H")
	if len(split) != 1 {
		buffer.WriteString(split[0] + "h ")
		input = split[1]
	}

	split = strings.Split(input, "M")
	if len(split) != 1 {
		buffer.WriteString(split[0] + "m ")
		input = split[1]
	}

	split = strings.Split(input, "S")
	if len(split) != 1 {
		buffer.WriteString(split[0] + "s")
	}

	return buffer.String()
}

func checkDurationCompliance(duration string) error {
	// MaxSongLengthMinutes

	parts := strings.Split(duration, " ")

	totalMinutes := 0

	for _, part := range parts {
		if strings.HasSuffix(part, "h") {
			hoursStr := strings.TrimSuffix(part, "h")
			hours, err := strconv.Atoi(hoursStr)
			if err != nil {
				return err
			}
			totalMinutes += hours * 60
		} else if strings.HasSuffix(part, "m") {
			minutesStr := strings.TrimSuffix(part, "m")
			minutes, err := strconv.Atoi(minutesStr)
			if err != nil {
				return err
			}
			totalMinutes += minutes
		}
	}

	if totalMinutes > int(config.CONFIG.Music.MaxSongLengthMinutes) {
		return fmt.Errorf(errSongLengthLimitExceeded, config.CONFIG.Music.MaxSongLengthMinutes)
	}

	return nil
}
