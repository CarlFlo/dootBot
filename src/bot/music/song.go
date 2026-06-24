package music

import (
	"fmt"
	"strings"
	"time"

	"github.com/CarlFlo/dootBot/src/database"
	"github.com/CarlFlo/malm"
)

type Song struct {
	ChannelID      string
	User           string
	Thumbnail      string
	ChannelName    string
	Title          string
	YoutubeVideoID string
	StreamURL      string
	Duration       time.Duration
}

func (s *Song) FetchStreamURL() error {
	if s.StreamURL != "" {
		return nil
	}

	if err := execYoutubeDL(s); err != nil {
		return err
	}

	var cache database.YoutubeCache
	if err := cache.UpdateStreamURL(s.YoutubeVideoID, s.StreamURL); err != nil {
		malm.Error("%s", err)
	}

	return nil
}

func (s *Song) GetDuration() string {
	output := fmt.Sprintf("%s", s.Duration)
	output = strings.Replace(output, "h", "h ", 1)
	output = strings.Replace(output, "m", "m ", 1)
	return output
}

func (s *Song) GetYoutubeURL() string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", s.YoutubeVideoID)
}
