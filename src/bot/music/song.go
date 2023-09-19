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
	User           string // Who requested the song
	Thumbnail      string
	ChannelName    string
	Title          string
	YoutubeVideoID string
	StreamURL      string
	Duration       time.Duration
}

func (s *Song) FetchStreamURL() error {

	// Todo. Move cache to the DB.
	// if streamURL was valid. Then URL should be in the object

	// We have a song in the cache
	if len(s.StreamURL) != 0 {
		return nil
	}

	// song.StreamURL contains the URL to the stream.

	// This function is slow. Takes a bit over 2 seconds
	if err := execYoutubeDL(s); err != nil {
		return err
	}

	var cache database.YoutubeCache
	err := cache.UpdateStreamURL(s.YoutubeVideoID, s.StreamURL)
	if err != nil {
		malm.Error("%s", err)
	}

	return nil
}

/*
func (s *Song) FetchStreamURL() error {

	// Todo. Move cache to the DB.
	// if streamURL was valid. Then URL should be in the object

	// song.StreamURL contains the URL to the stream.
	if streamURL := songCache.Check(s.YoutubeVideoID); len(streamURL) == 0 {
		// This function is slow. Takes a bit over 2 seconds
		if err := execYoutubeDL(s); err != nil {
			return err
		}
		songCache.Add(s)
		malm.Debug("[%s] cached - %s", s.Title, s.YoutubeVideoID)

		// Call:
		// func (c *YoutubeCache) UpdateStreamURL(videoID, streamURL string) {

	} else {
		s.StreamURL = streamURL
	}

	return nil
}
*/

/* SONG */

// GetDuration returns the duration of the song
func (s *Song) GetDuration() string {

	output := fmt.Sprintf("%s", s.Duration)

	output = strings.Replace(output, "h", "h ", 1)
	output = strings.Replace(output, "m", "m ", 1)
	// no need to do seconds

	return output
}

// GetYoutubeURL returns the full youtube url of the song
func (s *Song) GetYoutubeURL() string {

	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", s.YoutubeVideoID)
}
