package music

import (
	"testing"

	"github.com/CarlFlo/dootBot/src/test"
)

func TestParseStreamURL(t *testing.T) {
	streamURL, err := parseStreamURL("\nhttps://stream.example/audio\n")
	if err != nil {
		t.Fatalf("parseStreamURL returned unexpected error: %s", err)
	}

	test.Validate(t, streamURL, "https://stream.example/audio", "expected the first non-empty line to be used")
}

func TestParseStreamURLEmpty(t *testing.T) {
	_, err := parseStreamURL("  \n\t\n")
	if err == nil {
		t.Fatal("parseStreamURL should fail for empty output")
	}
}
