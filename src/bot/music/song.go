package music

import (
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgolink/v4/lavalink"
)

type Song struct {
	ChannelID   string
	User        string
	Thumbnail   string
	ChannelName string
	Title       string
	URL         string
	Duration    time.Duration
	Track       lavalink.Track
}

func NewSongFromTrack(track lavalink.Track, channelID, userID string) *Song {
	song := &Song{
		ChannelID:   channelID,
		User:        userID,
		Title:       track.Info.Title,
		ChannelName: track.Info.Author,
		Duration:    time.Duration(track.Info.Length.Milliseconds()) * time.Millisecond,
		Track:       track,
	}

	if track.Info.URI != nil {
		song.URL = *track.Info.URI
	}

	if track.Info.ArtworkURL != nil {
		song.Thumbnail = *track.Info.ArtworkURL
	}

	return song
}

func (s *Song) GetDuration() string {
	output := fmt.Sprintf("%s", s.Duration)
	output = strings.Replace(output, "h", "h ", 1)
	output = strings.Replace(output, "m", "m ", 1)
	return output
}

func (s *Song) GetYoutubeURL() string {
	return s.URL
}
