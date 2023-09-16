package music

import (
	"bytes"
	"encoding/json"
	"os/exec"

	"github.com/CarlFlo/malm"
)

type videoResponse struct {
	Formats []struct {
		Url string `json:"url"`
	} `json:"formats"`
}

func execYoutubeDL(song *Song) error {

	malm.Debug("Running youtube-DL")

	cmd := exec.Command("yt-dlp", song.YoutubeVideoID, "--skip-download", "--print-json", "--flat-playlist")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return err
	}

	var videoRes videoResponse
	err = json.NewDecoder(&out).Decode(&videoRes)
	if err != nil {
		return err
	}

	// The URL directly to the audio. Expires after 6 hours
	// yt-dlp uses index 3. Youtube-dl uses index 0
	song.StreamURL = videoRes.Formats[3].Url
	return nil
}
