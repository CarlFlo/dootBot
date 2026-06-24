package music

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var errEmptyStreamURL = errors.New("yt-dlp did not return a stream url")

func execYoutubeDL(song *Song) error {
	cmd := exec.Command("yt-dlp", song.GetYoutubeURL(), "--skip-download", "--get-url", "--no-playlist")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	streamURL, err := parseStreamURL(stdout.String())
	if err != nil {
		return err
	}

	song.StreamURL = streamURL
	return nil
}

func parseStreamURL(output string) (string, error) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line, nil
		}
	}

	return "", errEmptyStreamURL
}
