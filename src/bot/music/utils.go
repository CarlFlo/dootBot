package music

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/CarlFlo/DiscordMoneyBot/src/config"
	"github.com/CarlFlo/DiscordMoneyBot/src/utils"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

var (
	errDetectedNonYTURL      = errors.New("non youtube URL detected")
	errEmptyYTResult         = errors.New("empty youtube search result")
	errStatusYTSearchQuery   = "youtube search error - status code: %d - query: %s"
	errStatusYTSearchVideoID = "youtube search error - status code: %d - videoID: %s"
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

// Returns the title, thumbnail and channel of a youtube video
// error if there was any problem
func youtubeFindByVideoID(videoID string) (string, string, string, string, error) {

	res, err := http.Get(fmt.Sprintf(youtubeFindEndpoint, config.CONFIG.Music.YoutubeAPIKey, videoID))
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

	// TODO: if duration is too long (set in config) return error

	return title, thumbnail, channelName, duration, nil
}

func youtubeSearch(query string) (string, string, string, string, string, error) {

	query = url.QueryEscape(query)
	res, err := http.Get(fmt.Sprintf(youtubeSearchEndpoint, config.CONFIG.Music.YoutubeAPIKey, query))
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

func isMusicEnabled(s *discordgo.Session, m *discordgo.MessageCreate) bool {

	if !youtubeAPIKeyPresent {
		utils.SendMessageNeutral(s, m, "Music is currently disabled")
	}

	return youtubeAPIKeyPresent
}

func joinVoice(vi *VoiceInstance, s *discordgo.Session, m *discordgo.MessageCreate) *VoiceInstance {

	voiceChannelID := utils.FindVoiceChannel(s, m.Author.ID)
	if len(voiceChannelID) == 0 {
		utils.SendMessageFailure(s, m, "You are not in a voice channel")
		return nil
	}

	if vi == nil {
		// Instance not initialized
		musicMutex.Lock()
		vi = &VoiceInstance{}
		guildID := utils.GetGuild(s, m)
		instances[guildID] = vi
		vi.guildID = guildID
		vi.Session = s
		musicMutex.Unlock()
	}

	var err error
	vi.voice, err = s.ChannelVoiceJoin(vi.GetGuildID(), voiceChannelID, false, true)

	if err != nil {
		utils.SendMessageFailure(s, m, "Failed to join voice channel")
		vi.Stop()
		return nil
	}

	err = vi.voice.Speaking(false)
	if err != nil {
		malm.Error("%s", err)
		return nil
	}

	return vi
}

func leaveVoice(vi *VoiceInstance, s *discordgo.Session, m *discordgo.MessageCreate) {

	vi.Disconnect()

	musicMutex.Lock()
	delete(instances, vi.GetGuildID())
	musicMutex.Unlock()
}

func parseMusicInput(m *discordgo.MessageCreate, input string, song *Song) error {

	var title, thumbnail, channelName, videoID, duration string
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

		title, thumbnail, channelName, duration, err = youtubeFindByVideoID(videoID)
		if err != nil {
			return err
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
		title, thumbnail, channelName, videoID, duration, err = youtubeSearch(input)
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

	return nil
}

// formatYoutubeDuration formats the youtube duration string
// PT1H24M47S -> 1 hour 24 minutes and 47 seconds
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
