package utils

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/CarlFlo/dootBot/src/config"
)

var keyIndex = 0
var apiMutex sync.Mutex

func GetYoutubeAPIKey() string {

	apiMutex.Lock()
	defer apiMutex.Unlock()

	// Make a copy of the index to be used
	indexCopy := keyIndex

	// Update the index. Increment by 1
	keyIndex = (keyIndex + 1) % len(config.CONFIG.Music.YoutubeAPIKeys)

	return config.CONFIG.Music.YoutubeAPIKeys[indexCopy]
}

func ValidateYoutubeAPIKey() error {

	//endpoint := "https://www.googleapis.com/youtube/v3/search?part=snippet&q=YouTube+Data+API&type=video&key=%s"
	endpoint := "https://www.googleapis.com/youtube/v3/search?&key=%s"

	for i, apiKey := range config.CONFIG.Music.YoutubeAPIKeys {
		if len(apiKey) == 0 {
			return fmt.Errorf("at least one Youtube API key provided in the config is invalid (index: %d, key: '%s')", i, apiKey)
		}

		// Check if youtube will accept it
		res, err := http.Get(fmt.Sprintf(endpoint, apiKey))

		if err != nil {
			return err
		} else if res.StatusCode == 403 {
			return fmt.Errorf("status code: 403 (Forbidden) - check if the API key has exeeded its quota for key: '%s'", apiKey)
		} else if res.StatusCode != 200 {
			return fmt.Errorf("could not validate Youtube API key (key: '%s'). status code: %d", apiKey, res.StatusCode)
		}
	}

	return nil
}
