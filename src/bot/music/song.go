package music

import (
	"fmt"
	"strings"
	"sync"
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

	streamFetchMu      sync.Mutex
	streamFetchRunning bool
	streamFetchErr     error
	streamFetchWait    chan struct{}
}

func (s *Song) FetchStreamURL() error {
	s.streamFetchMu.Lock()

	if s.StreamURL != "" {
		s.streamFetchMu.Unlock()
		return nil
	}

	if s.streamFetchRunning {
		wait := s.streamFetchWait
		s.streamFetchMu.Unlock()

		<-wait

		s.streamFetchMu.Lock()
		defer s.streamFetchMu.Unlock()
		return s.streamFetchErr
	}

	s.streamFetchRunning = true
	s.streamFetchErr = nil
	s.streamFetchWait = make(chan struct{})
	wait := s.streamFetchWait
	s.streamFetchMu.Unlock()

	err := execYoutubeDL(s)
	if err == nil {
		var cache database.YoutubeCache
		if cacheErr := cache.UpdateStreamURL(s.YoutubeVideoID, s.StreamURL); cacheErr != nil {
			malm.Error("%s", cacheErr)
		}
	}

	s.streamFetchMu.Lock()
	s.streamFetchErr = err
	s.streamFetchRunning = false
	close(wait)
	s.streamFetchMu.Unlock()

	return err
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
