package music

import (
	"fmt"
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

/* SONG */

// GetDuration returns the duration of the song
func (s *Song) GetDuration() string {
	return s.duration
}

// GetYoutubeURL returns the full youtube url of the song
func (s *Song) GetYoutubeURL() string {

	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", s.YoutubeVideoID)
}
