package music

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

const (
	youtubeFindEndpoint   string = "https://www.googleapis.com/youtube/v3/videos?part=snippet&key=%s&id=%s"
	youtubeSearchEndpoint string = "https://www.googleapis.com/youtube/v3/search?part=snippet&type=video&key=%s&q=%s"
)

type youtubeResponseFind struct {
	Items []itemsFind
}

type itemsFind struct {
	Snippet snippet
	ID      string
}

type youtubeResponseSearch struct {
	Items []itemsSearch
}

type itemsSearch struct {
	Snippet snippet
	ID      id
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
	Width  string
	Height string
}

// Returns the title, thumbnail and channel of a youtube video
// error if there was any problem
func youtubeFindByVideoID(videoID string) (string, string, string, error) {

	res, err := http.Get(fmt.Sprintf(youtubeFindEndpoint, config.CONFIG.Music.YoutubeAPIKey, videoID))
	if err != nil {
		return "", "", "", err
	} else if res.StatusCode != 200 {
		return "", "", "", fmt.Errorf("youtube search error - status code: %d - videoID: %s", res.StatusCode, videoID)
	}
	defer res.Body.Close()

	var page youtubeResponseFind

	err = json.NewDecoder(res.Body).Decode(&page)
	if err != nil {
		return "", "", "", err
	}

	if len(page.Items) == 0 {
		return "", "", "", errors.New("empty youtube search result")
	}

	title := page.Items[0].Snippet.Title
	thumbnail := page.Items[0].Snippet.Thumbnails.Standard.Url
	channelName := page.Items[0].Snippet.ChannelTitle

	return title, thumbnail, channelName, nil
}

func youtubeSearch(query string) (string, string, string, string, error) {

	query = url.QueryEscape(query)
	res, err := http.Get(fmt.Sprintf(youtubeSearchEndpoint, config.CONFIG.Music.YoutubeAPIKey, query))
	if err != nil {
		return "", "", "", "", err
	} else if res.StatusCode != 200 {
		return "", "", "", "", fmt.Errorf("youtube search error - status code: %d - query: %s", res.StatusCode, query)
	}
	defer res.Body.Close()

	var page youtubeResponseSearch

	err = json.NewDecoder(res.Body).Decode(&page)
	if err != nil {
		return "", "", "", "", err
	}

	if len(page.Items) == 0 {
		return "", "", "", "", errors.New("empty youtube search result")
	}

	title := page.Items[0].Snippet.Title
	thumbnail := page.Items[0].Snippet.Thumbnails.Standard.Url
	channelName := page.Items[0].Snippet.ChannelTitle
	videoID := page.Items[0].ID.VideoId

	return title, thumbnail, channelName, videoID, nil
}

func isMusicEnabled(s *discordgo.Session, m *discordgo.MessageCreate) bool {

	if !youtubeAPIKeyPresent {
		s.ChannelMessageSend(m.ChannelID, "Music is currently disabled.")
		malm.Info("[Music] No Youtube API key provided in config")
	}

	return youtubeAPIKeyPresent
}
