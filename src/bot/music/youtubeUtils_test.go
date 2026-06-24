package music

import (
	"testing"

	"github.com/CarlFlo/dootBot/src/test"
	"github.com/disgoorg/disgolink/v4/lavalink"
)

func TestNewSongFromTrackPreservesMetadata(t *testing.T) {
	url := "https://example.com/track"
	artwork := "https://example.com/image.jpg"
	track := testTrack(url, artwork)

	song := NewSongFromTrack(track, "channel", "user")

	test.Validate(t, song.Title, "test track", "title should come from the lavalink track")
	test.Validate(t, song.URL, url, "url should come from the lavalink track")
	test.Validate(t, song.Thumbnail, artwork, "artwork should come from the lavalink track")
}

func testTrack(url, artwork string) lavalink.Track {
	return lavalink.Track{
		Encoded: "encoded",
		Info: lavalink.TrackInfo{
			Title:      "test track",
			Author:     "test author",
			Length:     5000,
			URI:        &url,
			ArtworkURL: &artwork,
		},
	}
}
