package music

import (
	"net/url"
	"testing"
	"time"

	"github.com/CarlFlo/dootBot/src/test"
)

func TestYoutubeTimeToDuration(t *testing.T) {

	input := "PT1H24M47S"
	duration := youtubeTimeToDuration(input)

	answer := time.Hour*1 + time.Minute*24 + time.Second*47

	test.Validate(t, duration, answer, "The durations should match")
}

func TestExtractVideoID(t *testing.T) {
	testCases := map[string]string{
		"https://www.youtube.com/watch?v=abc123": "abc123",
		"https://youtu.be/xyz987":                "xyz987",
	}

	for rawURL, expectedVideoID := range testCases {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			t.Fatalf("unable to parse test url %q: %s", rawURL, err)
		}

		videoID, err := extractVideoID(parsedURL)
		if err != nil {
			t.Fatalf("extractVideoID returned unexpected error for %q: %s", rawURL, err)
		}

		test.Validate(t, videoID, expectedVideoID, "the video IDs should match")
	}
}
