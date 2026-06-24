package music

import (
	"testing"

	"github.com/CarlFlo/dootBot/src/test"
)

func TestLooksLikeURL(t *testing.T) {
	test.Validate(t, looksLikeURL("https://example.com/audio"), true, "https URLs should be detected")
	test.Validate(t, looksLikeURL("plain search query"), false, "plain text should not be treated as a URL")
}

func TestBuildTrackIdentifier(t *testing.T) {
	test.Validate(t, buildTrackIdentifier("https://example.com/audio"), "https://example.com/audio", "urls should pass through unchanged")
	test.Validate(t, buildTrackIdentifier("never gonna give you up"), "ytsearch:never gonna give you up", "searches should use Lavalink's default source prefix")
}
