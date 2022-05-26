package music

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

const (
	youtubeFindEndpoint   string = "https://www.googleapis.com/youtube/v3/videos?part=snippet&key=%s&id=%s"
	youtubeSearchEndpoint string = "https://www.googleapis.com/youtube/v3/search?part=snippet&type=video&key=%s&q=%s"
)

type ytPage struct {
	Items []struct {
		Snippet struct {
			Title string
		}
		Thumbnails struct {
			Standard struct {
				Url    string
				Width  string
				Height string
			}
		}
		ChannelTitle string
	}
}

/*
type ytPage struct {
	Items []itemsFind
}

type itemsFind struct {
	Snippet snippet
}

type snippet struct {
	Title string
}
*/

// Returns the title, thumbnail and channel of a youtube video
// error if there was any problem
func youtubeFindByVideoID(videoID string) (string, string, string, error) {

	res, err := http.Get(fmt.Sprintf(youtubeFindEndpoint, config.CONFIG.Music.YoutubeAPIKey, videoID))
	if err != nil {
		return "", "", "", err
	}
	defer res.Body.Close()

	var page ytPage

	err = json.NewDecoder(res.Body).Decode(&page)
	if err != nil {
		return "", "", "", err
	}

	if len(page.Items) == 0 {
		return "", "", "", errors.New("empty youtube search result")
	}

	title := page.Items[0].Snippet.Title
	thumbnail := page.Items[0].Thumbnails.Standard.Url
	channelName := page.Items[0].ChannelTitle

	return title, thumbnail, channelName, nil
}

func isMusicEnabled(s *discordgo.Session, m *discordgo.MessageCreate) bool {

	if !youtubeAPIKeyPresent {
		s.ChannelMessageSend(m.ChannelID, "Music is currently disabled.")
		malm.Info("[Music] No Youtube API key provided in config")
	}

	return youtubeAPIKeyPresent
}
