package utils

import (
	"fmt"
	"net/http"

	"github.com/CarlFlo/DiscordMoneyBot/src/config"
)

func ValidateYoutubeAPIKey() error {

	if len(config.CONFIG.Music.YoutubeAPIKey) == 0 {
		return fmt.Errorf("no Youtube API key provided in the config")
	}

	endpoint := "https://www.googleapis.com/youtube/v3/search?part=snippet&q=YouTube+Data+API&type=video&key=%s"

	res, err := http.Get(fmt.Sprintf(endpoint, config.CONFIG.Music.YoutubeAPIKey))

	if err != nil {
		return err
	} else if res.StatusCode != 200 {
		return fmt.Errorf("could not validate Youtube API key. status code: %d", res.StatusCode)
	}

	return nil
}
