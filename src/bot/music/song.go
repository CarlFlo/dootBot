package music

import (
	"fmt"

	"github.com/CarlFlo/malm"
)

type Song struct {
	ChannelID      string
	User           string // Who requested the song
	Thumbnail      string
	ChannelName    string
	Title          string
	YoutubeVideoID string
	StreamURL      string
	duration       string
}

func (s *Song) FetchStreamURL() error {

	// song.StreamURL contains the URL to the stream.
	if streamURL := songCache.Check(s.YoutubeVideoID); len(streamURL) == 0 {
		// This function is slow. Takes a bit over 2 seconds
		if err := execYoutubeDL(s); err != nil {
			return err
		}
		songCache.Add(s)
	} else {
		s.StreamURL = streamURL
	}
	malm.Debug("[%s] cached and prepared", s.YoutubeVideoID)
	return nil
}

/* SONG */

// GetDuration returns the duration of the song
func (s *Song) GetDuration() string {
	return s.duration
}

// GetYoutubeURL returns the full youtube url of the song
func (s *Song) GetYoutubeURL() string {

	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", s.YoutubeVideoID)
}
